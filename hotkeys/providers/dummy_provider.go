// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"fmt"

	"github.com/AshBuk/dabri/v2/internal/logger"
)

// DummyKeyboardProvider implements a KeyboardEventProvider with no actual functionality
// Use as a fallback when no other providers are available
type DummyKeyboardProvider struct {
	callbacks map[string]func(bool) error
	logger    logger.Logger
}

// NewDummyKeyboardProvider creates a new DummyKeyboardProvider
func NewDummyKeyboardProvider(logger logger.Logger) *DummyKeyboardProvider {
	return &DummyKeyboardProvider{
		callbacks: make(map[string]func(bool) error),
		logger:    logger,
	}
}

// IsSupported returns true as the dummy provider is always supported as a fallback
func (p *DummyKeyboardProvider) IsSupported() bool {
	return true
}

// Start logs helpful instructions and fails, so the app can tell the user
// that hotkeys are unavailable
func (p *DummyKeyboardProvider) Start() error {
	p.logger.Warning("Using dummy keyboard provider; hotkeys will not be functional")
	p.logger.Info("")
	p.logger.Info("To enable hotkeys, try one of these solutions:")
	p.logger.Info("")
	p.logger.Info("🔧 Modern Desktop Environments (GNOME/KDE):")
	p.logger.Info("   - Ensure D-Bus session is running")
	p.logger.Info("   - Check if 'dbus-daemon --session' is active")
	p.logger.Info("")
	p.logger.Info("🔧 Other Desktop Environments (XFCE/i3/sway):")
	p.logger.Info("   - Add your user to the 'input' group: sudo usermod -a -G input $USER")
	p.logger.Info("   - Then logout and login again")
	p.logger.Info("   - Or run the application with sudo (not recommended)")
	p.logger.Info("")
	p.logger.Info("🔧 Alternative Solutions:")
	p.logger.Info("   - Use system hotkey tools like 'sxhkd' or 'xbindkeys'")
	p.logger.Info("   - Configure DE-specific keyboard shortcuts")
	p.logger.Info("   - Use the WebSocket interface for remote control")
	p.logger.Info("")

	return fmt.Errorf("no hotkey provider available")
}

// Stop does nothing as the dummy provider never listens
func (p *DummyKeyboardProvider) Stop() {}

// RegisterHotkey stores the callback but never calls it
func (p *DummyKeyboardProvider) RegisterHotkey(id, hotkey string, callback func(pressed bool) error) error {
	p.logger.Info("Registered action: %s (%s) (but it will not function with dummy provider)", id, hotkey)
	p.callbacks[id] = callback
	return nil
}
