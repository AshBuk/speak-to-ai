// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

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

// buildCommandArgs builds the command arguments for ffmpeg
func (f *FFmpegRecorder) buildCommandArgs() []string {
	// Basic arguments
	args := []string{
		"-y", // overwrite if temp file pre-created
		"-f", "alsa",
		"-i", f.config.Audio.Device,
		"-ar", fmt.Sprintf("%d", f.config.Audio.SampleRate),
		"-ac", fmt.Sprintf("%d", f.config.Audio.Channels),
		"-acodec", "pcm_s16le", // ensure 16-bit PCM output
	}

	// Add quality settings
	args = append(args, "-q:a", "0")

	// Configure output
	if f.streamingEnabled || f.useBuffer {
		// Output to stdout for streaming or buffer mode
		args = append(args, "-f", "wav", "-")
	} else {
		// Output to file
		args = append(args, "-f", "wav", f.outputFile)
	}

	return args
}
