// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

// No imports needed for interfaces

// AudioServiceInterface defines the contract for audio-related operations
type AudioServiceInterface interface {
	// Recording lifecycle
	HandleStartRecording() error
	HandleStopRecording() error
	IsRecording() bool
	GetLastTranscript() string

	// Streaming operations
	HandleStartStreamingRecording() error
	HandleStreamingResult(text string, isFinal bool)

	// VAD (Voice Activity Detection) operations
	HandleStartVADRecording() error

	// Model management
	EnsureModelAvailable() error
	EnsureAudioRecorderAvailable() error
	SwitchModel(modelType string) error

	// Cleanup
	Shutdown() error
}

// UIServiceInterface defines the contract for user interface operations
type UIServiceInterface interface {
	// Tray management
	SetRecordingState(isRecording bool)
	SetTooltip(tooltip string)
	ShowNotification(title, message string)

	// State updates
	UpdateRecordingUI(isRecording bool, level float64)
	SetError(message string)
	SetSuccess(message string)

	// Menu actions
	ShowConfigFile() error

	// Cleanup
	Shutdown() error
}

// IOServiceInterface defines the contract for input/output operations
type IOServiceInterface interface {
	// Text output
	OutputText(text string) error
	SetOutputMethod(method string) error

	// WebSocket operations
	BroadcastTranscription(text string, isFinal bool)
	StartWebSocketServer() error
	StopWebSocketServer() error

	// Output routing
	HandleTypingFallback(text string) error

	// Cleanup
	Shutdown() error
}

// ConfigServiceInterface defines the contract for configuration operations
type ConfigServiceInterface interface {
	// Configuration management
	LoadConfig(configFile string) error
	SaveConfig() error
	ReloadConfig() error
	GetConfig() interface{}

	// Settings updates
	UpdateVADSensitivity(sensitivity string) error
	UpdateLanguage(language string) error
	UpdateModelType(modelType string) error
	ToggleWorkflowNotifications() error
	ToggleStreaming() error
	ToggleVAD() error

	// Hotkey management
	RegisterHotkeys() error
	UnregisterHotkeys() error

	// Cleanup
	Shutdown() error
}

// ServiceContainer holds all service interfaces
type ServiceContainer struct {
	Audio  AudioServiceInterface
	UI     UIServiceInterface
	IO     IOServiceInterface
	Config ConfigServiceInterface
}

// NewServiceContainer creates a new service container with all services
func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{}
}

// Shutdown gracefully shuts down all services
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

	return lastErr
}
