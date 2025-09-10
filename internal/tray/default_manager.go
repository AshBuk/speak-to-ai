//go:build !systray

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package tray

import "github.com/AshBuk/speak-to-ai/config"

// CreateDefaultTrayManager creates the default tray manager
// based on available dependencies
func CreateDefaultTrayManager(onExit func(), onToggle func() error, onShowConfig func() error, onResetToDefaults func() error) TrayManagerInterface {
	// Use the mock implementation as fallback when systray is not available
	return CreateMockTrayManager(onExit, onToggle, onShowConfig, onResetToDefaults)
}

// CreateTrayManagerWithConfig creates tray manager with initial configuration
func CreateTrayManagerWithConfig(config *config.Config, onExit func(), onToggle func() error, onShowConfig func() error, onResetToDefaults func() error) TrayManagerInterface {
	trayManager := CreateDefaultTrayManager(onExit, onToggle, onShowConfig, onResetToDefaults)
	trayManager.UpdateSettings(config)
	return trayManager
}
