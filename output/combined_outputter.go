package output

import (
	"fmt"
)

// CombinedOutputter implements Outputter for both clipboard and typing
type CombinedOutputter struct {
	clipboardOutputter Outputter
	typeOutputter      Outputter
}

// NewCombinedOutputter creates a new combined outputter
func NewCombinedOutputter(clipboardTool, typeTool string) (Outputter, error) {
	clipboard, err := NewClipboardOutputter(clipboardTool)
	if err != nil {
		return nil, fmt.Errorf("failed to create clipboard outputter: %w", err)
	}

	typer, err := NewTypeOutputter(typeTool)
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
