// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package manager

import (
	"fmt"
	"os"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/utils"
	"github.com/AshBuk/speak-to-ai/whisper/interfaces"
	"github.com/AshBuk/speak-to-ai/whisper/providers"
)

// Implements the logic for managing the fixed `small-q5_1` Whisper model.
// It relies on a path resolver to find the model file and a downloader to fetch it
type ModelManager struct {
	config       *config.Config
	pathResolver *providers.ModelPathResolver
	downloader   interfaces.ModelDownloader
}

// Create a new manager responsible for the Whisper model
func NewModelManager(config *config.Config) *ModelManager {
	pathResolver := providers.NewModelPathResolver(config)
	downloader := providers.NewModelDownloader()

	return &ModelManager{
		config:       config,
		pathResolver: pathResolver,
		downloader:   downloader,
	}
}

// Verify that the `small-q5_1` model is present and valid.
// If not found, attempts to download it to user data directory
func (m *ModelManager) Initialize() error {
	modelPath := m.pathResolver.GetBundledModelPath()

	// Check if model exists
	if _, err := os.Stat(modelPath); err == nil {
		return m.ValidateModel(modelPath)
	}

	// Model not found - try to download to user data directory
	downloadPath := m.pathResolver.GetUserDataModelPath()
	if err := m.downloadModel(downloadPath); err != nil {
		return fmt.Errorf("model not found and download failed: %w", err)
	}

	return m.ValidateModel(downloadPath)
}

// Return the path to the `small-q5_1` model
func (m *ModelManager) GetModelPath() (string, error) {
	modelPath := m.pathResolver.GetBundledModelPath()
	if !utils.IsValidFile(modelPath) {
		return "", fmt.Errorf("small-q5_1 model not found at %s", modelPath)
	}
	return modelPath, nil
}

// Check if the model file at the given path is a valid model file
// Perform a basic sanity check based on file size
func (m *ModelManager) ValidateModel(modelPath string) error {
	if !utils.IsValidFile(modelPath) {
		return fmt.Errorf("model file not found: %s", modelPath)
	}
	size, err := utils.GetFileSize(modelPath)
	if err != nil {
		return fmt.Errorf("error checking model file: %w", err)
	}
	// Sanity check: the model should be at least 10MB
	if size < 10*1024*1024 {
		return fmt.Errorf("model file is too small (%d bytes), might be corrupted", size)
	}
	return nil
}

// downloadModel downloads the model to the specified path
func (m *ModelManager) downloadModel(destPath string) error {
	return m.downloader.Download(destPath)
}
