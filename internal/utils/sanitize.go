// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"regexp"
	"strings"
)

// Remove placeholder tokens and normalize whitespace.
// This package is broadly used across app and whisper layers
func SanitizeTranscript(input string) string {
	if input == "" {
		return ""
	}

	// Remove bracketed placeholders like [music], etc. (Unicode letters supported)
	tokenPattern := regexp.MustCompile(`(?i)\[[\p{L}0-9_\-]+\]`)
	cleaned := tokenPattern.ReplaceAllString(input, " ")

	cleaned = strings.Join(strings.Fields(cleaned), " ")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}
