package miruro

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/izu/izu-cli/internal/player/proxy"
	"github.com/izu/izu-cli/internal/provider"
)

const DefaultBaseURL = "http://localhost:8000"

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Name() string {
	return "miruro"
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) doRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "izu-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

type anilistSearchResponse struct {
	Data struct {
		Page struct {
			Media []struct {
				ID       int `json:"id"`
				Title    struct {
					Romaji  string `json:"romaji"`
					English string `json:"english"`
				} `json:"title"`
				CoverImage struct {
					Large string `json:"large"`
				} `json:"coverImage"`
				Format   string `json:"format"`
				Episodes *int  `json:"episodes"`
				Status   string `json:"status"`
			} `json:"media"`
		} `json:"Page"`
	} `json:"data"`
}

func (c *Client) Search(ctx context.Context, query string) ([]provider.SearchResult, error) {
	gqlQuery := `query ($search: String) { Page(page: 1, perPage: 20) { media(search: $search, type: ANIME, sort: SEARCH_MATCH) { id title { romaji english } coverImage { large } format episodes status } } }`

	payload := map[string]interface{}{
		"query": gqlQuery,
		"variables": map[string]string{
			"search": query,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://graphql.anilist.co", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AniList unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AniList unavailable (status %d)", resp.StatusCode)
	}

	var result anilistSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	results := make([]provider.SearchResult, 0, len(result.Data.Page.Media))
	for _, m := range result.Data.Page.Media {
		title := m.Title.English
		if title == "" {
			title = m.Title.Romaji
		}
		episodes := 0
		if m.Episodes != nil {
			episodes = *m.Episodes
		}
		results = append(results, provider.SearchResult{
			ID:       fmt.Sprintf("%d", m.ID),
			Title:    title,
			Image:    m.CoverImage.Large,
			Type:     m.Format,
			Episodes: episodes,
			Status:   m.Status,
		})
	}

	return results, nil
}

type infoResponse struct {
	ID          int `json:"id"`
	Title       struct {
		Romaji  string `json:"romaji"`
		English string `json:"english"`
		Native  string `json:"native"`
	} `json:"title"`
	Description string `json:"description"`
	CoverImage  struct {
		Large string `json:"large"`
	} `json:"coverImage"`
	Format   string `json:"format"`
	Episodes int    `json:"episodes"`
	Status   string `json:"status"`
	Genres   []string `json:"genres"`
}

func (c *Client) GetAnime(ctx context.Context, id string) (*provider.Anime, error) {
	url := fmt.Sprintf("%s/info/%s", c.baseURL, id)
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var resp infoResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	title := resp.Title.English
	if title == "" {
		title = resp.Title.Romaji
	}

	return &provider.Anime{
		ID:          fmt.Sprintf("%d", resp.ID),
		Title:       title,
		Japanese:    resp.Title.Native,
		Description: resp.Description,
		Image:       resp.CoverImage.Large,
		Type:        resp.Format,
		Episodes:    resp.Episodes,
		Status:      resp.Status,
		Genres:      resp.Genres,
	}, nil
}

type episodesResponse struct {
	Providers map[string]struct {
		Episodes map[string][]struct {
			ID     string `json:"id"`
			Number int    `json:"number"`
			Title  string `json:"title"`
		} `json:"episodes"`
	} `json:"providers"`
}

func (c *Client) GetEpisodes(ctx context.Context, animeID string, page int) (*provider.EpisodePage, error) {
	url := fmt.Sprintf("%s/episodes/%s", c.baseURL, animeID)
	data, err := c.doRequest(url)
	if err == nil {
		var resp episodesResponse
		if json.Unmarshal(data, &resp) == nil && len(resp.Providers) > 0 {
			// Collect all episodes from all providers
			var allEpisodes []provider.Episode
			for _, provData := range resp.Providers {
				for _, epList := range provData.Episodes {
					for _, ep := range epList {
						allEpisodes = append(allEpisodes, provider.Episode{
							ID:     ep.ID,
							Number: ep.Number,
							Title:  ep.Title,
						})
					}
				}
			}

			if len(allEpisodes) > 0 {
				// Prefer animepahe episodes — group by episode number
				epMap := map[int]provider.Episode{}
				for _, ep := range allEpisodes {
					if _, exists := epMap[ep.Number]; !exists {
						epMap[ep.Number] = ep
					}
					// Override with animepahe if slug contains it
					if len(ep.ID) > 0 && contains(ep.ID, "animepahe") {
						epMap[ep.Number] = ep
					}
				}

				var episodes []provider.Episode
				for _, ep := range epMap {
					episodes = append(episodes, ep)
				}

				sort.Slice(episodes, func(i, j int) bool {
					return episodes[i].Number < episodes[j].Number
				})

				pageSize := 25
				start := (page - 1) * pageSize
				if start >= len(episodes) {
					start = len(episodes)
				}
				end := start + pageSize
				if end > len(episodes) {
					end = len(episodes)
				}

				return &provider.EpisodePage{
					Episodes:    episodes[start:end],
					TotalPages:  (len(episodes) + pageSize - 1) / pageSize,
					CurrentPage: page,
					HasNext:     end < len(episodes),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("episodes not available — AniList may be temporarily down. Try another anime")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type streamEntry struct {
	URL     string `json:"url"`
	Type    string `json:"type"`
	Quality string `json:"quality"`
	Referer string `json:"referer"`
}

type streamResponse struct {
	Streams []streamEntry `json:"streams"`
	Subtitles []struct {
		File  string `json:"file"`
		Label string `json:"label"`
	} `json:"subtitles"`
	Intro *struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"intro"`
	Outro *struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"outro"`
}

func (c *Client) GetStream(ctx context.Context, episodeID string) (*provider.StreamInfo, error) {
	// episodeID is a Miruro path like "watch/kiwi/20/sub/animepahe-1"
	return c.fetchStream(episodeID)
}

func (c *Client) fetchStream(path string) (*provider.StreamInfo, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, path)
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var resp streamResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Streams) == 0 {
		return nil, fmt.Errorf("no streams")
	}

	// Find the best stream
	var bestStream *streamEntry
	for i, s := range resp.Streams {
		if s.Type == "hls" {
			bestStream = &resp.Streams[i]
			break
		}
	}
	if bestStream == nil {
		bestStream = &resp.Streams[0]
	}

	referer := bestStream.Referer
	if referer == "" {
		referer = "https://kwik.cx/"
	}

	// Start proxy and rewrite URL through it
	proxy.Start(referer, referer)
	proxiedURL := proxy.ProxyURL(bestStream.URL)

	info := &provider.StreamInfo{
		URL:     proxiedURL,
		Quality: bestStream.Quality,
		Format:  "hls",
		Headers: map[string]string{},
		Referer: referer,
	}

	for _, sub := range resp.Subtitles {
		info.Subtitles = append(info.Subtitles, provider.Subtitle{
			URL:    sub.File,
			Label:  sub.Label,
			Lang:   sub.Label,
			Format: "vtt",
		})
	}

	return info, nil
}


