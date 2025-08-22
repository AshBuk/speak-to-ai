package output

import (
	"os/exec"
	"testing"

	"github.com/AshBuk/speak-to-ai/config"
)

func TestNewTypeOutputter(t *testing.T) {
	tests := []struct {
		name        string
		typeTool    string
		expectError bool
	}{
		{
			name:        "nonexistent tool",
			typeTool:    "nonexistent-type-tool",
			expectError: true,
		},
		{
			name:        "empty tool name",
			typeTool:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			outputter, err := NewTypeOutputter(tt.typeTool, cfg)

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

func TestNewTypeOutputter_WithExistingTool(t *testing.T) {
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
	outputter, err := NewTypeOutputter(existingTool, cfg)

	if err != nil {
		t.Errorf("unexpected error with existing tool: %v", err)
	}
	if outputter == nil {
		t.Error("expected outputter to be created with existing tool")
	}

	typeOutputter, ok := outputter.(*TypeOutputter)
	if !ok {
		t.Error("expected TypeOutputter type")
	}
	if typeOutputter.typeTool != existingTool {
		t.Errorf("expected typeTool %s, got %s", existingTool, typeOutputter.typeTool)
	}
	if typeOutputter.config != cfg {
		t.Error("expected config to be set correctly")
	}
}

func TestTypeOutputter_TypeToActiveWindow_CommandNotAllowed(t *testing.T) {
	cfg := &config.Config{}
	// Create a mock config that doesn't allow any commands
	cfg.Security.AllowedCommands = []string{} // Empty list means no commands allowed

	outputter := &TypeOutputter{
		typeTool: "xdotool",
		config:   cfg,
	}

	err := outputter.TypeToActiveWindow("test text")
	if err == nil {
		t.Error("expected error for disallowed command")
	}
	if err != nil && err.Error() != "typing tool not allowed: xdotool" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTypeOutputter_TypeToActiveWindow_UnsupportedTool(t *testing.T) {
	cfg := &config.Config{}
	cfg.Security.AllowedCommands = []string{"unsupported-tool"} // Allow the tool

	outputter := &TypeOutputter{
		typeTool: "unsupported-tool",
		config:   cfg,
	}

	err := outputter.TypeToActiveWindow("test text")
	if err == nil {
		t.Error("expected error for unsupported tool")
	}
	if err != nil && err.Error() != "unsupported typing tool: unsupported-tool" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTypeOutputter_TypeToActiveWindow_SupportedTools(t *testing.T) {
	tests := []struct {
		name         string
		tool         string
		expectedArgs []string
	}{
		{
			name: "xdotool",
			tool: "xdotool",
		},
		{
			name: "wtype",
			tool: "wtype",
		},
		{
			name: "ydotool",
			tool: "ydotool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.Security.AllowedCommands = []string{tt.tool}

			outputter := &TypeOutputter{
				typeTool: tt.tool,
				config:   cfg,
			}

			// This will fail since the tools don't exist, but we're testing the command building logic
			err := outputter.TypeToActiveWindow("test text")

			// We expect an error since the tools don't exist, but it should be a "failed to type" error
			// not an "unsupported tool" error
			if err == nil {
				t.Error("expected error since tool doesn't exist")
			}
			if err != nil && err.Error() == "unsupported typing tool: "+tt.tool {
				t.Errorf("tool %s should be supported", tt.tool)
			}
		})
	}
}

func TestTypeOutputter_CopyToClipboard(t *testing.T) {
	cfg := &config.Config{}
	outputter := &TypeOutputter{
		typeTool: "xdotool",
		config:   cfg,
	}

	err := outputter.CopyToClipboard("test text")
	if err == nil {
		t.Error("expected error for unsupported operation")
	}
	expectedError := "copying to clipboard not supported by type outputter"
	if err.Error() != expectedError {
		t.Errorf("expected error message %q, got %q", expectedError, err.Error())
	}
}

func TestTypeOutputter_Interface(t *testing.T) {
	cfg := &config.Config{}
	outputter := &TypeOutputter{
		typeTool: "xdotool",
		config:   cfg,
	}

	// Verify it implements the Outputter interface
	var _ Outputter = outputter
}
