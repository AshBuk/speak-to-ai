// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package recorders

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/audio/processing"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// Implements the AudioRecorder interface using the `ffmpeg` command-line tool
type FFmpegRecorder struct {
	BaseRecorder
}

// Create a new instance of the ffmpeg-based recorder
func NewFFmpegRecorder(config *config.Config, logger logger.Logger, tempManager *processing.TempFileManager) *FFmpegRecorder {
	return &FFmpegRecorder{
		BaseRecorder: NewBaseRecorder(config, logger, tempManager),
	}
}

// Start an audio recording using the `ffmpeg` command
func (f *FFmpegRecorder) StartRecording() error {
	// Build the command-line arguments for ffmpeg
	args := f.buildBaseCommandArgs()

	// Use the BaseRecorder to execute the command and manage the process
	return f.ExecuteRecordingCommand("ffmpeg", args)
}

// Build the base command-line arguments for an ffmpeg recording process
func (f *FFmpegRecorder) buildBaseCommandArgs() []string {
	// Start with base arguments for input, performance, and logging
	args := []string{
		"-y", // Overwrite output file if it exists
		"-nostdin",
		"-hide_banner",
		"-loglevel", "error",
		"-f", "pulse", // Use the PulseAudio input format for broad compatibility (including PipeWire)
		"-thread_queue_size", "256",
		"-probesize", "32", // Reduce probing time for faster startup
		"-analyzeduration", "0", // Skip analysis phase
	}

	// Specify the input device
	device := f.config.Audio.Device
	args = append(args, "-i", device)

	// Specify audio format settings
	args = append(args,
		"-ar", fmt.Sprintf("%d", f.config.Audio.SampleRate),
		"-ac", "1",
	)

	// Specify quality settings
	args = append(args, "-q:a", "0")

	// Configure the output format
	if f.useBuffer {
		// Output raw PCM to a WAV container in stdout for in-memory processing
		args = append(args, "-acodec", "pcm_s16le")
		args = append(args, "-f", "wav", "-")
	} else {
		// Output to a WAV file. The path will be added by ExecuteRecordingCommand
		args = append(args, "-acodec", "pcm_s16le")
		args = append(args, "-f", "wav")
	}

	return args
}
