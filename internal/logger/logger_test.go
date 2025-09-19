// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package logger

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger(InfoLevel)

	if logger == nil {
		t.Fatal("NewDefaultLogger returned nil")
	}

	if logger.level != InfoLevel {
		t.Errorf("Expected level %v, got %v", InfoLevel, logger.level)
	}

	if logger.stdFlags == 0 {
		t.Error("Expected stdFlags to be set")
	}
}

func TestDefaultLogger_LogLevels(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	tests := []struct {
		name      string
		logLevel  LogLevel
		logMethod func(*DefaultLogger, string, ...interface{})
		message   string
		shouldLog bool
	}{
		{
			name:      "Debug level logs debug",
			logLevel:  DebugLevel,
			logMethod: (*DefaultLogger).Debug,
			message:   "debug message",
			shouldLog: true,
		},
		{
			name:      "Info level logs debug",
			logLevel:  InfoLevel,
			logMethod: (*DefaultLogger).Debug,
			message:   "debug message",
			shouldLog: false,
		},
		{
			name:      "Info level logs info",
			logLevel:  InfoLevel,
			logMethod: (*DefaultLogger).Info,
			message:   "info message",
			shouldLog: true,
		},
		{
			name:      "Warning level logs info",
			logLevel:  WarningLevel,
			logMethod: (*DefaultLogger).Info,
			message:   "info message",
			shouldLog: false,
		},
		{
			name:      "Warning level logs warning",
			logLevel:  WarningLevel,
			logMethod: (*DefaultLogger).Warning,
			message:   "warning message",
			shouldLog: true,
		},
		{
			name:      "Error level logs warning",
			logLevel:  ErrorLevel,
			logMethod: (*DefaultLogger).Warning,
			message:   "warning message",
			shouldLog: false,
		},
		{
			name:      "Error level logs error",
			logLevel:  ErrorLevel,
			logMethod: (*DefaultLogger).Error,
			message:   "error message",
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset buffer
			buf.Reset()

			logger := NewDefaultLogger(tt.logLevel)
			tt.logMethod(logger, tt.message)

			output := buf.String()

			if tt.shouldLog {
				if output == "" {
					t.Error("Expected log output but got none")
				}
				if !strings.Contains(output, tt.message) {
					t.Errorf("Expected log output to contain %q, got %q", tt.message, output)
				}
			} else {
				if output != "" {
					t.Errorf("Expected no log output but got %q", output)
				}
			}
		})
	}
}

func TestDefaultLogger_LogFormatting(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewDefaultLogger(DebugLevel)

	tests := []struct {
		name         string
		logMethod    func(*DefaultLogger, string, ...interface{})
		format       string
		args         []interface{}
		expectedText string
		prefix       string
	}{
		{
			name:         "Debug with formatting",
			logMethod:    (*DefaultLogger).Debug,
			format:       "User %s has %d items",
			args:         []interface{}{"John", 5},
			expectedText: "User John has 5 items",
			prefix:       "[DEBUG]",
		},
		{
			name:         "Info with formatting",
			logMethod:    (*DefaultLogger).Info,
			format:       "Processing file %s",
			args:         []interface{}{"test.txt"},
			expectedText: "Processing file test.txt",
			prefix:       "[INFO]",
		},
		{
			name:         "Warning with formatting",
			logMethod:    (*DefaultLogger).Warning,
			format:       "Low disk space: %d%% remaining",
			args:         []interface{}{15},
			expectedText: "Low disk space: 15% remaining",
			prefix:       "[WARNING]",
		},
		{
			name:         "Error with formatting",
			logMethod:    (*DefaultLogger).Error,
			format:       "Failed to connect to %s:%d",
			args:         []interface{}{"localhost", 8080},
			expectedText: "Failed to connect to localhost:8080",
			prefix:       "[ERROR]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset buffer
			buf.Reset()

			tt.logMethod(logger, tt.format, tt.args...)

			output := buf.String()

			if !strings.Contains(output, tt.expectedText) {
				t.Errorf("Expected log output to contain %q, got %q", tt.expectedText, output)
			}

			if !strings.Contains(output, tt.prefix) {
				t.Errorf("Expected log output to contain prefix %q, got %q", tt.prefix, output)
			}
		})
	}
}

func TestConfigure(t *testing.T) {
	config := Config{
		Level: InfoLevel,
	}

	logger, err := Configure(config)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be returned")
	}

	if logger.level != InfoLevel {
		t.Errorf("Expected level %v, got %v", InfoLevel, logger.level)
	}
}

func TestLogLevelConstants(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected int
	}{
		{
			name:     "DebugLevel",
			level:    DebugLevel,
			expected: 0,
		},
		{
			name:     "InfoLevel",
			level:    InfoLevel,
			expected: 1,
		},
		{
			name:     "WarningLevel",
			level:    WarningLevel,
			expected: 2,
		},
		{
			name:     "ErrorLevel",
			level:    ErrorLevel,
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.level) != tt.expected {
				t.Errorf("Expected %s to have value %d, got %d", tt.name, tt.expected, int(tt.level))
			}
		})
	}
}

func TestLogLevel_Ordering(t *testing.T) {
	// Test that log levels are properly ordered
	if DebugLevel >= InfoLevel {
		t.Error("DebugLevel should be less than InfoLevel")
	}

	if InfoLevel >= WarningLevel {
		t.Error("InfoLevel should be less than WarningLevel")
	}

	if WarningLevel >= ErrorLevel {
		t.Error("WarningLevel should be less than ErrorLevel")
	}
}

func TestDefaultLogger_Interface(t *testing.T) {
	// Test that DefaultLogger implements Logger interface
	var logger Logger = NewDefaultLogger(InfoLevel)

	// Test all interface methods
	logger.Debug("debug test")
	logger.Info("info test")
	logger.Warning("warning test")
	logger.Error("error test")
}

func TestDefaultLogger_ConcurrentAccess(t *testing.T) {
	// Test concurrent access to logger
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewDefaultLogger(InfoLevel)

	// Run multiple goroutines writing to logger
	const numGoroutines = 10
	const messagesPerGoroutine = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info("goroutine %d message %d", id, j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check that we got expected number of log messages
	output := buf.String()
	messageCount := strings.Count(output, "[INFO]")

	expectedMessages := numGoroutines * messagesPerGoroutine
	if messageCount != expectedMessages {
		t.Errorf("Expected %d log messages, got %d", expectedMessages, messageCount)
	}
}

func TestDefaultLogger_EmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewDefaultLogger(InfoLevel)

	logger.Info("")

	output := buf.String()

	if !strings.Contains(output, "[INFO]") {
		t.Error("Expected log output to contain [INFO] prefix even for empty message")
	}
}

func TestDefaultLogger_NilArgs(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewDefaultLogger(InfoLevel)

	// Test with nil args
	logger.Info("test message", nil)

	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Error("Expected log output to contain message even with nil args")
	}

	if !strings.Contains(output, "[INFO]") {
		t.Error("Expected log output to contain [INFO] prefix")
	}
}
