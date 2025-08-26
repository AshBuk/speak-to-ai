# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.1.2] - 2025-08-26

### Fixed
- AppImage packaging issues

### Changed
- Improved model type detection with switch statement

### Added
- GitHub issue templates for bug reports and feature requests
- GitHub labels configuration for project organization
- Pull request template for standardized contributions

## [0.1.1] - 2025-08-25

### Fixed
- AppImage system tray library bundling
- Flatpak shared library symlinks
- Whisper libraries bundling in packages

### Changed
- Unified artifact names in CI/CD pipeline

## [0.1.0] - 2025-08-19

### Added
- Initial public release
- Local speech-to-text using Whisper.cpp
- System tray integration
- Cross-platform support (X11/Wayland)
- Configurable hotkeys (AltGr + ,)
- Multiple output modes (typing/clipboard/combined)
- AppImage and Flatpak packages
- WebSocket API for integrations
- Comprehensive test suite and CI/CD

---

## How to Maintain This Changelog

### When to Update
- **Before each release** - move items from Unreleased to new version
- **After significant changes** - add items to Unreleased section
- **Never retroactively** - don't change past versions

### Format
```markdown
## [Version] - YYYY-MM-DD

### Added
- New features

### Changed  
- Changes in existing functionality

### Fixed
- Bug fixes

### Removed
- Removed features
```
