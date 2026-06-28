# Quick Start Guide

Get up and running with Govern library and the sample application in 5 minutes.

## Govern Library Quick Start

### Prerequisites

- **Go 1.26+** - [Download](https://go.dev/dl/)
- **Git** - [Download](https://git-scm.com/downloads/)

### 1. Install Govern Packages

```bash
# Add govern packages to your project
go get github.com/haipham22/govern/http
go get github.com/haipham22/govern/config
go get github.com/haipham22/govern/log
go get github.com/haipham22/govern/database/postgres
go get github.com/haipham22/govern/graceful
```

### 2. Basic HTTP Server

```go
package main

import (
    "context"
    "fmt"
    "net/http"

    "github.com/haipham22/govern/http"
    "github.com/haipham22/govern/http/middleware"
    "github.com/haipham22/govern/log"
    "go.uber.org/zap/zapcore"
)

func main() {
    // Create logger
    logger := log.New(log.WithLevel(zapcore.InfoLevel))

    // Create handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
    })

    // Add middleware
    handler = middleware.RequestLog(logger)(handler)
    handler = middleware.Recovery(logger)(handler)

    // Create server
    server := http.NewServer(":8080", handler)

    // Run server
    logger.Info("Starting server on :8080")
    if err := server.Start(context.Background()); err != nil {
        logger.Fatal(err)
    }
}
```

### 3. Configuration

```go
import "github.com/haipham22/govern/config"

type Config struct {
    Server struct {
        Port int `validate:"required,min=1,max=65535"`
    } `validate:"required"`
    Database struct {
        Host string `validate:"required"`
        Port int    `validate:"required,min=1,max=65535"`
    } `validate:"required"`
}

// Load configuration
cfg, err := config.Load[Config]("config.yaml",
    config.WithEnvFile(".env"),
    config.WithENVPrefix("APP"),
)
```

## Sample Application Quick Start

The repository includes a complete sample application demonstrating Govern usage with clean architecture.

### Prerequisites

- **mise** - Toolchain manager (recommended)
- **Docker** - For PostgreSQL database
- **Go 1.26+** - [Download](https://go.dev/dl/)

### 1. Navigate to Sample App

```bash
cd examples/golang-sample
```

### 2. Install Dependencies

```bash
# Install mise tools (including Go)
mise install

# Install Go dependencies
mise exec -- go mod download
```

### 3. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your settings
vim .env
```

Required variables:
- `APP_POSTGRES_DSN` - PostgreSQL connection string
- `APP_API_SECRET` - JWT signing secret (32+ characters)

### 4. Start PostgreSQL

```bash
# Start PostgreSQL with Docker
docker-compose up -d
```

### 5. Run Application

```bash
# Run server
make run

# Or directly
./bin/serverd

# Or with go run
mise exec -- go run . serve
```

The API will be available at `http://localhost:8080`

## Verify Installation

### Check Health Status

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok"
}
```

### Register a User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePassword123!",
    "confirm_password": "SecurePassword123!"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "SecurePassword123!"
  }'
```

### Access Protected Endpoint

```bash
curl http://localhost:8080/api/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## Next Steps

- **Govern Packages**: See [packages/](./packages/) for detailed package documentation
- **Sample App Guide**: See [golang-sample-guide.md](./golang-sample-guide.md) for complete sample app documentation
- **Development Guide**: See [development.md](./development.md) for testing and building
- **Code Standards**: See [code-standards.md](./code-standards.md) for contribution guidelines

## Package Documentation

Detailed documentation for each Govern package:

- [config](./packages/config.md) - Configuration loading
- [http](./packages/http.md) - HTTP server and middleware
- [database](./packages/database.md) - Database clients
- [log](./packages/log.md) - Structured logging
- [errors](./packages/errors.md) - Error handling
- [graceful](./packages/graceful.md) - Graceful shutdown
- [healthcheck](./packages/healthcheck.md) - Health checks
- [metrics](./packages/metrics.md) - Prometheus metrics
- [cron](./packages/cron.md) - Cron scheduler
- [mq](./packages/mq.md) - Message queue
- [retry](./packages/retry.md) - Retry logic

## Troubleshooting

### Port Already in Use

```bash
# Change port in .env
APP_PORT=8081

# Or kill process using port
lsof -ti:8080 | xargs kill
```

### Database Connection Failed

```bash
# Check PostgreSQL is running
docker-compose ps

# View logs
docker-compose logs -f postgres
```

### Module Download Issues

```bash
# Clear module cache
mise exec -- go clean -modcache

# Re-download dependencies
mise exec -- go mod download
```

## Getting Help

- **Govern Packages**: [docs/packages/](./packages/)
- **Sample App**: [examples/golang-sample/](../examples/golang-sample/)
- **GitHub Issues**: [github.com/haipham22/govern/issues](https://github.com/haipham22/govern/issues)
