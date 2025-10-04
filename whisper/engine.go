//go:build cgo

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package whisper

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/utils"
	"github.com/AshBuk/speak-to-ai/whisper/interfaces"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/go-audio/wav"
)

// Provides an interface for interacting with the whisper.cpp model.
// It encapsulates model loading, context management, and the transcription process
type WhisperEngine struct {
	config    *config.Config
	model     whisper.Model
	modelPath string
}

// Return the underlying whisper.Model for advanced or direct interactions
func (w *WhisperEngine) GetModel() interfaces.WhisperModel {
	return w.model
}

// Return the engine's configuration
func (w *WhisperEngine) GetConfig() *config.Config {
	return w.config
}

// Initialize and load a new Whisper model from the given path.
// Return an error if the model file is not found or fails to load
func NewWhisperEngine(config *config.Config, modelPath string) (*WhisperEngine, error) {
	if !utils.IsValidFile(modelPath) {
		return nil, fmt.Errorf("whisper model not found: %s", modelPath)
	}

	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load whisper model: %w", err)
	}

	return &WhisperEngine{
		config:    config,
		model:     model,
		modelPath: modelPath,
	}, nil
}

// Release the resources associated with the loaded Whisper model
func (w *WhisperEngine) Close() error {
	if w.model != nil {
		return w.model.Close()
	}
	return nil
}

// Perform speech-to-text conversion on the given audio file.
// Handle file validation, audio loading, context management, and processing
func (w *WhisperEngine) Transcribe(audioFile string) (string, error) {
	if !utils.IsValidFile(audioFile) {
		return "", fmt.Errorf("audio file not found or invalid: %s", audioFile)
	}

	// Enforce a reasonable file size limit to prevent excessive memory usage
	fileSize, err := utils.GetFileSize(audioFile)
	if err != nil {
		return "", fmt.Errorf("error checking audio file size: %w", err)
	}
	const maxFileSize int64 = 50 * 1024 * 1024 // 50MB
	if fileSize > maxFileSize {
		return "", fmt.Errorf("audio file too large (%d bytes), max allowed is %d bytes", fileSize, maxFileSize)
	}

	// Check for sufficient disk space before proceeding
	if err := utils.CheckDiskSpace(audioFile); err != nil {
		return "", fmt.Errorf("insufficient disk space: %w", err)
	}

	audioData, err := w.loadAudioData(audioFile)
	if err != nil {
		return "", fmt.Errorf("failed to load audio data: %w", err)
	}

	context, err := w.model.NewContext()
	if err != nil {
		return "", fmt.Errorf("failed to create whisper context: %w", err)
	}

	// Set the target language for transcription if specified in the config
	if lang := w.config.General.Language; lang != "" {
		if err := context.SetLanguage(lang); err != nil {
			return "", fmt.Errorf("failed to set language: %w", err)
		}
	}

	if err := context.Process(audioData, nil, nil, nil); err != nil {
		return "", fmt.Errorf("failed to process audio: %w", err)
	}

	// Collect all text segments from the processed audio
	var transcript strings.Builder
	for {
		segment, err := context.NextSegment()
		if err != nil {
			break // End of segments
		}
		transcript.WriteString(segment.Text)
		transcript.WriteString(" ")
	}

	result := strings.TrimSpace(transcript.String())
	result = utils.SanitizeTranscript(result)
	return result, nil
}

// Perform speech-to-text conversion with cancellation support.
// It wraps Transcribe in a goroutine and returns early if the provided context is cancelled.
func (w *WhisperEngine) TranscribeWithContext(ctx context.Context, audioFile string) (string, error) {
	type result struct {
		text string
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		txt, err := w.Transcribe(audioFile)
		ch <- result{text: txt, err: err}
	}()

	select {
	case r := <-ch:
		return r.text, r.err
	case <-ctx.Done():
		return "", fmt.Errorf("transcription cancelled: %w", ctx.Err())
	}
}

// Open a WAV file, decode it, and convert the PCM data
// into the float32 sample format required by the Whisper model
func (w *WhisperEngine) loadAudioData(audioFile string) ([]float32, error) {
	// Sanitize path to prevent directory traversal and other file-based attacks
	clean := filepath.Clean(audioFile)
	if clean != audioFile || strings.Contains(clean, "..") {
		return nil, fmt.Errorf("invalid audio file path")
	}
	// #nosec G304 -- Path is sanitized and validated by the caller.
	file, err := os.Open(clean)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close audio file: %v", err)
		}
	}()

	decoder := wav.NewDecoder(file)
	if decoder == nil {
		return nil, fmt.Errorf("failed to create WAV decoder")
	}

	audioBuffer, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to read audio buffer: %w", err)
	}

	// Convert integer PCM samples to float32 samples normalized to [-1.0, 1.0]
	samples := make([]float32, audioBuffer.NumFrames())
	for i := 0; i < audioBuffer.NumFrames(); i++ {
		intSample := audioBuffer.Data[i]
		samples[i] = float32(intSample) / 32768.0
	}
	return samples, nil
}
