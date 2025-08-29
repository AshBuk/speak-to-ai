// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/audio"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/hotkeys"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/notify"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	"github.com/AshBuk/speak-to-ai/output"
	"github.com/AshBuk/speak-to-ai/websocket"
	"github.com/AshBuk/speak-to-ai/whisper"
)

// Initialize the application components
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
			if err := a.HotkeyManager.SimulateHotkeyPress("stop_recording"); err != nil {
				a.Logger.Warning("Failed to simulate hotkey press: %v", err)
			}
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
			go func() {
				if err := a.WebSocketServer.Start(); err != nil {
					a.Logger.Info("Web interface disabled in desktop version. Integration in development: %v", err)
				}
			}()
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
