// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package logger

import (
	"log"
)

// LogLevel represents the severity level for log messages
// Lower values are more verbose (DebugLevel=0), higher values filter more (ErrorLevel=3)
type LogLevel int

const (
	DebugLevel   LogLevel = iota // Most verbose - development diagnostics
	InfoLevel                    // Normal operations and state changes
	WarningLevel                 // Potential issues or degraded functionality
	ErrorLevel                   // Critical errors requiring attention
)

// Logger defines the contract for structured logging across the application
// All logging must go through this interface to maintain consistency
// Usage pattern:
//   - Components accept logger via dependency injection (variadic ...logger.Logger)
//   - Default fallback: logger.NewDefaultLogger(logger.WarningLevel)
//   - Production: injected logger with appropriate level for context
type Logger interface {
	Debug(format string, args ...interface{})   // Detailed debugging information
	Info(format string, args ...interface{})    // General informational messages
	Warning(format string, args ...interface{}) // Warning messages about potential issues
	Error(format string, args ...interface{})   // Error messages about failures
}

// Config holds logger initialization settings
type Config struct {
	Level LogLevel // min level
}

// Thread-safe implementation wrapping log.Printf with level filtering
type DefaultLogger struct {
	level    LogLevel // min level
	stdFlags int      // formatting flags for log output
}

// NewDefaultLogger Constructor - creates logger with sensible defaults
// Output format: timestamp + file:line + [LEVEL] + message
func NewDefaultLogger(level LogLevel) *DefaultLogger {
	return &DefaultLogger{
		level:    level,
		stdFlags: log.LstdFlags | log.Lshortfile,
	}
}

// Configure Factory Method - creates and configures logger from Config struct
// Sets global log flags for consistent formatting across application
func Configure(config Config) (*DefaultLogger, error) {
	logger := NewDefaultLogger(config.Level)
	log.SetFlags(logger.stdFlags)
	return logger, nil
}

// Debug logs detailed diagnostic information
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	if l.level <= DebugLevel {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// Info logs general operational messages
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	if l.level <= InfoLevel {
		log.Printf("[INFO] "+format, args...)
	}
}

// Warning logs potential issues or degraded functionality
func (l *DefaultLogger) Warning(format string, args ...interface{}) {
	if l.level <= WarningLevel {
		log.Printf("[WARNING] "+format, args...)
	}
}

// Error logs critical failures requiring attention
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	if l.level <= ErrorLevel {
		log.Printf("[ERROR] "+format, args...)
	}
}
