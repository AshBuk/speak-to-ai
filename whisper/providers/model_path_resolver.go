// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"os"
	"path/filepath"

	"github.com/AshBuk/speak-to-ai/config"
)

const (
	// ModelFileName is the standard name for the whisper model file
	ModelFileName = "small-q5_1.bin"
	// AppDataDirName is the application data directory name
	AppDataDirName = "speak-to-ai"
	// ModelsDirName is the subdirectory for model files
	ModelsDirName = "models"
)

// Implements the logic for resolving the path to the bundled model
// based on the execution environment (e.g., AppImage, system package, dev)
type ModelPathResolver struct {
	config *config.Config
}

// Create a new resolver for locating the model file
func NewModelPathResolver(config *config.Config) *ModelPathResolver {
	return &ModelPathResolver{
		config: config,
	}
}

// GetBundledModelPath returns the path to the model file.
// Search order:
//  1. AppImage bundled path ($APPDIR/sources/language-models/)
//  2. User data directory (~/.local/share/speak-to-ai/models/)
//  3. Development path (sources/language-models/)
func (r *ModelPathResolver) GetBundledModelPath() string {
	// 1. AppImage environment
	if appDir := os.Getenv("APPDIR"); appDir != "" {
		path := filepath.Join(appDir, "sources/language-models", ModelFileName)
		if fileExists(path) {
			return path
		}
	}
	// 2. User data directory (XDG_DATA_HOME or ~/.local/share)
	userDataPath := r.getUserDataModelPath()
	if fileExists(userDataPath) {
		return userDataPath
	}
	// 3. Development path
	devPath := filepath.Join("sources/language-models", ModelFileName)
	if fileExists(devPath) {
		return devPath
	}
	// Return user data path as default (for download location)
	return userDataPath
}

// GetUserDataModelPath returns the path where model should be downloaded
// This is XDG_DATA_HOME/speak-to-ai/models/small-q5_1.bin
func (r *ModelPathResolver) GetUserDataModelPath() string {
	return r.getUserDataModelPath()
}

// getUserDataModelPath returns the XDG-compliant user data path for the model
func (r *ModelPathResolver) getUserDataModelPath() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home is unavailable
			return filepath.Join(ModelsDirName, ModelFileName)
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, AppDataDirName, ModelsDirName, ModelFileName)
}

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
