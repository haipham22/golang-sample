# healthcheck

Health check endpoints for Kubernetes readiness/liveness probes.

## Overview

The `healthcheck` package provides health check HTTP handlers for Kubernetes deployments with support for multiple check types and custom timeouts.

## Key Types

### Status

```go
type Status string

const (
    StatusPassing Status = "pass"
    StatusFailing Status = "fail"
    StatusWarning Status = "warn"
)
```

### Check Function

```go
type Check func(ctx context.Context) error
```

### Result

```go
type Result struct {
    Name      string        `json:"name"`
    Status    Status        `json:"status"`
    Message   string        `json:"message,omitempty"`
    Duration  time.Duration `json:"duration_ms,omitempty"`
    Timestamp time.Time     `json:"timestamp"`
}
```

### Response

```go
type Response struct {
    Status    Status            `json:"status"`
    Timestamp time.Time         `json:"timestamp"`
    Checks    map[string]Result `json:"checks,omitempty"`
    Duration  time.Duration     `json:"duration_ms,omitempty"`
}
```

## Key Functions

### Handler Creation

```go
// Create liveness handler (always returns 200)
func LivenessHandler() http.HandlerFunc

// Create readiness handler with checks
func ReadinessHandler(checks ...Check) http.HandlerFunc

// Create readiness handler with options
func ReadinessHandlerWithOptions(checks []Check, opts ...Option) http.HandlerFunc
```

### Options

```go
// Set timeout for individual checks
func WithTimeout(d time.Duration) Option

// Disable panic recovery for checks
func DisablePanic() Option
```

## Usage

### Basic Readiness Check

```go
import "github.com/haipham22/govern/healthcheck"

// Simple check
dbCheck := func(ctx context.Context) error {
    sqlDB, _ := db.DB()
    return sqlDB.PingContext(ctx)
}

// Register handler
http.HandleFunc("/readyz", healthcheck.ReadinessHandler(dbCheck))
http.HandleFunc("/livez", healthcheck.LivenessHandler())
```

### Multiple Checks

```go
// Database check
dbCheck := func(ctx context.Context) error {
    sqlDB, _ := db.db.DB()
    return sqlDB.PingContext(ctx)
}

// Redis check
redisCheck := func(ctx context.Context) error {
    return r.redis.Ping(ctx).Err()
}

// External API check
apiCheck := func(ctx context.Context) error {
    _, err := http.Get("https://api.example.com/health")
    return err
}

// Register with multiple checks
http.HandleFunc("/readyz", healthcheck.ReadinessHandler(
    dbCheck,
    redisCheck,
    apiCheck,
))
```

### With Options

```go
checks := []healthcheck.Check{
    dbCheck,
    redisCheck,
}

// With custom timeout
handler := healthcheck.ReadinessHandlerWithOptions(checks,
    healthcheck.WithTimeout(5*time.Second),
)

http.HandleFunc("/readyz", handler)
```

### With Echo Framework

```go
e := echo.New()

// Liveness
e.GET("/livez", echo.WrapHandler(healthcheck.LivenessHandler()))

// Readiness
e.GET("/readyz", echo.WrapHandler(
    healthcheck.ReadinessHandler(dbCheck, redisCheck),
))
```

### Named Checks with Metadata

```go
func namedCheck(name string, check healthcheck.Check) healthcheck.Check {
    return func(ctx context.Context) error {
        start := time.Now()
        err := check(ctx)
        duration := time.Since(start)
        
        if err != nil {
            log.Warnf("Health check %s failed: %v (took %v)", name, err, duration)
        } else {
            log.Infof("Health check %s passed (took %v)", name, duration)
        }
        
        return err
    }
}

checks := []healthcheck.Check{
    namedCheck("database", dbCheck),
    namedCheck("redis", redisCheck),
    namedCheck("external-api", apiCheck),
}

http.HandleFunc("/readyz", healthcheck.ReadinessHandler(checks...))
```

## Response Format

### Passing Response

```json
{
  "status": "pass",
  "timestamp": "2024-01-15T10:30:00Z",
  "checks": {
    "database": {
      "name": "database",
      "status": "pass",
      "duration_ms": 15,
      "timestamp": "2024-01-15T10:30:00Z"
    },
    "redis": {
      "name": "redis",
      "status": "pass",
      "duration_ms": 5,
      "timestamp": "2024-01-15T10:30:00Z"
    }
  },
  "duration_ms": 20
}
```

### Failing Response

```json
{
  "status": "fail",
  "timestamp": "2024-01-15T10:30:00Z",
  "checks": {
    "database": {
      "name": "database",
      "status": "fail",
      "message": "connection refused",
      "duration_ms": 1000,
      "timestamp": "2024-01-15T10:30:00Z"
    },
    "redis": {
      "name": "redis",
      "status": "pass",
      "duration_ms": 5,
      "timestamp": "2024-01-15T10:30:00Z"
    }
  },
  "duration_ms": 1005
}
```

## Kubernetes Integration

### Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      containers:
      - name: myapp
        image: myapp:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /livez
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 5
```

## Best Practices

1. **Liveness vs Readiness**
   - Liveness: Should always return 200 if app is running
   - Readiness: Should fail if dependencies are unavailable

2. **Fast Checks**
   - Keep checks under 1 second
   - Use timeouts to prevent hanging

3. **Idempotent**
   - Checks should not have side effects
   - Don't modify state

4. **Resource Cleanup**
   - Close connections created during checks
   - Don't leak goroutines

## References

- [healthcheck/handler.go](../../healthcheck/handler.go) - Handler implementation
- [healthcheck/types.go](../../healthcheck/types.go) - Type definitions
- [healthcheck/registry.go](../../healthcheck/registry.go) - Check registry
