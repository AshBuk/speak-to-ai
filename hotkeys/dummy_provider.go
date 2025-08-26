// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package hotkeys

import (
	"fmt"
	"log"
)

// DummyKeyboardProvider implements KeyboardEventProvider with no actual functionality
// Used as a fallback when no other providers are available
type DummyKeyboardProvider struct {
	callbacks   map[string]func() error
	isListening bool
}

// NewDummyKeyboardProvider creates a new DummyKeyboardProvider
func NewDummyKeyboardProvider() *DummyKeyboardProvider {
	return &DummyKeyboardProvider{
		callbacks:   make(map[string]func() error),
		isListening: false,
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
	log.Println("Warning: Using dummy keyboard provider. Hotkeys will not be functional.")
	log.Println("")
	log.Println("To enable hotkeys, try one of these solutions:")
	log.Println("")
	log.Println("🔧 Modern Desktop Environments (GNOME/KDE):")
	log.Println("   - Ensure D-Bus session is running")
	log.Println("   - Check if 'dbus-daemon --session' is active")
	log.Println("")
	log.Println("🔧 Other Desktop Environments (XFCE/i3/sway):")
	log.Println("   - Add your user to 'input' group: sudo usermod -a -G input $USER")
	log.Println("   - Then logout and login again")
	log.Println("   - Or run the application with sudo (not recommended)")
	log.Println("")
	log.Println("🔧 Alternative Solutions:")
	log.Println("   - Use system hotkey tools like 'sxhkd' or 'xbindkeys'")
	log.Println("   - Configure DE-specific keyboard shortcuts")
	log.Println("   - Use the WebSocket interface for remote control")
	log.Println("")

	return nil
}

// Stop does nothing but changes the state
func (p *DummyKeyboardProvider) Stop() {
	p.isListening = false
}

// RegisterHotkey just stores the callback but never calls it
func (p *DummyKeyboardProvider) RegisterHotkey(hotkey string, callback func() error) error {
	log.Printf("Registered hotkey: %s (but it will not function with dummy provider)", hotkey)
	p.callbacks[hotkey] = callback
	return nil
}
