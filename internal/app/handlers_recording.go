// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/constants"
	"github.com/AshBuk/speak-to-ai/internal/utils"
)

// handleStartRecording handles the start of recording
func (a *App) handleStartRecording() error {
	a.Logger.Info("Starting recording...")

	// Ensure model is available (lazy loading)
	if err := a.ensureModelAvailable(); err != nil {
		a.Logger.Error("Model not available: %v", err)
		a.setUIError(constants.MsgModelUnavailable)
		return fmt.Errorf("model not available: %w", err)
	}

	// Lazy reinitialization of audio recorder if method changed (auto-fallback)
	if err := a.ensureAudioRecorderAvailable(); err != nil {
		a.Logger.Error("Audio recorder not available: %v", err)
		a.setUIError(constants.MsgRecorderUnavailable)
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

			a.setUIRecording(levelPercentage)
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
		a.setUIProcessing(constants.MsgTranscribing)
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
		a.handleTranscriptionError(err)
		return
	}

	// Sanitize and store transcript
	sanitized := utils.SanitizeTranscript(transcript)
	a.LastTranscript = sanitized

	if sanitized == "" {
		a.handleEmptyTranscript()
		return
	}

	a.routeTranscriptOutput(sanitized)
	a.finalizeTranscription()
}

// handleTranscriptionError handles transcription errors
func (a *App) handleTranscriptionError(err error) {
	a.setUIError(constants.MsgTranscriptionFailed)
	a.notifyError(err)
	a.Logger.Error("Failed to transcribe audio: %v", err)
}

// handleEmptyTranscript handles empty transcript results
func (a *App) handleEmptyTranscript() {
	a.setUIReady()
	a.notifyInfo(constants.NotifyNoSpeech, constants.MsgTranscriptionEmpty)
	a.Logger.Info("Transcription completed: <empty>")
}

// routeTranscriptOutput routes transcript to configured output mode
func (a *App) routeTranscriptOutput(text string) {
	if a.OutputManager == nil {
		return
	}

	switch a.Config.Output.DefaultMode {
	case config.OutputModeClipboard:
		a.outputToClipboard(text)
	case config.OutputModeActiveWindow:
		a.outputToActiveWindow(text)
	case config.OutputModeCombined:
		a.outputToBoth(text)
	default:
		a.outputWithFallback(text)
	}
}

// outputToClipboard copies text to clipboard
func (a *App) outputToClipboard(text string) {
	if err := a.OutputManager.CopyToClipboard(text); err != nil {
		a.Logger.Warning("Failed to copy to clipboard: %v", err)
	}
}

// outputToActiveWindow types text to active window with fallback
func (a *App) outputToActiveWindow(text string) {
	if err := a.OutputManager.TypeToActiveWindow(text); err != nil {
		a.Logger.Warning("Failed to type to active window: %v", err)
		a.handleTypingFallback(text)
	}
}

// outputToBoth outputs to both clipboard and active window
func (a *App) outputToBoth(text string) {
	a.outputToClipboard(text)
	if err := a.OutputManager.TypeToActiveWindow(text); err != nil {
		a.Logger.Warning("Failed to type to active window: %v", err)
	}
}

// outputWithFallback outputs with automatic fallback to clipboard
func (a *App) outputWithFallback(text string) {
	if err := a.OutputManager.TypeToActiveWindow(text); err != nil {
		a.Logger.Warning("Failed to type to active window: %v", err)
		a.handleTypingFallback(text)
	}
}

// handleTypingFallback handles fallback when typing fails
func (a *App) handleTypingFallback(text string) {
	a.Logger.Info("Falling back to clipboard output")
	if clipErr := a.OutputManager.CopyToClipboard(text); clipErr != nil {
		a.Logger.Warning("Clipboard fallback also failed: %v", clipErr)
		a.notifyError(fmt.Errorf(constants.NotifyOutputBothFailed))
	} else {
		if a.Config.Notifications.EnableWorkflowNotifications {
			a.notifyInfo(constants.NotifyClipboard, constants.NotifyTypingFallback)
		}
	}
}

// finalizeTranscription completes the transcription process
func (a *App) finalizeTranscription() {
	a.setUIReady()

	// Show completion notification
	if a.NotifyManager != nil {
		if err := a.NotifyManager.NotifyTranscriptionComplete(); err != nil {
			a.Logger.Warning("failed to show transcription complete notification: %v", err)
		}
	}

	a.Logger.Info("Transcription completed: %s", a.LastTranscript)
}

// handleTranscriptionCancellation handles cancellation of transcription
func (a *App) handleTranscriptionCancellation(err error) {
	a.Logger.Warning("Transcription cancelled: %v", err)

	a.setUIWarning(constants.MsgTranscriptionCancelled)
	a.notifyInfo(constants.NotifyCancelled, constants.NotifyTranscriptionCancelled)
}
