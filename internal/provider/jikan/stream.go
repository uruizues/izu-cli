package jikan

import (
	"context"
	"fmt"

	"github.com/izu/izu-cli/internal/provider"
)

func (c *Client) GetStream(ctx context.Context, episodeID string) (*provider.StreamInfo, error) {
	// Jikan doesn't provide streaming links
	// This provider is for search/metadata only
	return nil, fmt.Errorf("jikan does not provide streaming links - use consumet for streaming")
}
