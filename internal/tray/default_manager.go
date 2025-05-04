package tray

// CreateDefaultTrayManager creates the default tray manager
// based on available dependencies
func CreateDefaultTrayManager(onExit func(), onToggle func() error) TrayManagerInterface {
	// Use the mock implementation when systray is unavailable
	return NewMockTrayManager(onExit, onToggle)
}
