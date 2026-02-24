package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/haipham22/govern/config"
	"go.uber.org/zap"
)

// EnvConfigMap defines the application configuration structure
type EnvConfigMap struct {
	App struct {
		Debug bool   `mapstructure:"debug"`
		Env   string `mapstructure:"env" validate:"required"`
	} `mapstructure:"app" validate:"required"`
	Postgres struct {
		DSN string `mapstructure:"dsn" validate:"required"`
	} `mapstructure:"postgres"`
	Redis struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"redis"`
	API struct {
		Secret string `mapstructure:"secret"`
	} `mapstructure:"api"`
}

// ENV is global variable for using config in other places
// Deprecated: Use dependency injection to pass config instead
var ENV *EnvConfigMap

// LoadConfig reads config file (YAML or .env) and loads to global ENV variable
// Uses govern/config for both YAML and .env files
func LoadConfig(cfgFile string, logger *zap.Logger) (*EnvConfigMap, error) {
	// Check if file is .env format (including .env.*, .test-env, etc.)
	if strings.Contains(cfgFile, ".env") || strings.Contains(cfgFile, "-env") {
		return loadFromEnv(cfgFile, logger)
	}

	// Use govern/config for YAML files
	cfg, err := config.LoadWithOptions[EnvConfigMap](cfgFile,
		config.WithENVPrefix("APP"),
		config.WithLogger(logger),
	)
	if err != nil {
		return nil, err
	}

	ENV = cfg
	return cfg, nil
}

// loadFromEnv loads configuration from .env file using govern/config
func loadFromEnv(cfgFile string, logger *zap.Logger) (*EnvConfigMap, error) {
	cfg, err := config.LoadFromEnvWithOptions[EnvConfigMap](cfgFile,
		config.WithENVPrefix("APP"),
		config.WithLogger(logger),
	)
	if err != nil {
		return nil, err
	}

	ENV = cfg
	return cfg, nil
}

// Validate validates the configuration and returns detailed errors
func (c *EnvConfigMap) Validate() error {
	v := validator.New()

	if err := v.Struct(c); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Custom validations
	if c.App.Env != "development" && c.App.Env != "staging" && c.App.Env != "production" {
		return fmt.Errorf("invalid APP_ENV: must be development, staging, or production, got: %s", c.App.Env)
	}

	if c.API.Secret != "" && len(c.API.Secret) < 32 {
		return fmt.Errorf("APP_API_SECRET must be at least 32 characters (got %d)", len(c.API.Secret))
	}

	return nil
}
