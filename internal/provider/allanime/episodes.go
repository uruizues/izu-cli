package allanime

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/izu/izu-cli/internal/provider"
)

const showQuery = `query ($showId: String!) {
  show(_id: $showId) {
    _id name englishName nativeName thumbnail description status
    availableEpisodes
  }
}`

type showResponse struct {
	Data struct {
		Show struct {
			ID                string            `json:"_id"`
			Name              string            `json:"name"`
			EnglishName       string            `json:"englishName"`
			NativeName        string            `json:"nativeName"`
			Thumbnail         string            `json:"thumbnail"`
			Description       string            `json:"description"`
			Status            string            `json:"status"`
			AvailableEpisodes map[string]int    `json:"availableEpisodes"`
		} `json:"show"`
	} `json:"data"`
}

func (c *Client) GetAnime(ctx context.Context, id string) (*provider.Anime, error) {
	variables := map[string]interface{}{
		"showId": id,
	}

	data, err := c.doGraphQL(showQuery, variables)
	if err != nil {
		return nil, err
	}

	var resp showResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	show := resp.Data.Show

	anime := &provider.Anime{
		ID:          show.ID,
		Title:       show.EnglishName,
		Japanese:    show.Name,
		Description: show.Description,
		Image:       show.Thumbnail,
		Type:        "TV",
		Status:      show.Status,
		Episodes:    show.AvailableEpisodes["sub"],
	}

	return anime, nil
}

func (c *Client) GetEpisodes(ctx context.Context, animeID string, page int) (*provider.EpisodePage, error) {
	anime, err := c.GetAnime(ctx, animeID)
	if err != nil {
		return nil, err
	}

	totalEpisodes := anime.Episodes
	if totalEpisodes == 0 {
		totalEpisodes = 1
	}

	var episodes []provider.Episode
	for i := 1; i <= totalEpisodes; i++ {
		epStr := strconv.Itoa(i)
		episodes = append(episodes, provider.Episode{
			ID:     animeID + "_ep" + epStr,
			Number: i,
			Title:  "Episode " + epStr,
		})
	}

	pageSize := 25
	start := (page - 1) * pageSize
	if start >= len(episodes) {
		start = len(episodes)
	}
	end := start + pageSize
	if end > len(episodes) {
		end = len(episodes)
	}

	return &provider.EpisodePage{
		Episodes:    episodes[start:end],
		TotalPages:  (len(episodes) + pageSize - 1) / pageSize,
		CurrentPage: page,
		HasNext:     end < len(episodes),
	}, nil
}
