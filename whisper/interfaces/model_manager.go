// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package interfaces

import (
	"github.com/AshBuk/speak-to-ai/config"
)

// ModelManager defines the interface for managing the fixed small-q5_1 Whisper model
type ModelManager interface {
	Initialize() error
	GetModelPath() (string, error)
	ValidateModel(modelPath string) error
}

// ModelPathResolver defines the interface for resolving the small-q5_1 model path
type ModelPathResolver interface {
	GetModelDir() string
	ShouldUseUserDirectory() bool
	IsBundledModelPath(modelPath string) bool
	GetUserModelsDirectory() string
	GetBundledModelPath() string
}

// WhisperModel represents a whisper model interface to avoid CGO dependency in interfaces
type WhisperModel interface{}

// WhisperEngine defines the interface for the base Whisper engine
type WhisperEngine interface {
	Transcribe(audioFile string) (string, error)
	Close() error
	GetModel() WhisperModel
	GetConfig() *config.Config // Strongly-typed config accessor
}
