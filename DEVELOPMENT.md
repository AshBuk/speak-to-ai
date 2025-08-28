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
make all                 # Build everything (deps + whisper + binary)
make build               # Build binary with whisper.cpp integration
make build-systray       # Build with system tray support
make test                # Run unit tests with CGO dependencies
make test-integration    # Run integration tests (fast mode, no CGO)
make test-integration-full # Run full integration tests (with audio/CGO)
make deps                # Download Go dependencies
make whisper-libs        # Build whisper.cpp libraries into ./lib
make check-tools         # Verify required tools
make appimage            # Build AppImage package
make flatpak             # Build Flatpak package
make clean               # Clean build artifacts
make fmt                 # Format Go code
make lint                # Run linter (via Docker)
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
make docker-up         # Start development services
make docker-dev        # Enter development container
make docker-down       # Stop all services
make docker-build      # Build all Docker images
make docker-appimage   # Build AppImage in container
make docker-flatpak    # Build Flatpak in container
make docker-clean      # Clean Docker resources
make docker-clean-all  # Clean everything including images
```

### 4. CI/CD - Production
GitHub Actions handle complex builds, releases, and distribution.

### Test Types & Current Coverage 

| Test Type | Command | Duration | Dependencies | Use Case |
|-----------|---------|----------|--------------|----------|
| **Unit Tests** | `make test` | ~2-5s | CGO, whisper.cpp | Core functionality |
| **Integration Fast** | `make test-integration` | ~0.3s | None | Development, CI |
| **Integration Full** | `make test-integration-full` | ~5-15s | Audio devices, CGO | QA, Production |


| Module | Coverage | Test Files | Integration Tests |
|--------|----------|------------|-------------------|
| **audio** | 57.1% | `*_test.go` | Audio devices, streaming |
| **output** | 68.3% | `*_test.go` | Clipboard, typing tools |
| **hotkeys** | 57.4% | `*_test.go` | D-Bus, evdev providers |
| **config** | 84.9% | `*_test.go` | File loading, validation |
| **internal/** | 87.5%+ | `*_test.go` | Platform detection |

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