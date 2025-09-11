// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package adapters

// HotkeyConfig defines the interface that a configuration must implement
// to be used with the hotkey system
type HotkeyConfig interface {
	GetStartRecordingHotkey() string
	GetProvider() string
	// GetActionHotkey returns configured hotkey string for a logical action name
	// Supported actions: "toggle_vad", "switch_model", "show_config", "reset_to_defaults"
	GetActionHotkey(action string) string
}

// ConfigAdapter adapts any config to HotkeyConfig
type ConfigAdapter struct {
	startRecording string
	provider       string
	// Hotkey for toggling VAD
	// toggleVAD       string
	switchModel     string
	showConfig      string
	resetToDefaults string
}

// NewConfigAdapter creates a new adapter from the given values
func NewConfigAdapter(startRecording string, provider string) *ConfigAdapter {
	return &ConfigAdapter{
		startRecording: startRecording,
		provider:       provider,
	}
}

// WithAdditionalHotkeys sets optional action hotkeys and returns the adapter (builder style)
func (c *ConfigAdapter) WithAdditionalHotkeys(toggleVAD, switchModel, showConfig, resetToDefaults string) *ConfigAdapter {
	// c.toggleVAD = toggleVAD
	c.switchModel = switchModel
	c.showConfig = showConfig
	c.resetToDefaults = resetToDefaults
	return c
}

// GetStartRecordingHotkey returns the start recording hotkey
func (c *ConfigAdapter) GetStartRecordingHotkey() string {
	return c.startRecording
}

// GetProvider returns provider override ("auto" | "dbus" | "evdev")
func (c *ConfigAdapter) GetProvider() string {
	if c.provider == "" {
		return "auto"
	}
	return c.provider
}

// GetActionHotkey implements action→hotkey string mapping
func (c *ConfigAdapter) GetActionHotkey(action string) string {
	switch action {
	// case "toggle_vad":
	// 	return c.toggleVAD
	case "switch_model":
		return c.switchModel
	case "show_config":
		return c.showConfig
	case "reset_to_defaults":
		return c.resetToDefaults
	default:
		return ""
	}
}
