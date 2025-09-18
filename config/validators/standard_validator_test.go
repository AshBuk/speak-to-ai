// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package validators

import (
	"testing"

	"github.com/AshBuk/speak-to-ai/config/models"
)

// setDefaultConfigForTest sets default values for testing
func setDefaultConfigForTest(config *models.Config) {
	config.General.Debug = false
	config.General.WhisperModel = "small-q5_1"
	config.General.ModelPath = "sources/language-models/small-q5_1.bin"
	config.General.TempAudioPath = "/tmp"
	config.General.Language = "en"
	config.General.LogFile = ""

	config.Audio.Device = "default"
	config.Audio.SampleRate = 16000
	config.Audio.Format = "s16le"
	config.Audio.RecordingMethod = "arecord"
	config.Audio.ExpectedDuration = 0
	config.Audio.MaxRecordingTime = 300
	// TODO: Next feature - VAD implementation
	// config.Audio.EnableVAD = false
	// config.Audio.VADSensitivity = "medium"
	// config.Audio.AutoStartStop = false

	config.Output.DefaultMode = models.OutputModeActiveWindow
	config.Output.ClipboardTool = "auto"
	config.Output.TypeTool = "auto"

	config.Security.AllowedCommands = []string{"arecord", "ffmpeg", "whisper", "xdotool", "wtype", "ydotool", "wl-copy", "wl-paste", "xclip", "notify-send", "xdg-open"}
	config.Security.CheckIntegrity = false
	config.Security.ConfigHash = ""
	config.Security.MaxTempFileSize = 50 * 1024 * 1024
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func() *models.Config
		expectError    bool
		expectedValues map[string]interface{}
	}{
		{
			name: "valid config",
			setupConfig: func() *models.Config {
				config := &models.Config{}
				setDefaultConfigForTest(config)
				return config
			},
			expectError: false,
			expectedValues: map[string]interface{}{
				"modelType":  "small",
				"sampleRate": 16000,
			},
		},
		{
			name: "path traversal attack in model path",
			setupConfig: func() *models.Config {
				config := &models.Config{}
				setDefaultConfigForTest(config)
				config.General.ModelPath = "../../../etc/passwd"
				return config
			},
			expectError: true,
			expectedValues: map[string]interface{}{
				"modelPath": "sources/language-models/small-q5_1.bin",
			},
		},
		{
			name: "invalid whisper model",
			setupConfig: func() *models.Config {
				config := &models.Config{}
				setDefaultConfigForTest(config)
				config.General.WhisperModel = "invalid"
				return config
			},
			expectError: true,
			expectedValues: map[string]interface{}{
				"whisperModel": "small-q5_1",
			},
		},
		{
			name: "invalid sample rate - too low",
			setupConfig: func() *models.Config {
				config := &models.Config{}
				setDefaultConfigForTest(config)
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
			setupConfig: func() *models.Config {
				config := &models.Config{}
				setDefaultConfigForTest(config)
				config.Audio.SampleRate = 100000
				return config
			},
			expectError: true,
			expectedValues: map[string]interface{}{
				"sampleRate": 16000,
			},
		},
		{
			name: "invalid recording method",
			setupConfig: func() *models.Config {
				config := &models.Config{}
				setDefaultConfigForTest(config)
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
			if whisperModel, ok := tt.expectedValues["whisperModel"]; ok {
				if config.General.WhisperModel != whisperModel {
					t.Errorf("expected WhisperModel %v, got %v", whisperModel, config.General.WhisperModel)
				}
			}
			if modelPath, ok := tt.expectedValues["modelPath"]; ok {
				if config.General.ModelPath != modelPath {
					t.Errorf("expected ModelPath %v, got %v", modelPath, config.General.ModelPath)
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
			// channels removed
			if recordingMethod, ok := tt.expectedValues["recordingMethod"]; ok {
				if config.Audio.RecordingMethod != recordingMethod {
					t.Errorf("expected RecordingMethod %v, got %v", recordingMethod, config.Audio.RecordingMethod)
				}
			}
		})
	}
}
