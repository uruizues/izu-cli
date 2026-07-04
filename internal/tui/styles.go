package tui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true).
			Margin(1, 0)

	SearchStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#6C5CE7")).
			Padding(0, 1)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DFE6E9"))

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#636E72")).
			Margin(1, 0)

	ProviderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00CEC9")).
			Bold(true)
)
