// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package outputters

import (
	"os"
	"os/exec"
	"slices"
	"testing"

	"github.com/AshBuk/dabri/v2/config"
	"github.com/AshBuk/dabri/v2/output/interfaces"
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
		clipboardTool: "xsel",
		config:        cfg,
	}
	err := outputter.CopyToClipboard("test text")
	if err == nil {
		t.Error("expected error for disallowed command")
	}
	if err != nil && err.Error() != "clipboard tool not allowed: xsel" {
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
		tool         string
		expectedArgs []string
	}{
		{tool: "xsel", expectedArgs: []string{"--clipboard", "--input"}},
		{tool: "wl-copy", expectedArgs: []string{""}},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			// A stub keeps the developer's real clipboard untouched
			captureFile := installFakeTool(t, tt.tool)
			cfg := &config.Config{}
			cfg.Security.AllowedCommands = []string{tt.tool}

			outputter := &ClipboardOutputter{
				clipboardTool: tt.tool,
				config:        cfg,
			}
			if err := outputter.CopyToClipboard("test text"); err != nil {
				t.Fatalf("expected fake %s to run successfully, got %v", tt.tool, err)
			}
			if args := readCapturedArgs(t, captureFile); !slices.Equal(args, tt.expectedArgs) {
				t.Errorf("expected args %q, got %q", tt.expectedArgs, args)
			}
			if stdin, err := os.ReadFile(captureFile + ".stdin"); err != nil || string(stdin) != "test text" {
				t.Errorf("expected text on stdin, got %q (%v)", stdin, err)
			}
		})
	}
}

func TestClipboardOutputter_TypeToActiveWindow(t *testing.T) {
	cfg := &config.Config{}
	outputter := &ClipboardOutputter{
		clipboardTool: "xsel",
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
	// Compile-time assertion: ClipboardOutputter implements interfaces.Outputter
	var _ interfaces.Outputter = (*ClipboardOutputter)(nil)
}
