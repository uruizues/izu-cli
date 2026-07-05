#!/bin/bash
# izu-cli launcher script

export PATH="$HOME/.local/bin:$PATH"

case "${1:-}" in
    ""|start|play)
        izu
        ;;
    izuapi)
        shift
        ./izuapi "$@"
        ;;
    api|api-start)
        ./izuapi start
        ;;
    api-stop)
        ./izuapi stop
        ;;
    api-status)
        ./izuapi status
        ;;
    search|s)
        izu --search "${2:-}"
        ;;
    version|v)
        izu --version
        ;;
    help|h)
        echo "izu-cli - Anime streaming CLI"
        echo ""
        echo "Usage: ./run.sh [command] [args]"
        echo ""
        echo "Commands:"
        echo "  (no args)  - Launch TUI"
        echo "  start      - Launch TUI"
        echo "  api        - Start Consumet API server"
        echo "  api-stop   - Stop API server"
        echo "  api-status - Check API status"
        echo "  search     - Search anime"
        echo "  version    - Show version"
        echo "  help       - Show this help"
        echo "  install    - Install to ~/.local/bin"
        echo "  build      - Build binary"
        echo "  update     - Pull latest and rebuild"
        echo "  config     - Edit config file"
        ;;
    install)
        mkdir -p ~/.local/bin
        cp izu ~/.local/bin/
        cp izuapi ~/.local/bin/
        chmod +x ~/.local/bin/izuapi
        echo "Installed:"
        echo "  ~/.local/bin/izu"
        echo "  ~/.local/bin/izuapi"
        ;;
    build)
        go build -o izu ./cmd/izu/
        echo "Built: ./izu"
        ;;
    update)
        git pull
        go build -o izu ./cmd/izu/
        echo "Updated and built"
        ;;
    config)
        ${EDITOR:-nano} ~/.config/izu-cli/config.yaml
        ;;
    *)
        izu "$@"
        ;;
esac
