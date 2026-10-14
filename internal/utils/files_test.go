// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"os"
	"testing"
)

func TestIsValidFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_file")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "existing file",
			path:     tempFile.Name(),
			expected: true,
		},
		{
			name:     "non-existing file",
			path:     "/non/existing/file.txt",
			expected: false,
		},
		{
			name:     "path traversal attempt",
			path:     "../../../etc/passwd",
			expected: false,
		},
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidFile(tt.path)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetFileSize(t *testing.T) {
	// Create a temporary file with known content
	tempFile, err := os.CreateTemp("", "test_size")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := "Hello, World!"
	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tempFile.Close()

	tests := []struct {
		name         string
		path         string
		expectedSize int64
		expectError  bool
	}{
		{
			name:         "existing file",
			path:         tempFile.Name(),
			expectedSize: int64(len(content)),
			expectError:  false,
		},
		{
			name:         "non-existing file",
			path:         "/non/existing/file.txt",
			expectedSize: 0,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := GetFileSize(tt.path)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && size != tt.expectedSize {
				t.Errorf("expected size %d, got %d", tt.expectedSize, size)
			}
		})
	}
}
