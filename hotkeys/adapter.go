package hotkeys

// HotkeyConfig defines the interface that a configuration must implement
// to be used with the hotkey system
type HotkeyConfig interface {
	GetStartRecordingHotkey() string
}

// ConfigAdapter adapts any config to HotkeyConfig
type ConfigAdapter struct {
	startRecording string
}

// NewConfigAdapter creates a new adapter from the given values
func NewConfigAdapter(startRecording string) *ConfigAdapter {
	return &ConfigAdapter{
		startRecording: startRecording,
	}
}

// GetStartRecordingHotkey returns the start recording hotkey
func (c *ConfigAdapter) GetStartRecordingHotkey() string {
	return c.startRecording
}
