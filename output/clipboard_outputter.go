package output

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/AshBuk/speak-to-ai/config"
)

// ClipboardOutputter implements Outputter for clipboard operations
type ClipboardOutputter struct {
	clipboardTool string
	config        *config.Config
}

// NewClipboardOutputter creates a new clipboard outputter
func NewClipboardOutputter(clipboardTool string, cfg *config.Config) (Outputter, error) {
	// Verify tool exists
	if _, err := exec.LookPath(clipboardTool); err != nil {
		return nil, fmt.Errorf("clipboard tool not found: %s", clipboardTool)
	}

	return &ClipboardOutputter{
		clipboardTool: clipboardTool,
		config:        cfg,
	}, nil
}

// CopyToClipboard copies text to the system clipboard
func (o *ClipboardOutputter) CopyToClipboard(text string) error {
	// Security: validate command before execution
	if !o.config.IsCommandAllowed(o.clipboardTool) {
		return fmt.Errorf("clipboard tool not allowed: %s", o.clipboardTool)
	}

	var cmd *exec.Cmd
	var args []string

	switch o.clipboardTool {
	case "xclip":
		args = []string{"-selection", "clipboard"}
	case "wl-copy":
		args = []string{} // wl-copy takes no additional args for basic operation
	default:
		return fmt.Errorf("unsupported clipboard tool: %s", o.clipboardTool)
	}

	// Security: sanitize arguments
	safeArgs := config.SanitizeCommandArgs(args)
	cmd = exec.Command(o.clipboardTool, safeArgs...)

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
