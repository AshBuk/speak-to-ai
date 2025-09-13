// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"github.com/AshBuk/speak-to-ai/internal/notify"
	"github.com/AshBuk/speak-to-ai/internal/tray"
	outputInterfaces "github.com/AshBuk/speak-to-ai/output/interfaces"
	"github.com/AshBuk/speak-to-ai/websocket"
)

// ServiceAssembler is responsible for building the ServiceContainer from the ready components
type ServiceAssembler struct {
	factoryConfig ServiceFactoryConfig
}

func NewServiceAssembler(config ServiceFactoryConfig) *ServiceAssembler {
	return &ServiceAssembler{factoryConfig: config}
}

// Assemble builds the ServiceContainer from the ready components
func (sa *ServiceAssembler) Assemble(components *Components) *ServiceContainer {
	container := NewServiceContainer()

	// Build services using concrete instances where needed
	configSvc := sa.createConfigService()
	hotkeysSvc := sa.createHotkeyService(components.HotkeyManager)
	audioSvc := sa.createAudioService(components)
	uiSvc := sa.createUIService(components.TrayManager, components.NotifyManager)

	// Wire UI to Config for notifications (concrete method)
	configSvc.SetUIService(uiSvc)

	ioSvc := sa.createIOService(components.OutputManager, components.WebSocketServer)

	// Assign to container interfaces
	container.Config = configSvc
	container.Hotkeys = hotkeysSvc
	container.Audio = audioSvc
	container.UI = uiSvc
	container.IO = ioSvc

	// Cross-dependencies
	sa.SetupDependencies(container)

	return container
}

// SetupDependencies sets up the dependencies between the services
func (sa *ServiceAssembler) SetupDependencies(container *ServiceContainer) {
	if audioSvc, ok := container.Audio.(*AudioService); ok {
		audioSvc.SetDependencies(container.UI, container.IO)
		audioSvc.SetConfig(container.Config)
	}
}

// createConfigService creates the ConfigService
func (sa *ServiceAssembler) createConfigService() *ConfigService {
	return NewConfigService(
		sa.factoryConfig.Logger,
		sa.factoryConfig.Config,
		sa.factoryConfig.ConfigFile,
	)
}

// createHotkeyService creates the HotkeyService
func (sa *ServiceAssembler) createHotkeyService(hotkeyManager HotkeyManagerInterface) *HotkeyService {
	return NewHotkeyService(
		sa.factoryConfig.Logger,
		hotkeyManager,
	)
}

// createAudioService creates the AudioService
func (sa *ServiceAssembler) createAudioService(components *Components) *AudioService {
	return NewAudioService(
		sa.factoryConfig.Logger,
		sa.factoryConfig.Config,
		components.Recorder,
		components.WhisperEngine,
		components.ModelManager,
	)
}

// createUIService creates the UIService
func (sa *ServiceAssembler) createUIService(trayManager tray.TrayManagerInterface, notifyManager *notify.NotificationManager) *UIService {
	return NewUIService(
		sa.factoryConfig.Logger,
		trayManager,
		notifyManager,
	)
}

// createIOService creates the IOService
func (sa *ServiceAssembler) createIOService(outputManager outputInterfaces.Outputter, webSocketServer *websocket.WebSocketServer) *IOService {
	return NewIOService(
		sa.factoryConfig.Logger,
		outputManager,
		webSocketServer,
	)
}
