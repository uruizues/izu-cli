package allanime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/izu/izu-cli/internal/provider"
	"github.com/izu/izu-cli/internal/provider/m3u8"
)

const episodeQuery = `query ($showId: String!, $translationType: VaildTranslationTypeEnumType!, $episodeString: String!) {
  episode(showId: $showId translationType: $translationType episodeString: $episodeString) {
    episodeString
    sourceUrls
    sourceNames
  }
}`

type episodeResponse struct {
	Data struct {
		Episode struct {
			EpisodeString string           `json:"episodeString"`
			SourceUrls    []sourceURLInfo  `json:"sourceUrls"`
		} `json:"episode"`
	} `json:"data"`
}

type sourceURLInfo struct {
	SourceName string `json:"sourceName"`
	SourceURL  string `json:"sourceUrl"`
}

type sourceProviderResponse struct {
	Links []sourceLink `json:"links"`
}

type sourceLink struct {
	Link      string          `json:"link"`
	Subtitles []subtitleEntry `json:"subtitles,omitempty"`
	Headers   struct {
		Referer string `json:"Referer"`
	} `json:"headers,omitempty"`
}

type subtitleEntry struct {
	Src   string `json:"src"`
	Label string `json:"label"`
	Lang  string `json:"lang"`
}

// Preferred source providers in order of preference
var preferredProviders = []string{"Yt-mp4", "S-Mp4", "Uv-mp4", "Ak", "Default"}

func (c *Client) GetStream(ctx context.Context, episodeID string) (*provider.StreamInfo, error) {
	// Parse episodeID to get showId and episode number
	// Format: {animeId}_ep{episodeNumber}
	parts := strings.SplitN(episodeID, "_ep", 2)
	if len(parts) != 2 {
		return &provider.StreamInfo{}, nil
	}
	showID := parts[0]
	episodeString := parts[1]

	variables := map[string]interface{}{
		"showId":          showID,
		"translationType": "sub",
		"episodeString":   episodeString,
	}

	data, err := c.doGraphQL(episodeQuery, variables)
	if err != nil {
		return nil, err
	}

	// Parse response, handling tobeparsed decryption
	parsed, err := ParseEpisodeResponse(data)
	if err != nil {
		return nil, fmt.Errorf("parse episode response: %w", err)
	}

	// Extract source URLs from the parsed response
	sources, err := extractSourceURLs(parsed)
	if err != nil {
		return nil, fmt.Errorf("extract source URLs: %w", err)
	}

	if len(sources) == 0 {
		return &provider.StreamInfo{
			Referer: Referer,
			Headers: map[string]string{
				"Referer": Referer,
				"Origin":  Referer,
			},
		}, nil
	}

	// Try each provider in preference order
	for _, prefName := range preferredProviders {
		for _, src := range sources {
			if src.SourceName != prefName {
				continue
			}
			streamInfo, err := c.resolveSource(ctx, src)
			if err != nil {
				continue
			}
			if streamInfo != nil && streamInfo.URL != "" {
				return streamInfo, nil
			}
		}
	}

	// Fallback: try first available source
	for _, src := range sources {
		streamInfo, err := c.resolveSource(ctx, src)
		if err != nil {
			continue
		}
		if streamInfo != nil && streamInfo.URL != "" {
			return streamInfo, nil
		}
	}

	return &provider.StreamInfo{
		Referer: Referer,
		Headers: map[string]string{
			"Referer": Referer,
			"Origin":  Referer,
		},
		Format: "hls",
	}, nil
}

func extractSourceURLs(parsed map[string]interface{}) ([]sourceURLInfo, error) {
	dataMap, ok := parsed["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing data field")
	}

	episodeMap, ok := dataMap["episode"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing episode field")
	}

	// sourceUrls can be a JSON array of objects or an array of strings
	switch urls := episodeMap["sourceUrls"].(type) {
	case []interface{}:
		var sources []sourceURLInfo
		for _, u := range urls {
			switch v := u.(type) {
			case map[string]interface{}:
				sources = append(sources, sourceURLInfo{
					SourceName: getString(v, "sourceName"),
					SourceURL:  getString(v, "sourceUrl"),
				})
			case string:
				sources = append(sources, sourceURLInfo{
					SourceURL: v,
				})
			}
		}
		return sources, nil
	default:
		return nil, fmt.Errorf("unexpected sourceUrls type")
	}
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c *Client) resolveSource(ctx context.Context, src sourceURLInfo) (*provider.StreamInfo, error) {
	sourceURL := src.SourceURL

	// Handle fast4speed URLs directly
	if strings.Contains(sourceURL, "tools.fast4speed.rsvp") {
		return &provider.StreamInfo{
			URL:     sourceURL,
			Referer: Referer,
			Headers: map[string]string{
				"Referer": Referer,
				"Origin":  Referer,
			},
			Quality: "1080",
			Format:  "hls",
		}, nil
	}

	// Decrypt the source URL path
	decryptedPath, err := decryptSourceURL(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("decrypt source URL: %w", err)
	}

	// Clean up the path
	decryptedPath = CleanProviderPath(decryptedPath)
	fullURL := ExtractProviderURL(decryptedPath)

	// Fetch the provider's JSON response
	providerData, err := c.fetchProviderJSON(ctx, fullURL)
	if err != nil {
		return nil, fmt.Errorf("fetch provider JSON: %w", err)
	}

	return c.processProviderLinks(ctx, providerData)
}

func (c *Client) fetchProviderJSON(ctx context.Context, url string) (*sourceProviderResponse, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Referer", "https://allmanga.to/")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Duration(attempt+1) * time.Millisecond)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if len(body) == 0 {
			lastErr = fmt.Errorf("empty response")
			time.Sleep(500 * time.Duration(attempt+1) * time.Millisecond)
			continue
		}

		var result sourceProviderResponse
		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = err
			continue
		}

		return &result, nil
	}
	return nil, fmt.Errorf("failed after 3 attempts: %w", lastErr)
}

func (c *Client) processProviderLinks(ctx context.Context, data *sourceProviderResponse) (*provider.StreamInfo, error) {
	if len(data.Links) == 0 {
		return nil, fmt.Errorf("no links in provider response")
	}

	link := data.Links[0]

	// Handle repackager.wixmp.com quality variants
	if strings.Contains(link.Link, "repackager.wixmp.com") {
		return c.processRepackagerLink(link.Link), nil
	}

	// Collect subtitles
	var subtitles []provider.Subtitle
	for _, sub := range link.Subtitles {
		subtitles = append(subtitles, provider.Subtitle{
			URL:    sub.Src,
			Lang:   sub.Lang,
			Label:  sub.Label,
			Format: "vtt",
		})
	}

	referer := link.Headers.Referer
	if referer == "" {
		referer = Referer
	}

	// Check if the link is itself an M3U8 URL
	if strings.HasSuffix(link.Link, ".m3u8") || strings.Contains(link.Link, ".m3u8") {
		return c.resolveM3U8(ctx, link.Link, referer, subtitles)
	}

	// Try to fetch as M3U8 anyway (some providers don't have .m3u8 extension)
	return c.resolveM3U8(ctx, link.Link, referer, subtitles)
}

func (c *Client) processRepackagerLink(link string) *provider.StreamInfo {
	// repackager.wixmp.com URLs have quality variants in the path
	// Format: prefix,quality1,quality2,...,suffix
	parts := strings.Split(link, ".urlset")
	if len(parts) < 2 {
		return &provider.StreamInfo{URL: link, Format: "hls"}
	}

	prefix := strings.Replace(parts[0], "repackager.wixmp.com/", "", 1)
	rest := strings.Split(parts[1], "/")
	if len(rest) < 2 {
		return &provider.StreamInfo{URL: link, Format: "hls"}
	}

	qualPart := rest[1]
	qualParts := strings.Split(qualPart, ",")
	if len(qualParts) < 3 {
		// Single quality
		quality := qualParts[0]
		return &provider.StreamInfo{
			URL:     prefix + quality + "." + strings.Join(rest[2:], "/"),
			Quality: strings.TrimSuffix(quality, "p"),
			Format:  "hls",
			Referer: Referer,
		}
	}

	// Multiple qualities - return the highest
	bestQuality := ""
	bestInt := 0
	for _, q := range qualParts {
		q = strings.TrimSuffix(q, "p")
		num := 0
		fmt.Sscanf(q, "%d", &num)
		if num > bestInt {
			bestInt = num
			bestQuality = q
		}
	}

	suffix := strings.Join(rest[2:], "/")
	return &provider.StreamInfo{
		URL:     prefix + bestQuality + "." + suffix,
		Quality: bestQuality,
		Format:  "hls",
		Referer: Referer,
	}
}

func (c *Client) resolveM3U8(ctx context.Context, m3u8URL, referer string, subtitles []provider.Subtitle) (*provider.StreamInfo, error) {
	headers := map[string]string{
		"Referer": referer,
		"Origin":  referer,
	}

	playlist, err := m3u8.FetchAndParse(c.httpClient, m3u8URL, headers)
	if err != nil {
		// If M3U8 parsing fails, return the URL directly (mpv can handle it)
		return &provider.StreamInfo{
			URL:       m3u8URL,
			Referer:   referer,
			Headers:   headers,
			Subtitles: subtitles,
			Quality:   "1080",
			Format:    "hls",
		}, nil
	}

	if playlist.IsMaster && len(playlist.Variants) > 0 {
		// Master playlist - pick the best variant
		best := playlist.BestVariant()
		if best != nil {
			resolution := best.Resolution
			if idx := strings.Index(resolution, "x"); idx >= 0 {
				resolution = resolution[idx+1:]
			}
			return &provider.StreamInfo{
				URL:       best.URL,
				Referer:   referer,
				Headers:   headers,
				Subtitles: subtitles,
				Quality:   resolution,
				Format:    "hls",
			}, nil
		}
	}

	// Media playlist or no variants - return the original URL
	return &provider.StreamInfo{
		URL:       m3u8URL,
		Referer:   referer,
		Headers:   headers,
		Subtitles: subtitles,
		Quality:   "1080",
		Format:    "hls",
	}, nil
}
