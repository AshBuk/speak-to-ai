// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"slices"
	"testing"

	"github.com/AshBuk/dabri/v2/hotkeys/utils"
	"github.com/AshBuk/dabri/v2/internal/testutils"
	evdev "github.com/holoplot/go-evdev"
)

func TestNewEvdevKeyboardProvider(t *testing.T) {
	provider := NewEvdevKeyboardProvider(testutils.NewMockLogger())
	if provider == nil {
		t.Fatal("NewEvdevKeyboardProvider returned nil")
	}

	if provider.callbacks == nil {
		t.Error("callbacks map not initialized")
	}
	if provider.stopListening == nil {
		t.Error("stopListening channel not initialized")
	}
	if provider.isListening {
		t.Error("should not be listening initially")
	}
	if provider.modifierState == nil {
		t.Error("modifierState map not initialized")
	}
}

func TestEvdevKeyboardProvider_RegisterHotkey(t *testing.T) {
	provider := NewEvdevKeyboardProvider(testutils.NewMockLogger())
	callbackCalled := false
	callback := func(bool) error {
		callbackCalled = true
		return nil
	}
	// Test registering a hotkey
	err := provider.RegisterHotkey("ctrl+shift+a", "ctrl+shift+a", callback)
	if err != nil {
		t.Errorf("unexpected error registering hotkey: %v", err)
	}
	// Check that callback is stored
	if len(provider.callbacks) != 1 {
		t.Errorf("expected 1 callback, got %d", len(provider.callbacks))
	}

	storedCallback, exists := provider.callbacks["ctrl+shift+a"]
	if !exists {
		t.Error("hotkey not found in callbacks")
	}

	// Test the stored callback
	if storedCallback != nil {
		err := storedCallback(true)
		if err != nil {
			t.Errorf("callback returned error: %v", err)
		}
		if !callbackCalled {
			t.Error("callback was not called")
		}
	}
}

func TestEvdevKeyboardProvider_Start_AlreadyStarted(t *testing.T) {
	provider := NewEvdevKeyboardProvider(testutils.NewMockLogger())
	// Set isListening to true to simulate already started
	provider.isListening = true

	err := provider.Start()
	if err == nil {
		t.Error("expected error when starting already started provider")
	}
	if err.Error() != "evdev keyboard provider already started" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestEvdevKeyboardProvider_Stop_NotStarted(t *testing.T) {
	provider := NewEvdevKeyboardProvider(testutils.NewMockLogger())
	// Stop should not panic even if not started
	provider.Stop()
	// Verify state
	if provider.isListening {
		t.Error("isListening should remain false")
	}
}

func TestEvdevKeyboardProvider_Stop_WhenStarted(t *testing.T) {
	provider := NewEvdevKeyboardProvider(testutils.NewMockLogger())
	// Simulate started state
	provider.isListening = true
	provider.stopListening = make(chan bool)

	provider.Stop()

	if provider.isListening {
		t.Error("expected isListening to be false after stop")
	}
	if provider.devices != nil {
		t.Error("expected devices to be nil after stop")
	}
}

func TestGetKeyName(t *testing.T) {
	tests := []struct {
		name     string
		keyCode  int
		expected string
	}{
		{
			name:     "escape key",
			keyCode:  1,
			expected: "esc",
		},
		{
			name:     "number 1",
			keyCode:  2,
			expected: "1",
		},
		{
			name:     "letter a",
			keyCode:  30,
			expected: "a",
		},
		{
			name:     "enter key",
			keyCode:  28,
			expected: "enter",
		},
		{
			name:     "space key",
			keyCode:  57,
			expected: "space",
		},
		{
			name:     "left ctrl",
			keyCode:  29,
			expected: "leftctrl",
		},
		{
			name:     "right ctrl",
			keyCode:  97,
			expected: "rightctrl",
		},
		{
			name:     "left alt",
			keyCode:  56,
			expected: "leftalt",
		},
		{
			name:     "right alt",
			keyCode:  100,
			expected: "rightalt",
		},
		{
			name:     "left shift",
			keyCode:  42,
			expected: "leftshift",
		},
		{
			name:     "right shift",
			keyCode:  54,
			expected: "rightshift",
		},
		{
			name:     "comma",
			keyCode:  51,
			expected: "comma",
		},
		{
			name:     "period",
			keyCode:  52,
			expected: "dot",
		},
		{
			name:     "unknown key",
			keyCode:  999,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.GetKeyName(tt.keyCode)
			if result != tt.expected {
				t.Errorf("GetKeyName(%d) = %q, want %q", tt.keyCode, result, tt.expected)
			}
		})
	}
}

// Test helper functions that are used by the evdev provider

func TestEvdevKeyboardProvider_PressAndRelease(t *testing.T) {
	p := NewEvdevKeyboardProvider(testutils.NewMockLogger())
	var got []bool
	_ = p.RegisterHotkey("toggle_recording", "alt+r", func(pressed bool) error {
		got = append(got, pressed)
		return nil
	})
	key := func(code uint16, value int32) {
		p.handleKeyEvent(0, &evdev.InputEvent{Type: evdev.EV_KEY, Code: evdev.EvCode(code), Value: value})
	}
	const alt, r = 56, 19

	key(r, 1) // no modifier: not a hotkey press
	key(r, 0) // release without a pending press is ignored
	key(alt, 1)
	key(r, 1)
	key(r, 2) // autorepeat is ignored
	key(alt, 0)
	key(r, 0) // release fires even after the modifier is let go

	if !slices.Equal(got, []bool{true, false}) {
		t.Errorf("callbacks = %v, want [true false]", got)
	}
}
