// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package constants

import (
	"strings"
	"testing"
)

func TestUIConstants(t *testing.T) {
	t.Run("Icons", func(t *testing.T) {
		// Test that all icon constants are non-empty
		iconTests := []struct {
			name  string
			value string
		}{
			{"IconReady", IconReady},
			{"IconError", IconError},
			{"IconRecording", IconRecording},
			{"IconProcessing", IconProcessing},
			{"IconWarning", IconWarning},
			{"IconDownload", IconDownload},
			{"IconInfo", IconInfo},
		}

		for _, test := range iconTests {
			if test.value == "" {
				t.Errorf("%s should not be empty", test.name)
			}
			if len(test.value) == 0 {
				t.Errorf("%s should have content", test.name)
			}
		}
	})

	t.Run("Messages", func(t *testing.T) {
		// Test that all message constants are non-empty and meaningful
		messageTests := []struct {
			name  string
			value string
		}{
			{"MsgReady", MsgReady},
			{"MsgRecording", MsgRecording},
			{"MsgTranscribing", MsgTranscribing},
			{"MsgModelUnavailable", MsgModelUnavailable},
			{"MsgRecorderUnavailable", MsgRecorderUnavailable},
			{"MsgTranscriptionFailed", MsgTranscriptionFailed},
			{"MsgTranscriptionEmpty", MsgTranscriptionEmpty},
			{"MsgModelSwitchFailed", MsgModelSwitchFailed},
			{"MsgTranscriptionCancelled", MsgTranscriptionCancelled},
		}

		for _, test := range messageTests {
			if test.value == "" {
				t.Errorf("%s should not be empty", test.name)
			}
			// Messages should be reasonable length (not too short or too long)
			if len(test.value) < 3 {
				t.Errorf("%s is too short: %s", test.name, test.value)
			}
			if len(test.value) > 100 {
				t.Errorf("%s is too long: %s", test.name, test.value)
			}
		}
	})

	t.Run("NotificationTitles", func(t *testing.T) {
		// Test notification title constants
		titleTests := []struct {
			name  string
			value string
		}{
			{"NotifyError", NotifyError},
			{"NotifySuccess", NotifySuccess},
			{"NotifyNoSpeech", NotifyNoSpeech},
			{"NotifyCancelled", NotifyCancelled},
			{"NotifyClipboard", NotifyClipboard},
			{"NotifyOutputFail", NotifyOutputFail},
		}

		for _, test := range titleTests {
			if test.value == "" {
				t.Errorf("%s should not be empty", test.name)
			}
			// Titles should be concise
			if len(test.value) < 2 {
				t.Errorf("%s is too short: %s", test.name, test.value)
			}
			if len(test.value) > 50 {
				t.Errorf("%s is too long for a title: %s", test.name, test.value)
			}
		}
	})

	t.Run("NotificationMessages", func(t *testing.T) {
		// Test notification message constants
		messageTests := []struct {
			name  string
			value string
		}{
			{"NotifyTypingFallback", NotifyTypingFallback},
			{"NotifyOutputBothFailed", NotifyOutputBothFailed},
			{"NotifyClipboardFallback", NotifyClipboardFallback},
			{"NotifyTranscriptionCancelled", NotifyTranscriptionCancelled},
		}

		for _, test := range messageTests {
			if test.value == "" {
				t.Errorf("%s should not be empty", test.name)
			}
			// Notification messages should be informative but not too long
			if len(test.value) < 10 {
				t.Errorf("%s is too short: %s", test.name, test.value)
			}
			if len(test.value) > 200 {
				t.Errorf("%s is too long for a notification: %s", test.name, test.value)
			}
		}
	})
}

func TestConstantUniqueness(t *testing.T) {
	// Test that different types of constants don't have conflicting values
	// This helps prevent confusion in the UI

	allConstants := map[string]string{
		"IconReady":      IconReady,
		"IconError":      IconError,
		"IconRecording":  IconRecording,
		"IconProcessing": IconProcessing,
		"IconWarning":    IconWarning,
		"IconDownload":   IconDownload,
		"IconInfo":       IconInfo,
	}

	// Icons should be unique (different emojis)
	seen := make(map[string]string)
	for name, value := range allConstants {
		if existing, exists := seen[value]; exists {
			t.Errorf("Duplicate constant value %s found in %s and %s", value, name, existing)
		}
		seen[value] = name
	}
}

func TestConstantConsistency(t *testing.T) {
	// Test logical consistency between related constants

	t.Run("RecordingMessages", func(t *testing.T) {
		// Recording message should contain "Recording"
		if !contains(MsgRecording, "Recording") {
			t.Errorf("MsgRecording should contain 'Recording', got: %s", MsgRecording)
		}
	})

	t.Run("TranscribingMessages", func(t *testing.T) {
		// Transcribing message should contain "Transcrib"
		if !contains(MsgTranscribing, "Transcrib") {
			t.Errorf("MsgTranscribing should contain 'Transcrib', got: %s", MsgTranscribing)
		}
	})

	t.Run("ErrorMessages", func(t *testing.T) {
		// Error-related messages should be descriptive
		errorMessages := []string{
			MsgModelUnavailable,
			MsgRecorderUnavailable,
			MsgTranscriptionFailed,
		}

		for i, msg := range errorMessages {
			if len(msg) < 5 {
				t.Errorf("Error message %d is too short: %s", i, msg)
			}
		}
	})
}

// Helper function to check if string contains substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			strings.Contains(strings.ToLower(s), strings.ToLower(substr)))
}
