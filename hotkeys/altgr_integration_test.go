//go:build linux

package hotkeys

import (
	"testing"
)

func TestAltGrHotkeyIntegration(t *testing.T) {
	tests := []struct {
		name         string
		hotkeyString string
		expectedKey  string
		expectedMods []string
	}{
		{
			name:         "altgr+comma combination",
			hotkeyString: "altgr+comma",
			expectedKey:  "comma",
			expectedMods: []string{"altgr"},
		},
		{
			name:         "AltGr+comma case insensitive",
			hotkeyString: "AltGr+comma",
			expectedKey:  "comma",
			expectedMods: []string{"altgr"}, // ParseHotkey normalizes to lowercase
		},
		{
			name:         "complex combination with altgr",
			hotkeyString: "ctrl+altgr+shift+a",
			expectedKey:  "a",
			expectedMods: []string{"ctrl", "altgr", "shift"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ParseHotkey function
			combo := ParseHotkey(tt.hotkeyString)

			if combo.Key != tt.expectedKey {
				t.Errorf("Expected key '%s', got '%s'", tt.expectedKey, combo.Key)
			}

			if len(combo.Modifiers) != len(tt.expectedMods) {
				t.Errorf("Expected %d modifiers, got %d", len(tt.expectedMods), len(combo.Modifiers))
			}

			// Check that modifiers match (order doesn't matter)
			for _, expectedMod := range tt.expectedMods {
				found := false
				for _, actualMod := range combo.Modifiers {
					if actualMod == expectedMod {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected modifier '%s' not found in %v", expectedMod, combo.Modifiers)
				}
			}
		})
	}
}

func TestAltGrEvdevConversion(t *testing.T) {
	// Test the full pipeline: hotkey string -> parsed combo -> evdev conversion
	hotkeyString := "altgr+comma"
	combo := ParseHotkey(hotkeyString)

	// Verify parsing
	if combo.Key != "comma" {
		t.Errorf("Expected key 'comma', got '%s'", combo.Key)
	}

	if len(combo.Modifiers) != 1 || combo.Modifiers[0] != "altgr" {
		t.Errorf("Expected modifiers ['altgr'], got %v", combo.Modifiers)
	}

	// Test conversion to evdev format
	for _, mod := range combo.Modifiers {
		evdevMod := ConvertModifierToEvdev(mod)
		if mod == "altgr" && evdevMod != "rightalt" {
			t.Errorf("Expected altgr to convert to rightalt, got '%s'", evdevMod)
		}
	}
}

func TestEvdevProvider_AltGrSupport(t *testing.T) {
	// Test that evdev provider can handle AltGr keys
	config := NewConfigAdapter("altgr+comma")
	provider := NewEvdevKeyboardProvider(config, EnvironmentWayland)

	// Test registering AltGr hotkey
	callback := func() error {
		return nil
	}

	err := provider.RegisterHotkey("altgr+comma", callback)
	// We expect this to fail in test environment (no /dev/input access)
	// but the important thing is that it doesn't panic
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}

	// Verify callback was registered
	if _, exists := provider.callbacks["altgr+comma"]; !exists {
		t.Error("AltGr hotkey callback should be registered")
	}
}
