// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	"github.com/AshBuk/speak-to-ai/internal/services"
	"github.com/AshBuk/speak-to-ai/internal/utils"
)

// Manages application lifecycle and context
type RuntimeContext struct {
	Ctx        context.Context
	Cancel     context.CancelFunc
	ShutdownCh chan os.Signal
	Logger     logger.Logger
}

// Create a new runtime context
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

// Represents the main application and its services
type App struct {
	Services *services.ServiceContainer
	Runtime  *RuntimeContext
}

// Create a new application instance
func NewApp(logger logger.Logger) *App {
	return &App{
		Services: services.NewServiceContainer(),
		Runtime:  NewRuntimeContext(logger),
	}
}

// Initialize all application services and dependencies
func (a *App) Initialize(configFile string, debug bool) error {
	a.Runtime.Logger.Info("Initializing application...")

	config, err := a.initializeConfig(configFile, debug)
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	if err := a.initializeServices(config, configFile); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	a.setupServiceDependencies()

	a.Runtime.Logger.Info("Application initialization complete")
	return nil
}

// Load and validate the application configuration
func (a *App) initializeConfig(configFile string, debug bool) (*config.Config, error) {
	a.Runtime.Logger.Info("Loading configuration from: %s", configFile)

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Override the debug flag from the command line if specified
	if debug {
		cfg.General.Debug = true
	}

	a.Runtime.Logger.Info("Configuration loaded successfully")
	return cfg, nil
}

// Create and configure all services
func (a *App) initializeServices(cfg *config.Config, cfgFilePath string) error {
	a.Runtime.Logger.Info("Initializing services with dependency injection...")

	environment := platform.DetectEnvironment()
	a.Runtime.Logger.Info("Detected environment: %s", environment)

	factory := services.NewServiceFactory(services.ServiceFactoryConfig{
		Logger:      a.Runtime.Logger,
		Config:      cfg,
		ConfigFile:  cfgFilePath,
		Environment: environment,
	})

	serviceContainer, err := factory.CreateServices()
	if err != nil {
		return fmt.Errorf("failed to create services: %w", err)
	}

	a.Services = serviceContainer

	if err := a.setupHotkeyCallbacks(); err != nil {
		a.Runtime.Logger.Warning("Failed to set up hotkey callbacks: %v", err)
	}

	a.Runtime.Logger.Info("Services initialized successfully")
	return nil
}

// Establish communication channels between services
func (a *App) setupServiceDependencies() {
	// Example: AudioService needs UIService for status updates and IOService for output
	if audioSvc, ok := a.Services.Audio.(*services.AudioService); ok {
		audioSvc.SetDependencies(a.Services.UI, a.Services.IO)
	}
}

// Connect hotkey manager events to their corresponding handler methods
func (a *App) setupHotkeyCallbacks() error {
	if a.Services == nil || a.Services.Hotkeys == nil {
		return fmt.Errorf("services not initialized")
	}

	if err := a.Services.Hotkeys.SetupHotkeyCallbacks(
		a.handleStartRecording,
		a.handleStopRecordingAndTranscribe,
		// TODO: Next feature - VAD implementation
		// a.handleToggleVAD,
		a.handleShowConfig,
		a.handleResetToDefaults,
	); err != nil {
		return fmt.Errorf("failed to set up hotkey callbacks: %w", err)
	}

	a.Runtime.Logger.Info("Hotkey callbacks configured successfully")
	return nil
}

// Start all application services
func (a *App) startServices() error {
	if a.Services == nil {
		return fmt.Errorf("services not initialized")
	}

	a.Runtime.Logger.Info("Starting application services...")

	if a.Services.IO != nil {
		if err := a.Services.IO.StartWebSocketServer(); err != nil {
			a.Runtime.Logger.Warning("Failed to start WebSocket server: %v", err)
		}
	}

	if a.Services.Hotkeys != nil {
		if err := a.Services.Hotkeys.RegisterHotkeys(); err != nil {
			a.Runtime.Logger.Warning("Failed to register hotkeys: %v", err)
		}
	}

	a.Runtime.Logger.Info("Application services started successfully")
	return nil
}

// Start the application and wait for a shutdown signal
func (a *App) RunAndWait() error {
	a.Runtime.Logger.Info("Starting application...")

	if err := a.startServices(); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait for a shutdown signal
	select {
	case <-a.Runtime.ShutdownCh:
		a.Runtime.Logger.Info("Received shutdown signal")
	case <-a.Runtime.Ctx.Done():
		a.Runtime.Logger.Info("Context cancelled")
	}

	return a.Shutdown()
}

// Gracefully stop all services
func (a *App) Shutdown() error {
	a.Runtime.Logger.Info("Shutting down application...")

	// Cancel the main context to signal all operations to stop
	a.Runtime.Cancel()

	if a.Services != nil {
		if err := a.Services.Shutdown(); err != nil {
			a.Runtime.Logger.Error("Error during service shutdown: %v", err)
			return err
		}
	}

	// Wait for long-lived goroutines to complete with a timeout
	if ok := utils.WaitAll(5 * time.Second); ok {
		a.Runtime.Logger.Info("Background tasks completed")
	} else {
		a.Runtime.Logger.Warning("Shutdown timeout - forcing exit")
	}

	a.Runtime.Logger.Info("Application shutdown complete")
	return nil
}

// Provide clean accessor methods for services

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
