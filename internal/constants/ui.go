// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package constants

// UI Icons used throughout the application
const (
	IconReady      = "✅"
	IconError      = "❌"
	IconRecording  = "🎤"
	IconProcessing = "🔄"
	IconWarning    = "⚠️"
	IconDownload   = "📥"
	IconInfo       = "ℹ️"
)

// Common UI Messages
const (
	MsgReady                  = "Ready"
	MsgRecording              = "Recording..."
	MsgTranscribing           = "Transcribing..."
	MsgModelUnavailable       = "Model unavailable"
	MsgRecorderUnavailable    = "Audio recorder unavailable"
	MsgTranscriptionFailed    = "Transcription failed"
	MsgTranscriptionEmpty     = "No speech detected in recording"
	MsgModelSwitchFailed      = "Model switch failed"
	MsgTranscriptionCancelled = "Transcription cancelled"
)

// Notification Titles
const (
	NotifyError      = "Error"
	NotifySuccess    = "Success"
	NotifyNoSpeech   = "No Speech"
	NotifyCancelled  = "Cancelled"
	NotifyClipboard  = "Output via Clipboard"
	NotifyOutputFail = "Output Failed"
)

// Notification Messages
const (
	NotifyTypingFallback         = "Typing not supported by compositor. Text copied to clipboard - press Ctrl+V to paste."
	NotifyOutputBothFailed       = "both typing and clipboard failed, check output configuration"
	NotifyClipboardFallback      = "Text copied to clipboard - press Ctrl+V to paste."
	NotifyTranscriptionCancelled = "Transcription was cancelled"
)
