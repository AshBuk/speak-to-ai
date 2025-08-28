// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

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

// DetectDesktopEnvironment detects the current desktop environment
func DetectDesktopEnvironment() string {
	// Check XDG_CURRENT_DESKTOP first (most reliable)
	if de := os.Getenv("XDG_CURRENT_DESKTOP"); de != "" {
		return de
	}

	// Fallback to legacy variables
	if de := os.Getenv("DESKTOP_SESSION"); de != "" {
		return de
	}

	return "Unknown"
}

// IsGNOMEWithWayland checks if running GNOME with Wayland
func IsGNOMEWithWayland() bool {
	de := DetectDesktopEnvironment()
	env := DetectEnvironment()

	return (de == "GNOME" || de == "ubuntu:GNOME") && env == EnvironmentWayland
}

// CheckAppIndicatorExtension checks if GNOME AppIndicator extension is available
func CheckAppIndicatorExtension() bool {
	// Check if gnome-extensions command is available
	if !UtilityExists("gnome-extensions") {
		return false
	}

	// Check if appindicator extension is installed and enabled
	cmd := exec.Command("gnome-extensions", "list", "--enabled")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Look for common appindicator extension IDs
	extensions := []string{
		"appindicatorsupport@rgcjonas.gmail.com",
		"ubuntu-appindicators@ubuntu.com",
		"appindicator@ubuntu.com",
	}

	outputStr := string(output)
	for _, ext := range extensions {
		if contains(outputStr, ext) {
			return true
		}
	}

	return false
}

// contains is a simple substring check helper
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
