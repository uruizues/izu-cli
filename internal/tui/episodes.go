package tui

import (
	"fmt"
	"strings"

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

	// Anime info header
	if m.anime != nil {
		s += TitleStyle.Render(m.anime.Title) + "\n"

		// Info line: type • status • episodes
		var info []string
		if m.anime.Type != "" {
			info = append(info, InfoLabel.Render(m.anime.Type))
		}
		if m.anime.Status != "" {
			info = append(info, m.anime.Status)
		}
		if m.anime.Episodes > 0 {
			info = append(info, fmt.Sprintf("%d episodes", m.anime.Episodes))
		}
		if len(info) > 0 {
			s += SubtitleStyle.Render(strings.Join(info, " • ")) + "\n"
		}

		// Genres
		if len(m.anime.Genres) > 0 {
			s += SubtitleStyle.Render(strings.Join(m.anime.Genres, " · ")) + "\n"
		}

		s += SeparatorStyle.Render(strings.Repeat("─", 50)) + "\n\n"
	}

	if m.loading {
		s += LoadingStyle.Render("Loading episodes...") + "\n"
		return s
	}

	if m.err != nil {
		s += ErrorStyle.Render("Error: " + m.err.Error()) + "\n"
		return s
	}

	// Episode list
	if len(m.episodes) == 0 {
		s += StatusBarStyle.Render("No episodes found.") + "\n"
		return s
	}

	// Show pagination info
	total := len(m.episodes)
	showing := total
	if showing > 25 {
		showing = 25
	}
	s += SubtitleStyle.Render(fmt.Sprintf("Episodes 1-%d of %d", showing, total)) + "\n\n"

	for i, ep := range m.episodes {
		if i >= 25 {
			break
		}

		epNum := fmt.Sprintf("EP %02d", ep.Number)

		if i == m.cursor {
			// Selected episode — highlighted
			s += SelectedStyle.Render(fmt.Sprintf(" %s ▶ %s", epNum, ep.Title)) + "\n"
		} else {
			// Normal episode
			s += NormalStyle.Render(fmt.Sprintf(" %s   %s", EpisodeBadge.Render(epNum), ep.Title)) + "\n"
		}
	}

	return s
}
