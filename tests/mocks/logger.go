// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package mocks

import "fmt"

// MockLogger implements Logger interface for testing
type MockLogger struct {
	messages []string
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		messages: make([]string, 0),
	}
}

// Debug logs a debug message
func (m *MockLogger) Debug(format string, args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf("[DEBUG] "+format, args...))
}

// Info logs an info message
func (m *MockLogger) Info(format string, args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf("[INFO] "+format, args...))
}

// Warning logs a warning message
func (m *MockLogger) Warning(format string, args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf("[WARNING] "+format, args...))
}

// Error logs an error message
func (m *MockLogger) Error(format string, args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf("[ERROR] "+format, args...))
}

// GetMessages returns all logged messages
func (m *MockLogger) GetMessages() []string {
	return m.messages
}

// Clear clears all logged messages
func (m *MockLogger) Clear() {
	m.messages = m.messages[:0]
}
