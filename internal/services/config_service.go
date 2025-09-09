// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"
	"strings"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/hotkeys/manager"
	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// ConfigService implements ConfigServiceInterface
type ConfigService struct {
	logger        logger.Logger
	config        *config.Config
	configFile    string
	hotkeyManager *manager.HotkeyManager
}

// NewConfigService creates a new ConfigService instance
func NewConfigService(
	logger logger.Logger,
	config *config.Config,
	configFile string,
	hotkeyManager *manager.HotkeyManager,
) *ConfigService {
	return &ConfigService{
		logger:        logger,
		config:        config,
		configFile:    configFile,
		hotkeyManager: hotkeyManager,
	}
}

// LoadConfig implements ConfigServiceInterface
func (cs *ConfigService) LoadConfig(configFile string) error {
	cs.logger.Info("Loading configuration from: %s", configFile)
	cs.configFile = configFile

	// Config is already loaded by factory, just update the file path
	return nil
}

// SaveConfig implements ConfigServiceInterface
func (cs *ConfigService) SaveConfig() error {
	cs.logger.Info("Saving configuration to: %s", cs.configFile)

	if cs.configFile == "" {
		return fmt.Errorf("no config file path set")
	}

	return config.SaveConfig(cs.configFile, cs.config)
}

// ReloadConfig implements ConfigServiceInterface
func (cs *ConfigService) ReloadConfig() error {
	cs.logger.Info("Reloading configuration...")

	if cs.configFile == "" {
		return fmt.Errorf("no config file path set")
	}

	newConfig, err := config.LoadConfig(cs.configFile)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	cs.config = newConfig
	return nil
}

// GetConfig implements ConfigServiceInterface
func (cs *ConfigService) GetConfig() interface{} {
	return cs.config
}

// UpdateVADSensitivity implements ConfigServiceInterface
func (cs *ConfigService) UpdateVADSensitivity(sensitivity string) error {
	cs.logger.Info("Updating VAD sensitivity to: %s", sensitivity)

	s := strings.ToLower(sensitivity)
	switch s {
	case "low", "medium", "high":
	default:
		return fmt.Errorf("invalid VAD sensitivity: %s", sensitivity)
	}

	if cs.config.Audio.VADSensitivity == s {
		return nil
	}

	old := cs.config.Audio.VADSensitivity
	cs.config.Audio.VADSensitivity = s

	if err := cs.SaveConfig(); err != nil {
		cs.config.Audio.VADSensitivity = old
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// UpdateLanguage implements ConfigServiceInterface
func (cs *ConfigService) UpdateLanguage(language string) error {
	cs.logger.Info("Updating language to: %s", language)

	if cs.config.General.Language == language {
		return nil
	}

	old := cs.config.General.Language
	cs.config.General.Language = language

	if err := cs.SaveConfig(); err != nil {
		cs.config.General.Language = old
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// UpdateModelType implements ConfigServiceInterface
func (cs *ConfigService) UpdateModelType(modelType string) error {
	cs.logger.Info("Updating model type to: %s", modelType)

	if cs.config.General.ModelType == modelType {
		return nil
	}

	old := cs.config.General.ModelType
	cs.config.General.ModelType = modelType

	if err := cs.SaveConfig(); err != nil {
		cs.config.General.ModelType = old
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// ToggleWorkflowNotifications implements ConfigServiceInterface
func (cs *ConfigService) ToggleWorkflowNotifications() error {
	cs.logger.Info("Toggling workflow notifications")

	cs.config.Notifications.EnableWorkflowNotifications = !cs.config.Notifications.EnableWorkflowNotifications

	return cs.SaveConfig()
}

// ToggleStreaming implements ConfigServiceInterface
func (cs *ConfigService) ToggleStreaming() error {
	cs.logger.Info("Toggling streaming mode")

	cs.config.Audio.EnableStreaming = !cs.config.Audio.EnableStreaming

	return cs.SaveConfig()
}

// ToggleVAD implements ConfigServiceInterface
func (cs *ConfigService) ToggleVAD() error {
	cs.logger.Info("Toggling VAD mode")

	cs.config.Audio.EnableVAD = !cs.config.Audio.EnableVAD

	return cs.SaveConfig()
}

// SetupHotkeyCallbacks configures hotkey callbacks with handler functions
func (cs *ConfigService) SetupHotkeyCallbacks(
	startRecording func() error,
	stopRecording func() error,
	toggleStreaming func() error,
	toggleVAD func() error,
	switchModel func() error,
	showConfig func() error,
	reloadConfig func() error,
) error {
	if cs.hotkeyManager == nil {
		return fmt.Errorf("hotkey manager not available")
	}

	cs.logger.Info("Setting up hotkey callbacks...")

	// Register the main recording callbacks
	cs.hotkeyManager.RegisterCallbacks(startRecording, stopRecording)

	// Register additional hotkey actions
	cs.hotkeyManager.RegisterHotkeyAction("toggle_streaming", toggleStreaming)
	cs.hotkeyManager.RegisterHotkeyAction("toggle_vad", toggleVAD)
	cs.hotkeyManager.RegisterHotkeyAction("switch_model", switchModel)
	cs.hotkeyManager.RegisterHotkeyAction("show_config", showConfig)
	cs.hotkeyManager.RegisterHotkeyAction("reload_config", reloadConfig)

	cs.logger.Info("Hotkey callbacks configured successfully")
	return nil
}

// RegisterHotkeys implements ConfigServiceInterface
func (cs *ConfigService) RegisterHotkeys() error {
	if cs.hotkeyManager == nil {
		return fmt.Errorf("hotkey manager not available")
	}

	cs.logger.Info("Registering hotkeys...")

	return cs.hotkeyManager.Start()
}

// UnregisterHotkeys implements ConfigServiceInterface
func (cs *ConfigService) UnregisterHotkeys() error {
	if cs.hotkeyManager == nil {
		return nil
	}

	cs.logger.Info("Unregistering hotkeys...")

	cs.hotkeyManager.Stop()
	return nil
}

// Shutdown implements ConfigServiceInterface
func (cs *ConfigService) Shutdown() error {
	var lastErr error

	// Unregister hotkeys
	if err := cs.UnregisterHotkeys(); err != nil {
		cs.logger.Error("Error unregistering hotkeys: %v", err)
		lastErr = err
	}

	// Save final configuration state
	if err := cs.SaveConfig(); err != nil {
		cs.logger.Error("Error saving config during shutdown: %v", err)
		lastErr = err
	}

	cs.logger.Info("ConfigService shutdown complete")
	return lastErr
}
