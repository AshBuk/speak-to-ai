package platform

import (
	"os"
	"os/exec"
)

// EnvironmentType represents the display server type
type EnvironmentType string

const (
	// EnvironmentX11 represents X11 display server
	EnvironmentX11 EnvironmentType = "X11"
	// EnvironmentWayland represents Wayland display server
	EnvironmentWayland EnvironmentType = "Wayland"
	// EnvironmentUnknown represents unknown display server
	EnvironmentUnknown EnvironmentType = "Unknown"
)

// DetectEnvironment detects the current display server environment
func DetectEnvironment() EnvironmentType {
	// Check for Wayland
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return EnvironmentWayland
	}

	// Check for X11
	if os.Getenv("DISPLAY") != "" {
		return EnvironmentX11
	}

	// If neither is detected, assume unknown
	return EnvironmentUnknown
}

// UtilityExists checks if a specified command/utility exists in PATH
func UtilityExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// CheckPrivileges checks if the application has the necessary privileges
// to access input devices
func CheckPrivileges() bool {
	// Implementation depends on platform-specific details
	// This is a simple placeholder
	return os.Geteuid() == 0
}

// EnsureDirectoryExists creates a directory if it doesn't exist
func EnsureDirectoryExists(path string) error {
	return os.MkdirAll(path, 0755)
}
