// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"
	"os/exec"

	"github.com/AshBuk/speak-to-ai/audio/factory"
	"github.com/AshBuk/speak-to-ai/audio/interfaces"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
	hotkeyInterfaces "github.com/AshBuk/speak-to-ai/hotkeys/interfaces"
	"github.com/AshBuk/speak-to-ai/hotkeys/manager"
	"github.com/AshBuk/speak-to-ai/internal/notify"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	"github.com/AshBuk/speak-to-ai/internal/tray"
	outputFactory "github.com/AshBuk/speak-to-ai/output/factory"
	outputInterfaces "github.com/AshBuk/speak-to-ai/output/interfaces"
	"github.com/AshBuk/speak-to-ai/output/outputters"
	"github.com/AshBuk/speak-to-ai/websocket"
	"github.com/AshBuk/speak-to-ai/whisper"
)

// ComponentFactory is responsible for creating components
type ComponentFactory struct {
	config ServiceFactoryConfig
}

func NewComponentFactory(config ServiceFactoryConfig) *ComponentFactory {
	return &ComponentFactory{config: config}
}

// Initializes all components
func (cf *ComponentFactory) InitializeComponents() (*Components, error) {
	components := &Components{}

	// Initialize model manager
	components.ModelManager = whisper.NewModelManager(cf.config.Config)
	if err := components.ModelManager.Initialize(); err != nil {
		cf.config.Logger.Warning("Failed to initialize model manager: %v", err)
	}

	// Override model path if provided
	if cf.config.ModelPath != "" {
		cf.config.Config.General.ModelPath = cf.config.ModelPath
	}

	// Ensure model is available
	if err := cf.ensureModelAvailable(components.ModelManager); err != nil {
		return nil, fmt.Errorf("failed to ensure model availability: %w", err)
	}

	// Get model file path
	modelFilePath, err := components.ModelManager.GetModelPath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve model path: %w", err)
	}
	cf.config.Logger.Info("Model path resolved: %s", modelFilePath)

	// Initialize audio recorder
	components.Recorder, err = factory.GetRecorder(cf.config.Config, cf.config.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize audio recorder: %w", err)
	}

	// Initialize whisper engine
	components.WhisperEngine, err = whisper.NewWhisperEngine(cf.config.Config, modelFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize whisper engine: %w", err)
	}

	// Initialize streaming engine if enabled
	if cf.config.Config.Audio.EnableStreaming {
		components.StreamingEngine, err = whisper.NewStreamingWhisperEngine(cf.config.Config, modelFilePath)
		if err != nil {
			cf.config.Logger.Warning("Failed to initialize streaming engine: %v", err)
			components.StreamingEngine = nil
		} else {
			cf.config.Logger.Info("Streaming transcription enabled")
		}
	}

	// Initialize output manager
	outputEnv := cf.convertEnvironmentType()
	components.OutputManager, err = outputFactory.GetOutputterFromConfig(cf.config.Config, outputEnv)
	if err != nil {
		cf.config.Logger.Warning("Failed to initialize text outputter: %v", err)
		// Fallback to clipboard only
		if fallbackOut := cf.createFallbackOutputManager(outputEnv); fallbackOut != nil {
			components.OutputManager = fallbackOut
		} else {
			return nil, fmt.Errorf("failed to initialize any output manager")
		}
	}

	// Initialize hotkey manager
	components.HotkeyManager = cf.createHotkeyManager()

	// Initialize WebSocket server (always initialized but may not be started)
	components.WebSocketServer = cf.createWebSocketServer(components.Recorder, components.WhisperEngine)

	// Initialize tray manager
	components.TrayManager = cf.createTrayManager()
	// Start tray manager (no-op in mock). Ensures systray is initialized early.
	if components.TrayManager != nil {
		components.TrayManager.Start()
		components.TrayManager.UpdateSettings(cf.config.Config)
	}

	// Initialize notification manager
	components.NotifyManager = notify.NewNotificationManager("Speak-to-AI", cf.config.Config)

	return components, nil
}

// ensureModelAvailable ensures the whisper model is available
func (cf *ComponentFactory) ensureModelAvailable(modelManager whisper.ModelManager) error {
	// Try to get the model path, which will download if needed
	_, err := modelManager.GetModelPath()
	if err != nil {
		cf.config.Logger.Info("Model not found locally, checking download...")
		return fmt.Errorf("failed to ensure model available: %w", err)
	}
	return nil
}

// convertEnvironmentType converts platform environment to output factory type
func (cf *ComponentFactory) convertEnvironmentType() outputFactory.EnvironmentType {
	switch cf.config.Environment {
	case platform.EnvironmentWayland:
		return outputFactory.EnvironmentWayland
	case platform.EnvironmentX11:
		return outputFactory.EnvironmentX11
	default:
		return outputFactory.EnvironmentX11
	}
}

// createFallbackOutputManager creates fallback clipboard-only output manager
func (cf *ComponentFactory) createFallbackOutputManager(outputEnv outputFactory.EnvironmentType) outputInterfaces.Outputter {
	clipboardTool := ""
	if outputEnv == outputFactory.EnvironmentWayland {
		if _, err := exec.LookPath("wl-copy"); err == nil {
			clipboardTool = "wl-copy"
		}
	}
	if clipboardTool == "" {
		if _, err := exec.LookPath("xclip"); err == nil {
			clipboardTool = "xclip"
		}
	}

	if clipboardTool != "" {
		cf.config.Logger.Info("Falling back to clipboard output using %s", clipboardTool)
		oldMode := cf.config.Config.Output.DefaultMode
		cf.config.Config.Output.DefaultMode = config.OutputModeClipboard
		cf.config.Config.Output.ClipboardTool = clipboardTool

		if out, err := outputters.NewClipboardOutputter(clipboardTool, cf.config.Config); err == nil {
			return out
		}

		// Restore original mode if fallback failed
		cf.config.Config.Output.DefaultMode = oldMode
	}

	return nil
}

// createHotkeyManager creates and configures hotkey manager
func (cf *ComponentFactory) createHotkeyManager() *manager.HotkeyManager {
	// Convert platform environment to hotkey interfaces environment
	var hotkeyEnv hotkeyInterfaces.EnvironmentType
	switch cf.config.Environment {
	case platform.EnvironmentWayland:
		hotkeyEnv = hotkeyInterfaces.EnvironmentWayland
	case platform.EnvironmentX11:
		hotkeyEnv = hotkeyInterfaces.EnvironmentX11
	default:
		hotkeyEnv = hotkeyInterfaces.EnvironmentX11
	}

	configAdapter := adapters.NewConfigAdapter(cf.config.Config.Hotkeys.StartRecording, cf.config.Config.Hotkeys.Provider).
		WithAdditionalHotkeys(
			cf.config.Config.Hotkeys.ToggleStreaming,
			cf.config.Config.Hotkeys.ToggleVAD,
			cf.config.Config.Hotkeys.SwitchModel,
			cf.config.Config.Hotkeys.ShowConfig,
			cf.config.Config.Hotkeys.ResetToDefaults,
		)

	return manager.NewHotkeyManager(configAdapter, hotkeyEnv)
}

// createWebSocketServer creates WebSocket server
func (cf *ComponentFactory) createWebSocketServer(recorder interfaces.AudioRecorder, whisperEngine *whisper.WhisperEngine) *websocket.WebSocketServer {
	return websocket.NewWebSocketServer(cf.config.Config, recorder, whisperEngine, cf.config.Logger)
}

// createTrayManager creates system tray manager
func (cf *ComponentFactory) createTrayManager() tray.TrayManagerInterface {
	// Create tray manager with placeholder callbacks (will be set later)
	return tray.CreateTrayManagerWithConfig(cf.config.Config,
		func() { // onExit
			cf.config.Logger.Info("Exit requested from tray")
		},
		func() error { // onToggle
			cf.config.Logger.Info("Toggle requested from tray")
			return nil
		},
		func() error { // onShowConfig
			cf.config.Logger.Info("Show config requested from tray")
			return nil
		},
		func() error { // onResetToDefaults
			cf.config.Logger.Info("Reset to defaults requested from tray")
			return nil
		})
}
