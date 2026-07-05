package tui

import (
	"fmt"
	"strings"

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
	s := SearchStyle.Render(m.input.View()) + "\n"

	if m.loading {
		s += LoadingStyle.Render("Searching...") + "\n"
		return s
	}

	if m.err != nil {
		s += ErrorStyle.Render("Error: " + m.err.Error()) + "\n"
		return s
	}

	if len(m.results) == 0 && !m.inputMode {
		s += StatusBarStyle.Render("No results found.") + "\n"
		return s
	}

	if len(m.results) == 0 {
		s += StatusBarStyle.Render("Type to search...") + "\n"
		return s
	}

	// Results header
	s += SubtitleStyle.Render(fmt.Sprintf("%d results", len(m.results))) + "\n\n"

	// Results list
	for i, result := range m.results {
		if i >= 20 {
			break
		}

		// Build info parts
		var parts []string
		if result.Type != "" {
			parts = append(parts, InfoLabel.Render(result.Type))
		}
		if result.Episodes > 0 {
			parts = append(parts, fmt.Sprintf("%d eps", result.Episodes))
		}
		if result.Status != "" {
			parts = append(parts, result.Status)
		}
		infoStr := ""
		if len(parts) > 0 {
			infoStr = " " + SubtitleStyle.Render("("+strings.Join(parts, " · ")+")")
		}

		if i == m.cursor {
			s += SelectedStyle.Render(fmt.Sprintf(" ▶ %s", result.Title)) + infoStr + "\n"
		} else {
			s += NormalStyle.Render(fmt.Sprintf("   %s", result.Title)) + infoStr + "\n"
		}
	}

	return s
}
