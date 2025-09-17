// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package mocks

import (
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
)

// MockAudioService implements AudioServiceInterface for testing
type MockAudioService struct {
	shutdownCalled bool
	shutdownError  error
}

func (m *MockAudioService) Shutdown() error {
	m.shutdownCalled = true
	return m.shutdownError
}

func (m *MockAudioService) HandleStartRecording() error { return nil }
func (m *MockAudioService) HandleStopRecording() error  { return nil }
func (m *MockAudioService) IsRecording() bool           { return false }

// TODO: Next feature - VAD implementation
// func (m *MockAudioService) HandleStartVADRecording() error                  { return nil }

// Test helper methods
func (m *MockAudioService) WasShutdownCalled() bool { return m.shutdownCalled }

func (m *MockAudioService) SetShutdownError(err error) { m.shutdownError = err }

// MockUIService implements UIServiceInterface for testing
type MockUIService struct {
	shutdownCalled bool
	shutdownError  error
}

func (m *MockUIService) Shutdown() error {
	m.shutdownCalled = true
	return m.shutdownError
}

func (m *MockUIService) SetRecordingState(isRecording bool)                {}
func (m *MockUIService) SetTooltip(tooltip string)                         {}
func (m *MockUIService) ShowNotification(title, message string)            {}
func (m *MockUIService) UpdateRecordingUI(isRecording bool, level float64) {}
func (m *MockUIService) SetError(message string)                           {}
func (m *MockUIService) SetSuccess(message string)                         {}
func (m *MockUIService) ShowConfigFile() error                             { return nil }
func (m *MockUIService) ShowAboutPage() error                              { return nil }
func (m *MockUIService) UpdateSettings(cfg *config.Config)                 {}

// Test helper methods
func (m *MockUIService) WasShutdownCalled() bool { return m.shutdownCalled }

func (m *MockUIService) SetShutdownError(err error) { m.shutdownError = err }

// MockIOService implements IOServiceInterface for testing
type MockIOService struct {
	shutdownCalled bool
	shutdownError  error
}

func (m *MockIOService) Shutdown() error {
	m.shutdownCalled = true
	return m.shutdownError
}

func (m *MockIOService) OutputText(text string) error           { return nil }
func (m *MockIOService) SetOutputMethod(method string) error    { return nil }
func (m *MockIOService) StartWebSocketServer() error            { return nil }
func (m *MockIOService) StopWebSocketServer() error             { return nil }
func (m *MockIOService) HandleTypingFallback(text string) error { return nil }

// Test helper methods
func (m *MockIOService) WasShutdownCalled() bool { return m.shutdownCalled }

func (m *MockIOService) SetShutdownError(err error) { m.shutdownError = err }

// MockConfigService implements ConfigServiceInterface for testing
type MockConfigService struct {
	shutdownCalled bool
	shutdownError  error
}

func (m *MockConfigService) Shutdown() error {
	m.shutdownCalled = true
	return m.shutdownError
}

func (m *MockConfigService) LoadConfig(configFile string) error { return nil }
func (m *MockConfigService) SaveConfig() error                  { return nil }
func (m *MockConfigService) ResetToDefaults() error             { return nil }
func (m *MockConfigService) GetConfig() *config.Config          { return &config.Config{} }

// TODO: Next feature - VAD implementation
// func (m *MockConfigService) UpdateVADSensitivity(sensitivity string) error { return nil }
func (m *MockConfigService) UpdateLanguage(language string) error { return nil }
func (m *MockConfigService) ToggleWorkflowNotifications() error   { return nil }

// TODO: Next feature - VAD implementation
// func (m *MockConfigService) ToggleVAD() error                              { return nil }
func (m *MockConfigService) UpdateRecordingMethod(method string) error { return nil }
func (m *MockConfigService) UpdateOutputMode(mode string) error        { return nil }
func (m *MockConfigService) UpdateHotkey(action, combo string) error   { return nil }

// Test helper methods
func (m *MockConfigService) WasShutdownCalled() bool { return m.shutdownCalled }

func (m *MockConfigService) SetShutdownError(err error) { m.shutdownError = err }

// MockHotkeyService implements HotkeyServiceInterface for testing
type MockHotkeyService struct {
	shutdownCalled bool
	shutdownError  error
}

func (m *MockHotkeyService) Shutdown() error {
	m.shutdownCalled = true
	return m.shutdownError
}

func (m *MockHotkeyService) SetupHotkeyCallbacks(
	startRecording func() error,
	stopRecording func() error,
	// TODO: Next feature - VAD implementation
	// toggleVAD func() error,
	showConfig func() error,
	resetToDefaults func() error,
) error {
	return nil
}
func (m *MockHotkeyService) RegisterHotkeys() error   { return nil }
func (m *MockHotkeyService) UnregisterHotkeys() error { return nil }
func (m *MockHotkeyService) ReloadFromConfig(startRecording, stopRecording func() error, _ func() adapters.HotkeyConfig) error {
	return nil
}

// CaptureOnce mock implementation
func (m *MockHotkeyService) CaptureOnce(timeoutMs int) (string, error) { return "alt+r", nil }

// SupportsCaptureOnce mock implementation
func (m *MockHotkeyService) SupportsCaptureOnce() bool { return true }

// Test helper methods
func (m *MockHotkeyService) WasShutdownCalled() bool { return m.shutdownCalled }

func (m *MockHotkeyService) SetShutdownError(err error) { m.shutdownError = err }
