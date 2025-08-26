// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package logger

import (
	"log"
	"os"
)

// LogLevel represents the level of logging
type LogLevel int

const (
	// Debug log level
	DebugLevel LogLevel = iota
	// Info log level
	InfoLevel
	// Warning log level
	WarningLevel
	// Error log level
	ErrorLevel
)

// Logger interface defines methods for logging at different levels
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// Config contains logger configuration
type Config struct {
	Level LogLevel
	File  string
}

// DefaultLogger implements the Logger interface using the standard log package
type DefaultLogger struct {
	level    LogLevel
	stdFlags int
}

// NewDefaultLogger creates a new default logger with the specified log level
func NewDefaultLogger(level LogLevel) *DefaultLogger {
	return &DefaultLogger{
		level:    level,
		stdFlags: log.LstdFlags | log.Lshortfile,
	}
}

// Configure sets up the logger with given configuration
func Configure(config Config) (*DefaultLogger, error) {
	logger := NewDefaultLogger(config.Level)
	log.SetFlags(logger.stdFlags)

	// If log file is specified, set up file logging
	if config.File != "" {
		f, err := os.OpenFile(config.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		log.SetOutput(f)
	}

	return logger, nil
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	if l.level <= DebugLevel {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// Info logs an informational message
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	if l.level <= InfoLevel {
		log.Printf("[INFO] "+format, args...)
	}
}

// Warning logs a warning message
func (l *DefaultLogger) Warning(format string, args ...interface{}) {
	if l.level <= WarningLevel {
		log.Printf("[WARNING] "+format, args...)
	}
}

// Error logs an error message
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	if l.level <= ErrorLevel {
		log.Printf("[ERROR] "+format, args...)
	}
}
