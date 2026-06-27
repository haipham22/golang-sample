package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// DefaultTimeFormat is the default time format for log timestamps.
	DefaultTimeFormat = "2006-01-02T15:04:05.000Z07:00"
)

var (
	// defaultLogger is the global sugar logger instance.
	defaultLogger *zap.SugaredLogger
)

func init() {
	// Initialize default logger with development config
	defaultLogger = New()
}

// Option configures the logger.
type Option func(*config)

type config struct {
	level       zapcore.Level
	encoding    string
	timeFormat  string
	output      zapcore.WriteSyncer
	errorOutput zapcore.WriteSyncer
}

// WithLevel sets the log level.
func WithLevel(level zapcore.Level) Option {
	return func(c *config) {
		c.level = level
	}
}

// WithLevelString sets the log level from string.
// Valid values: "debug", "info", "warn", "error", "fatal", "panic".
func WithLevelString(level string) Option {
	return func(c *config) {
		var l zapcore.Level
		if err := l.UnmarshalText([]byte(level)); err == nil {
			c.level = l
		}
	}
}

// WithEncoding sets the log encoding ("json" or "console").
func WithEncoding(encoding string) Option {
	return func(c *config) {
		c.encoding = encoding
	}
}

// WithTimeFormat sets the time format for timestamps.
func WithTimeFormat(format string) Option {
	return func(c *config) {
		c.timeFormat = format
	}
}

// WithOutput sets the output destination.
func WithOutput(w zapcore.WriteSyncer) Option {
	return func(c *config) {
		c.output = w
	}
}

// WithOutputFile sets the output to a file.
func WithOutputFile(path string) Option {
	return func(c *config) {
		if file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			c.output = file
		}
	}
}

// WithErrorOutput sets the error output destination.
func WithErrorOutput(w zapcore.WriteSyncer) Option {
	return func(c *config) {
		c.errorOutput = w
	}
}

// WithErrorOutputFile sets the error output to a file.
func WithErrorOutputFile(path string) Option {
	return func(c *config) {
		if file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			c.errorOutput = file
		}
	}
}

// WithDevelopment enables development mode (stacktraces on errors).
func WithDevelopment(enabled bool) Option {
	return func(c *config) {
		if enabled {
			c.level = zapcore.DebugLevel
		}
	}
}
