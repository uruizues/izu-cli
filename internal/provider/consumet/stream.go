package consumet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/izu/izu-cli/internal/provider"
)

type watchResponse struct {
	Sources []struct {
		URL    string `json:"url"`
		Quality string `json:"quality"`
		IsM3U8 bool   `json:"isM3U8"`
	} `json:"sources"`
	Subtitles []struct {
		URL  string `json:"url"`
		Lang string `json:"lang"`
	} `json:"subtitles"`
}

func (c *Client) GetStream(ctx context.Context, episodeID string) (*provider.StreamInfo, error) {
	u := fmt.Sprintf("%s/anime/%s/watch/%s", c.baseURL, c.provider, episodeID)

	data, err := c.doRequest(u)
	if err != nil {
		return nil, err
	}

	var resp watchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	info := &provider.StreamInfo{
		Referer: c.baseURL,
		Headers: map[string]string{
			"Referer": c.baseURL,
		},
	}

	if len(resp.Sources) > 0 {
		info.URL = resp.Sources[0].URL
		info.Quality = resp.Sources[0].Quality
		if resp.Sources[0].IsM3U8 {
			info.Format = "hls"
		} else {
			info.Format = "mp4"
		}
	}

	for _, sub := range resp.Subtitles {
		info.Subtitles = append(info.Subtitles, provider.Subtitle{
			URL:    sub.URL,
			Lang:   sub.Lang,
			Label:  sub.Lang,
			Format: "vtt",
		})
	}

	return info, nil
}
