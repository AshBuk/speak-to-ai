// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package interfaces

// AudioLevelCallback is called with audio level information
type AudioLevelCallback func(level float64)

// AudioRecorder interface for audio recording
type AudioRecorder interface {
	StartRecording() error
	StopRecording() (string, error)
	GetOutputFile() string
	CleanupFile() error
	SetAudioLevelCallback(callback AudioLevelCallback) // Sets callback for audio level monitoring
	GetAudioLevel() float64                            // Returns current audio level (0.0 to 1.0)
}
