// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"fmt"
	"time"

	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// DummyKeyboardProvider implements KeyboardEventProvider with no actual functionality
// Used as a fallback when no other providers are available
type DummyKeyboardProvider struct {
	callbacks   map[string]func() error
	isListening bool
	logger      logger.Logger
}

// NewDummyKeyboardProvider creates a new DummyKeyboardProvider
func NewDummyKeyboardProvider(logger logger.Logger) *DummyKeyboardProvider {
	return &DummyKeyboardProvider{
		callbacks:   make(map[string]func() error),
		isListening: false,
		logger:      logger,
	}
}

// IsSupported always returns true as the dummy provider is always supported
func (p *DummyKeyboardProvider) IsSupported() bool {
	return true
}

// Start does nothing but logs helpful instructions
func (p *DummyKeyboardProvider) Start() error {
	if p.isListening {
		return fmt.Errorf("dummy keyboard provider already started")
	}

	p.isListening = true
	p.logger.Warning("Using dummy keyboard provider. Hotkeys will not be functional.")
	p.logger.Info("")
	p.logger.Info("To enable hotkeys, try one of these solutions:")
	p.logger.Info("")
	p.logger.Info("🔧 Modern Desktop Environments (GNOME/KDE):")
	p.logger.Info("   - Ensure D-Bus session is running")
	p.logger.Info("   - Check if 'dbus-daemon --session' is active")
	p.logger.Info("")
	p.logger.Info("🔧 Other Desktop Environments (XFCE/i3/sway):")
	p.logger.Info("   - Add your user to 'input' group: sudo usermod -a -G input $USER")
	p.logger.Info("   - Then logout and login again")
	p.logger.Info("   - Or run the application with sudo (not recommended)")
	p.logger.Info("")
	p.logger.Info("🔧 Alternative Solutions:")
	p.logger.Info("   - Use system hotkey tools like 'sxhkd' or 'xbindkeys'")
	p.logger.Info("   - Configure DE-specific keyboard shortcuts")
	p.logger.Info("   - Use the WebSocket interface for remote control")
	p.logger.Info("")

	return nil
}

// Stop does nothing but changes the state
func (p *DummyKeyboardProvider) Stop() {
	p.isListening = false
}

// RegisterHotkey just stores the callback but never calls it
func (p *DummyKeyboardProvider) RegisterHotkey(hotkey string, callback func() error) error {
	p.logger.Info("Registered hotkey: %s (but it will not function with dummy provider)", hotkey)
	p.callbacks[hotkey] = callback
	return nil
}

// CaptureOnce is not supported in dummy provider
func (p *DummyKeyboardProvider) CaptureOnce(timeout time.Duration) (string, error) {
	return "", fmt.Errorf("captureOnce not supported in dummy provider")
}
