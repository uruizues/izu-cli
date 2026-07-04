package animepahe

import (
	"context"
	"encoding/json"

	"github.com/izu/izu-cli/internal/provider"
)

type episodeResponse struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	Results []struct {
		ID       string  `json:"id"`
		Session  string  `json:"session"`
		Number   float64 `json:"number"`
		Title    string  `json:"title"`
		Duration string  `json:"duration"`
		Snapshot string  `json:"snapshot"`
	} `json:"data"`
}

func (c *Client) GetAnime(ctx context.Context, id string) (*provider.Anime, error) {
	return &provider.Anime{
		ID: id,
	}, nil
}

func (c *Client) GetEpisodes(ctx context.Context, animeID string, page int) (*provider.EpisodePage, error) {
	data, err := c.GetRelease(animeID, page)
	if err != nil {
		return nil, err
	}

	var resp episodeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var episodes []provider.Episode
	for _, item := range resp.Results {
		episodes = append(episodes, provider.Episode{
			ID:       item.Session,
			Number:   int(item.Number),
			Title:    item.Title,
			Duration: item.Duration,
			Snapshot: item.Snapshot,
		})
	}

	return &provider.EpisodePage{
		Episodes:    episodes,
		TotalPages:  (resp.Total + 24) / 25,
		CurrentPage: resp.Page,
		HasNext:     resp.Page*25 < resp.Total,
	}, nil
}
