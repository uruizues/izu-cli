package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/izu/izu-cli/internal/provider"
	"github.com/izu/izu-cli/internal/storage"
)

type FavoritesModel struct {
	favorites []*provider.Anime
	cursor    int
	storage   storage.Storage
	loading   bool
	err       error
}

func NewFavoritesModel(s storage.Storage) FavoritesModel {
	return FavoritesModel{
		storage: s,
	}
}

func (m FavoritesModel) loadFavorites() tea.Cmd {
	return func() tea.Msg {
		favs, err := m.storage.GetFavorites()
		if err != nil {
			return favoritesErrMsg{err: err}
		}
		return favoritesListMsg{favorites: favs}
	}
}

type favoritesListMsg struct {
	favorites []*provider.Anime
}

type favoritesErrMsg struct {
	err error
}

func (m FavoritesModel) Update(msg tea.Msg) (FavoritesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case favoritesListMsg:
		m.favorites = msg.favorites
		m.loading = false
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.favorites)-1 {
				m.cursor++
			}
		case tea.KeyDelete, tea.KeyBackspace:
			if len(m.favorites) > 0 {
				fav := m.favorites[m.cursor]
				m.storage.RemoveFavorite(fav.ID)
				m.favorites = append(m.favorites[:m.cursor], m.favorites[m.cursor+1:]...)
				if m.cursor >= len(m.favorites) && m.cursor > 0 {
					m.cursor--
				}
			}
		}
	}
	return m, nil
}

func (m FavoritesModel) View() string {
	s := TitleStyle.Render("Favorites") + "\n\n"

	if m.loading {
		s += "Loading...\n"
		return s
	}

	if len(m.favorites) == 0 {
		s += StatusBarStyle.Render("No favorites yet. Press 'f' on an anime to add it.") + "\n"
		return s
	}

	for i, fav := range m.favorites {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
			s += SelectedStyle.Render(fmt.Sprintf("%s %s", cursor, fav.Title)) + "\n"
		} else {
			s += NormalStyle.Render(fmt.Sprintf("%s %s", cursor, fav.Title)) + "\n"
		}
	}

	return s
}
