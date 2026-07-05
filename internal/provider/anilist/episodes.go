package anilist

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/izu/izu-cli/internal/provider"
)

const animeQuery = `query($id: Int) {
  Media(id: $id, type: ANIME) {
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
    description
    genres
    streamingEpisodes {
      title
      thumbnail
      url
      site
    }
  }
}`

type animeResponse struct {
	Data struct {
		Media struct {
			ID       int `json:"id"`
			Title    struct {
				Romaji  string `json:"romaji"`
				English string `json:"english"`
				Native  string `json:"native"`
			} `json:"title"`
			CoverImage struct {
				Large string `json:"large"`
			} `json:"coverImage"`
			Format  string   `json:"format"`
			Episodes int     `json:"episodes"`
			Status   string   `json:"status"`
			Description string `json:"description"`
			Genres      []string `json:"genres"`
			StreamingEpisodes []struct {
				Title     string `json:"title"`
				Thumbnail string `json:"thumbnail"`
				URL       string `json:"url"`
				Site      string `json:"site"`
			} `json:"streamingEpisodes"`
		} `json:"Media"`
	} `json:"data"`
}

// Store streaming episodes globally for access in GetStream
var (
	streamingEpisodesMu sync.RWMutex
	streamingEpisodesMap = make(map[string][]struct {
		Title string
		URL   string
		Site  string
	})
)

func (c *Client) GetAnime(ctx context.Context, id string) (*provider.Anime, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid anime ID: %s", id)
	}

	variables := map[string]interface{}{
		"id": idInt,
	}

	data, err := c.doQuery(animeQuery, variables)
	if err != nil {
		return nil, err
	}

	var resp animeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	media := resp.Data.Media

	// Store streaming episodes for later use
	var eps []struct {
		Title string
		URL   string
		Site  string
	}
	for _, se := range media.StreamingEpisodes {
		eps = append(eps, struct {
			Title string
			URL   string
			Site  string
		}{
			Title: se.Title,
			URL:   se.URL,
			Site:  se.Site,
		})
	}
	streamingEpisodesMu.Lock()
	streamingEpisodesMap[id] = eps
	streamingEpisodesMu.Unlock()

	return &provider.Anime{
		ID:          fmt.Sprintf("%d", media.ID),
		Title:       media.Title.English,
		Japanese:    media.Title.Native,
		Description: media.Description,
		Image:       media.CoverImage.Large,
		Type:        media.Format,
		Episodes:    media.Episodes,
		Status:      media.Status,
		Genres:      media.Genres,
	}, nil
}

func (c *Client) GetEpisodes(ctx context.Context, animeID string, page int) (*provider.EpisodePage, error) {
	// First ensure we have the anime data with streaming episodes
	streamingEpisodesMu.RLock()
	_, exists := streamingEpisodesMap[animeID]
	streamingEpisodesMu.RUnlock()
	if !exists {
		_, err := c.GetAnime(ctx, animeID)
		if err != nil {
			return nil, err
		}
	}

	streamingEpisodesMu.RLock()
	streamEps := streamingEpisodesMap[animeID]
	streamingEpisodesMu.RUnlock()

	var episodes []provider.Episode
	for i, se := range streamEps {
		episodes = append(episodes, provider.Episode{
			ID:     fmt.Sprintf("%s:%d", animeID, i),
			Number: i + 1,
			Title:  se.Title,
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
