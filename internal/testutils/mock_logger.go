// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package testutils

import (
	"fmt"
	"sync"
)

// MockLogger implements Logger interface for testing
type MockLogger struct {
	mu       sync.Mutex
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
	m.mu.Lock()
	m.messages = append(m.messages, fmt.Sprintf("[DEBUG] "+format, args...))
	m.mu.Unlock()
}

// Info logs an info message
func (m *MockLogger) Info(format string, args ...interface{}) {
	m.mu.Lock()
	m.messages = append(m.messages, fmt.Sprintf("[INFO] "+format, args...))
	m.mu.Unlock()
}

// Warning logs a warning message
func (m *MockLogger) Warning(format string, args ...interface{}) {
	m.mu.Lock()
	m.messages = append(m.messages, fmt.Sprintf("[WARNING] "+format, args...))
	m.mu.Unlock()
}

// Error logs an error message
func (m *MockLogger) Error(format string, args ...interface{}) {
	m.mu.Lock()
	m.messages = append(m.messages, fmt.Sprintf("[ERROR] "+format, args...))
	m.mu.Unlock()
}

// Fatal logs a fatal message
func (m *MockLogger) Fatal(format string, args ...interface{}) {
	m.mu.Lock()
	m.messages = append(m.messages, fmt.Sprintf("[FATAL] "+format, args...))
	m.mu.Unlock()
}

// GetMessages returns all logged messages
func (m *MockLogger) GetMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]string, len(m.messages))
	copy(copied, m.messages)
	return copied
}

// Clear clears all logged messages
func (m *MockLogger) Clear() {
	m.mu.Lock()
	m.messages = m.messages[:0]
	m.mu.Unlock()
}
