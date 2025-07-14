package app

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AshBuk/speak-to-ai/audio"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/notify"
	"github.com/AshBuk/speak-to-ai/whisper"
)

// Mock implementations for testing
type MockAudioRecorder struct {
	startErr      error
	stopResult    string
	stopErr       error
	cleanupErr    error
	levelCallback audio.AudioLevelCallback
	audioLevel    float64
}

func (m *MockAudioRecorder) StartRecording() error {
	return m.startErr
}

func (m *MockAudioRecorder) StopRecording() (string, error) {
	return m.stopResult, m.stopErr
}

func (m *MockAudioRecorder) GetOutputFile() string {
	return m.stopResult
}

func (m *MockAudioRecorder) CleanupFile() error {
	return m.cleanupErr
}

func (m *MockAudioRecorder) UseStreaming() bool {
	return false
}

func (m *MockAudioRecorder) GetAudioStream() (io.Reader, error) {
	return strings.NewReader("mock audio data"), nil
}

func (m *MockAudioRecorder) SetAudioLevelCallback(callback audio.AudioLevelCallback) {
	m.levelCallback = callback
}

func (m *MockAudioRecorder) GetAudioLevel() float64 {
	return m.audioLevel
}

type MockWhisperEngine struct {
	transcribeResult string
	transcribeErr    error
}

func (m *MockWhisperEngine) Transcribe(audioFile string) (string, error) {
	return m.transcribeResult, m.transcribeErr
}

type MockOutputManager struct {
	typeErr error
}

func (m *MockOutputManager) TypeToActiveWindow(text string) error {
	return m.typeErr
}

func (m *MockOutputManager) CopyToClipboard(text string) error {
	return nil
}

type MockTrayManager struct {
	tooltip        string
	recordingState bool
	started        bool
	stopped        bool
	config         *config.Config
}

// GetTooltip returns the current tooltip
func (m *MockTrayManager) GetTooltip() string {
	return m.tooltip
}

// GetRecordingState returns the current recording state
func (m *MockTrayManager) GetRecordingState() bool {
	return m.recordingState
}

func (m *MockTrayManager) Start() {
	m.started = true
}

func (m *MockTrayManager) Stop() {
	m.stopped = true
}

func (m *MockTrayManager) SetTooltip(tooltip string) {
	m.tooltip = tooltip
}

func (m *MockTrayManager) SetRecordingState(recording bool) {
	m.recordingState = recording
}

func (m *MockTrayManager) UpdateSettings(config *config.Config) {
	m.config = config
}

type MockNotifyManager struct {
	notifications []string
}

// GetNotifications returns all notifications
func (m *MockNotifyManager) GetNotifications() []string {
	return m.notifications
}

// HasNotification checks if a specific notification exists
func (m *MockNotifyManager) HasNotification(notification string) bool {
	for _, n := range m.notifications {
		if n == notification {
			return true
		}
	}
	return false
}

func (m *MockNotifyManager) NotifyStartRecording() error {
	m.notifications = append(m.notifications, "start_recording")
	return nil
}

func (m *MockNotifyManager) NotifyStopRecording() error {
	m.notifications = append(m.notifications, "stop_recording")
	return nil
}

func (m *MockNotifyManager) NotifyTranscriptionComplete() error {
	m.notifications = append(m.notifications, "transcription_complete")
	return nil
}

func (m *MockNotifyManager) NotifyError(errMsg string) error {
	m.notifications = append(m.notifications, "error: "+errMsg)
	return nil
}

func (m *MockNotifyManager) ShowNotification(summary, body string) error {
	m.notifications = append(m.notifications, summary+": "+body)
	return nil
}

func (m *MockNotifyManager) IsAvailable() bool {
	return true
}

type MockModelManager struct {
	modelPath string
	modelErr  error
}

func (m *MockModelManager) GetModelPath() (string, error) {
	return m.modelPath, m.modelErr
}

func (m *MockModelManager) GetModelPathWithProgress(callback whisper.ProgressCallback) (string, error) {
	return m.modelPath, m.modelErr
}

func (m *MockModelManager) ValidateModel(modelPath string) error {
	return nil
}

func createTestApp() *App {
	app := NewApp("config.yaml", false, "", "", "")

	// Set up minimal configuration
	app.Config = &config.Config{}
	config.SetDefaultConfig(app.Config)

	app.Logger = &MockLogger{}
	app.Recorder = &MockAudioRecorder{}
	app.OutputManager = &MockOutputManager{}
	app.TrayManager = &MockTrayManager{}

	// Используем реальные типы для этих полей:
	app.WhisperEngine = whisper.NewWhisperEngine(app.Config, "/bin/true", "/dev/null")
	app.NotifyManager = notify.NewNotificationManager("test-app")
	app.ModelManager = whisper.NewModelManager(app.Config)

	return app
}

func TestApp_HandleStartRecording_Success(t *testing.T) {
	app := createTestApp()

	// Mock successful recording start
	mockRecorder := &MockAudioRecorder{}
	app.Recorder = mockRecorder

	err := app.handleStartRecording()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify tray state was updated
	mockTray := app.TrayManager.(*MockTrayManager)
	if !mockTray.GetRecordingState() {
		t.Error("Expected recording state to be true")
	}

	// Verify notification was sent
	if app.NotifyManager == nil {
		t.Error("NotifyManager should not be nil")
	}
	err = app.NotifyManager.NotifyStartRecording()
	if err != nil {
		t.Error(err)
	}
}

func TestApp_HandleStartRecording_RecorderError(t *testing.T) {
	app := createTestApp()

	// Mock recorder error
	mockRecorder := &MockAudioRecorder{startErr: errors.New("recording failed")}
	app.Recorder = mockRecorder

	err := app.handleStartRecording()

	if err == nil {
		t.Error("Expected error when recorder fails")
	}

	// Verify recording state was not set
	mockTray := app.TrayManager.(*MockTrayManager)
	if mockTray.recordingState {
		t.Error("Expected recording state to be false on error")
	}
}

func TestApp_HandleStopRecordingAndTranscribe_Success(t *testing.T) {
	app := createTestApp()

	// Mock successful recording stop
	mockRecorder := &MockAudioRecorder{
		stopResult: "/tmp/test.wav",
		stopErr:    nil,
	}
	app.Recorder = mockRecorder

	err := app.handleStopRecordingAndTranscribe()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify tray state was updated
	mockTray := app.TrayManager.(*MockTrayManager)
	if mockTray.recordingState {
		t.Error("Expected recording state to be false after stop")
	}

	if mockTray.tooltip != "🔄 Transcribing..." {
		t.Errorf("Expected transcribing tooltip, got %q", mockTray.tooltip)
	}

	// Verify notification was sent
	if app.NotifyManager == nil {
		t.Error("NotifyManager should not be nil")
	}
	err = app.NotifyManager.NotifyStopRecording()
	if err != nil {
		t.Error(err)
	}
}

func TestApp_HandleStopRecordingAndTranscribe_RecorderError(t *testing.T) {
	app := createTestApp()

	// Mock recorder error
	mockRecorder := &MockAudioRecorder{stopErr: errors.New("stop failed")}
	app.Recorder = mockRecorder

	err := app.handleStopRecordingAndTranscribe()

	if err == nil {
		t.Error("Expected error when recorder stop fails")
	}
}

func TestApp_HandleTranscriptionResult_Success(t *testing.T) {
	app := createTestApp()

	transcript := "Hello, world!"

	app.handleTranscriptionResult(transcript, nil)

	// Verify transcript was stored
	if app.LastTranscript != transcript {
		t.Errorf("Expected transcript %q, got %q", transcript, app.LastTranscript)
	}

	// Verify tray tooltip was reset
	mockTray := app.TrayManager.(*MockTrayManager)
	if mockTray.tooltip != "✅ Ready" {
		t.Errorf("Expected ready tooltip, got %q", mockTray.tooltip)
	}

	// Verify notification was sent
	if app.NotifyManager == nil {
		t.Error("NotifyManager should not be nil")
	}
	err := app.NotifyManager.NotifyTranscriptionComplete()
	if err != nil {
		t.Error(err)
	}
}

func TestApp_HandleTranscriptionResult_Error(t *testing.T) {
	app := createTestApp()

	testErr := errors.New("transcription failed")

	app.handleTranscriptionResult("", testErr)

	// Verify tray tooltip shows error
	mockTray := app.TrayManager.(*MockTrayManager)
	if mockTray.tooltip != "❌ Transcription failed" {
		t.Errorf("Expected error tooltip, got %q", mockTray.tooltip)
	}

	// Verify error notification was sent
	if app.NotifyManager == nil {
		t.Error("NotifyManager should not be nil")
	}
	err := app.NotifyManager.NotifyError("Transcription failed: transcription failed")
	if err != nil {
		t.Error(err)
	}
}

func TestApp_HandleTranscriptionCancellation(t *testing.T) {
	app := createTestApp()

	testErr := context.DeadlineExceeded

	app.handleTranscriptionCancellation(testErr)

	// Verify tray tooltip shows cancellation
	mockTray := app.TrayManager.(*MockTrayManager)
	if mockTray.tooltip != "⚠️  Transcription cancelled" {
		t.Errorf("Expected cancellation tooltip, got %q", mockTray.tooltip)
	}

	// Verify cancellation notification was sent
	if app.NotifyManager == nil {
		t.Error("NotifyManager should not be nil")
	}
	err := app.NotifyManager.NotifyError("Transcription was cancelled")
	if err != nil {
		t.Error(err)
	}
}

func TestApp_HandleShowConfig_Success(t *testing.T) {
	app := createTestApp()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Create the config file
	file, err := os.Create(configFile)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	file.Close()

	app.ConfigFile = configFile

	// Set EDITOR environment variable
	os.Setenv("EDITOR", "echo")
	defer os.Unsetenv("EDITOR")

	err = app.handleShowConfig()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestApp_HandleShowConfig_FileNotFound(t *testing.T) {
	app := createTestApp()

	app.ConfigFile = "/non/existent/config.yaml"

	err := app.handleShowConfig()

	if err == nil {
		t.Error("Expected error when config file doesn't exist")
	}
}

func TestApp_HandleReloadConfig_Success(t *testing.T) {
	app := createTestApp()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Create a valid config file
	configContent := `
general:
  debug: false
  model_path: "test.bin"
hotkeys:
  start_recording: "altgr+comma"
  stop_recording: "altgr+comma"
audio:
  device: "default"
  sample_rate: 16000
  recording_method: "arecord"
output:
  default_mode: "active_window"
web_server:
  enabled: false
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	app.ConfigFile = configFile

	err = app.handleReloadConfig()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify config was loaded
	if app.Config == nil {
		t.Error("Expected config to be loaded")
	}
}

func TestApp_AudioLevelCallback(t *testing.T) {
	app := createTestApp()

	mockRecorder := &MockAudioRecorder{}
	app.Recorder = mockRecorder

	// Start recording to set up callback
	err := app.handleStartRecording()
	if err != nil {
		t.Fatalf("Failed to start recording: %v", err)
	}

	// Simulate audio level callback
	if mockRecorder.levelCallback != nil {
		mockRecorder.levelCallback(0.5) // 50% level

		// Verify tray tooltip was updated
		mockTray := app.TrayManager.(*MockTrayManager)
		if mockTray.tooltip == "" {
			t.Error("Expected tooltip to be updated with audio level")
		}
	} else {
		t.Error("Expected audio level callback to be set")
	}
}
