// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"log"
	"path/filepath"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	outputFactory "github.com/AshBuk/speak-to-ai/output/factory"
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

// convertEnvironmentType converts platform.EnvironmentType to outputFactory.EnvironmentType
func (a *App) convertEnvironmentType() outputFactory.EnvironmentType {
	switch a.Environment {
	case platform.EnvironmentX11:
		return outputFactory.EnvironmentX11
	case platform.EnvironmentWayland:
		return outputFactory.EnvironmentWayland
	default:
		return outputFactory.EnvironmentUnknown
	}
}
