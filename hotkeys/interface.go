// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package hotkeys

// KeyboardEventProvider defines an interface for keyboard event sources
type KeyboardEventProvider interface {
	Start() error
	Stop()
	RegisterHotkey(hotkey string, callback func() error) error
	IsSupported() bool
}

// KeyCombination represents a hotkey combination
type KeyCombination struct {
	Modifiers []string // Modifier keys like "ctrl", "alt", "shift"
	Key       string   // Main key
}

// EnvironmentType defines the type of desktop environment
type EnvironmentType int

const (
	EnvironmentUnknown EnvironmentType = iota
	EnvironmentWayland
	EnvironmentX11
)
