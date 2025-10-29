# CLI Usage Guide

Speak-to-AI provides a dual-mode binary that works as both a background daemon and a command-line interface. This is useful for custom scripts and tiling WMs users.

## Quick Start

**Launch the daemon** (run once, keeps running in background):
```bash
./speak-to-ai-*.AppImage              # Launch daemon in background
```

**Use CLI commands** (while daemon is running):
```bash
./speak-to-ai-*.AppImage start        # Begin recording
./speak-to-ai-*.AppImage stop         # Stop and show transcript
./speak-to-ai-*.AppImage status       # Show recording status
./speak-to-ai-*.AppImage transcript   # Show last transcript
```

**Notes:**
- Transcript is printed to stdout
- If using `active_window` output mode, text is also typed into the active window
- To suppress duplicate output: `./speak-to-ai-*.AppImage stop >/dev/null`

---

## CLI Flags

### `--socket <path>`
Specify custom IPC socket path.

```bash
./speak-to-ai-*.AppImage --socket /tmp/custom.sock start     # Custom socket path
```

**Default:** `$XDG_RUNTIME_DIR/speak-to-ai.sock`

---

### `--json`
Output responses in JSON format for scripting.

```bash
./speak-to-ai-*.AppImage --json status                       # JSON output
```

---

### `--timeout <seconds>`
Override default timeout for the command.

```bash
./speak-to-ai-*.AppImage --timeout 120 stop                  # 120 second timeout
```

**Default timeouts:**
- `stop`: 60 seconds (transcription can take time)
- Other commands: 5 seconds

---

## Daemon Flags

When running as daemon (without CLI command):

### `--config <path>`
Specify custom configuration file path.

```bash
./speak-to-ai-*.AppImage --config ~/.config/speak-to-ai/custom.yaml    # Custom config
```

**Default path:** `~/.config/speak-to-ai/config.yaml`

---

### `--debug`
Enable debug logging.

```bash
./speak-to-ai-*.AppImage --debug                             # Debug mode
```

---

