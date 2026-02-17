// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package adapters

// Defines the contract a configuration must implement to be used by the hotkey system
type HotkeyConfig interface {
	GetStartRecordingHotkey() string
	GetProvider() string
	// Return the configured hotkey string for a given logical action name
	GetActionHotkey(action string) string
}

// Adapts a generic configuration to the HotkeyConfig interface
type ConfigAdapter struct {
	startRecording  string
	provider        string
	showConfig      string
	resetToDefaults string
}

// Create a new adapter from the given values
func NewConfigAdapter(startRecording string, provider string) *ConfigAdapter {
	return &ConfigAdapter{
		startRecording: startRecording,
		provider:       provider,
	}
}

// Set optional action hotkeys using a builder pattern
func (c *ConfigAdapter) WithAdditionalHotkeys(showConfig, resetToDefaults string) *ConfigAdapter {
	c.showConfig = showConfig
	c.resetToDefaults = resetToDefaults
	return c
}

// Return the start recording hotkey
func (c *ConfigAdapter) GetStartRecordingHotkey() string {
	return c.startRecording
}

// Return the provider override ("auto", "dbus", "evdev")
func (c *ConfigAdapter) GetProvider() string {
	if c.provider == "" {
		return "auto"
	}
	return c.provider
}

// Return the hotkey string for a given action name
func (c *ConfigAdapter) GetActionHotkey(action string) string {
	switch action {
	case "show_config":
		return c.showConfig
	case "reset_to_defaults":
		return c.resetToDefaults
	default:
		return ""
	}
}
