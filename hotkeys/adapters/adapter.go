// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package adapters

// HotkeyConfig defines the interface that a configuration must implement
// to be used with the hotkey system
type HotkeyConfig interface {
	GetStartRecordingHotkey() string
	GetProvider() string
}

// ConfigAdapter adapts any config to HotkeyConfig
type ConfigAdapter struct {
	startRecording string
	provider       string
}

// NewConfigAdapter creates a new adapter from the given values
func NewConfigAdapter(startRecording string, provider string) *ConfigAdapter {
	return &ConfigAdapter{
		startRecording: startRecording,
		provider:       provider,
	}
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
