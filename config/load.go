package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Load reads YAML file, applies ENV variable overrides, and validates the struct.
//
// The function expects a YAML file at the given path and will override any values
// with environment variables. ENV variables use the format: PARENT_CHILD=value
// (e.g., DATABASE_HOST=localhost overrides database.host in YAML).
//
// Example:
//
//	type Config struct {
//	    Server struct {
//	        Port int `validate:"required,min=1,max=65535"`
//	    } `validate:"required"`
//	    Database struct {
//	        Host string `validate:"required"`
//	        Port int    `validate:"required,min=1,max=65535"`
//	    } `validate:"required"`
//	}
//
//	cfg, err := config.Load[Config]("./config.yaml")
func Load[T any](path string) (*T, error) {
	return LoadWithOptions[T](path)
}

// LoadWithOptions reads YAML file, applies ENV variable overrides, and validates
// the struct with custom options.
//
// Options:
//   - WithEnvFile(path): Load .env file to override YAML values
//   - WithENVPrefix(prefix): Add prefix to ENV variable lookups (e.g., "APP")
//   - WithLogger(logger): Use custom logger for debug output
//
// Priority (highest to lowest):
//  1. System environment variables
//  2. .env file values (if WithEnvFile is set)
//  3. YAML file values
func LoadWithOptions[T any](path string, opts ...Option) (*T, error) {
	l := &loader{
		v:        viper.New(),
		logger:   zap.NewNop(),
		validate: validator.New(),
	}

	for _, opt := range opts {
		opt(l)
	}

	// Read YAML file first
	l.v.SetConfigFile(path)
	if err := l.v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Load .env file if path provided (overrides YAML values)
	if l.envPath != "" {
		envViper := viper.New()
		envViper.SetConfigFile(l.envPath)
		envViper.SetConfigType("env")
		if err := envViper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read .env: %w", err)
		}
		// Merge .env values into main viper instance (overrides YAML)
		// Convert underscore keys to dot notation for nested structures
		for _, key := range envViper.AllKeys() {
			// Convert DATABASE_HOST to database.host
			dottedKey := strings.ReplaceAll(key, "_", ".")
			l.v.Set(dottedKey, envViper.Get(key))
		}
	}

	// Bind ENV vars (flatten: database.host -> DATABASE_HOST)
	if l.envPrefix != "" {
		l.v.SetEnvPrefix(l.envPrefix)
	}
	l.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	l.v.AutomaticEnv()

	// Unmarshal
	var cfg T
	if err := l.v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	// Validate
	if err := l.validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	return &cfg, nil
}
