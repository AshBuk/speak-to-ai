// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package interfaces

import (
	"github.com/AshBuk/speak-to-ai/config"
)

// ModelManager defines the interface for managing Whisper models
type ModelManager interface {
	Initialize() error
	GetModelPath() (string, error)
	GetModelPathWithProgress(progressCallback ProgressCallback) (string, error)
	ValidateModel(modelPath string) error
	GetAvailableModels() map[string]*ModelInfo
	GetAllModels() map[string]*ModelInfo
	SwitchModel(modelName string) error
	GetActiveModel() string
	GetActiveModelPath() string
}

// ProgressCallback is called during model download with progress information
type ProgressCallback func(downloaded, total int64, percentage float64)

// ModelInfo represents information about a Whisper model
type ModelInfo struct {
	Path        string
	Name        string
	Type        string // tiny, base, small, medium, large
	Size        int64  // File size in bytes
	Available   bool   // Whether the model file exists
	Description string
}

// ModelPathResolver defines the interface for resolving model paths
type ModelPathResolver interface {
	GetModelDir() string
	ShouldUseUserDirectory() bool
	IsBundledModelPath(modelPath string) bool
	ExtractModelTypeFromPath(modelPath string) string
	GetUserModelsDirectory() string
	BuildModelPath(modelType, precision string) string
	BuildModelFileName(modelType, precision string) string
}

// ModelDownloader defines the interface for downloading models
type ModelDownloader interface {
	DownloadModelWithProgress(modelType, precision string, progressCallback ProgressCallback) (string, error)
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
