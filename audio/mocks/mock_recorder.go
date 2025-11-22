// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package mocks

import (
	"errors"
	"time"

	"github.com/AshBuk/speak-to-ai/audio/interfaces"
)

// MockAudioRecorder implements AudioRecorder interface for testing
type MockAudioRecorder struct {
	isRecording         bool
	outputFile          string
	cleanupCalled       bool
	audioLevel          float64
	audioLevelCallback  interfaces.AudioLevelCallback
	startError          error
	stopError           error
	cleanupError        error
	recordingResult     string
	recordingDuration   time.Duration
	simulateAudioLevels bool
	audioLevelSequence  []float64
	audioLevelIndex     int
}

// NewMockAudioRecorder creates a new mock audio recorder
func NewMockAudioRecorder() *MockAudioRecorder {
	return &MockAudioRecorder{
		outputFile:         "/tmp/test_audio.wav",
		audioLevel:         0.0,
		recordingResult:    "Test transcription result",
		recordingDuration:  time.Second * 3,
		audioLevelSequence: []float64{0.1, 0.3, 0.5, 0.7, 0.4, 0.2},
	}
}

// StartRecording simulates starting audio recording
func (m *MockAudioRecorder) StartRecording() error {
	if m.startError != nil {
		return m.startError
	}
	if m.isRecording {
		return errors.New("recording already in progress")
	}
	m.isRecording = true
	// Simulate audio level monitoring if callback is set
	if m.simulateAudioLevels && m.audioLevelCallback != nil {
		go m.simulateAudioLevelUpdates()
	}
	return nil
}

// StopRecording simulates stopping audio recording
func (m *MockAudioRecorder) StopRecording() (string, error) {
	if m.stopError != nil {
		return "", m.stopError
	}
	if !m.isRecording {
		return "", errors.New("no recording in progress")
	}
	m.isRecording = false
	return m.recordingResult, nil
}

// GetOutputFile returns the mock output file path
func (m *MockAudioRecorder) GetOutputFile() string {
	return m.outputFile
}

// CleanupFile simulates cleaning up the audio file
func (m *MockAudioRecorder) CleanupFile() error {
	if m.cleanupError != nil {
		return m.cleanupError
	}
	m.cleanupCalled = true
	return nil
}

// SetAudioLevelCallback sets the callback for audio level monitoring
func (m *MockAudioRecorder) SetAudioLevelCallback(callback interfaces.AudioLevelCallback) {
	m.audioLevelCallback = callback
}

// GetAudioLevel returns the current mock audio level
func (m *MockAudioRecorder) GetAudioLevel() float64 {
	return m.audioLevel
}

// Test helper methods

// SetStartError configures the mock to return an error on StartRecording
func (m *MockAudioRecorder) SetStartError(err error) {
	m.startError = err
}

// SetStopError configures the mock to return an error on StopRecording
func (m *MockAudioRecorder) SetStopError(err error) {
	m.stopError = err
}

// SetCleanupError configures the mock to return an error on CleanupFile
func (m *MockAudioRecorder) SetCleanupError(err error) {
	m.cleanupError = err
}

// SetRecordingResult sets the mock transcription result
func (m *MockAudioRecorder) SetRecordingResult(result string) {
	m.recordingResult = result
}

// SetOutputFile sets the mock output file path
func (m *MockAudioRecorder) SetOutputFile(path string) {
	m.outputFile = path
}

// IsRecording returns whether recording is in progress
func (m *MockAudioRecorder) IsRecording() bool {
	return m.isRecording
}

// WasCleanupCalled returns whether CleanupFile was called
func (m *MockAudioRecorder) WasCleanupCalled() bool {
	return m.cleanupCalled
}

// EnableAudioLevelSimulation enables automatic audio level updates
func (m *MockAudioRecorder) EnableAudioLevelSimulation() {
	m.simulateAudioLevels = true
}

// SetAudioLevelSequence sets a sequence of audio levels to simulate
func (m *MockAudioRecorder) SetAudioLevelSequence(levels []float64) {
	m.audioLevelSequence = levels
	m.audioLevelIndex = 0
}

// simulateAudioLevelUpdates simulates audio level changes during recording
func (m *MockAudioRecorder) simulateAudioLevelUpdates() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for m.isRecording {
		<-ticker.C
		if m.audioLevelIndex < len(m.audioLevelSequence) {
			m.audioLevel = m.audioLevelSequence[m.audioLevelIndex]
			m.audioLevelIndex++
		} else {
			m.audioLevelIndex = 0
		}
		if m.audioLevelCallback != nil {
			m.audioLevelCallback(m.audioLevel)
		}
	}
}

// SetAudioLevel manually sets the audio level (for testing)
func (m *MockAudioRecorder) SetAudioLevel(level float64) {
	m.audioLevel = level
	if m.audioLevelCallback != nil {
		m.audioLevelCallback(level)
	}
}

// Reset clears all mock state
func (m *MockAudioRecorder) Reset() {
	m.isRecording = false
	m.outputFile = "/tmp/test_audio.wav"
	m.cleanupCalled = false
	m.audioLevel = 0.0
	m.audioLevelCallback = nil
	m.startError = nil
	m.stopError = nil
	m.cleanupError = nil
	m.recordingResult = "Test transcription result"
	m.recordingDuration = time.Second * 3
	m.simulateAudioLevels = false
	m.audioLevelSequence = []float64{0.1, 0.3, 0.5, 0.7, 0.4, 0.2}
	m.audioLevelIndex = 0
}
