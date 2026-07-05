package jikan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/izu/izu-cli/internal/provider"
)

type searchResponse struct {
	Data []struct {
		MalID    int `json:"mal_id"`
		URL      string `json:"url"`
		Images   map[string]struct {
			ImageURL      string `json:"image_url"`
			SmallImageURL string `json:"small_image_url"`
			LargeImageURL string `json:"large_image_url"`
		} `json:"images"`
		Approved   bool   `json:"approved"`
		Titles     []struct {
			Type string `json:"type"`
			Title string `json:"title"`
		} `json:"titles"`
		Title         string `json:"title"`
		TitleEnglish  string `json:"title_english"`
		TitleJapanese string `json:"title_japanese"`
		Type          string `json:"type"`
		Source        string `json:"source"`
		Episodes      int    `json:"episodes"`
		Status        string `json:"status"`
		Score         float64 `json:"score"`
	} `json:"data"`
}

func (c *Client) Search(ctx context.Context, query string) ([]provider.SearchResult, error) {
	u := fmt.Sprintf("%s/anime?q=%s&limit=10", BaseURL, url.QueryEscape(query))

	data, err := c.doRequest(u)
	if err != nil {
		return nil, err
	}

	var resp searchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var results []provider.SearchResult
	for _, anime := range resp.Data {
		title := anime.Title
		if anime.TitleEnglish != "" {
			title = anime.TitleEnglish
		}

		image := ""
		if img, ok := anime.Images["jpg"]; ok {
			image = img.LargeImageURL
		}

		results = append(results, provider.SearchResult{
			ID:       fmt.Sprintf("%d", anime.MalID),
			Title:    title,
			Image:    image,
			Type:     anime.Type,
			Episodes: anime.Episodes,
			Status:   anime.Status,
		})
	}

	return results, nil
}
