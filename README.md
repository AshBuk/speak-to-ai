# Speak-to-AI

[![Build Releases](https://github.com/AshBuk/speak-to-ai/actions/workflows/build-releases.yml/badge.svg)](https://github.com/AshBuk/speak-to-ai/actions/workflows/build-releases.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/AshBuk/speak-to-ai?sort=semver)](https://github.com/AshBuk/speak-to-ai/releases)
[![Go Version](https://img.shields.io/badge/go-1.24-00ADD8?logo=go)](https://go.dev/)
[![OS](https://img.shields.io/badge/OS-Linux-34a853?logo=linux)](#-system-requirements)
[![Display](https://img.shields.io/badge/Display-Wayland%20%2F%20X11-ff69b4)](#-features)
[![Privacy](https://img.shields.io/badge/Privacy-Offline-blueviolet)](#-features)
[![AppImage](https://img.shields.io/badge/AppImage-available-0a7cff?logo=appimage)](https://github.com/AshBuk/speak-to-ai/releases)
[![Flatpak](https://img.shields.io/badge/Flatpak-available-4a90e2?logo=flatpak)](https://github.com/AshBuk/speak-to-ai/releases)

A minimalist, privacy-focused desktop application that enables voice input (speech to text) for redactors, IDE or AI assistants without sending your voice to the cloud. Uses the Whisper model locally for speech recognition. Written in Go, an optimized desktop application for Linux.

## Features

- **Cross-platform support** for X11 and Wayland
- **Desktop Environment Support**: Native integration with GNOME, KDE, and other Linux DEs
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

Get prebuilt packages on the [Releases](https://github.com/AshBuk/speak-to-ai/releases) page:

- AppImage: portable binary for most Linux distributions
- Flatpak: sandboxed install

### AppImage

Download the latest AppImage from [Releases](https://github.com/AshBuk/speak-to-ai/releases):

```bash
chmod +x speak-to-ai-*.AppImage
./speak-to-ai-*.AppImage
```

### Flatpak

Download and install the Flatpak from [Releases](https://github.com/AshBuk/speak-to-ai/releases):

```bash
# Download the file, then:
flatpak install --user io.github.ashbuk.speak-to-ai.flatpak
# Grant input device permissions (recommended)
flatpak override --user --device=input io.github.ashbuk.speak-to-ai
# Run the application
flatpak run io.github.ashbuk.speak-to-ai
```

## ✦ Configuration

Configuration file is automatically created at:
- **AppImage**: `~/.config/speak-to-ai/config.yaml`
- **Flatpak**: `~/.var/app/io.github.ashbuk.speak-to-ai/config/speak-to-ai/config.yaml`

### Desktop Environment Compatibility

#### GNOME & KDE (Recommended)
Global hotkeys work seamlessly using the `org.freedesktop.portal.GlobalShortcuts` portal:
- **GNOME**: Full native support, no additional configuration needed
- **KDE Plasma**: Full native support, no additional configuration needed

#### Other Desktop Environments
For DEs without GlobalShortcuts portal support (XFCE, MATE, i3, etc.):
- Hotkeys may be limited by Flatpak sandboxing
- Optional: Grant input device access for better hotkey support:

```bash
flatpak override --user --device=input io.github.ashbuk.speak-to-ai
```

Then restart the app. This is optional and only needed on DEs without GlobalShortcuts portal.

## ✦ Project Status

Stable builds are available. The app works offline with local Whisper models, supports X11/Wayland, and provides clipboard/typing output modes. Planned improvements: streaming API integration, model download UX, and additional language models.


## ✦ For Developers

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

## ✦ System Requirements

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
