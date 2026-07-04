package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/izu/izu-cli/internal/provider"
)

type SearchModel struct {
	input   textinput.Model
	results []provider.SearchResult
	cursor  int
	loading bool
	err     error
}

func NewSearchModel() SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search anime..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	return SearchModel{
		input: ti,
	}
}

func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if len(m.results) > 0 {
				m.cursor = 0
				return m, nil
			}
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
		}
	case searchResultsMsg:
		m.results = msg.results
		m.loading = false
		m.err = msg.err
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m SearchModel) View() string {
	s := SearchStyle.Render(m.input.View()) + "\n\n"

	if m.loading {
		s += "Searching...\n"
		return s
	}

	if m.err != nil {
		s += "Error: " + m.err.Error() + "\n"
		return s
	}

	for i, result := range m.results {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
			s += SelectedStyle.Render(cursor+" "+result.Title) + "\n"
		} else {
			s += NormalStyle.Render(cursor+" "+result.Title) + "\n"
		}
	}

	if len(m.results) == 0 {
		s += StatusBarStyle.Render("Type to search...") + "\n"
	}

	return s
}
