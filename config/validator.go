package config

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateConfig validates the configuration and corrects issues
func ValidateConfig(config *Config) error {
	var errors []string

	// Validate general settings
	if config.General.ModelPath != "" {
		// Sanitize path to avoid path traversal
		config.General.ModelPath = filepath.Clean(config.General.ModelPath)
		if strings.Contains(config.General.ModelPath, "..") {
			// Replace with default value if suspicious
			config.General.ModelPath = "sources/language-models/base.bin"
			errors = append(errors, "suspicious model path sanitized")
		}
	}

	if config.General.TempAudioPath != "" {
		// Sanitize path
		config.General.TempAudioPath = filepath.Clean(config.General.TempAudioPath)
		if strings.Contains(config.General.TempAudioPath, "..") {
			config.General.TempAudioPath = "/tmp"
			errors = append(errors, "suspicious temp audio path sanitized")
		}
	}

	// Validate model type
	validModelTypes := map[string]bool{
		"tiny":   true,
		"base":   true,
		"small":  true,
		"medium": true,
		"large":  true,
	}
	if !validModelTypes[config.General.ModelType] {
		config.General.ModelType = "base"
		errors = append(errors, fmt.Sprintf("invalid model type: %s, using 'base'", config.General.ModelType))
	}

	// Validate audio settings
	if config.Audio.SampleRate < 8000 || config.Audio.SampleRate > 48000 {
		config.Audio.SampleRate = 16000
		errors = append(errors, fmt.Sprintf("invalid sample rate: %d, using 16000", config.Audio.SampleRate))
	}

	if config.Audio.Channels < 1 || config.Audio.Channels > 2 {
		config.Audio.Channels = 1
		errors = append(errors, fmt.Sprintf("invalid channels: %d, using 1", config.Audio.Channels))
	}

	validRecordingMethods := map[string]bool{
		"arecord": true,
		"ffmpeg":  true,
	}
	if !validRecordingMethods[config.Audio.RecordingMethod] {
		config.Audio.RecordingMethod = "arecord"
		errors = append(errors, fmt.Sprintf("invalid recording method: %s, using 'arecord'", config.Audio.RecordingMethod))
	}

	// Validate max recording time
	if config.Audio.MaxRecordingTime <= 0 || config.Audio.MaxRecordingTime > 1800 {
		// Set to 5 minutes if invalid (too short or too long)
		config.Audio.MaxRecordingTime = 300
		errors = append(errors, fmt.Sprintf("invalid max recording time: %d, using 300", config.Audio.MaxRecordingTime))
	}

	// Validate web server settings
	if config.WebServer.Enabled {
		if config.WebServer.Port <= 0 || config.WebServer.Port > 65535 {
			config.WebServer.Port = 8080
			errors = append(errors, fmt.Sprintf("invalid port: %d, using 8080", config.WebServer.Port))
		}

		// Validate host
		if config.WebServer.Host == "" {
			config.WebServer.Host = "localhost"
		} else {
			// Basic host validation - allow only alphanumeric, dot, hyphen
			hostRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
			if !hostRegex.MatchString(config.WebServer.Host) {
				config.WebServer.Host = "localhost"
				errors = append(errors, fmt.Sprintf("invalid host: %s, using 'localhost'", config.WebServer.Host))
			}
		}
	}

	// Validate security settings
	if len(config.Security.AllowedCommands) == 0 {
		// Set default allowed commands
		config.Security.AllowedCommands = []string{"arecord", "ffmpeg", "whisper", "xdotool", "wl-copy", "xclip"}
	}

	// Return combined errors or nil if none
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation issues: %s", strings.Join(errors, "; "))
	}

	return nil
}

// IsCommandAllowed checks if a command is in the allowed list
func (c *Config) IsCommandAllowed(command string) bool {
	// Extract base command name
	base := filepath.Base(command)

	// Check if it's in allowed list
	for _, cmd := range c.Security.AllowedCommands {
		if cmd == base {
			return true
		}
	}

	return false
}

// SanitizeCommandArgs removes potentially dangerous arguments
func SanitizeCommandArgs(args []string) []string {
	sanitized := make([]string, 0, len(args))

	for _, arg := range args {
		// Filter out shell metacharacters and other dangerous constructs
		if !strings.ContainsAny(arg, "&|;$<>(){}[]") && !strings.Contains(arg, "..") {
			sanitized = append(sanitized, arg)
		}
	}

	return sanitized
}
