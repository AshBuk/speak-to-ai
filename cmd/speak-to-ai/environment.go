// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// Check for an AppImage environment and, if detected,
// modify the config file path to use the bundled configuration if available.
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

	if appDir == "" {
		logger.Warning("Running in AppImage but could not detect AppDir")
		return configPath
	}

	logger.Info("Running inside AppImage, base path: %s", appDir)

	if configPath == "config.yaml" {
		bundledConfig := filepath.Join(appDir, "config.yaml")
		if _, err := os.Stat(bundledConfig); err == nil {
			logger.Info("Using AppImage bundled config: %s", bundledConfig)
			return bundledConfig
		}
	}

	return configPath
}

// Check for a Flatpak environment and, if detected,
// modify the config file path to use the standard Flatpak configuration path.
func adjustPathsForFlatpak(logger logger.Logger, configPath string) string {
	flatpakInfo := os.Getenv("FLATPAK_ID")
	if flatpakInfo == "" {
		return configPath
	}

	logger.Info("Running inside Flatpak: %s", flatpakInfo)

	if configPath == "config.yaml" {
		flatpakConfigPath := "/app/share/speak-to-ai/config.yaml"
		if _, err := os.Stat(flatpakConfigPath); err == nil {
			logger.Info("Using Flatpak config: %s", flatpakConfigPath)
			return flatpakConfigPath
		}
	}

	return configPath
}
