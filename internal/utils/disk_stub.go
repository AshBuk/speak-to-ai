//go:build !linux

package utils

// CheckDiskSpace is a no-op on non-Linux platforms
func CheckDiskSpace(path string) error {
	return nil
}
