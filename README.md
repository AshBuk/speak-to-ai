<div align="center">

<img src="https://github.com/AshBuk/speak-to-ai/raw/master/icons/io.github.ashbuk.speak-to-ai.svg" width="180" height="180" alt="Speak to AI"/>

# Speak to AI

> Native Linux voice-to-text app 🗣️ 

</div>

<div align="center">

[![Go Reference](https://pkg.go.dev/badge/github.com/AshBuk/speak-to-ai.svg)](https://pkg.go.dev/github.com/AshBuk/speak-to-ai)

[![CI](https://github.com/AshBuk/speak-to-ai/actions/workflows/ci.yml/badge.svg)](https://github.com/AshBuk/speak-to-ai/actions/workflows/ci.yml)
[![Build Releases](https://github.com/AshBuk/speak-to-ai/actions/workflows/build-releases.yml/badge.svg)](https://github.com/AshBuk/speak-to-ai/actions/workflows/build-releases.yml)
[![Release](https://img.shields.io/github/v/release/AshBuk/speak-to-ai?sort=semver)](https://github.com/AshBuk/speak-to-ai/releases)
[![Go Version](https://img.shields.io/badge/go-1.24-00ADD8?logo=go)](https://go.dev/)
[![Security](https://snyk.io/test/github/AshBuk/speak-to-ai/badge.svg)](https://snyk.io/test/github/AshBuk/speak-to-ai)

[![OS](https://img.shields.io/badge/OS-Linux-34a853?logo=linux)](#-system-requirements)
[![Display](https://img.shields.io/badge/Display-Wayland%20%2F%20X11-ff69b4)](#-features)
[![Privacy](https://img.shields.io/badge/Privacy-Offline-blueviolet)](#-features)
[![AppImage](https://img.shields.io/badge/AppImage-available-0a7cff?logo=appimage)](https://github.com/AshBuk/speak-to-ai/releases)

</div>

 **A minimalist, privacy-focused desktop application for offline speech-to-text.
  Converts voice input directly into any active window (editors, browsers, IDEs, AI assistants) 
  Uses the Whisper model locally for speech recognition. 
  Written in Go, an optimized desktop application for Linux.**

https://github.com/user-attachments/assets/e8448f73-57f2-46dc-98f9-e36f685a6587

## Features

- **Offline speech-to-text**
- **Voice typing or clipboard mode**
- **Portable**: AppImage package
- **Cross-platform support** for X11 and Wayland
- **Desktop Environment Support**: Native integration with GNOME, KDE, and other Linux DEs
- **Privacy-first**: desktop, no data sent to external servers
- **Support: multi-language, global hotkeys, automatic typing, system tray integration, visual notifications, WebSocket API (optional)**

## ✦ Installation

### AppImage

Download the latest AppImage from [Releases](https://github.com/AshBuk/speak-to-ai/releases):

```bash
# Download the file, then:
 chmod +x speak-to-ai-*.AppImage
 # Ensure user is in input group for hotkeys to work:
 sudo usermod -a -G input $USER
 # then logout/login or reboot
 # Open via GUI or with terminal command:
 ./speak-to-ai-*.AppImage  
```

### CLI Helper

The AppImage bundles the CLI so you can call it directly:

```bash
./speak-to-ai-*.AppImage start    # Begin recording (requires the daemon to be running)
./speak-to-ai-*.AppImage stop     # Stop and print the transcript to stdout
./speak-to-ai-*.AppImage status   # Check whether recording is active
```

Note: these CLI commands communicate with the running speak-to-ai daemon over a Unix socket — the daemon must already be running for them to work. You can start the daemon simply by launching the AppImage once:

```bash
./speak-to-ai-*.AppImage
```

For source builds or packaging experiments, generate a standalone CLI binary:
```bash
make cli
./speak-to-ai-cli start
```

```bash
make cli
./speak-to-ai-cli start
```

The CLI communicates with an active `speak-to-ai` daemon through a secure Unix socket.  
Launch the main application once (AppImage or `make build`) and then trigger CLI commands from your window manager key bindings or scripts.  
Add the `--json` flag for machine-readable responses or `--socket` to point to a custom IPC socket path.

## Desktop Environment Compatibility

Help us test different desktop environments:

📋 **[Desktop Environment Support Guide](docs/Desktop_Environment_Support.md)**

**For system tray integration on GNOME, [install the AppIndicator extension](docs/Desktop_Environment_Support.md#for-system-tray-on-gnome---to-have-full-featured-ux-with-menu) ↑**
> KDE and other DEs have built-in system tray support out of the box

**For automatic typing on Wayland (GNOME and others) — [set up ydotool](docs/Desktop_Environment_Support.md#direct-typing-on-wayland---ydotool-setup-recommended-user-unit) ↑**
> X11 has native typing support with xdotool out of the box

> If automatic typing doesn't appear automatically, the app falls back to clipboard (Ctrl + V) mode

## ✦ Project Status

**AppImage** [release](https://github.com/AshBuk/speak-to-ai/releases) - main distribution format. I'd appreciate feedback about your experience on your system!

**Flatpak** bundle is planned.

For issues and bug reports: [GitHub Issues](https://github.com/AshBuk/speak-to-ai/issues)

See changes: [CHANGELOG.md](CHANGELOG.md)


## For Developers

Start onboarding with:

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) — system architecture and component design
- [DEVELOPMENT.md](docs/DEVELOPMENT.md) — development workflow and build instructions
- [CONTRIBUTING.md](docs/CONTRIBUTING.md) — contribution guidelines and how to help improve the project
- [docker/README.md](docker/README.md) — Docker-based development

Technical dive into architecture and engineering challenges: [Building Speak-to-AI on Hashnode](https://ashbuk.hashnode.dev/an-offline-voice-to-text-solution-for-linux-users-using-whispercpp-and-go)

## System Requirements

- **OS**: Linux (Ubuntu 20.04+, Fedora 35+, or similar)
- **Desktop**: X11 or Wayland environment
- **Audio**: Microphone/recording capability
- **Storage**: 277.2MB (whisper small q5 model, dependencies, go-binary)
- **Memory**: ~300MB RAM during operation
- **CPU**: AVX-capable processor (Intel/AMD 2011+)

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
