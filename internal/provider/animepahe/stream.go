package animepahe

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/izu/izu-cli/internal/provider"
)

type playResponse struct {
	Quality string `json:"quality"`
	Source  string `json:"source"`
}

func (c *Client) GetStream(ctx context.Context, episodeID string) (*provider.StreamInfo, error) {
	data, err := c.GetPlay(episodeID)
	if err != nil {
		return nil, err
	}

	var playData map[string]interface{}
	if err := json.Unmarshal(data, &playData); err != nil {
		return nil, err
	}

	info := &provider.StreamInfo{
		Referer: c.baseURL + "/",
		Headers: map[string]string{
			"Referer": c.baseURL + "/",
		},
		Format: "hls",
	}

	// Extract real stream URL from kwik.cx if available
	if source, ok := playData["data"].([]interface{}); ok && len(source) > 0 {
		if sourceObj, ok := source[0].(map[string]interface{}); ok {
			if key, ok := sourceObj["key"].(string); ok {
				streamURL, err := extractStreamURL(ctx, c.httpClient, key)
				if err == nil {
					info.URL = streamURL
				}
			}
			if label, ok := sourceObj["label"].(string); ok {
				info.Quality = label
			}
		}
	}

	if info.URL == "" {
		return nil, fmt.Errorf("could not extract stream URL for episode %s", episodeID)
	}

	return info, nil
}
