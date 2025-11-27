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
// It relies on a path resolver to find the bundled model file
type ModelManager struct {
	config       *config.Config
	pathResolver interfaces.ModelPathResolver
}

// Create a new manager responsible for the bundled Whisper model
func NewModelManager(config *config.Config) *ModelManager {
	pathResolver := providers.NewModelPathResolver(config)

	return &ModelManager{
		config:       config,
		pathResolver: pathResolver,
	}
}

// Verify that the bundled `small-q5_1` model is present and valid
func (m *ModelManager) Initialize() error {
	modelPath := m.pathResolver.GetBundledModelPath()
	if _, err := os.Stat(modelPath); err != nil {
		return fmt.Errorf("bundled small-q5_1 model not found at %s: %w", modelPath, err)
	}
	return m.ValidateModel(modelPath)
}

// Return the path to the bundled `small-q5_1` model
// Do not fall back to any other path; the bundled model must exist
func (m *ModelManager) GetModelPath() (string, error) {
	modelPath := m.pathResolver.GetBundledModelPath()
	if !utils.IsValidFile(modelPath) {
		return "", fmt.Errorf("bundled small-q5_1 model not found at %s", modelPath)
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
