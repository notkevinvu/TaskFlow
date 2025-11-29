package logger

import (
	"log/slog"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    LogLevel
		expected slog.Level
	}{
		{
			name:     "debug level",
			input:    LevelDebug,
			expected: slog.LevelDebug,
		},
		{
			name:     "info level",
			input:    LevelInfo,
			expected: slog.LevelInfo,
		},
		{
			name:     "warn level",
			input:    LevelWarn,
			expected: slog.LevelWarn,
		},
		{
			name:     "error level",
			input:    LevelError,
			expected: slog.LevelError,
		},
		{
			name:     "uppercase debug",
			input:    LogLevel("DEBUG"),
			expected: slog.LevelDebug,
		},
		{
			name:     "mixed case info",
			input:    LogLevel("InFo"),
			expected: slog.LevelInfo,
		},
		{
			name:     "invalid level defaults to info",
			input:    LogLevel("invalid"),
			expected: slog.LevelInfo,
		},
		{
			name:     "empty string defaults to info",
			input:    LogLevel(""),
			expected: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "text format with debug level",
			config: Config{
				Level:  LevelDebug,
				Format: "text",
			},
		},
		{
			name: "json format with info level",
			config: Config{
				Level:  LevelInfo,
				Format: "json",
			},
		},
		{
			name: "text format with error level",
			config: Config{
				Level:  LevelError,
				Format: "text",
			},
		},
		{
			name: "json format with warn level",
			config: Config{
				Level:  LevelWarn,
				Format: "json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.config)
			if logger == nil {
				t.Error("New() returned nil logger")
			}
		})
	}
}

func TestNewWithInvalidFormat(t *testing.T) {
	// Invalid format should default to text handler
	logger := New(Config{
		Level:  LevelInfo,
		Format: "invalid",
	})

	if logger == nil {
		t.Error("New() with invalid format returned nil logger")
	}
}

func BenchmarkParseLogLevel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseLogLevel(LevelInfo)
	}
}

func BenchmarkNew(b *testing.B) {
	cfg := Config{
		Level:  LevelInfo,
		Format: "json",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(cfg)
	}
}
