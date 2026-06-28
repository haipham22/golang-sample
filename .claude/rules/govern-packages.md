# Govern Package Preference Rules

**Prefer govern's own packages over external libraries or hand-rolled equivalents. Govern is the project's library — reuse it, don't reinvent it.**

---

## Overview

Govern (`github.com/haipham22/govern`) is the root module of this monorepo and provides production-ready packages for the common concerns of a Go service: HTTP, logging, config, errors, graceful shutdown, retry, cron, message queues, metrics, health checks, and database connections.

These packages are already a dependency (the sample app imports govern via `replace => ../../`). They share consistent conventions — functional options, `context.Context`-first APIs, `(T, cleanup, error)` for resources, and integration with `govern/log` + `govern/graceful`. Pulling in a third-party library for a concern govern already covers fragments those conventions and adds a dependency for no gain.

**Core rules:**
- ✅ **Before importing a new third-party package, check whether govern already covers the concern** — if it does, use govern
- ✅ Use the govern package's functional options instead of reaching for the underlying library directly (`postgres.New(dsn, ...opts)` over `gorm.Open`)
- ✅ Wire govern components together through their shared abstractions (`graceful.Service`, `*zap.SugaredLogger`, `cleanup`)
- ✅ Prefer the govern lifecycle (`graceful.Run`, `graceful.Service`) over ad-hoc signal handling
- ❌ Never hand-roll a solution for a concern govern provides (HTTP server, retry loop, health endpoint, error codes)
- ❌ Never add a competing library that duplicates a govern package's responsibility

---

## Package Map

**When you need X, use the govern package on the left — not the external/manual option on the right:**

| Concern | Use govern | Avoid / replace |
|---------|-----------|-----------------|
| HTTP server + middleware + shutdown | `govern/http` | hand-rolled `net/http.Server` + `signal.Notify` |
| Echo integration (JWT, Swagger) | `govern/http/echo` | re-implementing Echo middleware |
| JWT auth middleware | `govern/http/jwt` | rolling your own JWT validation |
| CORS / logging / recovery / request-id | `govern/http/middleware` | separate middleware libs |
| Structured logging | `govern/log` | configuring Zap from scratch, `log.Printf` |
| Config (YAML + env + validation) | `govern/config` | raw Viper boilerplate, manual env parsing |
| Error codes + wrapping | `govern/errors` | `pkg/errors`, bare `errors.New` |
| Graceful shutdown + goroutine mgmt | `govern/graceful` | `signal.Notify` + `sync.WaitGroup` |
| Retry with backoff | `govern/retry` | `cenkalti/backoff`, manual retry loops |
| Cron scheduling | `govern/cron` | raw `robfig/cron` or `gocron` without lifecycle |
| Task queue (asynq) | `govern/mq/asynq` | wiring asynq client/server by hand |
| Prometheus metrics | `govern/metrics` | manual Prometheus instrumentation |
| Health / liveness / readiness | `govern/healthcheck` | hand-built `/health` handlers |
| PostgreSQL connection | `govern/database/postgres` | direct `gorm.Open` |
| Redis connection | `govern/database/redis` | direct `redis.NewClient` |

---

## Logging — `govern/log`

```go
// GOOD — govern logger; consistent Zap Sugar API, functional options
import "github.com/haipham22/govern/log"

logger := log.New(log.WithLevelString("info"), log.WithEncoding("json"))
log.Infow("server started", "port", 8080)

// Inject the *zap.SugaredLogger into constructors (see dependency-injection.md)

// BAD — configuring zap by hand in every binary
logger, _ := zap.NewProduction()
defer logger.Sync()
sugar := logger.Sugar() // re-deriving the same thing govern already gives you

// BAD — stdlib logging
log.Printf("server started on %d", port) // no levels, no structure, no fields
```

---

## Config — `govern/config`

```go
// GOOD — govern: YAML + ENV override + validation in one generic call
import "github.com/haipham22/govern/config"

cfg, err := config.Load[AppConfig]("./config.yaml", config.WithENVPrefix("APP"))

// BAD — raw Viper with manual Unmarshal + separate validation pass
v := viper.New()
v.SetConfigFile("./config.yaml")
v.AutomaticEnv()
v.SetEnvPrefix("APP")
if err := v.ReadInConfig(); err != nil { return err }
var cfg AppConfig
if err := v.Unmarshal(&cfg); err != nil { return err } // + re-implement validation
```

---

## Errors — `govern/errors`

```go
// GOOD — govern: error codes flow through to the HTTP error handler
import "github.com/haipham22/govern/errors"

return errors.NewCode(errors.CodeNotFound, "user not found")
if errors.IsCode(err, errors.CodeNotFound) { ... }

// BAD — bare errors lose categorization; the handler can't map a status
return errors.New("user not found") // no code → 500 instead of 404

// BAD — pulling in github.com/pkg/errors when govern wraps the same idea
```

> The centralized HTTP error handler relies on govern error codes to pick the status — see [web-framework-rules.md](web-framework-rules.md).

---

## HTTP Server & Lifecycle — `govern/http` + `govern/graceful`

```go
// GOOD — govern server implements graceful.Service; pass to Run, done
import (
    "github.com/haipham22/govern/http"
    "github.com/haipham22/govern/graceful"
)

server := http.NewServer(":8080", handler,
    http.WithRequestID(),
    http.WithLogging(logger),
    http.WithRecovery(logger),
)
return graceful.Run(ctx, logger, 10*time.Second, server)

// BAD — hand-rolled server + signal + shutdown (govern already does this)
srv := &http.Server{Addr: ":8080", Handler: handler}
go srv.ListenAndServe()
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
<-sigCh
srv.Shutdown(ctx) // duplicated, no cleanup ordering, no fail-fast
```

- ✅ Implement `graceful.Service` for anything with a lifecycle (HTTP server, worker, consumer) — see [graceful.md](graceful.md)
- ✅ Use `graceful.Run` / `graceful.Manager` / `graceful.WorkerGroup` instead of raw goroutines + `sync.WaitGroup`

---

## Retry — `govern/retry`

```go
// GOOD — govern retry with backoff + context
import "github.com/haipham22/govern/retry"

err := retry.Do(func() error { return callAPI() },
    retry.MaxAttempts(5),
    retry.Backoff(retry.NewExponentialBackoff()),
)

// BAD — hand-rolled loop; no jitter, no context, reinvented backoff
for i := 0; i < 5; i++ {
    if err := callAPI(); err == nil { break }
    time.Sleep(time.Duration(i) * 100 * time.Millisecond)
}

// BAD — adding cenkalti/backoff when govern/retry covers the same ground
```

---

## Background Work — `govern/cron` & `govern/mq/asynq`

```go
// GOOD — govern cron implements graceful.Service; jobs drain on shutdown
scheduler, cleanup, _ := cron.New(cron.WithLogger(logger))
scheduler.DurationJob(5*time.Minute, cleanupJob)
graceful.Run(ctx, logger, 30*time.Second, scheduler)

// GOOD — govern asynq wrapper for the task queue; consistent options + lifecycle
//   (see mq/asynq/README.md)

// BAD — raw robfig/cron or gocron without graceful shutdown; jobs killed mid-run
```

---

## Observability — `govern/metrics` & `govern/healthcheck`

```go
// GOOD — govern metrics middleware auto-tracks requests, duration, size
handler := metrics.HTTPMiddleware(handler, "api")
http.Handle("/metrics", metrics.HandlerDefault())

// GOOD — govern healthcheck registry with per-check timeouts + JSON output
registry := healthcheck.New()
registry.Register("database", func(ctx context.Context) error { return db.PingContext(ctx) })
http.Handle("/health", registry.Handler)

// BAD — bespoke /health that returns 200 unconditionally; no dependency checks
// BAD — instrumenting Prometheus by hand instead of the metrics middleware
```

---

## Database Connections — `govern/database/postgres` & `govern/database/redis`

```go
// GOOD — govern postgres: returns (*gorm.DB, cleanup, error) with sane pool defaults
import "github.com/haipham22/govern/database/postgres"

db, cleanup, err := postgres.New(cfg.Postgres.DSN, postgres.WithMaxOpenConns(50))
defer cleanup()

// GOOD — govern redis: UniversalClient (standalone + cluster), DSN parsing, cleanup
import "github.com/haipham22/govern/database/redis"

client, cleanup, err := redis.New(cfg.Redis.URL)

// BAD — direct gorm.Open / redis.NewClient; no cleanup triple, no pool defaults
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // leaks pool if no cleanup
```

- ✅ Pass a single DSN/URL string and let the driver parse it — see [connection-dsn.md](connection-dsn.md)
- ✅ Always capture and defer the returned `cleanup`

---

## When an Exception Is OK

Govern wraps the underlying library, not every feature of it. An escape hatch is fine when:

- ✅ You need an **advanced option** govern's functional options don't expose — open a PR to add an option to govern rather than bypassing it
- ✅ Govern genuinely **doesn't cover** the concern (e.g., a specific client SDK) — then a third-party lib is the right call
- ✅ You're working **inside the govern library itself** (root module) — there you're *implementing* govern, not consuming it

**When you do add a new external dependency:** note in the PR/commit *why* no govern package fit, so the gap is visible for a future govern contribution.

---

## Anti-Patterns

```go
// BAD — duplicating a govern concern with a third-party lib
import "github.com/cenkalti/backoff"   // govern/retry already exists
import "github.com/pkg/errors"          // govern/errors already exists
import "github.com/robfig/cron/v3"      // govern/cron already exists

// BAD — using the underlying lib when a govern wrapper exists
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // use govern/database/postgres

// BAD — hand-rolling what govern/graceful already provides
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM) // use graceful.Run

// BAD — stdlib log instead of govern/log
log.Printf("started") // no levels, no structure
```

---

## Quick Reference

```go
// The govern stack at a glance
logger := log.New(log.WithLevelString("info"))                 // govern/log
cfg, _ := config.Load[AppConfig]("./config.yaml")               // govern/config
db, dbCleanup, _ := postgres.New(cfg.Postgres.DSN)             // govern/database/postgres
server := http.NewServer(":8080", handler, http.WithLogging(logger)) // govern/http
registry := healthcheck.New()                                  // govern/healthcheck
graceful.Run(ctx, logger, 10*time.Second, server)              // govern/graceful
```

| Need | First choice |
|------|--------------|
| Log | `govern/log` |
| Load config | `govern/config` |
| Categorize errors | `govern/errors` |
| HTTP server / middleware | `govern/http` |
| Shutdown / goroutines | `govern/graceful` |
| Retry a flaky call | `govern/retry` |
| Scheduled jobs | `govern/cron` |
| Async tasks | `govern/mq/asynq` |
| Prometheus metrics | `govern/metrics` |
| Health probes | `govern/healthcheck` |
| Postgres / Redis | `govern/database/postgres`, `govern/database/redis` |

---

## References

- Govern package READMEs: [`http/`](../../http/), [`log/`](../../log/), [`config/`](../../config/), [`errors/`](../../errors/), [`graceful/`](../../graceful/), [`retry/`](../../retry/), [`cron/`](../../cron/), [`mq/`](../../mq/), [`metrics/`](../../metrics/), [`healthcheck/`](../../healthcheck/), [`database/`](../../database/)
- [graceful.md](graceful.md) — lifecycle, `Run`, `Manager`, `WorkerGroup`
- [connection-dsn.md](connection-dsn.md) — DSN/URL config for govern connections
- [dependency-injection.md](dependency-injection.md) — wiring govern components at the composition root
- [web-framework-rules.md](web-framework-rules.md) — centralized error handler + govern error codes
