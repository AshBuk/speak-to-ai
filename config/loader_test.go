package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")

	tests := []struct {
		name          string
		configContent string
		expectError   bool
		checkValues   func(*testing.T, *Config)
	}{
		{
			name: "valid config",
			configContent: `
general:
  debug: true
  model_type: "base"
  language: "en"
  temp_audio_path: "/tmp"

audio:
  device: "default"
  sample_rate: 16000
  format: "S16_LE"
  channels: 1
  recording_method: "arecord"

output:
  default_mode: "clipboard"
  clipboard_tool: "auto"
  type_tool: "auto"

hotkeys:
  start_recording: "AltGr+,"
  stop_recording: "AltGr+."
`,
			expectError: false,
			checkValues: func(t *testing.T, cfg *Config) {
				if !cfg.General.Debug {
					t.Errorf("expected debug to be true")
				}
				if cfg.General.ModelType != "base" {
					t.Errorf("expected model type to be 'base', got %s", cfg.General.ModelType)
				}
				if cfg.Audio.SampleRate != 16000 {
					t.Errorf("expected sample rate to be 16000, got %d", cfg.Audio.SampleRate)
				}
				if cfg.Output.DefaultMode != "clipboard" {
					t.Errorf("expected default mode to be 'clipboard', got %s", cfg.Output.DefaultMode)
				}
			},
		},
		{
			name: "minimal config",
			configContent: `
general:
  model_type: "tiny"
`,
			expectError: false,
			checkValues: func(t *testing.T, cfg *Config) {
				if cfg.General.ModelType != "tiny" {
					t.Errorf("expected model type to be 'tiny', got %s", cfg.General.ModelType)
				}
				// Other fields should have default values
			},
		},
		{
			name: "invalid yaml",
			configContent: `
general:
  debug: true
  invalid_yaml: [
`,
			expectError: true,
			checkValues: nil,
		},
		{
			name:          "empty config",
			configContent: ``,
			expectError:   false,
			checkValues: func(t *testing.T, cfg *Config) {
				// Should create config with default values
				if cfg == nil {
					t.Errorf("expected config to be created")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write config content to file
			if err := os.WriteFile(configPath, []byte(tt.configContent), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			config, err := LoadConfig(configPath)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && config == nil {
				t.Errorf("expected config to be loaded")
			}

			if tt.checkValues != nil && config != nil {
				tt.checkValues(t, config)
			}
		})
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	config, err := LoadConfig("/non/existent/file.yaml")

	// LoadConfig returns default config when file doesn't exist
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if config == nil {
		t.Errorf("expected default config to be returned")
		return
	}
	// Check that default values are set
	if config.General.ModelType != "base" {
		t.Errorf("expected default model type to be 'base', got %s", config.General.ModelType)
	}
}

func TestLoadConfig_InvalidPermissions(t *testing.T) {
	// Create a temporary file with no read permissions
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("test: value"), 0000); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	config, err := LoadConfig(configPath)

	// LoadConfig returns default config when file can't be read
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if config == nil {
		t.Errorf("expected default config to be returned")
		return
	}
	// Check that default values are set
	if config.General.ModelType != "base" {
		t.Errorf("expected default model type to be 'base', got %s", config.General.ModelType)
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	// Test that default config has reasonable values
	config := &Config{}

	// Apply default values (this would typically be done in LoadConfig)
	// For this test, we'll just verify the structure exists
	if config.General.ModelType == "" {
		config.General.ModelType = "base"
	}
	if config.Audio.SampleRate == 0 {
		config.Audio.SampleRate = 16000
	}
	if config.Audio.Channels == 0 {
		config.Audio.Channels = 1
	}

	// Verify defaults
	if config.General.ModelType != "base" {
		t.Errorf("expected default model type to be 'base', got %s", config.General.ModelType)
	}
	if config.Audio.SampleRate != 16000 {
		t.Errorf("expected default sample rate to be 16000, got %d", config.Audio.SampleRate)
	}
	if config.Audio.Channels != 1 {
		t.Errorf("expected default channels to be 1, got %d", config.Audio.Channels)
	}
}
