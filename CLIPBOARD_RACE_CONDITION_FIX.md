# Clipboard Race Condition Fix

## Problem

User stops recording → immediately presses Ctrl+V → gets **old clipboard content** instead of whisper transcription (which appears 1-2 seconds later).

## Solution

**Block clipboard reads until transcription completes.**

### Implementation

1. **AudioService**: On stop recording → set `transcriptionInProgress = true`
2. **IOService**: If `transcriptionInProgress` → Ctrl+V waits for whisper result
3. **AudioService**: After transcription → write to clipboard → set `transcriptionInProgress = false`

### Technical Details

- Add boolean flag `transcriptionInProgress` to IOService
- Add result channel for whisper output
- Modify clipboard getter to wait when protected
- Automatic cleanup when transcription completes
- 5-second timeout failsafe

## Result

Ctrl+V always returns correct transcribed text, regardless of timing. No race condition.