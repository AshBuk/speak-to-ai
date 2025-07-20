package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// IsValidFile checks if a file exists and is accessible
func IsValidFile(path string) bool {
	// Check for path traversal attempts
	clean := filepath.Clean(path)
	if clean != path {
		return false
	}

	// Check file existence and access
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// CheckDiskSpace ensures there's enough disk space available
func CheckDiskSpace(path string) error {
	// Get directory stats
	dir := filepath.Dir(path)
	var stat syscall.Statfs_t
	err := syscall.Statfs(dir, &stat)
	if err != nil {
		return err
	}

	// Calculate available space
	available := stat.Bavail * uint64(stat.Bsize)

	// Require at least 100MB free
	const requiredSpace uint64 = 100 * 1024 * 1024
	if available < requiredSpace {
		return fmt.Errorf("insufficient disk space: %d bytes available, %d required", available, requiredSpace)
	}

	return nil
}
