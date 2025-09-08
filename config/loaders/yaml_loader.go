// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package loaders

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AshBuk/speak-to-ai/config/models"
	"github.com/AshBuk/speak-to-ai/config/validators"
	yaml "gopkg.in/yaml.v2"
)

// LoadConfig loads configuration from file
func LoadConfig(filename string) (*models.Config, error) {
	var config models.Config

	// Set default values
	SetDefaultConfig(&config)

	// Read configuration file
	// Sanitize and validate path
	clean := filepath.Clean(filename)
	if strings.Contains(clean, "..") {
		return nil, fmt.Errorf("invalid config path: %s", filename)
	}
	// #nosec G304 -- Safe: path is sanitized and controlled by application configuration.
	data, err := os.ReadFile(clean)
	if err != nil {
		log.Printf("Warning: could not read config file: %v", err)
		log.Println("Using default configuration")
		return &config, nil
	}

	// Parse YAML
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Validate configuration
	if err := validators.ValidateConfig(&config); err != nil {
		log.Printf("Configuration validation error: %v", err)
		log.Println("Using validated configuration with corrections")
	}

	return &config, nil
}

// SetDefaultConfig sets default values
func SetDefaultConfig(config *models.Config) {
	// General settings
	config.General.Debug = false
	config.General.ModelPath = "sources/language-models/base.bin" // Backward compatibility
	config.General.TempAudioPath = "/tmp"
	config.General.ModelType = "base"
	config.General.ModelPrecision = "f16"
	config.General.Language = "auto" // Auto-detect language
	config.General.LogFile = ""      // No log file by default

	// Multiple models support
	config.General.Models = []string{
		"sources/language-models/base.bin",
		"sources/language-models/small.bin",
		"sources/language-models/tiny.bin",
	}
	config.General.ActiveModel = "sources/language-models/base.bin" // Default to base model

	// Hotkey settings (defaults)
	config.Hotkeys.StartRecording = "ctrl+alt+r"     // Universal default
	config.Hotkeys.StopRecording = "ctrl+alt+r"      // Same combination for start/stop
	config.Hotkeys.ToggleStreaming = "altgr+shift+s" // AltGr + Shift + S
	config.Hotkeys.SwitchModel = "altgr+shift+m"     // AltGr + Shift + M
	config.Hotkeys.ToggleVAD = "altgr+shift+v"       // AltGr + Shift + V
	config.Hotkeys.ShowConfig = "altgr+shift+c"      // AltGr + Shift + C
	config.Hotkeys.ReloadConfig = "altgr+shift+r"    // AltGr + Shift + R

	// Audio settings
	config.Audio.Device = "default"
	config.Audio.SampleRate = 16000
	config.Audio.Format = "s16le"
	config.Audio.RecordingMethod = "arecord"
	config.Audio.ExpectedDuration = 0 // No expected duration by default
	config.Audio.EnableStreaming = false
	config.Audio.MaxRecordingTime = 300    // 5 minutes max by default
	config.Audio.EnableVAD = false         // VAD disabled by default for compatibility
	config.Audio.VADSensitivity = "medium" // Balanced VAD sensitivity
	config.Audio.AutoStartStop = false     // Manual control by default
	config.Audio.StreamingBufferMs = 1000  // 1 second streaming buffer

	// Output settings
	config.Output.DefaultMode = models.OutputModeActiveWindow
	config.Output.ClipboardTool = "auto" // auto-detect
	config.Output.TypeTool = "auto"      // auto-detect

	// Notification settings
	config.Notifications.EnableWorkflowNotifications = true // Enable workflow notifications by default

	// Web server settings
	config.WebServer.Enabled = false
	config.WebServer.Port = 8080
	config.WebServer.Host = "localhost"
	config.WebServer.AuthToken = "" // No auth by default
	config.WebServer.APIVersion = "v1"
	config.WebServer.LogRequests = true
	config.WebServer.CORSOrigins = "*" // Allow all origins by default
	config.WebServer.MaxClients = 10

	// Security settings
	config.Security.AllowedCommands = []string{"arecord", "ffmpeg", "whisper", "xdotool", "wtype", "ydotool", "wl-copy", "wl-paste", "xclip", "notify-send", "xdg-open"}
	config.Security.CheckIntegrity = false
	config.Security.ConfigHash = ""
	config.Security.MaxTempFileSize = 50 * 1024 * 1024 // 50MB by default
}

// SaveConfig writes the configuration back to disk in YAML format
func SaveConfig(filename string, config *models.Config) error {
	// Sanitize and validate path
	safe := filepath.Clean(filename)
	if strings.Contains(safe, "..") {
		return fmt.Errorf("invalid config path: %s", filename)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(safe), 0o750); err != nil {
		return err
	}

	// Write with restrictive permissions
	return os.WriteFile(safe, data, 0o600)
}
