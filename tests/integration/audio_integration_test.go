//go:build integration

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package integration

import (
	"os"
	"testing"
	"time"

	"github.com/AshBuk/dabri/v2/audio/factory"
	"github.com/AshBuk/dabri/v2/audio/processing"
	"github.com/AshBuk/dabri/v2/config"
	"github.com/AshBuk/dabri/v2/internal/logger"
)

func TestAudioRecordingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test real audio recording functionality
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)
	tempDir := t.TempDir()
	cfg.General.TempAudioPath = tempDir
	t.Run("arecord_real_device", func(t *testing.T) {
		// Test with real audio device if available
		cfg.Audio.RecordingMethod = "arecord"
		cfg.Audio.Device = "default"
		cfg.Audio.SampleRate = 16000
		// channels removed (mono enforced internally)

		testLogger := logger.NewDefaultLogger(logger.InfoLevel)
		tempManager := processing.NewTempFileManager(30 * time.Minute)
		recorder, err := factory.GetRecorder(cfg, testLogger, tempManager)
		if err != nil {
			t.Skipf("Audio recorder not available: %v", err)
		}

		err = recorder.StartRecording()
		if err != nil {
			t.Skipf("Could not start recording (no audio device?): %v", err)
		}

		// Record for a short time
		time.Sleep(100 * time.Millisecond)
		// Ensure cleanup happens even if test fails
		defer func() {
			recorder.StopRecording()
			recorder.CleanupFile()
		}()

		outputFile, err := recorder.StopRecording()
		if err != nil {
			t.Skipf("Failed to stop recording (audio device issue): %v", err)
		}
		// Check that file was created
		if outputFile != "" {
			if _, err := os.Stat(outputFile); err != nil {
				t.Errorf("Recording file not created: %v", err)
			}
		}
	})

	t.Run("ffmpeg_real_device", func(t *testing.T) {
		// Test with FFmpeg if available
		cfg.Audio.RecordingMethod = "ffmpeg"
		cfg.Audio.Device = "default"

		testLogger := logger.NewDefaultLogger(logger.InfoLevel)
		tempManager := processing.NewTempFileManager(30 * time.Minute)
		recorder, err := factory.GetRecorder(cfg, testLogger, tempManager)
		if err != nil {
			t.Skipf("FFmpeg recorder not available: %v", err)
		}

		err = recorder.StartRecording()
		if err != nil {
			t.Skipf("Could not start FFmpeg recording: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Ensure cleanup happens even if test fails
		defer func() {
			recorder.StopRecording()
			recorder.CleanupFile()
		}()

		outputFile, err := recorder.StopRecording()
		if err != nil {
			t.Skipf("Failed to stop FFmpeg recording (audio device issue): %v", err)
		}

		if outputFile != "" {
			if _, err := os.Stat(outputFile); err != nil {
				t.Errorf("FFmpeg recording file not created: %v", err)
			}
		}
	})
}
