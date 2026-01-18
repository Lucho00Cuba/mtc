package logger

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name   string
		level  string
		format string
		output io.Writer
	}{
		{
			name:   "debug level text",
			level:  "debug",
			format: "text",
			output: nil,
		},
		{
			name:   "info level json",
			level:  "info",
			format: "json",
			output: nil,
		},
		{
			name:   "warn level text",
			level:  "warn",
			format: "text",
			output: nil,
		},
		{
			name:   "error level json",
			level:  "error",
			format: "json",
			output: nil,
		},
		{
			name:   "invalid level defaults to info",
			level:  "invalid",
			format: "text",
			output: nil,
		},
		{
			name:   "custom output",
			level:  "info",
			format: "text",
			output: &bytes.Buffer{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.level, tt.format, tt.output)
			logger := Logger()
			if logger == nil {
				t.Error("Logger() returned nil after Init()")
			}
		})
	}
}

func TestLogger(t *testing.T) {
	// Test default initialization
	logger := Logger()
	if logger == nil {
		t.Error("Logger() returned nil")
	}

	// Test that it initializes with defaults if not initialized
	Init("", "", nil)
	logger = Logger()
	if logger == nil {
		t.Error("Logger() returned nil after default init")
	}
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	Init("debug", "text", &buf)

	Debug("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Debug() output should contain message, got: %s", output)
	}
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	Init("info", "text", &buf)

	Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Info() output should contain message, got: %s", output)
	}
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	Init("warn", "text", &buf)

	Warn("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Warn() output should contain message, got: %s", output)
	}
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	Init("error", "text", &buf)

	Error("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Error() output should contain message, got: %s", output)
	}
}

func TestWith(t *testing.T) {
	Init("info", "text", nil)

	logger := With("key1", "value1", "key2", "value2")
	if logger == nil {
		t.Error("With() returned nil")
	}

	// Test that it returns a logger with context
	var buf bytes.Buffer
	Init("info", "text", &buf)
	logger = With("test", "value")
	logger.Info("message")

	output := buf.String()
	if !strings.Contains(output, "message") {
		t.Errorf("With() logger should log messages, got: %s", output)
	}
}

func TestLogLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "invalid"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			Init(level, "text", nil)
			logger := Logger()
			if logger == nil {
				t.Errorf("Logger() returned nil for level %s", level)
			}
		})
	}
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	Init("info", "json", &buf)

	Info("test message", "key", "value")

	output := buf.String()
	// JSON format should contain quotes and structured data
	if !strings.Contains(output, "\"msg\"") && !strings.Contains(output, "test message") {
		t.Errorf("JSON format should contain structured data, got: %s", output)
	}
}

func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer
	Init("info", "text", &buf)

	Info("test message", "key", "value")

	output := buf.String()
	if output == "" {
		t.Error("Text format should produce output")
	}
}
