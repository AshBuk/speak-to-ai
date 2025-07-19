package audio

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
)

// TestNewArecordRecorder tests the creation of ArecordRecorder
func TestNewArecordRecorder(t *testing.T) {
	cfg := &config.Config{}
	cfg.Audio.Device = "default"
	cfg.Audio.Format = "S16_LE"
	cfg.Audio.SampleRate = 16000
	cfg.Audio.Channels = 1

	recorder := NewArecordRecorder(cfg)

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

	if recorder.config.Audio.SampleRate != 16000 {
		t.Errorf("Expected sample rate 16000, got %d", recorder.config.Audio.SampleRate)
	}
}

// TestArecordRecorder_buildCommandArgs tests command argument building with real scenarios
func TestArecordRecorder_buildCommandArgs(t *testing.T) {
	tests := []struct {
		name           string
		device         string
		format         string
		sampleRate     int
		channels       int
		useBuffer      bool
		streamingMode  bool
		expectRawMode  bool
		expectDevice   string
		expectFormat   string
	}{
		{
			name:          "basic file output",
			device:        "hw:0,0",
			format:        "S16_LE",
			sampleRate:    44100,
			channels:      2,
			useBuffer:     false,
			streamingMode: false,
			expectRawMode: false,
			expectDevice:  "hw:0,0",
			expectFormat:  "S16_LE",
		},
		{
			name:          "streaming mode should use raw format",
			device:        "default",
			format:        "S16_LE",
			sampleRate:    16000,
			channels:      1,
			useBuffer:     false,
			streamingMode: true,
			expectRawMode: true,
			expectDevice:  "default",
			expectFormat:  "S16_LE",
		},
		{
			name:          "buffer mode should use raw format",
			device:        "plughw:1,0",
			format:        "S24_LE",
			sampleRate:    48000,
			channels:      1,
			useBuffer:     true,
			streamingMode: false,
			expectRawMode: true,
			expectDevice:  "plughw:1,0",
			expectFormat:  "S24_LE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.Audio.Device = tt.device
			cfg.Audio.Format = tt.format
			cfg.Audio.SampleRate = tt.sampleRate
			cfg.Audio.Channels = tt.channels
			
			recorder := NewArecordRecorder(cfg)
			recorder.useBuffer = tt.useBuffer
			recorder.streamingEnabled = tt.streamingMode
			recorder.outputFile = "/tmp/test.wav"

			args := recorder.buildCommandArgs()

			// Verify device argument
			found := false
			for i, arg := range args {
				if arg == "-D" && i+1 < len(args) && args[i+1] == tt.expectDevice {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected device %s not found in args: %v", tt.expectDevice, args)
			}

			// Verify format argument
			found = false
			for i, arg := range args {
				if arg == "-f" && i+1 < len(args) && args[i+1] == tt.expectFormat {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected format %s not found in args: %v", tt.expectFormat, args)
			}

			// Check for raw mode when expected
			hasRawMode := false
			for i, arg := range args {
				if arg == "-t" && i+1 < len(args) && args[i+1] == "raw" {
					hasRawMode = true
					break
				}
			}
			if tt.expectRawMode && !hasRawMode {
				t.Error("Expected raw mode (-t raw) in streaming/buffer mode")
			}
			if !tt.expectRawMode && hasRawMode {
				t.Error("Did not expect raw mode in file output mode")
			}
		})
	}
}

// TestArecordRecorder_OutputFileHandling tests proper output file management
func TestArecordRecorder_OutputFileHandling(t *testing.T) {
	cfg := &config.Config{}
	cfg.Audio.Device = "default"
	cfg.Audio.Format = "S16_LE"
	cfg.Audio.SampleRate = 16000
	cfg.Audio.Channels = 1

	recorder := NewArecordRecorder(cfg)
	
	// Initially, no output file should be set
	if recorder.GetOutputFile() != "" {
		t.Errorf("Expected empty output file initially, got %s", recorder.GetOutputFile())
	}

	// Test that we can get/set output file through the interface
	tempDir := t.TempDir()
	expectedFile := filepath.Join(tempDir, "test_recording.wav")
	recorder.outputFile = expectedFile

	if recorder.GetOutputFile() != expectedFile {
		t.Errorf("Expected output file %s, got %s", expectedFile, recorder.GetOutputFile())
	}
}

// TestArecordRecorder_StreamingConfiguration tests streaming mode setup
func TestArecordRecorder_StreamingConfiguration(t *testing.T) {
	cfg := &config.Config{}
	cfg.Audio.Device = "default"
	cfg.Audio.EnableStreaming = true
	
	recorder := NewArecordRecorder(cfg)
	
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

// TestArecordRecorder_InvalidConfiguration tests error handling with bad configs
func TestArecordRecorder_InvalidConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		setupConfig func() *config.Config
		expectError bool
	}{
		{
			name: "empty device name",
			setupConfig: func() *config.Config {
				cfg := &config.Config{}
				cfg.Audio.Device = ""
				cfg.Audio.Format = "S16_LE"
				cfg.Audio.SampleRate = 16000
				cfg.Audio.Channels = 1
				return cfg
			},
			expectError: false, // arecord might handle empty device
		},
		{
			name: "zero sample rate",
			setupConfig: func() *config.Config {
				cfg := &config.Config{}
				cfg.Audio.Device = "default"
				cfg.Audio.Format = "S16_LE"
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
				cfg.Audio.Format = "S16_LE"
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
			recorder := NewArecordRecorder(cfg)
			
			// Build args to see if they're reasonable
			args := recorder.buildCommandArgs()
			
			if tt.expectError {
				// Check for obviously problematic arguments
				for i, arg := range args {
					if arg == "-r" && i+1 < len(args) && args[i+1] == "0" {
						t.Log("Correctly detected zero sample rate in args")
						return
					}
					if arg == "-c" && i+1 < len(args) && args[i+1] == "0" {
						t.Log("Correctly detected zero channels in args")
						return
					}
				}
				t.Error("Expected to detect invalid configuration in command args")
			}
		})
	}
}

// TestArecordRecorder_AudioLevelCallbacks tests audio level monitoring
func TestArecordRecorder_AudioLevelCallbacks(t *testing.T) {
	cfg := &config.Config{}
	cfg.Audio.Device = "default"
	recorder := NewArecordRecorder(cfg)
	
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
	testLevel := 0.75
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

// TestArecordRecorder_CleanupFile tests file cleanup functionality
func TestArecordRecorder_CleanupFile(t *testing.T) {
	cfg := &config.Config{}
	recorder := NewArecordRecorder(cfg)
	
	// Create a temporary file to test cleanup
	tempFile, err := os.CreateTemp("", "arecord_test_*.wav")
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

// TestArecordRecorder_BufferMode tests in-memory buffer functionality
func TestArecordRecorder_BufferMode(t *testing.T) {
	cfg := &config.Config{}
	cfg.Audio.ExpectedDuration = 5 // Short duration should trigger buffer mode
	cfg.Audio.SampleRate = 16000   // Low sample rate should trigger buffer mode
	
	recorder := NewArecordRecorder(cfg)
	
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