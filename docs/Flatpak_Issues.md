# Flatpak Issues Report

**Date**: 2025-09-21
**Tested Environment**: Fedora 42, GNOME/Wayland
**Status**: Beta

## Critical Issues:

### 1. System Tray Broken
```
systray error: failed to request name: org.freedesktop.DBus.Error.ServiceUnknown
```
- Added `--own-name=org.kde.StatusNotifierItem.*` but still fails
- fyne.io/systray cannot register in sandbox
- No tray icon appears

### 2. ydotool Not Working
```
Active window method failed: typing to active window not supported by clipboard outputter
```
- Permission `--filesystem=/tmp/.ydotool_socket:rw` added
- App cannot connect to ydotool socket
- Falls back to clipboard mode
- Host ydotool setup verified working (AppImage works fine)

### 3. Performance Issues
- Transcription much slower than AppImage
- Same model, same hardware
- AppImage: ~0.5-3 sec.
- Flatpak: ~4-10 sec.

### 4. About Page Fails
- About dialog doesn't open
- Temp file access restricted in sandbox

### 5. Lock File Problems
- Lock file persists between runs
- Shows "Another instance running"
- Need manual cleanup: `flatpak run --command=sh io.github.ashbuk.speak-to-ai -c "rm -f /run/user/1000/speak-to-ai.lock"`

## What Works
- ✅ Model loading (`/app/share/speak-to-ai/models/small-q5_1.bin`)
- ✅ Audio recording
- ✅ Speech recognition (but slow)
- ✅ Hotkeys via D-Bus portal
- ✅ Clipboard output
- ✅ Config persistence

## Code Changes Made
1. **Model path fix**: Added Flatpak detection in `whisper/providers/model_path_resolver.go`
2. **DBus permissions**: Added StatusNotifierItem own-name permissions

*Flatpak needs significant sandbox work for this type of functionality*