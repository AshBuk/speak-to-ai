// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// Get default config path using XDG Base Directory specification
func getDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.yaml" // fallback to current directory
	}
	return filepath.Join(home, ".config", "speak-to-ai", "config.yaml")
}

// Check for an AppImage environment and log detection.
// AppImage always uses XDG config path since AppImage filesystem is read-only.
func adjustPathsForAppImage(logger logger.Logger, configPath string) string {
	appImagePath := os.Getenv("APPIMAGE")
	if appImagePath == "" {
		return configPath
	}
	appDir := os.Getenv("APPDIR")
	if appDir == "" {
		argv0 := os.Getenv("ARGV0")
		if argv0 != "" && strings.HasSuffix(argv0, "/AppRun") {
			appDir = filepath.Dir(argv0)
		}
	}
	if appDir != "" {
		logger.Info("Running inside AppImage, base path: %s", appDir)
	}
	// AppImage filesystem is read-only (squashfs), always use XDG path for config
	// This ensures config changes (rebind, language, etc.) can be saved
	return getDefaultConfigPath()
}
