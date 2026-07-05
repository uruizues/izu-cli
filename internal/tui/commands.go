package tui

import (
	"context"
	"encoding/json"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/izu/izu-cli/internal/provider"
)

func (m Model) searchCmd(query string, p provider.Provider) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		results, err := p.Search(ctx, query)
		return searchResultsMsg{results: results, err: err}
	}
}

func (m Model) loadEpisodesCmd(id string, p provider.Provider) tea.Cmd {
	return func() tea.Msg {
		page, err := p.GetEpisodes(context.Background(), id, 1)
		if err != nil {
			return episodeListMsg{err: err}
		}
		return episodeListMsg{episodes: page.Episodes, err: nil}
	}
}

func (m Model) loadStreamCmd(episodeID string, p provider.Provider) tea.Cmd {
	return func() tea.Msg {
		info, err := p.GetStream(context.Background(), episodeID)
		return streamMsg{info: info, err: err}
	}
}

type ytDlpInfo struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
	Duration int    `json:"duration"`
}

func (m Model) watchURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("yt-dlp",
			"--dump-json",
			"--no-download",
			"--no-warnings",
			url,
		)

		output, err := cmd.Output()
		if err != nil {
			return streamMsg{err: err}
		}

		var info ytDlpInfo
		if err := json.Unmarshal(output, &info); err != nil {
			return streamMsg{err: err}
		}

		streamInfo := &provider.StreamInfo{
			URL:     info.URL,
			Quality: "best",
			Format:  "hls",
		}

		return streamMsg{info: streamInfo, err: nil}
	}
}
