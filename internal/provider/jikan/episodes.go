package jikan

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/izu/izu-cli/internal/provider"
)

type animeResponse struct {
	Data struct {
		MalID         int    `json:"mal_id"`
		Title         string `json:"title"`
		TitleJapanese string `json:"title_japanese"`
		Images        map[string]struct {
			LargeImageURL string `json:"large_image_url"`
		} `json:"images"`
		Type          string   `json:"type"`
		Episodes      int      `json:"episodes"`
		Status        string   `json:"status"`
		Aired         struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"aired"`
		Genres []struct {
			Name string `json:"name"`
		} `json:"genres"`
		Synopsis string `json:"synopsis"`
	} `json:"data"`
}

type episodesResponse struct {
	Data []struct {
		MalID    int    `json:"mal_id"`
		URL      string `json:"url"`
		Title    string `json:"title"`
		Episode  int    `json:"episode"`
	} `json:"data"`
	Pagination struct {
		LastVisiblePage int `json:"last_visible_page"`
		HasNextPage     bool `json:"has_next_page"`
	} `json:"pagination"`
}

func (c *Client) GetAnime(ctx context.Context, id string) (*provider.Anime, error) {
	u := fmt.Sprintf("%s/anime/%s", BaseURL, id)

	data, err := c.doRequest(u)
	if err != nil {
		return nil, err
	}

	var resp animeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	anime := resp.Data

	genres := make([]string, len(anime.Genres))
	for i, g := range anime.Genres {
		genres[i] = g.Name
	}

	image := ""
	if img, ok := anime.Images["jpg"]; ok {
		image = img.LargeImageURL
	}

	return &provider.Anime{
		ID:          strconv.Itoa(anime.MalID),
		Title:       anime.Title,
		Japanese:    anime.TitleJapanese,
		Description: anime.Synopsis,
		Image:       image,
		Type:        anime.Type,
		Episodes:    anime.Episodes,
		Status:      anime.Status,
		Genres:      genres,
	}, nil
}

func (c *Client) GetEpisodes(ctx context.Context, animeID string, page int) (*provider.EpisodePage, error) {
	u := fmt.Sprintf("%s/anime/%s/episodes?page=%d", BaseURL, animeID, page)

	data, err := c.doRequest(u)
	if err != nil {
		return nil, err
	}

	var resp episodesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var episodes []provider.Episode
	for _, ep := range resp.Data {
		episodes = append(episodes, provider.Episode{
			ID:     fmt.Sprintf("%d", ep.MalID),
			Number: ep.Episode,
			Title:  ep.Title,
		})
	}

	return &provider.EpisodePage{
		Episodes:    episodes,
		TotalPages:  resp.Pagination.LastVisiblePage,
		CurrentPage: page,
		HasNext:     resp.Pagination.HasNextPage,
	}, nil
}
