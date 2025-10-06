# Changelog

All notable changes to this project will be documented in this file.

---

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