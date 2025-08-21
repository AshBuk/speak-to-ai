# Development Guide

Concise commands for local development and CI-compatible builds.

## Build Architecture

```
Makefile → bash-scripts/ → docker/ → CI/CD
   ↑           ↑            ↑        ↑
simple     orchestration containers production
command    dependencies
```

### 1. Makefile - Simple Commands
Entry point for developers with proper CGO environment setup:
```bash
make build          # Build binary with whisper.cpp integration
make build-systray  # Build with system tray support
make test           # Run tests with CGO dependencies
make deps           # Download Go dependencies
make whisper-libs   # Build whisper.cpp libraries into ./lib
make appimage       # Build AppImage package
make flatpak        # Build Flatpak package
make clean          # Clean build artifacts
make fmt            # Format Go code
make lint           # Run linter (via Docker)
```

### 2. Bash Scripts - Orchestration & Dependencies  
Handle complex build logic and dependency management:
```bash
bash-scripts/build-appimage.sh     # AppImage creation with linuxdeploy fallbacks
bash-scripts/build-flatpak.sh      # Flatpak validation (CI builds actual package)
bash-scripts/dev-env.sh           # CGO environment configuration
bash-scripts/flatpak-runtime.sh   # Flatpak runtime wrapper
```

### 3. Docker - Containers
Reproducible builds across different environments:
```bash
make docker-build      # Build all Docker images
make docker-dev        # Enter development container
make docker-appimage   # Build AppImage in container
make docker-flatpak    # Build Flatpak in container
make docker-lint       # Run linter in container
make docker-clean      # Clean Docker resources
```

### 4. CI/CD - Production
GitHub Actions handle complex builds, releases, and distribution.

## Quick Start

```bash
# One-time setup
sudo apt-get update && sudo apt-get install -y build-essential cmake git pkg-config
make deps whisper-libs

# Development session
source bash-scripts/dev-env.sh
make build test
```

### Example Configuration

```yaml
# General settings
general:
  debug: false
  model_path: "~/.config/speak-to-ai/language-models/base.bin"
  language: "auto"  # Auto-detect or specify "en", "ru", etc.


# Audio settings
audio:
  device: "default"
  sample_rate: 16000
  recording_method: "arecord"  # Options: "arecord", "ffmpeg"

# Output settings
output:
  default_mode: "active_window"  # Options: "clipboard", "active_window", "combined"
  clipboard_tool: "auto"     # Options: "auto", "wl-copy", "xclip"
  type_tool: "auto"          # Options: "auto", "xdotool"

# WebSocket server settings (for future web integration)
web_server:
  enabled: false  # Enable for React web app integration
  port: 8080
  host: "localhost"
```