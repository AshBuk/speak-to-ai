// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package whisper

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/utils"
)

// ModelInfo represents information about a Whisper model
type ModelInfo struct {
	Path        string
	Name        string
	Type        string // tiny, base, small, medium, large
	Size        int64  // File size in bytes
	Available   bool   // Whether the model file exists
	Description string
}

// ModelManager handles downloading and managing Whisper models
type ModelManager struct {
	config      *config.Config
	models      map[string]*ModelInfo
	activeModel string
}

// ProgressCallback is called during model download with progress information
type ProgressCallback func(downloaded, total int64, percentage float64)

// NewModelManager creates a new model manager instance
func NewModelManager(config *config.Config) *ModelManager {
	return &ModelManager{
		config: config,
		models: make(map[string]*ModelInfo),
	}
}

// Initialize sets up the model manager
func (m *ModelManager) Initialize() error {
	// Load models from configuration
	if err := m.loadModelsFromConfig(); err != nil {
		return fmt.Errorf("failed to load models from config: %w", err)
	}

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
func (m *ModelManager) GetModelPathWithProgress(progressCallback ProgressCallback) (string, error) {
	// If model path is specified directly, use it
	if m.config.General.ModelPath != "" {
		// Check if it's a direct file path
		if utils.IsValidFile(m.config.General.ModelPath) {
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
	if utils.IsValidFile(modelPath) {
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
	defer func() { _ = out.Close() }()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download model: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

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
func (m *ModelManager) GetAvailableModels() map[string]*ModelInfo {
	result := make(map[string]*ModelInfo)
	for k, v := range m.models {
		if v.Available {
			result[k] = v
		}
	}
	return result
}

// GetAllModels returns all configured models (available or not)
func (m *ModelManager) GetAllModels() map[string]*ModelInfo {
	result := make(map[string]*ModelInfo)
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
func (m *ModelManager) loadModelsFromConfig() error {
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

	return nil
}

// createModelInfo creates ModelInfo from a file path
func (m *ModelManager) createModelInfo(modelPath string) *ModelInfo {
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

	// Determine model type from filename
	var modelType string
	switch {
	case contains(name, "tiny"):
		modelType = "tiny"
	case contains(name, "small"):
		modelType = "small"
	case contains(name, "medium"):
		modelType = "medium"
	case contains(name, "large"):
		modelType = "large"
	default:
		modelType = "base"
	}

	// Check if file exists and get size
	var size int64
	available := false
	if stat, err := os.Stat(absPath); err == nil {
		size = stat.Size()
		available = true
	}

	// Generate description
	description := m.generateModelDescription(modelType, size, available)

	return &ModelInfo{
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

// contains checks if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
