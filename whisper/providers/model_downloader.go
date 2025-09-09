// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/AshBuk/speak-to-ai/whisper/interfaces"
)

// ModelDownloader handles downloading models from remote sources
type ModelDownloader struct {
	pathResolver interfaces.ModelPathResolver
}

// NewModelDownloader creates a new model downloader
func NewModelDownloader(pathResolver interfaces.ModelPathResolver) *ModelDownloader {
	return &ModelDownloader{
		pathResolver: pathResolver,
	}
}

// DownloadModelWithProgress downloads a model from the server with progress reporting
func (d *ModelDownloader) DownloadModelWithProgress(modelType, precision string, progressCallback interfaces.ProgressCallback) (string, error) {
	// Create model directory if it doesn't exist
	modelDir := d.pathResolver.GetModelDir()
	if err := os.MkdirAll(modelDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create model directory: %w", err)
	}

	// Build model filename and path
	modelFile := d.pathResolver.BuildModelFileName(modelType, precision)
	modelPath := filepath.Join(modelDir, modelFile)

	// Build URL for downloading
	url := fmt.Sprintf("https://huggingface.co/ggerganov/whisper.cpp/resolve/main/%s", modelFile)

	// Create output file
	cleanPath := filepath.Clean(modelPath)
	out, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = out.Close() }()

	// Get the data
	parsed, perr := urlpkg.Parse(url)
	if perr != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("invalid download URL")
	}
	resp, err := http.Get(parsed.String())
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
	progressCallback interfaces.ProgressCallback
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
