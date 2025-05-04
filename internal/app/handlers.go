package app

import (
	"context"
	"time"
)

// handleStartRecording handles starting the recording
func (a *App) handleStartRecording() error {
	a.Logger.Info("Starting recording...")

	// Show notification
	if a.NotifyManager != nil && a.NotifyManager.IsAvailable() {
		if err := a.NotifyManager.NotifyStartRecording(); err != nil {
			a.Logger.Warning("Failed to show notification: %v", err)
		}
	}

	// Update tray icon
	if a.TrayManager != nil {
		a.TrayManager.SetRecordingState(true)
	}

	return a.Recorder.StartRecording()
}

// handleStopRecordingAndTranscribe handles stopping recording and transcribing the audio
func (a *App) handleStopRecordingAndTranscribe() error {
	a.Logger.Info("Stopping recording...")

	// Show notification
	if a.NotifyManager != nil && a.NotifyManager.IsAvailable() {
		if err := a.NotifyManager.NotifyStopRecording(); err != nil {
			a.Logger.Warning("Failed to show notification: %v", err)
		}
	}

	// Update tray icon
	if a.TrayManager != nil {
		a.TrayManager.SetRecordingState(false)
	}

	audioFile, err := a.Recorder.StopRecording()
	if err != nil {
		return err
	}

	// Process audio with whisper
	a.Logger.Info("Processing audio file: %s", audioFile)

	// Set a reasonable timeout for processing (e.g. 30 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use a channel to collect the result or error
	type result struct {
		transcript string
		err        error
	}

	resultCh := make(chan result, 1)

	go func() {
		transcript, err := a.WhisperEngine.Transcribe(audioFile)
		resultCh <- result{transcript, err}
	}()

	// Wait for result or timeout
	select {
	case r := <-resultCh:
		if r.err != nil {
			a.Logger.Error("Error processing audio: %v", r.err)

			// Show error notification
			if a.NotifyManager != nil && a.NotifyManager.IsAvailable() {
				a.NotifyManager.NotifyError("Failed to process audio")
			}

			return r.err
		}

		a.Logger.Info("Transcript: %s", r.transcript)

		// Send transcript to WebSocket clients if server is enabled
		if a.Config.WebServer.Enabled {
			a.WebSocketServer.BroadcastMessage("transcript", r.transcript)
		}

		// Store transcript for clipboard/paste operations
		a.LastTranscript = r.transcript

		// Automatically copy to clipboard if enabled
		if a.OutputManager != nil && a.Config.Output.DefaultMode == "clipboard" {
			if err := a.OutputManager.CopyToClipboard(r.transcript); err != nil {
				a.Logger.Warning("Failed to copy transcript to clipboard: %v", err)
			} else {
				// Show success notification
				if a.NotifyManager != nil && a.NotifyManager.IsAvailable() {
					a.NotifyManager.NotifyTranscriptionComplete()
				}
			}
		}

	case <-ctx.Done():
		// Show timeout notification
		if a.NotifyManager != nil && a.NotifyManager.IsAvailable() {
			a.NotifyManager.NotifyError("Transcription timed out")
		}

		return ctx.Err()
	}

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
