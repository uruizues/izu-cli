package allanime

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/izu/izu-cli/internal/provider"
)

const showQuery = `query ($showId: String!) {
  show(_id: $showId) {
    _id name englishName nativeName thumbnail description status
    availableEpisodes availableEpisodesDetail
  }
}`

type showResponse struct {
	Data struct {
		Show struct {
			ID                      string            `json:"_id"`
			Name                    string            `json:"name"`
			EnglishName             string            `json:"englishName"`
			NativeName              string            `json:"nativeName"`
			Thumbnail               string            `json:"thumbnail"`
			Description             string            `json:"description"`
			Status                  string            `json:"status"`
			AvailableEpisodes       map[string]string `json:"availableEpisodes"`
			AvailableEpisodesDetail map[string][]string `json:"availableEpisodesDetail"`
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
	}

	if sub, ok := show.AvailableEpisodes["sub"]; ok {
		fmt.Sscanf(sub, "%d", &anime.Episodes)
	}

	return anime, nil
}

func (c *Client) GetEpisodes(ctx context.Context, animeID string, page int) (*provider.EpisodePage, error) {
	anime, err := c.GetAnime(ctx, animeID)
	if err != nil {
		return nil, err
	}

	// Get episode list from show detail
	variables := map[string]interface{}{
		"showId": animeID,
	}

	data, err := c.doGraphQL(showQuery, variables)
	if err != nil {
		return nil, err
	}

	var resp showResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	epStrings := resp.Data.Show.AvailableEpisodesDetail["sub"]
	if len(epStrings) == 0 {
		epStrings = resp.Data.Show.AvailableEpisodesDetail["dub"]
	}

	var episodes []provider.Episode
	for _, epStr := range epStrings {
		epNum, _ := strconv.Atoi(epStr)
		episodes = append(episodes, provider.Episode{
			ID:     animeID + "_ep" + epStr,
			Number: epNum,
			Title:  "Episode " + epStr,
		})
	}

	// Simple pagination
	pageSize := 25
	start := (page - 1) * pageSize
	if start >= len(episodes) {
		start = len(episodes)
	}
	end := start + pageSize
	if end > len(episodes) {
		end = len(episodes)
	}

	_ = anime // used for metadata fetch above

	return &provider.EpisodePage{
		Episodes:    episodes[start:end],
		TotalPages:  (len(episodes) + pageSize - 1) / pageSize,
		CurrentPage: page,
		HasNext:     end < len(episodes),
	}, nil
}
