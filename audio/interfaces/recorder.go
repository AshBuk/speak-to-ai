// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package interfaces

// AudioRecorder defines the contract for an audio recorder
type AudioRecorder interface {
	// Start the audio recording process
	StartRecording() error
	// Stop the audio recording and return the path to the output file
	StopRecording() (string, error)
	// Return the path to the recorded audio file
	GetOutputFile() string
	// Remove the temporary audio file
	CleanupFile() error
}
