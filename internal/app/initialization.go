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
func (a *App) Initialize(configFile string, debug bool, whisperPath, modelPath, quantizePath string) error {
	// Set up initial logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Speak-to-AI daemon...")

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
	return a.initializeComponents(whisperPath, modelPath)
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
func (a *App) initializeComponents(whisperPath, modelPath string) error {
	var err error

	// Initialize model manager
	a.ModelManager = whisper.NewModelManager(a.Config)

	// Override model path if provided via command line
	if modelPath != "" {
		a.Config.General.ModelPath = modelPath
	}

	// Get model path from config (no blocking download)
	modelFilePath := a.Config.General.ModelPath
	a.Logger.Info("Model path configured: %s", modelFilePath)

	// Initialize audio recorder
	a.Recorder, err = audio.GetRecorder(a.Config)
	if err != nil {
		return err
	}

	// Initialize whisper engine
	a.WhisperEngine = whisper.NewWhisperEngine(a.Config, whisperPath, modelFilePath)

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
	a.NotifyManager = notify.NewNotificationManager("Speak-to-AI")

	// Initialize system tray
	a.initializeTrayManager()

	return nil
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

	// Create the appropriate tray manager
	a.TrayManager = tray.CreateDefaultTrayManager(exitFunc, toggleFunc, showConfigFunc, reloadConfigFunc)
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
