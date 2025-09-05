// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package interfaces

// Outputter defines the interface for text output operations
type Outputter interface {
	// CopyToClipboard copies text to the system clipboard
	CopyToClipboard(text string) error

	// TypeToActiveWindow simulates typing in the currently active window
	TypeToActiveWindow(text string) error
}
