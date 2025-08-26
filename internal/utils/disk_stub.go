//go:build !linux

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

// CheckDiskSpace is a no-op on non-Linux platforms
func CheckDiskSpace(path string) error {
	return nil
}
