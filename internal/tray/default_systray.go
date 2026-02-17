//go:build systray

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package tray

import (
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// CreateDefaultTrayManager creates the default tray manager
// based on available dependencies.
func CreateDefaultTrayManager(logger logger.Logger) TrayManagerInterface {
	// Use the real systray implementation
	iconMicOff := GetIconMicOff(logger)
	iconMicOn := GetIconMicOn(logger)

	return NewTrayManager(iconMicOff, iconMicOn, logger)
}

// CreateTrayManagerWithConfig creates tray manager with initial configuration.
func CreateTrayManagerWithConfig(config *config.Config, logger logger.Logger) TrayManagerInterface {
	trayManager := CreateDefaultTrayManager(logger)
	trayManager.UpdateSettings(config)
	return trayManager
}
