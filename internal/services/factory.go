// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/audio/interfaces"
	"github.com/AshBuk/speak-to-ai/audio/processing"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/hotkeys/manager"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/notify"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	"github.com/AshBuk/speak-to-ai/internal/tray"
	outputInterfaces "github.com/AshBuk/speak-to-ai/output/interfaces"
	"github.com/AshBuk/speak-to-ai/websocket"
	"github.com/AshBuk/speak-to-ai/whisper"
)

// ServiceFactoryConfig holds all dependencies needed to create services
type ServiceFactoryConfig struct {
	Logger      logger.Logger
	Config      *config.Config
	ConfigFile  string
	Environment platform.EnvironmentType
}

// Components holds all initialized application components
type Components struct {
	ModelManager    whisper.ModelManager
	Recorder        interfaces.AudioRecorder
	WhisperEngine   *whisper.WhisperEngine
	OutputManager   outputInterfaces.Outputter
	HotkeyManager   *manager.HotkeyManager
	WebSocketServer *websocket.WebSocketServer
	TrayManager     tray.TrayManagerInterface
	NotifyManager   *notify.NotificationManager
	TempFileManager *processing.TempFileManager
}

// ServiceFactory creates and configures all services with proper dependency injection
type ServiceFactory struct {
	config ServiceFactoryConfig
}

// Create a new service factory
func NewServiceFactory(config ServiceFactoryConfig) *ServiceFactory {
	return &ServiceFactory{
		config: config,
	}
}

// CreateServices creates and configures all services
func (sf *ServiceFactory) CreateServices() (*ServiceContainer, error) {
	// Initialize components via ComponentFactory
	compFactory := NewComponentFactory(sf.config)
	components, err := compFactory.InitializeComponents()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	// Assemble services from components
	assembler := NewServiceAssembler(sf.config)
	container := assembler.Assemble(components)

	// Wire tray callbacks
	wirer := NewCallbackWirer(sf.config.Logger)
	wirer.Wire(container, components)

	return container, nil
}
