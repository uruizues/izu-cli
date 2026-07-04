package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Escape     key.Binding
	Search     key.Binding
	Tab        key.Binding
	Quit       key.Binding
	Download   key.Binding
	Favorite   key.Binding
	History    key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Escape:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Search:   key.NewBinding(key.WithKeys("/", "ctrl+f"), key.WithHelp("/", "search")),
		Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch provider")),
		Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Download: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "download")),
		Favorite: key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "favorite")),
		History:  key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "history")),
	}
}
