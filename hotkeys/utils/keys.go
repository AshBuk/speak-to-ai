// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"strings"
)

// GetKeyName converts key code to key name using common mapping
func GetKeyName(keyCode int) string {
	keyMap := map[int]string{
		1:   "esc",
		2:   "1",
		3:   "2",
		4:   "3",
		5:   "4",
		6:   "5",
		7:   "6",
		8:   "7",
		9:   "8",
		10:  "9",
		11:  "0",
		12:  "minus",
		13:  "equal",
		14:  "backspace",
		15:  "tab",
		16:  "q",
		17:  "w",
		18:  "e",
		19:  "r",
		20:  "t",
		21:  "y",
		22:  "u",
		23:  "i",
		24:  "o",
		25:  "p",
		26:  "leftbrace",
		27:  "rightbrace",
		28:  "enter",
		29:  "leftctrl",
		30:  "a",
		31:  "s",
		32:  "d",
		33:  "f",
		34:  "g",
		35:  "h",
		36:  "j",
		37:  "k",
		38:  "l",
		39:  "semicolon",
		40:  "apostrophe",
		41:  "grave",
		42:  "leftshift",
		43:  "backslash",
		44:  "z",
		45:  "x",
		46:  "c",
		47:  "v",
		48:  "b",
		49:  "n",
		50:  "m",
		51:  "comma",
		52:  "dot",
		53:  "slash",
		54:  "rightshift",
		55:  "kpasterisk",
		56:  "leftalt",
		57:  "space",
		58:  "capslock",
		59:  "f1",
		60:  "f2",
		61:  "f3",
		62:  "f4",
		63:  "f5",
		64:  "f6",
		65:  "f7",
		66:  "f8",
		67:  "f9",
		68:  "f10",
		69:  "numlock",
		70:  "scrolllock",
		71:  "kp7",
		72:  "kp8",
		73:  "kp9",
		74:  "kpminus",
		75:  "kp4",
		76:  "kp5",
		77:  "kp6",
		78:  "kpplus",
		79:  "kp1",
		80:  "kp2",
		81:  "kp3",
		82:  "kp0",
		83:  "kpdot",
		97:  "rightctrl",
		100: "rightalt",
		125: "leftmeta",
		126: "rightmeta",
	}

	if name, ok := keyMap[keyCode]; ok {
		return name
	}
	return ""
}

// IsModifierPressed determines if a modifier (generic or side-specific) is pressed
// based on the modifier state map
func IsModifierPressed(mod string, state map[string]bool) bool {
	m := strings.ToLower(mod)
	switch m {
	case "ctrl":
		return state["leftctrl"] || state["rightctrl"]
	case "leftctrl":
		return state["leftctrl"]
	case "rightctrl":
		return state["rightctrl"]
	case "alt":
		return state["leftalt"] || state["rightalt"]
	case "leftalt":
		return state["leftalt"]
	case "rightalt", "altgr":
		return state["rightalt"]
	case "shift":
		return state["leftshift"] || state["rightshift"]
	case "leftshift":
		return state["leftshift"]
	case "rightshift":
		return state["rightshift"]
	case "super", "meta", "win":
		return state["leftmeta"] || state["rightmeta"]
	case "leftmeta":
		return state["leftmeta"]
	case "rightmeta":
		return state["rightmeta"]
	default:
		// Fallback to simple mapping for any other names
		return state[ConvertModifierToEvdev(m)]
	}
}

// BuildModifierState builds a modifier state map from evdev state for CaptureOnce
func BuildModifierState(modState map[string]bool) []string {
	mods := make([]string, 0, 5)

	if modState["leftctrl"] || modState["rightctrl"] {
		mods = append(mods, "ctrl")
	}
	if modState["leftshift"] || modState["rightshift"] {
		mods = append(mods, "shift")
	}
	if modState["leftalt"] || modState["rightalt"] {
		mods = append(mods, "alt")
	}
	if modState["rightalt"] {
		mods = append(mods, "altgr")
	}
	if modState["leftmeta"] || modState["rightmeta"] {
		mods = append(mods, "super")
	}

	return mods
}

// IsModifierKey checks if a key name is a modifier or meta key for tracking state
func IsModifierKey(keyName string) bool {
	switch strings.ToLower(keyName) {
	case "leftctrl", "rightctrl", "leftalt", "rightalt",
		"leftshift", "rightshift", "leftmeta", "rightmeta":
		return true
	}
	return IsModifier(keyName)
}

// CheckCancelCondition checks if Esc was pressed without modifiers (for CaptureOnce)
func CheckCancelCondition(keyName string, modState map[string]bool) bool {
	if !strings.EqualFold(keyName, "esc") {
		return false
	}

	// Check if no modifiers are pressed
	return !modState["leftctrl"] && !modState["rightctrl"] &&
		!modState["leftshift"] && !modState["rightshift"] &&
		!modState["leftalt"] && !modState["rightalt"] &&
		!modState["leftmeta"] && !modState["rightmeta"]
}

// BuildHotkeyString builds a hotkey string from modifiers and key
func BuildHotkeyString(mods []string, keyName string) string {
	combo := strings.ToLower(keyName)
	if len(mods) > 0 {
		combo = strings.Join(mods, "+") + "+" + combo
	}
	return NormalizeHotkey(combo)
}
