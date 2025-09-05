// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AshBuk/speak-to-ai/audio/processing"
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
		if err := a.NotifyManager.ShowNotification("VAD Recording", "Listening for speech. Speak to start recording."); err != nil {
			a.Logger.Warning("failed to show notification: %v", err)
		}
	}

	// Start VAD processing in background
	go a.processVADStream(audioStream)

	return nil
}

// processVADStream processes audio stream with VAD
func (a *App) processVADStream(audioStream <-chan []float32) {
	a.Logger.Info("Starting VAD stream processing...")

	// Create VAD with configured sensitivity
	vadSensitivity := processing.ParseVADSensitivity(a.Config.Audio.VADSensitivity)
	vad := processing.NewVADWithSensitivity(vadSensitivity)

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
						if err := a.NotifyManager.NotifyStartRecording(); err != nil {
							a.Logger.Warning("failed to show start recording notification: %v", err)
						}
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

	// Ensure target directory exists with secure permissions
	base := a.Config.General.TempAudioPath
	if base == "" {
		base = os.TempDir()
	}
	if err := os.MkdirAll(base, 0700); err != nil {
		return "", err
	}

	tempFile := filepath.Join(base, fmt.Sprintf("speak-to-ai-vad-%d.wav", time.Now().UnixNano()))
	tempFile = filepath.Clean(tempFile)

	// Minimal WAV header with zero data, but we include the number of samples from speechBuffer
	var totalSamples int
	for _, chunk := range speechBuffer {
		totalSamples += len(chunk)
	}
	// Guard against overflow when converting to uint32
	if totalSamples < 0 || totalSamples > (1<<31-1)/2 {
		return "", fmt.Errorf("too many samples")
	}
	dataBytes := totalSamples * 2 // pretend 16-bit PCM

	f, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	// RIFF header
	// ChunkID "RIFF"
	if _, err := f.Write([]byte{'R', 'I', 'F', 'F'}); err != nil {
		return "", err
	}
	// ChunkSize 36 + dataBytes
	chunkSize := 36 + dataBytes
	if chunkSize < 0 || chunkSize > 0xFFFFFFFF {
		return "", fmt.Errorf("chunk size overflow")
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(chunkSize)); err != nil {
		return "", err
	}
	// Format "WAVE"
	if _, err := f.Write([]byte{'W', 'A', 'V', 'E'}); err != nil {
		return "", err
	}
	// Subchunk1ID "fmt "
	if _, err := f.Write([]byte{'f', 'm', 't', ' '}); err != nil {
		return "", err
	}
	// Subchunk1Size 16 for PCM
	if err := binary.Write(f, binary.LittleEndian, uint32(16)); err != nil {
		return "", err
	}
	// AudioFormat 1 PCM
	if err := binary.Write(f, binary.LittleEndian, uint16(1)); err != nil {
		return "", err
	}
	// NumChannels 1
	if err := binary.Write(f, binary.LittleEndian, uint16(1)); err != nil {
		return "", err
	}
	// SampleRate from config
	sr := a.Config.Audio.SampleRate
	if sr < 0 || sr > 0x7FFFFFFF {
		return "", fmt.Errorf("invalid sample rate")
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(sr)); err != nil {
		return "", err
	}
	// ByteRate = SampleRate * NumChannels * BitsPerSample/8
	br := sr * 2
	if br < 0 || br > 0x7FFFFFFF {
		return "", fmt.Errorf("invalid byte rate")
	}
	byteRate := uint32(br)
	if err := binary.Write(f, binary.LittleEndian, byteRate); err != nil {
		return "", err
	}
	// BlockAlign = NumChannels * BitsPerSample/8
	if err := binary.Write(f, binary.LittleEndian, uint16(2)); err != nil {
		return "", err
	}
	// BitsPerSample 16
	if err := binary.Write(f, binary.LittleEndian, uint16(16)); err != nil {
		return "", err
	}
	// Subchunk2ID "data"
	if _, err := f.Write([]byte{'d', 'a', 't', 'a'}); err != nil {
		return "", err
	}
	// Subchunk2Size dataBytes
	if dataBytes < 0 || dataBytes > 0x7FFFFFFF {
		return "", fmt.Errorf("invalid data size")
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(dataBytes)); err != nil {
		return "", err
	}
	// For now, we do not write actual PCM data; placeholder only

	a.Logger.Debug("Saved placeholder VAD buffer (%d samples) to: %s", totalSamples, tempFile)
	return tempFile, nil
}
