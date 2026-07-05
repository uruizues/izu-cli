package consumet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/izu/izu-cli/internal/provider"
)

type searchResponse struct {
	HasNextPage bool `json:"hasNextPage"`
	Results     []struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		URL         string `json:"url"`
		Image       string `json:"image"`
		ReleaseDate string `json:"releaseDate"`
		SubOrDub    string `json:"subOrDub"`
	} `json:"results"`
}

func (c *Client) Search(ctx context.Context, query string) ([]provider.SearchResult, error) {
	if c.provider == "multi" {
		return c.searchAllProviders(ctx, query)
	}

	return c.searchSingle(ctx, query)
}

func (c *Client) searchSingle(ctx context.Context, query string) ([]provider.SearchResult, error) {
	u := fmt.Sprintf("%s/anime/%s/%s?page=1", c.baseURL, c.provider, url.PathEscape(query))

	data, err := c.doRequest(u)
	if err != nil {
		return nil, err
	}

	var resp searchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var results []provider.SearchResult
	for _, r := range resp.Results {
		results = append(results, provider.SearchResult{
			ID:    r.ID,
			Title: r.Title,
			Image: r.Image,
			Type:  r.SubOrDub,
		})
	}

	return results, nil
}
