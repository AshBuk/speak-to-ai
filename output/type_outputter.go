package output

import (
	"fmt"
	"os/exec"

	"github.com/AshBuk/speak-to-ai/config"
)

// TypeOutputter implements Outputter for typing in active window
type TypeOutputter struct {
	typeTool string
	config   *config.Config
}

// NewTypeOutputter creates a new type outputter
func NewTypeOutputter(typeTool string, cfg *config.Config) (Outputter, error) {
	// Verify tool exists
	if _, err := exec.LookPath(typeTool); err != nil {
		return nil, fmt.Errorf("type tool not found: %s", typeTool)
	}

	return &TypeOutputter{
		typeTool: typeTool,
		config:   cfg,
	}, nil
}

// TypeToActiveWindow simulates typing in the currently active window
func (o *TypeOutputter) TypeToActiveWindow(text string) error {
	// Security: validate command before execution
	if !o.config.IsCommandAllowed(o.typeTool) {
		return fmt.Errorf("typing tool not allowed: %s", o.typeTool)
	}

	var cmd *exec.Cmd
	var args []string

	switch o.typeTool {
	case "xdotool":
		args = []string{"type", "--clearmodifiers", text}
	case "wtype":
		// wtype: Wayland typing tool
		args = []string{text}
	case "ydotool":
		// ydotool: Universal input automation tool (works on both X11 and Wayland)
		args = []string{"type", text}
	default:
		return fmt.Errorf("unsupported typing tool: %s", o.typeTool)
	}

	// Security: sanitize arguments (text is user content, needs special handling)
	safeArgs := config.SanitizeCommandArgs(args)
	cmd = exec.Command(o.typeTool, safeArgs...)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to type text with %s: %w, output: %s", o.typeTool, err, string(output))
	}

	return nil
}

// CopyToClipboard is a no-op for type outputter
func (o *TypeOutputter) CopyToClipboard(text string) error {
	// Not supported by type outputter
	return fmt.Errorf("copying to clipboard not supported by type outputter")
}
