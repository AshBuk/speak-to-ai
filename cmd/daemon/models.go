package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ModelManager handles whisper models downloading and conversion
type ModelManager struct {
	config      *Config
	modelsDir   string
	quantizeBin string
	modelCache  map[string]*ModelInfo
	cacheMutex  sync.RWMutex
}

// ModelInfo contains information about a whisper model
type ModelInfo struct {
	Name      string
	Size      int64  // Size in bytes
	Format    string // "ggml" or "gguf"
	Quantized bool
	Precision string // "f16", "q4_0", etc.
	Path      string
	Checksum  string // SHA-256 hash
}

// Standard model URLs and their expected checksums
var modelURLs = map[string]string{
	"tiny":   "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin",
	"base":   "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin",
	"small":  "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin",
	"medium": "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-medium.bin",
	"large":  "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large.bin",
}

// Expected checksums for standard models
var modelChecksums = map[string]string{
	// These are placeholder checksums and should be replaced with actual values
	// The actual checksums would depend on the specific model files from HuggingFace
	"tiny":   "",
	"base":   "",
	"small":  "",
	"medium": "",
	"large":  "",
}

// NewModelManager creates a new instance of ModelManager
func NewModelManager(config *Config, modelsDir, quantizeBin string) *ModelManager {
	return &ModelManager{
		config:      config,
		modelsDir:   modelsDir,
		quantizeBin: quantizeBin,
		modelCache:  make(map[string]*ModelInfo),
	}
}

// PreloadModels preloads models specified in the configuration
func (m *ModelManager) PreloadModels() error {
	modelType := m.config.General.ModelType
	precision := m.config.General.ModelPrecision

	log.Printf("Preloading model: %s with precision %s", modelType, precision)

	_, err := m.GetModelPath(modelType, precision)
	if err != nil {
		return fmt.Errorf("failed to preload model %s-%s: %w", modelType, precision, err)
	}

	log.Printf("Model %s-%s preloaded successfully", modelType, precision)
	return nil
}

// GetAvailableModels returns a list of available models in the models directory
func (m *ModelManager) GetAvailableModels() ([]ModelInfo, error) {
	// Lock for cache reading
	m.cacheMutex.RLock()
	if len(m.modelCache) > 0 {
		// Convert cache map to slice
		models := make([]ModelInfo, 0, len(m.modelCache))
		for _, model := range m.modelCache {
			models = append(models, *model)
		}
		m.cacheMutex.RUnlock()
		return models, nil
	}
	m.cacheMutex.RUnlock()

	// No cache, need to read from disk and build cache
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	var models []ModelInfo

	// Ensure models directory exists
	if err := os.MkdirAll(m.modelsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create models directory: %w", err)
	}

	// Read the models directory
	files, err := os.ReadDir(m.modelsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read models directory: %w", err)
	}

	// Filter for model files
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasSuffix(name, ".bin") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			log.Printf("Failed to get info for %s: %v", name, err)
			continue
		}

		// Parse model information from filename
		baseName := strings.TrimSuffix(name, filepath.Ext(name))
		parts := strings.Split(baseName, "-")

		model := ModelInfo{
			Path: filepath.Join(m.modelsDir, name),
			Size: info.Size(),
		}

		// Calculate checksum
		checksum, err := calculateFileChecksum(model.Path)
		if err != nil {
			log.Printf("Warning: Failed to calculate checksum for %s: %v", model.Path, err)
		}
		model.Checksum = checksum

		if len(parts) >= 2 {
			// Format is typically like "ggml-base.bin" or "ggml-base-q4_0.bin"
			model.Name = parts[1]   // e.g., "base"
			model.Format = parts[0] // e.g., "ggml"

			// Check if it's quantized
			if len(parts) >= 3 {
				model.Quantized = true
				model.Precision = parts[2] // e.g., "q4_0"
			} else {
				model.Quantized = false
				model.Precision = "f16" // Default precision
			}
		} else {
			// If we can't parse the name format, just use the base name
			model.Name = baseName
			model.Format = "unknown"
			model.Quantized = false
		}

		models = append(models, model)

		// Add to cache
		cacheKey := fmt.Sprintf("%s-%s", model.Name, model.Precision)
		m.modelCache[cacheKey] = &model
	}

	return models, nil
}

// calculateFileChecksum calculates SHA-256 checksum of a file
func calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// DownloadModel downloads a model from a URL with checksum validation
func (m *ModelManager) DownloadModel(modelType string) (string, error) {
	url, exists := modelURLs[modelType]
	if !exists {
		return "", fmt.Errorf("unknown model type: %s", modelType)
	}

	// Ensure models directory exists
	if err := os.MkdirAll(m.modelsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create models directory: %w", err)
	}

	// Determine output path
	outputPath := filepath.Join(m.modelsDir, fmt.Sprintf("ggml-%s.bin", modelType))

	log.Printf("Downloading model %s from %s to %s...", modelType, url, outputPath)

	// Create output file
	out, err := os.Create(outputPath)
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

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Use a TeeReader to compute hash while downloading
	hash := sha256.New()
	teeReader := io.TeeReader(resp.Body, hash)

	// Write the body to file
	_, err = io.Copy(out, teeReader)
	if err != nil {
		return "", fmt.Errorf("failed to save model: %w", err)
	}

	// Get the computed checksum
	checksum := hex.EncodeToString(hash.Sum(nil))

	// Validate checksum if we have an expected value
	if expectedChecksum, ok := modelChecksums[modelType]; ok && expectedChecksum != "" {
		if checksum != expectedChecksum {
			// If checksum doesn't match, remove the file and return an error
			os.Remove(outputPath)
			return "", fmt.Errorf("checksum validation failed for %s", modelType)
		}
	}

	log.Printf("Model downloaded successfully to %s (checksum: %s)", outputPath, checksum)

	// Update model cache
	info := &ModelInfo{
		Name:      modelType,
		Path:      outputPath,
		Format:    "ggml",
		Precision: "f16",
		Quantized: false,
		Checksum:  checksum,
	}

	cacheKey := fmt.Sprintf("%s-f16", modelType)
	m.cacheMutex.Lock()
	m.modelCache[cacheKey] = info
	m.cacheMutex.Unlock()

	return outputPath, nil
}

// QuantizeModel quantizes a model to a lower precision
func (m *ModelManager) QuantizeModel(modelPath, precision string) (string, error) {
	if m.quantizeBin == "" {
		return "", fmt.Errorf("quantize binary not specified")
	}

	if _, err := os.Stat(m.quantizeBin); os.IsNotExist(err) {
		return "", fmt.Errorf("quantize binary not found at %s", m.quantizeBin)
	}

	// Get base model name without extension
	modelName := filepath.Base(modelPath)
	modelName = strings.TrimSuffix(modelName, filepath.Ext(modelName))

	// Output path with precision suffix
	outputPath := filepath.Join(m.modelsDir, fmt.Sprintf("%s-%s.bin", modelName, precision))

	// Check if the quantized model already exists
	if _, err := os.Stat(outputPath); err == nil {
		log.Printf("Quantized model %s already exists", outputPath)
		return outputPath, nil
	}

	log.Printf("Quantizing model %s to %s precision...", modelPath, precision)

	// Create command for quantization
	cmd := exec.Command(m.quantizeBin, modelPath, outputPath, precision)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Run the command
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to quantize model: %w\nstdout: %s\nstderr: %s",
			err, outBuf.String(), errBuf.String())
	}

	// Calculate checksum for the new quantized model
	checksum, err := calculateFileChecksum(outputPath)
	if err != nil {
		log.Printf("Warning: Failed to calculate checksum for %s: %v", outputPath, err)
	}

	// Update model cache
	parts := strings.Split(modelName, "-")
	if len(parts) >= 2 {
		baseModelName := parts[1] // e.g., "base" from "ggml-base"

		info := &ModelInfo{
			Name:      baseModelName,
			Path:      outputPath,
			Format:    parts[0], // e.g., "ggml"
			Precision: precision,
			Quantized: true,
			Checksum:  checksum,
		}

		cacheKey := fmt.Sprintf("%s-%s", baseModelName, precision)
		m.cacheMutex.Lock()
		m.modelCache[cacheKey] = info
		m.cacheMutex.Unlock()
	}

	log.Printf("Model quantized successfully to %s", outputPath)
	return outputPath, nil
}

// GetModelPath returns the path to the specified model, downloading it if necessary
func (m *ModelManager) GetModelPath(modelType, precision string) (string, error) {
	// If precision is "auto" or empty, use the default model
	if precision == "auto" || precision == "" {
		precision = "f16" // Default precision
	}

	// Check cache first
	cacheKey := fmt.Sprintf("%s-%s", modelType, precision)
	m.cacheMutex.RLock()
	if model, ok := m.modelCache[cacheKey]; ok {
		path := model.Path
		m.cacheMutex.RUnlock()

		// Verify the file exists
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		// File doesn't exist, continue to regular path
	} else {
		m.cacheMutex.RUnlock()
	}

	// Check if the model exists on disk
	models, err := m.GetAvailableModels()
	if err != nil {
		return "", err
	}

	// Look for the requested model
	for _, model := range models {
		if model.Name == modelType {
			if !strings.HasPrefix(model.Precision, precision) {
				// Found the model but with different precision
				continue
			}
			return model.Path, nil
		}
	}

	// Model not found, try to download it
	log.Printf("Model %s with precision %s not found, downloading...", modelType, precision)
	modelPath, err := m.DownloadModel(modelType)
	if err != nil {
		return "", err
	}

	// If we need to quantize
	if precision != "f16" {
		modelPath, err = m.QuantizeModel(modelPath, precision)
		if err != nil {
			return "", err
		}
	}

	return modelPath, nil
}
