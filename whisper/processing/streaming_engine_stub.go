//go:build !cgo || nocgo

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package processing

import (
	"context"
	"errors"
	"time"

	"github.com/AshBuk/speak-to-ai/whisper/interfaces"
)

type StreamingWhisperEngine struct {
	interfaces.WhisperEngine
}

type TranscriptionResult struct {
	Text        string
	IsConfirmed bool
	Confidence  float64
	Timestamp   time.Time
}

func NewStreamingWhisperEngine(baseEngine interfaces.WhisperEngine) *StreamingWhisperEngine {
	return nil // unavailable: built without cgo
}

func (s *StreamingWhisperEngine) SetPartialResultCallback(callback func(string, bool)) {}

func (s *StreamingWhisperEngine) TranscribeChunk(audioData []float32) (*TranscriptionResult, error) {
	return nil, errors.New("transcription unavailable: built without cgo")
}

func (s *StreamingWhisperEngine) TranscribeStream(ctx context.Context, audioStream <-chan []float32, resultStream chan<- *TranscriptionResult) error {
	return errors.New("transcription unavailable: built without cgo")
}

func (s *StreamingWhisperEngine) Reset() {}

func (s *StreamingWhisperEngine) GetLastConfirmedTranscript() string { return "" }

func (s *StreamingWhisperEngine) SetAgreementThreshold(threshold int) {}
