package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Config structure for storing application configuration
type Config struct {
	// General settings
	General struct {
		Debug          bool   `yaml:"debug"`
		ModelPath      string `yaml:"model_path"`
		TempAudioPath  string `yaml:"temp_audio_path"`
		ModelType      string `yaml:"model_type"`      // 'tiny', 'base', 'small', 'medium', 'large'
		ModelPrecision string `yaml:"model_precision"` // 'f16', 'q4_0', etc.
		Language       string `yaml:"language"`        // Language for recognition (en, ..., etc.)
		LogFile        string `yaml:"log_file"`        // Path to log file
	} `yaml:"general"`

	// Hotkey settings
	Hotkeys struct {
		StartRecording   string `yaml:"start_recording"`
		StopRecording    string `yaml:"stop_recording"`
		CopyToClipboard  string `yaml:"copy_to_clipboard"`
		PasteToActiveApp string `yaml:"paste_to_active_app"`
	} `yaml:"hotkeys"`

	// Audio recording settings
	Audio struct {
		Device          string `yaml:"device"`
		SampleRate      int    `yaml:"sample_rate"`
		Format          string `yaml:"format"`
		Channels        int    `yaml:"channels"`
		RecordingMethod string `yaml:"recording_method"` // 'arecord', 'ffmpeg', 'go-native'
	} `yaml:"audio"`

	// Text output settings
	Output struct {
		DefaultMode   string `yaml:"default_mode"`   // 'clipboard', 'active_window', 'web'
		ClipboardTool string `yaml:"clipboard_tool"` // 'wl-copy', 'xclip'
		TypeTool      string `yaml:"type_tool"`      // 'xdotool', 'wl-clipboard', 'dbus'
	} `yaml:"output"`

	// Web server settings
	WebServer struct {
		Enabled bool   `yaml:"enabled"`
		Port    int    `yaml:"port"`
		Host    string `yaml:"host"`
	} `yaml:"web_server"`
}

// LoadConfig loads configuration from file
func LoadConfig(filename string) (*Config, error) {
	var config Config

	// Set default values
	setDefaultConfig(&config)

	// Read configuration file
	data, err := os.ReadFile(filename)
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

	return &config, nil
}

// setDefaultConfig sets default values
func setDefaultConfig(config *Config) {
	// General settings
	config.General.Debug = false
	config.General.ModelPath = "sources/language-models/base.bin"
	config.General.TempAudioPath = "/tmp"
	config.General.ModelType = "base"
	config.General.ModelPrecision = "f16"
	config.General.Language = "auto" // Auto-detect language
	config.General.LogFile = ""      // No log file by default

	// Hotkey settings (defaults)
	config.Hotkeys.StartRecording = "ctrl+shift+space"
	config.Hotkeys.StopRecording = "ctrl+shift+space" // Same combination for start/stop
	config.Hotkeys.CopyToClipboard = "ctrl+shift+c"
	config.Hotkeys.PasteToActiveApp = "ctrl+shift+v"

	// Audio settings
	config.Audio.Device = "default"
	config.Audio.SampleRate = 16000
	config.Audio.Format = "s16le"
	config.Audio.Channels = 1
	config.Audio.RecordingMethod = "arecord"

	// Output settings
	config.Output.DefaultMode = "clipboard"
	config.Output.ClipboardTool = "auto" // auto-detect
	config.Output.TypeTool = "auto"      // auto-detect

	// Web server settings
	config.WebServer.Enabled = true
	config.WebServer.Port = 8080
	config.WebServer.Host = "localhost"
}
