// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package manager

import "github.com/AshBuk/speak-to-ai/internal/logger"

// mockLogger implements logger.Logger for testing
type mockLogger struct{}

func newMockLogger() logger.Logger {
	return &mockLogger{}
}

func (m *mockLogger) Info(format string, args ...interface{})    {}
func (m *mockLogger) Warning(format string, args ...interface{}) {}
func (m *mockLogger) Error(format string, args ...interface{})   {}
func (m *mockLogger) Debug(format string, args ...interface{})   {}
func (m *mockLogger) Fatal(format string, args ...interface{})   {}
