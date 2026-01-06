// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// ModelDownloadURL is the URL to download the whisper small-q5_1 model
	ModelDownloadURL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small-q5_1.bin"
	// MinModelSize is the minimum expected model size (180 MB)
	MinModelSize = 180 * 1024 * 1024
)

// ModelDownloader handles downloading the whisper model from Hugging Face
type ModelDownloader struct {
	url string
}

// NewModelDownloader creates a new downloader with the default model URL
func NewModelDownloader() *ModelDownloader {
	return &ModelDownloader{url: ModelDownloadURL}
}

// Download downloads the model to the specified path
// Creates parent directories if they don't exist
func (d *ModelDownloader) Download(destPath string) error {
	// Create parent directories
	dir := filepath.Dir(destPath)
	// #nosec G301 -- Model directory needs to be readable by the application
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create model directory %s: %w", dir, err)
	}

	// Create temporary file for atomic download
	tmpPath := destPath + ".tmp"

	// Download to temporary file
	if err := d.downloadToFile(tmpPath); err != nil {
		_ = os.Remove(tmpPath) // Clean up on error
		return err
	}

	// Verify download size
	info, err := os.Stat(tmpPath)
	if err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to stat downloaded file: %w", err)
	}
	if info.Size() < MinModelSize {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("downloaded model is too small (%d bytes), expected at least %d bytes", info.Size(), MinModelSize)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, destPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to move model to final location: %w", err)
	}
	return nil
}

// downloadToFile downloads the model URL to the specified file
func (d *ModelDownloader) downloadToFile(path string) error {
	// Create HTTP request
	resp, err := http.Get(d.url) // #nosec G107 -- URL is a constant, not user input
	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download model: HTTP %d", resp.StatusCode)
	}

	// Create output file
	// #nosec G304 -- path is constructed internally, not from user input
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create model file: %w", err)
	}
	defer func() { _ = out.Close() }()

	// Copy response body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write model file: %w", err)
	}
	return nil
}

// GetModelURL returns the download URL
func (d *ModelDownloader) GetModelURL() string {
	return d.url
}
