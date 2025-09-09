// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"regexp"
	"strings"
)

// SanitizeTranscript removes placeholder tokens and normalizes whitespace.
// This package is broadly used across app and whisper layers.
func SanitizeTranscript(input string) string {
	if input == "" {
		return ""
	}

	tokenPattern := regexp.MustCompile(`(?i)\[[a-z0-9_\-]+\]`)
	cleaned := tokenPattern.ReplaceAllString(input, " ")

	cleaned = strings.Join(strings.Fields(cleaned), " ")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}
