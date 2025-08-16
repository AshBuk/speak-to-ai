package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
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

	// Start asynchronous transcription
	go a.transcribeAsync(audioFile)

	return nil
}

// transcribeAsync performs transcription in a separate goroutine with cancellable context
func (a *App) transcribeAsync(audioFile string) {
	// Create context with timeout for transcription
	ctx, cancel := context.WithTimeout(a.Ctx, 2*time.Minute)
	defer cancel()

	// Channel for transcription result
	type transcriptionResult struct {
		transcript string
		err        error
	}
	resultChan := make(chan transcriptionResult, 1)

	// Start transcription in another goroutine
	go func() {
		transcript, err := a.WhisperEngine.Transcribe(audioFile)
		select {
		case resultChan <- transcriptionResult{transcript: transcript, err: err}:
		case <-ctx.Done():
			// Context cancelled, don't send result
		}
	}()

	// Wait for result or cancellation
	select {
	case result := <-resultChan:
		a.handleTranscriptionResult(result.transcript, result.err)
	case <-ctx.Done():
		a.handleTranscriptionCancellation(ctx.Err())
	}
}

// handleTranscriptionResult handles the result of transcription
func (a *App) handleTranscriptionResult(transcript string, err error) {
	if err != nil {
		// Reset tray state on error
		if a.TrayManager != nil {
			a.TrayManager.SetTooltip("❌ Transcription failed")
		}

		// Show error notification
		if a.NotifyManager != nil {
			a.NotifyManager.ShowNotification("Error", fmt.Sprintf("Transcription failed: %v", err))
		}

		a.Logger.Error("Failed to transcribe audio: %v", err)
		return
	}

	// Store transcript
	a.LastTranscript = transcript

	// Route the transcript according to configured output mode
	if a.OutputManager != nil {
		switch a.Config.Output.DefaultMode {
		case "clipboard":
			if err := a.OutputManager.CopyToClipboard(transcript); err != nil {
				a.Logger.Warning("Failed to copy to clipboard: %v", err)
			}
		case "active_window":
			if err := a.OutputManager.TypeToActiveWindow(transcript); err != nil {
				a.Logger.Warning("Failed to type to active window: %v", err)
			}
		case "combined":
			if err := a.OutputManager.CopyToClipboard(transcript); err != nil {
				a.Logger.Warning("Failed to copy to clipboard: %v", err)
			}
			if err := a.OutputManager.TypeToActiveWindow(transcript); err != nil {
				a.Logger.Warning("Failed to type to active window: %v", err)
			}
		default:
			if err := a.OutputManager.TypeToActiveWindow(transcript); err != nil {
				a.Logger.Warning("Failed to type to active window: %v", err)
			}
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
}

// handleTranscriptionCancellation handles cancellation of transcription
func (a *App) handleTranscriptionCancellation(err error) {
	a.Logger.Warning("Transcription cancelled: %v", err)

	// Reset tray state
	if a.TrayManager != nil {
		a.TrayManager.SetTooltip("⚠️  Transcription cancelled")
	}

	// Show cancellation notification
	if a.NotifyManager != nil {
		a.NotifyManager.ShowNotification("Cancelled", "Transcription was cancelled")
	}
}

// handleShowConfig handles showing the configuration file
func (a *App) handleShowConfig() error {
	a.Logger.Info("Opening configuration file: %s", a.ConfigFile)

	// Show notification about config file location
	if a.NotifyManager != nil {
		a.NotifyManager.ShowNotification("Configuration File", fmt.Sprintf("Opening: %s", a.ConfigFile))
	}

	// Get editor from environment variable
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Fallback to xdg-open
		editor = "xdg-open"
		a.Logger.Debug("$EDITOR not set, using xdg-open as fallback")
	} else {
		a.Logger.Debug("Using editor from $EDITOR: %s", editor)
	}

	// Security: allowlist check on editor
	if !a.Config.IsCommandAllowed(editor) {
		return fmt.Errorf("command not allowed: %s", editor)
	}

	// Check if config file exists
	if _, err := os.Stat(a.ConfigFile); os.IsNotExist(err) {
		errMsg := fmt.Sprintf("Configuration file not found: %s", a.ConfigFile)
		a.Logger.Error(errMsg)
		if a.NotifyManager != nil {
			a.NotifyManager.ShowNotification("Error", errMsg)
		}
		return fmt.Errorf("config file not found: %s", a.ConfigFile)
	}

	// Sanitize args (config file path)
	args := config.SanitizeCommandArgs([]string{a.ConfigFile})
	if len(args) != 1 {
		return fmt.Errorf("invalid config file path")
	}

	// Start editor in background
	cmd := exec.Command(editor, args[0])

	// For GUI applications, detach from parent process
	if editor == "xdg-open" {
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
	}

	err := cmd.Start()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to open config file with %s: %v", editor, err)
		a.Logger.Error(errMsg)
		if a.NotifyManager != nil {
			a.NotifyManager.ShowNotification("Error", errMsg)
		}
		return fmt.Errorf("failed to open config file: %w", err)
	}

	a.Logger.Info("Successfully opened config file with %s", editor)
	return nil
}

// handleReloadConfig handles reloading the configuration
func (a *App) handleReloadConfig() error {
	a.Logger.Info("Reloading configuration from: %s", a.ConfigFile)

	// Show notification about config reload
	if a.NotifyManager != nil {
		a.NotifyManager.ShowNotification("Configuration", "Reloading configuration...")
	}

	// Load new configuration
	newConfig, err := config.LoadConfig(a.ConfigFile)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to reload config: %v", err)
		a.Logger.Error(errMsg)
		if a.NotifyManager != nil {
			a.NotifyManager.ShowNotification("Error", errMsg)
		}
		return fmt.Errorf("failed to reload config: %w", err)
	}

	// Store old config for comparison
	oldConfig := a.Config
	a.Config = newConfig

	// Reinitialize components that depend on configuration
	err = a.reinitializeComponents(oldConfig)
	if err != nil {
		// Rollback to old config on failure
		a.Config = oldConfig
		errMsg := fmt.Sprintf("Failed to reinitialize components: %v", err)
		a.Logger.Error(errMsg)
		if a.NotifyManager != nil {
			a.NotifyManager.ShowNotification("Error", errMsg)
		}
		return fmt.Errorf("failed to reinitialize components: %w", err)
	}

	// Success notification
	if a.NotifyManager != nil {
		a.NotifyManager.ShowNotification("Configuration", "Configuration reloaded successfully!")
	}

	a.Logger.Info("Configuration reloaded successfully")
	return nil
}
