package whisper

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/AshBuk/speak-to-ai/config"
)

// ModelManager handles downloading and managing Whisper models
type ModelManager struct {
	config *config.Config
}

// NewModelManager creates a new model manager instance
func NewModelManager(config *config.Config) *ModelManager {
	return &ModelManager{
		config: config,
	}
}

// GetModelPath returns the path to the requested model, downloading it if needed
func (m *ModelManager) GetModelPath() (string, error) {
	// If model path is specified directly, use it
	if m.config.General.ModelPath != "" {
		// Check if it's a direct file path
		if isValidFile(m.config.General.ModelPath) {
			return m.config.General.ModelPath, nil
		}
	}

	// Otherwise, use the standard model naming convention
	modelType := m.config.General.ModelType
	precision := m.config.General.ModelPrecision

	// Build the model filename
	modelFile := fmt.Sprintf("ggml-model-%s.%s.bin", modelType, precision)
	modelPath := filepath.Join(m.getModelDir(), modelFile)

	// Check if model exists
	if isValidFile(modelPath) {
		return modelPath, nil
	}

	// If not, try to download it
	return m.downloadModel(modelType, precision)
}

// getModelDir returns the directory for storing models
func (m *ModelManager) getModelDir() string {
	dir := m.config.General.ModelPath
	if dir == "" {
		dir = "models"
	}

	// If the path looks like a file (has extension), return its directory
	if filepath.Ext(dir) != "" {
		return filepath.Dir(dir)
	}

	return dir
}

// downloadModel downloads a model from the server
func (m *ModelManager) downloadModel(modelType, precision string) (string, error) {
	// Create model directory if it doesn't exist
	modelDir := m.getModelDir()
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create model directory: %w", err)
	}

	// Build model filename
	modelFile := fmt.Sprintf("ggml-model-%s.%s.bin", modelType, precision)
	modelPath := filepath.Join(modelDir, modelFile)

	// Build URL for downloading
	url := fmt.Sprintf("https://huggingface.co/ggerganov/whisper.cpp/resolve/main/%s", modelFile)

	// Create output file
	out, err := os.Create(modelPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download model: %w", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save model: %w", err)
	}

	return modelPath, nil
}

// ValidateModel checks if a model file is valid
func (m *ModelManager) ValidateModel(modelPath string) error {
	// Check file exists
	if !isValidFile(modelPath) {
		return fmt.Errorf("model file not found: %s", modelPath)
	}

	// Get file size
	size, err := getFileSize(modelPath)
	if err != nil {
		return fmt.Errorf("error checking model file: %w", err)
	}

	// Basic size check (models should be at least 10MB)
	if size < 10*1024*1024 {
		return fmt.Errorf("model file is too small (%d bytes), might be corrupted", size)
	}

	return nil
}
