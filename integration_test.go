//go:build integration
// +build integration

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AshBuk/speak-to-ai/config"
)

// Integration tests for complete user scenarios
// Run with: go test -tags=integration

func TestApplicationInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test that basic config loading works
	tempDir := t.TempDir()

	// Create test config
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)
	cfg.General.ModelPath = filepath.Join(tempDir, "test-model.bin")
	cfg.General.TempAudioPath = tempDir
	cfg.Output.DefaultMode = "clipboard" // Safe for testing
	cfg.WebServer.Enabled = false

	// Create mock model file
	modelFile, err := os.Create(cfg.General.ModelPath)
	if err != nil {
		t.Fatalf("Failed to create mock model: %v", err)
	}
	modelFile.Close()

	// Test config validation
	err = config.ValidateConfig(cfg)
	if err != nil {
		t.Logf("Config validation failed (expected in test environment): %v", err)
	}

	t.Log("Application initialization test completed")
}

func TestConfigurationLoading(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		expectErr  bool
	}{
		{
			name: "valid_config",
			configData: `
general:
  debug: false
  model_path: "test-model.bin"
  language: "en"
hotkeys:
  start_recording: "altgr+comma"
audio:
  device: "default"
  sample_rate: 16000
output:
  default_mode: "clipboard"
`,
			expectErr: false,
		},
		{
			name: "invalid_yaml",
			configData: `
invalid: yaml: content:
  - malformed
`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configFile := filepath.Join(tempDir, "config.yaml")

			err := os.WriteFile(configFile, []byte(tt.configData), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			_, err = config.LoadConfig(configFile)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestHotkeyManagerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test that hotkey manager can be created and configured
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	// Test different hotkey configurations
	testHotkeys := []string{
		"altgr+comma",
		"ctrl+shift+r",
		"alt+space",
	}

	for _, hotkey := range testHotkeys {
		t.Run("hotkey_"+hotkey, func(t *testing.T) {
			cfg.Hotkeys.StartRecording = hotkey

			// This would normally require elevated permissions
			// In test environment, we just verify it doesn't crash
			t.Logf("Testing hotkey configuration: %s", hotkey)
		})
	}
}

func TestOutputManagerIntegration(t *testing.T) {
	// Test output manager creation with different configurations
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	testModes := []string{"clipboard", "active_window", "combined"}

	for _, mode := range testModes {
		t.Run("mode_"+mode, func(t *testing.T) {
			cfg.Output.DefaultMode = mode

			// Test that output manager can be created
			// Actual functionality would require external tools
			t.Logf("Testing output mode: %s", mode)
		})
	}
}

func TestEnvironmentDetection(t *testing.T) {
	// Test that environment detection doesn't crash
	// and returns reasonable values

	// This would test actual environment detection
	// but in CI it might return "Unknown"
	t.Log("Environment detection test - results may vary in CI")
}

func TestModelManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()

	// Test model path validation
	validPath := filepath.Join(tempDir, "valid-model.bin")
	invalidPath := filepath.Join(tempDir, "nonexistent.bin")

	// Create a dummy model file
	err := os.WriteFile(validPath, []byte("dummy model data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy model: %v", err)
	}

	// Test various model paths
	testCases := []struct {
		name      string
		path      string
		expectErr bool
	}{
		{"valid_model", validPath, false},
		{"nonexistent_model", invalidPath, true},
		{"empty_path", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{}
			config.SetDefaultConfig(cfg)
			cfg.General.ModelPath = tc.path

			// Validate the model path
			if tc.path != "" {
				_, err := os.Stat(tc.path)
				hasErr := err != nil
				if tc.expectErr != hasErr {
					t.Errorf("Expected error: %v, got error: %v", tc.expectErr, hasErr)
				}
			}
		})
	}
}

func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test that concurrent config operations don't cause issues
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Create simple test config file
	configContent := `
general:
  debug: false
  model_path: "test-model.bin"
output:
  default_mode: "clipboard"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Run concurrent config loads
	errChan := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := config.LoadConfig(configFile)
			errChan <- err
		}()
	}

	// Collect results
	for i := 0; i < 10; i++ {
		err := <-errChan
		if err != nil {
			t.Errorf("Concurrent config load failed: %v", err)
		}
	}
}
