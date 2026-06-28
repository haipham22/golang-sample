# http

HTTP server with graceful shutdown and middleware support.

## Overview

The `http` package provides HTTP server utilities with graceful shutdown, middleware chaining, and integration with the Govern ecosystem.

## Key Types

### Server Interface

```go
type Server interface {
    graceful.Service
    Server() *http.Server
    Listen() (net.Listener, error)
    Use(middleware ...Middleware)
}
```

### Middleware Type

```go
type Middleware func(http.Handler) http.Handler
```

## Key Functions

### Server Creation

```go
// Create new server with graceful shutdown
func NewServer(addr string, handler http.Handler, opts ...ServerOption) Server
```

### Server Options

```go
// Set address
func WithAddress(addr string) ServerOption

// Set handler
func WithHandler(handler http.Handler) ServerOption

// Set timeouts
func WithTimeouts(read, write, idle time.Duration) ServerOption

// Set shutdown timeout
func WithShutdownTimeout(timeout time.Duration) ServerOption

// Set logger
func WithLogger(logger *zap.SugaredLogger) ServerOption

// Add middleware
func WithMiddleware(middleware ...Middleware) ServerOption
```

## Usage

### Basic Server

```go
import "github.com/haipham22/govern/http"

// Create handler
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
})

// Create server
server := http.NewServer(":8080", handler)

// Start server (blocks until shutdown)
err := server.Start(context.Background())
```

### With Middleware

```go
// Create handler
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
})

// Create server with middleware
server := http.NewServer(":8080", handler,
    http.WithMiddleware(
        middleware.RequestLog(logger),
        middleware.Recovery(logger),
        middleware.CORS(),
    ),
)

err := server.Start(context.Background())
```

### With Custom Options

```go
server := http.NewServer("", handler,
    http.WithAddress(":8080"),
    http.WithTimeouts(10*time.Second, 10*time.Second, 60*time.Second),
    http.WithShutdownTimeout(30*time.Second),
    http.WithLogger(logger),
)
```

### Using Server Interface

```go
// Get underlying http.Server
httpServer := server.Server()

// Create listener for testing
listener, err := server.Listen()
if err != nil {
    log.Fatal(err)
}
defer listener.Close()

// Start server with listener
go server.Server().Serve(listener)
```

## Subpackages

### echo

Echo framework integration with JWT authentication and Swagger UI.

**Key Features:**
- JWT middleware for Echo
- Swagger UI integration
- Handler wrapping utilities
- Context helpers for current user

```go
import "github.com/haipham22/govern/http/echo"

// JWT middleware
jwtConfig := &echo.JWTMiddlewareConfig{
    Config:         echo.DefaultConfig(),
    TokenExtractor: echo.DefaultTokenExtractor,
    SkipPaths:      []string{"/health", "/login"},
}
jwtConfig.Config.Secret = "your-secret-key"

e := echo.New()
e.Use(echo.JWTMiddleware(jwtConfig))

// Get current user
func handler(c echo.Context) error {
    claims, ok := echo.GetCurrentUser(c)
    if !ok {
        return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
    }
    return c.JSON(http.StatusOK, claims)
}
```

### middleware

Common HTTP middleware for request logging, recovery, CORS, and more.

**Available Middleware:**
- `RequestLog` - Structured request logging
- `Recovery` - Panic recovery with logging
- `CORS` - Cross-origin resource sharing
- `Security` - Security headers
- `Compression` - Response compression
- `Timeout` - Request timeout
- `RateLimit` - Rate limiting

```go
import "github.com/haipham22/govern/http/middleware"

// Apply middleware
handler = middleware.RequestLog(logger)(handler)
handler = middleware.Recovery(logger)(handler)
handler = middleware.CORS()(handler)
handler = middleware.Security()(handler)
handler = middleware.Compression()(handler)

// Or with server
server := http.NewServer(":8080", handler,
    http.WithMiddleware(
        middleware.RequestLog(logger),
        middleware.Recovery(logger),
        middleware.CORS(),
        middleware.Security(),
    ),
)
```

### jwt

JWT authentication middleware for HTTP servers.

```go
import "github.com/haipham22/govern/http/jwt"

// JWT middleware
middleware := jwt.Middleware(jwt.DefaultConfig())
middleware.SetSecret("your-secret-key")

// Apply to handler
protectedHandler := middleware(handler)
```

## Graceful Shutdown

### With graceful.Run

```go
import "github.com/haipham22/govern/graceful"

server := http.NewServer(":8080", handler)

// Use with graceful.Run
graceful.Run(ctx, logger, 30*time.Second, server)
```

### Manual Shutdown

```go
server := http.NewServer(":8080", handler)

// Start in goroutine
go func() {
    if err := server.Start(context.Background()); err != nil {
        log.Printf("Server error: %v", err)
    }
}()

// Handle shutdown signal
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
<-sigChan

// Shutdown gracefully
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
server.Shutdown(ctx)
```

## Integration with Govern

### With log package

```go
import "github.com/haipham22/govern/log"

logger := log.New(
    log.WithLevel(zapcore.InfoLevel),
    log.WithEncoding("json"),
)

server := http.NewServer(":8080", handler,
    http.WithLogger(logger),
    http.WithMiddleware(middleware.RequestLog(logger)),
)
```

### With metrics package

```go
import "github.com/haipham22/govern/metrics"

server := http.NewServer(":8080", handler,
    http.WithMiddleware(metrics.Middleware()),
)

http.Handle("/metrics", metrics.HandlerDefault())
```

### With healthcheck package

```go
import "github.com/haipham22/govern/healthcheck"

http.HandleFunc("/health", healthcheck.LivenessHandler())
http.HandleFunc("/readyz", healthcheck.ReadinessHandler(dbCheck))

handler = http.NewServeMux()
handler.HandleFunc("/api", apiHandler)
handler.HandleFunc("/health", healthcheck.LivenessHandler())
handler.HandleFunc("/readyz", healthcheck.ReadinessHandler(dbCheck))

server := http.NewServer(":8080", handler)
```

## Common Patterns

### REST API Server

```go
// Create API handler
apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // API logic
})

// Apply middleware
handler := middleware.RequestLog(logger)(apiHandler)
handler = middleware.Recovery(logger)(handler)
handler = middleware.CORS()(handler)
handler = middleware.Security()(handler)

// Create server
server := http.NewServer(":8080", handler,
    http.WithTimeouts(10*time.Second, 10*time.Second, 60*time.Second),
    http.WithLogger(logger),
)

// Add health checks
http.Handle("/health", healthcheck.LivenessHandler())
http.Handle("/readyz", healthcheck.ReadinessHandler(dbCheck))
http.Handle("/metrics", metrics.HandlerDefault())

// Start server
graceful.Run(ctx, logger, 30*time.Second, server)
```

### Echo Framework

```go
import "github.com/haipham22/govern/http/echo"

e := echo.New()

// JWT middleware
jwtConfig := &echo.JWTMiddlewareConfig{
    Config:         echo.DefaultConfig(),
    TokenExtractor: echo.DefaultTokenExtractor,
    SkipPaths:      []string{"/health", "/login"},
}
jwtConfig.Config.Secret = "your-secret-key"
e.Use(echo.JWTMiddleware(jwtConfig))

// Routes
e.GET("/health", func(c echo.Context) error {
    return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
})

e.GET("/api/users", func(c echo.Context) error {
    claims := echo.MustGetCurrentUser(c)
    return c.JSON(http.StatusOK, claims)
})

// Start server
server := http.NewServer(":8080", e)
graceful.Run(ctx, logger, 30*time.Second, server)
```

## Best Practices

1. **Use graceful shutdown** - Always use graceful.Run or handle signals properly
2. **Add middleware** - Apply logging, recovery, CORS middleware
3. **Set timeouts** - Configure read, write, and idle timeouts
4. **Health checks** - Provide /health and /readyz endpoints
5. **Structured logging** - Use RequestLog middleware for request tracking
6. **Security headers** - Apply Security middleware for production

## References

- [http/server.go](../../http/server.go) - Server implementation
- [http/middleware.go](../../http/middleware.go) - Middleware utilities
- [http/options.go](../../http/options.go) - Server options
- [http/echo/doc.go](../../http/echo/doc.go) - Echo integration
- [http/middleware/](../../http/middleware/) - Middleware implementations
