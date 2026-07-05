package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Title style — bold red header
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true).
			Margin(0, 0, 0, 0)

	// Subtitle style — dimmed text
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#636E72")).
			Italic(true)

	// Search input box
	SearchStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#6C5CE7")).
			Padding(0, 1).
			Margin(0, 0, 1, 0)

	// Selected item — bright with background
	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#6C5CE7")).
			Bold(true).
			Padding(0, 1)

	// Normal item
	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DFE6E9"))

	// Episode number badge
	EpisodeBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00CEC9")).
			Bold(true).
			Width(6)

	// Episode title
	EpisodeTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DFE6E9"))

	// Status bar at bottom
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#636E72")).
			Margin(1, 0)

	// Provider name
	ProviderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00CEC9")).
			Bold(true)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	// Loading spinner text
	LoadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FDCB6E")).
			Italic(true)

	// Separator line
	SeparatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2D3436"))

	// Info label (genres, status, etc.)
	InfoLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00CEC9")).
			Bold(true)

	// Info value
	InfoValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DFE6E9"))

	// Box container
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6C5CE7")).
			Padding(1, 2)
)
