package config

// SetDefaultConfig sets default values
func SetDefaultConfig(config *Config) {
	// General settings
	config.General.Debug = false
	config.General.ModelPath = "sources/language-models/base.bin"
	config.General.TempAudioPath = "/tmp"
	config.General.ModelType = "base"
	config.General.ModelPrecision = "f16"
	config.General.Language = "auto" // Auto-detect language
	config.General.LogFile = ""      // No log file by default

	// Hotkey settings (defaults)
	config.Hotkeys.StartRecording = "alt+comma" // AltGr + , (comma)
	config.Hotkeys.StopRecording = "alt+comma"  // Same combination for start/stop
	config.Hotkeys.CopilotKey = "alt+comma"     // Also set as copilot key for Windows keyboards
	config.Hotkeys.CopyToClipboard = "ctrl+shift+c"
	config.Hotkeys.PasteToActiveApp = "ctrl+shift+v"

	// Audio settings
	config.Audio.Device = "default"
	config.Audio.SampleRate = 16000
	config.Audio.Format = "s16le"
	config.Audio.Channels = 1
	config.Audio.RecordingMethod = "arecord"
	config.Audio.ExpectedDuration = 0 // No expected duration by default
	config.Audio.EnableStreaming = false
	config.Audio.MaxRecordingTime = 300 // 5 minutes max by default

	// Output settings
	config.Output.DefaultMode = "clipboard"
	config.Output.ClipboardTool = "auto" // auto-detect
	config.Output.TypeTool = "auto"      // auto-detect

	// Web server settings
	config.WebServer.Enabled = true
	config.WebServer.Port = 8080
	config.WebServer.Host = "localhost"
	config.WebServer.AuthToken = "" // No auth by default
	config.WebServer.APIVersion = "v1"
	config.WebServer.LogRequests = true
	config.WebServer.CORSOrigins = "*" // Allow all origins by default
	config.WebServer.MaxClients = 10

	// Security settings
	config.Security.AllowedCommands = []string{"arecord", "ffmpeg", "whisper", "xdotool", "wl-copy", "xclip", "notify-send"}
	config.Security.CheckIntegrity = false
	config.Security.ConfigHash = ""
	config.Security.MaxTempFileSize = 50 * 1024 * 1024 // 50MB by default
}
