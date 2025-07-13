package notify

import (
	"fmt"
	"os/exec"
)

// NotificationManager handles desktop notifications
type NotificationManager struct {
	appName string
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(appName string) *NotificationManager {
	return &NotificationManager{
		appName: appName,
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
	cmd := exec.Command("notify-send",
		"--app-name", nm.appName,
		"--icon", icon,
		summary, body)

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
