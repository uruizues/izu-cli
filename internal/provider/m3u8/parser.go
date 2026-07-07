package m3u8

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// StreamVariant represents a single quality variant from a master M3U8 playlist.
type StreamVariant struct {
	URL        string
	Resolution string // e.g. "1920x1080"
	Bandwidth  int
	Codecs     string
	FrameRate  float64
}

// PlaylistInfo holds metadata about a parsed M3U8 playlist.
type PlaylistInfo struct {
	IsMaster  bool
	Variants  []StreamVariant
	Segments  int
	TotalTime time.Duration
	BaseURL   string
}

// FetchAndParse fetches an M3U8 URL and parses it as a master or media playlist.
func FetchAndParse(client *http.Client, m3u8URL string, headers map[string]string) (*PlaylistInfo, error) {
	req, err := http.NewRequest("GET", m3u8URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch m3u8: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("m3u8 fetch returned %d", resp.StatusCode)
	}

	return Parse(resp.Body, m3u8URL)
}

// Parse reads an M3U8 playlist from r, using baseURL to resolve relative URLs.
func Parse(r io.Reader, baseURL string) (*PlaylistInfo, error) {
	info := &PlaylistInfo{
		BaseURL: baseURL,
	}

	scanner := bufio.NewScanner(r)
	var currentBandwidth int
	var currentResolution string
	var currentCodecs string
	var currentFrameRate float64
	var foundExtM3U bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if line == "#EXTM3U" {
			foundExtM3U = true
			continue
		}

		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			info.IsMaster = true
			attrs := parseAttributes(line[len("#EXT-X-STREAM-INF:"):])
			if bw, ok := attrs["BANDWIDTH"]; ok {
				currentBandwidth, _ = strconv.Atoi(bw)
			}
			if res, ok := attrs["RESOLUTION"]; ok {
				currentResolution = res
			}
			if codecs, ok := attrs["CODECS"]; ok {
				currentCodecs = strings.Trim(codecs, "\"")
			}
			if fr, ok := attrs["FRAME-RATE"]; ok {
				currentFrameRate, _ = strconv.ParseFloat(fr, 64)
			}
			continue
		}

		// Non-comment, non-tag line after STREAM-INF is the URL
		if info.IsMaster && currentBandwidth > 0 {
			resolvedURL := resolveURL(baseURL, line)
			info.Variants = append(info.Variants, StreamVariant{
				URL:        resolvedURL,
				Resolution: currentResolution,
				Bandwidth:  currentBandwidth,
				Codecs:     currentCodecs,
				FrameRate:  currentFrameRate,
			})
			currentBandwidth = 0
			currentResolution = ""
			currentCodecs = ""
			currentFrameRate = 0
			continue
		}

		// Media playlist: count segments via EXTINF
		if strings.HasPrefix(line, "#EXTINF:") {
			info.Segments++
			// Extract duration from EXTINF:duration,title
			durationStr := line[len("#EXTINF:"):]
			if idx := strings.Index(durationStr, ","); idx >= 0 {
				durationStr = durationStr[:idx]
			}
			durationStr = strings.TrimSpace(durationStr)
			if d, err := strconv.ParseFloat(durationStr, 64); err == nil {
				info.TotalTime += time.Duration(d * float64(time.Second))
			}
			continue
		}
	}

	if !foundExtM3U && len(info.Variants) == 0 && info.Segments == 0 {
		return nil, fmt.Errorf("not a valid M3U8 playlist")
	}

	return info, nil
}

// BestVariant returns the highest-bandwidth variant, or the first if all have 0 bandwidth.
func (p *PlaylistInfo) BestVariant() *StreamVariant {
	if len(p.Variants) == 0 {
		return nil
	}
	best := &p.Variants[0]
	for i := range p.Variants {
		if p.Variants[i].Bandwidth > best.Bandwidth {
			best = &p.Variants[i]
		}
	}
	return best
}

// parseAttributes parses EXT-X-STREAM-INF attributes like BANDWIDTH=1000000,RESOLUTION=1920x1080
func parseAttributes(s string) map[string]string {
	attrs := make(map[string]string)
	// Handle quoted strings
	for {
		eqIdx := strings.Index(s, "=")
		if eqIdx < 0 {
			break
		}
		key := strings.TrimSpace(s[:eqIdx])
		s = s[eqIdx+1:]

		if len(s) > 0 && s[0] == '"' {
			// Quoted value
			endIdx := strings.Index(s[1:], "\"")
			if endIdx < 0 {
				attrs[key] = strings.Trim(s, "\"")
				break
			}
			attrs[key] = s[1 : endIdx+1]
			s = s[endIdx+2:]
			if len(s) > 0 && s[0] == ',' {
				s = s[1:]
			}
		} else {
			commaIdx := strings.Index(s, ",")
			if commaIdx < 0 {
				attrs[key] = strings.TrimSpace(s)
				break
			}
			attrs[key] = strings.TrimSpace(s[:commaIdx])
			s = s[commaIdx+1:]
		}
	}
	return attrs
}

// resolveURL resolves a potentially relative URL against a base.
func resolveURL(base, ref string) string {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	u, err := url.Parse(base)
	if err != nil {
		return ref
	}
	refU, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return u.ResolveReference(refU).String()
}
