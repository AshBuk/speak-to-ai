// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/internal/platform"
	"github.com/AshBuk/speak-to-ai/internal/tray"
)

// ensureModelAvailable ensures the model is available, downloading if necessary
func (a *App) ensureModelAvailable() error {
	// Check if model already exists
	if _, err := a.ModelManager.GetModelPath(); err == nil {
		return nil
	}

	// Model doesn't exist, need to download
	a.Logger.Info("Model not found, downloading...")

	// Show notification about download starting
	if a.NotifyManager != nil {
		if err := a.NotifyManager.ShowNotification("Speak-to-AI", "Downloading Whisper model for first use..."); err != nil {
			a.Logger.Warning("Failed to show notification: %v", err)
		}
	}

	// Create progress callback
	progressCallback := func(downloaded, total int64, percentage float64) {
		// Format sizes for display
		downloadedMB := float64(downloaded) / (1024 * 1024)
		totalMB := float64(total) / (1024 * 1024)

		// Log progress
		a.Logger.Info("Download progress: %.1f%% (%.1f MB / %.1f MB)",
			percentage, downloadedMB, totalMB)

		// Update tray tooltip if available
		if a.TrayManager != nil {
			status := fmt.Sprintf("📥 Downloading model: %.1f%%", percentage)
			a.TrayManager.SetTooltip(status)
		}
	}

	// Download with progress
	modelPath, err := a.ModelManager.GetModelPathWithProgress(progressCallback)
	if err != nil {
		// Reset tray tooltip on error
		if a.TrayManager != nil {
			a.TrayManager.SetTooltip("❌ Model download failed")
		}
		return fmt.Errorf("failed to download model: %w", err)
	}

	// Show completion notification
	if a.NotifyManager != nil {
		if err := a.NotifyManager.ShowNotification("Speak-to-AI", "Model downloaded successfully!"); err != nil {
			a.Logger.Warning("Failed to show notification: %v", err)
		}
	}

	// Reset tray tooltip
	if a.TrayManager != nil {
		a.TrayManager.SetTooltip("✅ Ready")
	}

	a.Logger.Info("Model downloaded successfully: %s", modelPath)
	return nil
}

// initializeTrayManager initializes the system tray manager
func (a *App) initializeTrayManager() {
	// Check if GNOME+Wayland without AppIndicator extension
	if platform.IsGNOMEWithWayland() && !platform.CheckAppIndicatorExtension() {
		a.Logger.Info("GNOME with Wayland detected without AppIndicator extension")

		// Show helpful notification to user
		if a.NotifyManager != nil {
			message := "System tray requires extension. Install with:\nsudo dnf install gnome-shell-extension-appindicator\n\nThen enable in Extensions app or run:\ngnome-extensions enable appindicatorsupport@rgcjonas.gmail.com"
			if err := a.NotifyManager.ShowNotification("ℹ️ Speak-to-AI Setup", message); err != nil {
				a.Logger.Warning("Failed to show setup notification: %v", err)
			}
		}

		// Also log instructions
		a.Logger.Info("To enable system tray icon:")
		a.Logger.Info("1. Install: sudo dnf install gnome-shell-extension-appindicator")
		a.Logger.Info("2. Enable: gnome-extensions enable appindicatorsupport@rgcjonas.gmail.com")
		a.Logger.Info("3. Or use GNOME Extensions app")
		a.Logger.Info("Using hotkeys for now: %s to toggle recording", a.Config.Hotkeys.StartRecording)
	}

	// Create a toggle function for the tray
	toggleFunc := func() error {
		if a.HotkeyManager.IsRecording() {
			return a.HotkeyManager.SimulateHotkeyPress("stop_recording")
		}
		return a.HotkeyManager.SimulateHotkeyPress("start_recording")
	}

	// Create exit function
	exitFunc := func() {
		a.Cancel() // Trigger application shutdown
	}

	// Create show config function
	showConfigFunc := func() error {
		return a.handleShowConfig()
	}

	// Create reload config function
	reloadConfigFunc := func() error {
		return a.handleReloadConfig()
	}

	// Create the appropriate tray manager with configuration
	a.TrayManager = tray.CreateTrayManagerWithConfig(a.Config, exitFunc, toggleFunc, showConfigFunc, reloadConfigFunc)
}
