// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package outputters

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/output/interfaces"
)

// ClipboardOutputter implements interfaces.Outputter for clipboard operations
type ClipboardOutputter struct {
	clipboardTool string
	config        *config.Config
}

// NewClipboardOutputter creates a new clipboard outputter
func NewClipboardOutputter(clipboardTool string, cfg *config.Config) (interfaces.Outputter, error) {
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
	if !config.IsCommandAllowed(o.config, o.clipboardTool) {
		return fmt.Errorf("clipboard tool not allowed: %s", o.clipboardTool)
	}

	var cmd *exec.Cmd
	var args []string

	switch o.clipboardTool {
	case "xsel":
		args = []string{"--clipboard", "--input"}
	case "wl-copy":
		args = []string{} // wl-copy takes no additional args for basic operation
	default:
		return fmt.Errorf("unsupported clipboard tool: %s", o.clipboardTool)
	}

	// Security: sanitize arguments
	safeArgs := config.SanitizeCommandArgs(args)
	// #nosec G204 -- Safe: tool is allowlisted and arguments are sanitized.
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

// GetToolNames returns the actual tool names being used
func (o *ClipboardOutputter) GetToolNames() (clipboardTool, typeTool string) {
	return o.clipboardTool, ""
}
