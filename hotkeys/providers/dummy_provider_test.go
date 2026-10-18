// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"testing"

	"github.com/AshBuk/dabri/v2/hotkeys/interfaces"
	"github.com/AshBuk/dabri/v2/internal/testutils"
)

func TestNewDummyKeyboardProvider(t *testing.T) {
	provider := NewDummyKeyboardProvider(testutils.NewMockLogger())

	if provider == nil {
		t.Fatal("NewDummyKeyboardProvider returned nil")
	}

	if provider.callbacks == nil {
		t.Error("callbacks map should be initialized")
	}

	// Test that it implements the interface
	var _ interfaces.KeyboardEventProvider = provider
}

func TestDummyKeyboardProvider_IsSupported(t *testing.T) {
	provider := NewDummyKeyboardProvider(testutils.NewMockLogger())

	// Dummy provider should always be supported
	if !provider.IsSupported() {
		t.Error("DummyKeyboardProvider should always be supported")
	}

	// Test multiple calls
	for i := 0; i < 5; i++ {
		if !provider.IsSupported() {
			t.Errorf("IsSupported should be consistent, failed on call %d", i+1)
		}
	}
}

func TestDummyKeyboardProvider_Start(t *testing.T) {
	provider := NewDummyKeyboardProvider(testutils.NewMockLogger())

	// Start fails so the app can report that hotkeys are unavailable
	if err := provider.Start(); err == nil {
		t.Error("Start should return an error")
	}
}

func TestDummyKeyboardProvider_RegisterHotkey(t *testing.T) {
	provider := NewDummyKeyboardProvider(testutils.NewMockLogger())
	callbackCalled := false

	callback := func(bool) error {
		callbackCalled = true
		return nil
	}

	// Register a hotkey
	err := provider.RegisterHotkey("ctrl+r", "ctrl+r", callback)
	if err != nil {
		t.Errorf("RegisterHotkey should not return error, got: %v", err)
	}

	// Check callback is stored
	if len(provider.callbacks) != 1 {
		t.Errorf("Expected 1 callback, got %d", len(provider.callbacks))
	}

	storedCallback, exists := provider.callbacks["ctrl+r"]
	if !exists {
		t.Error("Callback should be stored with correct key")
	}

	// Test calling the stored callback
	err = storedCallback(true)
	if err != nil {
		t.Errorf("Stored callback returned error: %v", err)
	}

	if !callbackCalled {
		t.Error("Callback should have been called")
	}
}

func TestDummyKeyboardProvider_RegisterMultipleHotkeys(t *testing.T) {
	provider := NewDummyKeyboardProvider(testutils.NewMockLogger())

	hotkeys := []string{
		"ctrl+r",
		"alt+f",
		"shift+space",
		"F12",
		"ctrl+shift+r",
	}

	callbackResults := make(map[string]bool)

	// Register multiple hotkeys
	for _, hotkey := range hotkeys {
		localHotkey := hotkey // capture for closure
		callback := func(bool) error {
			callbackResults[localHotkey] = true
			return nil
		}

		err := provider.RegisterHotkey(hotkey, hotkey, callback)
		if err != nil {
			t.Errorf("RegisterHotkey failed for '%s': %v", hotkey, err)
		}
	}

	// Check all callbacks are stored
	if len(provider.callbacks) != len(hotkeys) {
		t.Errorf("Expected %d callbacks, got %d", len(hotkeys), len(provider.callbacks))
	}

	// Test all callbacks
	for _, hotkey := range hotkeys {
		callback, exists := provider.callbacks[hotkey]
		if !exists {
			t.Errorf("Callback for '%s' not found", hotkey)
			continue
		}

		err := callback(true)
		if err != nil {
			t.Errorf("Callback for '%s' returned error: %v", hotkey, err)
		}
	}

	// Verify all callbacks were called
	for _, hotkey := range hotkeys {
		if !callbackResults[hotkey] {
			t.Errorf("Callback for '%s' was not called", hotkey)
		}
	}
}

func TestDummyKeyboardProvider_OverwriteHotkey(t *testing.T) {
	provider := NewDummyKeyboardProvider(testutils.NewMockLogger())

	firstCallbackCalled := false
	secondCallbackCalled := false

	firstCallback := func(bool) error {
		firstCallbackCalled = true
		return nil
	}

	secondCallback := func(bool) error {
		secondCallbackCalled = true
		return nil
	}

	// Register first callback
	err := provider.RegisterHotkey("ctrl+r", "ctrl+r", firstCallback)
	if err != nil {
		t.Errorf("RegisterHotkey failed: %v", err)
	}

	// Register second callback for same hotkey (should overwrite)
	err = provider.RegisterHotkey("ctrl+r", "ctrl+r", secondCallback)
	if err != nil {
		t.Errorf("RegisterHotkey failed: %v", err)
	}

	// Should only have one callback stored
	if len(provider.callbacks) != 1 {
		t.Errorf("Expected 1 callback, got %d", len(provider.callbacks))
	}

	// Call the stored callback
	callback := provider.callbacks["ctrl+r"]
	err = callback(true)
	if err != nil {
		t.Errorf("Callback returned error: %v", err)
	}

	// Only second callback should have been called
	if firstCallbackCalled {
		t.Error("First callback should not have been called")
	}
	if !secondCallbackCalled {
		t.Error("Second callback should have been called")
	}
}
