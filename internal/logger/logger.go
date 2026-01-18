// Package logger provides structured logging functionality for the MTC application.
// It wraps the standard library's slog package to provide a simple, consistent logging interface
// with support for multiple log levels (debug, info, warn, error) and output formats (text, JSON).
package logger

import (
	"io"
	"log/slog"
	"os"
)

var (
	// defaultLogger is the default logger instance used throughout the application.
	defaultLogger *slog.Logger

	// logLevel is the current log level threshold.
	// Messages below this level will not be logged.
	logLevel slog.Level = slog.LevelInfo
)

// Init initializes the logger with the specified level and format.
// If format is "json", logs will be in JSON format; otherwise, human-readable text.
// If output is nil, os.Stderr is used.
func Init(level string, format string, output io.Writer) {
	if output == nil {
		output = os.Stderr
	}

	// Parse log level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	if format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	defaultLogger = slog.New(handler)
}

// Logger returns the default logger instance.
// If defaultLogger is nil, it initializes a new logger with default settings
// (info level, text format, stderr output). In tests, the logger should be
// initialized via init() functions in test files to avoid unwanted output.
//
// Returns the default logger instance.
func Logger() *slog.Logger {
	if defaultLogger == nil {
		// Initialize with defaults if not already initialized
		// In tests, this should be initialized via init() functions in test files
		Init("info", "text", nil)
	}
	return defaultLogger
}

// Debug logs a debug message with optional key-value pairs.
// The message is only logged if the current log level is debug or lower.
//
// Parameters:
//   - msg: The log message
//   - args: Optional key-value pairs for structured logging (e.g., "key", value)
func Debug(msg string, args ...any) {
	Logger().Debug(msg, args...)
}

// Info logs an info message with optional key-value pairs.
// The message is only logged if the current log level is info or lower.
//
// Parameters:
//   - msg: The log message
//   - args: Optional key-value pairs for structured logging (e.g., "key", value)
func Info(msg string, args ...any) {
	Logger().Info(msg, args...)
}

// Warn logs a warning message with optional key-value pairs.
// The message is only logged if the current log level is warn or lower.
//
// Parameters:
//   - msg: The log message
//   - args: Optional key-value pairs for structured logging (e.g., "key", value)
func Warn(msg string, args ...any) {
	Logger().Warn(msg, args...)
}

// Error logs an error message with optional key-value pairs.
// Error messages are always logged regardless of log level.
//
// Parameters:
//   - msg: The log message
//   - args: Optional key-value pairs for structured logging (e.g., "key", value)
func Error(msg string, args ...any) {
	Logger().Error(msg, args...)
}

// With returns a logger with the given key-value pairs added to its context.
// All subsequent log messages from the returned logger will include these
// key-value pairs. This is useful for adding contextual information like
// operation names, request IDs, or file paths.
//
// Parameters:
//   - args: Key-value pairs to add to the logger context (e.g., "path", "/tmp/file")
//
// Returns a new logger instance with the context added.
func With(args ...any) *slog.Logger {
	return Logger().With(args...)
}
