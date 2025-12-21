// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package interfaces

// Defines the contract for text output operations
type Outputter interface {
	// Copy text to the system clipboard
	CopyToClipboard(text string) error
	// Simulate typing text into the currently active window
	TypeToActiveWindow(text string) error
	// Return the names of the underlying tools being used
	GetToolNames() (clipboardTool, typeTool string)
}
