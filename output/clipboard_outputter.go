package output

import (
	"fmt"
	"os/exec"
	"strings"
)

// ClipboardOutputter implements Outputter for clipboard operations
type ClipboardOutputter struct {
	clipboardTool string
}

// NewClipboardOutputter creates a new clipboard outputter
func NewClipboardOutputter(clipboardTool string) (Outputter, error) {
	// Verify tool exists
	if _, err := exec.LookPath(clipboardTool); err != nil {
		return nil, fmt.Errorf("clipboard tool not found: %s", clipboardTool)
	}

	return &ClipboardOutputter{
		clipboardTool: clipboardTool,
	}, nil
}

// CopyToClipboard copies text to the system clipboard
func (o *ClipboardOutputter) CopyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch o.clipboardTool {
	case "xclip":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "wl-copy":
		cmd = exec.Command("wl-copy")
	default:
		return fmt.Errorf("unsupported clipboard tool: %s", o.clipboardTool)
	}

	// Pipe text to the command
	cmd.Stdin = strings.NewReader(text)

	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}

// TypeToActiveWindow is a no-op for clipboard outputter
func (o *ClipboardOutputter) TypeToActiveWindow(text string) error {
	// Not supported by clipboard outputter
	return fmt.Errorf("typing to active window not supported by clipboard outputter")
}
