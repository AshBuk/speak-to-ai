// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"strings"
	"testing"
)

func TestSanitizeTranscript(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "No tokens to remove",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "Single token removal",
			input:    "Hello [MUSIC] world",
			expected: "Hello world",
		},
		{
			name:     "Multiple tokens removal",
			input:    "Hello [MUSIC] world [NOISE] how are you [SILENCE]",
			expected: "Hello world how are you",
		},
		{
			name:     "Case insensitive token removal",
			input:    "Hello [music] world [NOISE] test [Silence]",
			expected: "Hello world test",
		},
		{
			name:     "Tokens with numbers and underscores",
			input:    "Hello [music_1] world [NOISE_2] test [silence_end]",
			expected: "Hello world test",
		},
		{
			name:     "Tokens with hyphens",
			input:    "Hello [background-music] world [non-speech] test",
			expected: "Hello world test",
		},
		{
			name:     "Multiple whitespace normalization",
			input:    "Hello    world   how  are   you",
			expected: "Hello world how are you",
		},
		{
			name:     "Leading and trailing whitespace",
			input:    "   Hello world   ",
			expected: "Hello world",
		},
		{
			name:     "Tokens with extra whitespace",
			input:    "Hello  [MUSIC]   world   [NOISE]  test",
			expected: "Hello world test",
		},
		{
			name:     "Only tokens",
			input:    "[MUSIC] [NOISE] [SILENCE]",
			expected: "",
		},
		{
			name:     "Mixed whitespace types",
			input:    "Hello\t[MUSIC]\nworld\r\n[NOISE]\ttest",
			expected: "Hello world test",
		},
		{
			name:     "Consecutive tokens",
			input:    "Hello [MUSIC][NOISE] world",
			expected: "Hello world",
		},
		{
			name:     "Token at beginning",
			input:    "[START] Hello world",
			expected: "Hello world",
		},
		{
			name:     "Token at end",
			input:    "Hello world [END]",
			expected: "Hello world",
		},
		{
			name:     "Complex real-world example",
			input:    " [MUSIC]  Hello   there [NOISE] how are you doing [SILENCE] today?  [BACKGROUND]  ",
			expected: "Hello there how are you doing today?",
		},
		{
			name:     "Invalid brackets (tokens with numbers at start are still removed)",
			input:    "Hello [123ABC] world [_invalid] test",
			expected: "Hello world test",
		},
		{
			name:     "Partial brackets",
			input:    "Hello [MUSIC world NOISE] test",
			expected: "Hello [MUSIC world NOISE] test",
		},
		{
			name:     "Nested brackets",
			input:    "Hello [[MUSIC]] world",
			expected: "Hello [ ] world",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeTranscript(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestSanitizeTranscript_Performance(t *testing.T) {
	// Test with a large input to ensure reasonable performance
	largeInput := strings.Repeat("Hello [MUSIC] world [NOISE] ", 1000)

	result := SanitizeTranscript(largeInput)

	// Should not panic and should produce expected result
	expectedPattern := strings.Repeat("Hello world ", 1000)
	expectedPattern = strings.TrimSpace(expectedPattern)

	if result != expectedPattern {
		t.Errorf("Performance test failed: result length %d, expected length %d",
			len(result), len(expectedPattern))
	}
}

func TestSanitizeTranscript_SecurityEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		desc     string
	}{
		{
			name:     "Very long token",
			input:    "Hello [" + strings.Repeat("A", 100) + "] world",
			expected: "Hello world",
			desc:     "Should handle very long tokens",
		},
		{
			name:     "Many consecutive tokens",
			input:    "Hello " + strings.Repeat("[TOKEN] ", 50) + "world",
			expected: "Hello world",
			desc:     "Should handle many consecutive tokens",
		},
		{
			name:     "Unicode in content (not in tokens)",
			input:    "Привет [MUSIC] мир",
			expected: "Привет мир",
			desc:     "Should preserve Unicode content",
		},
		{
			name:     "Special regex characters in content",
			input:    "Hello $^.*+?{}|()[] [MUSIC] world",
			expected: "Hello $^.*+?{}|()[] world",
			desc:     "Should not break on regex special characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeTranscript(tc.input)
			if result != tc.expected {
				t.Errorf("%s: Expected %q, got %q", tc.desc, tc.expected, result)
			}
		})
	}
}

// Benchmark to ensure reasonable performance
func BenchmarkSanitizeTranscript(b *testing.B) {
	input := "Hello [MUSIC] world [NOISE] how are you [SILENCE] today?"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeTranscript(input)
	}
}

func BenchmarkSanitizeTranscript_LargeInput(b *testing.B) {
	input := strings.Repeat("Hello [MUSIC] world [NOISE] ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeTranscript(input)
	}
}
