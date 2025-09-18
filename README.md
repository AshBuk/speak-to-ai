<div align="center">

<img src="https://github.com/AshBuk/speak-to-ai/raw/master/icons/io.github.ashbuk.speak-to-ai.svg" width="180" height="180" alt="Speak to AI"/>

# Speak to AI

> Native Linux voice-to-text app 🗣️ 

</div>

<div align="center">

[![CI](https://github.com/AshBuk/speak-to-ai/actions/workflows/ci.yml/badge.svg)](https://github.com/AshBuk/speak-to-ai/actions/workflows/ci.yml)
[![Build Releases](https://github.com/AshBuk/speak-to-ai/actions/workflows/build-releases.yml/badge.svg)](https://github.com/AshBuk/speak-to-ai/actions/workflows/build-releases.yml)
[![Release](https://img.shields.io/github/v/release/AshBuk/speak-to-ai?sort=semver)](https://github.com/AshBuk/speak-to-ai/releases)
[![Go Version](https://img.shields.io/badge/go-1.24-00ADD8?logo=go)](https://go.dev/)
[![Security](https://snyk.io/test/github/AshBuk/speak-to-ai/badge.svg)](https://snyk.io/test/github/AshBuk/speak-to-ai)

[![OS](https://img.shields.io/badge/OS-Linux-34a853?logo=linux)](#-system-requirements)
[![Display](https://img.shields.io/badge/Display-Wayland%20%2F%20X11-ff69b4)](#-features)
[![Privacy](https://img.shields.io/badge/Privacy-Offline-blueviolet)](#-features)
[![AppImage](https://img.shields.io/badge/AppImage-available-0a7cff?logo=appimage)](https://github.com/AshBuk/speak-to-ai/releases)
[![Flatpak](https://img.shields.io/badge/Flatpak-available-4a90e2?logo=flatpak)](https://github.com/AshBuk/speak-to-ai/releases)

</div>

 **A minimalist, privacy-focused desktop application that enables voice input (speech to text) for redactors, IDE or AI assistants without sending your voice to the cloud. Uses the Whisper model locally for speech recognition. Written in Go, an optimized desktop application for Linux.**

## Features

- **Portable**: available as AppImage and Flatpak
- **Cross-platform support** for X11 and Wayland
- **Desktop Environment Support**: Native integration with GNOME, KDE, and other Linux DEs
- **Privacy-first**: desktop, no data sent to external servers
- **Key binding support, Automatic typing and Clipboard support, WebSocket API (optional, for external integrations), Visual notifications**

## ✦ Installation

### AppImage

Download the latest AppImage from [Releases](https://github.com/AshBuk/speak-to-ai/releases):

```bash
# Download the file, then:
chmod +x speak-to-ai-*.AppImage
# Ensure user is in input group for hotkeys to work:
sudo usermod -a -G input $USER
# then reboot
# Open via GUI or command:
./speak-to-ai-*.AppImage  # Replace with your downloaded version
```

### Flatpak 

Download and install the Flatpak from [Releases](https://github.com/AshBuk/speak-to-ai/releases):

```bash
# Download the file, then:
flatpak install --user io.github.ashbuk.speak-to-ai.flatpak  # Replace with your downloaded version
# Run the application
flatpak run io.github.ashbuk.speak-to-ai
```

## Desktop Environment Compatibility

GNOME (Wayland/X11) and KDE Plasma (Wayland/X11) have native support. Help us test different desktop environments:

📋 **[Desktop Environment Support Guide](docs/Desktop_Environment_Support.md)**

**For system tray integration on GNOME**:
```bash
# Ubuntu/Debian
sudo apt install gnome-shell-extension-appindicator
# Fedora
sudo dnf install gnome-shell-extension-appindicator
# Arch Linux
sudo pacman -S gnome-shell-extension-appindicator
```
*KDE and other DEs have built-in system tray support*

**Text typing for GNOME/Wayland**

If text doesn't appear in applications automatically, the app falls back to clipboard mode. To enable direct typing:

```bash
# Install ydotool
sudo dnf install ydotool  # Fedora
sudo apt install ydotool  # Ubuntu/Debian

# Add user to input group
sudo usermod -a -G input $USER
# Reboot required

# Enable daemon (logout/login if required)
sudo systemctl enable --now ydotoold
```

## ✦ Project Status

Functionality and Go code are ready. Currently improving UI/UX as for now it's more geek-friendly than user-friendly. Working on quality AppImage and Flatpak builds.


## ✦ For Developers

Start onboarding with:

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) — system architecture and component design
- [DEVELOPMENT.md](docs/DEVELOPMENT.md) — development workflow and build instructions
- [CONTRIBUTING.md](docs/CONTRIBUTING.md) — contribution guidelines and how to help improve the project
- [docker/README.md](docker/README.md) — Docker-based development


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
## ✦ License

MIT — see `LICENSE`.

## Sponsor

[![Sponsor](https://img.shields.io/badge/Sponsor-💖-pink?style=for-the-badge&logo=github)](https://github.com/sponsors/AshBuk) [![PayPal](https://img.shields.io/badge/PayPal-00457C?style=for-the-badge&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?hosted_button_id=R3HZH8DX7SCJG)

If you find Speak-to-AI useful, please consider supporting development. Your support helps improve the app, real-time streaming, and security hardening.

