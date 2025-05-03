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

// Start does nothing but logs a warning
func (p *DummyKeyboardProvider) Start() error {
	if p.isListening {
		return fmt.Errorf("dummy keyboard provider already started")
	}

	p.isListening = true
	log.Println("Warning: Using dummy keyboard provider. Hotkeys will not be functional.")
	log.Println("To use hotkeys, please run the application with appropriate permissions or in a supported environment.")

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
