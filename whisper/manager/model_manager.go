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

// ModelManager handles the fixed small-q5_1 Whisper model
type ModelManager struct {
	config       *config.Config
	pathResolver interfaces.ModelPathResolver
}

// NewModelManager creates a new model manager instance
func NewModelManager(config *config.Config) *ModelManager {
	pathResolver := providers.NewModelPathResolver(config)

	return &ModelManager{
		config:       config,
		pathResolver: pathResolver,
	}
}

// Initialize sets up the model manager for the small-q5_1 model
func (m *ModelManager) Initialize() error {
	// Validate the bundled small-q5_1 model exists
	modelPath := m.pathResolver.GetBundledModelPath()
	if _, err := os.Stat(modelPath); err != nil {
		return fmt.Errorf("bundled small-q5_1 model not found at %s: %w", modelPath, err)
	}
	// Model exists, validate it
	return m.ValidateModel(modelPath)
}

// GetModelPath returns the path to the bundled small-q5_1 model
func (m *ModelManager) GetModelPath() (string, error) {
	// Always use bundled model path - no config path fallback
	modelPath := m.pathResolver.GetBundledModelPath()

	// Verify the bundled model exists
	if !utils.IsValidFile(modelPath) {
		return "", fmt.Errorf("bundled small-q5_1 model not found at %s", modelPath)
	}

	return modelPath, nil
}

// ValidateModel checks if a model file is valid
func (m *ModelManager) ValidateModel(modelPath string) error {
	// Check file exists
	if !utils.IsValidFile(modelPath) {
		return fmt.Errorf("model file not found: %s", modelPath)
	}

	// Get file size
	size, err := utils.GetFileSize(modelPath)
	if err != nil {
		return fmt.Errorf("error checking model file: %w", err)
	}

	// Basic size check (models should be at least 10MB)
	if size < 10*1024*1024 {
		return fmt.Errorf("model file is too small (%d bytes), might be corrupted", size)
	}

	return nil
}
