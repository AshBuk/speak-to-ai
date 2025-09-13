// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package outputters

import (
	"errors"
	"strings"
)

// MockOutputter implements interfaces.Outputter interface for testing
type MockOutputter struct {
	clipboardContent     string
	typedContent         string
	clipboardError       error
	typeError            error
	clipboardCallCount   int
	typeCallCount        int
	clipboardCallHistory []string
	typeCallHistory      []string
}

// NewMockOutputter creates a new mock outputter
func NewMockOutputter() *MockOutputter {
	return &MockOutputter{
		clipboardCallHistory: make([]string, 0),
		typeCallHistory:      make([]string, 0),
	}
}

// CopyToClipboard simulates copying text to clipboard
func (m *MockOutputter) CopyToClipboard(text string) error {
	if m.clipboardError != nil {
		return m.clipboardError
	}

	m.clipboardContent = text
	m.clipboardCallCount++
	m.clipboardCallHistory = append(m.clipboardCallHistory, text)

	return nil
}

// TypeToActiveWindow simulates typing text to active window
func (m *MockOutputter) TypeToActiveWindow(text string) error {
	if m.typeError != nil {
		return m.typeError
	}

	m.typedContent = text
	m.typeCallCount++
	m.typeCallHistory = append(m.typeCallHistory, text)

	return nil
}

// GetToolNames returns mock tool names
func (m *MockOutputter) GetToolNames() (clipboardTool, typeTool string) {
	return "mock-clipboard", "mock-type"
}

// Test helper methods

// SetClipboardError configures the mock to return an error on CopyToClipboard
func (m *MockOutputter) SetClipboardError(err error) {
	m.clipboardError = err
}

// SetTypeError configures the mock to return an error on TypeToActiveWindow
func (m *MockOutputter) SetTypeError(err error) {
	m.typeError = err
}

// GetClipboardContent returns the last text copied to clipboard
func (m *MockOutputter) GetClipboardContent() string {
	return m.clipboardContent
}

// GetTypedContent returns the last text typed to active window
func (m *MockOutputter) GetTypedContent() string {
	return m.typedContent
}

// GetClipboardCallCount returns the number of times CopyToClipboard was called
func (m *MockOutputter) GetClipboardCallCount() int {
	return m.clipboardCallCount
}

// GetTypeCallCount returns the number of times TypeToActiveWindow was called
func (m *MockOutputter) GetTypeCallCount() int {
	return m.typeCallCount
}

// GetClipboardCallHistory returns all texts copied to clipboard
func (m *MockOutputter) GetClipboardCallHistory() []string {
	return m.clipboardCallHistory
}

// GetTypeCallHistory returns all texts typed to active window
func (m *MockOutputter) GetTypeCallHistory() []string {
	return m.typeCallHistory
}

// Reset clears all mock state
func (m *MockOutputter) Reset() {
	m.clipboardContent = ""
	m.typedContent = ""
	m.clipboardError = nil
	m.typeError = nil
	m.clipboardCallCount = 0
	m.typeCallCount = 0
	m.clipboardCallHistory = make([]string, 0)
	m.typeCallHistory = make([]string, 0)
}

// WasClipboardCalled returns true if CopyToClipboard was called
func (m *MockOutputter) WasClipboardCalled() bool {
	return m.clipboardCallCount > 0
}

// WasTypeCalled returns true if TypeToActiveWindow was called
func (m *MockOutputter) WasTypeCalled() bool {
	return m.typeCallCount > 0
}

// GetLastClipboardCall returns the last text copied to clipboard
func (m *MockOutputter) GetLastClipboardCall() string {
	if len(m.clipboardCallHistory) == 0 {
		return ""
	}
	return m.clipboardCallHistory[len(m.clipboardCallHistory)-1]
}

// GetLastTypeCall returns the last text typed to active window
func (m *MockOutputter) GetLastTypeCall() string {
	if len(m.typeCallHistory) == 0 {
		return ""
	}
	return m.typeCallHistory[len(m.typeCallHistory)-1]
}

// ContainsClipboardText checks if any clipboard call contained the given text
func (m *MockOutputter) ContainsClipboardText(text string) bool {
	for _, call := range m.clipboardCallHistory {
		if strings.Contains(call, text) {
			return true
		}
	}
	return false
}

// ContainsTypeText checks if any type call contained the given text
func (m *MockOutputter) ContainsTypeText(text string) bool {
	for _, call := range m.typeCallHistory {
		if strings.Contains(call, text) {
			return true
		}
	}
	return false
}

// MockOutputterWithErrors provides pre-configured error scenarios
type MockOutputterWithErrors struct {
	*MockOutputter
}

// NewMockOutputterWithErrors creates a mock outputter with common error scenarios
func NewMockOutputterWithErrors() *MockOutputterWithErrors {
	return &MockOutputterWithErrors{
		MockOutputter: NewMockOutputter(),
	}
}

// SimulateClipboardUnavailable simulates clipboard service unavailable
func (m *MockOutputterWithErrors) SimulateClipboardUnavailable() {
	m.SetClipboardError(errors.New("clipboard service unavailable"))
}

// SimulatePermissionDenied simulates permission denied error
func (m *MockOutputterWithErrors) SimulatePermissionDenied() {
	m.SetTypeError(errors.New("permission denied: cannot access active window"))
}

// SimulateTimeout simulates operation timeout
func (m *MockOutputterWithErrors) SimulateTimeout() {
	m.SetClipboardError(errors.New("operation timed out"))
	m.SetTypeError(errors.New("operation timed out"))
}

// SimulateInvalidInput simulates invalid input error
func (m *MockOutputterWithErrors) SimulateInvalidInput() {
	m.SetClipboardError(errors.New("invalid input: text is empty"))
	m.SetTypeError(errors.New("invalid input: text is empty"))
}
