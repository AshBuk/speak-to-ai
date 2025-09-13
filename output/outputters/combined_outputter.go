// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package outputters

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/output/interfaces"
)

// CombinedOutputter implements interfaces.Outputter for both clipboard and typing
type CombinedOutputter struct {
	clipboardOutputter interfaces.Outputter
	typeOutputter      interfaces.Outputter
}

// NewCombinedOutputter creates a new combined outputter
func NewCombinedOutputter(clipboardTool, typeTool string, cfg *config.Config) (interfaces.Outputter, error) {
	clipboard, err := NewClipboardOutputter(clipboardTool, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create clipboard outputter: %w", err)
	}

	typer, err := NewTypeOutputter(typeTool, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create type outputter: %w", err)
	}

	return &CombinedOutputter{
		clipboardOutputter: clipboard,
		typeOutputter:      typer,
	}, nil
}

// CopyToClipboard copies text to the system clipboard
func (o *CombinedOutputter) CopyToClipboard(text string) error {
	return o.clipboardOutputter.CopyToClipboard(text)
}

// TypeToActiveWindow simulates typing in the currently active window
func (o *CombinedOutputter) TypeToActiveWindow(text string) error {
	return o.typeOutputter.TypeToActiveWindow(text)
}

// GetToolNames returns the actual tool names being used
func (o *CombinedOutputter) GetToolNames() (clipboardTool, typeTool string) {
	clipboardTool, _ = o.clipboardOutputter.GetToolNames()
	_, typeTool = o.typeOutputter.GetToolNames()
	return clipboardTool, typeTool
}
