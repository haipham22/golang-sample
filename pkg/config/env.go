package config

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/haipham22/govern/config"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// EnvConfigMap defines the application configuration structure
type EnvConfigMap struct {
	App struct {
		Debug bool   `mapstructure:"debug" validate:"required"`
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
var ENV *EnvConfigMap

// LoadConfig reads config file (YAML or .env) and loads to global ENV variable
// Uses govern/config for YAML files or viper for .env files
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

// loadFromEnv loads configuration from .env file using viper
func loadFromEnv(cfgFile string, _ *zap.Logger) (*EnvConfigMap, error) {
	v := viper.New()

	// Read .env file
	v.SetConfigFile(cfgFile)
	v.SetConfigType("env")

	// Enable ENV variable reading with APP_ prefix
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// Unmarshal into struct
	var cfg EnvConfigMap
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Validate
	if err := validator.New().Struct(&cfg); err != nil {
		return nil, err
	}

	ENV = &cfg
	return &cfg, nil
}
