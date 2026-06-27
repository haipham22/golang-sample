# cron

Cron scheduler integration with [gocron v2](https://github.com/go-co-op/gocron) and graceful lifecycle management.

## Features

- ✅ Implements `graceful.Service` interface for clean lifecycle management
- ✅ Thread-safe scheduler start/stop using atomic operations
- ✅ Functional options pattern for configuration
- ✅ Integration with govern's logging (Zap)
- ✅ Multiple job types support (Duration, Cron, Daily, Weekly, Monthly, OneTime, RandomDuration)
- ✅ Configurable location/timezone support
- ✅ Graceful shutdown with timeout

## Installation

```bash
go get github.com/haipham22/govern/cron
```

## Quick Start

```go
package main

import (
    "context"
    "time"

    "github.com/haipham22/govern/cron"
    "github.com/haipham22/govern/graceful"
    "github.com/haipham22/govern/log"
)

func main() {
    logger := log.New()

    // Create cron scheduler
    scheduler, cleanup, _ := cron.New(
        cron.WithLogger(logger),
    )
    defer cleanup()

    // Add a duration-based job (runs every 5 minutes)
    _, _ = scheduler.DurationJob(5*time.Minute, func() {
        logger.Info("Running cleanup job")
        // Your job logic here
    })

    // Run with graceful shutdown
    graceful.Run(
        context.Background(),
        logger,
        30*time.Second,
        scheduler,
    )
}
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithLogger(logger)` | Set Zap logger | `log.Default()` |
| `WithLocation(loc)` | Set timezone | `time.Local` |
| `WithStopTimeout(d)` | Shutdown timeout | `30s` |

## Job Types

### Duration Job
Runs at fixed intervals:
```go
scheduler.DurationJob(10*time.Second, myFunc, arg1, arg2)
```

### Cron Job
Runs using cron expression:
```go
scheduler.CronJob("0 * * * *", true, myFunc) // Every hour
```

### Daily Job
Runs daily at specific times:
```go
import gocronv2 "github.com/go-co-op/gocron/v2"

scheduler.DailyJob(1, gocronv2.NewAtTimes(gocronv2.NewAtTime(9, 0, 0)), myFunc)
```

### More Job Types
- `WeeklyJob` - Weekly on specific weekdays
- `MonthlyJob` - Monthly on specific days
- `OneTimeJob` - Run once at specific time
- `RandomDurationJob` - Random interval between min/max

## Graceful Integration

```go
package main

import (
    "context"
    "time"

    "github.com/haipham22/govern/cron"
    "github.com/haipham22/govern/graceful"
    "github.com/haipham22/govern/http"
    "github.com/haipham22/govern/log"
)

func main() {
    logger := log.New()

    // Create HTTP server
    server, _ := http.NewServer(":8080")

    // Create cron scheduler
    scheduler, cleanup, _ := cron.New(cron.WithLogger(logger))
    defer cleanup()

    // Add jobs
    _, _ = scheduler.DurationJob(time.Hour, cleanupJob)

    // Run both services with graceful shutdown
    graceful.Run(
        context.Background(),
        logger,
        30*time.Second,
        server,
        scheduler,
    )
}
```

## Testing

```bash
go test ./cron/...
```

## Wire Integration

[Wire](https://github.com/google/wire) is a compile-time dependency injection tool for Go. Here's how to integrate govern's cron scheduler with Wire:

### Setup

Install Wire:
```bash
go install github.com/google/wire/cmd/wire@latest
```

Add to `go.mod`:
```go
// +build tools
package wire

import (
    "github.com/google/wire/wire"
    "github.com/haipham22/govern/cron"
    "github.com/haipham22/govern/log"
)
```

### Provider Functions

Create provider functions in `cmd/app/wire.go`:

```go
//go:build wireinject
// +build !wire

package main

import (
    "time"

    "github.com/google/wire/wire"
    "github.com/haipham22/govern/cron"
    "github.com/haipham22/govern/log"
    "github.com/haipham22/govern/database/redis"
)

// ProviderSet is the collection of providers for Wire
var ProviderSet = wire.NewSet(
    NewLogger,
    NewRedisClient,
    NewCronScheduler,
    NewCleanupJob,
)

// NewLogger creates a Zap logger
func NewLogger() *zap.SugaredLogger {
    return log.New()
}

// NewRedisClient creates a Redis client
func NewRedisClient() (redis.UniversalClient, func(), error) {
    return redis.New("localhost:6379")
}

// NewCronScheduler creates a cron scheduler with dependencies
func NewCronScheduler(logger *zap.SugaredLogger, redisClient redis.UniversalClient) (*cron.Scheduler, func(), error) {
    return cron.New(
        cron.WithLogger(logger),
        cron.WithDistributedLocker(NewRedisLocker(redisClient)),
    )
}

// NewCleanupJob creates a cleanup job handler
func NewCleanupJob(db *sql.DB) cron.JobHandler {
    return &CleanupHandler{db: db}
}
```

### Wire Configuration

Create `cmd/app/wire.go`:

```go
//go:build !wire
// +build !wire

package main

import (
    "github.com/google/wire/wire"
)

//go:generate go run github.com/google/wire/cmd/wire
func initApp(name string) (*App, error) {
    wire.Build(
        ProviderSet,
        NewApp,
    )
    return &App{}, nil
}
```

### Application Structure

```go
//cmd/app/app.go

type App struct {
    Logger      *zap.SugaredLogger
    Redis       redis.UniversalClient
    Scheduler   *cron.Scheduler
    HTTPServer  *http.Server
}

func NewApp(
    logger *zap.SugaredLogger,
    redisClient redis.UniversalClient,
    scheduler *cron.Scheduler,
) *App {
    return &App{
        Logger:     logger,
        Redis:      redisClient,
        Scheduler:  scheduler,
    }
}

func (a *App) Start(ctx context.Context) error {
    // Register jobs
    _, _ = a.Scheduler.DurationJob(time.Hour, func() {
        // Cleanup logic
    })

    // Start scheduler
    return a.Scheduler.Start(ctx)
}

func (a *App) Stop(ctx context.Context) error {
    return a.Scheduler.Shutdown(ctx)
}
```

### Full Example

```go
//cmd/app/main.go

package main

func main() {
    app, err := initApp("myapp")
    if err != nil {
        panic(err)
    }

    ctx := context.Background()
    if err := app.Start(ctx); err != nil {
        panic(err)
    }

    // Wait for shutdown signal
    <-ctx.Done()

    if err := app.Stop(ctx); err != nil {
        panic(err)
    }
}
```

Generate Wire code:
```bash
cd cmd/app
wire
```

### Benefits of Wire Integration

- ✅ **Compile-time safety**: Dependency injection errors caught at compile time
- ✅ **Explicit dependencies**: All dependencies clearly defined in provider functions
- ✅ **Easy testing**: Mock providers for unit tests
- ✅ **No runtime reflection**: Zero overhead at runtime
- ✅ **Clean initialization**: Automatic dependency graph resolution

### Testing with Wire

Create test providers in `cmd/app/wire_test.go`:

```go
//go:build wireinject

package main

import (
    "github.com/google/wire/wire"
    "github.com/haipham22/govern/cron"
    "github.com/haipham22/govern/log"
    "testing"
)

var MockProviderSet = wire.NewSet(
    NewMockLogger,
    NewMockCronScheduler,
)

func NewMockLogger() *zap.SugaredLogger {
    return zap.NewNop().Sugar()
}

func NewMockCronScheduler(logger *zap.SugaredLogger) (*cron.Scheduler, func(), error) {
    return cron.New(cron.WithLogger(logger))
}

func TestNewApp(t *testing.T) {
    // Initialize with test providers
    app, err := initAppTest("test")
    assert.NoError(t, err)
    assert.NotNil(t, app)
}
```

## API Documentation

See [GoDoc](https://pkg.go.dev/github.com/haipham22/govern/cron) for full API reference.

## Roadmap

- [ ] Phase 2: Job Registration (all 7 job types)
- [ ] Phase 3: Distributed Locker (Redis implementation)
- [ ] Phase 4: Prometheus Metrics
- [ ] Phase 5: Singleton scheduler (`cron.Default()`)

## License

TBD
