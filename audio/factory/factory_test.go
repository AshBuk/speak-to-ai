// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package factory

import (
	"testing"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/testutils"
)

func TestNewAudioRecorderFactory(t *testing.T) {
	config := &config.Config{}
	mockLogger := testutils.NewMockLogger()
	factory := NewAudioRecorderFactory(config, mockLogger)

	if factory.config != config {
		t.Errorf("expected config to be set correctly")
	}
}

func TestAudioRecorderFactory_CreateRecorder(t *testing.T) {
	tests := []struct {
		name            string
		recordingMethod string
		expectError     bool
		expectedType    string
	}{
		{
			name:            "arecord recorder",
			recordingMethod: "arecord",
			expectError:     false,
			expectedType:    "*audio.ArecordRecorder",
		},
		{
			name:            "ffmpeg recorder",
			recordingMethod: "ffmpeg",
			expectError:     false,
			expectedType:    "*audio.FFmpegRecorder",
		},
		{
			name:            "unsupported method",
			recordingMethod: "unsupported",
			expectError:     true,
			expectedType:    "",
		},
		{
			name:            "empty method",
			recordingMethod: "",
			expectError:     true,
			expectedType:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &config.Config{}
			config.Audio.RecordingMethod = tt.recordingMethod

			mockLogger := testutils.NewMockLogger()
			factory := NewAudioRecorderFactory(config, mockLogger)
			recorder, err := factory.CreateRecorder()

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && recorder == nil {
				t.Errorf("expected recorder to be created")
			}
		})
	}
}

func TestGetRecorder(t *testing.T) {
	tests := []struct {
		name            string
		recordingMethod string
		expectError     bool
	}{
		{
			name:            "valid arecord config",
			recordingMethod: "arecord",
			expectError:     false,
		},
		{
			name:            "valid ffmpeg config",
			recordingMethod: "ffmpeg",
			expectError:     false,
		},
		{
			name:            "invalid config",
			recordingMethod: "invalid",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &config.Config{}
			config.Audio.RecordingMethod = tt.recordingMethod

			mockLogger := testutils.NewMockLogger()
			recorder, err := GetRecorder(config, mockLogger)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && recorder == nil {
				t.Errorf("expected recorder to be created")
			}
		})
	}
}
