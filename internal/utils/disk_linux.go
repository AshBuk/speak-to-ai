//go:build linux

package utils

import (
	"fmt"
	"path/filepath"
	"syscall"
)

// CheckDiskSpace ensures there's enough disk space available (Linux implementation)
func CheckDiskSpace(path string) error {
	dir := filepath.Dir(path)
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return err
	}

	available := stat.Bavail * uint64(stat.Bsize)
	const requiredSpace uint64 = 100 * 1024 * 1024
	if available < requiredSpace {
		return fmt.Errorf("insufficient disk space: %d bytes available, %d required", available, requiredSpace)
	}
	return nil
}
