// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package adapters

import (
	"testing"
)

func TestNewConfigAdapter(t *testing.T) {
	tests := []struct {
		name           string
		startRecording string
		provider       string
	}{
		{
			name:           "standard hotkey",
			startRecording: "ctrl+shift+r",
			provider:       "auto",
		},
		{
			name:           "single key",
			startRecording: "F12",
			provider:       "dbus",
		},
		{
			name:           "complex hotkey",
			startRecording: "altgr+comma",
			provider:       "evdev",
		},
		{
			name:           "empty hotkey",
			startRecording: "",
			provider:       "auto",
		},
		{
			name:           "special characters",
			startRecording: "ctrl+alt+/",
			provider:       "auto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.startRecording, tt.provider)

			if adapter == nil {
				t.Fatalf("NewConfigAdapter returned nil")
			}

			if adapter.startRecording != tt.startRecording {
				t.Errorf("Expected startRecording '%s', got '%s'", tt.startRecording, adapter.startRecording)
			}
			if adapter.provider != tt.provider {
				t.Errorf("Expected provider '%s', got '%s'", tt.provider, adapter.provider)
			}

			// Test that the adapter implements HotkeyConfig interface
			var _ HotkeyConfig = adapter
		})
	}
}

func TestConfigAdapter_GetStartRecordingHotkey(t *testing.T) {
	tests := []struct {
		name           string
		startRecording string
		expected       string
	}{
		{
			name:           "standard hotkey",
			startRecording: "ctrl+shift+r",
			expected:       "ctrl+shift+r",
		},
		{
			name:           "single key",
			startRecording: "Space",
			expected:       "Space",
		},
		{
			name:           "empty hotkey",
			startRecording: "",
			expected:       "",
		},
		{
			name:           "unicode characters",
			startRecording: "ctrl+ñ",
			expected:       "ctrl+ñ",
		},
		{
			name:           "mixed case",
			startRecording: "Ctrl+Shift+R",
			expected:       "Ctrl+Shift+R",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.startRecording, "auto")
			result := adapter.GetStartRecordingHotkey()

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestConfigAdapter_InterfaceCompliance(t *testing.T) {
	// Test that ConfigAdapter properly implements HotkeyConfig interface
	adapter := NewConfigAdapter("test+hotkey", "auto")

	// This should compile without issues if the interface is implemented correctly
	var config HotkeyConfig = adapter

	hotkey := config.GetStartRecordingHotkey()
	expected := "test+hotkey"

	if hotkey != expected {
		t.Errorf("Interface method returned '%s', expected '%s'", hotkey, expected)
	}
}

func TestConfigAdapter_Immutability(t *testing.T) {
	// Test that the adapter's internal state cannot be modified externally
	originalHotkey := "ctrl+r"
	adapter := NewConfigAdapter(originalHotkey, "auto")

	// Get the hotkey multiple times to ensure it doesn't change
	first := adapter.GetStartRecordingHotkey()
	second := adapter.GetStartRecordingHotkey()
	third := adapter.GetStartRecordingHotkey()

	if first != originalHotkey || second != originalHotkey || third != originalHotkey {
		t.Errorf("Adapter state changed: first='%s', second='%s', third='%s', expected='%s'",
			first, second, third, originalHotkey)
	}
}

func TestConfigAdapter_MultipleInstances(t *testing.T) {
	// Test that multiple adapters are independent
	adapter1 := NewConfigAdapter("ctrl+1", "auto")
	adapter2 := NewConfigAdapter("ctrl+2", "auto")
	adapter3 := NewConfigAdapter("ctrl+3", "auto")

	if adapter1.GetStartRecordingHotkey() != "ctrl+1" {
		t.Errorf("Adapter1 hotkey incorrect: got '%s'", adapter1.GetStartRecordingHotkey())
	}
	if adapter2.GetStartRecordingHotkey() != "ctrl+2" {
		t.Errorf("Adapter2 hotkey incorrect: got '%s'", adapter2.GetStartRecordingHotkey())
	}
	if adapter3.GetStartRecordingHotkey() != "ctrl+3" {
		t.Errorf("Adapter3 hotkey incorrect: got '%s'", adapter3.GetStartRecordingHotkey())
	}

	// Verify they are different instances
	if adapter1 == adapter2 || adapter2 == adapter3 || adapter1 == adapter3 {
		t.Error("Adapters should be different instances")
	}
}

func TestConfigAdapter_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "very long hotkey",
			input:    "ctrl+alt+shift+super+meta+hyper+F12",
			expected: "ctrl+alt+shift+super+meta+hyper+F12",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "   ",
		},
		{
			name:     "special symbols",
			input:    "ctrl+!@#$%^&*()",
			expected: "ctrl+!@#$%^&*()",
		},
		{
			name:     "numbers",
			input:    "ctrl+1+2+3",
			expected: "ctrl+1+2+3",
		},
		{
			name:     "mixed separators",
			input:    "ctrl-alt+shift_super",
			expected: "ctrl-alt+shift_super",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.input, "auto")
			result := adapter.GetStartRecordingHotkey()

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestConfigAdapter_GetProvider_DefaultAuto(t *testing.T) {
	adapter := NewConfigAdapter("ctrl+r", "")
	if p := adapter.GetProvider(); p != "auto" {
		t.Errorf("expected default provider 'auto', got '%s'", p)
	}
}
