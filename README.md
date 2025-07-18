# 🎤 Speak-to-AI

[![Build Releases](https://github.com/AshBuk/speak-to-ai/actions/workflows/build-releases.yml/badge.svg)](https://github.com/AshBuk/speak-to-ai/actions/workflows/build-releases.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A minimalist, privacy-focused desktop application that enables voice input for AI assistants without sending your voice to the cloud. Uses the Whisper model locally for speech recognition.

## ✨ Features

- 🎤 **100% Offline** speech recognition using Whisper.cpp
- 🔧 **System tray integration** with recording status (🎤 / 💤)
- ⌨️ **Microsoft Copilot key support** (AltGr + ,) and customizable hotkeys
- 📋 **Multiple output modes**: clipboard copy, direct typing simulation
- 🖥️ **Cross-platform support** for X11 and Wayland
- 🔒 **Privacy-first**: no data sent to external servers
- 🔔 **Visual notifications** for recording status
- 📦 **Portable**: available as AppImage and Flatpak

## 🚀 Installation

### AppImage (Recommended)

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

## ⌨️ Default Hotkeys

- **AltGr + **, (comma): Start/Stop recording (Microsoft Copilot key)
- **Alt + **, (comma): Alternative hotkey

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

# Hotkey settings
hotkeys:
  start_recording: "altgr+comma"    # Microsoft Copilot key
  alt_start_recording: "alt+comma"  # Alternative

# Audio settings
audio:
  device: "default"
  sample_rate: 16000
  recording_method: "arecord"  # Options: "arecord", "ffmpeg"

# Output settings
output:
  default_mode: "clipboard"  # Options: "clipboard", "active_window"
  clipboard_tool: "auto"     # Options: "auto", "wl-copy", "xclip"
```

## 🔨 Building from Source

### Prerequisites

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

# Build executable
go build -o speak-to-ai cmd/daemon/main.go

# Build AppImage
bash bash-scripts/build-appimage.sh

# Build Flatpak (requires flatpak-builder)
bash bash-scripts/build-flatpak.sh
```

## 🏗️ Architecture

```
    A[Hotkey] --> B[Local Daemon (Go)]
    B --> C[whisper.cpp Execution]
    C --> D[Transcript (stdout)]
    D --> E{Mode}
    E -->|Clipboard| F1[wl-copy / xclip]
    E -->|Simulated Input| F2[xdotool / wl-clipboard]
```

### Components

- **Local Daemon**: Go application handling hotkeys, audio recording, and output
- **Whisper Engine**: Uses `whisper.cpp` binary for speech recognition
- **Audio Recording**: Supports `arecord` and `ffmpeg` backends
- **Text Output**: Clipboard copy or typing simulation via system tools

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

**Made with ❤️ for privacy-conscious Linux users**

---
