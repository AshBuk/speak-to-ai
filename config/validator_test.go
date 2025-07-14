package config

import (
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func() *Config
		expectError    bool
		expectedValues map[string]interface{}
	}{
		{
			name: "valid config",
			setupConfig: func() *Config {
				config := &Config{}
				SetDefaultConfig(config)
				return config
			},
			expectError: false,
			expectedValues: map[string]interface{}{
				"modelType":  "base",
				"sampleRate": 16000,
			},
		},
		{
			name: "path traversal attack in model path",
			setupConfig: func() *Config {
				config := &Config{}
				SetDefaultConfig(config)
				config.General.ModelPath = "../../../etc/passwd"
				return config
			},
			expectError: true,
			expectedValues: map[string]interface{}{
				"modelPath": "sources/language-models/base.bin",
			},
		},
		{
			name: "invalid model type",
			setupConfig: func() *Config {
				config := &Config{}
				SetDefaultConfig(config)
				config.General.ModelType = "invalid"
				return config
			},
			expectError: true,
			expectedValues: map[string]interface{}{
				"modelType": "base",
			},
		},
		{
			name: "invalid sample rate - too low",
			setupConfig: func() *Config {
				config := &Config{}
				SetDefaultConfig(config)
				config.Audio.SampleRate = 1000
				return config
			},
			expectError: true,
			expectedValues: map[string]interface{}{
				"sampleRate": 16000,
			},
		},
		{
			name: "invalid sample rate - too high",
			setupConfig: func() *Config {
				config := &Config{}
				SetDefaultConfig(config)
				config.Audio.SampleRate = 100000
				return config
			},
			expectError: true,
			expectedValues: map[string]interface{}{
				"sampleRate": 16000,
			},
		},
		{
			name: "invalid channels",
			setupConfig: func() *Config {
				config := &Config{}
				SetDefaultConfig(config)
				config.Audio.Channels = 5
				return config
			},
			expectError: true,
			expectedValues: map[string]interface{}{
				"channels": 1,
			},
		},
		{
			name: "invalid recording method",
			setupConfig: func() *Config {
				config := &Config{}
				SetDefaultConfig(config)
				config.Audio.RecordingMethod = "invalid"
				return config
			},
			expectError: true,
			expectedValues: map[string]interface{}{
				"recordingMethod": "arecord",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupConfig()
			err := ValidateConfig(config)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check expected values
			if modelType, ok := tt.expectedValues["modelType"]; ok {
				if config.General.ModelType != modelType {
					t.Errorf("expected ModelType %v, got %v", modelType, config.General.ModelType)
				}
			}
			if sampleRate, ok := tt.expectedValues["sampleRate"]; ok {
				if config.Audio.SampleRate != sampleRate {
					t.Errorf("expected SampleRate %v, got %v", sampleRate, config.Audio.SampleRate)
				}
			}
			if modelPath, ok := tt.expectedValues["modelPath"]; ok {
				if config.General.ModelPath != modelPath {
					t.Errorf("expected ModelPath %v, got %v", modelPath, config.General.ModelPath)
				}
			}
			if channels, ok := tt.expectedValues["channels"]; ok {
				if config.Audio.Channels != channels {
					t.Errorf("expected Channels %v, got %v", channels, config.Audio.Channels)
				}
			}
			if recordingMethod, ok := tt.expectedValues["recordingMethod"]; ok {
				if config.Audio.RecordingMethod != recordingMethod {
					t.Errorf("expected RecordingMethod %v, got %v", recordingMethod, config.Audio.RecordingMethod)
				}
			}
		})
	}
}

func TestIsCommandAllowed(t *testing.T) {
	config := &Config{}
	config.Security.AllowedCommands = []string{"echo", "ls", "cat"}

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"allowed command", "echo", true},
		{"allowed command with path", "/bin/echo", true},
		{"disallowed command", "rm", false},
		{"empty command", "", false},
		{"malicious command", "rm -rf /", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.IsCommandAllowed(tt.command)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSanitizeCommandArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "clean args",
			args:     []string{"--help", "--version"},
			expected: []string{"--help", "--version"},
		},
		{
			name:     "args with path traversal",
			args:     []string{"../../../etc/passwd", "--help"},
			expected: []string{"--help"},
		},
		{
			name:     "args with dangerous chars",
			args:     []string{"file.txt", "$(rm -rf /)", "--safe"},
			expected: []string{"file.txt", "--safe"},
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeCommandArgs(tt.args)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d args, got %d", len(tt.expected), len(result))
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("expected arg %d to be %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}
