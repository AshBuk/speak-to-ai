// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/AshBuk/dabri/v2/config"
)

const (
	// DefaultLockFileName is the default name for the lock file
	DefaultLockFileName = "dabri.lock"
)

// LockFile represents a file-based application lock
type LockFile struct {
	path string
	file *os.File
}

// NewLockFile creates a new lock file instance
func NewLockFile(path string) *LockFile {
	return &LockFile{
		path: path,
	}
}

// GetDefaultLockPath returns the default lock file path
func GetDefaultLockPath() string {
	// Try user runtime directory first (preferred for XDG systems)
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		return filepath.Join(runtimeDir, DefaultLockFileName)
	}

	// Fall back to user config directory
	if configDir, err := config.EnsureConfigDir(); err == nil {
		return filepath.Join(configDir, DefaultLockFileName)
	}

	// Last resort: temp directory
	return filepath.Join(os.TempDir(), DefaultLockFileName)
}

// TryLock attempts to acquire the lock
func (lf *LockFile) TryLock() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(lf.path), 0700); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Try to open/create the lock file; truncate only after the lock is held,
	// so a second instance does not wipe the running instance's PID
	file, err := os.OpenFile(lf.path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}
	// Try to acquire exclusive lock
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		if err == syscall.EWOULDBLOCK {
			return fmt.Errorf("another instance of dabri is already running")
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	// Write current PID to lock file, replacing a stale one
	if err := file.Truncate(0); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to truncate lock file: %w", err)
	}
	if _, err := file.WriteString(strconv.Itoa(os.Getpid())); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to write PID to lock file: %w", err)
	}

	lf.file = file
	return nil
}

// Unlock releases the lock
func (lf *LockFile) Unlock() error {
	if lf.file == nil {
		return nil
	}
	// Release the lock
	_ = syscall.Flock(int(lf.file.Fd()), syscall.LOCK_UN)
	// Close the file
	if err := lf.file.Close(); err != nil {
		return fmt.Errorf("failed to close lock file: %w", err)
	}

	lf.file = nil
	// Remove the lock file
	if err := os.Remove(lf.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	return nil
}
