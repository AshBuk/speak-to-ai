// Package speaktoai provides a high-level overview of the Speak-to-AI project.
//
// Speak-to-AI is a minimalist, privacy-focused desktop daemon written in Go
// that converts speech to text offline using local Whisper models.
//
// Core responsibilities:
//   - Global hotkeys using DBus GlobalShortcuts portal (primary) or evdev (fallback)
//   - Audio recording via arecord/ffmpeg backends
//   - Local transcription using go-whisper (whisper.cpp)
//   - Text output routing: clipboard, active window typing, or combined
//   - X11 and Wayland support with smart tool selection (xdotool, wtype, ydotool)
//
// Packaging:
//   - AppImage and Flatpak packages with first-run configuration and model copy
//   - Flatpak uses minimal permissions and desktop portals where possible
//
// Testing strategy:
//   - Unit tests colocated with packages (default go test ./...)
//   - Integration tests in tests/integration (run with -tags=integration)
//
// For more details, see README.md, DEVELOPMENT.md, and docker/README.md.
package speaktoai
