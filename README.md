<p align="center">
  <img src="https://consumet.org/images/consumetlogo.png" width="120">
</p>

<h1 align="center">izu-cli</h1>

<p align="center">
  Terminal anime streaming client with TUI interface
</p>

<p align="center">
  <img src="https://img.shields.io/badge/go-1.22+-00ADD8?style=flat&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
  <img src="https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-blue" alt="Platform">
</p>

---

## Features

- **Search** anime across AniList with instant results
- **Stream** episodes directly via mpv player
- **Multi-provider fallback** — tries multiple sources if one fails
- **Subtitles** — automatic subtitle loading
- **Favorites & History** — saved locally in SQLite
- **Discord Rich Presence** — show what you're watching
- **TUI interface** — navigate entirely with keyboard

## Installation

### Linux

**Prerequisites:**
```bash
# Ubuntu/Debian
sudo apt install golang mpv python3 git

# Arch Linux
sudo pacman -S go mpv python git

# Fedora
sudo dnf install golang mpv python3 git
```

**Build from source:**
```bash
git clone https://github.com/izu/izu-cli.git
cd izu-cli
go build -o izu ./cmd/izu/
chmod +x izu

# Install to PATH
mkdir -p ~/.local/bin
cp izu ~/.local/bin/
cp izuapi ~/.local/bin/
```

Add to PATH (if not already):
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**Or use Makefile:**
```bash
make build
sudo make install
```

### macOS

**Prerequisites:**
```bash
# Using Homebrew
brew install go mpv python3 git
```

**Build from source:**
```bash
git clone https://github.com/izu/izu-cli.git
cd izu-cli
go build -o izu ./cmd/izu/
chmod +x izu

# Install to PATH
mkdir -p ~/.local/bin
cp izu ~/.local/bin/
cp izuapi ~/.local/bin/
```

Add to PATH:
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### Windows

**Prerequisites:**
1. Install [Go](https://go.dev/dl/) 1.22+
2. Install [mpv](https://sourceforge.net/projects/mpv-player-windows/) — add to PATH
3. Install [Python 3](https://www.python.org/downloads/) — add to PATH
4. Install [Git](https://git-scm.com/download/win)

**Build from source (PowerShell):**
```powershell
git clone https://github.com/izu/izu-cli.git
cd izu-cli
go build -o izu.exe ./cmd/izu/

# Create install directory
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.local\bin"

# Copy binaries
Copy-Item izu.exe "$env:USERPROFILE\.local\bin\"
Copy-Item izuapi.bat "$env:USERPROFILE\.local\bin\"

# Add to PATH (run once)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\.local\bin", "User")
```

## Setup Miruro API

izu-cli uses [Miruro API](https://github.com/walterwhite-69/Miruro-API) as its streaming backend. You need to run it locally.

```bash
# First-time setup
izuapi setup

# Start the API
izuapi start

# Check status
izuapi status

# Stop when done
izuapi stop
```

The API runs on `http://localhost:8000` by default.

### Miruro API Commands

| Command | Description |
|---------|-------------|
| `izuapi setup` | Clone repo and install Python dependencies |
| `izuapi start` | Start the API server |
| `izuapi stop` | Stop the API server |
| `izuapi restart` | Restart the API server |
| `izuapi status` | Check if API is running |
| `izuapi logs` | View API logs |
| `izuapi update` | Pull latest changes |

## Usage

```bash
# Launch TUI
izu

# Show version
izu --version
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `/` or `Ctrl+F` | Start search |
| `Enter` | Select / Play |
| `Esc` | Go back |
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Tab` | Switch provider |
| `f` | Favorites |
| `h` | History |
| `w` | Watch from URL |
| `q` / `Ctrl+C` | Quit |

## Configuration

Config file: `~/.config/izu-cli/config.yaml`

```yaml
general:
  provider: "miruro"
  theme: "dark"
  language: "en"

player:
  binary: "mpv"
  volume: 100

consumet:
  base_url: "http://localhost:8000"
  enabled: true
  provider: "miruro"
```

## Architecture

```
izu-cli/
├── cmd/izu/              # Entry point
├── internal/
│   ├── config/           # Viper config management
│   ├── provider/
│   │   ├── miruro/       # Miruro API client (primary)
│   │   ├── allanime/     # AllAnime GraphQL client
│   │   ├── anilist/      # AniList GraphQL client
│   │   ├── consumet/     # Consumet API client
│   │   └── jikan/        # Jikan (MyAnimeList) client
│   ├── player/mpv/       # mpv IPC client
│   ├── storage/          # SQLite history & favorites
│   └── tui/              # Bubbletea TUI
├── izuapi                # API server launcher script
├── run.sh                # Unified launcher
└── Makefile
```

## Troubleshooting

**"API not running" error:**
```bash
izuapi start
```

**mpv doesn't play / 403 errors:**
- Make sure Miruro API is running (`izuapi status`)
- Try restarting: `izuapi restart`

**No search results:**
- Check internet connection
- Verify API is running: `curl http://localhost:8000/`

**Build fails:**
- Ensure Go 1.22+ is installed: `go version`
- Try: `go clean && go build -o izu ./cmd/izu/`

## Related

- [Miruro API](https://github.com/walterwhite-69/Miruro-API) — Streaming backend
- [AniList](https://anilist.co) — Anime metadata
- [mpv](https://mpv.io) — Video player

## License

MIT
