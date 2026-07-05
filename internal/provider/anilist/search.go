package anilist

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/izu/izu-cli/internal/provider"
)

const searchQuery = `query($search: String, $page: Int, $perPage: Int) {
  Page(page: $page, perPage: $perPage) {
    media(search: $search, type: ANIME) {
      id
      title {
        romaji
        english
        native
      }
      coverImage {
        large
      }
      format
      episodes
      status
    }
  }
}`

type searchResponse struct {
	Data struct {
		Page struct {
			Media []struct {
				ID       int `json:"id"`
				Title    struct {
					Romaji  string `json:"romaji"`
					English string `json:"english"`
					Native  string `json:"native"`
				} `json:"title"`
				CoverImage struct {
					Large string `json:"large"`
				} `json:"coverImage"`
				Format  string `json:"format"`
				Episodes int   `json:"episodes"`
				Status   string `json:"status"`
			} `json:"media"`
		} `json:"Page"`
	} `json:"data"`
}

func (c *Client) Search(ctx context.Context, query string) ([]provider.SearchResult, error) {
	variables := map[string]interface{}{
		"search":  query,
		"page":    1,
		"perPage": 10,
	}

	data, err := c.doQuery(searchQuery, variables)
	if err != nil {
		return nil, err
	}

	var resp searchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var results []provider.SearchResult
	for _, media := range resp.Data.Page.Media {
		title := media.Title.English
		if title == "" {
			title = media.Title.Romaji
		}

		results = append(results, provider.SearchResult{
			ID:       fmt.Sprintf("%d", media.ID),
			Title:    title,
			Image:    media.CoverImage.Large,
			Type:     media.Format,
			Episodes: media.Episodes,
			Status:   media.Status,
		})
	}

	return results, nil
}
