# Changelog

All notable changes to this project will be documented in this file.

## [1.3.4] - 2025-12-07

### Security
- **gosec integration:** SAST security scanning with gosec for continuous security monitoring
- **CI** scans on every PR with SARIF reports uploaded to GitHub Security tab
- **Zero vulnerabilities:** Current scan shows 0 security issues across 10,241 lines of code

---
## [1.3.3] - 2025-11-27

### Maintainability
- Improving long-term project maintainability: Key design components are now clarified so developers can understand the flow more quickly.
- Includes: comments, formatting, code style, and light stylistic refactoring.
- Excludes: any functional changes

Based on #55.

---
## [1.3.2] - 2025-11-13

### Maintainability
- **AppImageHub compatibility:** Renamed the AppImage file to follow the standard nomenclature (`speak-to-ai-VERSION-ARCH.AppImage`) for better catalog integration.
- **Reduced cyclomatic complexity:** Refactored several modules (validators, factories, providers, recorders) to improve code readability and maintainability.
- **Enhanced linting:** Upgraded `golangci-lint` to v2.6.1 and enabled the `gocyclo` linter for stricter complexity checks.

---
## [1.3.1] - 2025-10-31

### Dependencies
- **Go 1.24.1:** Updated Go dependencies to latest compatible versions
- **Bump Whisper.cpp to v1.8.2:** Improved performance and stability

### DevOps
- **Simplified Docker workflow:** Docker-first approach with streamlined commands for easy onboarding
- **Simplified and clearer Docker-based make targets**

---
## [1.3.0] - 2025-10-29

### Features

- **Dual-Mode Architecture:** Single binary now supports both daemon mode (background service with system tray) and CLI mode (command-line interface for scripting)
- **CLI Commands:** See [CLI Usage Guide](docs/CLI_USAGE.md)
- **Tiling WM Support:** (i3, sway, bspwm, etc.) through CLI hotkey bindings
- **IPC Communication:** CLI commands communicate with daemon via Unix socket for low-latency operations
- **JSON Output:** `--json` flag for machine-readable responses, perfect for automation and scripts
- **High-Resolution Icon:** Upgraded tray icon to 500x500 resolution for crisp display on HiDPI screens

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