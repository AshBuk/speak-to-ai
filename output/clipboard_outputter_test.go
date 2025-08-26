// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package output

import (
	"os/exec"
	"testing"

	"github.com/AshBuk/speak-to-ai/config"
)

func TestNewClipboardOutputter(t *testing.T) {
	tests := []struct {
		name          string
		clipboardTool string
		expectError   bool
	}{
		{
			name:          "nonexistent tool",
			clipboardTool: "nonexistent-clipboard-tool",
			expectError:   true,
		},
		{
			name:          "empty tool name",
			clipboardTool: "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			outputter, err := NewClipboardOutputter(tt.clipboardTool, cfg)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && outputter == nil {
				t.Errorf("expected outputter to be created")
			}
		})
	}
}

func TestNewClipboardOutputter_WithExistingTool(t *testing.T) {
	// Find an existing command for testing
	var existingTool string
	for _, tool := range []string{"cat", "echo", "true"} {
		if _, err := exec.LookPath(tool); err == nil {
			existingTool = tool
			break
		}
	}

	if existingTool == "" {
		t.Skip("No basic tools available for testing")
	}

	cfg := &config.Config{}
	outputter, err := NewClipboardOutputter(existingTool, cfg)

	if err != nil {
		t.Errorf("unexpected error with existing tool: %v", err)
	}
	if outputter == nil {
		t.Error("expected outputter to be created with existing tool")
	}

	clipboardOutputter, ok := outputter.(*ClipboardOutputter)
	if !ok {
		t.Error("expected ClipboardOutputter type")
	}
	if clipboardOutputter.clipboardTool != existingTool {
		t.Errorf("expected clipboardTool %s, got %s", existingTool, clipboardOutputter.clipboardTool)
	}
	if clipboardOutputter.config != cfg {
		t.Error("expected config to be set correctly")
	}
}

func TestClipboardOutputter_CopyToClipboard_CommandNotAllowed(t *testing.T) {
	cfg := &config.Config{}
	// Create a mock config that doesn't allow any commands
	cfg.Security.AllowedCommands = []string{} // Empty list means no commands allowed

	outputter := &ClipboardOutputter{
		clipboardTool: "xclip",
		config:        cfg,
	}

	err := outputter.CopyToClipboard("test text")
	if err == nil {
		t.Error("expected error for disallowed command")
	}
	if err != nil && err.Error() != "clipboard tool not allowed: xclip" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestClipboardOutputter_CopyToClipboard_UnsupportedTool(t *testing.T) {
	cfg := &config.Config{}
	cfg.Security.AllowedCommands = []string{"unsupported-tool"} // Allow the tool

	outputter := &ClipboardOutputter{
		clipboardTool: "unsupported-tool",
		config:        cfg,
	}

	err := outputter.CopyToClipboard("test text")
	if err == nil {
		t.Error("expected error for unsupported tool")
	}
	if err != nil && err.Error() != "unsupported clipboard tool: unsupported-tool" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestClipboardOutputter_CopyToClipboard_SupportedTools(t *testing.T) {
	tests := []struct {
		name string
		tool string
	}{
		{
			name: "xclip",
			tool: "xclip",
		},
		{
			name: "wl-copy",
			tool: "wl-copy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.Security.AllowedCommands = []string{tt.tool}

			outputter := &ClipboardOutputter{
				clipboardTool: tt.tool,
				config:        cfg,
			}

			// This will fail since the tools don't exist, but we're testing the command building logic
			err := outputter.CopyToClipboard("test text")

			// We expect an error since the tools don't exist, but it should be a "failed to copy" error
			// not an "unsupported tool" error
			if err == nil {
				t.Error("expected error since tool doesn't exist")
			}
			if err != nil && err.Error() == "unsupported clipboard tool: "+tt.tool {
				t.Errorf("tool %s should be supported", tt.tool)
			}
		})
	}
}

func TestClipboardOutputter_TypeToActiveWindow(t *testing.T) {
	cfg := &config.Config{}
	outputter := &ClipboardOutputter{
		clipboardTool: "xclip",
		config:        cfg,
	}

	err := outputter.TypeToActiveWindow("test text")
	if err == nil {
		t.Error("expected error for unsupported operation")
	}
	expectedError := "typing to active window not supported by clipboard outputter"
	if err.Error() != expectedError {
		t.Errorf("expected error message %q, got %q", expectedError, err.Error())
	}
}

func TestClipboardOutputter_Interface(t *testing.T) {
	cfg := &config.Config{}
	outputter := &ClipboardOutputter{
		clipboardTool: "xclip",
		config:        cfg,
	}

	// Verify it implements the Outputter interface
	var _ Outputter = outputter
}
