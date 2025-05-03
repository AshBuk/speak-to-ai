package output

import (
	"fmt"
	"os/exec"
)

// TypeOutputter implements Outputter for typing in active window
type TypeOutputter struct {
	typeTool string
}

// NewTypeOutputter creates a new type outputter
func NewTypeOutputter(typeTool string) (Outputter, error) {
	// Verify tool exists
	if _, err := exec.LookPath(typeTool); err != nil {
		return nil, fmt.Errorf("type tool not found: %s", typeTool)
	}

	return &TypeOutputter{
		typeTool: typeTool,
	}, nil
}

// TypeToActiveWindow simulates typing in the currently active window
func (o *TypeOutputter) TypeToActiveWindow(text string) error {
	var cmd *exec.Cmd

	switch o.typeTool {
	case "xdotool":
		cmd = exec.Command("xdotool", "type", "--clearmodifiers", text)
	default:
		return fmt.Errorf("unsupported typing tool: %s", o.typeTool)
	}

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to type text: %w, output: %s", err, string(output))
	}

	return nil
}

// CopyToClipboard is a no-op for type outputter
func (o *TypeOutputter) CopyToClipboard(text string) error {
	// Not supported by type outputter
	return fmt.Errorf("copying to clipboard not supported by type outputter")
}
