// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package hotkeys

import (
	"testing"
)

func TestNewConfigAdapter(t *testing.T) {
	tests := []struct {
		name           string
		startRecording string
	}{
		{
			name:           "standard hotkey",
			startRecording: "ctrl+shift+r",
		},
		{
			name:           "single key",
			startRecording: "F12",
		},
		{
			name:           "complex hotkey",
			startRecording: "altgr+comma",
		},
		{
			name:           "empty hotkey",
			startRecording: "",
		},
		{
			name:           "special characters",
			startRecording: "ctrl+alt+/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.startRecording)

			if adapter == nil {
				t.Fatalf("NewConfigAdapter returned nil")
			}

			if adapter.startRecording != tt.startRecording {
				t.Errorf("Expected startRecording '%s', got '%s'", tt.startRecording, adapter.startRecording)
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
			adapter := NewConfigAdapter(tt.startRecording)
			result := adapter.GetStartRecordingHotkey()

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestConfigAdapter_InterfaceCompliance(t *testing.T) {
	// Test that ConfigAdapter properly implements HotkeyConfig interface
	adapter := NewConfigAdapter("test+hotkey")

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
	adapter := NewConfigAdapter(originalHotkey)

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
	adapter1 := NewConfigAdapter("ctrl+1")
	adapter2 := NewConfigAdapter("ctrl+2")
	adapter3 := NewConfigAdapter("ctrl+3")

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
			adapter := NewConfigAdapter(tt.input)
			result := adapter.GetStartRecordingHotkey()

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
