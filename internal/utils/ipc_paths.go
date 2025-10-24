// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"os"
	"path/filepath"
)

const (
	// DefaultSocketFileName is the default IPC socket filename.
	DefaultSocketFileName = "speak-to-ai.sock"
)

// GetDefaultSocketPath returns the default IPC socket path.
func GetDefaultSocketPath() string {
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		return filepath.Join(runtimeDir, DefaultSocketFileName)
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		socketDir := filepath.Join(homeDir, ".config", "speak-to-ai")
		if err := os.MkdirAll(socketDir, 0o700); err == nil {
			return filepath.Join(socketDir, DefaultSocketFileName)
		}
	}

	return filepath.Join(os.TempDir(), DefaultSocketFileName)
}
