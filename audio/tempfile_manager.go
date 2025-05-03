package audio

import (
	"log"
	"os"
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
