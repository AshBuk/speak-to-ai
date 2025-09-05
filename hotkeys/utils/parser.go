// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"strings"

	"github.com/AshBuk/speak-to-ai/hotkeys/interfaces"
)

// ParseHotkey parses a hotkey string like "ctrl+alt+r" into a KeyCombination.
// The last part is treated as the main key, everything else as modifiers.
//
// Examples:
//   - "r" -> KeyCombination{Key: "r", Modifiers: []}
//   - "ctrl+r" -> KeyCombination{Key: "r", Modifiers: ["ctrl"]}
//   - "ctrl+alt+r" -> KeyCombination{Key: "r", Modifiers: ["ctrl", "alt"]}
func ParseHotkey(hotkeyStr string) interfaces.KeyCombination {
	combo := interfaces.KeyCombination{}
	parts := strings.Split(hotkeyStr, "+")
	// If there's only one part, it's just a key
	if len(parts) == 1 {
		combo.Key = strings.TrimSpace(parts[0])
		return combo
	}
	// Last part is the key, the rest are modifiers
	combo.Key = strings.TrimSpace(parts[len(parts)-1])
	for i := 0; i < len(parts)-1; i++ {
		modifier := strings.ToLower(strings.TrimSpace(parts[i]))
		combo.Modifiers = append(combo.Modifiers, modifier)
	}
	return combo
}

// IsModifier checks if a key name represents a modifier key
func IsModifier(keyName string) bool {
	modifiers := map[string]bool{
		"ctrl":       true,
		"alt":        true,
		"shift":      true,
		"super":      true,
		"meta":       true,
		"win":        true,
		"altgr":      true, // AltGr modifier for international keyboards
		"hyper":      true, // Hyper modifier
		"leftctrl":   true,
		"rightctrl":  true,
		"leftalt":    true,
		"rightalt":   true,
		"leftshift":  true,
		"rightshift": true,
	}
	return modifiers[strings.ToLower(keyName)]
}

// ConvertModifierToEvdev converts common modifier names to evdev key names
func ConvertModifierToEvdev(modifier string) string {
	modifierMap := map[string]string{
		"ctrl":  "leftctrl",
		"alt":   "leftalt",
		"shift": "leftshift",
		"super": "leftmeta",
		"meta":  "leftmeta",
		"win":   "leftmeta",
		"altgr": "rightalt",
	}

	if evdevName, ok := modifierMap[strings.ToLower(modifier)]; ok {
		return evdevName
	}

	return strings.ToLower(modifier)
}
