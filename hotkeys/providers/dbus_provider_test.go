// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"testing"

	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
	"github.com/AshBuk/speak-to-ai/hotkeys/interfaces"
	"github.com/AshBuk/speak-to-ai/internal/testutils"
)

func TestNewDbusKeyboardProvider(t *testing.T) {
	config := adapters.NewConfigAdapter("ctrl+shift+r", "auto")
	env := interfaces.EnvironmentWayland

	provider := NewDbusKeyboardProvider(config, env, testutils.NewMockLogger())

	if provider == nil {
		t.Fatal("NewDbusKeyboardProvider returned nil")
	}

	if provider.callbacks == nil {
		t.Error("callbacks map not initialized")
	}
	if provider.isListening {
		t.Error("should not be listening initially")
	}
}

func TestDbusKeyboardProvider_IsSupported_NewTest(t *testing.T) {
	config := adapters.NewConfigAdapter("ctrl+shift+r", "auto")
	provider := NewDbusKeyboardProvider(config, interfaces.EnvironmentWayland, testutils.NewMockLogger())

	// Test IsSupported - this will likely return false in test environment
	supported := provider.IsSupported()

	// We don't assert true/false here since it depends on the test environment
	// Just verify the method doesn't panic and returns a boolean
	if supported {
		t.Log("D-Bus portal GlobalShortcuts is supported in test environment")
	} else {
		t.Log("D-Bus portal GlobalShortcuts is not supported in test environment (expected)")
	}
}

func TestDbusKeyboardProvider_RegisterHotkey(t *testing.T) {
	config := adapters.NewConfigAdapter("ctrl+shift+r", "auto")
	provider := NewDbusKeyboardProvider(config, interfaces.EnvironmentWayland, testutils.NewMockLogger())

	callbackCalled := false
	callback := func() error {
		callbackCalled = true
		return nil
	}

	// Test registering a hotkey
	err := provider.RegisterHotkey("ctrl+shift+a", callback)
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
		err := storedCallback()
		if err != nil {
			t.Errorf("callback returned error: %v", err)
		}
		if !callbackCalled {
			t.Error("callback was not called")
		}
	}
}

func TestDbusKeyboardProvider_RegisterHotkey_Duplicate(t *testing.T) {
	config := adapters.NewConfigAdapter("ctrl+shift+r", "auto")
	provider := NewDbusKeyboardProvider(config, interfaces.EnvironmentWayland, testutils.NewMockLogger())

	callback := func() error { return nil }

	// Register first hotkey
	err := provider.RegisterHotkey("ctrl+a", callback)
	if err != nil {
		t.Errorf("unexpected error registering first hotkey: %v", err)
	}

	// Try to register the same hotkey again
	err = provider.RegisterHotkey("ctrl+a", callback)
	if err == nil {
		t.Error("expected error when registering duplicate hotkey")
	}
	if err.Error() != "hotkey ctrl+a already registered" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDbusKeyboardProvider_Start_AlreadyStarted(t *testing.T) {
	config := adapters.NewConfigAdapter("ctrl+shift+r", "auto")
	provider := NewDbusKeyboardProvider(config, interfaces.EnvironmentWayland, testutils.NewMockLogger())

	// Set isListening to true to simulate already started
	provider.isListening = true

	err := provider.Start()
	if err == nil {
		t.Error("expected error when starting already started provider")
	}
	if err.Error() != "D-Bus keyboard provider already started" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDbusKeyboardProvider_Stop_NotStarted(t *testing.T) {
	config := adapters.NewConfigAdapter("ctrl+shift+r", "auto")
	provider := NewDbusKeyboardProvider(config, interfaces.EnvironmentWayland, testutils.NewMockLogger())

	// Stop should not panic even if not started
	provider.Stop()

	// Verify state
	if provider.isListening {
		t.Error("isListening should remain false")
	}
}

func TestDbusKeyboardProvider_Stop_WhenStarted(t *testing.T) {
	config := adapters.NewConfigAdapter("ctrl+shift+r", "auto")
	provider := NewDbusKeyboardProvider(config, interfaces.EnvironmentWayland, testutils.NewMockLogger())

	// Simulate started state
	provider.isListening = true

	provider.Stop()

	if provider.isListening {
		t.Error("expected isListening to be false after stop")
	}
	if provider.conn != nil {
		t.Error("expected connection to be nil after stop")
	}
}

func TestContainsGlobalShortcuts(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected bool
	}{
		{
			name:     "contains GlobalShortcuts",
			data:     `<interface name="org.freedesktop.portal.GlobalShortcuts">`,
			expected: true,
		},
		{
			name:     "does not contain GlobalShortcuts",
			data:     `<interface name="org.freedesktop.portal.FileChooser">`,
			expected: false,
		},
		{
			name:     "empty data",
			data:     "",
			expected: false,
		},
		{
			name:     "contains partial match",
			data:     "Some text with Global but not the full interface",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsGlobalShortcuts(tt.data)
			if result != tt.expected {
				t.Errorf("containsGlobalShortcuts(%q) = %v, want %v", tt.data, result, tt.expected)
			}
		})
	}
}
