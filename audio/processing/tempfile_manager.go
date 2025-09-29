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

// Manages the lifecycle of temporary audio files
type TempFileManager struct {
	tempFiles      map[string]time.Time
	mutex          sync.Mutex
	cleanupTimeout time.Duration
	running        bool
	stopChan       chan bool
}

var (
	// global singleton instance
	tempFileManager *TempFileManager
	managerOnce     sync.Once
)

// Return the global TempFileManager instance, creating it if necessary
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

// Add a file to the manager for tracking and eventual cleanup
func (t *TempFileManager) AddFile(path string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.tempFiles[path] = time.Now()
}

// Remove a file from tracking and optionally delete it from disk
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

// Run a background routine to periodically remove old files
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

// Remove any tracked files that are older than the configured timeout
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

// Remove all tracked temporary files from disk immediately
func (t *TempFileManager) CleanupAll() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for path := range t.tempFiles {
		if _, err := os.Stat(path); err == nil {
			if err := os.Remove(path); err != nil {
				log.Printf("Error removing temp file %s: %v", path, err)
			}
		}
		delete(t.tempFiles, path)
	}
}

// Stop the background cleanup routine
func (t *TempFileManager) Stop() {
	if t.running {
		t.stopChan <- true
		close(t.stopChan)
	}
}

// Create a new temporary .wav file and register it for cleanup
func (t *TempFileManager) CreateTempWav(baseDir string) (string, error) {
	// Determine the directory for the temporary file
	dir := baseDir
	if dir == "" {
		dir = os.TempDir()
	}

	// Ensure the target directory exists
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate a unique filename based on the current timestamp
	timestamp := time.Now().Format("20060102-150405")
	path := filepath.Join(dir, fmt.Sprintf("audio_%s.wav", timestamp))
	cleaned := filepath.Clean(path)
	// Ensure the path remains within the base directory to prevent traversal attacks
	if !strings.HasPrefix(cleaned+string(os.PathSeparator), filepath.Clean(dir)+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe temp file path outside base dir")
	}

	// Pre-create the file to reserve the path
	// #nosec G304 -- Safe: path is constructed, cleaned, and verified under a controlled base directory.
	f, err := os.OpenFile(cleaned, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	_ = f.Close()

	// Track the file for automatic cleanup
	t.AddFile(cleaned)
	return cleaned, nil
}
