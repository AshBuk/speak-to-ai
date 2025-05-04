package app

import (
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

	// Get model path from manager for immediate usage
	modelFilePath, err := a.ModelManager.GetModelPath()
	if err != nil {
		a.Logger.Warning("Failed to get model path: %v", err)
		a.Logger.Info("Will continue startup and attempt to load model later")
	}

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

	// Create the appropriate tray manager
	a.TrayManager = tray.CreateDefaultTrayManager(exitFunc, toggleFunc)
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
		a.Config.Hotkeys.CopyToClipboard,
		a.Config.Hotkeys.PasteToActiveApp,
	)

	// Initialize hotkey manager with environment information
	a.HotkeyManager = hotkeys.NewHotkeyManager(hotkeyConfig, hotkeyEnv)
}
