# Changelog

All notable changes to this project will be documented in this file.

---
## [1.3.0] - 2025-10-29

### Features

- **Dual-Mode Architecture:** Single binary now supports both daemon mode (background service with system tray) and CLI mode (command-line interface for scripting)
- **CLI Commands:** See [CLI Usage Guide](docs/CLI_USAGE.md)
- **Tiling WM Support:** (i3, sway, bspwm, etc.) through CLI hotkey bindings
- **IPC Communication:** CLI commands communicate with daemon via Unix socket for low-latency operations
- **JSON Output:** `--json` flag for machine-readable responses, perfect for automation and scripts
- **High-Resolution Icon:** Upgraded tray icon to 500x500 resolution for crisp display on HiDPI screens

### Contributors

Thanks to [@BigMitchGit](https://github.com/BigMitchGit) for implementing CLI/Daemon unification and Unix socket IPC! (PR [#49](https://github.com/AshBuk/speak-to-ai/pull/49)).


## [1.2.0] - 2025-10-16

### Features

- **Audio Stability:** Enhanced `ffmpeg` recorder for reliable use with PulseAudio/PipeWire, preventing empty or clipped recordings during concurrent audio activity (e.g., video calls).
- **Recorder Reliability:** Added a warm-up phase to ensure valid audio payloads and a robust stop-retry mechanism to prevent data loss in short recordings.
- **Low Latency:** Optimized PulseAudio input buffering to reduce audio capture latency.

*For more details, see the audio pipeline diagram: [docs/AUDIO_PIPELINE_DIAGRAM.txt](docs/AUDIO_PIPELINE_DIAGRAM.txt).*


## [1.1.0] - 2025-10-06

### Features

- **Configuration:** Temporary file manager with a default 30-minute cleanup now configurable via `audio.temp_file_cleanup_time`.
- **Performance:** Improved goroutine lifecycle management for better optimization and memory usage, following best practices.

### Fixes

- **Build:** Corrected a build failure caused by a breaking change in the `systray` dependency.

## [1.0.0] - 2025-09-21

### Added

- Voice-to-text functionality with offline Whisper.cpp
- System tray integration
- Global hotkey support
- Multiple output modes (clipboard, typing)
- AppImage packaging
- Cross-platform Linux desktop environment compatibility