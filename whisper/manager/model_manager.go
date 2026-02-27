// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package manager

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/constants"
	"github.com/AshBuk/speak-to-ai/internal/utils"
	"github.com/AshBuk/speak-to-ai/whisper/providers"
)

// Manages Whisper model lifecycle: resolution, download, and validation
type ModelManager struct {
	config *config.Config
}

// Create a new manager responsible for the Whisper model
func NewModelManager(config *config.Config) *ModelManager {
	return &ModelManager{config: config}
}

// Initialize validates the configured model is present, downloading if needed
func (m *ModelManager) Initialize() error {
	_, err := m.resolveModel(m.configuredModelID())
	return err
}

// GetModelPath returns the absolute path to the configured model file
func (m *ModelManager) GetModelPath() (string, error) {
	return m.resolveModel(m.configuredModelID())
}

// SwitchModel resolves (and downloads if needed) a model by ID, returning its path
func (m *ModelManager) SwitchModel(modelID string) (string, error) {
	return m.resolveModel(modelID)
}

// ValidateModel checks if the model file at the given path is valid (basic size check)
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

// resolveModel finds or downloads a model by ID, returning its validated path
func (m *ModelManager) resolveModel(modelID string) (string, error) {
	def := constants.ModelByID(modelID)
	if def == nil {
		return "", fmt.Errorf("unknown model ID: %s", modelID)
	}

	resolver := providers.NewModelPathResolver(m.config, def.FileName)

	// Check bundled / user data / dev paths
	modelPath := resolver.GetBundledModelPath()
	if utils.IsValidFile(modelPath) {
		if err := m.ValidateModel(modelPath); err == nil {
			return modelPath, nil
		}
	}

	// Download to user data directory
	downloadPath := resolver.GetUserDataModelPath()
	dl := providers.NewModelDownloaderForURL(def.URL, def.MinSize)
	if err := dl.Download(downloadPath); err != nil {
		return "", fmt.Errorf("failed to download model %s: %w", modelID, err)
	}

	if err := m.ValidateModel(downloadPath); err != nil {
		return "", fmt.Errorf("downloaded model %s failed validation: %w", modelID, err)
	}
	return downloadPath, nil
}

// configuredModelID returns the model ID from config, falling back to default
func (m *ModelManager) configuredModelID() string {
	if id := m.config.General.WhisperModel; id != "" {
		return id
	}
	return constants.DefaultModelID
}
