// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package interfaces

import "time"

// Defines the contract for a keyboard event source
type KeyboardEventProvider interface {
	// Start listening for keyboard events
	Start() error
	// Stop listening for keyboard events
	Stop()
	// Register a hotkey combination and its associated callback function
	RegisterHotkey(hotkey string, callback func() error) error
	// Check if the provider is supported on the current system
	IsSupported() bool
	// Capture a single hotkey combination within a given timeout
	CaptureOnce(timeout time.Duration) (string, error)
	// Check if the provider supports one-shot hotkey capture
	SupportsCaptureOnce() bool
}

// Represents a hotkey combination
type KeyCombination struct {
	Modifiers []string // Modifier keys like "ctrl", "alt", "shift"
	Key       string   // The primary, non-modifier key
}

// Defines the type of desktop environment
type EnvironmentType int

const (
	EnvironmentUnknown EnvironmentType = iota
	EnvironmentWayland
	EnvironmentX11
)
