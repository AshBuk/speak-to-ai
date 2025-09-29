// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"os"
	"path/filepath"

	"github.com/AshBuk/speak-to-ai/config"
)

// Implements the logic for resolving the path to the bundled model
// based on the execution environment (e.g., Flatpak, AppImage)
type ModelPathResolver struct {
	config *config.Config
}

// Create a new resolver for locating the model file
func NewModelPathResolver(config *config.Config) *ModelPathResolver {
	return &ModelPathResolver{
		config: config,
	}
}

// Return the platform-specific path to the bundled `small-q5_1.bin` model.
// Check for Flatpak and AppImage environments to determine the correct relative path
func (r *ModelPathResolver) GetBundledModelPath() string {
	// In a Flatpak environment, the model is at a fixed path within the sandbox
	if os.Getenv("FLATPAK_ID") != "" {
		return "/app/share/speak-to-ai/models/small-q5_1.bin"
	}

	// In an AppImage environment, the model is located relative to the AppDir
	if appDir := os.Getenv("APPDIR"); appDir != "" {
		return filepath.Join(appDir, "sources/language-models/small-q5_1.bin")
	}

	// For other environments (like local development), the model is expected
	// to be in the standard source location
	return "sources/language-models/small-q5_1.bin"
}
