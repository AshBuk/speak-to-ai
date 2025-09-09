// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package processing

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TempFileManager handles temporary audio files lifecycle
type TempFileManager struct {
	tempFiles      map[string]time.Time
	mutex          sync.Mutex
	cleanupTimeout time.Duration
	running        bool
	stopChan       chan bool
}

var (
	// Global singleton instance
	tempFileManager *TempFileManager
	managerOnce     sync.Once
)

// GetTempFileManager returns the global TempFileManager instance
func GetTempFileManager() *TempFileManager {
	managerOnce.Do(func() {
		tempFileManager = &TempFileManager{
			tempFiles:      make(map[string]time.Time),
			cleanupTimeout: 30 * time.Minute, // Default timeout
			stopChan:       make(chan bool),
		}
		go tempFileManager.cleanupRoutine()
	})
	return tempFileManager
}

// AddFile adds a file to the manager for tracking
func (t *TempFileManager) AddFile(path string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.tempFiles[path] = time.Now()
}

// RemoveFile removes a file from tracking and optionally deletes it
func (t *TempFileManager) RemoveFile(path string, shouldDelete bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	delete(t.tempFiles, path)

	if shouldDelete {
		if _, err := os.Stat(path); err == nil {
			if err := os.Remove(path); err != nil {
				log.Printf("Error removing temp file %s: %v", path, err)
			}
		}
	}
}

// cleanupRoutine periodically checks for old files to remove
func (t *TempFileManager) cleanupRoutine() {
	t.running = true
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.cleanupOldFiles()
		case <-t.stopChan:
			t.running = false
			return
		}
	}
}

// cleanupOldFiles removes files older than the timeout
func (t *TempFileManager) cleanupOldFiles() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	now := time.Now()
	for path, timestamp := range t.tempFiles {
		if now.Sub(timestamp) > t.cleanupTimeout {
			if _, err := os.Stat(path); err == nil {
				if err := os.Remove(path); err != nil {
					log.Printf("Error removing old temp file %s: %v", path, err)
				}
			}
			delete(t.tempFiles, path)
		}
	}
}

// Stop shuts down the cleanup routine
func (t *TempFileManager) Stop() {
	if t.running {
		t.stopChan <- true
		close(t.stopChan)
	}
}

// CreateTempWav creates a temp .wav file in provided base dir (or os.TempDir) and registers it
func (t *TempFileManager) CreateTempWav(baseDir string) (string, error) {
	// Determine directory
	dir := baseDir
	if dir == "" {
		dir = os.TempDir()
	}

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate unique name
	timestamp := time.Now().Format("20060102-150405")
	path := filepath.Join(dir, fmt.Sprintf("audio_%s.wav", timestamp))
	cleaned := filepath.Clean(path)
	// Ensure the path remains within the base directory to avoid traversal
	if !strings.HasPrefix(cleaned+string(os.PathSeparator), filepath.Clean(dir)+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe temp file path outside base dir")
	}

	// Pre-create file for downstream checks
	// #nosec G304 -- Safe: path is constructed, cleaned, and verified under controlled base directory
	f, err := os.OpenFile(cleaned, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	_ = f.Close()

	// Track file for cleanup
	t.AddFile(cleaned)
	return cleaned, nil
}
