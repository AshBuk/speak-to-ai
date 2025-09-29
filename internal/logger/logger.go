// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package logger

import (
	"log"
)

// Represents the level of logging
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
)

// Defines the contract for a logger that supports different log levels
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// Contains logger configuration
type Config struct {
	Level LogLevel
}

// Implements the Logger interface using the standard log package
type DefaultLogger struct {
	level    LogLevel
	stdFlags int
}

// Create a new default logger with the specified log level
func NewDefaultLogger(level LogLevel) *DefaultLogger {
	return &DefaultLogger{
		level:    level,
		stdFlags: log.LstdFlags | log.Lshortfile,
	}
}

// Set up the logger with the given configuration
func Configure(config Config) (*DefaultLogger, error) {
	logger := NewDefaultLogger(config.Level)
	log.SetFlags(logger.stdFlags)
	return logger, nil
}

// Log a debug message
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	if l.level <= DebugLevel {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// Log an informational message
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	if l.level <= InfoLevel {
		log.Printf("[INFO] "+format, args...)
	}
}

// Log a warning message
func (l *DefaultLogger) Warning(format string, args ...interface{}) {
	if l.level <= WarningLevel {
		log.Printf("[WARNING] "+format, args...)
	}
}

// Log an error message
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	if l.level <= ErrorLevel {
		log.Printf("[ERROR] "+format, args...)
	}
}
