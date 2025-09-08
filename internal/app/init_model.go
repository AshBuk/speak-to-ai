// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
	"time"

	"github.com/AshBuk/speak-to-ai/audio/factory"
	"github.com/AshBuk/speak-to-ai/config"
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
		a.notify("Speak-to-AI", "Downloading Whisper model for first use...")
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
		a.notify("Speak-to-AI", "Model downloaded successfully!")
	}

	// Reset tray tooltip
	if a.TrayManager != nil {
		a.TrayManager.SetTooltip("✅ Ready")
	}

	a.Logger.Info("Model downloaded successfully: %s", modelPath)
	return nil
}

// ensureAudioRecorderAvailable ensures audio recorder matches current config (lazy reinitialization)
func (a *App) ensureAudioRecorderAvailable() error {
	// Add a flag to track if recorder needs reinitialization
	if a.audioRecorderNeedsReinit {
		a.Logger.Info("Audio recorder method changed to %s, reinitializing...", a.Config.Audio.RecordingMethod)

		// Reinitialize audio recorder
		recorder, err := factory.GetRecorder(a.Config)
		if err != nil {
			return fmt.Errorf("failed to reinitialize audio recorder: %w", err)
		}

		a.Recorder = recorder
		a.audioRecorderNeedsReinit = false // Reset flag
		a.Logger.Info("Audio recorder reinitialized successfully with method: %s", a.Config.Audio.RecordingMethod)

		// Update tray settings asynchronously to prevent UI deadlock
		if a.TrayManager != nil {
			go func() {
				a.TrayManager.UpdateSettings(a.Config)
				a.Logger.Info("Tray settings updated after audio recorder change")
			}()
		}
	}

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
			a.notify("ℹ️ Speak-to-AI", msg)
		}
		a.TrayManager = tray.CreateMockTrayManager(exitFunc, toggleFunc, showConfigFunc, reloadConfigFunc)
		return
	}

	// Create the tray manager with configuration
	a.TrayManager = tray.CreateTrayManagerWithConfig(a.Config, exitFunc, toggleFunc, showConfigFunc, reloadConfigFunc)

	// Wire audio actions: recorder selection + test recording
	a.TrayManager.SetAudioActions(
		func(method string) error {
			// Update config and set reinitialization flag
			a.Config.Audio.RecordingMethod = method
			a.audioRecorderNeedsReinit = true

			// Immediately reflect in tray UI
			if a.TrayManager != nil {
				go a.TrayManager.UpdateSettings(a.Config)
			}
			// Persist selection to config file
			if a.ConfigFile != "" {
				if err := config.SaveConfig(a.ConfigFile, a.Config); err != nil {
					a.Logger.Warning("failed to save config after recorder selection: %v", err)
				}
			}

			a.Logger.Info("Audio recorder method changed via tray to: %s", method)
			return nil
		},
		func() error {
			// Perform a short 3s test recording and show result via notification
			recorder, err := factory.GetRecorder(a.Config)
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
				a.notify("Audio Test", fmt.Sprintf("Saved test recording: %s", file))
			}
			return nil
		},
	)
}
