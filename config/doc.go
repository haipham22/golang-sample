package config

/*
Config - Configuration loading with YAML, .env, and ENV variable support.

This package provides a simple, type-safe way to load configuration from YAML files,
.env files, and environment variables with validation.

## Usage

### Loading from YAML

	type Config struct {
	    Server struct {
	        Host string `validate:"required"`
	        Port int    `validate:"required,min=1,max=65535"`
	    } `validate:"required"`
	    Database struct {
	        Host string `validate:"required"`
	        Port int    `validate:"required,min=1,max=65535"`
	    } `validate:"required"`
	}

	cfg, err := config.Load[Config]("config.yaml")

### Loading from .env file

	# .env file
	SERVER_HOST=localhost
	SERVER_PORT=8080
	DATABASE_HOST=localhost
	DATABASE_PORT=5432

	type Config struct {
	    ServerHost string `mapstructure:"SERVER_HOST" validate:"required"`
	    ServerPort int    `mapstructure:"SERVER_PORT" validate:"required,min=1,max=65535"`
	    DatabaseHost string `mapstructure:"DATABASE_HOST" validate:"required"`
	    DatabasePort int    `mapstructure:"DATABASE_PORT" validate:"required,min=1,max=65535"`
	}

	cfg, err := config.LoadFromEnv[Config](".env")

### Combining YAML + .env + ENV

	cfg, err := config.LoadWithOptions[Config](
	    "config.yaml",
	    config.WithEnvFile(".env"),
	    config.WithENVPrefix("APP"),
	)

## Priority (highest to lowest)

1. System environment variables
2. .env file values (if WithEnvFile is set)
3. YAML file values

## Environment Variable Format

For YAML configs, use underscore-separated env vars to override nested values:

	YAML:           ENV Variable:
	database.host -> DATABASE_HOST
	server.port   -> SERVER_PORT

For .env configs, use uppercase keys with mapstructure tags:

	.env:                    Struct Tag:
	DATABASE_HOST=localhost -> mapstructure:"DATABASE_HOST"
*/
