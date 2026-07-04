package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/izu/izu-cli/internal/provider"
	"github.com/izu/izu-cli/internal/storage"
)

type screen int

const (
	screenSearch screen = iota
	screenEpisodes
	screenPlayer
	screenFavorites
	screenHistory
)

type Model struct {
	screen      screen
	search      SearchModel
	episodes    EpisodesModel
	favorites   FavoritesModel
	history     HistoryModel
	keys        KeyMap
	providers   []provider.Provider
	providerIdx int
	width       int
	height      int
	ready       bool
	err         error
}

func NewModel(providers []provider.Provider, s storage.Storage) Model {
	return Model{
		screen:    screenSearch,
		search:    NewSearchModel(),
		episodes:  NewEpisodesModel(),
		favorites: NewFavoritesModel(s),
		history:   NewHistoryModel(s),
		keys:      DefaultKeyMap(),
		providers: providers,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Tab):
			if len(m.providers) > 1 {
				m.providerIdx = (m.providerIdx + 1) % len(m.providers)
			}
			return m, nil
		case key.Matches(msg, m.keys.Escape):
			if m.screen == screenEpisodes {
				m.screen = screenSearch
				return m, nil
			}
			if m.screen == screenFavorites || m.screen == screenHistory {
				m.screen = screenSearch
				return m, nil
			}
		case key.Matches(msg, m.keys.Favorite):
			if m.screen == screenSearch {
				m.screen = screenFavorites
				cmds = append(cmds, m.favorites.loadFavorites())
			}
		case key.Matches(msg, m.keys.History):
			if m.screen == screenSearch {
				m.screen = screenHistory
				cmds = append(cmds, m.history.loadHistory())
			}
		}
	}

	switch m.screen {
	case screenSearch:
		newSearch, cmd := m.search.Update(msg)
		m.search = newSearch
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case screenEpisodes:
		newEpisodes, cmd := m.episodes.Update(msg)
		m.episodes = newEpisodes
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case screenFavorites:
		newFavs, cmd := m.favorites.Update(msg)
		m.favorites = newFavs
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case screenHistory:
		newHist, cmd := m.history.Update(msg)
		m.history = newHist
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var s string
	s += TitleStyle.Render("iz u - anime cli") + "\n"
	if len(m.providers) > 0 {
		s += ProviderStyle.Render("Provider: "+m.providers[m.providerIdx].Name()) + "\n"
	}
	s += "\n"

	switch m.screen {
	case screenSearch:
		s += m.search.View()
	case screenEpisodes:
		s += m.episodes.View()
	case screenFavorites:
		s += m.favorites.View()
	case screenHistory:
		s += m.history.View()
	}

	s += "\n" + StatusBarStyle.Render("[f]avorites [h]istory [tab]switch provider [q]uit")

	return s
}
