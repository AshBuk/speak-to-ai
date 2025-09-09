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
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/notify"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	"github.com/AshBuk/speak-to-ai/internal/tray"
	outputFactory "github.com/AshBuk/speak-to-ai/output/factory"
	outputInterfaces "github.com/AshBuk/speak-to-ai/output/interfaces"
	"github.com/AshBuk/speak-to-ai/output/outputters"
	"github.com/AshBuk/speak-to-ai/websocket"
	"github.com/AshBuk/speak-to-ai/whisper"
)

// ServiceFactoryConfig holds all dependencies needed to create services
type ServiceFactoryConfig struct {
	Logger      logger.Logger
	Config      *config.Config
	ConfigFile  string
	Environment platform.EnvironmentType
	ModelPath   string
}

// Components holds all initialized application components
type Components struct {
	ModelManager    whisper.ModelManager
	Recorder        interfaces.AudioRecorder
	WhisperEngine   *whisper.WhisperEngine
	StreamingEngine *whisper.StreamingWhisperEngine
	OutputManager   outputInterfaces.Outputter
	HotkeyManager   *manager.HotkeyManager
	WebSocketServer *websocket.WebSocketServer
	TrayManager     tray.TrayManagerInterface
	NotifyManager   *notify.NotificationManager
}

// ServiceFactory creates and configures all services with proper dependency injection
type ServiceFactory struct {
	config ServiceFactoryConfig
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(config ServiceFactoryConfig) *ServiceFactory {
	return &ServiceFactory{
		config: config,
	}
}

// CreateServices creates and configures all services
func (sf *ServiceFactory) CreateServices() (*ServiceContainer, error) {
	container := NewServiceContainer()

	// Create all components first
	components, err := sf.initializeComponents()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	// Create ConfigService and HotkeyService
	configService := sf.createConfigService()
	container.Config = configService

	hotkeyService := sf.createHotkeyService(components.HotkeyManager)
	container.Hotkeys = hotkeyService

	// Create AudioService
	audioService := sf.createAudioService(components)
	container.Audio = audioService

	// Create UIService
	uiService := sf.createUIService(components.TrayManager, components.NotifyManager)
	container.UI = uiService

	// Create IOService
	ioService := sf.createIOService(components.OutputManager, components.WebSocketServer)
	container.IO = ioService

	return container, nil
}

// initializeComponents initializes all application components
func (sf *ServiceFactory) initializeComponents() (*Components, error) {
	components := &Components{}

	// Initialize model manager
	components.ModelManager = whisper.NewModelManager(sf.config.Config)
	if err := components.ModelManager.Initialize(); err != nil {
		sf.config.Logger.Warning("Failed to initialize model manager: %v", err)
	}

	// Override model path if provided
	if sf.config.ModelPath != "" {
		sf.config.Config.General.ModelPath = sf.config.ModelPath
	}

	// Ensure model is available
	if err := sf.ensureModelAvailable(components.ModelManager); err != nil {
		return nil, fmt.Errorf("failed to ensure model availability: %w", err)
	}

	// Get model file path
	modelFilePath, err := components.ModelManager.GetModelPath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve model path: %w", err)
	}
	sf.config.Logger.Info("Model path resolved: %s", modelFilePath)

	// Initialize audio recorder
	components.Recorder, err = factory.GetRecorder(sf.config.Config, sf.config.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize audio recorder: %w", err)
	}

	// Initialize whisper engine
	components.WhisperEngine, err = whisper.NewWhisperEngine(sf.config.Config, modelFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize whisper engine: %w", err)
	}

	// Initialize streaming engine if enabled
	if sf.config.Config.Audio.EnableStreaming {
		components.StreamingEngine, err = whisper.NewStreamingWhisperEngine(sf.config.Config, modelFilePath)
		if err != nil {
			sf.config.Logger.Warning("Failed to initialize streaming engine: %v", err)
			components.StreamingEngine = nil
		} else {
			sf.config.Logger.Info("Streaming transcription enabled")
		}
	}

	// Initialize output manager
	outputEnv := sf.convertEnvironmentType()
	components.OutputManager, err = outputFactory.GetOutputterFromConfig(sf.config.Config, outputEnv)
	if err != nil {
		sf.config.Logger.Warning("Failed to initialize text outputter: %v", err)
		// Fallback to clipboard only
		if fallbackOut := sf.createFallbackOutputManager(outputEnv); fallbackOut != nil {
			components.OutputManager = fallbackOut
		} else {
			return nil, fmt.Errorf("failed to initialize any output manager")
		}
	}

	// Initialize hotkey manager
	components.HotkeyManager = sf.createHotkeyManager()

	// Initialize WebSocket server (always initialized but may not be started)
	components.WebSocketServer = sf.createWebSocketServer(components.Recorder, components.WhisperEngine)

	// Initialize tray manager
	components.TrayManager = sf.createTrayManager()

	// Initialize notification manager
	components.NotifyManager = notify.NewNotificationManager("Speak-to-AI", sf.config.Config)

	return components, nil
}

// createConfigService creates a ConfigService instance for configuration management only
func (sf *ServiceFactory) createConfigService() *ConfigService {
	return NewConfigService(
		sf.config.Logger,
		sf.config.Config,
		sf.config.ConfigFile,
	)
}

// createHotkeyService creates a HotkeyService instance for hotkey management
func (sf *ServiceFactory) createHotkeyService(hotkeyManager *manager.HotkeyManager) *HotkeyService {
	return NewHotkeyService(
		sf.config.Logger,
		hotkeyManager,
	)
}

// createAudioService creates an AudioService instance
func (sf *ServiceFactory) createAudioService(components *Components) *AudioService {
	return NewAudioService(
		sf.config.Logger,
		sf.config.Config,
		components.Recorder,
		components.WhisperEngine,
		components.StreamingEngine,
		components.ModelManager,
	)
}

// createUIService creates a UIService instance
func (sf *ServiceFactory) createUIService(trayManager tray.TrayManagerInterface, notifyManager *notify.NotificationManager) *UIService {
	return NewUIService(
		sf.config.Logger,
		trayManager,
		notifyManager,
	)
}

// createIOService creates an IOService instance
func (sf *ServiceFactory) createIOService(outputManager outputInterfaces.Outputter, webSocketServer *websocket.WebSocketServer) *IOService {
	return NewIOService(
		sf.config.Logger,
		outputManager,
		webSocketServer,
	)
}

// Helper methods for component initialization

// ensureModelAvailable ensures the whisper model is available
func (sf *ServiceFactory) ensureModelAvailable(modelManager whisper.ModelManager) error {
	// Try to get the model path, which will download if needed
	_, err := modelManager.GetModelPath()
	if err != nil {
		sf.config.Logger.Info("Model not found locally, checking download...")
		return fmt.Errorf("failed to ensure model available: %w", err)
	}
	return nil
}

// convertEnvironmentType converts platform environment to output factory type
func (sf *ServiceFactory) convertEnvironmentType() outputFactory.EnvironmentType {
	switch sf.config.Environment {
	case platform.EnvironmentWayland:
		return outputFactory.EnvironmentWayland
	case platform.EnvironmentX11:
		return outputFactory.EnvironmentX11
	default:
		return outputFactory.EnvironmentX11
	}
}

// createFallbackOutputManager creates fallback clipboard-only output manager
func (sf *ServiceFactory) createFallbackOutputManager(outputEnv outputFactory.EnvironmentType) outputInterfaces.Outputter {
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
		sf.config.Logger.Info("Falling back to clipboard output using %s", clipboardTool)
		oldMode := sf.config.Config.Output.DefaultMode
		sf.config.Config.Output.DefaultMode = config.OutputModeClipboard
		sf.config.Config.Output.ClipboardTool = clipboardTool

		if out, err := outputters.NewClipboardOutputter(clipboardTool, sf.config.Config); err == nil {
			return out
		}

		// Restore original mode if fallback failed
		sf.config.Config.Output.DefaultMode = oldMode
	}

	return nil
}

// createHotkeyManager creates and configures hotkey manager
func (sf *ServiceFactory) createHotkeyManager() *manager.HotkeyManager {
	// Convert platform environment to hotkey interfaces environment
	var hotkeyEnv hotkeyInterfaces.EnvironmentType
	switch sf.config.Environment {
	case platform.EnvironmentWayland:
		hotkeyEnv = hotkeyInterfaces.EnvironmentWayland
	case platform.EnvironmentX11:
		hotkeyEnv = hotkeyInterfaces.EnvironmentX11
	default:
		hotkeyEnv = hotkeyInterfaces.EnvironmentX11
	}

	configAdapter := adapters.NewConfigAdapter(sf.config.Config.Hotkeys.StartRecording, sf.config.Config.Hotkeys.Provider)

	return manager.NewHotkeyManager(configAdapter, hotkeyEnv)
}

// createWebSocketServer creates WebSocket server
func (sf *ServiceFactory) createWebSocketServer(recorder interfaces.AudioRecorder, whisperEngine *whisper.WhisperEngine) *websocket.WebSocketServer {
	return websocket.NewWebSocketServer(sf.config.Config, recorder, whisperEngine, sf.config.Logger)
}

// createTrayManager creates system tray manager
func (sf *ServiceFactory) createTrayManager() tray.TrayManagerInterface {
	// Create tray manager with placeholder callbacks (will be set later)
	return tray.CreateTrayManagerWithConfig(sf.config.Config,
		func() { // onExit
			sf.config.Logger.Info("Exit requested from tray")
		},
		func() error { // onToggle
			sf.config.Logger.Info("Toggle requested from tray")
			return nil
		},
		func() error { // onShowConfig
			sf.config.Logger.Info("Show config requested from tray")
			return nil
		},
		func() error { // onReloadConfig
			sf.config.Logger.Info("Reload config requested from tray")
			return nil
		})
}

// SetupServiceDependencies configures cross-service dependencies
func (sf *ServiceFactory) SetupServiceDependencies(container *ServiceContainer) {
	// Set up AudioService dependencies
	if audioSvc, ok := container.Audio.(*AudioService); ok {
		audioSvc.SetDependencies(container.UI, container.IO)
	}

	// Additional cross-dependencies can be set up here as needed
}
