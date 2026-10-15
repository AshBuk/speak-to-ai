// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package factory

import (
	"fmt"
	"os/exec"

	"github.com/AshBuk/dabri/v2/audio/interfaces"
	"github.com/AshBuk/dabri/v2/audio/processing"
	"github.com/AshBuk/dabri/v2/audio/recorders"
	"github.com/AshBuk/dabri/v2/config"
	"github.com/AshBuk/dabri/v2/internal/logger"
)

// AudioRecorderFactory creates audio recorder instances based on configuration
// Specialized subfactory used by services.NewContainer
//
// Factory Hierarchy:
//
//	services.NewContainer (internal/services/factory.go)
//	    │
//	    └── uses → AudioRecorderFactory (this file)
//	                  │
//	                  └── creates → ArecordRecorder or FFmpegRecorder
//
// Design:
//   - Factory Method: creates different recorder implementations (arecord, ffmpeg)
//   - Strategy: recorders implement common AudioRecorder interface
//   - Diagnostics Pattern: pre-creation system validation
//
// Usage:
//
//	GetRecorder(config, logger, tempManager) // simple creation
type AudioRecorderFactory struct {
	config      *config.Config
	logger      logger.Logger
	tempManager *processing.TempFileManager
}

// NewAudioRecorderFactory Constructor - initializes factory with required dependencies
func NewAudioRecorderFactory(config *config.Config, logger logger.Logger, tempManager *processing.TempFileManager) *AudioRecorderFactory {
	return &AudioRecorderFactory{ // Injected:
		config:      config,      // (recorder selection)
		logger:      logger,      // (diagnostics)
		tempManager: tempManager, // (file handling)
	}
}

// CreateRecorder Factory Method implementation - creates recorder instance from config
// Returns AudioRecorder interface with arecord or ffmpeg concrete implementation
// Runs diagnostics before creation to aid troubleshooting if recorder fails later
func (f *AudioRecorderFactory) CreateRecorder() (interfaces.AudioRecorder, error) {
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

// DiagnoseAudioSystem Pre-creation validation - checks tools and devices before recording
// Logs warnings for missing tools (ffmpeg/arecord) and available audio devices
// Helps troubleshoot "recorder created but fails to record" scenarios
func (f *AudioRecorderFactory) DiagnoseAudioSystem() {
	f.logger.Info("[AUDIO DIAGNOSTICS] Recording method: %s, Device: %s",
		f.config.Audio.RecordingMethod, f.config.Audio.Device)
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

// GetRecorder Convenience function - one-line recorder creation without factory instance
// Calls CreateRecorder() - no fallback, fails if configured method unavailable
func GetRecorder(config *config.Config, logger logger.Logger, tempManager *processing.TempFileManager) (interfaces.AudioRecorder, error) {
	factory := NewAudioRecorderFactory(config, logger, tempManager)
	return factory.CreateRecorder()
}
