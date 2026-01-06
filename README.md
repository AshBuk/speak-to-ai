<div align="center">
<img src="https://github.com/AshBuk/speak-to-ai/raw/master/icons/io.github.ashbuk.speak-to-ai.svg" width="180" height="180" alt="Speak to AI"/>
<h1>Speak to AI</h1>
  
> 🗣️ **Native Linux Voice-To-Text App**

<p style="margin-bottom: 12px;">
<a href="https://pkg.go.dev/github.com/AshBuk/speak-to-ai">
  <img src="https://pkg.go.dev/badge/github.com/AshBuk/speak-to-ai.svg" alt="Go Reference">
</a>
<a href="https://goreportcard.com/report/github.com/AshBuk/speak-to-ai">
  <img src="https://goreportcard.com/badge/github.com/AshBuk/speak-to-ai" alt="Go Report Card">
</a>
<a href="https://go.dev/">
  <img src="https://img.shields.io/badge/go-1.24-00ADD8?logo=go" alt="Go Version">
</a>
</p>
  
[![CI](https://github.com/AshBuk/speak-to-ai/actions/workflows/ci.yml/badge.svg)](https://github.com/AshBuk/speak-to-ai/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/AshBuk/speak-to-ai?sort=semver)](https://github.com/AshBuk/speak-to-ai/releases)
[![AppImage](https://img.shields.io/badge/AppImage-available-0a7cff?logo=appimage)](https://github.com/AshBuk/speak-to-ai/releases)
[![AUR](https://img.shields.io/aur/version/speak-to-ai?logo=archlinux)](https://aur.archlinux.org/packages/speak-to-ai)
[![COPR](https://img.shields.io/badge/COPR-Fedora-51a2da?logo=fedora)](https://copr.fedorainfracloud.org/coprs/ashbuk/speak-to-ai/)

</div>

**Speak to AI** is a **minimalist**, **privacy-focused** desktop application for **offline voice recognition** directly into any active window (editors, browsers, IDEs, AI assistants).  

Written in pure **[Go](https://github.com/golang/go)**, it leverages **[whisper.cpp](https://github.com/ggerganov/whisper.cpp)** for fast, offline transcription.
The architecture is built from the ground up without external frameworks, featuring a **custom dependency injection factory** and a **minimal set of dependencies**, ensuring **lean and maintainable**.

https://github.com/user-attachments/assets/e8448f73-57f2-46dc-98f9-e36f685a6587

## Features

[![Privacy](https://img.shields.io/badge/Privacy-Offline-blueviolet)](#-features)
[![Security](https://snyk.io/test/github/AshBuk/speak-to-ai/badge.svg)](https://snyk.io/test/github/AshBuk/speak-to-ai)
[![gosec](https://img.shields.io/badge/gosec-passing-brightgreen)](https://github.com/AshBuk/speak-to-ai/security/code-scanning)

▸ Speak to AI runs quietly in the background and integrates into the system tray for convenient management. 

▸ It can also be invoked as a CLI tool (see **[CLI Usage Guide](docs/CLI_USAGE.md)**) for scripting purposes. 

▸ For integration enthusiasts, a WebSocket server is available at `localhost:8080`. Enable it in your config with web_server enabled: true (disabled by default).

- **Offline speech-to-text, privacy-first**: all processing happens locally
- **Portable**: AppImage package
- **Cross-platform support** for X11 and Wayland
- **Linux DEs**: native integration with GNOME, KDE, and others
- **Voice typing or clipboard mode**
- **Flexible audio recording**: arecord (ALSA) or ffmpeg (PulseAudio/PipeWire), see [audio pipeline](docs/AUDIO_PIPELINE_DIAGRAM.txt)
- **Multi-language support, custom hotkey binding, visual notifications**

## Beyond Minimalism

Intuitive minimalist UX, **robust STT infrastructure**. A foundation for voice-controlled automation:

- **Dual API**: Unix socket IPC + WebSocket — script locally or integrate remotely
- **Interface-driven**: 50+ contracts — swap STT engines, add I/O methods, extend hotkey providers
- **Daemon + CLI**: background hub + stateless commands — perfect for IoT pipelines
- **Graceful degradation**: provider fallbacks, optional components, no crashes

```bash
# Voice command → smart home action
transcript=$(speak-to-ai stop-recording | jq -r '.data.transcript')
[[ "$transcript" == *"lights off"* ]] && curl -X POST http://hub/lights/off
```

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

### Arch Linux [AUR](https://aur.archlinux.org/packages/speak-to-ai):
```bash
yay -S speak-to-ai
```

### Fedora [COPR](https://copr.fedorainfracloud.org/coprs/ashbuk/speak-to-ai/):
```bash
sudo dnf copr enable ashbuk/speak-to-ai
sudo dnf install speak-to-ai
```

## Desktop Environment Compatibility

[![OS](https://img.shields.io/badge/OS-Linux-34a853?logo=linux)](#-system-requirements)
[![Display](https://img.shields.io/badge/Display-Wayland%20%2F%20X11-ff69b4)](#-features)

📋 **[Desktop Environment Support Guide](docs/Desktop_Environment_Support.md)** - help us test different desktop environments!

**For system tray integration on GNOME — [install the AppIndicator extension](docs/Desktop_Environment_Support.md#for-system-tray-on-gnome---to-have-full-featured-ux-with-menu) ↑**
> KDE and other DEs have built-in system tray support out of the box

**For automatic typing on GNOME — [see setup guide](docs/Desktop_Environment_Support.md#direct-typing-on-wayland---tool-options) ↑**  
> **Other Wayland compositors** (KDE, Hyprland, Sway, etc.): wtype works without setup — automatically detected!  
> **X11**: Native support with xdotool out of the box

> If automatic typing doesn't appear automatically, the app falls back to clipboard (Ctrl + V) mode

For issues and bug reports: [GitHub Issues](https://github.com/AshBuk/speak-to-ai/issues)

See changes: [CHANGELOG.md](CHANGELOG.md)

### System Requirements

| Category | Requirement |
|----------|-------------|
| **OS** | Linux with glibc 2.35+ |
| **Desktop** | X11 or Wayland |
| **Audio** | Microphone capability |
| **Storage** | ~277MB |
| **Memory** | ~300MB RAM |
| **CPU** | AVX-capable (Intel/AMD 2011+) |

<details>
<summary><b>📋 Supported Distributions</b></summary>

| Family | Distributions |
|--------|---------------|
| **Ubuntu-based** | Ubuntu 22.04+, Linux Mint 21+, Pop!_OS 22.04+, Elementary OS 7+, Zorin OS 17+ |
| **Debian-based** | Debian 12+ |
| **Fedora** | Fedora 36+ |
| **Rolling release** | Arch Linux, Manjaro, EndeavourOS, openSUSE Tumbleweed |

</details>


## For Developers

Start onboarding with:

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) — system architecture and component design
- [DEVELOPMENT.md](docs/DEVELOPMENT.md) — development workflow and build instructions
- [CONTRIBUTING.md](docs/CONTRIBUTING.md) — contribution guidelines and how to help improve the project
- [docker/README.md](docker/README.md) — Docker-based development

Technical dive into architecture and engineering challenges: [Building Speak-to-AI on Hashnode](https://ashbuk.hashnode.dev/an-offline-voice-to-text-solution-for-linux-users-using-whispercpp-and-go)

## ✦ Acknowledgments

- [whisper.cpp](https://github.com/ggerganov/whisper.cpp) for the excellent C++ implementation of OpenAI Whisper
- [fyne.io/systray](https://github.com/fyne-io/systray) for cross-platform system tray support
- [ydotool](https://github.com/ReimuNotMoe/ydotool) and [wtype](https://github.com/atx/wtype) for Wayland-compatible input automation
- OpenAI for the original Whisper model

## ✦ MIT [LICENSE](LICENSE)

If you use this project, please link back to this repo and ⭐ it if it helped you.
- Consider contributing back improvements

---

Sharing with the community for privacy-conscious Linux users

---

## Sponsor

[![Sponsor](https://img.shields.io/badge/Sponsor-💖-pink?style=for-the-badge&logo=github)](https://github.com/sponsors/AshBuk) [![PayPal](https://img.shields.io/badge/PayPal-00457C?style=for-the-badge&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?hosted_button_id=R3HZH8DX7SCJG)

Please consider supporting development
