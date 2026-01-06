// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package interfaces

import (
	"context"

	"github.com/AshBuk/speak-to-ai/config"
)

// Defines the contract for managing the lifecycle of the fixed small-q5_1 Whisper model
type ModelManager interface {
	// Initialize and validate the bundled model
	Initialize() error
	// Return the absolute path to the validated small-q5_1 model file
	GetModelPath() (string, error)
	// Check if the model file at the given path is valid
	ValidateModel(modelPath string) error
}

// Defines the contract for resolving the path to the small-q5_1 model
type ModelPathResolver interface {
	// Return the platform-specific path to the model file (searches multiple locations)
	GetBundledModelPath() string
	// Return the user data directory path for downloading the model
	GetUserDataModelPath() string
}

// Defines the contract for downloading the whisper model
type ModelDownloader interface {
	// Download the model to the specified destination path
	Download(destPath string) error
	// Return the download URL
	GetModelURL() string
}

// Represents a whisper.Model type without a direct CGO dependency
type WhisperModel interface{}

// Defines the contract for a speech-to-text transcription engine
type WhisperEngine interface {
	// Transcribe an audio file and return the resulting text
	Transcribe(audioFile string) (string, error)
	// Add cancellation support for long running operations
	TranscribeWithContext(ctx context.Context, audioFile string) (string, error)
	// Close the engine and release any associated resources
	Close() error
	// Return the underlying model object
	GetModel() WhisperModel
	// Return the configuration used by the engine
	GetConfig() *config.Config
}
