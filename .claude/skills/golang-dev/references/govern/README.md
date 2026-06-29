# Govern Package Rules

One file per govern package. Applies only if app depends on `github.com/haipham22/govern`.

## Index

| Package | File | Concern |
|---|---|---|
| `govern/http` | [http.md](http.md) | HTTP server + shutdown |
| `govern/http/echo` | [http-echo.md](http-echo.md) | Echo integration |
| `govern/http/jwt` | [http-jwt.md](http-jwt.md) | JWT auth middleware |
| `govern/http/middleware` | [http-middleware.md](http-middleware.md) | CORS/log/recover/request ID |
| `govern/log` | [log.md](log.md) | Structured logging |
| `govern/config` | [config.md](config.md) | Config loading |
| `govern/errors` | [errors.md](errors.md) | Error codes + wrapping |
| `govern/graceful` | [graceful.md](graceful.md) | Graceful shutdown |
| `govern/retry` | [retry.md](retry.md) | Retry with backoff |
| `govern/cron` | [cron.md](cron.md) | Cron scheduler |
| `govern/mq/asynq` | [mq-asynq.md](mq-asynq.md) | Task queue |
| `govern/metrics` | [metrics.md](metrics.md) | Prometheus metrics |
| `govern/healthcheck` | [healthcheck.md](healthcheck.md) | Health probes |
| `govern/database/postgres` | [database-postgres.md](database-postgres.md) | Postgres connection |
| `govern/database/redis` | [database-redis.md](database-redis.md) | Redis connection |

## Rule

Before adding a new dependency or writing glue, check whether govern already owns that concern. Use govern when it fits. Escape only when app needs behavior govern does not expose.
