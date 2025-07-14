//go:build systray
// +build systray

package tray

// CreateDefaultTrayManager creates the default tray manager
// based on available dependencies
func CreateDefaultTrayManager(onExit func(), onToggle func() error, onShowConfig func() error, onReloadConfig func() error) TrayManagerInterface {
	// Use the real systray implementation
	iconMicOff := GetIconMicOff()
	iconMicOn := GetIconMicOn()

	return NewTrayManager(iconMicOff, iconMicOn, onExit, onToggle, onShowConfig, onReloadConfig)
}
