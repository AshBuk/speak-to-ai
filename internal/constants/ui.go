// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package constants

// UI Icons used throughout the application
const (
	IconReady     = "✅"
	IconError     = "❌"
	IconRecording = "🔴"
	IconStop      = "🟥"
	IconConfig    = "↺"
	TraySettings  = "⚙️"
)

// Common UI Messages
const (
	MsgReady                  = "Ready"
	MsgRecording              = "Recording..."
	MsgTranscribing           = "Transcribing..."
	MsgTranscriptionComplete  = "Transcription complete"
	MsgModelUnavailable       = "Model unavailable"
	MsgRecorderUnavailable    = "Audio recorder unavailable"
	MsgTranscriptionFailed    = "Transcription failed"
	MsgTranscriptionEmpty     = "No speech was detected in the recording"
	MsgNoSpeechDetected       = "No speech detected"
	MsgModelSwitchFailed      = "Model switch failed"
	MsgTranscriptionCancelled = "Transcription cancelled"
)

// Notification Titles
const (
	NotifyError             = "Error"
	NotifySuccess           = "Success"
	NotifyNoSpeech          = "No Speech"
	NotifyConfigReset       = "Configuration Reset"
	NotifyRecordingStarted  = "Recording Started"
	NotifyRecordingStopped  = "Recording stopped"
	NotifyTranscriptionDone = "Transcription Complete"
	NotifyTranscriptionErr  = "Transcription Error"
)

// Notification Messages
const (
	NotifyTypingFallback         = "Typing not supported by compositor. Text copied to clipboard - press Ctrl+V to paste."
	NotifyOutputBothFailed       = "both typing and clipboard failed, check output configuration"
	NotifyClipboardFallback      = "Text copied to clipboard - press Ctrl+V to paste."
	NotifyTranscriptionCancelled = "Transcription was cancelled"
	NotifyConfigResetSuccess     = "Settings reset to defaults successfully"
	NotifyRecordingStartMsg      = "Recording started"
	NotifyRecordingStopMsg       = "Transcribing audio..."
	NotifyTranscriptionMsg       = "Text copied to clipboard"
	NotifyAppName                = "Speak-to-AI"
)

// Notification Full Titles (with icons)
const (
	NotifyTitleRecordingStart = IconRecording + " " + NotifyAppName
	NotifyTitleRecordingStop  = IconStop + " " + NotifyRecordingStopped
	NotifyTitleTranscription  = IconReady + " " + NotifyTranscriptionDone
	NotifyTitleError          = IconError + " " + NotifyError
	NotifyTitleConfigReset    = IconConfig + " " + NotifyConfigReset
)
