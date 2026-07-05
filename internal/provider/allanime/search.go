package allanime

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/izu/izu-cli/internal/provider"
)

const searchQuery = `query($search: SearchInput $limit: Int $page: Int $translationType: VaildTranslationTypeEnumType $countryOrigin: VaildCountryOriginEnumType) {
  shows(search: $search limit: $limit page: $page translationType: $translationType countryOrigin: $countryOrigin) {
    pageInfo { total }
    edges { _id name englishName nativeName availableEpisodes }
  }
}`

type searchResponse struct {
	Data struct {
		Shows struct {
			PageInfo struct {
				Total int `json:"total"`
			} `json:"pageInfo"`
			Edges []struct {
				ID                 string         `json:"_id"`
				Name               string         `json:"name"`
				EnglishName        string         `json:"englishName"`
				NativeName         string         `json:"nativeName"`
				AvailableEpisodes map[string]int `json:"availableEpisodes"`
			} `json:"edges"`
		} `json:"shows"`
	} `json:"data"`
}

func (c *Client) Search(ctx context.Context, query string) ([]provider.SearchResult, error) {
	variables := map[string]interface{}{
		"search": map[string]interface{}{
			"allowAdult":   false,
			"allowUnknown": false,
			"query":        query,
		},
		"limit":           26,
		"page":            1,
		"translationType": "sub",
		"countryOrigin":   "ALL",
	}

	data, err := c.doGraphQL(searchQuery, variables)
	if err != nil {
		return nil, err
	}

	var resp searchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var results []provider.SearchResult
	for _, edge := range resp.Data.Shows.Edges {
		epCount := edge.AvailableEpisodes["sub"]

		title := edge.Name
		if edge.EnglishName != "" {
			title = edge.EnglishName
		}

		results = append(results, provider.SearchResult{
			ID:       edge.ID,
			Title:    title,
			Type:     "TV",
			Episodes: epCount,
			Status:   "Airing",
		})
	}

	return results, nil
}
