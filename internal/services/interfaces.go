// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"github.com/AshBuk/speak-to-ai/audio/processing"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
)

// Defines the contract for audio-related operations
type AudioServiceInterface interface {
	// Recording lifecycle
	HandleStartRecording() error
	HandleStopRecording() error
	IsRecording() bool

	// TODO: Next feature - VAD implementation
	// VAD (Voice Activity Detection) operations
	// HandleStartVADRecording() error

	// Cleanup
	Shutdown() error
}

// Defines the contract for user interface operations
type UIServiceInterface interface {
	// Tray management
	SetRecordingState(isRecording bool)
	ShowNotification(title, message string)
	UpdateSettings(config *config.Config)

	// State updates
	UpdateRecordingUI(isRecording bool, level float64)
	SetError(message string)
	SetSuccess(message string)

	// Menu actions
	ShowConfigFile() error
	ShowAboutPage() error

	// Cleanup
	Shutdown() error
}

// Defines the contract for input/output operations
type IOServiceInterface interface {
	// Text output
	OutputText(text string) error
	SetOutputMethod(method string) error

	// WebSocket operations
	StartWebSocketServer() error
	StopWebSocketServer() error

	// Cleanup
	Shutdown() error
}

// Defines the contract for configuration operations
type ConfigServiceInterface interface {
	// Configuration management
	LoadConfig(configFile string) error
	SaveConfig() error
	ResetToDefaults() error
	GetConfig() *config.Config

	// Settings updates
	// TODO: Next feature - VAD implementation
	// UpdateVADSensitivity(sensitivity string) error
	// ToggleVAD() error
	UpdateLanguage(language string) error
	ToggleWorkflowNotifications() error
	UpdateRecordingMethod(method string) error
	UpdateOutputMode(mode string) error
	UpdateHotkey(action, combo string) error

	// Cleanup
	Shutdown() error
}

// Defines the contract for hotkey operations
type HotkeyServiceInterface interface {
	// Hotkey callback setup
	SetupHotkeyCallbacks(
		startRecording func() error,
		stopRecording func() error,
		// toggleVAD func() error,
		showConfig func() error,
		reloadConfig func() error,
	) error

	// Hotkey lifecycle
	RegisterHotkeys() error
	UnregisterHotkeys() error

	// Hotkey reloading
	ReloadFromConfig(startRecording, stopRecording func() error, configProvider func() adapters.HotkeyConfig) error

	// One-shot capture for rebind flow
	CaptureOnce(timeoutMs int) (string, error)

	// Capability flags
	SupportsCaptureOnce() bool

	// Cleanup
	Shutdown() error
}

// ServiceContainer holds all service interfaces
type ServiceContainer struct {
	Audio           AudioServiceInterface
	UI              UIServiceInterface
	IO              IOServiceInterface
	Config          ConfigServiceInterface
	Hotkeys         HotkeyServiceInterface
	TempFileManager *processing.TempFileManager
}

// Create a new service container with all services
func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{}
}

// Gracefully shut down all services
func (sc *ServiceContainer) Shutdown() error {
	var lastErr error

	if sc.Audio != nil {
		if err := sc.Audio.Shutdown(); err != nil {
			lastErr = err
		}
	}

	if sc.UI != nil {
		if err := sc.UI.Shutdown(); err != nil {
			lastErr = err
		}
	}

	if sc.IO != nil {
		if err := sc.IO.Shutdown(); err != nil {
			lastErr = err
		}
	}

	if sc.Config != nil {
		if err := sc.Config.Shutdown(); err != nil {
			lastErr = err
		}
	}

	if sc.Hotkeys != nil {
		if err := sc.Hotkeys.Shutdown(); err != nil {
			lastErr = err
		}
	}

	if sc.TempFileManager != nil {
		sc.TempFileManager.Stop()
		sc.TempFileManager.CleanupAll()
	}

	return lastErr
}
