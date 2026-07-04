package storage

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/izu/izu-cli/internal/provider"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage() (*SQLiteStorage, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	dbDir := filepath.Join(configDir, "izu-cli")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dbDir, "izu-cli.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	s := &SQLiteStorage{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

func (s *SQLiteStorage) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS history (
			id TEXT PRIMARY KEY,
			anime_id TEXT NOT NULL,
			anime_title TEXT NOT NULL,
			anime_image TEXT DEFAULT '',
			episode_id TEXT NOT NULL,
			episode_number INTEGER NOT NULL,
			position INTEGER DEFAULT 0,
			duration INTEGER DEFAULT 0,
			provider TEXT NOT NULL,
			watched_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS favorites (
			id TEXT PRIMARY KEY,
			data TEXT NOT NULL
		)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

func (s *SQLiteStorage) AddHistory(entry *HistoryEntry) error {
	if entry.ID == "" {
		entry.ID = entry.AnimeID + ":" + entry.EpisodeID
	}
	if entry.WatchedAt.IsZero() {
		entry.WatchedAt = time.Now()
	}

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO history (id, anime_id, anime_title, anime_image, episode_id, episode_number, position, duration, provider, watched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.AnimeID, entry.AnimeTitle, entry.AnimeImage,
		entry.EpisodeID, entry.EpisodeNum, int64(entry.Position), int64(entry.Duration),
		entry.Provider, entry.WatchedAt,
	)
	return err
}

func (s *SQLiteStorage) GetHistory(limit, offset int) ([]*HistoryEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, anime_id, anime_title, anime_image, episode_id, episode_number, position, duration, provider, watched_at
		FROM history
		ORDER BY watched_at DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*HistoryEntry
	for rows.Next() {
		e := &HistoryEntry{}
		var pos, dur int64
		if err := rows.Scan(&e.ID, &e.AnimeID, &e.AnimeTitle, &e.AnimeImage,
			&e.EpisodeID, &e.EpisodeNum, &pos, &dur, &e.Provider, &e.WatchedAt); err != nil {
			return nil, err
		}
		e.Position = time.Duration(pos)
		e.Duration = time.Duration(dur)
		entries = append(entries, e)
	}

	return entries, nil
}

func (s *SQLiteStorage) AddFavorite(anime *provider.Anime) error {
	data, err := json.Marshal(anime)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`INSERT OR REPLACE INTO favorites (id, data) VALUES (?, ?)`, anime.ID, string(data))
	return err
}

func (s *SQLiteStorage) RemoveFavorite(id string) error {
	_, err := s.db.Exec(`DELETE FROM favorites WHERE id = ?`, id)
	return err
}

func (s *SQLiteStorage) GetFavorites() ([]*provider.Anime, error) {
	rows, err := s.db.Query(`SELECT data FROM favorites`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []*provider.Anime
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		a := &provider.Anime{}
		if err := json.Unmarshal([]byte(data), a); err != nil {
			return nil, err
		}
		favorites = append(favorites, a)
	}

	return favorites, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
