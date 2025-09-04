// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package audio

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
)

// AudioRecorderFactory creates appropriate audio recorder instances
type AudioRecorderFactory struct {
	config *config.Config
}

// NewAudioRecorderFactory creates a new factory instance
func NewAudioRecorderFactory(config *config.Config) *AudioRecorderFactory {
	return &AudioRecorderFactory{
		config: config,
	}
}

// CreateRecorder creates a recorder based on config settings
func (f *AudioRecorderFactory) CreateRecorder() (AudioRecorder, error) {
	// Diagnose audio system before creating recorder
	f.DiagnoseAudioSystem()

	switch f.config.Audio.RecordingMethod {
	case "arecord":
		return NewArecordRecorder(f.config), nil
	case "ffmpeg":
		return NewFFmpegRecorder(f.config), nil
	default:
		return nil, fmt.Errorf("unsupported recording method: %s", f.config.Audio.RecordingMethod)
	}
}

// DiagnoseAudioSystem performs audio system diagnostics
func (f *AudioRecorderFactory) DiagnoseAudioSystem() {
	log.Printf("[AUDIO DIAGNOSTICS] Recording method: %s, Device: %s",
		f.config.Audio.RecordingMethod, f.config.Audio.Device)

	// Check if commands are available and perform method-specific diagnostics
	switch f.config.Audio.RecordingMethod {
	case "ffmpeg":
		if _, err := exec.LookPath("ffmpeg"); err != nil {
			log.Printf("[AUDIO DIAGNOSTICS] WARNING: ffmpeg not found: %v", err)
		}
		// Check PulseAudio sources for ffmpeg
		if out, err := exec.Command("pactl", "list", "short", "sources").Output(); err == nil {
			log.Printf("[AUDIO DIAGNOSTICS] PulseAudio sources available:")
			log.Printf("%s", string(out))
		} else {
			log.Printf("[AUDIO DIAGNOSTICS] WARNING: Cannot list PulseAudio sources: %v", err)
		}
	case "arecord":
		if _, err := exec.LookPath("arecord"); err != nil {
			log.Printf("[AUDIO DIAGNOSTICS] WARNING: arecord not found: %v", err)
		}
		// Check ALSA devices for arecord
		if out, err := exec.Command("arecord", "-l").Output(); err == nil {
			log.Printf("[AUDIO DIAGNOSTICS] ALSA capture devices:")
			log.Printf("%s", string(out))
		} else {
			log.Printf("[AUDIO DIAGNOSTICS] WARNING: Cannot list ALSA devices: %v", err)
		}
	}
}

// TestRecorderMethod tests if a recording method works properly
func (f *AudioRecorderFactory) TestRecorderMethod(method string) error {
	// Create temporary test config
	testConfig := *f.config
	testConfig.Audio.RecordingMethod = method

	log.Printf("[AUDIO TEST] Testing %s recorder...", method)

	var testArgs []string
	var cmdName string

	switch method {
	case "ffmpeg":
		cmdName = "ffmpeg"
		testArgs = []string{
			"-y", "-f", "pulse", "-i", testConfig.Audio.Device,
			"-ar", "16000", "-ac", "1", "-acodec", "pcm_s16le",
			"-t", "0.5", "-f", "null", "-"}
	case "arecord":
		cmdName = "arecord"
		testArgs = []string{
			"-D", testConfig.Audio.Device, "-f", "S16_LE",
			"-r", "16000", "-c", "1", "-d", "1", "/dev/null"}
	default:
		return fmt.Errorf("unsupported test method: %s", method)
	}

	// Run test command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdName, testArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("[AUDIO TEST] %s test failed: %v", method, err)
		log.Printf("[AUDIO TEST] %s output: %s", method, string(output))
		return fmt.Errorf("%s test failed: %w", method, err)
	}

	log.Printf("[AUDIO TEST] %s test passed", method)
	return nil
}

// CreateRecorderWithFallback creates a recorder with automatic fallback
func (f *AudioRecorderFactory) CreateRecorderWithFallback() (AudioRecorder, error) {
	// First, try the configured method
	recorder, err := f.CreateRecorder()
	if err != nil {
		return nil, err
	}

	// Test if the method actually works
	testErr := f.TestRecorderMethod(f.config.Audio.RecordingMethod)
	if testErr == nil {
		log.Printf("[AUDIO] Using configured recorder: %s", f.config.Audio.RecordingMethod)
		return recorder, nil
	}

	log.Printf("[AUDIO] Configured method %s failed test: %v", f.config.Audio.RecordingMethod, testErr)

	// Try fallback methods
	fallbacks := []string{"arecord", "ffmpeg"}
	originalMethod := f.config.Audio.RecordingMethod

	for _, method := range fallbacks {
		if method == originalMethod {
			continue // Skip the already failed method
		}

		if f.TestRecorderMethod(method) == nil {
			log.Printf("[AUDIO] Automatically switching to %s (fallback)", method)
			f.config.Audio.RecordingMethod = method
			return f.CreateRecorder()
		}
	}

	return nil, fmt.Errorf("no working audio recorder found (tested: %s)", f.config.Audio.RecordingMethod)
}

// GetRecorder is a convenience function to create a recorder directly from config
func GetRecorder(config *config.Config) (AudioRecorder, error) {
	factory := NewAudioRecorderFactory(config)
	return factory.CreateRecorder()
}

// GetRecorderWithFallback is a convenience function to create a recorder with fallback
func GetRecorderWithFallback(config *config.Config) (AudioRecorder, error) {
	factory := NewAudioRecorderFactory(config)
	return factory.CreateRecorderWithFallback()
}
