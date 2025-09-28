// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

// Package config provides configuration management functionality with support for multiple
// configuration formats, validation, and security features.
//
// Subpackages:
//   - models:     Defines the core configuration data structures.
//   - loaders:    Handles loading and saving configuration from/to different formats (e.g., YAML).
//   - validators: Implements validation logic to ensure configuration integrity.
//   - security:   Provides security-related utilities like integrity checks and command validation.
package config

import (
	"github.com/AshBuk/speak-to-ai/config/loaders"
	"github.com/AshBuk/speak-to-ai/config/models"
	"github.com/AshBuk/speak-to-ai/config/security"
	"github.com/AshBuk/speak-to-ai/config/validators"
)

// Config is a type alias for the main configuration structure defined in the models package.
// This provides a convenient way to reference the configuration type without importing the models package directly.
type Config = models.Config

// Output mode constants, aliased from the models package for convenience.
const (
	OutputModeClipboard    = models.OutputModeClipboard
	OutputModeActiveWindow = models.OutputModeActiveWindow
)

// Load configuration from the specified file using the configured loader.
func LoadConfig(filename string) (*Config, error) {
	return loaders.LoadConfig(filename)
}

// Write the configuration to the specified file.
func SaveConfig(filename string, config *Config) error {
	return loaders.SaveConfig(filename, config)
}

// Apply the default values to a configuration object.
func SetDefaultConfig(config *Config) {
	loaders.SetDefaultConfig(config)
}

// Check the configuration for correctness and apply corrections if necessary.
func ValidateConfig(config *Config) error {
	return validators.ValidateConfig(config)
}

// Check if a command is permitted by the security policy.
func IsCommandAllowed(config *Config, command string) bool {
	return security.IsCommandAllowed(config, command)
}

// Remove potentially unsafe arguments from a command.
func SanitizeCommandArgs(args []string) []string {
	return security.SanitizeCommandArgs(args)
}

// Verify if the configuration file has been tampered with.
func VerifyConfigIntegrity(filename string, config *Config) error {
	return security.VerifyConfigIntegrity(filename, config)
}

// Calculate and update the integrity hash for the configuration file.
func UpdateConfigHash(filename string, config *Config) error {
	return security.UpdateConfigHash(filename, config)
}

// Compute the SHA-256 hash of a file.
func CalculateFileHash(filename string) (string, error) {
	return security.CalculateFileHash(filename)
}

// Enforce that a file does not exceed the configured size limit.
func EnforceFileSizeLimit(filename string, config *Config) error {
	return security.EnforceFileSizeLimit(filename, config)
}
