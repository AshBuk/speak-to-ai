// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	"github.com/AshBuk/speak-to-ai/internal/services"
)

// RuntimeContext manages application lifecycle and context
type RuntimeContext struct {
	Ctx        context.Context
	Cancel     context.CancelFunc
	ShutdownCh chan os.Signal
	Logger     logger.Logger
}

// NewRuntimeContext creates a new runtime context
func NewRuntimeContext(logger logger.Logger) *RuntimeContext {
	ctx, cancel := context.WithCancel(context.Background())
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

	return &RuntimeContext{
		Ctx:        ctx,
		Cancel:     cancel,
		ShutdownCh: shutdownCh,
		Logger:     logger,
	}
}

// App represents the application with service-based architecture
type App struct {
	Services *services.ServiceContainer
	Runtime  *RuntimeContext
}

// NewApp creates a new application instance
func NewApp(logger logger.Logger) *App {
	return &App{
		Services: services.NewServiceContainer(),
		Runtime:  NewRuntimeContext(logger),
	}
}

// Initialize sets up all services with dependency injection
func (a *App) Initialize(configFile string, debug bool, modelPath, quantizePath string) error {
	a.Runtime.Logger.Info("Initializing application...")

	// Load configuration
	config, err := a.initializeConfig(configFile, debug)
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Initialize services with dependency injection
	if err := a.initializeServices(config, modelPath, configFile); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	// Set up service cross-dependencies
	a.setupServiceDependencies()

	a.Runtime.Logger.Info("Application initialization complete")
	return nil
}

// initializeConfig loads and validates configuration
func (a *App) initializeConfig(configFile string, debug bool) (*config.Config, error) {
	a.Runtime.Logger.Info("Loading configuration from: %s", configFile)

	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Override debug flag from command line if specified
	if debug {
		cfg.General.Debug = true
	}

	a.Runtime.Logger.Info("Configuration loaded successfully")
	return cfg, nil
}

// initializeServices creates and configures all services
func (a *App) initializeServices(cfg *config.Config, modelPath string, cfgFilePath string) error {
	a.Runtime.Logger.Info("Initializing services with dependency injection...")

	// Detect environment
	environment := platform.DetectEnvironment()
	a.Runtime.Logger.Info("Detected environment: %s", environment)

	// Create service factory
	factory := services.NewServiceFactory(services.ServiceFactoryConfig{
		Logger:      a.Runtime.Logger,
		Config:      cfg,
		ConfigFile:  cfgFilePath,
		Environment: environment,
		ModelPath:   modelPath,
	})

	// Create services
	serviceContainer, err := factory.CreateServices()
	if err != nil {
		return fmt.Errorf("failed to create services: %w", err)
	}

	a.Services = serviceContainer

	// Set up hotkey callbacks
	if err := a.setupHotkeyCallbacks(); err != nil {
		a.Runtime.Logger.Warning("Failed to set up hotkey callbacks: %v", err)
	}

	a.Runtime.Logger.Info("Services initialized successfully")
	return nil
}

// setupServiceDependencies establishes cross-service communication
func (a *App) setupServiceDependencies() {
	// Set up cross-dependencies between services
	// Example: AudioService needs UIService for status updates
	// Example: AudioService needs IOService for output

	if audioSvc, ok := a.Services.Audio.(*services.AudioService); ok {
		audioSvc.SetDependencies(a.Services.UI, a.Services.IO)
	}
}

// setupHotkeyCallbacks connects hotkey manager with handler methods
func (a *App) setupHotkeyCallbacks() error {
	if a.Services == nil || a.Services.Hotkeys == nil {
		return fmt.Errorf("services not initialized")
	}

	// Set up hotkey callbacks by connecting to handler methods
	// The hotkey manager will call these handlers when hotkeys are triggered
	if err := a.Services.Hotkeys.SetupHotkeyCallbacks(
		a.handleStartRecording,
		a.handleStopRecordingAndTranscribe,
		// TODO: Next feature - VAD implementation
		// a.handleToggleVAD,
		a.handleSwitchModel,
		a.handleShowConfig,
		a.handleResetToDefaults,
	); err != nil {
		return fmt.Errorf("failed to set up hotkey callbacks: %w", err)
	}

	a.Runtime.Logger.Info("Hotkey callbacks configured successfully")
	return nil
}

// startServices starts all application services
func (a *App) startServices() error {
	if a.Services == nil {
		return fmt.Errorf("services not initialized")
	}

	a.Runtime.Logger.Info("Starting application services...")

	// Start WebSocket server if IO service supports it
	if a.Services.IO != nil {
		if err := a.Services.IO.StartWebSocketServer(); err != nil {
			a.Runtime.Logger.Warning("Failed to start WebSocket server: %v", err)
		}
	}

	// Register hotkeys if hotkey service supports it
	if a.Services.Hotkeys != nil {
		if err := a.Services.Hotkeys.RegisterHotkeys(); err != nil {
			a.Runtime.Logger.Warning("Failed to register hotkeys: %v", err)
		}
	}

	a.Runtime.Logger.Info("Application services started successfully")
	return nil
}

// RunAndWait starts the application and waits for shutdown signal
func (a *App) RunAndWait() error {
	a.Runtime.Logger.Info("Starting application...")

	// Start all services
	if err := a.startServices(); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait for shutdown signal
	select {
	case <-a.Runtime.ShutdownCh:
		a.Runtime.Logger.Info("Received shutdown signal")
	case <-a.Runtime.Ctx.Done():
		a.Runtime.Logger.Info("Context cancelled")
	}

	return a.Shutdown()
}

// Shutdown gracefully stops all services
func (a *App) Shutdown() error {
	a.Runtime.Logger.Info("Shutting down application...")

	// Cancel context to stop all operations
	a.Runtime.Cancel()

	// Shutdown all services
	if a.Services != nil {
		if err := a.Services.Shutdown(); err != nil {
			a.Runtime.Logger.Error("Error during service shutdown: %v", err)
			return err
		}
	}

	a.Runtime.Logger.Info("Application shutdown complete")
	return nil
}

// Service accessor methods for clean API
func (a *App) Audio() services.AudioServiceInterface {
	if a.Services == nil {
		return nil
	}
	return a.Services.Audio
}

func (a *App) UI() services.UIServiceInterface {
	if a.Services == nil {
		return nil
	}
	return a.Services.UI
}

func (a *App) IO() services.IOServiceInterface {
	if a.Services == nil {
		return nil
	}
	return a.Services.IO
}

func (a *App) Config() services.ConfigServiceInterface {
	if a.Services == nil {
		return nil
	}
	return a.Services.Config
}

func (a *App) Hotkeys() services.HotkeyServiceInterface {
	if a.Services == nil {
		return nil
	}
	return a.Services.Hotkeys
}
