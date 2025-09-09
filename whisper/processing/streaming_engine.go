//go:build cgo

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package processing

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/utils"
	"github.com/AshBuk/speak-to-ai/whisper/interfaces"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// StreamingWhisperEngine provides real-time streaming transcription
type StreamingWhisperEngine struct {
	interfaces.WhisperEngine
	agreementCache     map[string]int     // Cache for iterative agreement
	lastTranscript     string             // Last agreed transcript
	mutex              sync.RWMutex       // Protects agreement cache
	agreementThreshold int                // Minimum agreements for consensus
	onPartialResult    func(string, bool) // Callback for partial results (text, isConfirmed)

	// Direct access to whisper components for streaming
	config *config.Config
	model  whisper.Model
}

// TranscriptionResult holds both partial and confirmed transcription results
type TranscriptionResult struct {
	Text        string
	IsConfirmed bool
	Confidence  float64
	Timestamp   time.Time
}

// NewStreamingWhisperEngine creates a new streaming whisper engine
func NewStreamingWhisperEngine(baseEngine interfaces.WhisperEngine) *StreamingWhisperEngine {
	// Extract model and config from base engine
	modelInterface := baseEngine.GetModel()
	cfg := baseEngine.GetConfig()

	whisperModel, modelOk := modelInterface.(whisper.Model)

	if !modelOk || whisperModel == nil || cfg == nil {
		// Return a minimal engine that will fail gracefully
		return &StreamingWhisperEngine{
			WhisperEngine:      baseEngine,
			agreementCache:     make(map[string]int),
			agreementThreshold: 2,
			mutex:              sync.RWMutex{},
			config:             cfg,          // May be nil but that's okay
			model:              whisperModel, // May be nil but better than crashing
		}
	}

	return &StreamingWhisperEngine{
		WhisperEngine:      baseEngine,
		agreementCache:     make(map[string]int),
		agreementThreshold: 2, // Require 2 consecutive agreements
		mutex:              sync.RWMutex{},
		config:             cfg,
		model:              whisperModel,
	}
}

// SetPartialResultCallback sets callback for receiving partial transcription results
func (s *StreamingWhisperEngine) SetPartialResultCallback(callback func(string, bool)) {
	s.onPartialResult = callback
}

// TranscribeChunk transcribes an audio chunk and applies iterative agreement
func (s *StreamingWhisperEngine) TranscribeChunk(audioData []float32) (*TranscriptionResult, error) {
	if len(audioData) == 0 {
		return nil, fmt.Errorf("empty audio chunk")
	}

	// Create context for transcription
	context, err := s.model.NewContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create whisper context: %w", err)
	}

	// Set language if specified
	if lang := s.config.General.Language; lang != "" && lang != "auto" {
		if err := context.SetLanguage(lang); err != nil {
			return nil, fmt.Errorf("failed to set language: %w", err)
		}
	}

	// Process audio data
	if err := context.Process(audioData, nil, nil, nil); err != nil {
		return nil, fmt.Errorf("failed to process audio chunk: %w", err)
	}

	// Extract transcript
	var transcript strings.Builder
	for {
		segment, err := context.NextSegment()
		if err != nil {
			break
		}
		transcript.WriteString(segment.Text)
		transcript.WriteString(" ")
	}

	result := strings.TrimSpace(transcript.String())
	result = utils.SanitizeTranscript(result)

	// Apply iterative agreement logic
	transcriptionResult := s.processWithAgreement(result)

	// Send callback if provided
	if s.onPartialResult != nil {
		s.onPartialResult(transcriptionResult.Text, transcriptionResult.IsConfirmed)
	}

	return transcriptionResult, nil
}

// processWithAgreement implements iterative agreement for improved accuracy
func (s *StreamingWhisperEngine) processWithAgreement(newTranscript string) *TranscriptionResult {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	result := &TranscriptionResult{
		Text:      newTranscript,
		Timestamp: time.Now(),
	}

	// If transcript is empty, return immediately
	if newTranscript == "" {
		return result
	}

	// Normalize transcript for comparison
	normalized := strings.ToLower(strings.TrimSpace(newTranscript))

	// Check if this matches our last confirmed transcript
	if normalized == strings.ToLower(s.lastTranscript) {
		result.IsConfirmed = true
		result.Confidence = 1.0
		return result
	}

	// Update agreement cache
	s.agreementCache[normalized]++

	// Clean old entries to prevent memory bloat
	if len(s.agreementCache) > 100 {
		s.cleanAgreementCache()
	}

	// Check if we have enough agreements for this transcript
	if count, exists := s.agreementCache[normalized]; exists && count >= s.agreementThreshold {
		// We have consensus!
		result.IsConfirmed = true
		result.Confidence = float64(count) / float64(s.agreementThreshold+2) // Cap at reasonable confidence
		if result.Confidence > 1.0 {
			result.Confidence = 1.0
		}

		// Update last confirmed transcript
		s.lastTranscript = newTranscript

		// Clear cache for this transcript
		delete(s.agreementCache, normalized)

		return result
	}

	// Not enough agreements yet
	result.IsConfirmed = false
	result.Confidence = float64(s.agreementCache[normalized]) / float64(s.agreementThreshold)

	return result
}

// cleanAgreementCache removes old entries to prevent memory bloat
func (s *StreamingWhisperEngine) cleanAgreementCache() {
	// Simple strategy: remove entries with count = 1 (least likely to reach consensus)
	for key, count := range s.agreementCache {
		if count == 1 {
			delete(s.agreementCache, key)
			// Only remove a few at a time
			if len(s.agreementCache) <= 50 {
				break
			}
		}
	}
}

// TranscribeStream processes continuous audio stream with real-time transcription
func (s *StreamingWhisperEngine) TranscribeStream(ctx context.Context, audioStream <-chan []float32, resultStream chan<- *TranscriptionResult) error {
	defer close(resultStream)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case audioChunk, ok := <-audioStream:
			if !ok {
				// Stream closed
				return nil
			}

			// Transcribe chunk
			result, err := s.TranscribeChunk(audioChunk)
			if err != nil {
				// Log error but continue processing
				continue
			}

			// Send result if not empty
			if result.Text != "" {
				select {
				case resultStream <- result:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}

// Reset resets the streaming engine state
func (s *StreamingWhisperEngine) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Clear agreement cache
	for key := range s.agreementCache {
		delete(s.agreementCache, key)
	}

	// Reset last transcript
	s.lastTranscript = ""
}

// GetLastConfirmedTranscript returns the last confirmed transcript
func (s *StreamingWhisperEngine) GetLastConfirmedTranscript() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.lastTranscript
}

// SetAgreementThreshold sets the number of agreements required for consensus
func (s *StreamingWhisperEngine) SetAgreementThreshold(threshold int) {
	if threshold < 1 {
		threshold = 1
	}
	if threshold > 10 {
		threshold = 10
	}
	s.agreementThreshold = threshold
}

// sanitizeTranscript removed in favor of textutil.SanitizeTranscript
