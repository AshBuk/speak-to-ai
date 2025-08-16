package notify

import (
	"fmt"
	"os/exec"

	"github.com/AshBuk/speak-to-ai/config"
)

// NotificationManager handles desktop notifications
type NotificationManager struct {
	appName string
	config  *config.Config
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(appName string, cfg *config.Config) *NotificationManager {
	return &NotificationManager{
		appName: appName,
		config:  cfg,
	}
}

// NotifyStartRecording shows a notification when recording starts
func (nm *NotificationManager) NotifyStartRecording() error {
	return nm.sendNotification("🎤 Speak-to-AI", "Recording started", "notification-microphone-sensitivity-high")
}

// NotifyStopRecording shows a notification when recording stops
func (nm *NotificationManager) NotifyStopRecording() error {
	return nm.sendNotification("🛑 Recording stopped", "Transcribing audio...", "notification-microphone-sensitivity-muted")
}

// NotifyTranscriptionComplete shows a notification when transcription is complete
func (nm *NotificationManager) NotifyTranscriptionComplete() error {
	return nm.sendNotification("✅ Transcription complete", "Text copied to clipboard", "edit-copy")
}

// NotifyError shows an error notification
func (nm *NotificationManager) NotifyError(errMsg string) error {
	return nm.sendNotification("❌ Error", errMsg, "dialog-error")
}

// ShowNotification shows a generic notification
func (nm *NotificationManager) ShowNotification(summary, body string) error {
	return nm.sendNotification(summary, body, "dialog-information")
}

// sendNotification sends a notification with the given parameters
func (nm *NotificationManager) sendNotification(summary, body, icon string) error {
	// Security: validate command before execution
	if !nm.config.IsCommandAllowed("notify-send") {
		return fmt.Errorf("notify-send command not allowed")
	}

	args := []string{
		"--app-name", nm.appName,
		"--icon", icon,
		summary, body,
	}

	// Security: sanitize arguments
	safeArgs := config.SanitizeCommandArgs(args)
	cmd := exec.Command("notify-send", safeArgs...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// IsAvailable checks if the notification system is available
func (nm *NotificationManager) IsAvailable() bool {
	_, err := exec.LookPath("notify-send")
	return err == nil
}
