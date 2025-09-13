// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

// Package config provides configuration management functionality with support for multiple
// configuration formats, validation, and security features.
//
// Subpackages:
//   - models: Configuration data structures and types
//   - loaders: Configuration loader implementations (YAML)
//   - validators: Configuration validation implementations
//   - security: Security utilities and integrity checks
package config

import (
	"github.com/AshBuk/speak-to-ai/config/loaders"
	"github.com/AshBuk/speak-to-ai/config/models"
	"github.com/AshBuk/speak-to-ai/config/security"
	"github.com/AshBuk/speak-to-ai/config/validators"
)

// Config type alias for convenience
type Config = models.Config

// Output mode constants
const (
	OutputModeClipboard    = models.OutputModeClipboard
	OutputModeActiveWindow = models.OutputModeActiveWindow
)

// LoadConfig loads configuration from file
func LoadConfig(filename string) (*Config, error) {
	return loaders.LoadConfig(filename)
}

// SaveConfig writes configuration to file
func SaveConfig(filename string, config *Config) error {
	return loaders.SaveConfig(filename, config)
}

// SetDefaultConfig sets default values
func SetDefaultConfig(config *Config) {
	loaders.SetDefaultConfig(config)
}

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	return validators.ValidateConfig(config)
}

// IsCommandAllowed checks if a command is allowed
func IsCommandAllowed(config *Config, command string) bool {
	return security.IsCommandAllowed(config, command)
}

// SanitizeCommandArgs removes potentially dangerous arguments
func SanitizeCommandArgs(args []string) []string {
	return security.SanitizeCommandArgs(args)
}

// VerifyConfigIntegrity checks if the config file has been tampered with
func VerifyConfigIntegrity(filename string, config *Config) error {
	return security.VerifyConfigIntegrity(filename, config)
}

// UpdateConfigHash updates the hash in the config
func UpdateConfigHash(filename string, config *Config) error {
	return security.UpdateConfigHash(filename, config)
}

// CalculateFileHash calculates SHA-256 hash of a file
func CalculateFileHash(filename string) (string, error) {
	return security.CalculateFileHash(filename)
}

// EnforceFileSizeLimit checks if a file exceeds the maximum allowed size
func EnforceFileSizeLimit(filename string, config *Config) error {
	return security.EnforceFileSizeLimit(filename, config)
}
