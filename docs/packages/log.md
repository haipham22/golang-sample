# log

Structured logging with Zap.

## Overview

The `log` package provides structured logging utilities built on `go.uber.org/zap` with sensible defaults and easy configuration.

## Key Functions

### Logger Creation

```go
// Create new logger with options
func New(opts ...Option) *zap.SugaredLogger

// Get default global logger
func Default() *zap.SugaredLogger

// Set default global logger
func SetDefault(logger *zap.SugaredLogger)

// Sync buffered log entries
func Sync() error
```

### Helper Functions

```go
// Log with context (request_id, user_id, etc.)
func WithContext(ctx context.Context, fields ...Field) *zap.SugaredLogger

// Extract context values for logging
func ExtractRequestID(ctx context.Context) string
func ExtractUserID(ctx context.Context) int64
```

## Options

```go
// Set log level
func WithLevel(level zapcore.Level) Option

// Set encoding format (json/console)
func WithEncoding(encoding string) Option

// Set time format
func WithTimeFormat(format string) Option

// Set output writer
func WithOutput(w zapcore.WriteSyncer) Option

// Set error output writer
func WithErrorOutput(w zapcore.WriteSyncer) Option

// Set field encoder
func WithFieldEncoder(encoder zapcore.Encoder) Option

// Enable caller (file:line)
func WithCaller(enabled bool) Option

// Enable stacktrace for error level
func WithStacktrace(enabled bool) Option
```

## Defaults

```go
const DefaultTimeFormat = "2006-01-02T15:04:05.000Z07:00"

Default logger:
  - Level: Info
  - Encoding: console
  - Output: stdout
  - Error Output: stderr
  - Time Format: ISO8601
  - Caller: enabled
  - Stacktrace: ErrorLevel only
```

## Usage

### Basic Usage

```go
import "github.com/haipham22/govern/log"

// Use default logger
log.Default().Info("Application started")
log.Default().Error("Failed to connect", "error", err)

// Create custom logger
logger := log.New(
    log.WithLevel(zapcore.DebugLevel),
    log.WithEncoding("json"),
)

logger.Info("Starting server", "port", 8080)
```

### Structured Logging

```go
logger := log.Default()

// Info with fields
logger.Infow("User created",
    "user_id", userID,
    "username", username,
    "email", email,
)

// Error with context
logger.Errorw("Database query failed",
    "query", query,
    "error", err,
    "duration_ms", time.Since(start).Milliseconds(),
)

// Debug with fields
logger.Debugw("Processing request",
    "request_id", requestID,
    "path", path,
    "method", method,
)
```

### With Context

```go
import "github.com/haipham22/govern/log"

// Add context values
ctx := context.WithValue(context.Background(), "request_id", "abc-123")
ctx = context.WithValue(ctx, "user_id", int64(456))

logger := log.WithContext(ctx)
logger.Info("Processing request")
// Output: {"request_id":"abc-123","user_id":456,"msg":"Processing request"}

// Extract specific context values
requestID := log.ExtractRequestID(ctx)
userID := log.ExtractUserID(ctx)
logger.Infof("Request %s from user %d", requestID, userID)
```

### Level Logging

```go
logger := log.Default()

// Debug level
logger.Debug("Detailed diagnostic info")

// Info level
logger.Info("Application started", "port", 8080)

// Warn level
logger.Warn("High memory usage", "usage_mb", 512)

// Error level
logger.Error("Failed to connect", "host", host, "error", err)

// Fatal level (calls os.Exit)
logger.Fatal("Cannot start", "error", err)
```

### Configuration

```go
// Production JSON logger
productionLogger := log.New(
    log.WithLevel(zapcore.InfoLevel),
    log.WithEncoding("json"),
    log.WithOutput(os.Stdout),
)

// Development console logger
developmentLogger := log.New(
    log.WithLevel(zapcore.DebugLevel),
    log.WithEncoding("console"),
    log.WithCaller(true),
    log.WithStacktrace(true),
)

// File output logger
file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
fileLogger := log.New(
    log.WithOutput(file),
    log.WithErrorOutput(os.Stderr),
)

// Set as global default
log.SetDefault(productionLogger)
```

### Time Formats

```go
// ISO8601 (default)
logger := log.New(
    log.WithTimeFormat("2006-01-02T15:04:05.000Z07:00"),
)

// Unix timestamp
logger := log.New(
    log.WithTimeFormat("unix"),
)

// Custom format
logger := log.New(
    log.WithTimeFormat("2006-01-02 15:04:05"),
)
```

### Error Logging

```go
// Log error with stack trace
logger := log.New(
    log.WithStacktrace(zapcore.ErrorLevel),
)

// Log error
if err != nil {
    logger.Errorw("Operation failed",
        "operation", "save_user",
        "error", err,
        // Stack trace automatically included
    )
}

// Wrap and log
if err != nil {
    wrappedErr := fmt.Errorf("failed to save user: %w", err)
    logger.Error("Save failed", "error", wrappedErr)
}
```

## Integration with Govern Packages

### With HTTP Middleware

```go
import "github.com/haipham22/govern/http/middleware"

// Add request logging to HTTP server
server := http.NewServer(":8080", handler,
    http.WithMiddleware(middleware.RequestLog(logger)),
)
```

### With Graceful Shutdown

```go
import "github.com/haipham22/govern/graceful"

logger := log.New()
mgr := graceful.NewManager(ctx)

mgr.Go(func(ctx context.Context) error {
    logger.Info("Starting service")
    return service.Start(ctx)
})

mgr.Defer(func(ctx context.Context) error {
    logger.Info("Shutting down service")
    return service.Shutdown(ctx)
})
```

## Output Examples

### Console Output

```console
2024-01-15T10:30:00.000Z INFO    User created user_id=123 username=john email=john@example.com
2024-01-15T10:30:01.000Z ERROR   Database query failed query="SELECT * FROM users" error=connection refused duration_ms=500
```

### JSON Output

```json
{"ts":"2024-01-15T10:30:00.000Z","level":"INFO","caller":"main.go:45","msg":"User created","user_id":123,"username":"john","email":"john@example.com"}
{"ts":"2024-01-15T10:30:01.000Z","level":"ERROR","caller":"db.go:78","msg":"Database query failed","query":"SELECT * FROM users","error":"connection refused","duration_ms":500}
```

## Best Practices

1. **Use structured logging** - Use `Infow`, `Errorw` instead of `Infof`, `Errorf`
2. **Add context** - Include request_id, user_id, and other relevant fields
3. **Log levels** - Use appropriate levels (Debug, Info, Warn, Error)
4. **Avoid sensitive data** - Don't log passwords, tokens, PII
5. **Performance** - Use `Debug` level sparingly in production
6. **Sync before exit** - Call `log.Sync()` before application exit

## References

- [log/logger.go](../../log/logger.go) - Logger implementation
- [log/config.go](../../log/config.go) - Configuration options
- [log/helpers.go](../../log/helpers.go) - Helper functions
