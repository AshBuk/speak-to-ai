// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
	"time"

	"github.com/AshBuk/speak-to-ai/audio"
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
	// Define shared functions for tray manager
	exitFunc := func() {
		a.Cancel() // Trigger application shutdown
	}
	toggleFunc := func() error {
		if a.HotkeyManager.IsRecording() {
			return a.HotkeyManager.SimulateHotkeyPress("stop_recording")
		}
		return a.HotkeyManager.SimulateHotkeyPress("start_recording")
	}
	showConfigFunc := func() error {
		return a.handleShowConfig()
	}
	reloadConfigFunc := func() error {
		return a.handleReloadConfig()
	}

	// Fallback to mock tray if StatusNotifier watcher is not available
	if !platform.HasStatusNotifierWatcher() {
		a.Logger.Info("StatusNotifier watcher not found; using mock tray")
		if a.NotifyManager != nil {
			msg := "System tray support is not available. On GNOME, install and enable the AppIndicator extension."
			if err := a.NotifyManager.ShowNotification("ℹ️ Speak-to-AI", msg); err != nil {
				a.Logger.Warning("Failed to show notification: %v", err)
			}
		}
		a.TrayManager = tray.CreateMockTrayManager(exitFunc, toggleFunc, showConfigFunc, reloadConfigFunc)
		return
	}

	// Create the tray manager with configuration
	a.TrayManager = tray.CreateTrayManagerWithConfig(a.Config, exitFunc, toggleFunc, showConfigFunc, reloadConfigFunc)

	// Wire audio actions: recorder selection + test recording
	a.TrayManager.SetAudioActions(
		func(method string) error {
			// Update config and reinitialize recorder at runtime
			old := a.Config
			newCfg := *a.Config
			newCfg.Audio.RecordingMethod = method
			a.Config = &newCfg
			if err := a.reinitializeComponents(old); err != nil {
				// rollback on error
				a.Config = old
				return err
			}
			return nil
		},
		func() error {
			// Perform a short 3s test recording and show result via notification
			recorder, err := audio.GetRecorder(a.Config)
			if err != nil {
				return err
			}
			// Start
			if err := recorder.StartRecording(); err != nil {
				return err
			}
			// Sleep for ~3s non-blocking using context
			select {
			case <-a.Ctx.Done():
				return a.Ctx.Err()
			case <-time.After(3 * time.Second):
			}
			// Stop
			file, err := recorder.StopRecording()
			if err != nil {
				return err
			}
			// Notify
			if a.NotifyManager != nil {
				_ = a.NotifyManager.ShowNotification("Audio Test", fmt.Sprintf("Saved test recording: %s", file))
			}
			return nil
		},
	)
}
