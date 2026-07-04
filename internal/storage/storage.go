package storage

import (
	"time"

	"github.com/izu/izu-cli/internal/provider"
)

type Storage interface {
	AddHistory(entry *HistoryEntry) error
	GetHistory(limit, offset int) ([]*HistoryEntry, error)
	AddFavorite(anime *provider.Anime) error
	RemoveFavorite(id string) error
	GetFavorites() ([]*provider.Anime, error)
	Close() error
}

type HistoryEntry struct {
	ID         string        `json:"id"`
	AnimeID    string        `json:"anime_id"`
	AnimeTitle string        `json:"anime_title"`
	AnimeImage string        `json:"anime_image"`
	EpisodeID  string        `json:"episode_id"`
	EpisodeNum int           `json:"episode_number"`
	Position   time.Duration `json:"position"`
	Duration   time.Duration `json:"duration"`
	Provider   string        `json:"provider"`
	WatchedAt  time.Time     `json:"watched_at"`
}
