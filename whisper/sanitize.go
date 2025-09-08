//go:build cgo

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
	tokenPattern := regexp.MustCompile(`\[[A-Z_]+\]`)
	cleaned := tokenPattern.ReplaceAllString(input, " ")

	// Collapse multiple spaces
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	// Trim
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}
