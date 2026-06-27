# http

Production-ready HTTP server with graceful shutdown, middleware support, and sensible defaults.

## Features

- Graceful shutdown with configurable timeout
- Middleware chain with execution order control
- Sensible timeout defaults (10s read/write, 60s idle)
- Structured logging integration
- TLS/HTTPS support
- Request ID generation
- Built-in middleware: logging, recovery, CORS, JWT
- Compatible with net/http handlers

## Installation

```bash
go get github.com/haipham22/govern/http
```

## Quick Start

### Basic Server

```go
package main

import (
    "context"
    "net/http"
    "github.com/haipham22/govern/http"
)

func main() {
    // Create a simple handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // Create server with default settings
    server := http.NewServer(":8080", handler)

    // Start server (blocks until shutdown signal)
    server.Start(context.Background())
}
```

### With Middleware

```go
import (
    "github.com/haipham22/govern/http"
    "github.com/haipham22/govern/http/middleware"
    "go.uber.org/zap"
)

func main() {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello with middleware!"))
    })

    logger := zap.NewExample().Sugar()

    server := http.NewServer(":8080", handler,
        // Built-in middleware options
        http.WithRequestID(),
        http.WithLogging(logger),
        http.WithRecovery(logger),
        http.WithCORS(&middleware.CORSConfig{
            AllowedOrigins: []string{"*"},
        }),
    )

    server.Start(context.Background())
}
```

### Custom Middleware

```go
// Define custom middleware
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func main() {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Authenticated!"))
    })

    server := http.NewServer(":8080", handler)

    // Add custom middleware
    server.Use(authMiddleware)

    server.Start(context.Background())
}
```

### Middleware Chain

```go
import "github.com/haipham22/govern/http"

func main() {
    // Create reusable middleware chain
    chain := http.NewChain(
        loggingMiddleware,
        authMiddleware,
        recoveryMiddleware,
    )

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Chain applied!"))
    })

    // Apply chain to handler
    server := http.NewServer(":8080", chain.Then(handler))

    server.Start(context.Background())
}
```

### Server Configuration

```go
import "time"

server := http.NewServer(":8080", handler,
    // Timeouts
    http.WithTimeout(30*time.Second, 30*time.Second, 120*time.Second),
    http.WithReadHeaderTimeout(5*time.Second),
    http.WithShutdownTimeout(30*time.Second),

    // Server settings
    http.WithMaxHeaderBytes(1 << 20), // 1MB
    http.WithLogger(customLogger),

    // TLS
    http.WithTLS("/path/to/cert.pem", "/path/to/key.pem"),
)

// Start with TLS
server.Start(context.Background())
```

### Graceful Shutdown

```go
import (
    "context"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    server := http.NewServer(":8080", handler)

    // Start server in background
    go func() {
        if err := server.Start(context.Background()); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()

    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    // Trigger graceful shutdown
    server.Shutdown(context.Background())
}
```

## Server Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithTimeout(read, write, idle)` | Set read/write/idle timeouts | 10s/10s/60s |
| `WithReadHeaderTimeout(timeout)` | Set header read timeout | 5s |
| `WithShutdownTimeout(timeout)` | Max wait for connections to close | 30s |
| `WithLogger(logger)` | Set custom logger | zap default |
| `WithMaxHeaderBytes(bytes)` | Max header bytes | http.DefaultMaxHeaderBytes |
| `WithBaseContext(fn)` | Set base context factory | nil |
| `WithTLS(certFile, keyFile)` | Configure TLS | nil |
| `WithHTTPServerOptions(opts)` | Set raw http.Server options | nil |
| `WithJWT(config)` | Add JWT authentication | nil |
| `WithLogging(logger)` | Add request logging | nil |
| `WithRecovery(logger)` | Add panic recovery | nil |
| `WithCORS(config)` | Add CORS support | nil |
| `WithRequestID()` | Add request ID generation | nil |

## API Reference

### Types

```go
// Server interface with graceful shutdown support
type Server interface {
    graceful.Service
    Server() *http.Server
    Listen() (net.Listener, error)
    Use(middleware ...Middleware)
}

// Middleware function type
type Middleware func(http.Handler) http.Handler

// Middleware chain for reusable composition
type Chain struct {
    middlewares []Middleware
}
```

### Functions

```go
// Create new server with options
func NewServer(addr string, handler http.Handler, opts ...ServerOption) Server

// Create middleware chain
func NewChain(middlewares ...Middleware) *Chain
func (c *Chain) Then(h http.Handler) http.Handler
func (c *Chain) Append(middlewares ...Middleware) *Chain
func (c *Chain) Prepend(middlewares ...Middleware) *Chain
```

## Testing

```go
func TestHandler(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    server := http.NewServer(":0", handler) // :0 for random port

    // Get listener for testing
    listener, err := server.Listen()
    if err != nil {
        t.Fatal(err)
    }
    defer listener.Close()

    // Test against listener.Addr().String()
}
```

## See Also

- [`http/echo`](./echo/) - Echo framework utilities (JWT, Swagger, handler wrapping)
- [`http/jwt`](./jwt/) - JWT authentication middleware
- [`http/middleware`](./middleware/) - Common middleware implementations
- [`graceful`](../graceful/) - Goroutine and process management

## License

TBD
