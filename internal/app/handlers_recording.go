// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/utils"
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

	// Lazy reinitialization of audio recorder if method changed (auto-fallback)
	if err := a.ensureAudioRecorderAvailable(); err != nil {
		a.Logger.Error("Audio recorder not available: %v", err)
		if a.TrayManager != nil {
			a.TrayManager.SetTooltip("❌ Audio recorder unavailable")
		}
		return fmt.Errorf("audio recorder not available: %w", err)
	}

	// If VAD auto start/stop is enabled, use streaming mode
	if a.Config.Audio.EnableVAD && a.Config.Audio.AutoStartStop {
		return a.handleStartVADRecording()
	}

	// If streaming is enabled, use streaming transcription
	if a.Config.Audio.EnableStreaming && a.StreamingEngine != nil {
		return a.handleStartStreamingRecording()
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
		if err := a.NotifyManager.NotifyStartRecording(); err != nil {
			a.Logger.Warning("failed to show start recording notification: %v", err)
		}
	}

	return nil
}

// handleStopRecordingAndTranscribe handles stopping recording and transcription
func (a *App) handleStopRecordingAndTranscribe() error {
	a.Logger.Info("Stopping recording and transcribing...")

	// Stop recording
	audioFile, err := a.Recorder.StopRecording()
	if err != nil {
		a.Logger.Warning("StopRecording returned error: %v", err)
		// Keep async to avoid potential deadlock, but centralize logic
		go a.handleRecordingError(err)

		// Auto-fallback to arecord if using ffmpeg (deferred to avoid deadlock)
		if a.Config.Audio.RecordingMethod == "ffmpeg" {
			// Simply update config and set reinitialization flag
			a.Config.Audio.RecordingMethod = "arecord"
			a.audioRecorderNeedsReinit = true

			// Update tray settings asynchronously to reflect selection immediately
			if a.TrayManager != nil {
				go a.TrayManager.UpdateSettings(a.Config)
			}
			// Persist new recorder selection
			if err := config.SaveConfig(a.ConfigFile, a.Config); err != nil {
				a.Logger.Warning("failed to save config after fallback: %v", err)
			}

			if a.NotifyManager != nil {
				go a.notify("Audio Fallback", "Switched to arecord due to ffmpeg capture error. Try recording again.")
			}
			a.Logger.Info("Auto-fallback: switched to arecord due to ffmpeg failure")
		}

		return fmt.Errorf("failed to stop recording: %w", err)
	}

	// Update tray state
	if a.TrayManager != nil {
		a.TrayManager.SetRecordingState(false)
		a.TrayManager.SetTooltip("🔄 Transcribing...")
	}

	// Show notification
	if a.NotifyManager != nil {
		if err := a.NotifyManager.NotifyStopRecording(); err != nil {
			a.Logger.Warning("failed to show stop recording notification: %v", err)
		}
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
			if err := a.NotifyManager.ShowNotification("Error", fmt.Sprintf("Transcription failed: %v", err)); err != nil {
				a.Logger.Warning("failed to show notification: %v", err)
			}
		}

		a.Logger.Error("Failed to transcribe audio: %v", err)
		return
	}

	// Sanitize and store transcript
	sanitized := utils.SanitizeTranscript(transcript)
	a.LastTranscript = sanitized

	// Do not output empty transcripts
	if sanitized == "" {
		if a.TrayManager != nil {
			a.TrayManager.SetTooltip("✅ Ready")
		}
		if a.NotifyManager != nil {
			_ = a.NotifyManager.ShowNotification("No Speech", "No speech detected in recording")
		}
		a.Logger.Info("Transcription completed: <empty>")
		return
	}

	// Route the transcript according to configured output mode
	if a.OutputManager != nil {
		switch a.Config.Output.DefaultMode {
		case config.OutputModeClipboard:
			if err := a.OutputManager.CopyToClipboard(sanitized); err != nil {
				a.Logger.Warning("Failed to copy to clipboard: %v", err)
			}
		case config.OutputModeActiveWindow:
			if err := a.OutputManager.TypeToActiveWindow(sanitized); err != nil {
				a.Logger.Warning("Failed to type to active window: %v", err)
				// Fallback to clipboard if typing fails
				a.Logger.Info("Falling back to clipboard output")
				if clipErr := a.OutputManager.CopyToClipboard(sanitized); clipErr != nil {
					a.Logger.Warning("Clipboard fallback also failed: %v", clipErr)
					if a.NotifyManager != nil {
						_ = a.NotifyManager.ShowNotification("Output Failed", "Both typing and clipboard failed. Check output configuration.")
					}
				} else {
					if a.NotifyManager != nil && a.Config.Notifications.EnableWorkflowNotifications {
						_ = a.NotifyManager.ShowNotification("Output via Clipboard", "Typing not supported by compositor. Text copied to clipboard - press Ctrl+V to paste.")
					}
				}
			}
		case config.OutputModeCombined:
			if err := a.OutputManager.CopyToClipboard(sanitized); err != nil {
				a.Logger.Warning("Failed to copy to clipboard: %v", err)
			}
			if err := a.OutputManager.TypeToActiveWindow(sanitized); err != nil {
				a.Logger.Warning("Failed to type to active window: %v", err)
			}
		default:
			if err := a.OutputManager.TypeToActiveWindow(sanitized); err != nil {
				a.Logger.Warning("Failed to type to active window: %v", err)
				// Fallback to clipboard if typing fails
				a.Logger.Info("Falling back to clipboard output")
				if clipErr := a.OutputManager.CopyToClipboard(sanitized); clipErr != nil {
					a.Logger.Warning("Clipboard fallback also failed: %v", clipErr)
				} else {
					if a.NotifyManager != nil {
						_ = a.NotifyManager.ShowNotification("Output via Clipboard", "Text copied to clipboard - press Ctrl+V to paste.")
					}
				}
			}
		}
	}

	// Reset tray state
	if a.TrayManager != nil {
		a.TrayManager.SetTooltip("✅ Ready")
	}

	// Show completion notification
	if a.NotifyManager != nil {
		if err := a.NotifyManager.NotifyTranscriptionComplete(); err != nil {
			a.Logger.Warning("failed to show transcription complete notification: %v", err)
		}
	}

	a.Logger.Info("Transcription completed: %s", sanitized)
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
		if err := a.NotifyManager.ShowNotification("Cancelled", "Transcription was cancelled"); err != nil {
			a.Logger.Warning("failed to show notification: %v", err)
		}
	}
}
