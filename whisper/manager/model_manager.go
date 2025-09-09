// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/utils"
	"github.com/AshBuk/speak-to-ai/whisper/interfaces"
	"github.com/AshBuk/speak-to-ai/whisper/providers"
)

// ModelManager handles downloading and managing Whisper models
type ModelManager struct {
	config       *config.Config
	models       map[string]*interfaces.ModelInfo
	activeModel  string
	pathResolver interfaces.ModelPathResolver
	downloader   interfaces.ModelDownloader
}

// NewModelManager creates a new model manager instance
func NewModelManager(config *config.Config) *ModelManager {
	pathResolver := providers.NewModelPathResolver(config)
	downloader := providers.NewModelDownloader(pathResolver)

	return &ModelManager{
		config:       config,
		models:       make(map[string]*interfaces.ModelInfo),
		pathResolver: pathResolver,
		downloader:   downloader,
	}
}

// Initialize sets up the model manager
func (m *ModelManager) Initialize() error {
	// Load models from configuration
	m.loadModelsFromConfig()

	// Set active model
	if err := m.setActiveModel(); err != nil {
		return fmt.Errorf("failed to set active model: %w", err)
	}

	return nil
}

// GetModelPath returns the path to the requested model, downloading it if needed
func (m *ModelManager) GetModelPath() (string, error) {
	// If we have multiple models configured, use the active one
	if len(m.models) > 0 {
		return m.GetActiveModelPath(), nil
	}
	// Fallback to legacy behavior
	return m.GetModelPathWithProgress(nil)
}

// GetModelPathWithProgress returns the path to the requested model with progress callback
func (m *ModelManager) GetModelPathWithProgress(progressCallback interfaces.ProgressCallback) (string, error) {
	// If model path is specified directly, use it ONLY if it matches the requested model type
	if m.config.General.ModelPath != "" && utils.IsValidFile(m.config.General.ModelPath) {
		// For bundled models, check if the type matches what user requested
		if m.pathResolver.IsBundledModelPath(m.config.General.ModelPath) {
			bundledType := m.pathResolver.ExtractModelTypeFromPath(m.config.General.ModelPath)
			if bundledType == m.config.General.ModelType {
				// Bundled model matches requested type, use it
				return m.config.General.ModelPath, nil
			}
			// Bundled model doesn't match, continue to download logic below
		} else {
			// Non-bundled model path, use it as-is
			return m.config.General.ModelPath, nil
		}
	}

	// Use the standard model naming convention for downloading
	modelType := m.config.General.ModelType
	precision := m.config.General.ModelPrecision

	// Build the model path using path resolver
	modelPath := m.pathResolver.BuildModelPath(modelType, precision)

	// Check if model exists
	if utils.IsValidFile(modelPath) {
		return modelPath, nil
	}

	// If not, try to download it
	return m.downloader.DownloadModelWithProgress(modelType, precision, progressCallback)
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

// GetAvailableModels returns list of available models
func (m *ModelManager) GetAvailableModels() map[string]*interfaces.ModelInfo {
	result := make(map[string]*interfaces.ModelInfo)
	for k, v := range m.models {
		if v.Available {
			result[k] = v
		}
	}
	return result
}

// GetAllModels returns all configured models (available or not)
func (m *ModelManager) GetAllModels() map[string]*interfaces.ModelInfo {
	result := make(map[string]*interfaces.ModelInfo)
	for k, v := range m.models {
		result[k] = v
	}
	return result
}

// SwitchModel switches to a different model
func (m *ModelManager) SwitchModel(modelName string) error {
	model, exists := m.models[modelName]
	if !exists {
		return fmt.Errorf("model not found: %s", modelName)
	}

	if !model.Available {
		return fmt.Errorf("model not available: %s", modelName)
	}

	m.activeModel = modelName
	return nil
}

// GetActiveModel returns the currently active model name
func (m *ModelManager) GetActiveModel() string {
	return m.activeModel
}

// GetActiveModelPath returns the path of the currently active model
func (m *ModelManager) GetActiveModelPath() string {
	if m.activeModel == "" {
		return m.config.General.ModelPath // Fallback
	}

	if model, exists := m.models[m.activeModel]; exists {
		return model.Path
	}

	return m.config.General.ModelPath
}

// loadModelsFromConfig loads model information from configuration
func (m *ModelManager) loadModelsFromConfig() {
	// Load from models array if available
	if len(m.config.General.Models) > 0 {
		for _, modelPath := range m.config.General.Models {
			modelInfo := m.createModelInfo(modelPath)
			m.models[modelInfo.Name] = modelInfo
		}
	} else {
		// Fallback to single model path for backward compatibility
		modelInfo := m.createModelInfo(m.config.General.ModelPath)
		m.models[modelInfo.Name] = modelInfo
	}
}

// createModelInfo creates ModelInfo from a file path
func (m *ModelManager) createModelInfo(modelPath string) *interfaces.ModelInfo {
	// Resolve absolute path
	absPath := modelPath
	if !filepath.IsAbs(modelPath) {
		if cwd, err := os.Getwd(); err == nil {
			absPath = filepath.Join(cwd, modelPath)
		}
	}

	// Extract model name and type from filename
	basename := filepath.Base(absPath)
	name := basename
	if ext := filepath.Ext(basename); ext != "" {
		name = basename[:len(basename)-len(ext)]
	}

	// Use path resolver to determine model type from filename
	modelType := m.pathResolver.ExtractModelTypeFromPath(absPath)

	// Check if file exists and get size
	var size int64
	available := false
	if stat, err := os.Stat(absPath); err == nil {
		size = stat.Size()
		available = true
	}

	// Generate description
	description := m.generateModelDescription(modelType, size, available)

	return &interfaces.ModelInfo{
		Path:        absPath,
		Name:        name,
		Type:        modelType,
		Size:        size,
		Available:   available,
		Description: description,
	}
}

// generateModelDescription generates a human-readable description for a model
func (m *ModelManager) generateModelDescription(modelType string, size int64, available bool) string {
	if !available {
		return fmt.Sprintf("%s model (not available)", modelType)
	}

	sizeStr := formatFileSize(size)

	switch modelType {
	case "tiny":
		return fmt.Sprintf("Tiny model (%s) - Fastest, basic quality", sizeStr)
	case "base":
		return fmt.Sprintf("Base model (%s) - Balanced speed/accuracy", sizeStr)
	case "small":
		return fmt.Sprintf("Small model (%s) - Better quality, slower", sizeStr)
	case "medium":
		return fmt.Sprintf("Medium model (%s) - High quality, moderate speed", sizeStr)
	case "large":
		return fmt.Sprintf("Large model (%s) - Best quality, slowest", sizeStr)
	default:
		return fmt.Sprintf("%s model (%s)", modelType, sizeStr)
	}
}

// setActiveModel sets the active model from configuration
func (m *ModelManager) setActiveModel() error {
	// Use configured active model if available
	if m.config.General.ActiveModel != "" {
		for name, model := range m.models {
			if model.Path == m.config.General.ActiveModel && model.Available {
				m.activeModel = name
				return nil
			}
		}
	}

	// Fallback: find first available model
	for name, model := range m.models {
		if model.Available {
			m.activeModel = name
			return nil
		}
	}

	return fmt.Errorf("no available models found")
}

// formatFileSize formats file size in human readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
