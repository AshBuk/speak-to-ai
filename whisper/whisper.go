// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

// Package whisper provides speech-to-text functionality using Whisper models
package whisper

import (
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/whisper/interfaces"
	"github.com/AshBuk/speak-to-ai/whisper/manager"
	"github.com/AshBuk/speak-to-ai/whisper/processing"
)

// Re-export key interfaces and types for external use
type (
	ModelInfo              = interfaces.ModelInfo
	ProgressCallback       = interfaces.ProgressCallback
	ModelManager           = interfaces.ModelManager
	StreamingWhisperEngine = processing.StreamingWhisperEngine
	TranscriptionResult    = processing.TranscriptionResult
)

// NewModelManager creates a new model manager instance
func NewModelManager(config *config.Config) ModelManager {
	return manager.NewModelManager(config)
}

// NewStreamingWhisperEngine creates a new streaming whisper engine
func NewStreamingWhisperEngine(config *config.Config, modelPath string) (*StreamingWhisperEngine, error) {
	baseEngine, err := NewWhisperEngine(config, modelPath)
	if err != nil {
		return nil, err
	}

	return processing.NewStreamingWhisperEngine(baseEngine), nil
}
