// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
	"os"
	"os/exec"

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

	// Get editor from environment variable
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Fallback to xdg-open
		editor = "xdg-open"
		a.Logger.Debug("$EDITOR not set, using xdg-open as fallback")
	} else {
		a.Logger.Debug("Using editor from $EDITOR: %s", editor)
	}

	// Security: allowlist check on editor
	if !a.Config.IsCommandAllowed(editor) {
		return fmt.Errorf("command not allowed: %s", editor)
	}

	// Check if config file exists
	if _, err := os.Stat(a.ConfigFile); os.IsNotExist(err) {
		errMsg := fmt.Sprintf("Configuration file not found: %s", a.ConfigFile)
		a.Logger.Error(errMsg)
		if a.NotifyManager != nil {
			if err := a.NotifyManager.ShowNotification("Error", errMsg); err != nil {
				a.Logger.Warning("failed to show notification: %v", err)
			}
		}
		return fmt.Errorf("config file not found: %s", a.ConfigFile)
	}

	// Sanitize args (config file path)
	args := config.SanitizeCommandArgs([]string{a.ConfigFile})
	if len(args) != 1 {
		return fmt.Errorf("invalid config file path")
	}

	// Start editor in background
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
			if err := a.NotifyManager.ShowNotification("Error", errMsg); err != nil {
				a.Logger.Warning("failed to show notification: %v", err)
			}
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
		if err := a.NotifyManager.ShowNotification("Configuration", "Reloading configuration..."); err != nil {
			a.Logger.Warning("failed to show notification: %v", err)
		}
	}

	// Load new configuration
	newConfig, err := config.LoadConfig(a.ConfigFile)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to reload config: %v", err)
		a.Logger.Error(errMsg)
		if a.NotifyManager != nil {
			if err := a.NotifyManager.ShowNotification("Error", errMsg); err != nil {
				a.Logger.Warning("failed to show notification: %v", err)
			}
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
			if err := a.NotifyManager.ShowNotification("Error", errMsg); err != nil {
				a.Logger.Warning("failed to show notification: %v", err)
			}
		}
		return fmt.Errorf("failed to reinitialize components: %w", err)
	}

	// Success notification
	if a.NotifyManager != nil {
		if err := a.NotifyManager.ShowNotification("Configuration", "Configuration reloaded successfully!"); err != nil {
			a.Logger.Warning("failed to show notification: %v", err)
		}
	}

	a.Logger.Info("Configuration reloaded successfully")
	return nil
}
