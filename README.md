# 🎤 Speak-to-AI

[![Build Releases](https://github.com/AshBuk/speak-to-ai/actions/workflows/build-releases.yml/badge.svg)](https://github.com/AshBuk/speak-to-ai/actions/workflows/build-releases.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A minimalist, privacy-focused desktop application that enables voice input (speech to text) for redactors, IDE or AI assistants without sending your voice to the cloud. Uses the Whisper model locally for speech recognition.

## ✨ Features

- 🖥️ **Cross-platform support** for X11 and Wayland
- 🔒 **Privacy-first**: no data sent to external servers
- 📦 **Portable**: available as AppImage and Flatpak

- **100% Offline** speech recognition using Whisper.cpp
- **System tray integration** with recording status (🎤 / 💤)
- **Key binding support** (AltGr + ,) and customizable hotkeys
- **Automatic typing** in active window after transcription
- **Clipboard support** for copying transcribed text
- **WebSocket API** for external integrations (optional)
- **Visual notifications** for statuses


## 🚀 Installation

### AppImage

Download the latest AppImage from [Releases](https://github.com/AshBuk/speak-to-ai/releases):

```bash
chmod +x speak-to-ai-*.AppImage
./speak-to-ai-*.AppImage
```

### Flatpak

Download and install the Flatpak from [Releases](https://github.com/AshBuk/speak-to-ai/releases):

```bash
flatpak install speak-to-ai-*.flatpak
flatpak run io.github.ashbuk.speak-to-ai
```

## 🔧 Configuration

Configuration file is automatically created at:
- **AppImage**: `~/.config/speak-to-ai/config.yaml`
- **Flatpak**: `~/.var/app/io.github.ashbuk.speak-to-ai/config/speak-to-ai/config.yaml`

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

## 🔨 Building from Source

### Prerequisites (for developers)

- Go 1.21+
- Linux development libraries:
  ```bash
  # Ubuntu/Debian
  sudo apt install libasound2-dev libx11-dev libxext-dev libxi-dev libxrandr-dev
  
  # Fedora
  sudo dnf install alsa-lib-devel libX11-devel libXext-devel libXi-devel libXrandr-devel
  ```

### Build Commands

```bash
# Clone repository
git clone https://github.com/AshBuk/speak-to-ai.git
cd speak-to-ai

# Build everything (recommended)
make all

# Or build individual components
make build           # Build executable only
make build-systray   # Build with system tray support
make whisper-libs    # Build whisper.cpp libraries only
make test           # Run tests

# Build packages
make appimage       # Build AppImage
make flatpak        # Build Flatpak (requires flatpak-builder)

# Other commands
make clean          # Clean build artifacts
make help           # Show all available targets
```

## 🏗️ Architecture & Components

- **Local Daemon**: Go application handling hotkeys, audio recording, and output
- **Whisper Engine**: Uses `whisper.cpp` binary for speech recognition
- **Audio Recording**: Supports `arecord` and `ffmpeg` backends
- **Text Output**: 
  - **Active Window Mode**: Automatically types transcribed text into the currently active window
  - **Clipboard Mode**: Copies transcribed text to system clipboard
  - **Combined Mode**: Both typing and clipboard operations
- **WebSocket Server**: Provides API for external applications (optional, port 8080)

## 📋 System Requirements

- **OS**: Linux (Ubuntu 20.04+, Fedora 35+, or similar)
- **Desktop**: X11 or Wayland environment
- **Audio**: Microphone/recording capability
- **Storage**: ~200MB for model and dependencies
- **Memory**: ~500MB RAM during operation

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [whisper.cpp](https://github.com/ggerganov/whisper.cpp) for the excellent C++ implementation of OpenAI Whisper
- [getlantern/systray](https://github.com/getlantern/systray) for cross-platform system tray support
- OpenAI for the original Whisper model

---

**Sharing with the community with ❤️, for privacy-conscious Linux users**

---
