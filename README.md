# Speak to AI

**Speak to AI** is an open-source local speech-to-text application for **Linux** that converts speech into text and inserts it into AI chats and other applications.  
The application runs in the background as a daemon, providing system-wide access via hotkeys for a seamless experience.

---

## Features

- **100% Local & Private**  
  Speech processing happens entirely on your device — no internet required, no data sent to external servers.

- **Whisper.cpp Integration**  
  Uses an optimized implementation of OpenAI's Whisper model for efficient CPU-based transcription.

- **Minimalist Design**  
  Lightweight and fast. Built with simplicity, performance, and low system resource usage in mind.

- **Multiple Output Methods**  
  - 📋 Copy to clipboard (`wl-copy`, `xclip`)  
  - ⌨️ Simulate input into active window (`xdotool`, `wl-clipboard`)

- **Hotkey Support**  
  - Built-in support for Copilot key (or Alt+Comma)
  - Trigger voice input from any application with customizable global hotkeys

- **Desktop Integration**
  - System tray indicator (🎤 / 💤)
  - Desktop notifications
  - Automatic configuration on first launch

- **Linux Compatibility**  
  Fully supports both **Wayland** and **X11** display servers.

---

## 📋 Installation

### Pre-built Binaries

Download AppImage from the Releases page and make it executable:

```bash
chmod +x Speak-to-AI-*.AppImage
```

Run the AppImage:

```bash
./Speak-to-AI-*.AppImage
```

The first time you run the application, it will automatically:
1. Create a configuration file at `~/.config/speak-to-ai/config.yaml`
2. Set up the application to use the embedded Whisper model

---

## 🔧 Configuration

After installation, you can customize the configuration by editing:

```
~/.config/speak-to-ai/config.yaml
```

### Key Configuration Options:

- **Hotkeys**
  - Default recording toggle: `Copilot` or `alt+comma` keys 
  - Copy to clipboard: `ctrl+shift+c`
  - Paste to active window: `ctrl+shift+v`

- **Audio Settings**
  - Recording tool: `arecord` (or `ffmpeg`)
  - Sample rate, format, and channels

- **Output Settings**
  - Default mode: `clipboard` (options: `clipboard`, `active_window`)

---

## 🧩 Dependencies

All necessary dependencies are included in the AppImage! No additional installation required.

- `whisper.cpp` (included)
- `xclip` for X11 clipboard (included)
- `wl-clipboard` for Wayland clipboard
- `notify-send` for desktop notifications (included)

## 🛠️ Building from Source

If you want to build from source, refer to the Development Documentation.

---
