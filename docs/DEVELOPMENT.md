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
make test                # Run unit tests via Docker (CGO + whisper.cpp)
make test-integration    # Run integration tests via Docker (fast mode)
make test-integration-full # Run full integration tests via Docker (audio/CGO)
make deps                # Download Go dependencies
make whisper-libs        # Build whisper.cpp libraries into ./lib
make check-tools         # Verify required tools
make appimage            # Build AppImage package
make flatpak             # Build Flatpak package
make clean               # Clean build artifacts
make fmt                 # Format Go code (go fmt + goimports)
make lint                # Run linter and code quality checks
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
make docker-dev-stop   # Stop development environment
make docker-down       # Stop all services
make docker-shell      # Open shell in dev container
make docker-build      # Build all Docker images
make docker-build-dev  # Build development Docker image
make docker-build-lint # Using official golangci-lint image
make docker-whisper    # Build whisper.cpp libraries in Docker
make docker-appimage   # Build AppImage in container
make docker-flatpak    # Build Flatpak in container
make docker-ci         # Run full CI pipeline
make docker-logs       # Show Docker logs
make docker-ps         # Show Docker containers
make docker-clean      # Clean Docker resources
make docker-clean-all  # Clean everything including images
```

### 4. CI/CD - Production
GitHub Actions handle complex builds, releases, and distribution.

### Test Types & Current Coverage 

| Test Type | Command | Duration | Dependencies | Use Case |
|-----------|---------|----------|--------------|----------|
| **Unit Tests** | `make test` | ~3-8s | Docker dev image | Core functionality |
| **Integration Fast** | `make test-integration` | ~2-5s | Docker dev image | Development, CI |
| **Integration Full** | `make test-integration-full` | ~8-20s | Docker dev image + audio | QA, Production |


| Module | Test Coverage | Test Files | Integration Tests |
|--------|---------------|------------|-------------------|
| **audio** | Comprehensive | `factory/`, `recorders/`, `interfaces/` | Audio devices, recording |
| **output** | Comprehensive | `factory/`, `outputters/` | Clipboard, typing tools |
| **hotkeys** | Comprehensive | `manager/`, `providers/`, `adapters/` | D-Bus, evdev fallback providers |
| **config** | Comprehensive | `loaders/`, `validators/`, `security/` | File loading, validation |
| **internal/** | Comprehensive | `services/`, `platform/`, `utils/` | Platform detection, services |

## Hotkeys Architecture

### Provider Chain, Override & Fallback
```
DBus GlobalShortcuts (GNOME/KDE) → Evdev (i3/XFCE/MATE/AppImage)
      ↑                                  ↑
   preferred                        fallback
   (portal)                       (direct input)
```

### Module Structure
```
hotkeys/
├── adapters/            # Hotkey adapters and utilities
├── interfaces/          # Provider interface definitions  
├── manager/             # Manager with fallback logic
│   ├── manager.go       # Main orchestrator
│   └── provider_fallback.go # Fallback registration
├── providers/           # Hotkey providers implementation
│   ├── dbus_provider.go # GlobalShortcuts portal (GNOME/KDE)
│   ├── evdev_provider.go # Direct input devices (i3/XFCE/AppImage)
│   └── dummy_provider.go # Testing provider
├── utils/               # Parser utilities
└── hotkeys.go          # Main entry point
```

### Example Configuration

```yaml
# General settings
general:
  debug: false
  model_path: "sources/language-models/small-q5_1.bin"  # Default small-q5_1 model
  temp_audio_path: "/tmp"
  model_precision: "q5_1"
  language: "en"  # Recognition language: "en", "ru", "de", "fr", "es", "he", etc.
  log_file: "logs/speak-to-ai.log"  # Log file path (empty to disable)

  # Single model configuration - only small-q5_1
  models:
    - "sources/language-models/small-q5_1.bin"      # Fixed model - small q5_1 quantized
  active_model: "sources/language-models/small-q5_1.bin"  # Fixed active model

# Hotkeys settings
hotkeys:
  start_recording: "ctrl+alt+r"       # Main recording hotkey
  stop_recording: "ctrl+alt+r"        # Same combination for start/stop
  show_config: "altgr+shift+c"        # Open configuration file
  reset_to_defaults: "altgr+shift+r"  # Reset all settings to defaults
  # toggle_vad: "altgr+shift+v"       # Toggle VAD (future feature)

# Audio settings
audio:
  device: "default"
  sample_rate: 16000            # Optimal for speech recognition
  format: "s16le"               # Audio recorded in mono (1 channel)
  recording_method: "arecord"   # Options: "arecord", "ffmpeg"
  # Future VAD implementation
  # enable_vad: false           # Enable Voice Activity Detection
  # vad_sensitivity: "medium"   # VAD sensitivity: "low", "medium", "high"

# Output settings
output:
  default_mode: "active_window"  # Options: "clipboard", "active_window", "web"
  clipboard_tool: "auto"        # Options: "auto", "wl-copy", "xclip"
  type_tool: "auto"             # Options: "auto", "xdotool", "wl-clipboard", "dbus"

# WebSocket server settings
web_server:
  enabled: false  # Disabled by default - enable only if needed
  port: 8080
  host: "localhost"
```