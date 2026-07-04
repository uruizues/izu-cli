package animepahe

import (
	"context"
	"encoding/json"

	"github.com/izu/izu-cli/internal/provider"
)

type searchResponse struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	Results []struct {
		ID       string `json:"id"`
		Session  string `json:"session"`
		Title    string `json:"title"`
		Image    string `json:"image"`
		Type     string `json:"type"`
		Episodes int    `json:"episodes"`
		Status   string `json:"status"`
	} `json:"data"`
}

func (c *Client) Search(ctx context.Context, query string) ([]provider.SearchResult, error) {
	data, err := c.doSearch(query)
	if err != nil {
		return nil, err
	}

	var resp searchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var results []provider.SearchResult
	for _, item := range resp.Results {
		results = append(results, provider.SearchResult{
			ID:       item.Session,
			Title:    item.Title,
			Image:    item.Image,
			Type:     item.Type,
			Episodes: item.Episodes,
			Status:   item.Status,
		})
	}

	return results, nil
}
