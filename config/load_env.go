package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// LoadFromEnv loads .env file and validates the struct (no YAML file).
//
// IMPORTANT: .env files use flat key names with underscores representing nesting.
// Keys in the .env file must be UPPERCASE and map to struct fields via mapstructure tags.
//
// Example .env file:
//
//	SERVER_HOST=localhost
//	SERVER_PORT=8080
//	DATABASE_HOST=localhost
//	DATABASE_PORT=5432
//
// Example Config struct:
//
//	type Config struct {
//	    ServerHost string `mapstructure:"SERVER_HOST" validate:"required"`
//	    ServerPort int    `mapstructure:"SERVER_PORT" validate:"required,min=1,max=65535"`
//	    DatabaseHost string `mapstructure:"DATABASE_HOST" validate:"required"`
//	    DatabasePort int    `mapstructure:"DATABASE_PORT" validate:"required,min=1,max=65535"`
//	}
//
//	cfg, err := config.LoadFromEnv[Config](".env")
func LoadFromEnv[T any](path string) (*T, error) {
	return LoadFromEnvWithOptions[T](path)
}

// LoadFromEnvWithOptions loads .env file and validates the struct with custom options.
//
// Options:
//   - WithENVPrefix(prefix): Add prefix to .env variable lookups (e.g., "APP")
//   - WithLogger(logger): Use custom logger for debug output
//
// When using WithENVPrefix, the prefix is automatically added to all lookups.
// For example, WithENVPrefix("APP") will look for APP_SERVER_HOST in .env.
func LoadFromEnvWithOptions[T any](path string, opts ...Option) (*T, error) {
	l := &loader{
		v:        viper.New(),
		logger:   zap.NewNop(),
		validate: validator.New(),
	}

	for _, opt := range opts {
		opt(l)
	}

	// Read .env file using Viper
	l.v.SetConfigFile(path)
	l.v.SetConfigType("env")
	if err := l.v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read .env: %w", err)
	}

	// Bind ENV vars (flatten: POSTGRES_HOST -> postgres.host)
	if l.envPrefix != "" {
		l.v.SetEnvPrefix(l.envPrefix)
	}
	l.v.SetEnvKeyReplacer(strings.NewReplacer("_", "."))
	l.v.AutomaticEnv()

	// Unmarshal from .env (with underscore to dot mapping)
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
