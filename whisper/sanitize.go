// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package whisper

import (
	"regexp"
	"strings"
)

// sanitizeTranscript removes placeholder tokens and normalizes whitespace.
func sanitizeTranscript(input string) string {
	if input == "" {
		return ""
	}

	// Remove tokens like [BLANK_AUDIO], [NO_SPEECH], [MUSIC], etc.
	tokenPattern := regexp.MustCompile(`(?i)\[[a-z0-9_\-]+\]`)
	cleaned := tokenPattern.ReplaceAllString(input, " ")

	// Collapse multiple spaces
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	// Trim
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

// SanitizeTranscript provides a public API for transcript sanitization.
// It removes placeholder tokens and normalizes whitespace.
func SanitizeTranscript(input string) string { // exported for use by app layer
	return sanitizeTranscript(input)
}
