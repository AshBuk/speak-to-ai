// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package audio

import (
	"io"
)

// AudioLevelCallback is called with audio level information
type AudioLevelCallback func(level float64)

// AudioRecorder interface for audio recording
type AudioRecorder interface {
	StartRecording() error
	StopRecording() (string, error)
	GetOutputFile() string
	CleanupFile() error
	UseStreaming() bool                                 // Indicates if the recorder supports streaming mode
	GetAudioStream() (io.Reader, error)                 // Returns the audio stream for streaming mode
	SetAudioLevelCallback(callback AudioLevelCallback)  // Sets callback for audio level monitoring
	GetAudioLevel() float64                             // Returns current audio level (0.0 to 1.0)
	StartStreamingRecording() (<-chan []float32, error) // Starts streaming recording, returns audio chunks channel
	StopStreamingRecording() error                      // Stops streaming recording
}
