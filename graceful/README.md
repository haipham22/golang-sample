# graceful

Graceful shutdown and goroutine management for Go applications.

## Features

- **Manager**: Coordinate goroutines with cleanup hooks
- **WorkerGroup**: Bounded concurrency for job processing
- **Server**: HTTP server with graceful shutdown
- Signal-aware context (SIGINT, SIGTERM)
- Fail-fast mode for error propagation
- LIFO cleanup execution
- Structured logging with `govern/log`

## Installation

```bash
go get github.com/haipham22/govern/graceful
```

## Usage

### Manager - Basic Example

```go
import (
    "github.com/haipham22/govern/graceful"
    "github.com/haipham22/govern/log"
)

m := graceful.NewManager(nil)

// Run a background goroutine
m.Go(func(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            // Do work
        }
    }
})

// Register cleanup hook
m.Defer(func(ctx context.Context) error {
    log.Info("Cleaning up resources")
    return nil
})

// Wait for shutdown signal
err := m.Wait()
shErr := m.Shutdown(10 * time.Second)
```

### Manager with Custom Logger

```go
logger := log.New(log.WithLevelString("debug"))
m := graceful.NewManager(nil, graceful.WithLogger(logger))
```

### WorkerGroup - Bounded Concurrency

```go
m := graceful.NewManager(nil)
wg := graceful.NewWorkerGroup(10) // Max 10 concurrent jobs

m.Go(func(ctx context.Context) error {
    for i := 0; i < 100; i++ {
        if !wg.TryGo(ctx, func(ctx context.Context) {
            log.Infow("Processing job", "index", i)
        }) {
            break // Shutdown started
        }
    }
    return nil
})

// Wait for inflight jobs during shutdown
m.Defer(func(ctx context.Context) error {
    return wg.Drain(5 * time.Second)
})
```

### HTTP Server

```go
handler := http.NewServeMux()
handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
})

server := graceful.NewServer(
    ":8080",
    handler,
    graceful.WithShutdownTimeout(10*time.Second),
)

server.Start()
```

### Exit Helper

```go
m := graceful.NewManager(nil)
// ... setup goroutines and cleanup

graceful.ExitOnSignal(m, 10*time.Second)
```

## API

### Manager

| Method | Description |
|--------|-------------|
| `NewManager(parent, opts...)` | Create new Manager |
| `Context()` | Get shutdown-aware context |
| `Go(fn)` | Run managed goroutine |
| `Defer(fn)` | Register cleanup hook (LIFO) |
| `Wait()` | Wait for signal or error |
| `Shutdown(timeout)` | Graceful shutdown |
| `InitiateShutdown()` | Trigger shutdown manually |

### Manager Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithFailFast(v)` | Shutdown on first error | `true` |
| `WithLogger(logger)` | Custom logger | `log.Default()` |

### WorkerGroup

| Method | Description |
|--------|-------------|
| `NewWorkerGroup(concurrency)` | Create bounded pool |
| `TryGo(ctx, fn)` | Try to start job |
| `Drain(timeout)` | Wait for inflight jobs |

### Server Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithShutdownTimeout` | Max wait time | 30s |
| `WithServerLogger(logger)` | Custom logger | `log.Default()` |
| `WithServerOptions` | http.Server options | - |
