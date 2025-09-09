// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AshBuk/speak-to-ai/config"
)

// handleShowConfig handles showing the configuration file
func (a *App) handleShowConfig() error {
	a.Logger.Info("Opening configuration file: %s", a.ConfigFile)

	// Show notification about config file location
	if a.NotifyManager != nil {
		if err := a.NotifyManager.ShowNotification("Configuration File", fmt.Sprintf("Opening: %s", a.ConfigFile)); err != nil {
			a.Logger.Warning("failed to show notification: %v", err)
		}
	}

	// Prefer xdg-open for better compatibility and security
	editor := "xdg-open"
	a.Logger.Debug("Using xdg-open to open config file")

	// Security: allowlist check on editor
	if !config.IsCommandAllowed(a.Config, editor) {
		return fmt.Errorf("command not allowed: %s", editor)
	}

	// Resolve candidate absolute path with fallbacks (XDG, Flatpak)
	absPath := a.ConfigFile
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		xdg := os.Getenv("XDG_CONFIG_HOME")
		if xdg == "" {
			if home, herr := os.UserHomeDir(); herr == nil {
				xdg = filepath.Join(home, ".config")
			}
		}
		candidate := filepath.Join(xdg, "speak-to-ai", "config.yaml")
		if _, cerr := os.Stat(candidate); cerr == nil {
			absPath = candidate
		} else {
			flatpakPath := filepath.Join(os.Getenv("HOME"), ".var", "app", "io.github.ashbuk.speak-to-ai", "config", "speak-to-ai", "config.yaml")
			if _, ferr := os.Stat(flatpakPath); ferr == nil {
				absPath = flatpakPath
			}
		}
	}
	if p, err := filepath.Abs(absPath); err == nil {
		absPath = p
	}
	args := config.SanitizeCommandArgs([]string{absPath})
	if len(args) != 1 {
		return fmt.Errorf("invalid config file path")
	}

	// Start editor in background
	// #nosec G204 -- Safe: editor is allowlisted constant (xdg-open) and args are sanitized.
	cmd := exec.Command(editor, args[0])

	// For GUI applications, detach from parent process
	if editor == "xdg-open" {
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
	}

	err := cmd.Start()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to open config file with %s: %v", editor, err)
		a.Logger.Error(errMsg)
		if a.NotifyManager != nil {
			a.notify("Error", errMsg)
		}
		return fmt.Errorf("failed to open config file: %w", err)
	}

	a.Logger.Info("Successfully opened config file with %s", editor)
	return nil
}

// handleReloadConfig handles reloading the configuration
func (a *App) handleReloadConfig() error {
	a.Logger.Info("Reloading configuration from: %s", a.ConfigFile)

	// Show notification about config reload
	if a.NotifyManager != nil {
		if err := a.NotifyManager.ShowNotification("Configuration", "Reloading..."); err != nil {
			a.Logger.Warning("failed to show notification: %v", err)
		}
	}

	// Load new configuration
	newConfig, err := config.LoadConfig(a.ConfigFile)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to reload config: %v", err)
		a.Logger.Error(errMsg)
		if a.NotifyManager != nil {
			a.notify("Error", errMsg)
		}
		return fmt.Errorf("failed to reload config: %w", err)
	}

	// Store old config for comparison
	oldConfig := a.Config
	a.Config = newConfig

	// Reinitialize components that depend on configuration
	err = a.reinitializeComponents(oldConfig)
	if err != nil {
		// Rollback to old config on failure
		a.Config = oldConfig
		errMsg := fmt.Sprintf("Failed to reinitialize components: %v", err)
		a.Logger.Error(errMsg)
		if a.NotifyManager != nil {
			a.notify("Error", errMsg)
		}
		return fmt.Errorf("failed to reinitialize components: %w", err)
	}

	// Success notification
	if a.NotifyManager != nil {
		if err := a.NotifyManager.ShowNotification("Configuration", "Reloaded successfully!"); err != nil {
			a.Logger.Warning("failed to show notification: %v", err)
		}
	}

	a.Logger.Info("Configuration reloaded successfully")
	return nil
}
