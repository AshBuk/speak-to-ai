package whisper

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
)

// ModelManager handles downloading and managing Whisper models
type ModelManager struct {
	config *config.Config
}

// ProgressCallback is called during model download with progress information
type ProgressCallback func(downloaded, total int64, percentage float64)

// NewModelManager creates a new model manager instance
func NewModelManager(config *config.Config) *ModelManager {
	return &ModelManager{
		config: config,
	}
}

// GetModelPath returns the path to the requested model, downloading it if needed
func (m *ModelManager) GetModelPath() (string, error) {
	return m.GetModelPathWithProgress(nil)
}

// GetModelPathWithProgress returns the path to the requested model with progress callback
func (m *ModelManager) GetModelPathWithProgress(progressCallback ProgressCallback) (string, error) {
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
	return m.downloadModelWithProgress(modelType, precision, progressCallback)
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

// downloadModelWithProgress downloads a model from the server with progress reporting
func (m *ModelManager) downloadModelWithProgress(modelType, precision string, progressCallback ProgressCallback) (string, error) {
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

	// Get content length for progress tracking
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		// Try to get from header
		if lengthStr := resp.Header.Get("Content-Length"); lengthStr != "" {
			if parsed, err := strconv.ParseInt(lengthStr, 10, 64); err == nil {
				contentLength = parsed
			}
		}
	}

	// Create progress reader if callback provided
	var reader io.Reader = resp.Body
	if progressCallback != nil && contentLength > 0 {
		reader = &progressReader{
			reader:           resp.Body,
			total:            contentLength,
			progressCallback: progressCallback,
		}
	}

	// Write the body to file
	_, err = io.Copy(out, reader)
	if err != nil {
		return "", fmt.Errorf("failed to save model: %w", err)
	}

	return modelPath, nil
}

// progressReader wraps an io.Reader to report download progress
type progressReader struct {
	reader           io.Reader
	total            int64
	downloaded       int64
	progressCallback ProgressCallback
	lastReportTime   time.Time
}

// Read implements io.Reader interface with progress reporting
func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.downloaded += int64(n)

	// Report progress every 100ms to avoid too frequent updates
	now := time.Now()
	if pr.progressCallback != nil && (now.Sub(pr.lastReportTime) > 100*time.Millisecond || err == io.EOF) {
		percentage := float64(pr.downloaded) / float64(pr.total) * 100
		pr.progressCallback(pr.downloaded, pr.total, percentage)
		pr.lastReportTime = now
	}

	return n, err
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
