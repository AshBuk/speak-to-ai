// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package validators

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AshBuk/speak-to-ai/config/models"
)

// Inspect the configuration for invalid or unsafe values.
// It automatically corrects offending values to safe defaults and returns an error
// that aggregates all validation issues found. This ensures the application can
// always run with a sane configuration
func ValidateConfig(config *models.Config) error {
	var errors []string

	// The whisper model is fixed to a specific version for consistency
	if config.General.WhisperModel != "small-q5_1" {
		errors = append(errors, fmt.Sprintf("invalid whisper model: %s, using 'small-q5_1'", config.General.WhisperModel))
		config.General.WhisperModel = "small-q5_1"
	}

	if config.General.TempAudioPath != "" {
		// Sanitize path to prevent directory traversal
		config.General.TempAudioPath = filepath.Clean(config.General.TempAudioPath)
		if strings.Contains(config.General.TempAudioPath, "..") {
			config.General.TempAudioPath = "/tmp"
			errors = append(errors, "suspicious temp audio path sanitized to /tmp")
		}
	}

	// Audio sample rate must be within a reasonable range for audio processing
	if config.Audio.SampleRate < 8000 || config.Audio.SampleRate > 48000 {
		errors = append(errors, fmt.Sprintf("invalid sample rate: %d, correcting to 16000", config.Audio.SampleRate))
		config.Audio.SampleRate = 16000
	}

	// Ensure only supported recording methods are used
	validRecordingMethods := map[string]bool{
		"arecord": true,
		"ffmpeg":  true,
	}
	if !validRecordingMethods[config.Audio.RecordingMethod] {
		errors = append(errors, fmt.Sprintf("invalid recording method: %s, correcting to 'arecord'", config.Audio.RecordingMethod))
		config.Audio.RecordingMethod = "arecord"
	}

	// Max recording time is capped to prevent accidental resource exhaustion
	if config.Audio.MaxRecordingTime <= 0 || config.Audio.MaxRecordingTime > 1800 { // 30 minutes
		errors = append(errors, fmt.Sprintf("invalid max recording time: %d, correcting to 300s", config.Audio.MaxRecordingTime))
		config.Audio.MaxRecordingTime = 300 // 5 minutes
	}

	// Validate web server settings if it's enabled
	if config.WebServer.Enabled {
		if config.WebServer.Port <= 0 || config.WebServer.Port > 65535 {
			errors = append(errors, fmt.Sprintf("invalid port: %d, correcting to 8080", config.WebServer.Port))
			config.WebServer.Port = 8080
		}

		// Host must be a valid hostname
		if config.WebServer.Host == "" {
			config.WebServer.Host = "localhost"
		} else {
			// Basic validation to prevent injection of malicious characters
			hostRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
			if !hostRegex.MatchString(config.WebServer.Host) {
				errors = append(errors, fmt.Sprintf("invalid host: %s, correcting to 'localhost'", config.WebServer.Host))
				config.WebServer.Host = "localhost"
			}
		}
	}

	// Ensure there's always a baseline of allowed commands for security
	if len(config.Security.AllowedCommands) == 0 {
		config.Security.AllowedCommands = []string{"arecord", "ffmpeg", "whisper", "xdotool", "wl-copy", "xsel"}
		errors = append(errors, "allowed_commands was empty, populated with defaults")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation issues: %s", strings.Join(errors, "; "))
	}

	return nil
}
