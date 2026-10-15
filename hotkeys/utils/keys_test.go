// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import "testing"

func TestIsModifierPressed(t *testing.T) {
	state := map[string]bool{"rightctrl": true, "leftalt": true, "leftmeta": true}
	tests := map[string]bool{
		"ctrl": true, "Ctrl": true, "leftctrl": false, "rightctrl": true,
		"alt": true, "leftalt": true, "rightalt": false, "altgr": false,
		"shift": false, "leftshift": false, "rightshift": false,
		"super": true, "meta": true, "win": true, "leftmeta": true, "rightmeta": false,
		"capslock": false,
	}
	for mod, want := range tests {
		if got := IsModifierPressed(mod, state); got != want {
			t.Errorf("IsModifierPressed(%q) = %v, want %v", mod, got, want)
		}
	}
	if !IsModifierPressed("altgr", map[string]bool{"rightalt": true}) {
		t.Error("altgr should match rightalt")
	}
}
