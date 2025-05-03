package audio

import (
	"io"
)

// AudioRecorder interface for audio recording
type AudioRecorder interface {
	StartRecording() error
	StopRecording() (string, error)
	GetOutputFile() string
	CleanupFile() error
	UseStreaming() bool                 // Indicates if the recorder supports streaming mode
	GetAudioStream() (io.Reader, error) // Returns the audio stream for streaming mode
}
