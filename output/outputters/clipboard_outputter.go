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

// Implements the Outputter interface for clipboard operations
type ClipboardOutputter struct {
	clipboardTool string
	config        *config.Config
}

// Create a new clipboard outputter
func NewClipboardOutputter(clipboardTool string, cfg *config.Config) (interfaces.Outputter, error) {
	// Verify the required tool exists in the system's PATH
	if _, err := exec.LookPath(clipboardTool); err != nil {
		return nil, fmt.Errorf("clipboard tool not found: %s", clipboardTool)
	}

	return &ClipboardOutputter{
		clipboardTool: clipboardTool,
		config:        cfg,
	}, nil
}

// Copy text to the system clipboard using the configured tool
func (o *ClipboardOutputter) CopyToClipboard(text string) error {
	// Security: validate the command before execution
	if !config.IsCommandAllowed(o.config, o.clipboardTool) {
		return fmt.Errorf("clipboard tool not allowed: %s", o.clipboardTool)
	}

	var cmd *exec.Cmd
	var args []string

	switch o.clipboardTool {
	case "xsel":
		args = []string{"--clipboard", "--input"}
	case "wl-copy":
		args = []string{} // wl-copy takes no additional args for this operation
	default:
		return fmt.Errorf("unsupported clipboard tool: %s", o.clipboardTool)
	}

	// Security: sanitize arguments
	safeArgs := config.SanitizeCommandArgs(args)
	// #nosec G204 -- Safe: tool is from an allowlist and arguments are sanitized
	cmd = exec.Command(o.clipboardTool, safeArgs...)

	// Pipe the text to the command's standard input
	cmd.Stdin = strings.NewReader(text)

	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}

// Return an error as typing is not supported by this outputter
func (o *ClipboardOutputter) TypeToActiveWindow(text string) error {
	return fmt.Errorf("typing to active window not supported by clipboard outputter")
}

// Return the name of the clipboard tool being used
func (o *ClipboardOutputter) GetToolNames() (clipboardTool, typeTool string) {
	return o.clipboardTool, ""
}
