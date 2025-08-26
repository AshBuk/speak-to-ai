// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
	"time"

	"github.com/AshBuk/speak-to-ai/audio"
)

// handleStartVADRecording handles VAD-based automatic recording
func (a *App) handleStartVADRecording() error {
	a.Logger.Info("Starting VAD recording...")

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
	if a.TrayManager != nil {
		a.TrayManager.SetRecordingState(true)
		a.TrayManager.SetTooltip("🎧 Listening for speech...")
	}

	// Show notification
	if a.NotifyManager != nil {
		a.NotifyManager.ShowNotification("VAD Recording", "Listening for speech. Speak to start recording.")
	}

	// Start VAD processing in background
	go a.processVADStream(audioStream)

	return nil
}

// processVADStream processes audio stream with VAD
func (a *App) processVADStream(audioStream <-chan []float32) {
	a.Logger.Info("Starting VAD stream processing...")

	// Create VAD with configured sensitivity
	vadSensitivity := audio.ParseVADSensitivity(a.Config.Audio.VADSensitivity)
	vad := audio.NewVADWithSensitivity(vadSensitivity)

	speechBuffer := make([][]float32, 0)
	isRecordingSpeech := false
	silenceCounter := 0
	maxSilenceFrames := 20 // ~1 second of silence to stop recording

	for {
		select {
		case chunk, ok := <-audioStream:
			if !ok {
				a.Logger.Info("Audio stream closed")
				return
			}

			// Check for speech activity
			speechActive := vad.IsSpeechActive(chunk)

			if speechActive {
				silenceCounter = 0
				if !isRecordingSpeech {
					// Start recording speech
					isRecordingSpeech = true
					speechBuffer = speechBuffer[:0] // Clear buffer
					a.Logger.Info("Speech detected, starting recording")

					if a.TrayManager != nil {
						a.TrayManager.SetTooltip("🎤 Recording speech...")
					}
					if a.NotifyManager != nil {
						a.NotifyManager.NotifyStartRecording()
					}
				}
				// Add chunk to speech buffer
				speechBuffer = append(speechBuffer, chunk)
			} else if isRecordingSpeech {
				silenceCounter++
				speechBuffer = append(speechBuffer, chunk) // Include some silence

				if silenceCounter >= maxSilenceFrames {
					// End of speech detected
					isRecordingSpeech = false
					a.Logger.Info("End of speech detected, processing...")

					if a.TrayManager != nil {
						a.TrayManager.SetTooltip("🔄 Transcribing...")
					}

					// Process collected speech
					go a.processSpeechBuffer(speechBuffer)
					speechBuffer = speechBuffer[:0]
				}
			}

		case <-a.Ctx.Done():
			a.Logger.Info("VAD processing cancelled")
			return
		}
	}
}

// processSpeechBuffer processes collected speech chunks
func (a *App) processSpeechBuffer(speechBuffer [][]float32) {
	a.Logger.Info("Processing speech buffer with %d chunks", len(speechBuffer))

	if len(speechBuffer) == 0 {
		return
	}

	// Convert chunks to single audio file
	// This would need to be implemented based on your audio format
	// For now, we'll use the existing transcription system

	// Save buffer to temporary file and transcribe
	tempFile, err := a.saveSpeechBufferToFile(speechBuffer)
	if err != nil {
		a.Logger.Error("Failed to save speech buffer: %v", err)
		if a.TrayManager != nil {
			a.TrayManager.SetTooltip("❌ Processing failed")
		}
		return
	}

	// Transcribe the audio
	go a.transcribeAsync(tempFile)
}

// saveSpeechBufferToFile saves speech chunks to a temporary audio file
func (a *App) saveSpeechBufferToFile(speechBuffer [][]float32) (string, error) {
	// This is a simplified implementation
	// In reality, you'd need to properly encode the audio data

	tempFile := fmt.Sprintf("%s/speak-to-ai-vad-%d.wav",
		a.Config.General.TempAudioPath, time.Now().UnixNano())

	// For now, return the temp file path
	// The actual audio encoding would need to be implemented
	// based on your specific audio format requirements

	a.Logger.Debug("Would save speech buffer to: %s", tempFile)
	return tempFile, nil
}
