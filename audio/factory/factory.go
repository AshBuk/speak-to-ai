// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package factory

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/AshBuk/speak-to-ai/audio/interfaces"
	"github.com/AshBuk/speak-to-ai/audio/processing"
	"github.com/AshBuk/speak-to-ai/audio/recorders"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// Creates appropriate audio recorder instances based on configuration
type AudioRecorderFactory struct {
	config      *config.Config
	logger      logger.Logger
	tempManager *processing.TempFileManager
}

// Create a new factory instance
func NewAudioRecorderFactory(config *config.Config, logger logger.Logger, tempManager *processing.TempFileManager) *AudioRecorderFactory {
	return &AudioRecorderFactory{
		config:      config,
		logger:      logger,
		tempManager: tempManager,
	}
}

// Create a recorder based on the method specified in the configuration
func (f *AudioRecorderFactory) CreateRecorder() (interfaces.AudioRecorder, error) {
	// Diagnose the audio system before creating the recorder
	f.DiagnoseAudioSystem()

	switch f.config.Audio.RecordingMethod {
	case "arecord":
		return recorders.NewArecordRecorder(f.config, f.logger, f.tempManager), nil
	case "ffmpeg":
		return recorders.NewFFmpegRecorder(f.config, f.logger, f.tempManager), nil
	default:
		return nil, fmt.Errorf("unsupported recording method: %s", f.config.Audio.RecordingMethod)
	}
}

// Perform audio system diagnostics to help debug configuration issues
func (f *AudioRecorderFactory) DiagnoseAudioSystem() {
	f.logger.Info("[AUDIO DIAGNOSTICS] Recording method: %s, Device: %s",
		f.config.Audio.RecordingMethod, f.config.Audio.Device)

	// Check if the required command-line tools are available and log device info
	switch f.config.Audio.RecordingMethod {
	case "ffmpeg":
		if _, err := exec.LookPath("ffmpeg"); err != nil {
			f.logger.Warning("[AUDIO DIAGNOSTICS] WARNING: ffmpeg not found: %v", err)
		}
		// Check PulseAudio sources for ffmpeg
		if out, err := exec.Command("pactl", "list", "short", "sources").Output(); err == nil {
			f.logger.Info("[AUDIO DIAGNOSTICS] PulseAudio sources available:")
			f.logger.Info("%s", string(out))
		} else {
			f.logger.Warning("[AUDIO DIAGNOSTICS] WARNING: Cannot list PulseAudio sources: %v", err)
		}
	case "arecord":
		if _, err := exec.LookPath("arecord"); err != nil {
			f.logger.Warning("[AUDIO DIAGNOSTICS] WARNING: arecord not found: %v", err)
		}
		// Check ALSA devices for arecord
		if out, err := exec.Command("arecord", "-l").Output(); err == nil {
			f.logger.Info("[AUDIO DIAGNOSTICS] ALSA capture devices:")
			f.logger.Info("%s", string(out))
		} else {
			f.logger.Warning("[AUDIO DIAGNOSTICS] WARNING: Cannot list ALSA devices: %v", err)
		}
	}
}

// Test if a specific recording method works correctly
func (f *AudioRecorderFactory) TestRecorderMethod(method string) error {
	// Create a temporary test configuration
	testConfig := *f.config
	testConfig.Audio.RecordingMethod = method

	f.logger.Info("[AUDIO TEST] Testing %s recorder...", method)

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

	// Run the test command with a timeout to prevent hangs
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Security: validate the command and sanitize arguments before execution
	if !config.IsCommandAllowed(&testConfig, cmdName) {
		return fmt.Errorf("command not allowed: %s", cmdName)
	}
	safeArgs := config.SanitizeCommandArgs(testArgs)

	// #nosec G204 -- Safe: command is allowlisted and arguments are sanitized.
	cmd := exec.CommandContext(ctx, cmdName, safeArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		f.logger.Error("[AUDIO TEST] %s test failed: %v", method, err)
		f.logger.Info("[AUDIO TEST] %s output: %s", method, string(output))
		return fmt.Errorf("%s test failed: %w", method, err)
	}

	f.logger.Info("[AUDIO TEST] %s test passed", method)
	return nil
}

// Create a recorder using the configured method, but fall back to other supported
// methods if the primary one fails a functionality test
func (f *AudioRecorderFactory) CreateRecorderWithFallback() (interfaces.AudioRecorder, error) {
	// First, try the configured method
	recorder, err := f.CreateRecorder()
	if err != nil {
		return nil, err
	}

	// Test if the method actually works
	testErr := f.TestRecorderMethod(f.config.Audio.RecordingMethod)
	if testErr == nil {
		f.logger.Info("[AUDIO] Using configured recorder: %s", f.config.Audio.RecordingMethod)
		return recorder, nil
	}

	f.logger.Warning("[AUDIO] Configured method %s failed test: %v", f.config.Audio.RecordingMethod, testErr)

	// Try fallback methods if the configured one fails
	fallbacks := []string{"arecord", "ffmpeg"}
	originalMethod := f.config.Audio.RecordingMethod

	for _, method := range fallbacks {
		if method == originalMethod {
			continue // Skip the already failed method
		}

		if f.TestRecorderMethod(method) == nil {
			f.logger.Info("[AUDIO] Automatically switching to %s (fallback)", method)
			f.config.Audio.RecordingMethod = method
			return f.CreateRecorder()
		}
	}

	return nil, fmt.Errorf("no working audio recorder found (tested: %s)", f.config.Audio.RecordingMethod)
}

// Create a recorder directly from a configuration
func GetRecorder(config *config.Config, logger logger.Logger, tempManager *processing.TempFileManager) (interfaces.AudioRecorder, error) {
	factory := NewAudioRecorderFactory(config, logger, tempManager)
	return factory.CreateRecorder()
}

// Create a recorder with automatic fallback functionality
func GetRecorderWithFallback(config *config.Config, logger logger.Logger, tempManager *processing.TempFileManager) (interfaces.AudioRecorder, error) {
	factory := NewAudioRecorderFactory(config, logger, tempManager)
	return factory.CreateRecorderWithFallback()
}
