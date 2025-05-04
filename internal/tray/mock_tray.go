//go:build !systray
// +build !systray

package tray

import (
	"log"
)

// MockTrayManager implements a mock version of TrayManager without external dependencies
type MockTrayManager struct {
	isRecording bool
	onExit      func()
	onToggle    func() error
}

// NewMockTrayManager creates a new mock tray manager instance
func NewMockTrayManager(onExit func(), onToggle func() error) TrayManagerInterface {
	return &MockTrayManager{
		isRecording: false,
		onExit:      onExit,
		onToggle:    onToggle,
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

// Stop stops the mock tray manager (no-op)
func (tm *MockTrayManager) Stop() {
	log.Println("Mock tray manager stopped")
}
