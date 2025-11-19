// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package notify

import (
	"fmt"
	"os/exec"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/constants"
)

// Manages the sending of desktop notifications
type NotificationManager struct {
	appName string
	config  *config.Config
}

// Create a new notification manager
func NewNotificationManager(appName string, cfg *config.Config) *NotificationManager {
	return &NotificationManager{
		appName: appName,
		config:  cfg,
	}
}

// Show a notification when recording starts
func (nm *NotificationManager) NotifyStartRecording() error {
	if !nm.config.Notifications.EnableWorkflowNotifications {
		return nil // Skip if workflow notifications are disabled
	}
	return nm.sendNotification(constants.NotifyTitleRecordingStart, constants.NotifyRecordingStartMsg, "notification-microphone-sensitivity-high")
}

// Show a notification when recording stops
func (nm *NotificationManager) NotifyStopRecording() error {
	if !nm.config.Notifications.EnableWorkflowNotifications {
		return nil // Skip if workflow notifications are disabled
	}
	return nm.sendNotification(constants.NotifyTitleRecordingStop, constants.NotifyRecordingStopMsg, "notification-microphone-sensitivity-muted")
}

// Show a notification when transcription is complete
func (nm *NotificationManager) NotifyTranscriptionComplete() error {
	if !nm.config.Notifications.EnableWorkflowNotifications {
		return nil // Skip if workflow notifications are disabled
	}
	// Select message based on current output mode
	body := constants.NotifyTranscriptionMsg
	if nm.config.Output.DefaultMode == config.OutputModeActiveWindow {
		body = constants.NotifyTranscriptionTypedMsg
	}
	return nm.sendNotification(constants.NotifyTitleTranscription, body, "edit-copy")
}

// Show an error notification
func (nm *NotificationManager) NotifyError(errMsg string) error {
	return nm.sendNotification(constants.NotifyTitleError, errMsg, "dialog-error")
}

// Show a notification when the configuration is reset to defaults
func (nm *NotificationManager) NotifyConfigurationReset() error {
	return nm.sendNotification(constants.NotifyTitleConfigReset, constants.NotifyConfigResetSuccess, "preferences-system")
}

// Show a generic notification
func (nm *NotificationManager) ShowNotification(summary, body string) error {
	return nm.sendNotification(summary, body, "dialog-information")
}

// Send a notification using the notify-send command
func (nm *NotificationManager) sendNotification(summary, body, icon string) error {
	// Security: validate command before execution
	if !config.IsCommandAllowed(nm.config, "notify-send") {
		return fmt.Errorf("notify-send command not allowed")
	}

	args := []string{
		"--app-name", nm.appName,
		"--icon", icon,
		summary, body,
	}

	// Security: sanitize arguments
	safeArgs := config.SanitizeCommandArgs(args)
	// #nosec G204 -- Safe: notify-send is from an allowlist and arguments are sanitized
	cmd := exec.Command("notify-send", safeArgs...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// Check if the notification system is available
func (nm *NotificationManager) IsAvailable() bool {
	_, err := exec.LookPath("notify-send")
	return err == nil
}
