package anilist

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/izu/izu-cli/internal/provider"
)

func (c *Client) GetStream(ctx context.Context, episodeID string) (*provider.StreamInfo, error) {
	// Parse episode ID format: "animeID:episodeIndex"
	parts := strings.SplitN(episodeID, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid episode ID format: %s", episodeID)
	}

	animeID := parts[0]
	epIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid episode index: %s", parts[1])
	}

	// Get streaming episodes
	streamingEpisodesMu.RLock()
	streamEps, ok := streamingEpisodesMap[animeID]
	streamingEpisodesMu.RUnlock()
	if !ok {
		// Try to fetch anime data
		_, err := c.GetAnime(ctx, animeID)
		if err != nil {
			return nil, err
		}
		streamingEpisodesMu.RLock()
		streamEps = streamingEpisodesMap[animeID]
		streamingEpisodesMu.RUnlock()
	}

	if epIndex >= len(streamEps) {
		return nil, fmt.Errorf("episode index %d out of range (max: %d)", epIndex, len(streamEps)-1)
	}

	ep := streamEps[epIndex]

	return &provider.StreamInfo{
		URL:     ep.URL,
		Quality: "best",
		Format:  "hls",
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		},
	}, nil
}
