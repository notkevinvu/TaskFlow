package logger

import (
	"log/slog"
	"os"
	"strings"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Config holds logger configuration
type Config struct {
	Level  LogLevel
	Format string // "json" or "text"
}

// New creates a new structured logger
func New(cfg Config) *slog.Logger {
	// Parse log level
	level := parseLogLevel(cfg.Level)

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: level,
		// Add source location for errors and above
		// Note: This is a static config based on minimum log level, not per-message
		// All logs at ERROR level and above will include source location
		AddSource: level >= slog.LevelError,
	}

	// Create handler based on format
	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level LogLevel) slog.Level {
	switch strings.ToLower(string(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

