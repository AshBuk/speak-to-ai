package audio

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/config"
)

// ArecordRecorder implements AudioRecorder using arecord
type ArecordRecorder struct {
	BaseRecorder
}

// NewArecordRecorder creates a new instance of ArecordRecorder
func NewArecordRecorder(config *config.Config) *ArecordRecorder {
	return &ArecordRecorder{
		BaseRecorder: NewBaseRecorder(config),
	}
}

// StartRecording starts audio recording
func (a *ArecordRecorder) StartRecording() error {
	// Prepare arecord command
	args := []string{
		"-D", a.config.Audio.Device,
		"-f", a.config.Audio.Format,
		"-r", fmt.Sprintf("%d", a.config.Audio.SampleRate),
		"-c", fmt.Sprintf("%d", a.config.Audio.Channels),
	}

	if a.useBuffer {
		// Add args to output to stdout
		args = append(args, "-t", "raw")
	} else {
		// Add output file
		args = append(args, a.outputFile)
	}

	// Start command using template method
	return a.StartProcessTemplate("arecord", args, true)
}

// StopRecording stops recording and returns the path to the recorded file
func (a *ArecordRecorder) StopRecording() (string, error) {
	if err := a.StopProcess(); err != nil {
		return "", err
	}

	// Return the file path
	return a.outputFile, nil
}
