package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// DEBUG level for detailed debug information
	DEBUG LogLevel = iota
	// INFO level for general information
	INFO
	// WARNING level for warnings
	WARNING
	// ERROR level for errors
	ERROR
	// FATAL level for fatal errors
	FATAL
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger is a structured logger for the application
type Logger struct {
	level      LogLevel
	logger     *log.Logger
	prefix     string
	logFile    *os.File
	fileLogger *log.Logger
	useFile    bool
}

// NewLogger creates a new logger instance
func NewLogger(prefix string, level LogLevel) *Logger {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	return &Logger{
		level:   level,
		logger:  logger,
		prefix:  prefix,
		useFile: false,
	}
}

// SetLogFile sets a file for logging in addition to stdout
func (l *Logger) SetLogFile(logFilePath string) error {
	// Ensure directory exists
	dir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Close any existing log file
	if l.logFile != nil {
		l.logFile.Close()
	}

	l.logFile = f
	l.fileLogger = log.New(f, "", log.LstdFlags)
	l.useFile = true

	return nil
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// log logs a message at the specified level
func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	if level < l.level {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Use base filename only
	file = filepath.Base(file)

	// Format message
	message := fmt.Sprintf(format, v...)
	logMessage := fmt.Sprintf("[%s] [%s] [%s:%d] %s", level.String(), l.prefix, file, line, message)

	// Log to stdout
	l.logger.Println(logMessage)

	// Log to file if enabled
	if l.useFile && l.fileLogger != nil {
		l.fileLogger.Println(logMessage)
	}

	// Exit on fatal errors
	if level == FATAL {
		if l.logFile != nil {
			l.logFile.Close()
		}
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DEBUG, format, v...)
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(INFO, format, v...)
}

// Warning logs a warning message
func (l *Logger) Warning(format string, v ...interface{}) {
	l.log(WARNING, format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ERROR, format, v...)
}

// Fatal logs a fatal error message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(FATAL, format, v...)
}

// Close closes the logger and any open files
func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
		l.logFile = nil
	}
}

// MultiWriter returns a writer that duplicates writes to stdout and the log file
func (l *Logger) MultiWriter() io.Writer {
	if l.useFile && l.logFile != nil {
		return io.MultiWriter(os.Stdout, l.logFile)
	}
	return os.Stdout
}

// GlobalLogger is the default logger instance
var GlobalLogger = NewLogger("DAEMON", INFO)

// ConfigureLogger configures the global logger based on application settings
func ConfigureLogger(config *Config) error {
	// Set log level based on debug flag
	if config.General.Debug {
		GlobalLogger.SetLevel(DEBUG)
	} else {
		GlobalLogger.SetLevel(INFO)
	}

	// Setup log file if specified
	if config.General.LogFile != "" {
		if err := GlobalLogger.SetLogFile(config.General.LogFile); err != nil {
			return err
		}
	}

	return nil
}

// Default logging functions that use the global logger

// Debug logs a debug message using the global logger
func Debug(format string, v ...interface{}) {
	GlobalLogger.Debug(format, v...)
}

// Info logs an info message using the global logger
func Info(format string, v ...interface{}) {
	GlobalLogger.Info(format, v...)
}

// Warning logs a warning message using the global logger
func Warning(format string, v ...interface{}) {
	GlobalLogger.Warning(format, v...)
}

// Error logs an error message using the global logger
func Error(format string, v ...interface{}) {
	GlobalLogger.Error(format, v...)
}

// Fatal logs a fatal error message and exits using the global logger
func Fatal(format string, v ...interface{}) {
	GlobalLogger.Fatal(format, v...)
}
