// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package processing

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// TestNewTempFileManager tests manager creation
func TestNewTempFileManager(t *testing.T) {
	manager := NewTempFileManager(30 * time.Minute)
	if manager == nil {
		t.Fatal("Expected manager to be created, got nil")
	}
	if manager.tempFiles == nil {
		t.Error("Expected tempFiles map to be initialized")
	}
	if manager.stopChan == nil {
		t.Error("Expected stopChan to be initialized")
	}
	if manager.cleanupTimeout != 30*time.Minute {
		t.Error("Expected cleanupTimeout to be set to 30 minutes")
	}
}

// TestTempFileManager_AddFile tests adding files to tracking
func TestTempFileManager_AddFile(t *testing.T) {
	manager := NewTempFileManager(30 * time.Minute)

	testPath := "/tmp/test_audio.wav"
	manager.AddFile(testPath)
	manager.mutex.Lock()
	_, exists := manager.tempFiles[testPath]
	manager.mutex.Unlock()

	if !exists {
		t.Errorf("Expected file %s to be tracked", testPath)
	}
}

// TestTempFileManager_RemoveFile tests file removal
func TestTempFileManager_RemoveFile(t *testing.T) {
	manager := NewTempFileManager(30 * time.Minute)

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_audio_*.wav")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	testPath := tmpFile.Name()
	tmpFile.Close()
	// Add file to tracking
	manager.AddFile(testPath)

	// Remove file (with deletion)
	manager.RemoveFile(testPath, true)
	// Verify file is no longer tracked
	manager.mutex.Lock()
	_, exists := manager.tempFiles[testPath]
	manager.mutex.Unlock()

	if exists {
		t.Error("Expected file to be removed from tracking")
	}
	// Verify file was deleted from disk
	if _, err := os.Stat(testPath); !os.IsNotExist(err) {
		t.Error("Expected file to be deleted from disk")
	}
}

// TestTempFileManager_CleanupOldFiles tests automatic cleanup of old files
func TestTempFileManager_CleanupOldFiles(t *testing.T) {
	// Use very short timeout for testing
	manager := NewTempFileManager(100 * time.Millisecond)

	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_audio_*.wav")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	testPath := tmpFile.Name()
	tmpFile.Close()
	// Add file with old timestamp
	manager.mutex.Lock()
	manager.tempFiles[testPath] = time.Now().Add(-200 * time.Millisecond)
	manager.mutex.Unlock()
	// Run cleanup
	manager.cleanupOldFiles()
	// Verify file was removed
	manager.mutex.Lock()
	_, exists := manager.tempFiles[testPath]
	manager.mutex.Unlock()

	if exists {
		t.Error("Expected old file to be removed")
	}
}

// TestTempFileManager_CreateTempWav tests temporary WAV file creation
func TestTempFileManager_CreateTempWav(t *testing.T) {
	manager := NewTempFileManager(30 * time.Minute)

	// Create temp file in system temp directory
	path, err := manager.CreateTempWav("")
	if err != nil {
		t.Fatalf("Failed to create temp wav: %v", err)
	}
	// Verify file exists
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		t.Errorf("Expected file to exist at %s", path)
	}
	// Verify file is tracked
	manager.mutex.Lock()
	_, exists := manager.tempFiles[path]
	manager.mutex.Unlock()

	if !exists {
		t.Error("Expected file to be tracked")
	}
	// Cleanup
	manager.RemoveFile(path, true)
}

// TestTempFileManager_CreateTempWav_CustomDir tests file creation in custom directory
func TestTempFileManager_CreateTempWav_CustomDir(t *testing.T) {
	manager := NewTempFileManager(30 * time.Minute)

	// Create custom directory
	customDir := "/tmp/test_audio_dir"
	defer os.RemoveAll(customDir)

	// Create temp file in custom directory
	path, err := manager.CreateTempWav(customDir)
	if err != nil {
		t.Fatalf("Failed to create temp wav in custom dir: %v", err)
	}
	// Verify file exists
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		t.Errorf("Expected file to exist at %s", path)
	}
	// Cleanup
	manager.RemoveFile(path, true)
}

// TestTempFileManager_CleanupAll tests cleanup of all tracked files
func TestTempFileManager_CleanupAll(t *testing.T) {
	manager := NewTempFileManager(30 * time.Minute)

	// Create multiple temp files
	var paths []string
	for i := 0; i < 3; i++ {
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("test_audio_%d_*.wav", i))
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		path := tmpFile.Name()
		tmpFile.Close()
		paths = append(paths, path)
		manager.AddFile(path)
	}

	// Cleanup all files
	manager.CleanupAll()

	// Verify all files are no longer tracked
	manager.mutex.Lock()
	trackedCount := len(manager.tempFiles)
	manager.mutex.Unlock()

	if trackedCount != 0 {
		t.Errorf("Expected 0 tracked files, got %d", trackedCount)
	}
	// Verify all files were deleted
	for _, path := range paths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Expected file %s to be deleted", path)
		}
	}
}

// TestTempFileManager_ConcurrentAccess tests thread-safe operations
func TestTempFileManager_ConcurrentAccess(t *testing.T) {
	manager := NewTempFileManager(30 * time.Minute)

	var wg sync.WaitGroup
	concurrency := 10

	// Concurrent AddFile operations
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			path := fmt.Sprintf("/tmp/test_audio_%d.wav", id)
			manager.AddFile(path)
		}(i)
	}

	wg.Wait()

	// Verify all files were added
	manager.mutex.Lock()
	count := len(manager.tempFiles)
	manager.mutex.Unlock()

	if count != concurrency {
		t.Errorf("Expected %d files, got %d", concurrency, count)
	}
}

// TestTempFileManager_StartStop tests background routine lifecycle
func TestTempFileManager_StartStop(t *testing.T) {
	manager := NewTempFileManager(5 * time.Minute)

	// Start the background routine
	manager.Start()
	if !manager.running {
		t.Error("Expected manager to be running after Start()")
	}
	// Stop the background routine
	manager.Stop()

	// Give it a moment to stop
	time.Sleep(100 * time.Millisecond)
	if manager.running {
		t.Error("Expected manager to be stopped after Stop()")
	}
}

// TestTempFileManager_StopIdempotent tests that Stop() can be called multiple times safely
func TestTempFileManager_StopIdempotent(t *testing.T) {
	manager := NewTempFileManager(5 * time.Minute)
	manager.Start()

	// Stop multiple times should not panic
	manager.Stop()

	// Second stop should be safe (no-op) - protected by stopClosed flag
	defer func() {
		if r := recover(); r != nil {
			t.Error("Stop() panicked on second call")
		}
	}()
	manager.Stop()
}
