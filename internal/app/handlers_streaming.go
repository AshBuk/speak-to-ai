// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/utils"
	"github.com/AshBuk/speak-to-ai/whisper"
)

// handleStartStreamingRecording handles streaming transcription mode
func (a *App) handleStartStreamingRecording() error {
	a.Logger.Info("Starting streaming recording...")

	// Check if recorder supports streaming
	if !a.Recorder.UseStreaming() {
		a.Logger.Warning("Recorder doesn't support streaming, falling back to manual mode")
		return fmt.Errorf("streaming not supported, use manual recording mode")
	}

	// Start streaming recording
	audioStream, err := a.Recorder.StartStreamingRecording()
	if err != nil {
		return fmt.Errorf("failed to start streaming recording: %w", err)
	}

	// Update tray state
	a.updateUIState(true, "🎤 Streaming transcription active...")

	// Show notification
	if a.NotifyManager != nil {
		a.notify("Streaming Mode", "Real-time transcription started. Speak normally.")
	}

	// Start streaming processing in background
	go a.processStreamingTranscription(audioStream)

	return nil
}

// processStreamingTranscription processes streaming transcription
func (a *App) processStreamingTranscription(audioStream <-chan []float32) {
	a.Logger.Info("Starting streaming transcription processing...")

	// Create result channel
	resultStream := make(chan *whisper.TranscriptionResult, 10)

	// Set up partial result callback
	a.StreamingEngine.SetPartialResultCallback(func(text string, isConfirmed bool) {
		if text == "" {
			return
		}

		// Update tray with current partial result
		if a.TrayManager != nil {
			var status string
			if isConfirmed {
				status = "✅"
			} else {
				status = "🔄"
			}
			a.TrayManager.SetTooltip(fmt.Sprintf("%s %s", status, text[:min(50, len(text))]))
		}
	})

	// Start transcription in background
	go func() {
		if err := a.StreamingEngine.TranscribeStream(a.Ctx, audioStream, resultStream); err != nil {
			a.Logger.Error("Streaming transcription error: %v", err)
		}
	}()

	// Process results
	for {
		select {
		case result, ok := <-resultStream:
			if !ok {
				a.Logger.Info("Streaming transcription completed")
				return
			}

			// Process confirmed results
			if result.IsConfirmed && result.Text != "" {
				a.handleConfirmedTranscription(result.Text)
			}

		case <-a.Ctx.Done():
			a.Logger.Info("Streaming transcription cancelled")
			return
		}
	}
}

// handleConfirmedTranscription processes confirmed transcription results
func (a *App) handleConfirmedTranscription(text string) {
	sanitized := utils.SanitizeTranscript(text)
	if sanitized == "" {
		if a.TrayManager != nil {
			a.TrayManager.SetTooltip("✅ Ready")
		}
		if a.NotifyManager != nil {
			a.notify("No Speech", "No speech detected in recording")
		}
		a.Logger.Info("Confirmed transcription: <empty>")
		return
	}

	a.Logger.Info("Confirmed transcription: %s", sanitized)

	// Store transcript
	a.LastTranscript = sanitized

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
			}
		}
	}

	// Show completion notification
	if a.NotifyManager != nil {
		if err := a.NotifyManager.NotifyTranscriptionComplete(); err != nil {
			a.Logger.Warning("failed to show transcription complete notification: %v", err)
		}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
