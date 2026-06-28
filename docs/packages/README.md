# Govern Packages

Complete documentation for all Govern library packages.

## Available Packages

| Package | Description | Documentation |
|---------|-------------|----------------|
| [config](./config.md) | Configuration loading with YAML, .env, and ENV variable support | [config.md](./config.md) |
| [cron](./cron.md) | Cron scheduler with graceful shutdown and job lifecycle management | [cron.md](./cron.md) |
| [database](./database.md) | Database client integrations for PostgreSQL and Redis | [database.md](./database.md) |
| [errors](./errors.md) | Standardized error handling with error codes and wrapping | [errors.md](./errors.md) |
| [graceful](./graceful.md) | Graceful shutdown for Go services and applications | [graceful.md](./graceful.md) |
| [healthcheck](./healthcheck.md) | Health check endpoints for Kubernetes readiness/liveness probes | [healthcheck.md](./healthcheck.md) |
| [http](./http.md) | HTTP server with graceful shutdown and middleware support | [http.md](./http.md) |
| [log](./log.md) | Structured logging with Zap | [log.md](./log.md) |
| [metrics](./metrics.md) | Prometheus metrics integration with registry and middleware | [metrics.md](./metrics.md) |
| [mq](./mq.md) | Message queue integration with Asynq for background task processing | [mq.md](./mq.md) |
| [retry](./retry.md) | Exponential backoff retry with configurable policies | [retry.md](./retry.md) |

## Quick Reference

### Core Services

```go
// Configuration
import "github.com/haipham22/govern/config"
cfg, err := config.Load[Config]("config.yaml")

// Logging
import "github.com/haipham22/govern/log"
logger := log.New()

// Errors
import "github.com/haipham22/govern/errors"
err = errors.WrapCode(errors.CodeNotFound, err)
```

### Infrastructure

```go
// Database
import "github.com/haipham22/govern/database/postgres"
db, cleanup, err := postgres.New(dsn)

// Redis
import "github.com/haipham22/govern/database/redis"
client, cleanup, err := redis.New("localhost:6379")

// HTTP Server
import "github.com/haipham22/govern/http"
server := http.NewServer(":8080", handler)
```

### Processing

```go
// Background Jobs
import "github.com/haipham22/govern/cron"
scheduler, cleanup, err := cron.New()

// Task Queue
import "github.com/haipham22/govern/mq/asynq"
server, cleanup, err := asynq.NewServer(redisClient, mux)

// Retry Logic
import "github.com/haipham22/govern/retry"
err := retry.Do(func() error { return callAPI() })
```

### Observability

```go
// Health Checks
import "github.com/haipham22/govern/healthcheck"
http.HandleFunc("/readyz", healthcheck.ReadinessHandler(dbCheck))

// Metrics
import "github.com/haipham22/govern/metrics"
metrics.MustRegisterDefault(myMetric)
http.Handle("/metrics", metrics.HandlerDefault())
```

### Lifecycle Management

```go
// Graceful Shutdown
import "github.com/haipham22/govern/graceful"
graceful.Run(ctx, logger, 30*time.Second, server)
```

## Package Categories

### Configuration & Logging
- [config](./config.md) - Load and validate configuration
- [log](./log.md) - Structured logging with Zap

### Data & Storage
- [database/postgres](./database.md) - PostgreSQL with GORM
- [database/redis](./database.md) - Redis client

### HTTP & API
- [http](./http.md) - HTTP server with middleware
- [http/echo](./http.md) - Echo framework integration
- [http/middleware](./http.md) - Common middleware

### Background Processing
- [cron](./cron.md) - Scheduled job execution
- [mq/asynq](./mq.md) - Task queue processing

### Resilience
- [errors](./errors.md) - Error codes and wrapping
- [retry](./retry.md) - Exponential backoff retry
- [graceful](./graceful.md) - Graceful shutdown

### Observability
- [healthcheck](./healthcheck.md) - Health check endpoints
- [metrics](./metrics.md) - Prometheus metrics

## Integration Examples

### Web Service

```go
import (
    "github.com/haipham22/govern/config"
    "github.com/haipham22/govern/database/postgres"
    "github.com/haipham22/govern/graceful"
    "github.com/haipham22/govern/healthcheck"
    "github.com/haipham22/govern/http"
    "github.com/haipham22/govern/http/middleware"
    "github.com/haipham22/govern/log"
    "github.com/haipham22/govern/metrics"
)

// Load configuration
cfg, _ := config.Load[Config]("config.yaml")

// Setup logger
logger := log.New(log.WithLevel(zapcore.InfoLevel))

// Setup database
db, cleanup, _ := postgres.New(cfg.Database.DSN)
defer cleanup()

// Setup health checks
http.HandleFunc("/readyz", healthcheck.ReadinessHandler(
    func(ctx context.Context) error {
        return db.PingContext(ctx)
    },
))

// Setup metrics
metrics.MustRegisterDefault(requestsTotal)
http.Handle("/metrics", metrics.HandlerDefault())

// Setup HTTP server
handler := myAPIHandler(db)
handler = middleware.RequestLog(logger)(handler)
handler = middleware.Recovery(logger)(handler)

server := http.NewServer(":8080", handler)

// Run with graceful shutdown
graceful.Run(ctx, logger, 30*time.Second, server)
```

### Background Worker

```go
import (
    "github.com/haipham22/govern/config"
    "github.com/haipham22/govern/cron"
    "github.com/haipham22/govern/database/redis"
    "github.com/haipham22/govern/graceful"
    "github.com/haipham22/govern/log"
    "github.com/haipham22/govern/mq/asynq"
)

// Load configuration
cfg, _ := config.Load[Config]("config.yaml")

// Setup logger
logger := log.New()

// Setup Redis
redisClient, cleanup, _ := redis.New(cfg.Redis.Addr)
defer cleanup()

// Setup task queue
mux := asynq.NewTaskMux()
mux.HandleFunc("email:send", handleSendEmail)

server, cleanup, _ := asynq.NewServer(redisClient, mux,
    asynq.WithConcurrency(10),
    asynq.WithLogger(logger),
)
defer cleanup()

// Setup cron scheduler
scheduler, cleanup, _ := cron.New(cron.WithLogger(logger))
defer cleanup()

scheduler.DurationJob(1*time.Hour, func() {
    cleanupOldData()
})

// Run with graceful shutdown
graceful.Run(ctx, logger, 30*time.Second, server, scheduler)
```

## See Also

- [Development Guide](../development.md) - Testing, building, workflow
- [Code Standards](../code-standards.md) - Naming, style, best practices
- [Sample Application](../../examples/golang-sample/) - Complete working example
- [README](../../README.md) - Project overview and quick start
