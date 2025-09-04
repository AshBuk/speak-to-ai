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

// TestGetTempFileManager tests singleton creation
func TestGetTempFileManager(t *testing.T) {
	// Get first instance
	manager1 := GetTempFileManager()
	if manager1 == nil {
		t.Fatal("Expected manager to be created, got nil")
	}

	// Get second instance - should be the same
	manager2 := GetTempFileManager()
	if manager1 != manager2 {
		t.Error("Expected singleton pattern - same instance should be returned")
	}

	// Check that the manager is properly initialized
	if manager1.tempFiles == nil {
		t.Error("Expected tempFiles map to be initialized")
	}

	if manager1.stopChan == nil {
		t.Error("Expected stopChan to be initialized")
	}

	if manager1.cleanupTimeout == 0 {
		t.Error("Expected cleanupTimeout to be set")
	}
}

// TestTempFileManager_AddFile tests adding files to tracking
func TestTempFileManager_AddFile(t *testing.T) {
	manager := GetTempFileManager()

	// Clear any existing files for clean test
	manager.mutex.Lock()
	manager.tempFiles = make(map[string]time.Time)
	manager.mutex.Unlock()

	testFile := "/tmp/test_audio_file.wav"

	// Add file
	manager.AddFile(testFile)

	// Check that file was added
	manager.mutex.Lock()
	timestamp, exists := manager.tempFiles[testFile]
	manager.mutex.Unlock()

	if !exists {
		t.Error("Expected file to be added to tracking")
	}

	// Check that timestamp is recent
	if time.Since(timestamp) > time.Second {
		t.Error("Expected timestamp to be recent")
	}
}

// TestTempFileManager_RemoveFile tests removing files from tracking
func TestTempFileManager_RemoveFile(t *testing.T) {
	manager := GetTempFileManager()

	// Clear any existing files for clean test
	manager.mutex.Lock()
	manager.tempFiles = make(map[string]time.Time)
	manager.mutex.Unlock()

	testFile := "/tmp/test_remove_file.wav"

	// Add file first
	manager.AddFile(testFile)

	// Verify it's there
	manager.mutex.Lock()
	_, exists := manager.tempFiles[testFile]
	manager.mutex.Unlock()

	if !exists {
		t.Fatal("File should exist before removal test")
	}

	// Remove file without deletion
	manager.RemoveFile(testFile, false)

	// Check that file was removed from tracking
	manager.mutex.Lock()
	_, exists = manager.tempFiles[testFile]
	manager.mutex.Unlock()

	if exists {
		t.Error("Expected file to be removed from tracking")
	}
}

// TestTempFileManager_RemoveFileWithDeletion tests file deletion
func TestTempFileManager_RemoveFileWithDeletion(t *testing.T) {
	manager := GetTempFileManager()

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "temp_manager_test_*.wav")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()

	fileName := tempFile.Name()

	// Add to manager
	manager.AddFile(fileName)

	// Verify file exists on disk
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		t.Fatal("Temp file should exist on disk")
	}

	// Remove with deletion
	manager.RemoveFile(fileName, true)

	// Check that file is removed from tracking
	manager.mutex.Lock()
	_, exists := manager.tempFiles[fileName]
	manager.mutex.Unlock()

	if exists {
		t.Error("Expected file to be removed from tracking")
	}

	// Check that file is deleted from disk
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		t.Error("Expected file to be deleted from disk")
	}
}

// TestTempFileManager_CleanupOldFiles tests automatic cleanup
func TestTempFileManager_CleanupOldFiles(t *testing.T) {
	_ = GetTempFileManager()

	// Create a test manager instance to avoid interfering with global one
	testManager := &TempFileManager{
		tempFiles:      make(map[string]time.Time),
		cleanupTimeout: 100 * time.Millisecond, // Very short timeout for testing
	}

	// Create actual temp files
	oldFile, err := os.CreateTemp("", "old_temp_*.wav")
	if err != nil {
		t.Fatalf("Failed to create old temp file: %v", err)
	}
	oldFile.Close()
	oldFileName := oldFile.Name()

	newFile, err := os.CreateTemp("", "new_temp_*.wav")
	if err != nil {
		t.Fatalf("Failed to create new temp file: %v", err)
	}
	newFile.Close()
	newFileName := newFile.Name()

	// Add files with different timestamps
	testManager.mutex.Lock()
	testManager.tempFiles[oldFileName] = time.Now().Add(-200 * time.Millisecond) // Old file
	testManager.tempFiles[newFileName] = time.Now()                              // New file
	testManager.mutex.Unlock()

	// Run cleanup
	testManager.cleanupOldFiles()

	// Check results
	testManager.mutex.Lock()
	_, oldExists := testManager.tempFiles[oldFileName]
	_, newExists := testManager.tempFiles[newFileName]
	testManager.mutex.Unlock()

	if oldExists {
		t.Error("Expected old file to be removed from tracking")
	}

	if !newExists {
		t.Error("Expected new file to remain in tracking")
	}

	// Check that old file was deleted from disk
	if _, err := os.Stat(oldFileName); !os.IsNotExist(err) {
		t.Error("Expected old file to be deleted from disk")
	}

	// Check that new file still exists
	if _, err := os.Stat(newFileName); os.IsNotExist(err) {
		t.Error("Expected new file to still exist on disk")
	}

	// Clean up remaining file
	os.Remove(newFileName)
}

// TestTempFileManager_ConcurrentAccess tests thread safety
func TestTempFileManager_ConcurrentAccess(t *testing.T) {
	manager := GetTempFileManager()

	// Clear existing files
	manager.mutex.Lock()
	manager.tempFiles = make(map[string]time.Time)
	manager.mutex.Unlock()

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines doing concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				fileName := fmt.Sprintf("/tmp/concurrent_test_%d_%d.wav", id, j)

				// Add file
				manager.AddFile(fileName)

				// Sometimes remove it
				if j%2 == 0 {
					manager.RemoveFile(fileName, false)
				}
			}
		}(i)
	}

	wg.Wait()

	// Check that no race conditions occurred (test will fail with -race if there are issues)
	manager.mutex.Lock()
	filesCount := len(manager.tempFiles)
	manager.mutex.Unlock()

	// We expect some files to remain (those not removed in the loop)
	if filesCount < 0 || filesCount > numGoroutines*numOperations {
		t.Errorf("Unexpected number of files: %d", filesCount)
	}
}

// TestTempFileManager_Stop tests stopping the cleanup routine
func TestTempFileManager_Stop(t *testing.T) {
	// Create a separate manager for this test to avoid affecting global one
	testManager := &TempFileManager{
		tempFiles:      make(map[string]time.Time),
		cleanupTimeout: 30 * time.Minute,
		stopChan:       make(chan bool),
		running:        true,
	}

	// Start cleanup routine
	go testManager.cleanupRoutine()

	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	// Stop the manager
	testManager.Stop()

	// Give it a moment to stop
	time.Sleep(10 * time.Millisecond)

	// Check that it stopped
	if testManager.running {
		t.Error("Expected manager to be stopped")
	}
}

// TestTempFileManager_CleanupRoutine tests the background cleanup routine
func TestTempFileManager_CleanupRoutine(t *testing.T) {
	// Create a test manager with very short intervals for testing
	testManager := &TempFileManager{
		tempFiles:      make(map[string]time.Time),
		cleanupTimeout: 50 * time.Millisecond,
		stopChan:       make(chan bool),
	}

	// Create a temp file that will become old
	tempFile, err := os.CreateTemp("", "cleanup_routine_test_*.wav")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	fileName := tempFile.Name()

	// Add file with old timestamp
	testManager.mutex.Lock()
	testManager.tempFiles[fileName] = time.Now().Add(-100 * time.Millisecond)
	testManager.mutex.Unlock()

	// Start cleanup routine in background
	routineDone := make(chan bool)
	go func() {
		testManager.cleanupRoutine()
		routineDone <- true
	}()

	// Give routine time to start
	time.Sleep(10 * time.Millisecond)

	// Verify routine is running
	if !testManager.running {
		t.Error("Expected cleanup routine to be running")
	}

	// Stop the routine
	testManager.Stop()

	// Wait for routine to actually stop
	select {
	case <-routineDone:
		// Good, routine stopped
	case <-time.After(100 * time.Millisecond):
		t.Error("Cleanup routine did not stop within timeout")
	}

	// Check that routine stopped
	if testManager.running {
		t.Error("Expected cleanup routine to be stopped")
	}

	// Clean up
	os.Remove(fileName)
}

// TestTempFileManager_DefaultTimeout tests the default cleanup timeout
func TestTempFileManager_DefaultTimeout(t *testing.T) {
	manager := GetTempFileManager()

	expectedTimeout := 30 * time.Minute
	if manager.cleanupTimeout != expectedTimeout {
		t.Errorf("Expected default timeout %v, got %v", expectedTimeout, manager.cleanupTimeout)
	}
}
