// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AshBuk/speak-to-ai/config"
)

// ModelPathResolver handles all path resolution logic for models
type ModelPathResolver struct {
	config *config.Config
}

// NewModelPathResolver creates a new model path resolver
func NewModelPathResolver(config *config.Config) *ModelPathResolver {
	return &ModelPathResolver{
		config: config,
	}
}

// GetModelDir returns the directory for storing models
func (r *ModelPathResolver) GetModelDir() string {
	// Check if we need to use user directory for downloaded models
	if r.ShouldUseUserDirectory() {
		return r.GetUserModelsDirectory()
	}

	dir := r.config.General.ModelPath
	if dir == "" {
		dir = "models"
	}

	// If the path looks like a file (has extension), return its directory
	if filepath.Ext(dir) != "" {
		return filepath.Dir(dir)
	}

	return dir
}

// ShouldUseUserDirectory determines if we should use user directory for models
func (r *ModelPathResolver) ShouldUseUserDirectory() bool {
	// Always use bundled model since we only support small-q5_1
	// No need to check compatibility since whisper_model is fixed
	return false
}

// IsBundledModelPath checks if path points to a bundled model
func (r *ModelPathResolver) IsBundledModelPath(modelPath string) bool {
	// Check for AppImage bundled paths
	if strings.Contains(modelPath, "sources/language-models/") {
		return true
	}
	// Check for Flatpak bundled paths
	if strings.Contains(modelPath, "/app/share/speak-to-ai/models/") {
		return true
	}
	return false
}

// ExtractModelTypeFromPath always returns "small" for the fixed small-q5_1 model
func (r *ModelPathResolver) ExtractModelTypeFromPath(modelPath string) string {
	// We only support small-q5_1 model in this simple implementation
	return "small"
}

// GetUserModelsDirectory returns the user's models directory
func (r *ModelPathResolver) GetUserModelsDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return "models"
	}
	return filepath.Join(homeDir, ".config", "speak-to-ai", "language-models")
}

// BuildModelPath constructs the full path for a model file
func (r *ModelPathResolver) BuildModelPath(modelType, precision string) string {
	modelFile := r.BuildModelFileName(modelType, precision)
	return filepath.Join(r.GetModelDir(), modelFile)
}

// BuildModelFileName constructs the filename for a model
func (r *ModelPathResolver) BuildModelFileName(modelType, precision string) string {
	return fmt.Sprintf("ggml-model-%s.%s.bin", modelType, precision)
}

// GetBundledModelPath returns the path to the bundled small-q5_1 model
func (r *ModelPathResolver) GetBundledModelPath() string {
	// Check if running in Flatpak and use the built-in model path
	if os.Getenv("FLATPAK_ID") != "" {
		flatpakModelPath := "/app/share/speak-to-ai/models/small-q5_1.bin"
		if _, err := os.Stat(flatpakModelPath); err == nil {
			return flatpakModelPath
		}
	}

	// Use the configured model path if it points to small-q5_1
	if r.config.General.ModelPath != "" && strings.Contains(r.config.General.ModelPath, "small-q5_1") {
		return r.config.General.ModelPath
	}

	// Build the standard path for small-q5_1 model
	return filepath.Join(r.GetModelDir(), "ggml-model-small.q5_1.bin")
}
