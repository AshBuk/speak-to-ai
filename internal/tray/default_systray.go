//go:build systray
// +build systray

package tray

// CreateDefaultTrayManager creates the default tray manager
// based on available dependencies
func CreateDefaultTrayManager(onExit func(), onToggle func() error) TrayManagerInterface {
	// Try to create real tray manager with proper icon validation
	if iconMicOff := GetIconMicOff(); len(iconMicOff) > 0 {
		if iconMicOn := GetIconMicOn(); len(iconMicOn) > 0 {
			return NewTrayManager(iconMicOff, iconMicOn, onExit, onToggle)
		}
	}

	// Fallback to mock if icons are not available
	return NewMockTrayManager(onExit, onToggle)
}
