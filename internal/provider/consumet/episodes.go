package consumet

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/izu/izu-cli/internal/provider"
)

type infoResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	JapaneseTitle string `json:"japaneseTitle"`
	Image        string `json:"image"`
	Description  string `json:"description"`
	ReleaseDate  string `json:"releaseDate"`
	Genres       []string `json:"genres"`
	OtherTitles  []string `json:"otherTitles"`
	Episodes     []struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Number string `json:"number"`
	} `json:"episodes"`
}

func (c *Client) GetAnime(ctx context.Context, id string) (*provider.Anime, error) {
	u := fmt.Sprintf("%s/anime/%s/info/%s", c.baseURL, c.provider, id)

	data, err := c.doRequest(u)
	if err != nil {
		return nil, err
	}

	var resp infoResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &provider.Anime{
		ID:          resp.ID,
		Title:       resp.Title,
		Japanese:    resp.JapaneseTitle,
		Description: resp.Description,
		Image:       resp.Image,
		Genres:      resp.Genres,
		Episodes:    len(resp.Episodes),
	}, nil
}

func (c *Client) GetEpisodes(ctx context.Context, animeID string, page int) (*provider.EpisodePage, error) {
	u := fmt.Sprintf("%s/anime/%s/info/%s", c.baseURL, c.provider, animeID)

	data, err := c.doRequest(u)
	if err != nil {
		return nil, err
	}

	var resp infoResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var episodes []provider.Episode
	for _, ep := range resp.Episodes {
		num, _ := strconv.Atoi(ep.Number)
		episodes = append(episodes, provider.Episode{
			ID:     ep.ID,
			Number: num,
			Title:  ep.Title,
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
