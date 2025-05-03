package hotkeys

// HotkeyConfig defines the interface that a configuration must implement
// to be used with the hotkey system
type HotkeyConfig interface {
	GetStartRecordingHotkey() string
	GetCopyToClipboardHotkey() string
	GetPasteToActiveAppHotkey() string
}

// ConfigAdapter adapts any config to HotkeyConfig
type ConfigAdapter struct {
	startRecording   string
	copyToClipboard  string
	pasteToActiveApp string
}

// NewConfigAdapter creates a new adapter from the given values
func NewConfigAdapter(startRecording, copyToClipboard, pasteToActiveApp string) *ConfigAdapter {
	return &ConfigAdapter{
		startRecording:   startRecording,
		copyToClipboard:  copyToClipboard,
		pasteToActiveApp: pasteToActiveApp,
	}
}

// GetStartRecordingHotkey returns the start recording hotkey
func (c *ConfigAdapter) GetStartRecordingHotkey() string {
	return c.startRecording
}

// GetCopyToClipboardHotkey returns the copy to clipboard hotkey
func (c *ConfigAdapter) GetCopyToClipboardHotkey() string {
	return c.copyToClipboard
}

// GetPasteToActiveAppHotkey returns the paste to active app hotkey
func (c *ConfigAdapter) GetPasteToActiveAppHotkey() string {
	return c.pasteToActiveApp
}
