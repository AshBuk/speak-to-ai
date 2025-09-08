// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/AshBuk/speak-to-ai/audio/interfaces"
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

// App represents the application with all its components
type App struct {
	Logger                   logger.Logger
	Config                   *config.Config
	ConfigFile               string // Path to the configuration file
	Environment              platform.EnvironmentType
	ModelManager             *whisper.ModelManager
	Recorder                 interfaces.AudioRecorder
	WhisperEngine            *whisper.WhisperEngine
	StreamingEngine          *whisper.StreamingWhisperEngine
	OutputManager            outputInterfaces.Outputter
	HotkeyManager            *manager.HotkeyManager
	WebSocketServer          *websocket.WebSocketServer
	TrayManager              tray.TrayManagerInterface
	NotifyManager            *notify.NotificationManager
	LastTranscript           string
	audioRecorderNeedsReinit bool // Flag to trigger lazy audio recorder reinitialization
	ShutdownCh               chan os.Signal
	Ctx                      context.Context
	Cancel                   context.CancelFunc
}

// NewApp creates a new application instance with the specified configuration
func NewApp(configFile string, debug bool, modelPath, quantizePath string) *App {
	app := &App{
		ShutdownCh: make(chan os.Signal, 1),
	}

	// Create context with cancellation for clean shutdown
	app.Ctx, app.Cancel = context.WithCancel(context.Background())

	// Set up signal handling
	signal.Notify(app.ShutdownCh, os.Interrupt, syscall.SIGTERM)

	return app
}

// updateUIState updates tray recording indicator and tooltip if available
func (a *App) updateUIState(isRecording bool, tooltip string) {
	if a.TrayManager != nil {
		a.TrayManager.SetRecordingState(isRecording)
		if tooltip != "" {
			a.TrayManager.SetTooltip(tooltip)
		}
	}
}

// resetRecordingState resets HotkeyManager state if available
func (a *App) resetRecordingState() {
	if a.HotkeyManager != nil {
		a.HotkeyManager.ResetRecordingState()
	}
}

// notify shows a notification with basic error logging
func (a *App) notify(title, message string) {
	if a.NotifyManager != nil {
		if err := a.NotifyManager.ShowNotification(title, message); err != nil {
			a.Logger.Warning("Failed to show notification: %v", err)
		}
	}
}

// handleRecordingError centralizes UI reset and error notification for recording failures
func (a *App) handleRecordingError(err error) {
	a.updateUIState(false, "⚠️ Recording failed")
	a.resetRecordingState()
	a.notify("Recording Error", fmt.Sprintf("%v", err))
}

// updateTraySettings updates tray UI to reflect current configuration
func (a *App) updateTraySettings() {
	if a.TrayManager != nil {
		a.TrayManager.UpdateSettings(a.Config)
		a.Logger.Info("Tray settings updated")
	}
}
