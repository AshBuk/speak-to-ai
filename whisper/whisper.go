// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

// Provides a high-level facade for interacting with the speech-to-text
// functionality. It abstracts the underlying implementation details of model management
// and transcription, exposing a simple entry point for external packages.
package whisper

import (
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/whisper/interfaces"
	"github.com/AshBuk/speak-to-ai/whisper/manager"
)

// Re-export key interfaces for convenience
type (
	ModelManager = interfaces.ModelManager
)

// Create a new manager for the bundled Whisper model
func NewModelManager(config *config.Config) ModelManager {
	return manager.NewModelManager(config)
}
