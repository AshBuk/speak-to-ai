package config

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
		Device           string `yaml:"device"`
		SampleRate       int    `yaml:"sample_rate"`
		Format           string `yaml:"format"`
		Channels         int    `yaml:"channels"`
		RecordingMethod  string `yaml:"recording_method"`   // 'arecord', 'ffmpeg', 'go-native'
		ExpectedDuration int    `yaml:"expected_duration"`  // Expected recording duration in seconds
		EnableStreaming  bool   `yaml:"enable_streaming"`   // Whether to enable audio streaming
		MaxRecordingTime int    `yaml:"max_recording_time"` // Max recording time in seconds
	} `yaml:"audio"`

	// Text output settings
	Output struct {
		DefaultMode   string `yaml:"default_mode"`   // 'clipboard', 'active_window', 'web'
		ClipboardTool string `yaml:"clipboard_tool"` // 'wl-copy', 'xclip'
		TypeTool      string `yaml:"type_tool"`      // 'xdotool', 'wl-clipboard', 'dbus'
	} `yaml:"output"`

	// Web server settings
	WebServer struct {
		Enabled     bool   `yaml:"enabled"`
		Port        int    `yaml:"port"`
		Host        string `yaml:"host"`
		AuthToken   string `yaml:"auth_token"`   // Optional auth token
		APIVersion  string `yaml:"api_version"`  // API version
		LogRequests bool   `yaml:"log_requests"` // Whether to log requests
		CORSOrigins string `yaml:"cors_origins"` // Allowed origins for CORS
		MaxClients  int    `yaml:"max_clients"`  // Maximum number of clients
	} `yaml:"web_server"`

	// Security settings
	Security struct {
		AllowedCommands []string `yaml:"allowed_commands"`   // Whitelist of allowed commands
		CheckIntegrity  bool     `yaml:"check_integrity"`    // Whether to check config integrity
		ConfigHash      string   `yaml:"config_hash"`        // Hash for integrity check
		MaxTempFileSize int64    `yaml:"max_temp_file_size"` // Max temp file size in bytes
	} `yaml:"security"`
}

// LoadConfig loads configuration from file
func LoadConfig(filename string) (*Config, error) {
	var config Config

	// Set default values
	SetDefaultConfig(&config)

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

	// Validate configuration
	if err := ValidateConfig(&config); err != nil {
		log.Printf("Configuration validation error: %v", err)
		log.Println("Using validated configuration with corrections")
	}

	return &config, nil
}
