package tui

import (
	"time"

	"github.com/izu/izu-cli/internal/provider"
)

type searchResultsMsg struct {
	results []provider.SearchResult
	err     error
}

type searchErrMsg struct {
	err error
}

type episodeListMsg struct {
	episodes []provider.Episode
	err      error
}

type streamMsg struct {
	info *provider.StreamInfo
	err  error
}

type tickMsg time.Time
