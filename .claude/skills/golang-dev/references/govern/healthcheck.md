# govern/healthcheck

Import: `github.com/haipham22/govern/healthcheck`

Health check registry for liveness/readiness probes. Concurrent checks, per-check timeout, panic recovery, JSON output.

## Use When

- App exposes liveness/readiness probes.

## Setup

```go
import "github.com/haipham22/govern/healthcheck"

registry := healthcheck.New()

registry.Register("database", func(ctx context.Context) error {
    return db.PingContext(ctx)
})

registry.Register("redis", func(ctx context.Context) error {
    return rdb.Ping(ctx).Err()
}, healthcheck.WithTimeout(2*time.Second))

http.HandleFunc("/health", registry.Handler)
http.HandleFunc("/healthz", healthcheck.Liveness) // always 200
```

## Query Parameters

- `?name=checkname` — run only one check.

## Status Codes

- `200 OK` — all checks pass
- `503 Service Unavailable` — any check failing

## Response

```json
{
  "status": "pass",
  "timestamp": "2024-01-01T00:00:00Z",
  "checks": {
    "database": { "name": "database", "status": "pass", "duration_ms": 5 },
    "redis":    { "name": "redis",    "status": "fail", "message": "connection refused" }
  }
}
```

## Rules

- ✅ Liveness (`/healthz`) returns 200 if process alive — no dependency checks.
- ✅ Readiness (`/health`) checks dependencies (DB, cache) — 503 if unready.
- ✅ Set per-check timeouts.
- ✅ Register only real dependencies.
- ❌ Do not return 200 from readiness when DB is down.
- ❌ Do not run heavy/slow checks synchronously in hot path.

## Avoid

- Static `/health` returning 200 unconditionally.
- Missing dependency checks on readiness.

## Reference

Source: [`healthcheck/`](../../../../../../../healthcheck/).
