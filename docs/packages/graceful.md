# graceful

Graceful shutdown for Go services and applications.

## Overview

The `graceful` package provides lifecycle management for Go services with controlled startup and shutdown sequences.

## Key Types

### Service Interface

```go
type Service interface {
    Start(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

### ServiceFunc

```go
type ServiceFunc struct {
    StartFunc    func(ctx context.Context) error
    ShutdownFunc func(ctx context.Context) error
}
```

### Manager

```go
type Manager struct {
    // Internal state
}
```

## Key Functions

### Service Creation

```go
// Create Service from functions
func FromFunc(
    start func(ctx context.Context) error,
    shutdown func(ctx context.Context) error,
) Service
```

### Manager Functions

```go
// Create new Manager
func NewManager(ctx context.Context) *Manager

// Register service startup
func (m *Manager) Go(fn func(ctx context.Context) error)

// Register cleanup function
func (m *Manager) Defer(fn func(ctx context.Context) error)

// Wait for all services to complete
func (m *Manager) Wait() error

// Shutdown all services with timeout
func (m *Manager) Shutdown(timeout time.Duration) error

// Run service with graceful shutdown
func Run(ctx context.Context, logger *zap.SugaredLogger, timeout time.Duration, services ...Service)
```

## Usage

### Implementing Service Interface

```go
type MyServer struct {
    server *http.Server
}

func (s *MyServer) Start(ctx context.Context) error {
    log.Info("Starting server")
    return s.server.ListenAndServe()
}

func (s *MyServer) Shutdown(ctx context.Context) error {
    log.Info("Shutting down server")
    return s.server.Shutdown(ctx)
}
```

### Using ServiceFunc

```go
// Simple functions without defining a type
service := graceful.FromFunc(
    func(ctx context.Context) error {
        log.Info("Starting service")
        return runServer(ctx)
    },
    func(ctx context.Context) error {
        log.Info("Shutting down service")
        return cleanup()
    },
)
```

### Using Manager

```go
mgr := graceful.NewManager(ctx)

// Start multiple services concurrently
mgr.Go(func(ctx context.Context) error {
    return httpServer.Start(ctx)
})

mgr.Go(func(ctx context.Context) error {
    return grpcServer.Start(ctx)
})

// Register cleanup
mgr.Defer(func(ctx context.Context) error {
    return database.Close()
})

// Wait for completion
if err := mgr.Wait(); err != nil {
    log.Errorf("Service error: %v", err)
}

// Shutdown with timeout
mgr.Shutdown(30 * time.Second)
```

### Using Run Helper

```go
// Single service
graceful.Run(ctx, logger, 30*time.Second, myService)

// Multiple services (runs in sequence)
httpServer := &HTTPServer{...}
grpcServer := &GRPCServer{...}

graceful.Run(ctx, logger, 30*time.Second, httpServer, grpcServer)
```

### Combining Multiple Services

```go
// Custom service that manages multiple services
type App struct {
    httpServer  graceful.Service
    dbService   graceful.Service
    mqService   graceful.Service
}

func (a *App) Start(ctx context.Context) error {
    mgr := graceful.NewManager(ctx)

    mgr.Go(func(ctx context.Context) error {
        return a.httpServer.Start(ctx)
    })

    mgr.Go(func(ctx context.Context) error {
        return a.dbService.Start(ctx)
    })

    mgr.Go(func(ctx context.Context) error {
        return a.mqService.Start(ctx)
    })

    return mgr.Wait()
}

func (a *App) Shutdown(ctx context.Context) error {
    // Shutdown in reverse order
    if err := a.mqService.Shutdown(ctx); err != nil {
        return err
    }

    if err := a.dbService.Shutdown(ctx); err != nil {
        return err
    }

    return a.httpServer.Shutdown(ctx)
}
```

## Signal Handling

The `Run` function handles OS signals automatically:

```go
// Handles SIGINT and SIGTERM
graceful.Run(context.Background(), logger, 30*time.Second, service)

// When signal received:
// 1. Context cancellation triggered
// 2. Service.Shutdown() called
// 3. Waits up to timeout for completion
// 4. Exits
```

## Common Patterns

### HTTP Server

```go
type HTTPServer struct {
    server *http.Server
}

func NewHTTPServer(addr string, handler http.Handler) *HTTPServer {
    return &HTTPServer{
        server: &http.Server{
            Addr:    addr,
            Handler: handler,
        },
    }
}

func (s *HTTPServer) Start(ctx context.Context) error {
    log.Infof("Starting HTTP server on %s", s.server.Addr)
    
    // Listen for shutdown signal
    go func() {
        <-ctx.Done()
        s.Shutdown(context.Background())
    }()
    
    return s.server.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
    log.Info("Shutting down HTTP server")
    return s.server.Shutdown(ctx)
}

// Usage
server := NewHTTPServer(":8080", handler)
graceful.Run(ctx, logger, 30*time.Second, server)
```

### Database Connection

```go
type DatabaseService struct {
    db *gorm.DB
}

func (s *DatabaseService) Start(ctx context.Context) error {
    log.Info("Connecting to database")
    // Connection already established
    return nil
}

func (s *DatabaseService) Shutdown(ctx context.Context) error {
    log.Info("Closing database connection")
    sqlDB, _ := s.db.DB()
    return sqlDB.Close()
}
```

### Worker Process

```go
type WorkerService struct {
    jobs <-chan Job
}

func (s *WorkerService) Start(ctx context.Context) error {
    log.Info("Starting worker")
    
    for {
        select {
        case <-ctx.Done():
            return nil
        case job := <-s.jobs:
            if err := process(job); err != nil {
                return err
            }
        }
    }
}

func (s *WorkerService) Shutdown(ctx context.Context) error {
    log.Info("Shutting down worker")
    // Drain remaining jobs
    return nil
}
```

## Integration with Govern Packages

The `graceful` package integrates with other govern packages:

- `http.Server` - Implements `graceful.Service`
- `cron.Scheduler` - Implements `graceful.Service`
- `mq/asynq.Server` - Implements `graceful.Service`

## References

- [graceful/service.go](../../graceful/service.go) - Service interface
- [graceful/runner.go](../../graceful/runner.go) - Manager implementation
- [graceful/worker-group.go](../../graceful/worker-group.go) - Worker group utilities
