package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type loader struct {
	v         *viper.Viper
	logger    *zap.Logger
	validate  *validator.Validate
	envPrefix string
	envPath   string
}

// Option configures the loader.
type Option func(*loader)

// WithENVPrefix sets a prefix for environment variable lookups.
//
// Example: WithENVPrefix("APP") will look for APP_DATABASE_HOST
// instead of DATABASE_HOST in environment variables.
//
// This is useful when you want to namespace your application's
// environment variables to avoid conflicts.
func WithENVPrefix(prefix string) Option {
	return func(l *loader) {
		l.envPrefix = prefix
	}
}

// WithEnvFile specifies a .env file path to load before reading the YAML config.
//
// The .env file values will override YAML defaults but can be overridden
// by system environment variables. This is useful for local development
// or for providing environment-specific configuration.
//
// Example .env file format:
//
//	DATABASE_HOST=localhost
//	DATABASE_PORT=5432
func WithEnvFile(path string) Option {
	return func(l *loader) {
		l.envPath = path
	}
}

// WithLogger sets a custom logger for debug output during config loading.
//
// By default, config loading uses a no-op logger. Set a custom logger
// to see debug information about file reading, ENV binding, and validation.
func WithLogger(logger *zap.Logger) Option {
	return func(l *loader) {
		l.logger = logger
	}
}
