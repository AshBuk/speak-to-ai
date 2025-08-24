package app

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/AshBuk/speak-to-ai/audio"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/hotkeys"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/notify"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	"github.com/AshBuk/speak-to-ai/internal/tray"
	"github.com/AshBuk/speak-to-ai/output"
	"github.com/AshBuk/speak-to-ai/websocket"
	"github.com/AshBuk/speak-to-ai/whisper"
)

// Initialize initializes the application and all its components
func (a *App) Initialize(configFile string, debug bool, modelPath, quantizePath string) error {
	// Set up initial logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Speak-to-AI daemon...")

	// Store config file path
	a.ConfigFile = configFile

	// Detect environment (X11 or Wayland)
	a.Environment = platform.DetectEnvironment()
	log.Printf("Detected environment: %s", a.Environment)

	// Load configuration
	var err error
	a.Config, err = config.LoadConfig(configFile)
	if err != nil {
		return err
	}

	// Override debug flag from command line if specified
	if debug {
		a.Config.General.Debug = true
	}

	// Create directories for models and temporary files
	a.ensureDirectories()

	// Initialize logger
	logConfig := logger.Config{
		Level: logger.InfoLevel,
	}

	// Set debug level if enabled
	if a.Config.General.Debug {
		logConfig.Level = logger.DebugLevel
	}

	defaultLogger, err := logger.Configure(logConfig)
	if err != nil {
		return err
	}
	a.Logger = defaultLogger

	// Initialize components (continue with initialization)
	return a.initializeComponents(modelPath)
}

// ensureDirectories creates necessary directories for the application
func (a *App) ensureDirectories() {
	// Create model directory if it doesn't exist
	modelsDir := filepath.Dir(a.Config.General.ModelPath)
	if err := platform.EnsureDirectoryExists(modelsDir); err != nil {
		log.Printf("Warning: Failed to create models directory: %v", err)
	}

	// Create temp directory if it doesn't exist
	if err := platform.EnsureDirectoryExists(a.Config.General.TempAudioPath); err != nil {
		log.Printf("Warning: Failed to create temp directory: %v", err)
	}
}

// initializeComponents initializes the application components
func (a *App) initializeComponents(modelPath string) error {
	var err error

	// Initialize model manager
	a.ModelManager = whisper.NewModelManager(a.Config)

	// Initialize the model manager to load model information
	if err := a.ModelManager.Initialize(); err != nil {
		a.Logger.Warning("Failed to initialize model manager: %v", err)
	}

	// Override model path if provided via command line
	if modelPath != "" {
		a.Config.General.ModelPath = modelPath
	}

	// Ensure model is available before initializing whisper engine
	if err := a.ensureModelAvailable(); err != nil {
		return fmt.Errorf("failed to ensure model availability: %w", err)
	}

	// Resolve concrete model file path (after ensure/download)
	modelFilePath, err := a.ModelManager.GetModelPath()
	if err != nil {
		return fmt.Errorf("failed to resolve model path: %w", err)
	}
	a.Logger.Info("Model path resolved: %s", modelFilePath)

	// Initialize audio recorder
	a.Recorder, err = audio.GetRecorder(a.Config)
	if err != nil {
		return err
	}

	// Initialize whisper engine
	a.WhisperEngine, err = whisper.NewWhisperEngine(a.Config, modelFilePath)
	if err != nil {
		return fmt.Errorf("failed to initialize whisper engine: %w", err)
	}

	// Initialize streaming engine if enabled
	if a.Config.Audio.EnableStreaming {
		a.StreamingEngine, err = whisper.NewStreamingWhisperEngine(a.Config, modelFilePath)
		if err != nil {
			a.Logger.Warning("Failed to initialize streaming engine: %v", err)
			a.StreamingEngine = nil
		} else {
			a.Logger.Info("Streaming transcription enabled")
		}
	}

	// Initialize output manager based on environment
	outputEnv := a.convertEnvironmentType()
	a.OutputManager, err = output.GetOutputterFromConfig(a.Config, outputEnv)
	if err != nil {
		a.Logger.Warning("Failed to initialize text outputter: %v", err)
	}

	// Initialize hotkey manager
	a.initializeHotkeyManager()

	// Initialize WebSocket server
	a.WebSocketServer = websocket.NewWebSocketServer(a.Config, a.Recorder, a.WhisperEngine, a.Logger)

	// Initialize notification manager
	a.NotifyManager = notify.NewNotificationManager("Speak-to-AI", a.Config)

	// Initialize system tray
	a.initializeTrayManager()

	return nil
}

// reinitializeComponents reinitializes components that depend on configuration
func (a *App) reinitializeComponents(oldConfig *config.Config) error {
	var err error

	// Check if debug level changed
	if oldConfig.General.Debug != a.Config.General.Debug {
		// Update logger level
		logConfig := logger.Config{
			Level: logger.InfoLevel,
		}
		if a.Config.General.Debug {
			logConfig.Level = logger.DebugLevel
		}

		newLogger, err := logger.Configure(logConfig)
		if err != nil {
			return fmt.Errorf("failed to reconfigure logger: %w", err)
		}
		a.Logger = newLogger
		a.Logger.Info("Logger reconfigured with debug=%v", a.Config.General.Debug)
	}

	// Reinitialize HotkeyManager if hotkey settings changed
	if oldConfig.Hotkeys.StartRecording != a.Config.Hotkeys.StartRecording ||
		oldConfig.Hotkeys.StopRecording != a.Config.Hotkeys.StopRecording {

		a.Logger.Info("Hotkey settings changed, reinitializing hotkey manager...")

		// Stop old hotkey manager
		if a.HotkeyManager != nil {
			a.HotkeyManager.Stop()
		}

		// Initialize new hotkey manager
		a.initializeHotkeyManager()

		// Register callbacks and start
		a.registerCallbacks()
		if err := a.HotkeyManager.Start(); err != nil {
			a.Logger.Warning("Failed to start hotkey manager: %v", err)
		} else {
			a.Logger.Info("Hotkey manager reinitialized successfully")
		}
	}

	// Reinitialize AudioRecorder if audio settings changed
	if oldConfig.Audio.Device != a.Config.Audio.Device ||
		oldConfig.Audio.SampleRate != a.Config.Audio.SampleRate ||
		oldConfig.Audio.RecordingMethod != a.Config.Audio.RecordingMethod ||
		oldConfig.Audio.Channels != a.Config.Audio.Channels {

		a.Logger.Info("Audio settings changed, reinitializing audio recorder...")

		// Stop current recording if active
		if a.HotkeyManager != nil && a.HotkeyManager.IsRecording() {
			a.Logger.Warning("Stopping active recording for audio reconfiguration")
			a.HotkeyManager.SimulateHotkeyPress("stop_recording")
		}

		// Reinitialize audio recorder
		a.Recorder, err = audio.GetRecorder(a.Config)
		if err != nil {
			return fmt.Errorf("failed to reinitialize audio recorder: %w", err)
		}
		a.Logger.Info("Audio recorder reinitialized successfully")
	}

	// Reinitialize OutputManager if output settings changed
	if oldConfig.Output.DefaultMode != a.Config.Output.DefaultMode ||
		oldConfig.Output.ClipboardTool != a.Config.Output.ClipboardTool ||
		oldConfig.Output.TypeTool != a.Config.Output.TypeTool {

		a.Logger.Info("Output settings changed, reinitializing output manager...")

		outputEnv := a.convertEnvironmentType()
		a.OutputManager, err = output.GetOutputterFromConfig(a.Config, outputEnv)
		if err != nil {
			a.Logger.Warning("Failed to reinitialize output manager: %v", err)
		} else {
			a.Logger.Info("Output manager reinitialized successfully")
		}
	}

	// Reinitialize WebSocket server if settings changed
	if oldConfig.WebServer.Enabled != a.Config.WebServer.Enabled ||
		oldConfig.WebServer.Port != a.Config.WebServer.Port ||
		oldConfig.WebServer.Host != a.Config.WebServer.Host {

		a.Logger.Info("WebSocket server settings changed, reinitializing...")

		// Stop old server
		if a.WebSocketServer != nil {
			a.WebSocketServer.Stop()
		}

		// Create new server
		a.WebSocketServer = websocket.NewWebSocketServer(a.Config, a.Recorder, a.WhisperEngine, a.Logger)

		// Start new server if enabled
		if a.Config.WebServer.Enabled {
			go a.WebSocketServer.Start()
		}
		a.Logger.Info("WebSocket server reinitialized successfully")
	}

	// Update tray settings display
	if a.TrayManager != nil {
		a.TrayManager.UpdateSettings(a.Config)
		a.Logger.Info("Tray settings updated")
	}

	return nil
}

// registerCallbacks registers all the hotkey callbacks (moved from shutdown.go for reuse)
func (a *App) registerCallbacks() {
	a.HotkeyManager.RegisterCallbacks(
		// Record start callback
		a.handleStartRecording,
		// Record stop and transcribe callback
		a.handleStopRecordingAndTranscribe,
	)

	// Register additional hotkeys
	if a.Config.Hotkeys.ToggleStreaming != "" {
		a.HotkeyManager.RegisterHotkeyAction(a.Config.Hotkeys.ToggleStreaming, a.handleToggleStreaming)
	}
	if a.Config.Hotkeys.SwitchModel != "" {
		a.HotkeyManager.RegisterHotkeyAction(a.Config.Hotkeys.SwitchModel, a.handleSwitchModel)
	}
	if a.Config.Hotkeys.ToggleVAD != "" {
		a.HotkeyManager.RegisterHotkeyAction(a.Config.Hotkeys.ToggleVAD, a.handleToggleVAD)
	}
	if a.Config.Hotkeys.ShowConfig != "" {
		a.HotkeyManager.RegisterHotkeyAction(a.Config.Hotkeys.ShowConfig, a.handleShowConfig)
	}
	if a.Config.Hotkeys.ReloadConfig != "" {
		a.HotkeyManager.RegisterHotkeyAction(a.Config.Hotkeys.ReloadConfig, a.handleReloadConfig)
	}
}

// ensureModelAvailable ensures the model is available, downloading if necessary
func (a *App) ensureModelAvailable() error {
	// Check if model already exists
	if _, err := a.ModelManager.GetModelPath(); err == nil {
		return nil
	}

	// Model doesn't exist, need to download
	a.Logger.Info("Model not found, downloading...")

	// Show notification about download starting
	if a.NotifyManager != nil {
		a.NotifyManager.ShowNotification("Speak-to-AI", "Downloading Whisper model for first use...")
	}

	// Create progress callback
	progressCallback := func(downloaded, total int64, percentage float64) {
		// Format sizes for display
		downloadedMB := float64(downloaded) / (1024 * 1024)
		totalMB := float64(total) / (1024 * 1024)

		// Log progress
		a.Logger.Info("Download progress: %.1f%% (%.1f MB / %.1f MB)",
			percentage, downloadedMB, totalMB)

		// Update tray tooltip if available
		if a.TrayManager != nil {
			status := fmt.Sprintf("📥 Downloading model: %.1f%%", percentage)
			a.TrayManager.SetTooltip(status)
		}
	}

	// Download with progress
	modelPath, err := a.ModelManager.GetModelPathWithProgress(progressCallback)
	if err != nil {
		// Reset tray tooltip on error
		if a.TrayManager != nil {
			a.TrayManager.SetTooltip("❌ Model download failed")
		}
		return fmt.Errorf("failed to download model: %w", err)
	}

	// Show completion notification
	if a.NotifyManager != nil {
		a.NotifyManager.ShowNotification("Speak-to-AI", "Model downloaded successfully!")
	}

	// Reset tray tooltip
	if a.TrayManager != nil {
		a.TrayManager.SetTooltip("✅ Ready")
	}

	a.Logger.Info("Model downloaded successfully: %s", modelPath)
	return nil
}

// initializeTrayManager initializes the system tray manager
func (a *App) initializeTrayManager() {
	// Create a toggle function for the tray
	toggleFunc := func() error {
		if a.HotkeyManager.IsRecording() {
			return a.HotkeyManager.SimulateHotkeyPress("stop_recording")
		}
		return a.HotkeyManager.SimulateHotkeyPress("start_recording")
	}

	// Create exit function
	exitFunc := func() {
		a.Cancel() // Trigger application shutdown
	}

	// Create show config function
	showConfigFunc := func() error {
		return a.handleShowConfig()
	}

	// Create reload config function
	reloadConfigFunc := func() error {
		return a.handleReloadConfig()
	}

	// Create the appropriate tray manager with configuration
	a.TrayManager = tray.CreateTrayManagerWithConfig(a.Config, exitFunc, toggleFunc, showConfigFunc, reloadConfigFunc)
}

// convertEnvironmentType converts platform.EnvironmentType to output.EnvironmentType
func (a *App) convertEnvironmentType() output.EnvironmentType {
	switch a.Environment {
	case platform.EnvironmentX11:
		return output.EnvironmentX11
	case platform.EnvironmentWayland:
		return output.EnvironmentWayland
	default:
		return output.EnvironmentUnknown
	}
}

// initializeHotkeyManager initializes the hotkey manager component
func (a *App) initializeHotkeyManager() {
	// Convert platform.EnvironmentType to hotkeys.EnvironmentType
	var hotkeyEnv hotkeys.EnvironmentType
	switch a.Environment {
	case platform.EnvironmentX11:
		hotkeyEnv = hotkeys.EnvironmentX11
	case platform.EnvironmentWayland:
		hotkeyEnv = hotkeys.EnvironmentWayland
	default:
		hotkeyEnv = hotkeys.EnvironmentUnknown
	}

	// Create hotkey config adapter
	hotkeyConfig := hotkeys.NewConfigAdapter(
		a.Config.Hotkeys.StartRecording,
	)

	// Initialize hotkey manager with environment information
	a.HotkeyManager = hotkeys.NewHotkeyManager(hotkeyConfig, hotkeyEnv)
}
