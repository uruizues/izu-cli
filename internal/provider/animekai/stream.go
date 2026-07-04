package animekai

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/izu/izu-cli/internal/provider"
)

type streamResponse struct {
    Link struct {
        URL     string `json:"url"`
        Type    string `json:"type"`
        Quality string `json:"quality"`
    } `json:"link"`
    Subtitles []struct {
        File     string `json:"file"`
        Label    string `json:"label"`
        Language string `json:"language"`
    } `json:"subtitles"`
}

func (c *Client) GetStream(ctx context.Context, episodeID string) (*provider.StreamInfo, error) {
    data, err := c.doRequest("GET", fmt.Sprintf("/ajax/links/view?id=%s", episodeID), nil)
    if err != nil {
        return nil, err
    }

    var resp streamResponse
    if err := json.Unmarshal(data, &resp); err != nil {
        return nil, err
    }

    info := &provider.StreamInfo{
        URL:     resp.Link.URL,
        Referer: c.baseURL + "/",
        Headers: map[string]string{
            "Referer": c.baseURL + "/",
        },
        Quality: resp.Link.Quality,
        Format:  "hls",
    }

    for _, sub := range resp.Subtitles {
        info.Subtitles = append(info.Subtitles, provider.Subtitle{
            URL:    sub.File,
            Lang:   sub.Language,
            Label:  sub.Label,
            Format: "vtt",
        })
    }

    return info, nil
}

func (c *Client) Name() string                                        { return "animekai" }
func (c *Client) GetAnime(ctx context.Context, id string) (*provider.Anime, error) {
    return nil, nil
}
func (c *Client) GetEpisodes(ctx context.Context, animeID string, page int) (*provider.EpisodePage, error) {
    return nil, nil
}
func (c *Client) Close() error { return nil }
