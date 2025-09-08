// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package recorders

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
	// Build arecord command arguments (output file will be set in ExecuteRecordingCommand)
	// Determine format: use FLOAT_LE for streaming to match float32 pipeline
	formatArg := a.getArecordFormat()
	if a.streamingEnabled {
		formatArg = "FLOAT_LE"
	}

	baseArgs := []string{
		"-D", a.config.Audio.Device,
		"-f", formatArg,
		"-r", fmt.Sprintf("%d", a.config.Audio.SampleRate),
		"-c", "1",
	}

	if a.streamingEnabled {
		// For streaming we prefer FLOAT_LE to match float32 pipeline
		// Override format to FLOAT if needed
		baseArgs = append(baseArgs, "-t", "raw")
	} else if a.useBuffer {
		// Keep WAV header for buffer mode compatibility
		baseArgs = append(baseArgs, "-t", "wav")
	} else {
		// Output to file with WAV header - file path will be added in ExecuteRecordingCommand
		baseArgs = append(baseArgs, "-t", "wav")
	}

	// Use BaseRecorder's ExecuteRecordingCommand for all process management
	return a.ExecuteRecordingCommand("arecord", baseArgs)
}

// getArecordFormat converts format from ffmpeg style to arecord style
func (a *ArecordRecorder) getArecordFormat() string {
	format := a.config.Audio.Format
	switch format {
	case "s16le":
		return "S16_LE"
	case "s24le":
		return "S24_LE"
	case "s32le":
		return "S32_LE"
	default:
		return format
	}
}
