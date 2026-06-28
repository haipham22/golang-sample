# config

Configuration loading with YAML, .env, and environment variable support.

## Overview

The `config` package provides type-safe configuration loading from YAML files, .env files, and environment variables with validation using `go-playground/validator`.

## Key Functions

### Load Functions

```go
// Load YAML file with ENV variable overrides
func Load[T any](path string) (*T, error)

// Load YAML with custom options (.env file, ENV prefix, logger)
func LoadWithOptions[T any](path string, opts ...Option) (*T, error)

// Load from .env file only (no YAML)
func LoadFromEnv[T any](path string) (*T, error)

// Load from .env with custom options
func LoadFromEnvWithOptions[T any](path string, opts ...Option) (*T, error)
```

### Options

```go
// Add prefix to ENV variable lookups (e.g., "APP")
func WithENVPrefix(prefix string) Option

// Load .env file to override YAML values
func WithEnvFile(path string) Option

// Set custom logger for debug output
func WithLogger(logger *zap.Logger) Option
```

## Usage

### YAML Configuration

```go
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
```

### .env Configuration

```go
// .env file:
// SERVER_HOST=localhost
// SERVER_PORT=8080
// DATABASE_HOST=localhost
// DATABASE_PORT=5432

type Config struct {
    ServerHost string `mapstructure:"SERVER_HOST" validate:"required"`
    ServerPort int    `mapstructure:"SERVER_PORT" validate:"required,min=1,max=65535"`
    DatabaseHost string `mapstructure:"DATABASE_HOST" validate:"required"`
    DatabasePort int    `mapstructure:"DATABASE_PORT" validate:"required,min=1,max=65535"`
}

cfg, err := config.LoadFromEnv[Config](".env")
```

### Combined YAML + .env + ENV

```go
cfg, err := config.LoadWithOptions[Config](
    "config.yaml",
    config.WithEnvFile(".env"),
    config.WithENVPrefix("APP"),
)
```

## Priority (highest to lowest)

1. System environment variables
2. .env file values (if `WithEnvFile` is set)
3. YAML file values

## Environment Variable Format

### For YAML configs
Use underscore-separated env vars to override nested values:

```yaml
# YAML:
database:
  host: localhost
  port: 5432

# ENV Variable:
DATABASE_HOST=production-db
DATABASE_PORT=5433
```

### For .env configs
Use uppercase keys with mapstructure tags:

```go
// .env: DATABASE_HOST=localhost
type Config struct {
    DatabaseHost string `mapstructure:"DATABASE_HOST" validate:"required"`
}
```

## Validation

The package uses `go-playground/validator` for struct validation. Add validation tags to your struct fields:

```go
type Config struct {
    Port int `validate:"required,min=1,max=65535"`
    Host string `validate:"required,hostname"`
}
```

## References

- [config/load.go](../../config/load.go) - Main load functions
- [config/load_env.go](../../config/load_env.go) - .env load functions
- [config/options.go](../../config/options.go) - Configuration options
