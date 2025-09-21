// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"os"
	"path/filepath"

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

// GetBundledModelPath returns the path to the bundled small-q5_1 model
func (r *ModelPathResolver) GetBundledModelPath() string {
	// Flatpak: use built-in model path
	if os.Getenv("FLATPAK_ID") != "" {
		return "/app/share/speak-to-ai/models/small-q5_1.bin"
	}

	// AppImage: use built-in model path
	if appDir := os.Getenv("APPDIR"); appDir != "" {
		return filepath.Join(appDir, "sources/language-models/small-q5_1.bin")
	}

	// Should not reach here - model must be bundled
	return ""
}
