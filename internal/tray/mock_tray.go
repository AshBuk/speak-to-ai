//go:build !systray

package tray

import (
	"log"

	"github.com/AshBuk/speak-to-ai/config"
)

// MockTrayManager implements a mock version of TrayManager without external dependencies
type MockTrayManager struct {
	isRecording    bool
	onExit         func()
	onToggle       func() error
	onShowConfig   func() error
	onReloadConfig func() error
}

// NewMockTrayManager creates a new mock tray manager instance
func NewMockTrayManager(onExit func(), onToggle func() error, onShowConfig func() error, onReloadConfig func() error) TrayManagerInterface {
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
	log.Println("Mock tray manager started (no actual system tray is shown)")
}

// SetRecordingState updates the mock tray state
func (tm *MockTrayManager) SetRecordingState(isRecording bool) {
	tm.isRecording = isRecording
	if isRecording {
		log.Println("Tray icon: Recording ON")
	} else {
		log.Println("Tray icon: Recording OFF")
	}
}

// SetTooltip sets the tooltip text (mock implementation)
func (tm *MockTrayManager) SetTooltip(tooltip string) {
	log.Printf("Tray tooltip: %s", tooltip)
}

// UpdateSettings updates the settings display (mock implementation)
func (tm *MockTrayManager) UpdateSettings(config *config.Config) {
	log.Printf("Mock tray: Settings updated with config")
}

// Stop stops the mock tray manager (no-op)
func (tm *MockTrayManager) Stop() {
	log.Println("Mock tray manager stopped")
}
