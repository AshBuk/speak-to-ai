// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package models

// Output mode constants to avoid magic strings throughout the codebase
const (
	OutputModeClipboard    = "clipboard"
	OutputModeActiveWindow = "active_window"
	OutputModeCombined     = "combined"
)

// Config structure for storing application configuration
type Config struct {
	// General settings
	General struct {
		Debug          bool     `yaml:"debug"`
		ModelPath      string   `yaml:"model_path"` // Primary model path (for backward compatibility)
		TempAudioPath  string   `yaml:"temp_audio_path"`
		ModelType      string   `yaml:"model_type"`      // 'tiny', 'base', 'small', 'medium', 'large'
		ModelPrecision string   `yaml:"model_precision"` // 'f16', 'q4_0', etc.
		Language       string   `yaml:"language"`        // Language for recognition (en, ..., etc.)
		LogFile        string   `yaml:"log_file"`        // Path to log file
		Models         []string `yaml:"models"`          // Array of available model paths
		ActiveModel    string   `yaml:"active_model"`    // Currently active model
	} `yaml:"general"`

	// Hotkey settings
	Hotkeys struct {
		// Provider override: "auto" | "dbus" | "evdev"
		Provider        string `yaml:"provider"`
		StartRecording  string `yaml:"start_recording"`
		StopRecording   string `yaml:"stop_recording"`
		ToggleStreaming string `yaml:"toggle_streaming"`
		SwitchModel     string `yaml:"switch_model"`
		ToggleVAD       string `yaml:"toggle_vad"`
		ShowConfig      string `yaml:"show_config"`
		ResetToDefaults string `yaml:"reset_to_defaults"`
	} `yaml:"hotkeys"`

	// Audio recording settings
	Audio struct {
		Device            string `yaml:"device"`
		SampleRate        int    `yaml:"sample_rate"`
		Format            string `yaml:"format"`
		RecordingMethod   string `yaml:"recording_method"`    // 'arecord', 'ffmpeg', 'go-native'
		ExpectedDuration  int    `yaml:"expected_duration"`   // Expected recording duration in seconds
		EnableStreaming   bool   `yaml:"enable_streaming"`    // Whether to enable streaming transcription
		MaxRecordingTime  int    `yaml:"max_recording_time"`  // Max recording time in seconds
		EnableVAD         bool   `yaml:"enable_vad"`          // Enable Voice Activity Detection
		VADSensitivity    string `yaml:"vad_sensitivity"`     // VAD sensitivity: 'low', 'medium', 'high'
		AutoStartStop     bool   `yaml:"auto_start_stop"`     // Auto start/stop recording based on VAD
		StreamingBufferMs int    `yaml:"streaming_buffer_ms"` // Streaming buffer size in milliseconds
	} `yaml:"audio"`

	// Text output settings
	Output struct {
		DefaultMode   string `yaml:"default_mode"`   // 'clipboard', 'active_window', 'web'
		ClipboardTool string `yaml:"clipboard_tool"` // 'wl-copy', 'xclip'
		TypeTool      string `yaml:"type_tool"`      // 'xdotool', 'wl-clipboard', 'dbus'
	} `yaml:"output"`

	// Notification settings
	Notifications struct {
		EnableWorkflowNotifications bool `yaml:"enable_workflow_notifications"` // Enable workflow notifications (recording, transcription)
	} `yaml:"notifications"`

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
