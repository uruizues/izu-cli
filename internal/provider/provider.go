package provider

import "context"

type Provider interface {
    Name() string
    Search(ctx context.Context, query string) ([]SearchResult, error)
    GetAnime(ctx context.Context, id string) (*Anime, error)
    GetEpisodes(ctx context.Context, animeID string, page int) (*EpisodePage, error)
    GetStream(ctx context.Context, episodeID string) (*StreamInfo, error)
    Close() error
}

type SearchResult struct {
    ID       string `json:"id"`
    Title    string `json:"title"`
    Image    string `json:"image"`
    Type     string `json:"type"`
    Episodes int    `json:"episodes"`
    Status   string `json:"status"`
}

type Anime struct {
    ID          string   `json:"id"`
    Title       string   `json:"title"`
    Japanese    string   `json:"japanese"`
    Synonyms    []string `json:"synonyms"`
    Description string   `json:"description"`
    Image       string   `json:"image"`
    Type        string   `json:"type"`
    Episodes    int      `json:"episodes"`
    Status      string   `json:"status"`
    Aired       string   `json:"aired"`
    Genres      []string `json:"genres"`
    Source      string   `json:"source"`
}

type EpisodePage struct {
    Episodes    []Episode `json:"episodes"`
    TotalPages  int       `json:"total_pages"`
    CurrentPage int       `json:"current_page"`
    HasNext     bool      `json:"has_next"`
}

type Episode struct {
    ID       string `json:"id"`
    Number   int    `json:"number"`
    Title    string `json:"title"`
    Duration string `json:"duration"`
    Snapshot string `json:"snapshot"`
}

type StreamInfo struct {
    URL       string            `json:"url"`
    Referer   string            `json:"referer"`
    Headers   map[string]string `json:"headers"`
    Subtitles []Subtitle        `json:"subtitles"`
    Quality   string            `json:"quality"`
    Format    string            `json:"format"`
}

type Subtitle struct {
    URL    string `json:"url"`
    Lang   string `json:"lang"`
    Label  string `json:"label"`
    Format string `json:"format"`
}

// StreamSource represents a single video stream with quality and playback metadata.
// Ported from anipy-cli's ProviderStream pattern.
type StreamSource struct {
    URL        string     `json:"url"`
    Resolution string     `json:"resolution"`
    Referrer   string     `json:"referrer"`
    Subtitles  []Subtitle `json:"subtitles,omitempty"`
    Language   string     `json:"language"` // "sub" or "dub"
    Bandwidth  int        `json:"bandwidth,omitempty"`
}
