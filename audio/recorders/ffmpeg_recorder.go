// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package recorders

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
	// Build ffmpeg command arguments (output file will be set in ExecuteRecordingCommand)
	args := f.buildBaseCommandArgs()

	// Use BaseRecorder's ExecuteRecordingCommand for all process management
	return f.ExecuteRecordingCommand("ffmpeg", args)
}

// buildBaseCommandArgs builds the base command arguments for ffmpeg (without output file)
func (f *FFmpegRecorder) buildBaseCommandArgs() []string {
	// Basic arguments with improved PipeWire compatibility
	args := []string{
		"-y",          // overwrite if temp file pre-created
		"-f", "pulse", // Use pulse for PipeWire compatibility
		"-thread_queue_size", "64",
		"-probesize", "32", // Reduce probing time
		"-analyzeduration", "0", // Skip analysis phase
	}

	// Handle device specification for better PipeWire support
	device := f.config.Audio.Device
	if device == "default" {
		// Try to activate a specific source in PipeWire if default fails
		args = append(args, "-i", "default")
	} else {
		args = append(args, "-i", device)
	}

	// Audio format settings
	args = append(args,
		"-ar", fmt.Sprintf("%d", f.config.Audio.SampleRate),
		"-ac", "1",
	)

	// Add quality settings
	args = append(args, "-q:a", "0")

	// Configure output format
	if f.streamingEnabled {
		// Stream raw float32 for easier chunk decoding in streaming mode
		args = append(args, "-acodec", "pcm_f32le")
		args = append(args, "-f", "f32le", "-")
	} else if f.useBuffer {
		// Keep WAV header in buffer mode to avoid breaking audio level monitor
		args = append(args, "-acodec", "pcm_s16le")
		args = append(args, "-f", "wav", "-")
	} else {
		// Output to file - file path will be added by ExecuteRecordingCommand
		args = append(args, "-acodec", "pcm_s16le")
		args = append(args, "-f", "wav")
	}

	return args
}
