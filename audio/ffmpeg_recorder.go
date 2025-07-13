package audio

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/config"
)

// FFmpegRecorder implements AudioRecorder using ffmpeg
type FFmpegRecorder struct {
	BaseRecorder
}

// NewFFmpegRecorder creates a new instance of FFmpegRecorder
func NewFFmpegRecorder(config *config.Config) *FFmpegRecorder {
	return &FFmpegRecorder{
		BaseRecorder: NewBaseRecorder(config),
	}
}

// StartRecording starts audio recording
func (f *FFmpegRecorder) StartRecording() error {
	// Build ffmpeg command arguments
	args := f.buildCommandArgs()

	// Use BaseRecorder's ExecuteRecordingCommand for all process management
	return f.ExecuteRecordingCommand("ffmpeg", args)
}

// StopRecording stops recording and returns the path to the recorded file
func (f *FFmpegRecorder) StopRecording() (string, error) {
	// Close streaming pipe if enabled
	if f.streamingEnabled && f.pipeWriter != nil {
		defer f.pipeWriter.Close()
	}

	// Stop the recording process using BaseRecorder
	if err := f.StopProcess(); err != nil {
		return "", err
	}

	return f.outputFile, nil
}

// buildCommandArgs builds the command arguments for ffmpeg
func (f *FFmpegRecorder) buildCommandArgs() []string {
	// Basic arguments
	args := []string{
		"-f", "alsa",
		"-i", f.config.Audio.Device,
		"-ar", fmt.Sprintf("%d", f.config.Audio.SampleRate),
		"-ac", fmt.Sprintf("%d", f.config.Audio.Channels),
	}

	// Add quality settings
	args = append(args, "-q:a", "0")

	// Configure output
	if f.streamingEnabled || f.useBuffer {
		// Output to stdout for streaming or buffer mode
		args = append(args, "-f", "wav", "-")
	} else {
		// Output to file
		args = append(args, f.outputFile)
	}

	return args
}
