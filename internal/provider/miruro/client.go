package miruro

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/izu/izu-cli/internal/provider"
)

const DefaultBaseURL = "http://localhost:8000"

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Name() string {
	return "miruro"
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) doRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "izu-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

type jikanSearchResponse struct {
	Data []struct {
		MALID   int `json:"mal_id"`
		Title   string `json:"title"`
		TitleEnglish *string `json:"title_english"`
		Images struct {
			JPG struct {
				ImageURL string `json:"image_url"`
			} `json:"jpg"`
		} `json:"images"`
		Type     string `json:"type"`
		Episodes *int   `json:"episodes"`
		Status   string `json:"status"`
	} `json:"data"`
}

func (c *Client) Search(ctx context.Context, query string) ([]provider.SearchResult, error) {
	jikanURL := fmt.Sprintf("https://api.jikan.moe/v4/anime?q=%s&limit=20", url.QueryEscape(query))
	data, err := c.doRequest(jikanURL)
	if err != nil {
		return nil, err
	}

	var resp jikanSearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	results := make([]provider.SearchResult, 0, len(resp.Data))
	for _, r := range resp.Data {
		title := r.Title
		if r.TitleEnglish != nil && *r.TitleEnglish != "" {
			title = *r.TitleEnglish
		}

		image := r.Images.JPG.ImageURL

		episodes := 0
		if r.Episodes != nil {
			episodes = *r.Episodes
		}

		results = append(results, provider.SearchResult{
			ID:       fmt.Sprintf("%d", r.MALID),
			Title:    title,
			Image:    image,
			Type:     r.Type,
			Episodes: episodes,
			Status:   r.Status,
		})
	}

	return results, nil
}

type infoResponse struct {
	ID          int `json:"id"`
	Title       struct {
		Romaji  string `json:"romaji"`
		English string `json:"english"`
		Native  string `json:"native"`
	} `json:"title"`
	Description string `json:"description"`
	CoverImage  struct {
		Large string `json:"large"`
	} `json:"coverImage"`
	Format   string `json:"format"`
	Episodes int    `json:"episodes"`
	Status   string `json:"status"`
	Genres   []string `json:"genres"`
}

func (c *Client) GetAnime(ctx context.Context, id string) (*provider.Anime, error) {
	url := fmt.Sprintf("%s/info/%s", c.baseURL, id)
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var resp infoResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	title := resp.Title.English
	if title == "" {
		title = resp.Title.Romaji
	}

	return &provider.Anime{
		ID:          fmt.Sprintf("%d", resp.ID),
		Title:       title,
		Japanese:    resp.Title.Native,
		Description: resp.Description,
		Image:       resp.CoverImage.Large,
		Type:        resp.Format,
		Episodes:    resp.Episodes,
		Status:      resp.Status,
		Genres:      resp.Genres,
	}, nil
}

type episodesResponse struct {
	Providers map[string]struct {
		Episodes map[string][]struct {
			ID     string `json:"id"`
			Number int    `json:"number"`
			Title  string `json:"title"`
		} `json:"episodes"`
	} `json:"providers"`
}

func (c *Client) GetEpisodes(ctx context.Context, animeID string, page int) (*provider.EpisodePage, error) {
	// Try Miruro first
	url := fmt.Sprintf("%s/episodes/%s", c.baseURL, animeID)
	data, err := c.doRequest(url)
	if err == nil {
		var resp episodesResponse
		if json.Unmarshal(data, &resp) == nil {
			var episodes []provider.Episode
			for _, provData := range resp.Providers {
				for _, epList := range provData.Episodes {
					for _, ep := range epList {
						episodes = append(episodes, provider.Episode{
							ID:     ep.ID,
							Number: ep.Number,
							Title:  ep.Title,
						})
					}
					break
				}
				break
			}

			sort.Slice(episodes, func(i, j int) bool {
				return episodes[i].Number < episodes[j].Number
			})

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
	}

	// Fallback: generate episodes from Jikan metadata
	jikanURL := fmt.Sprintf("https://api.jikan.moe/v4/anime/%s", animeID)
	jikanData, err := c.doRequest(jikanURL)
	if err != nil {
		return nil, fmt.Errorf("anime not available for streaming")
	}

	var jikanResp struct {
		Data struct {
			Title    string `json:"title"`
			Episodes *int   `json:"episodes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(jikanData, &jikanResp); err != nil {
		return nil, fmt.Errorf("anime not available for streaming")
	}

	total := 1
	if jikanResp.Data.Episodes != nil && *jikanResp.Data.Episodes > 0 {
		total = *jikanResp.Data.Episodes
	}

	var episodes []provider.Episode
	for i := 1; i <= total; i++ {
		episodes = append(episodes, provider.Episode{
			ID:     fmt.Sprintf("%s_ep%d", animeID, i),
			Number: i,
			Title:  fmt.Sprintf("Episode %d", i),
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

type streamResponse struct {
	Streams []struct {
		URL      string `json:"url"`
		Type     string `json:"type"`
		Quality  string `json:"quality"`
		Referer  string `json:"referer"`
	} `json:"streams"`
	Subtitles []struct {
		File  string `json:"file"`
		Label string `json:"label"`
	} `json:"subtitles"`
	Intro *struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"intro"`
	Outro *struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"outro"`
}

func (c *Client) GetStream(ctx context.Context, episodeID string) (*provider.StreamInfo, error) {
	// episodeID is a Miruro path like "watch/kiwi/20/sub/animepahe-1"
	return c.fetchStream(episodeID)
}

func (c *Client) fetchStream(path string) (*provider.StreamInfo, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, path)
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var resp streamResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Streams) == 0 {
		return nil, fmt.Errorf("no streams")
	}

	info := &provider.StreamInfo{
		Headers: map[string]string{},
	}

	for _, s := range resp.Streams {
		if s.Type == "hls" {
			info.URL = s.URL
			info.Quality = s.Quality
			info.Format = "hls"
			if s.Referer != "" {
				info.Referer = s.Referer
				info.Headers["Referer"] = s.Referer
				info.Headers["Origin"] = s.Referer
			}
			break
		}
	}

	if info.URL == "" && len(resp.Streams) > 0 {
		s := resp.Streams[0]
		info.URL = s.URL
		info.Quality = s.Quality
		info.Format = "mp4"
		if s.Referer != "" {
			info.Referer = s.Referer
			info.Headers["Referer"] = s.Referer
			info.Headers["Origin"] = s.Referer
		}
	}

	for _, sub := range resp.Subtitles {
		info.Subtitles = append(info.Subtitles, provider.Subtitle{
			URL:    sub.File,
			Label:  sub.Label,
			Lang:   sub.Label,
			Format: "vtt",
		})
	}

	return info, nil
}
