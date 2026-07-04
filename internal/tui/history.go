package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/izu/izu-cli/internal/storage"
)

type HistoryModel struct {
	history []*storage.HistoryEntry
	cursor  int
	storage storage.Storage
	loading bool
	err     error
}

func NewHistoryModel(s storage.Storage) HistoryModel {
	return HistoryModel{
		storage: s,
	}
}

func (m HistoryModel) loadHistory() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.storage.GetHistory(50, 0)
		if err != nil {
			return historyErrMsg{err: err}
		}
		return historyListMsg{history: entries}
	}
}

type historyListMsg struct {
	history []*storage.HistoryEntry
}

type historyErrMsg struct {
	err error
}

func (m HistoryModel) Update(msg tea.Msg) (HistoryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case historyListMsg:
		m.history = msg.history
		m.loading = false
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.history)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m HistoryModel) View() string {
	s := TitleStyle.Render("History") + "\n\n"

	if m.loading {
		s += "Loading...\n"
		return s
	}

	if len(m.history) == 0 {
		s += StatusBarStyle.Render("No watch history yet.") + "\n"
		return s
	}

	for i, entry := range m.history {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		timeStr := entry.WatchedAt.Format(time.RFC822)
		line := fmt.Sprintf("%s %s - EP %02d (%s)", cursor, entry.AnimeTitle, entry.EpisodeNum, timeStr)
		if i == m.cursor {
			s += SelectedStyle.Render(line) + "\n"
		} else {
			s += NormalStyle.Render(line) + "\n"
		}
	}

	return s
}
