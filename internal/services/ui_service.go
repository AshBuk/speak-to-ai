// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/assets"
	"github.com/AshBuk/speak-to-ai/internal/constants"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/notify"
	"github.com/AshBuk/speak-to-ai/internal/tray"
)

// UIService implements UIServiceInterface
type UIService struct {
	logger        logger.Logger
	trayManager   tray.TrayManagerInterface
	notifyManager *notify.NotificationManager
	config        *config.Config
}

// NewUIService creates a new UIService instance
func NewUIService(
	logger logger.Logger,
	trayManager tray.TrayManagerInterface,
	notifyManager *notify.NotificationManager,
	config *config.Config,
) *UIService {
	return &UIService{
		logger:        logger,
		trayManager:   trayManager,
		notifyManager: notifyManager,
		config:        config,
	}
}

// SetRecordingState implements UIServiceInterface
func (us *UIService) SetRecordingState(isRecording bool) {
	if us.trayManager != nil {
		us.trayManager.SetRecordingState(isRecording)
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

// UpdateSettings updates tray UI with current configuration
func (us *UIService) UpdateSettings(cfg *config.Config) {
	if us.trayManager != nil {
		us.trayManager.UpdateSettings(cfg)
	}
}

// UpdateRecordingUI implements UIServiceInterface
func (us *UIService) UpdateRecordingUI(isRecording bool, level float64) {
	us.SetRecordingState(isRecording)

}

// SetError implements UIServiceInterface
func (us *UIService) SetError(message string) {
	us.logger.Error("UI Error: %s", message)

	if us.notifyManager != nil {
		if err := us.sendNotification(constants.NotifyError, message, "dialog-error"); err != nil {
			us.logger.Warning("Failed to show error notification: %v", err)
		}
	}
}

// SetSuccess implements UIServiceInterface
func (us *UIService) SetSuccess(message string) {
	us.logger.Info("UI Success: %s", message)

	if us.notifyManager != nil {
		if err := us.sendNotification(constants.NotifySuccess, message, "dialog-ok-apply"); err != nil {
			us.logger.Warning("Failed to show success notification: %v", err)
		}
	}
}

// ShowAboutPage opens the About page in the default browser
func (us *UIService) ShowAboutPage() error {
	us.logger.Info("Showing about page...")

	tmpFile, err := os.CreateTemp("", "speak-to-ai-about-*.html")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	// Note: temp file for browser access

	if _, err := tmpFile.WriteString(assets.AboutHTML); err != nil {
		return fmt.Errorf("failed to write HTML content: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	return us.openWithSystem(tmpFile.Name())
}

// ShowConfigFile implements UIServiceInterface
func (us *UIService) ShowConfigFile() error {
	us.logger.Info("Showing config file...")

	path, ok := us.getConfigPath()
	if !ok {
		us.logger.Warning("Could not determine config file path")
		return fmt.Errorf("could not determine config file path")
	}

	// Open with system handler using sanitized environment (AppImage-safe)
	if err := us.openWithSystem(path); err != nil {
		us.logger.Info("Config file open failed, trying directory: %v", err)
		// As a fallback, try to open the containing directory
		dir := filepath.Dir(path)
		if err := us.openWithSystem(dir); err != nil {
			us.logger.Error("Failed to open config directory as fallback: %v", err)
			return fmt.Errorf("failed to open config file or directory: %w", err)
		}
	}

	if us.notifyManager != nil {
		if err := us.sendNotification("Configuration", "Config file opened", "preferences-desktop"); err != nil {
			us.logger.Error("Failed to show config notification: %v", err)
		}
	}

	return nil
}

// Open file or directory via the system default handler.
// It strips AppImage-related environment variables to avoid host app linkage issues.
func (us *UIService) openWithSystem(target string) error {
	// #nosec G204 -- Safe: xdg-open is a system tool; arguments are not shell-interpreted.
	cmd := exec.Command("xdg-open", target)

	// AppImage environment cleanup for host app compatibility
	filtered := make([]string, 0, len(os.Environ()))
	for _, env := range os.Environ() {
		// Skip AppImage-specific variables that break host apps
		if !strings.HasPrefix(env, "LD_LIBRARY_PATH=") &&
			!strings.HasPrefix(env, "LD_PRELOAD=") &&
			!strings.HasPrefix(env, "APPIMAGE=") &&
			!strings.HasPrefix(env, "APPDIR=") &&
			!strings.HasPrefix(env, "ARGV0=") {
			filtered = append(filtered, env)
		}
	}
	cmd.Env = filtered

	return cmd.Start()
}

// getConfigPath retrieves the effective config file path
func (us *UIService) getConfigPath() (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", false
	}
	appImagePath := filepath.Join(home, ".config", "speak-to-ai", "config.yaml")
	flatpakPath := filepath.Join(home, ".var", "app", "io.github.ashbuk.speak-to-ai", "config", "speak-to-ai", "config.yaml")

	// Prefer existing file
	if _, err := os.Stat(appImagePath); err == nil {
		return appImagePath, true
	}
	if _, err := os.Stat(flatpakPath); err == nil {
		return flatpakPath, true
	}
	// Fallback to default AppImage path even if not present (opener will show dialog)
	return appImagePath, true
}

// sendNotification is a helper method for sending notifications
func (us *UIService) sendNotification(title, message, _ string) error {
	if us.notifyManager == nil {
		return fmt.Errorf("notification manager not available")
	}

	// Use appropriate notification method based on title/context
	switch title {
	case constants.NotifyError:
		return us.notifyManager.NotifyError(message)
	case constants.NotifyRecordingStarted:
		return us.notifyManager.NotifyStartRecording()
	case constants.NotifyRecordingStopped:
		return us.notifyManager.NotifyStopRecording()
	case constants.NotifyTranscriptionDone, constants.NotifySuccess:
		return us.notifyManager.NotifyTranscriptionComplete()
	case constants.NotifyConfigReset:
		return us.notifyManager.NotifyConfigurationReset()
	default:
		// Generic notification - show to user
		us.logger.Info("Notification: %s - %s", title, message)
		return us.notifyManager.ShowNotification(title, message)
	}
}

// Shutdown implements UIServiceInterface
func (us *UIService) Shutdown() error {
	us.logger.Info("UIService shutdown complete")
	return nil
}
