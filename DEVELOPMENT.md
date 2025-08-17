# Development Guide

Concise commands for local development and CI-compatible builds.

## One-time setup

```bash
# Install base tools
sudo apt-get update && sudo apt-get install -y build-essential cmake git pkg-config

# Go modules
make deps

# Build local whisper headers/libs into ./lib
make whisper-libs
```

## Dev session

```bash
# 1) Configure CGO env to use ./lib (whisper headers/libs)
source bash-scripts/dev-env.sh

# 2) Build / test
make build
make test
```

Under the hood, `dev-env.sh` sets `CGO_ENABLED=1`, `CGO_CFLAGS`, `CGO_LDFLAGS`, `LD_LIBRARY_PATH` so the compiler and runtime can locate `whisper.h` and `libwhisper.so` (+ ggml).

## Make targets

```bash
make build          # deps + whisper-libs + build binary
make build-systray  # build with systray tag
make test           # run tests (CGO env pre-configured)
make fmt            # format code (gofmt)
make lint           # run linter (Docker)
make clean          # clean artifacts
```

## Notes
- If whisper.cpp changes, re-run `make whisper-libs`.
- CI uses `golangci/golangci-lint-action@v6` for linting; whisper libs are built in the test job.


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