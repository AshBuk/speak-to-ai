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
	// Build arecord command arguments
	args := a.buildCommandArgs()

	// Use BaseRecorder's ExecuteRecordingCommand for all process management
	return a.ExecuteRecordingCommand("arecord", args)
}

// StopRecording stops recording and returns the path to the recorded file
func (a *ArecordRecorder) StopRecording() (string, error) {
	// Close streaming pipe if enabled
	if a.streamingEnabled && a.pipeWriter != nil {
		defer a.pipeWriter.Close()
	}

	// Stop the recording process using BaseRecorder
	if err := a.StopProcess(); err != nil {
		return "", err
	}

	return a.outputFile, nil
}

// buildCommandArgs builds the command arguments for arecord
func (a *ArecordRecorder) buildCommandArgs() []string {
	args := []string{
		"-D", a.config.Audio.Device,
		"-f", a.config.Audio.Format,
		"-r", fmt.Sprintf("%d", a.config.Audio.SampleRate),
		"-c", fmt.Sprintf("%d", a.config.Audio.Channels),
	}

	if a.useBuffer || a.streamingEnabled {
		// Output to stdout for buffer/streaming mode
		args = append(args, "-t", "raw")
	} else {
		// Output to file
		args = append(args, a.outputFile)
	}

	return args
}
