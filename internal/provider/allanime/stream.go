package allanime

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/izu/izu-cli/internal/provider"
)

const episodeQuery = `query ($showId: String!, $translationType: VaildTranslationTypeEnumType!, $episodeString: String!) {
  episode(showId: $showId translationType: $translationType episodeString: $episodeString) {
    episodeString
    sourceUrls
  }
}`

type episodeResponse struct {
	Data struct {
		Episode struct {
			EpisodeString string   `json:"episodeString"`
			SourceUrls    []string `json:"sourceUrls"`
		} `json:"episode"`
	} `json:"data"`
	Tobeparsed string `json:"tobeparsed"`
}

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

	// Try to decrypt if tobeparsed is present
	var resp episodeResponse
	if err := json.Unmarshal(data, &resp); err == nil && resp.Tobeparsed != "" {
		decrypted, err := decryptTobeparsed(resp.Tobeparsed)
		if err == nil {
			json.Unmarshal([]byte(decrypted), &resp)
		}
	}

	// Extract source URLs
	info := &provider.StreamInfo{
		Referer: Referer,
		Headers: map[string]string{
			"Referer": Referer,
			"Origin":  Referer,
		},
		Format: "hls",
	}

	if len(resp.Data.Episode.SourceUrls) > 0 {
		sourceURL := resp.Data.Episode.SourceUrls[0]

		// Decode hex mapping if needed
		if strings.HasPrefix(sourceURL, "--") {
			decoded, err := decodeHexMapping(sourceURL)
			if err == nil {
				sourceURL = decoded
			}
		}

		info.URL = ExtractProviderURL(CleanProviderPath(sourceURL))
	}

	return info, nil
}
