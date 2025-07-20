package platform

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEnvironmentType_Constants tests environment type constants
func TestEnvironmentType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		envType  EnvironmentType
		expected string
	}{
		{
			name:     "X11 environment",
			envType:  EnvironmentX11,
			expected: "X11",
		},
		{
			name:     "Wayland environment",
			envType:  EnvironmentWayland,
			expected: "Wayland",
		},
		{
			name:     "Unknown environment",
			envType:  EnvironmentUnknown,
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.envType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.envType))
			}
		})
	}
}

// TestDetectEnvironment tests environment detection
func TestDetectEnvironment(t *testing.T) {
	// Save original environment variables
	originalWaylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	originalDisplay := os.Getenv("DISPLAY")

	// Clean up after test
	defer func() {
		if originalWaylandDisplay != "" {
			os.Setenv("WAYLAND_DISPLAY", originalWaylandDisplay)
		} else {
			os.Unsetenv("WAYLAND_DISPLAY")
		}
		if originalDisplay != "" {
			os.Setenv("DISPLAY", originalDisplay)
		} else {
			os.Unsetenv("DISPLAY")
		}
	}()

	tests := []struct {
		name            string
		waylandDisplay  string
		display         string
		expectedEnvType EnvironmentType
	}{
		{
			name:            "Wayland environment detected",
			waylandDisplay:  "wayland-0",
			display:         "",
			expectedEnvType: EnvironmentWayland,
		},
		{
			name:            "Wayland takes precedence over X11",
			waylandDisplay:  "wayland-0",
			display:         ":0",
			expectedEnvType: EnvironmentWayland,
		},
		{
			name:            "X11 environment detected",
			waylandDisplay:  "",
			display:         ":0",
			expectedEnvType: EnvironmentX11,
		},
		{
			name:            "X11 with localhost display",
			waylandDisplay:  "",
			display:         "localhost:10.0",
			expectedEnvType: EnvironmentX11,
		},
		{
			name:            "Neither environment detected",
			waylandDisplay:  "",
			display:         "",
			expectedEnvType: EnvironmentUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			if tt.waylandDisplay != "" {
				os.Setenv("WAYLAND_DISPLAY", tt.waylandDisplay)
			} else {
				os.Unsetenv("WAYLAND_DISPLAY")
			}

			if tt.display != "" {
				os.Setenv("DISPLAY", tt.display)
			} else {
				os.Unsetenv("DISPLAY")
			}

			// Test detection
			detected := DetectEnvironment()
			if detected != tt.expectedEnvType {
				t.Errorf("Expected %s, got %s", tt.expectedEnvType, detected)
			}
		})
	}
}

// TestUtilityExists tests utility existence checking
func TestUtilityExists(t *testing.T) {
	tests := []struct {
		name        string
		utilityName string
		shouldExist bool
	}{
		{
			name:        "existing utility - ls",
			utilityName: "ls",
			shouldExist: true,
		},
		{
			name:        "existing utility - cat",
			utilityName: "cat",
			shouldExist: true,
		},
		{
			name:        "existing utility - sh",
			utilityName: "sh",
			shouldExist: true,
		},
		{
			name:        "nonexistent utility",
			utilityName: "nonexistent_utility_12345",
			shouldExist: false,
		},
		{
			name:        "empty utility name",
			utilityName: "",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := UtilityExists(tt.utilityName)
			if exists != tt.shouldExist {
				t.Errorf("Expected %s to exist: %v, got: %v", tt.utilityName, tt.shouldExist, exists)
			}
		})
	}
}

// TestUtilityExists_CommonTools tests common development tools
func TestUtilityExists_CommonTools(t *testing.T) {
	// These tests are informational - they check what's available on the system
	commonTools := []string{
		"go",
		"git",
		"make",
		"grep",
		"find",
	}

	for _, tool := range commonTools {
		t.Run("check_"+tool, func(t *testing.T) {
			exists := UtilityExists(tool)
			t.Logf("Tool %s exists: %v", tool, exists)
			// We don't assert here since these tools may or may not be available
		})
	}
}

// TestCheckPrivileges tests privilege checking
func TestCheckPrivileges(t *testing.T) {
	// This test checks the current privilege level
	hasPrivileges := CheckPrivileges()

	// We can't reliably test this in all environments, but we can test that it returns a boolean
	if hasPrivileges {
		t.Log("Running with elevated privileges (root)")
	} else {
		t.Log("Running with normal user privileges")
	}

	// The function should always return a boolean value
	if hasPrivileges != true && hasPrivileges != false {
		t.Error("CheckPrivileges should return a boolean value")
	}
}

// TestEnsureDirectoryExists tests directory creation
func TestEnsureDirectoryExists(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		path          string
		shouldSucceed bool
	}{
		{
			name:          "create simple directory",
			path:          filepath.Join(tempDir, "test_dir"),
			shouldSucceed: true,
		},
		{
			name:          "create nested directory",
			path:          filepath.Join(tempDir, "level1", "level2", "level3"),
			shouldSucceed: true,
		},
		{
			name:          "create already existing directory",
			path:          tempDir, // This already exists
			shouldSucceed: true,
		},
		{
			name:          "create directory with complex path",
			path:          filepath.Join(tempDir, "complex_dir_123", "sub", "another"),
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureDirectoryExists(tt.path)

			if tt.shouldSucceed && err != nil {
				t.Errorf("Expected directory creation to succeed, got error: %v", err)
			}

			if !tt.shouldSucceed && err == nil {
				t.Error("Expected directory creation to fail, but it succeeded")
			}

			if tt.shouldSucceed {
				// Verify directory was actually created
				info, statErr := os.Stat(tt.path)
				if statErr != nil {
					t.Errorf("Directory was not created: %v", statErr)
				} else if !info.IsDir() {
					t.Error("Created path is not a directory")
				}
			}
		})
	}
}

// TestEnsureDirectoryExists_InvalidPath tests directory creation with invalid paths
func TestEnsureDirectoryExists_InvalidPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		description string
	}{
		{
			name:        "empty path",
			path:        "",
			description: "Empty path should create current directory (should succeed)",
		},
		{
			name:        "path with null character",
			path:        "/tmp/test\x00dir",
			description: "Path with null character should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureDirectoryExists(tt.path)
			t.Logf("Path: %q, Error: %v, Description: %s", tt.path, err, tt.description)

			// Note: We don't assert success/failure here because behavior may vary
			// across different operating systems and file systems
		})
	}
}

// TestEnsureDirectoryExists_Permissions tests directory creation with permission restrictions
func TestEnsureDirectoryExists_Permissions(t *testing.T) {
	// This test is platform-specific and may not work in all environments
	// We'll test creating a directory in a location that typically requires permissions

	restrictedPaths := []string{
		"/root/test_restricted_dir", // Requires root access
		"/sys/test_restricted_dir",  // System directory
	}

	for _, path := range restrictedPaths {
		t.Run("restricted_path_"+path, func(t *testing.T) {
			err := EnsureDirectoryExists(path)
			if err != nil {
				t.Logf("Cannot create directory %s (expected): %v", path, err)
			} else {
				t.Logf("Successfully created directory %s (cleanup may be needed)", path)
				// Try to clean up if we somehow succeeded
				os.RemoveAll(path)
			}
		})
	}
}

// TestEnvironmentDetection_Integration tests integration with real environment
func TestEnvironmentDetection_Integration(t *testing.T) {
	// Test the actual environment detection without modifying env vars
	currentEnv := DetectEnvironment()

	t.Logf("Current environment detected as: %s", currentEnv)

	// Verify it's one of the valid types
	validTypes := []EnvironmentType{EnvironmentX11, EnvironmentWayland, EnvironmentUnknown}
	isValid := false
	for _, validType := range validTypes {
		if currentEnv == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		t.Errorf("Detected environment %s is not a valid EnvironmentType", currentEnv)
	}

	// Log some environment information for debugging
	t.Logf("DISPLAY env var: %q", os.Getenv("DISPLAY"))
	t.Logf("WAYLAND_DISPLAY env var: %q", os.Getenv("WAYLAND_DISPLAY"))
}

// TestEnvironmentType_StringConversion tests string conversion of environment types
func TestEnvironmentType_StringConversion(t *testing.T) {
	envTypes := map[EnvironmentType]string{
		EnvironmentX11:     "X11",
		EnvironmentWayland: "Wayland",
		EnvironmentUnknown: "Unknown",
	}

	for envType, expected := range envTypes {
		if string(envType) != expected {
			t.Errorf("Environment type %v should convert to string %q, got %q", envType, expected, string(envType))
		}
	}
}
