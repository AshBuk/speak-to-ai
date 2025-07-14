package app

import (
	"syscall"
	"testing"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	"github.com/AshBuk/speak-to-ai/output"
)

func TestNewApp(t *testing.T) {
	configFile := "config.yaml"
	debug := true
	whisperPath := "/usr/bin/whisper"
	modelPath := "/path/to/model"
	quantizePath := "/path/to/quantize"

	app := NewApp(configFile, debug, whisperPath, modelPath, quantizePath)

	if app == nil {
		t.Fatal("NewApp returned nil")
	}

	if app.ShutdownCh == nil {
		t.Error("ShutdownCh should be initialized")
	}

	if app.Ctx == nil {
		t.Error("Context should be initialized")
	}

	if app.Cancel == nil {
		t.Error("Cancel function should be initialized")
	}

	// Test that context is not cancelled initially
	select {
	case <-app.Ctx.Done():
		t.Error("Context should not be cancelled initially")
	default:
		// Expected behavior
	}
}

func TestApp_ContextCancellation(t *testing.T) {
	app := NewApp("config.yaml", false, "", "", "")

	// Cancel the context
	app.Cancel()

	// Wait for context to be done
	select {
	case <-app.Ctx.Done():
		// Expected behavior
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled after calling Cancel()")
	}
}

func TestApp_SignalHandling(t *testing.T) {
	app := NewApp("config.yaml", false, "", "", "")

	// Test that signal channel is properly configured
	if cap(app.ShutdownCh) == 0 {
		t.Error("ShutdownCh should be buffered")
	}

	// Send a signal to the channel
	app.ShutdownCh <- syscall.SIGTERM

	// Read from the channel
	select {
	case sig := <-app.ShutdownCh:
		if sig != syscall.SIGTERM {
			t.Errorf("Expected SIGTERM, got %v", sig)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Should receive signal from channel")
	}
}

func TestApp_ComponentInitialization(t *testing.T) {
	app := NewApp("config.yaml", false, "", "", "")

	// Test initial state
	if app.Logger != nil {
		t.Error("Logger should be nil initially")
	}

	if app.Config != nil {
		t.Error("Config should be nil initially")
	}

	if app.Recorder != nil {
		t.Error("Recorder should be nil initially")
	}

	if app.WhisperEngine != nil {
		t.Error("WhisperEngine should be nil initially")
	}

	if app.HotkeyManager != nil {
		t.Error("HotkeyManager should be nil initially")
	}

	if app.TrayManager != nil {
		t.Error("TrayManager should be nil initially")
	}

	if app.NotifyManager != nil {
		t.Error("NotifyManager should be nil initially")
	}

	if app.WebSocketServer != nil {
		t.Error("WebSocketServer should be nil initially")
	}

	if app.OutputManager != nil {
		t.Error("OutputManager should be nil initially")
	}

	if app.ModelManager != nil {
		t.Error("ModelManager should be nil initially")
	}
}

func TestApp_ConfigFile(t *testing.T) {
	configFile := "test-config.yaml"
	app := NewApp(configFile, false, "", "", "")

	// ConfigFile should be set after initialization
	if app.ConfigFile != "" {
		t.Error("ConfigFile should be empty initially")
	}
}

func TestApp_Environment(t *testing.T) {
	app := NewApp("config.yaml", false, "", "", "")

	// Environment should be unknown initially
	if app.Environment != "" {
		t.Error("Environment should be empty initially")
	}
}

func TestApp_LastTranscript(t *testing.T) {
	app := NewApp("config.yaml", false, "", "", "")

	// Test LastTranscript field
	if app.LastTranscript != "" {
		t.Error("LastTranscript should be empty initially")
	}

	// Test setting LastTranscript
	testTranscript := "Hello, world!"
	app.LastTranscript = testTranscript

	if app.LastTranscript != testTranscript {
		t.Errorf("Expected LastTranscript to be %q, got %q", testTranscript, app.LastTranscript)
	}
}

// MockComponents for testing
type MockLogger struct {
	messages []string
}

func (m *MockLogger) Debug(format string, args ...interface{}) {
	m.messages = append(m.messages, "DEBUG: "+format)
}

func (m *MockLogger) Info(format string, args ...interface{}) {
	m.messages = append(m.messages, "INFO: "+format)
}

func (m *MockLogger) Warning(format string, args ...interface{}) {
	m.messages = append(m.messages, "WARNING: "+format)
}

func (m *MockLogger) Error(format string, args ...interface{}) {
	m.messages = append(m.messages, "ERROR: "+format)
}

func TestApp_ComponentAssignment(t *testing.T) {
	app := NewApp("config.yaml", false, "", "", "")

	// Test component assignment
	mockLogger := &MockLogger{}
	app.Logger = mockLogger

	testConfig := &config.Config{}
	app.Config = testConfig

	app.Environment = platform.EnvironmentX11
	app.ConfigFile = "test-config.yaml"

	// Verify assignments
	if app.Logger != mockLogger {
		t.Error("Logger assignment failed")
	}

	if app.Config != testConfig {
		t.Error("Config assignment failed")
	}

	if app.Environment != platform.EnvironmentX11 {
		t.Error("Environment assignment failed")
	}

	if app.ConfigFile != "test-config.yaml" {
		t.Error("ConfigFile assignment failed")
	}
}

func TestApp_ConvertEnvironmentType(t *testing.T) {
	app := NewApp("config.yaml", false, "", "", "")

	tests := []struct {
		name     string
		input    platform.EnvironmentType
		expected output.EnvironmentType
	}{
		{
			name:     "X11 environment",
			input:    platform.EnvironmentX11,
			expected: output.EnvironmentX11,
		},
		{
			name:     "Wayland environment",
			input:    platform.EnvironmentWayland,
			expected: output.EnvironmentWayland,
		},
		{
			name:     "Unknown environment",
			input:    platform.EnvironmentUnknown,
			expected: output.EnvironmentUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.Environment = tt.input
			result := app.convertEnvironmentType()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
