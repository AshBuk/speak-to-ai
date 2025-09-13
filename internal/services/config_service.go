// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/constants"
	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// ConfigService implements ConfigServiceInterface for configuration management only
type ConfigService struct {
	logger     logger.Logger
	config     *config.Config
	configFile string
	uiService  UIServiceInterface
}

// NewConfigService creates a new ConfigService instance
func NewConfigService(
	logger logger.Logger,
	config *config.Config,
	configFile string,
) *ConfigService {
	return &ConfigService{
		logger:     logger,
		config:     config,
		configFile: configFile,
		uiService:  nil, // Will be set later via dependency injection
	}
}

// SetUIService sets the UI service for notifications (dependency injection)
func (cs *ConfigService) SetUIService(uiService UIServiceInterface) {
	cs.uiService = uiService
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

// ResetToDefaults implements ConfigServiceInterface
func (cs *ConfigService) ResetToDefaults() error {
	cs.logger.Info("Resetting configuration to defaults...")

	// Reset existing config in-place to keep pointer stable across app
	config.SetDefaultConfig(cs.config)

	// Save to file immediately
	if err := cs.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save default config: %w", err)
	}

	cs.logger.Info("Configuration reset to defaults successfully")

	// Show success notification via UI service and refresh UI
	if cs.uiService != nil {
		cs.uiService.ShowNotification(constants.NotifyConfigReset, constants.NotifyConfigResetSuccess)
		cs.uiService.UpdateSettings(cs.config)
	}

	return nil
}

// GetConfig implements ConfigServiceInterface
func (cs *ConfigService) GetConfig() *config.Config {
	return cs.config
}

// TODO: Next feature - VAD implementation
// UpdateVADSensitivity implements ConfigServiceInterface
// func (cs *ConfigService) UpdateVADSensitivity(sensitivity string) error {
//	cs.logger.Info("Updating VAD sensitivity to: %s", sensitivity)
//
//	s := strings.ToLower(sensitivity)
//	switch s {
//	case "low", "medium", "high":
//	default:
//		return fmt.Errorf("invalid VAD sensitivity: %s", sensitivity)
//	}
//
//	if cs.config.Audio.VADSensitivity == s {
//		return nil
//	}
//
//	old := cs.config.Audio.VADSensitivity
//	cs.config.Audio.VADSensitivity = s
//
//	if err := cs.SaveConfig(); err != nil {
//		cs.config.Audio.VADSensitivity = old
//		return fmt.Errorf("failed to save config: %w", err)
//	}
//
//	return nil
// }

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

// UpdateOutputMode implements ConfigServiceInterface
func (cs *ConfigService) UpdateOutputMode(mode string) error {
	cs.logger.Info("Updating output mode to: %s", mode)

	if cs.config.Output.DefaultMode == mode {
		return nil
	}

	old := cs.config.Output.DefaultMode
	cs.config.Output.DefaultMode = mode

	if err := cs.SaveConfig(); err != nil {
		cs.config.Output.DefaultMode = old
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

// TODO: Next feature - VAD implementation
// ToggleVAD implements ConfigServiceInterface
// func (cs *ConfigService) ToggleVAD() error {
//	cs.logger.Info("Toggling VAD mode")
//
//	cs.config.Audio.EnableVAD = !cs.config.Audio.EnableVAD
//
//	return cs.SaveConfig()
// }

// UpdateRecordingMethod updates and persists the audio recording method
func (cs *ConfigService) UpdateRecordingMethod(method string) error {
	cs.logger.Info("Updating recording method to: %s", method)

	if method != "arecord" && method != "ffmpeg" {
		return fmt.Errorf("invalid recording method: %s", method)
	}

	if cs.config.Audio.RecordingMethod == method {
		return nil
	}

	old := cs.config.Audio.RecordingMethod
	cs.config.Audio.RecordingMethod = method

	if err := cs.SaveConfig(); err != nil {
		cs.config.Audio.RecordingMethod = old
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// UpdateHotkey updates a hotkey action string and persists config
func (cs *ConfigService) UpdateHotkey(action, combo string) error {
	if cs == nil || cs.config == nil {
		return fmt.Errorf("config service not available")
	}
	oldStart := cs.config.Hotkeys.StartRecording
	oldStop := cs.config.Hotkeys.StopRecording
	oldSwitch := cs.config.Hotkeys.SwitchModel
	oldShow := cs.config.Hotkeys.ShowConfig
	oldReset := cs.config.Hotkeys.ResetToDefaults

	switch action {
	case "start_recording", "stop_recording":
		cs.config.Hotkeys.StartRecording = combo
		cs.config.Hotkeys.StopRecording = combo
	case "switch_model":
		cs.config.Hotkeys.SwitchModel = combo
	case "show_config":
		cs.config.Hotkeys.ShowConfig = combo
	case "reset_to_defaults":
		cs.config.Hotkeys.ResetToDefaults = combo
	default:
		return fmt.Errorf("unknown hotkey action: %s", action)
	}

	if err := cs.SaveConfig(); err != nil {
		// rollback
		cs.config.Hotkeys.StartRecording = oldStart
		cs.config.Hotkeys.StopRecording = oldStop
		cs.config.Hotkeys.SwitchModel = oldSwitch
		cs.config.Hotkeys.ShowConfig = oldShow
		cs.config.Hotkeys.ResetToDefaults = oldReset
		return fmt.Errorf("failed to save hotkey: %w", err)
	}

	// Refresh UI display
	if cs.uiService != nil {
		cs.uiService.UpdateSettings(cs.config)
	}
	return nil
}

// Shutdown implements ConfigServiceInterface
func (cs *ConfigService) Shutdown() error {
	// Save final configuration state
	if err := cs.SaveConfig(); err != nil {
		cs.logger.Error("Error saving config during shutdown: %v", err)
		return err
	}

	cs.logger.Info("ConfigService shutdown complete")
	return nil
}
