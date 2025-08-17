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
	config.Hotkeys.StartRecording = "altgr+comma" // AltGr + , (comma) - Microsoft Copilot key
	config.Hotkeys.StopRecording = "altgr+comma"  // Same combination for start/stop

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
	config.Output.DefaultMode = OutputModeActiveWindow
	config.Output.ClipboardTool = "auto" // auto-detect
	config.Output.TypeTool = "auto"      // auto-detect

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
