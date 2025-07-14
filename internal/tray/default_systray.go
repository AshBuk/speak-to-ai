//go:build systray

package tray

import "github.com/AshBuk/speak-to-ai/config"

// CreateDefaultTrayManager creates the default tray manager
// based on available dependencies
func CreateDefaultTrayManager(onExit func(), onToggle func() error, onShowConfig func() error, onReloadConfig func() error) TrayManagerInterface {
	// Use the real systray implementation
	iconMicOff := GetIconMicOff()
	iconMicOn := GetIconMicOn()

	return NewTrayManager(iconMicOff, iconMicOn, onExit, onToggle, onShowConfig, onReloadConfig)
}

// CreateTrayManagerWithConfig creates tray manager with initial configuration
func CreateTrayManagerWithConfig(config *config.Config, onExit func(), onToggle func() error, onShowConfig func() error, onReloadConfig func() error) TrayManagerInterface {
	trayManager := CreateDefaultTrayManager(onExit, onToggle, onShowConfig, onReloadConfig)
	trayManager.UpdateSettings(config)
	return trayManager
}
