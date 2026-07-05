package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/izu/izu-cli/internal/player"
	"github.com/izu/izu-cli/internal/player/mpv"
	"github.com/izu/izu-cli/internal/provider"
	"github.com/izu/izu-cli/internal/rpc"
	"github.com/izu/izu-cli/internal/storage"
)

type screen int

const (
	screenSearch screen = iota
	screenEpisodes
	screenPlayer
	screenFavorites
	screenHistory
	screenWatchURL
)

type Model struct {
	screen      screen
	search      SearchModel
	episodes    EpisodesModel
	favorites   FavoritesModel
	history     HistoryModel
	watchURL    textinput.Model
	keys        KeyMap
	providers   []provider.Provider
	providerIdx int
	player      player.Player
	playerCmd   context.CancelFunc
	discord     *rpc.DiscordRPC
	width       int
	height      int
	ready       bool
	err         error
}

func NewModel(providers []provider.Provider, s storage.Storage, discord *rpc.DiscordRPC) Model {
	urlInput := textinput.New()
	urlInput.Placeholder = "Paste anime URL (animepahe, zoro, etc.)"
	urlInput.CharLimit = 512
	urlInput.Width = 60

	m := Model{
		screen:    screenSearch,
		search:    NewSearchModel(),
		episodes:  NewEpisodesModel(),
		favorites: NewFavoritesModel(s),
		history:   NewHistoryModel(s),
		watchURL:  urlInput,
		keys:      DefaultKeyMap(),
		providers: providers,
		player:    mpv.New("mpv", []string{}, "/tmp/izu-mpv-socket"),
		discord:   discord,
	}
	m.search.EnterInputMode()
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
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
		// === WATCH URL MODE ===
		if m.screen == screenWatchURL {
			switch msg.Type {
			case tea.KeyEsc:
				m.screen = screenSearch
				m.search.EnterInputMode()
				return m, textinput.Blink
			case tea.KeyEnter:
				url := m.watchURL.Value()
				if url != "" {
					m.screen = screenSearch
					m.search.EnterInputMode()
					return m, m.watchURLCmd(url)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.watchURL, cmd = m.watchURL.Update(msg)
				return m, cmd
			}
		}

		// === INPUT MODE: ALL keys go to search, NO global bindings ===
		if m.screen == screenSearch && m.search.inputMode {
			switch msg.Type {
			case tea.KeyEsc:
				m.search.ExitInputMode()
				return m, nil
			case tea.KeyEnter:
				// Submit search query — exit input mode, trigger search
				query := m.search.input.Value()
				if query != "" {
					m.search.ExitInputMode()
					m.search.loading = true
					provider := m.providers[m.providerIdx]
					cmds = append(cmds, m.searchCmd(query, provider))
					return m, tea.Batch(cmds...)
				}
				return m, nil
			default:
				newSearch, cmd := m.search.Update(msg)
				m.search = newSearch
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return m, tea.Batch(cmds...)
			}
		}

		// === GLOBAL KEYBINDINGS (only when NOT in input mode) ===
		switch {
		case key.Matches(msg, m.keys.Quit):
			if m.playerCmd != nil {
				m.playerCmd()
				m.playerCmd = nil
			}
			if m.player != nil && m.player.IsRunning() {
				m.player.Stop()
			}
			return m, tea.Quit

		case key.Matches(msg, m.keys.Tab):
			if len(m.providers) > 1 {
				m.providerIdx = (m.providerIdx + 1) % len(m.providers)
			}
			return m, nil

		case key.Matches(msg, m.keys.Escape):
			switch m.screen {
			case screenEpisodes:
				m.screen = screenSearch
				m.search.EnterInputMode()
				return m, textinput.Blink
			case screenFavorites, screenHistory:
				m.screen = screenSearch
				return m, nil
			}

		case key.Matches(msg, m.keys.Search):
			if m.screen == screenSearch {
				m.search.EnterInputMode()
				return m, textinput.Blink
			}

		case key.Matches(msg, m.keys.Enter):
			if m.screen == screenEpisodes && len(m.episodes.episodes) > 0 {
				ep := m.episodes.episodes[m.episodes.cursor]
				provider := m.providers[m.providerIdx]
				cmds = append(cmds, m.loadStreamCmd(ep.ID, provider))
			}
			if m.screen == screenSearch && len(m.search.results) > 0 {
				result := m.search.results[m.search.cursor]
				provider := m.providers[m.providerIdx]
				cmds = append(cmds, m.loadEpisodesCmd(result.ID, provider))
			}

		case key.Matches(msg, m.keys.Up):
			if m.screen == screenSearch && !m.search.inputMode {
				if m.search.cursor > 0 {
					m.search.cursor--
				}
			}
			if m.screen == screenEpisodes {
				if m.episodes.cursor > 0 {
					m.episodes.cursor--
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Down):
			if m.screen == screenSearch && !m.search.inputMode {
				if m.search.cursor < len(m.search.results)-1 {
					m.search.cursor++
				}
			}
			if m.screen == screenEpisodes {
				if m.episodes.cursor < len(m.episodes.episodes)-1 {
					m.episodes.cursor++
				}
			}
			return m, nil

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

		case key.Matches(msg, m.keys.WatchURL):
			if m.screen == screenSearch {
				m.screen = screenWatchURL
				m.watchURL.SetValue("")
				m.watchURL.Focus()
				return m, textinput.Blink
			}
		}

	case streamMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if m.playerCmd != nil {
			m.playerCmd()
		}
		ctx, cancel := context.WithCancel(context.Background())
		m.playerCmd = cancel
		go func() {
			m.player.Play(ctx, msg.info, player.PlayOptions{})
		}()
		// Set Discord RPC activity
		if m.discord != nil && m.episodes.anime != nil {
			epNum := 0
			if len(m.episodes.episodes) > 0 && m.episodes.cursor < len(m.episodes.episodes) {
				epNum = m.episodes.episodes[m.episodes.cursor].Number
			}
			m.discord.SetWatching(m.episodes.anime.Title, epNum)
		}
		return m, nil

	case episodeListMsg:
		m.episodes.episodes = msg.episodes
		m.episodes.loading = false
		m.episodes.err = msg.err
		m.screen = screenEpisodes

	case searchResultsMsg:
		m.search.results = msg.results
		m.search.loading = false
		m.search.err = msg.err
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return LoadingStyle.Render("Initializing...")
	}

	var s string

	// Header
	s += TitleStyle.Render("iz u") + SubtitleStyle.Render(" — anime cli") + "\n"
	if len(m.providers) > 0 {
		s += ProviderStyle.Render("● " + m.providers[m.providerIdx].Name()) + "\n"
	}
	s += SeparatorStyle.Render(strings.Repeat("─", 50)) + "\n"

	// Screen content
	switch m.screen {
	case screenSearch:
		s += m.search.View()
	case screenEpisodes:
		s += m.episodes.View()
	case screenFavorites:
		s += m.favorites.View()
	case screenHistory:
		s += m.history.View()
	case screenWatchURL:
		s += TitleStyle.Render("Watch from URL") + "\n\n"
		s += m.watchURL.View() + "\n"
		s += SubtitleStyle.Render("Paste URL from anime site and press Enter") + "\n"
	}

	// Status bar
	s += SeparatorStyle.Render(strings.Repeat("─", 50)) + "\n"
	if m.screen == screenSearch && m.search.inputMode {
		s += StatusBarStyle.Render("[esc] exit search   [enter] search")
	} else if m.screen == screenEpisodes {
		s += StatusBarStyle.Render("[↑↓] navigate   [enter] play   [esc] back")
	} else if m.screen == screenWatchURL {
		s += StatusBarStyle.Render("[enter] play   [esc] cancel")
	} else {
		s += StatusBarStyle.Render("[/] search   [w] watch URL   [f] favorites   [h] history   [tab] provider   [q] quit")
	}

	// Error display
	if m.err != nil {
		s += "\n" + ErrorStyle.Render("Error: "+m.err.Error())
	}

	return s
}
