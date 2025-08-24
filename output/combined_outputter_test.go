package output

import (
	"testing"

	"github.com/AshBuk/speak-to-ai/config"
)

func TestNewCombinedOutputter(t *testing.T) {
	tests := []struct {
		name          string
		clipboardTool string
		typeTool      string
		expectError   bool
	}{
		{
			name:          "both tools nonexistent",
			clipboardTool: "nonexistent-clipboard",
			typeTool:      "nonexistent-type",
			expectError:   true,
		},
		{
			name:          "clipboard tool nonexistent",
			clipboardTool: "nonexistent-clipboard",
			typeTool:      "echo", // echo should exist
			expectError:   true,
		},
		{
			name:          "type tool nonexistent",
			clipboardTool: "echo", // echo should exist
			typeTool:      "nonexistent-type",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			outputter, err := NewCombinedOutputter(tt.clipboardTool, tt.typeTool, cfg)

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

func TestNewCombinedOutputter_WithExistingTools(t *testing.T) {
	// Use simple existing commands for testing
	cfg := &config.Config{}
	outputter, err := NewCombinedOutputter("echo", "echo", cfg)

	if err != nil {
		t.Errorf("unexpected error with existing tools: %v", err)
	}
	if outputter == nil {
		t.Error("expected outputter to be created with existing tools")
	}

	combinedOutputter, ok := outputter.(*CombinedOutputter)
	if !ok {
		t.Error("expected CombinedOutputter type")
	}
	if combinedOutputter.clipboardOutputter == nil {
		t.Error("expected clipboard outputter to be initialized")
	}
	if combinedOutputter.typeOutputter == nil {
		t.Error("expected type outputter to be initialized")
	}
}

func TestCombinedOutputter_CopyToClipboard(t *testing.T) {
	// Create mock outputters
	mockClipboard := NewMockOutputter()
	mockType := NewMockOutputter()

	outputter := &CombinedOutputter{
		clipboardOutputter: mockClipboard,
		typeOutputter:      mockType,
	}

	err := outputter.CopyToClipboard("test text")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify that clipboard outputter was called
	if !mockClipboard.WasClipboardCalled() {
		t.Error("expected clipboard outputter CopyToClipboard to be called")
	}
	if mockClipboard.GetLastClipboardCall() != "test text" {
		t.Errorf("expected clipboard text 'test text', got '%s'", mockClipboard.GetLastClipboardCall())
	}

	// Verify that type outputter was NOT called for clipboard
	if mockType.WasClipboardCalled() {
		t.Error("expected type outputter CopyToClipboard NOT to be called")
	}
}

func TestCombinedOutputter_TypeToActiveWindow(t *testing.T) {
	// Create mock outputters
	mockClipboard := NewMockOutputter()
	mockType := NewMockOutputter()

	outputter := &CombinedOutputter{
		clipboardOutputter: mockClipboard,
		typeOutputter:      mockType,
	}

	err := outputter.TypeToActiveWindow("test text")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify that type outputter was called
	if !mockType.WasTypeCalled() {
		t.Error("expected type outputter TypeToActiveWindow to be called")
	}
	if mockType.GetLastTypeCall() != "test text" {
		t.Errorf("expected type text 'test text', got '%s'", mockType.GetLastTypeCall())
	}

	// Verify that clipboard outputter was NOT called for typing
	if mockClipboard.WasTypeCalled() {
		t.Error("expected clipboard outputter TypeToActiveWindow NOT to be called")
	}
}

func TestCombinedOutputter_Interface(t *testing.T) {
	mockClipboard := NewMockOutputter()
	mockType := NewMockOutputter()

	outputter := &CombinedOutputter{
		clipboardOutputter: mockClipboard,
		typeOutputter:      mockType,
	}

	// Verify it implements the Outputter interface
	var _ Outputter = outputter
}
