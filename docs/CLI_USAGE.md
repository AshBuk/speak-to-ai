# CLI Usage Guide

Speak-to-AI provides a dual-mode binary that works as both a background daemon and a command-line interface. This is useful for custom scripts and tiling WMs users.

## Quick Start

**Launch the daemon** (run once, keeps running in background):
```bash
speak-to-ai                          # Launch daemon in background
```

**Use CLI commands** (while daemon is running):
```bash
speak-to-ai start                    # Begin recording
speak-to-ai stop                     # Stop and show transcript
speak-to-ai toggle                   # Toggle recording (start/stop with one command)
speak-to-ai status                   # Show state and configuration
speak-to-ai transcript               # Show last transcript

# Whisper model selection
speak-to-ai model list               # List available models (works without daemon)
speak-to-ai model set <model-id>     # Switch whisper model (requires daemon)
speak-to-ai model set base-q5_1      # ~57 MB, fast
speak-to-ai model set small-q5_1     # ~181 MB, default
speak-to-ai model set medium-q5_0    # ~539 MB
speak-to-ai model set large-v3-q5_0  # ~1.1 GB, best quality
speak-to-ai model delete <model-id>  # Delete a downloaded model (cannot delete active)
```

**Notes:**
- `toggle` is ideal for binding to a single DE shortcut — one key to start and stop recording
- Transcript is printed to stdout
- If using `active_window` output mode, text is also typed into the active window
- To suppress duplicate output: `speak-to-ai stop >/dev/null`

---

## CLI Flags

### `--socket <path>`
Specify custom IPC socket path.

```bash
speak-to-ai --socket /tmp/custom.sock start     # Custom socket path
```

**Default:** `$XDG_RUNTIME_DIR/speak-to-ai.sock`

---

### `--json`
Output responses in JSON format for scripting.

```bash
speak-to-ai --json status                       # JSON output
```

---

### `--timeout <seconds>`
Override default timeout for the command.

```bash
speak-to-ai --timeout 120 stop                  # 120 second timeout
```

**Default timeouts:**
- `stop`, `toggle`: 60 seconds (transcription can take time)
- Other commands: 5 seconds

---

## Daemon Flags

When running as daemon (without CLI command):

### `--config <path>`
Specify custom configuration file path.

```bash
speak-to-ai --config ~/.config/speak-to-ai/custom.yaml    # Custom config
```

**Default path:** `~/.config/speak-to-ai/config.yaml`

---

### `--debug`
Enable debug logging.

```bash
speak-to-ai --debug                             # Debug mode
```

---
