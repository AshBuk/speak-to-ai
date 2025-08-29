// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
)

// VerifyConfigIntegrity checks if the config file has been tampered with
func VerifyConfigIntegrity(filename string, config *Config) error {
	// Skip check if not enabled
	if !config.Security.CheckIntegrity {
		return nil
	}

	// If no hash is set, just return
	if config.Security.ConfigHash == "" {
		return nil
	}

	// Calculate hash of the config file
	hash, err := calculateFileHash(filename)
	if err != nil {
		return fmt.Errorf("failed to calculate config file hash: %w", err)
	}

	// Compare hashes
	if hash != config.Security.ConfigHash {
		return fmt.Errorf("config file integrity check failed: hash mismatch")
	}

	return nil
}

// UpdateConfigHash updates the hash in the config
func UpdateConfigHash(filename string, config *Config) error {
	// Calculate hash
	hash, err := calculateFileHash(filename)
	if err != nil {
		return fmt.Errorf("failed to calculate config file hash: %w", err)
	}

	// Update hash in config
	config.Security.ConfigHash = hash
	return nil
}

// calculateFileHash calculates SHA-256 hash of a file
func calculateFileHash(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Warning: failed to close file %s: %v", filename, err)
		}
	}()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// EnforceFileSizeLimit checks if a file exceeds the maximum allowed size
func EnforceFileSizeLimit(filename string, config *Config) error {
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if info.Size() > config.Security.MaxTempFileSize {
		return fmt.Errorf("file size exceeds limit: %d bytes (limit: %d bytes)",
			info.Size(), config.Security.MaxTempFileSize)
	}

	return nil
}
