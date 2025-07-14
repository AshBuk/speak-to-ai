//go:build !systray
// +build !systray

package tray

// CreateDefaultTrayManager creates the default tray manager
// based on available dependencies
func CreateDefaultTrayManager(onExit func(), onToggle func() error, onShowConfig func() error, onReloadConfig func() error) TrayManagerInterface {
	// Use the mock implementation as fallback when systray is not available
	return NewMockTrayManager(onExit, onToggle, onShowConfig, onReloadConfig)
}
