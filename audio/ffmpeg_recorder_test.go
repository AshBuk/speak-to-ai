// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package audio

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
)

// TestNewFFmpegRecorder tests the creation of FFmpegRecorder
func TestNewFFmpegRecorder(t *testing.T) {
	cfg := &config.Config{}
	cfg.Audio.Device = "default"
	cfg.Audio.Format = "S16_LE"
	cfg.Audio.SampleRate = 16000
	cfg.Audio.Channels = 1

	recorder := NewFFmpegRecorder(cfg)

	if recorder == nil {
		t.Fatal("Expected recorder to be created, got nil")
	}

	if recorder.config != cfg {
		t.Error("Expected config to be set correctly")
	}

	// Test that BaseRecorder is properly initialized
	if recorder.config.Audio.Device != "default" {
		t.Errorf("Expected device 'default', got %s", recorder.config.Audio.Device)
	}
}

// TestFFmpegRecorder_buildCommandArgs tests command argument building with real scenarios
func TestFFmpegRecorder_buildCommandArgs(t *testing.T) {
	tests := []struct {
		name          string
		device        string
		sampleRate    int
		channels      int
		useBuffer     bool
		streamingMode bool
		expectWavMode bool
		expectDevice  string
		expectRate    string
	}{
		{
			name:          "basic file output",
			device:        "hw:0,0",
			sampleRate:    44100,
			channels:      2,
			useBuffer:     false,
			streamingMode: false,
			expectWavMode: false,
			expectDevice:  "hw:0,0",
			expectRate:    "44100",
		},
		{
			name:          "streaming mode should output to stdout",
			device:        "default",
			sampleRate:    16000,
			channels:      1,
			useBuffer:     false,
			streamingMode: true,
			expectWavMode: true,
			expectDevice:  "default",
			expectRate:    "16000",
		},
		{
			name:          "buffer mode should output to stdout",
			device:        "plughw:1,0",
			sampleRate:    48000,
			channels:      1,
			useBuffer:     true,
			streamingMode: false,
			expectWavMode: true,
			expectDevice:  "plughw:1,0",
			expectRate:    "48000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.Audio.Device = tt.device
			cfg.Audio.SampleRate = tt.sampleRate
			cfg.Audio.Channels = tt.channels

			recorder := NewFFmpegRecorder(cfg)
			recorder.useBuffer = tt.useBuffer
			recorder.streamingEnabled = tt.streamingMode
			recorder.outputFile = "/tmp/test.wav"

			args := recorder.buildCommandArgs()

			// Verify device argument
			found := false
			for i, arg := range args {
				if arg == "-i" && i+1 < len(args) && args[i+1] == tt.expectDevice {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected device %s not found in args: %v", tt.expectDevice, args)
			}

			// Verify sample rate argument
			found = false
			for i, arg := range args {
				if arg == "-ar" && i+1 < len(args) && args[i+1] == tt.expectRate {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected sample rate %s not found in args: %v", tt.expectRate, args)
			}

			// Check for wav mode when expected (streaming/buffer)
			hasWavMode := false
			hasStdout := false
			for i, arg := range args {
				if arg == "-f" && i+1 < len(args) && args[i+1] == "wav" {
					hasWavMode = true
				}
				if arg == "-" {
					hasStdout = true
				}
			}
			if tt.expectWavMode && (!hasWavMode || !hasStdout) {
				t.Error("Expected wav mode with stdout (-f wav -) in streaming/buffer mode")
			}
			if !tt.expectWavMode && hasWavMode && hasStdout {
				t.Error("Did not expect wav stdout mode in file output mode")
			}

			// Verify quality setting is always present
			found = false
			for i, arg := range args {
				if arg == "-q:a" && i+1 < len(args) && args[i+1] == "0" {
					found = true
					break
				}
			}
			if !found {
				t.Error("Expected quality setting -q:a 0 in command args")
			}
		})
	}
}

// TestFFmpegRecorder_OutputFileHandling tests proper output file management
func TestFFmpegRecorder_OutputFileHandling(t *testing.T) {
	cfg := &config.Config{}
	cfg.Audio.Device = "default"
	cfg.Audio.SampleRate = 16000
	cfg.Audio.Channels = 1

	recorder := NewFFmpegRecorder(cfg)

	// Initially, no output file should be set
	if recorder.GetOutputFile() != "" {
		t.Errorf("Expected empty output file initially, got %s", recorder.GetOutputFile())
	}

	// Test that we can get/set output file through the interface
	tempDir := t.TempDir()
	expectedFile := filepath.Join(tempDir, "test_ffmpeg_recording.wav")
	recorder.outputFile = expectedFile

	if recorder.GetOutputFile() != expectedFile {
		t.Errorf("Expected output file %s, got %s", expectedFile, recorder.GetOutputFile())
	}
}

// TestFFmpegRecorder_StopRecording tests the stop recording functionality
func TestFFmpegRecorder_StopRecording(t *testing.T) {
	cfg := &config.Config{}
	cfg.Audio.Device = "default"
	cfg.Audio.SampleRate = 16000
	cfg.Audio.Channels = 1

	recorder := NewFFmpegRecorder(cfg)

	// Set up output file before testing stop
	tempDir := t.TempDir()
	expectedFile := filepath.Join(tempDir, "test_ffmpeg_output.wav")
	recorder.outputFile = expectedFile

	// Test stopping without starting (should not panic)
	outputFile, err := recorder.StopRecording()

	// When no recording process is started, StopProcess() returns error
	// and the function returns empty string
	if err != nil {
		// This is expected - no recording was started
		t.Logf("StopRecording returned error as expected: %v", err)
		if outputFile != "" {
			t.Errorf("Expected empty output file on error, got %s", outputFile)
		}
	} else {
		// If no error, should return the file path
		if outputFile != expectedFile {
			t.Errorf("Expected output file %s, got %s", expectedFile, outputFile)
		}
	}
}

// TestFFmpegRecorder_StreamingConfiguration tests streaming mode setup
func TestFFmpegRecorder_StreamingConfiguration(t *testing.T) {
	cfg := &config.Config{}
	cfg.Audio.Device = "default"
	cfg.Audio.EnableStreaming = true

	recorder := NewFFmpegRecorder(cfg)

	// Should inherit streaming setting from config
	if !recorder.UseStreaming() {
		t.Error("Expected streaming to be enabled from config")
	}

	// Test manual streaming toggle
	recorder.streamingEnabled = false
	if recorder.UseStreaming() {
		t.Error("Expected streaming to be disabled after manual toggle")
	}

	recorder.streamingEnabled = true
	if !recorder.UseStreaming() {
		t.Error("Expected streaming to be enabled after manual toggle")
	}
}

// TestFFmpegRecorder_InvalidConfiguration tests error handling with bad configs
func TestFFmpegRecorder_InvalidConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *config.Config
		expectError bool
	}{
		{
			name: "empty device name",
			setupConfig: func() *config.Config {
				cfg := &config.Config{}
				cfg.Audio.Device = ""
				cfg.Audio.SampleRate = 16000
				cfg.Audio.Channels = 1
				return cfg
			},
			expectError: false, // ffmpeg might handle empty device
		},
		{
			name: "zero sample rate",
			setupConfig: func() *config.Config {
				cfg := &config.Config{}
				cfg.Audio.Device = "default"
				cfg.Audio.SampleRate = 0
				cfg.Audio.Channels = 1
				return cfg
			},
			expectError: true, // This should cause issues
		},
		{
			name: "zero channels",
			setupConfig: func() *config.Config {
				cfg := &config.Config{}
				cfg.Audio.Device = "default"
				cfg.Audio.SampleRate = 16000
				cfg.Audio.Channels = 0
				return cfg
			},
			expectError: true, // This should cause issues
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()
			recorder := NewFFmpegRecorder(cfg)

			// Build args to see if they're reasonable
			args := recorder.buildCommandArgs()

			if tt.expectError {
				// Check for obviously problematic arguments
				for i, arg := range args {
					if arg == "-ar" && i+1 < len(args) && args[i+1] == "0" {
						t.Log("Correctly detected zero sample rate in args")
						return
					}
					if arg == "-ac" && i+1 < len(args) && args[i+1] == "0" {
						t.Log("Correctly detected zero channels in args")
						return
					}
				}
				t.Error("Expected to detect invalid configuration in command args")
			}
		})
	}
}

// TestFFmpegRecorder_AudioLevelCallbacks tests audio level monitoring
func TestFFmpegRecorder_AudioLevelCallbacks(t *testing.T) {
	cfg := &config.Config{}
	cfg.Audio.Device = "default"
	recorder := NewFFmpegRecorder(cfg)

	// Test initial audio level
	if recorder.GetAudioLevel() != 0.0 {
		t.Errorf("Expected initial audio level 0.0, got %f", recorder.GetAudioLevel())
	}

	// Test callback setting
	var receivedLevel float64
	var callbackCalled bool

	recorder.SetAudioLevelCallback(func(level float64) {
		receivedLevel = level
		callbackCalled = true
	})

	// Test audio level update
	testLevel := 0.85
	recorder.updateAudioLevel(testLevel)

	// Give callback time to execute
	time.Sleep(1 * time.Millisecond)

	if !callbackCalled {
		t.Error("Expected audio level callback to be called")
	}

	if receivedLevel != testLevel {
		t.Errorf("Expected callback to receive level %f, got %f", testLevel, receivedLevel)
	}

	if recorder.GetAudioLevel() != testLevel {
		t.Errorf("Expected stored audio level %f, got %f", testLevel, recorder.GetAudioLevel())
	}
}

// TestFFmpegRecorder_CleanupFile tests file cleanup functionality
func TestFFmpegRecorder_CleanupFile(t *testing.T) {
	cfg := &config.Config{}
	recorder := NewFFmpegRecorder(cfg)

	// Create a temporary file to test cleanup
	tempFile, err := os.CreateTemp("", "ffmpeg_test_*.wav")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()

	fileName := tempFile.Name()
	recorder.outputFile = fileName

	// Verify file exists
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		t.Fatal("Temp file should exist before cleanup")
	}

	// Test cleanup
	err = recorder.CleanupFile()
	if err != nil {
		t.Errorf("CleanupFile returned error: %v", err)
	}

	// File should be removed
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		t.Error("File should be removed after cleanup")
		os.Remove(fileName) // Clean up in case test failed
	}
}

// TestFFmpegRecorder_BufferMode tests in-memory buffer functionality
func TestFFmpegRecorder_BufferMode(t *testing.T) {
	cfg := &config.Config{}
	cfg.Audio.ExpectedDuration = 5 // Short duration should trigger buffer mode
	cfg.Audio.SampleRate = 16000   // Low sample rate should trigger buffer mode

	recorder := NewFFmpegRecorder(cfg)

	// Should be using buffer mode for short, low-quality recordings
	if !recorder.useBuffer {
		t.Error("Expected buffer mode to be enabled for short recordings")
	}

	// Test buffer access
	stream, err := recorder.GetAudioStream()
	if err != nil {
		t.Errorf("GetAudioStream returned error: %v", err)
	}

	if stream == nil {
		t.Error("Expected audio stream to be available")
	}
}
