# config

Ultra-minimal config helper for Govern library.

## Features

- ✅ Read YAML configuration files
- ✅ Override with environment variables
- ✅ Struct validation with tags
- ✅ Generic API (works with any struct)
- ✅ ~150 lines of code

## Installation

```bash
go get github.com/haipham22/govern/config
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/haipham22/govern/config"
)

// Define your config structure
type Config struct {
    Server struct {
        Host string `validate:"required"`
        Port int    `validate:"required,min=1,max=65535"`
    } `validate:"required"`
    Postgres struct {
        Host     string `validate:"required"`
        Port     int    `validate:"required,min=1,max=65535"`
        Database string `validate:"required"`
    } `validate:"required"`
}

func main() {
    // Load config from YAML file (with ENV override)
    cfg, err := config.Load[Config]("./config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // Use it directly
    fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
}
```

### YAML File

**config.yaml**:
```yaml
server:
  host: "0.0.0.0"
  port: 8080

postgres:
  host: "localhost"
  port: 5432
  database: "myapp"
```

### Environment Variable Override

Environment variables override YAML values:

```bash
# Override server port
SERVER_PORT=9090 ./app

# Override postgres host
POSTGRES_HOST=prod.db.local ./app
```

ENV naming: `SECTION_KEY` → `section.key` (e.g., `SERVER_PORT` → `server.port`)

### With Options

```go
// Use ENV prefix (APP_SERVER_PORT instead of SERVER_PORT)
cfg, err := config.LoadWithOptions[Config](
    "./config.yaml",
    config.WithENVPrefix("APP"),
)

// Use custom logger
logger := zap.NewNop()
cfg, err := config.LoadWithOptions[Config](
    "./config.yaml",
    config.WithLogger(logger),
)
```

## Validation

Uses standard struct validation tags:

```go
type Config struct {
    Host string `validate:"required,hostname|ip"`
    Port int    `validate:"required,min=1,max=65535"`
}
```

Common tags:
- `required` - field must be present
- `min=X` - minimum value
- `max=X` - maximum value
- `oneof=foo bar` - must be one of the values

See [validator documentation](https://github.com/go-playground/validator) for more tags.

## API

```go
// Load reads YAML file, overrides with ENV, validates
func Load[T any](path string) (*T, error)

// LoadWithOptions same as Load with custom options
func LoadWithOptions[T any](path string, opts ...Option) (*T, error)

// WithENVPrefix sets ENV variable prefix
func WithENVPrefix(prefix string) Option

// WithLogger sets custom logger
func WithLogger(logger *zap.Logger) Option
```

## License

TBD
