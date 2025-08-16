//go:build !cgo || nocgo

package whisper

import (
	"errors"

	"github.com/AshBuk/speak-to-ai/config"
)

// WhisperEngine is a no-cgo stub that fails gracefully when CGO is disabled.
type WhisperEngine struct {
	config    *config.Config
	modelPath string
}

// NewWhisperEngine returns an error indicating that CGO is required.
func NewWhisperEngine(config *config.Config, modelPath string) (*WhisperEngine, error) {
	return nil, errors.New("whisper engine is unavailable: built without cgo")
}

// Close is a no-op in the stub implementation.
func (w *WhisperEngine) Close() error { return nil }

// Transcribe returns an error in the stub implementation.
func (w *WhisperEngine) Transcribe(audioFile string) (string, error) {
	return "", errors.New("transcription unavailable: built without cgo")
}
