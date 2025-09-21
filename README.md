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

 **A minimalist, privacy-focused desktop application for offline speech-to-text.
  Converts voice input directly into any active window (editors, browsers, IDEs, AI assistants) 
  Uses the Whisper model locally for speech recognition. 
  Written in Go, an optimized desktop application for Linux.**

https://github.com/user-attachments/assets/ff2ef633-40df-46be-b9c8-ab8ae16e0101

## Features

- **Portable**: available as AppImage and Flatpak
- **Cross-platform support** for X11 and Wayland
- **Desktop Environment Support**: Native integration with GNOME, KDE, and other Linux DEs
- **Privacy-first**: desktop, no data sent to external servers
- **Support: multi-language, global hotkeys, automatic typing & clipboard, system tray integration, visual notifications, WebSocket API (optional)**

## ✦ Installation

### AppImage

Download the latest AppImage from [Releases](https://github.com/AshBuk/speak-to-ai/releases):

```bash
# Download the file, then:
chmod +x speak-to-ai-v1.0.0.AppImage  # Replace with your downloaded version
# Ensure user is in input group for hotkeys to work:
sudo usermod -a -G input $USER
# then reboot
# Open via GUI or command:
./speak-to-ai-v1.0.0.AppImage
```

### Flatpak 

Download and install the Flatpak from [Releases](https://github.com/AshBuk/speak-to-ai/releases):

```bash
# Download the file, then:
flatpak install --user speak-to-ai-v1.0.0.flatpak  # Replace with your downloaded version
# Run the application:
flatpak run io.github.ashbuk.speak-to-ai
```

## Desktop Environment Compatibility

Help us test different desktop environments:

📋 **[Desktop Environment Support Guide](docs/Desktop_Environment_Support.md)**

**For system tray integration on GNOME, [install the AppIndicator extension](docs/Desktop_Environment_Support.md#for-system-tray-on-gnome---to-have-full-featured-ux-with-menu) ↑**
> KDE and other DEs have built-in system tray support out of the box

**For automatic typing on Wayland (GNOME and others) — [set up ydotool](docs/Desktop_Environment_Support.md#direct-typing-on-wayland---ydotool-setup-recommended-user-unit) ↑**
> X11 has native typing support with xdotool out of the box

> If automatic typing doesn't appear automatically, the app falls back to clipboard (Ctrl + V) mode

## ✦ Project Status

[v1.0.0 release](https://github.com/AshBuk/speak-to-ai/releases) with working functionality, tested on my Fedora 42 (GNOME/Wayland) and Ubuntu 24. I'd appreciate feedback about your experience on your system!

For issues and bug reports: [GitHub Issues](https://github.com/AshBuk/speak-to-ai/issues)

See changes: [CHANGELOG.md](CHANGELOG.md)


## For Developers

Start onboarding with:

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) — system architecture and component design
- [DEVELOPMENT.md](docs/DEVELOPMENT.md) — development workflow and build instructions
- [CONTRIBUTING.md](docs/CONTRIBUTING.md) — contribution guidelines and how to help improve the project
- [docker/README.md](docker/README.md) — Docker-based development


## System Requirements

- **OS**: Linux (Ubuntu 20.04+, Fedora 35+, or similar)
- **Desktop**: X11 or Wayland environment
- **Audio**: Microphone/recording capability
- **Storage**: 277.5MB (whisper small q5 model, dependencies, go-binary)
- **Memory**: ~500MB RAM during operation

## ✦ Acknowledgments

- [whisper.cpp](https://github.com/ggerganov/whisper.cpp) for the excellent C++ implementation of OpenAI Whisper
- [fyne.io/systray](https://github.com/fyne-io/systray) for cross-platform system tray support
- OpenAI for the original Whisper model

---

Sharing with the community for privacy-conscious Linux users

---
## ✦ License

MIT — see `LICENSE`.

## Sponsor

[![Sponsor](https://img.shields.io/badge/Sponsor-💖-pink?style=for-the-badge&logo=github)](https://github.com/sponsors/AshBuk) [![PayPal](https://img.shields.io/badge/PayPal-00457C?style=for-the-badge&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?hosted_button_id=R3HZH8DX7SCJG)

If you find Speak-to-AI useful, please consider supporting development.