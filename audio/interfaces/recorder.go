// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package interfaces

// Defines a callback function for receiving audio level information
type AudioLevelCallback func(level float64)

// Defines the contract for an audio recorder
type AudioRecorder interface {
	// Start the audio recording process
	StartRecording() error
	// Stop the audio recording and return the path to the output file
	StopRecording() (string, error)
	// Return the path to the recorded audio file
	GetOutputFile() string
	// Remove the temporary audio file
	CleanupFile() error
	// Set a callback function for audio level monitoring
	SetAudioLevelCallback(callback AudioLevelCallback)
	// Return the current audio level, typically normalized from 0.0 to 1.0
	GetAudioLevel() float64
}
