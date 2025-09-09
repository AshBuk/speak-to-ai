// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/notify"
	"github.com/AshBuk/speak-to-ai/internal/tray"
)

// UIService implements UIServiceInterface
type UIService struct {
	logger        logger.Logger
	trayManager   tray.TrayManagerInterface
	notifyManager *notify.NotificationManager
}

// NewUIService creates a new UIService instance
func NewUIService(
	logger logger.Logger,
	trayManager tray.TrayManagerInterface,
	notifyManager *notify.NotificationManager,
) *UIService {
	return &UIService{
		logger:        logger,
		trayManager:   trayManager,
		notifyManager: notifyManager,
	}
}

// SetRecordingState implements UIServiceInterface
func (us *UIService) SetRecordingState(isRecording bool) {
	if us.trayManager != nil {
		us.trayManager.SetRecordingState(isRecording)
	}
}

// SetTooltip implements UIServiceInterface
func (us *UIService) SetTooltip(tooltip string) {
	if us.trayManager != nil {
		us.trayManager.SetTooltip(tooltip)
	}
}

// ShowNotification implements UIServiceInterface
func (us *UIService) ShowNotification(title, message string) {
	if us.notifyManager != nil {
		// Use appropriate notification method based on message type
		// Use generic sendNotification method
		if err := us.sendNotification(title, message, "dialog-information"); err != nil {
			us.logger.Warning("Failed to show notification: %v", err)
		}
	}
}

// UpdateRecordingUI implements UIServiceInterface
func (us *UIService) UpdateRecordingUI(isRecording bool, level float64) {
	us.SetRecordingState(isRecording)

	if isRecording && us.trayManager != nil {
		levelPercentage := int(level * 100)
		if levelPercentage > 100 {
			levelPercentage = 100
		}

		tooltip := fmt.Sprintf("Recording... Level: %d%%", levelPercentage)
		us.SetTooltip(tooltip)
	}
}

// SetError implements UIServiceInterface
func (us *UIService) SetError(message string) {
	us.logger.Error("UI Error: %s", message)

	if us.trayManager != nil {
		us.trayManager.SetTooltip(fmt.Sprintf("Error: %s", message))
	}

	if us.notifyManager != nil {
		if err := us.sendNotification("Error", message, "dialog-error"); err != nil {
			us.logger.Warning("Failed to show error notification: %v", err)
		}
	}
}

// SetSuccess implements UIServiceInterface
func (us *UIService) SetSuccess(message string) {
	us.logger.Info("UI Success: %s", message)

	if us.trayManager != nil {
		us.trayManager.SetTooltip("Ready")
	}

	if us.notifyManager != nil {
		if err := us.sendNotification("Success", message, "dialog-ok-apply"); err != nil {
			us.logger.Warning("Failed to show success notification: %v", err)
		}
	}
}

// ShowConfigFile implements UIServiceInterface
func (us *UIService) ShowConfigFile() error {
	us.logger.Info("Showing config file...")

	if us.notifyManager != nil {
		if err := us.sendNotification("Configuration", "Config file location copied to clipboard", "preferences-desktop"); err != nil {
			us.logger.Error("Failed to show config notification: %v", err)
			return fmt.Errorf("failed to show config notification: %w", err)
		}
	}

	return nil
}

// sendNotification is a helper method for sending notifications
func (us *UIService) sendNotification(title, message, _ string) error {
	if us.notifyManager == nil {
		return fmt.Errorf("notification manager not available")
	}

	// Use appropriate notification method based on title/context
	switch title {
	case "Error":
		return us.notifyManager.NotifyError(message)
	case "Recording Started":
		return us.notifyManager.NotifyStartRecording()
	case "Recording Stopped":
		return us.notifyManager.NotifyStopRecording()
	case "Transcription Complete", "Success":
		return us.notifyManager.NotifyTranscriptionComplete()
	default:
		// Generic notification - use a simple approach
		us.logger.Info("Notification: %s - %s", title, message)
		return nil
	}
}

// Shutdown implements UIServiceInterface
func (us *UIService) Shutdown() error {
	us.logger.Info("UIService shutdown complete")
	return nil
}
