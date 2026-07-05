package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/izu/izu-cli/internal/provider"
)

type EpisodesModel struct {
	anime    *provider.Anime
	episodes []provider.Episode
	cursor   int
	page     int
	loading  bool
	err      error
}

func NewEpisodesModel() EpisodesModel {
	return EpisodesModel{}
}

func (m EpisodesModel) Update(msg tea.Msg) (EpisodesModel, tea.Cmd) {
	switch msg.(type) {
	case episodeListMsg:
		// Handled directly by app.go now
	}
	return m, nil
}

func (m EpisodesModel) View() string {
	if m.anime == nil && !m.loading && m.err == nil && len(m.episodes) == 0 {
		return ""
	}

	s := ""
	if m.anime != nil {
		s += TitleStyle.Render(m.anime.Title) + "\n"
		s += StatusBarStyle.Render(fmt.Sprintf("%s • %d episodes", m.anime.Type, m.anime.Episodes)) + "\n\n"
	}

	if m.loading {
		s += "Loading episodes...\n"
		return s
	}

	if m.err != nil {
		s += "Error: " + m.err.Error() + "\n"
		return s
	}

	for i, ep := range m.episodes {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
			s += SelectedStyle.Render(fmt.Sprintf("%s EP %02d - %s", cursor, ep.Number, ep.Title)) + "\n"
		} else {
			s += NormalStyle.Render(fmt.Sprintf("%s EP %02d - %s", cursor, ep.Number, ep.Title)) + "\n"
		}
	}

	if len(m.episodes) == 0 {
		s += StatusBarStyle.Render("No episodes found.") + "\n"
	}

	return s
}
