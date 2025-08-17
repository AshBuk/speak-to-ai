# Speak-to-AI

[![Build Releases](https://github.com/AshBuk/speak-to-ai/actions/workflows/build-releases.yml/badge.svg)](https://github.com/AshBuk/speak-to-ai/actions/workflows/build-releases.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A minimalist, privacy-focused desktop application that enables voice input (speech to text) for redactors, IDE or AI assistants without sending your voice to the cloud. Uses the Whisper model locally for speech recognition.

## Features

- **Cross-platform support** for X11 and Wayland
- **Privacy-first**: no data sent to external servers
- **Portable**: available as AppImage and Flatpak

- **100% Offline** speech recognition using Whisper.cpp
- **System tray integration** with recording status (🎤 / 💤)
- **Key binding support** (AltGr + ,) and customizable hotkeys
- **Automatic typing** in active window after transcription
- **Clipboard support** for copying transcribed text
- **WebSocket API** for external integrations (optional)
- **Visual notifications** for statuses


## ✦ Installation

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

## Configuration

Configuration file is automatically created at:
- **AppImage**: `~/.config/speak-to-ai/config.yaml`
- **Flatpak**: `~/.var/app/io.github.ashbuk.speak-to-ai/config/speak-to-ai/config.yaml`

### Hotkeys in Flatpak

- On GNOME/KDE, global hotkeys use the `org.freedesktop.portal.GlobalShortcuts` portal and work out of the box (no extra permissions needed).
- On other desktop environments where the portal is not available, hotkeys may be limited by sandboxing. You can opt-in to input device access:

```bash
flatpak override --user --device=input io.github.ashbuk.speak-to-ai
```

Then restart the app. This is optional and only needed on DEs without GlobalShortcuts portal.

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

## Project Status

Stable builds are available. The app works offline with local Whisper models, supports X11/Wayland, and provides clipboard/typing output modes. Planned improvements: streaming API integration, model download UX, and additional language models.

## Downloads

Get prebuilt packages on the [Releases](https://github.com/AshBuk/speak-to-ai/releases) page:

- AppImage: portable binary for most Linux distributions
- Flatpak: sandboxed install

Follow the steps in the Installation section above.

## For Developers

Developer documentation has moved to:

- DEVELOPMENT.md — development workflow and build instructions
- docker/README.md — Docker-based development

## ✦ Architecture & Components

- **Local Daemon**: Go application handling hotkeys, audio recording, and output
- **Whisper Engine**: Uses `whisper.cpp` binary for speech recognition
- **Audio Recording**: Supports `arecord` and `ffmpeg` backends
- **Text Output**: 
  - **Active Window Mode**: Automatically types transcribed text into the currently active window
  - **Clipboard Mode**: Copies transcribed text to system clipboard
  - **Combined Mode**: Both typing and clipboard operations
- **WebSocket Server**: Provides API for external applications (optional, port 8080)

## System Requirements

- **OS**: Linux (Ubuntu 20.04+, Fedora 35+, or similar)
- **Desktop**: X11 or Wayland environment
- **Audio**: Microphone/recording capability
- **Storage**: ~200MB for model and dependencies
- **Memory**: ~500MB RAM during operation

## ✦ Acknowledgments

- [whisper.cpp](https://github.com/ggerganov/whisper.cpp) for the excellent C++ implementation of OpenAI Whisper
- [getlantern/systray](https://github.com/getlantern/systray) for cross-platform system tray support
- OpenAI for the original Whisper model

---

Sharing with the community for privacy-conscious Linux users

---

## ❤️ Sponsors

If you find Speak-to-AI useful, please consider supporting development via [GitHub Sponsors](https://github.com/sponsors/AshBuk). Your support helps improve Wayland hotkeys, real-time streaming, and security hardening.

✦ License
MIT — see LICENSE.
