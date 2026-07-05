package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/izu/izu-cli/internal/provider"
)

type SearchModel struct {
	input     textinput.Model
	results   []provider.SearchResult
	cursor    int
	loading   bool
	err       error
	inputMode bool // true = accept text input, managed by app.go
}

func NewSearchModel() SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search anime..."
	ti.CharLimit = 256
	ti.Width = 50
	// Focus the textinput once and never touch focus state again.
	// We gate key routing via inputMode in app.go instead.
	ti.Focus()
	return SearchModel{
		input:     ti,
		inputMode: true,
	}
}

func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Always pass keys to the textinput — key routing is controlled
		// by inputMode in app.go, not by textinput's internal focus state.
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

	case searchResultsMsg:
		m.results = msg.results
		m.loading = false
		m.err = msg.err
	}

	return m, nil
}

// EnterInputMode activates text input. Called by app.go.
func (m *SearchModel) EnterInputMode() {
	m.inputMode = true
}

// ExitInputMode deactivates text input. Called by app.go.
func (m *SearchModel) ExitInputMode() {
	m.inputMode = false
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
