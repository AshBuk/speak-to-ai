package app

import (
	"fmt"
)

// handleStartRecording handles the start of recording
func (a *App) handleStartRecording() error {
	a.Logger.Info("Starting recording...")

	// Ensure model is available (lazy loading)
	if err := a.ensureModelAvailable(); err != nil {
		a.Logger.Error("Model not available: %v", err)
		if a.TrayManager != nil {
			a.TrayManager.SetTooltip("❌ Model unavailable")
		}
		return fmt.Errorf("model not available: %w", err)
	}

	// Set up audio level monitoring
	a.Recorder.SetAudioLevelCallback(func(level float64) {
		// Update tray tooltip with audio level
		if a.TrayManager != nil {
			levelPercentage := int(level * 100)
			if levelPercentage > 100 {
				levelPercentage = 100
			}

			// Create visual level indicator
			var levelBar string
			bars := levelPercentage / 10
			for i := 0; i < 10; i++ {
				if i < bars {
					levelBar += "█"
				} else {
					levelBar += "░"
				}
			}

			tooltip := fmt.Sprintf("🎤 Recording... Level: %s %d%%", levelBar, levelPercentage)
			a.TrayManager.SetTooltip(tooltip)
		}

		// Log level for debugging
		a.Logger.Debug("Audio level: %.2f", level)
	})

	// Start recording
	if err := a.Recorder.StartRecording(); err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	// Update tray state
	if a.TrayManager != nil {
		a.TrayManager.SetRecordingState(true)
	}

	// Show notification
	if a.NotifyManager != nil {
		a.NotifyManager.NotifyStartRecording()
	}

	return nil
}

// handleStopRecordingAndTranscribe handles stopping recording and transcription
func (a *App) handleStopRecordingAndTranscribe() error {
	a.Logger.Info("Stopping recording and transcribing...")

	// Stop recording
	audioFile, err := a.Recorder.StopRecording()
	if err != nil {
		return fmt.Errorf("failed to stop recording: %w", err)
	}

	// Update tray state
	if a.TrayManager != nil {
		a.TrayManager.SetRecordingState(false)
		a.TrayManager.SetTooltip("🔄 Transcribing...")
	}

	// Show notification
	if a.NotifyManager != nil {
		a.NotifyManager.NotifyStopRecording()
	}

	// Transcribe audio
	transcript, err := a.WhisperEngine.Transcribe(audioFile)
	if err != nil {
		// Reset tray state on error
		if a.TrayManager != nil {
			a.TrayManager.SetTooltip("❌ Transcription failed")
		}
		return fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Store transcript
	a.LastTranscript = transcript

	// Automatically type the transcript to the active window
	if a.OutputManager != nil {
		if err := a.OutputManager.TypeToActiveWindow(transcript); err != nil {
			a.Logger.Warning("Failed to type to active window: %v", err)
		}
	}

	// Reset tray state
	if a.TrayManager != nil {
		a.TrayManager.SetTooltip("✅ Ready")
	}

	// Show completion notification
	if a.NotifyManager != nil {
		a.NotifyManager.NotifyTranscriptionComplete()
	}

	a.Logger.Info("Transcription completed: %s", transcript)
	return nil
}

// handleCopyToClipboard handles copying the transcript to clipboard
func (a *App) handleCopyToClipboard() error {
	if a.OutputManager == nil {
		a.Logger.Info("Output manager not initialized")
		return nil
	}

	if a.LastTranscript == "" {
		a.Logger.Info("No transcript available to copy")
		return nil
	}

	a.Logger.Info("Copying transcript to clipboard")
	return a.OutputManager.CopyToClipboard(a.LastTranscript)
}

// handlePasteToActiveWindow handles pasting the transcript to the active window
func (a *App) handlePasteToActiveWindow() error {
	if a.OutputManager == nil {
		a.Logger.Info("Output manager not initialized")
		return nil
	}

	if a.LastTranscript == "" {
		a.Logger.Info("No transcript available to paste")
		return nil
	}

	a.Logger.Info("Pasting transcript to active application")
	return a.OutputManager.TypeToActiveWindow(a.LastTranscript)
}

// handleShowConfig handles showing the configuration file
func (a *App) handleShowConfig() error {
	a.Logger.Info("Opening configuration file")

	// Show notification about config file location
	if a.NotifyManager != nil {
		a.NotifyManager.ShowNotification("Configuration File", "Config file: config.yaml")
	}

	// TODO: Open config file in default editor
	// For now, just log the location
	a.Logger.Info("Configuration file location: config.yaml")

	return nil
}

// handleReloadConfig handles reloading the configuration
func (a *App) handleReloadConfig() error {
	a.Logger.Info("Reloading configuration...")

	// Show notification about config reload
	if a.NotifyManager != nil {
		a.NotifyManager.ShowNotification("Configuration", "Reloading configuration...")
	}

	// TODO: Implement config reload logic
	// For now, just log the action
	a.Logger.Info("Configuration reload requested (not yet implemented)")

	return nil
}
