// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package tray

import (
	"log"

	"github.com/AshBuk/speak-to-ai/config"
)

// MockTrayManager implements a mock version of TrayManager without external dependencies
type MockTrayManager struct {
	isRecording            bool
	onExit                 func()
	onToggle               func() error
	onShowConfig           func() error
	onReloadConfig         func() error
	onSelectRecorder       func(method string) error
	onTestRecording        func() error
	onSelectVADSens        func(sensitivity string) error
	onSelectLang           func(language string) error
	onSelectModel          func(modelType string) error
	onToggleWorkflowNotify func() error
}

// CreateMockTrayManager creates a mock tray manager that doesn't use systray
func CreateMockTrayManager(onExit func(), onToggle func() error, onShowConfig func() error, onReloadConfig func() error) TrayManagerInterface {
	return &MockTrayManager{
		isRecording:    false,
		onExit:         onExit,
		onToggle:       onToggle,
		onShowConfig:   onShowConfig,
		onReloadConfig: onReloadConfig,
	}
}

// Start initializes and starts the mock system tray (no-op)
func (tm *MockTrayManager) Start() {
	log.Println("Mock tray started (no actual system tray is shown)")
}

func (tm *MockTrayManager) SetRecordingState(isRecording bool) {
	tm.isRecording = isRecording
	if isRecording {
		log.Println("Mock tray: Recording ON")
	} else {
		log.Println("Mock tray: Recording OFF")
	}
}

func (tm *MockTrayManager) SetTooltip(tooltip string) {
	log.Printf("Mock tray tooltip: %s", tooltip)
}

func (tm *MockTrayManager) UpdateSettings(config *config.Config) {
	log.Println("Mock tray: Settings updated")
}

func (tm *MockTrayManager) Stop() {
	log.Println("Mock tray stopped")
}

// SetAudioActions sets callbacks for audio-related actions (mock implementation)
func (tm *MockTrayManager) SetAudioActions(onSelectRecorder func(method string) error, onTestRecording func() error) {
	tm.onSelectRecorder = onSelectRecorder
	tm.onTestRecording = onTestRecording
	log.Println("Mock tray: audio actions set")
}

// SetSettingsActions sets callbacks for settings (mock implementation)
func (tm *MockTrayManager) SetSettingsActions(
	onSelectVADSensitivity func(sensitivity string) error,
	onSelectLanguage func(language string) error,
	onSelectModelType func(modelType string) error,
	onToggleWorkflowNotifications func() error,
) {
	tm.onSelectVADSens = onSelectVADSensitivity
	tm.onSelectLang = onSelectLanguage
	tm.onSelectModel = onSelectModelType
	tm.onToggleWorkflowNotify = onToggleWorkflowNotifications
	log.Println("Mock tray: settings actions set")
}
