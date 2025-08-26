// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"os"
	"path/filepath"
)

// IsValidFile checks if a file exists and is accessible
func IsValidFile(path string) bool {
	// Check for path traversal attempts
	clean := filepath.Clean(path)
	if clean != path {
		return false
	}

	// Check file existence and access
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
