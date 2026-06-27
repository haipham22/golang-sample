# healthcheck

Health check registry for HTTP services with liveness/readiness probes.

## Features

- Concurrent health check execution
- Per-check timeout configuration
- Panic recovery for individual checks
- JSON response with detailed results
- Single check or all checks endpoint

## Usage

```go
import "github.com/haipham22/govern/healthcheck"

registry := healthcheck.New()

// Register checks
registry.Register("database", func(ctx context.Context) error {
    return db.PingContext(ctx)
})

registry.Register("redis", func(ctx context.Context) error {
    return rdb.Ping(ctx).Err()
}, healthcheck.WithTimeout(2*time.Second))

// HTTP handler
http.HandleFunc("/health", registry.Handler)
http.HandleFunc("/healthz", healthcheck.Liveness) // always returns 200
```

## Response Format

```json
{
  "status": "pass",
  "timestamp": "2024-01-01T00:00:00Z",
  "checks": {
    "database": {
      "name": "database",
      "status": "pass",
      "duration_ms": 5000000
    },
    "redis": {
      "name": "redis",
      "status": "fail",
      "message": "connection refused"
    }
  }
}
```

## Query Parameters

- `?name=checkname` - run only the specified check

## Status Codes

- 200 OK - all checks passing
- 503 Service Unavailable - any check failing
