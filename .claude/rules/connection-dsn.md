# Connection DSN Rules

**Rules for configuring external resources (Postgres, Redis, Kafka, RabbitMQ, …) via a single DSN/URL connection string instead of decomposed fields.**

---

## Overview

A **DSN** (Data Source Name) or **URL** is one string holding every connection parameter (host, port, user, password, db, ssl, timeouts). The project standardizes on this for all connection-based resources:

| Resource | Env var | Format |
|----------|---------|--------|
| **Postgres** | `APP_POSTGRES_DSN` | key=value: `host=... user=... dbname=... port=... sslmode=...` |
| **Redis** | `APP_REDIS_URL` | URI: `redis://[user:pass@]host:port/db` |

Source: [`pkg/config/env.go`](../../examples/golang-sample/pkg/config/env.go), [`.env.example`](../../examples/golang-sample/.env.example).

**Core rules:**
- ✅ Configure each resource with ONE connection string (DSN or URL), not separate host/port/user/… fields
- ✅ Name it `APP_<RESOURCE>_DSN` (Postgres-style) or `APP_<RESOURCE>_URL` (URI-style)
- ✅ Store it as a single `string` field with `validate:"required"`
- ✅ Pass the raw string to the driver and let the driver parse it
- ✅ Treat the string as a **secret** (contains password) — env var only, never committed/logged
- ❌ Never decompose a DSN back into host/port/user/password config fields
- ❌ Never hand-build the connection string via `fmt.Sprintf` from fields at the call site

---

## Why Prefer a DSN/URL

- **Library-native** — `gorm.Open(postgres.Open(dsn))` and `redis.ParseURL(url)` parse the string directly; no glue code
- **One env var** — simpler config, fewer validators, one value per environment (dev/staging/prod)
- **Carries all options** — `sslmode`, `connect_timeout`, `search_path`, `application_name`, … that decomposed fields would miss
- **Trivial to pass** — `New(dsn)` instead of `New(host, port, user, pass, db, ssl, …)`
- **Swappable** — change one var to point at a different cluster; no struct migrations

---

## Naming Convention

```text
APP_<RESOURCE>_DSN    # key=value style (Postgres, MySQL)
APP_<RESOURCE>_URL    # URI scheme style (redis://, postgres://, amqp://, mongodb://)
```

**Examples:**
```bash
APP_POSTGRES_DSN="host=localhost user=postgres password=CHANGE_ME dbname=golang_sample port=5432 sslmode=disable"
APP_REDIS_URL="redis://localhost:6379/0"
APP_RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
```

**Rules:**
- ✅ Prefix with `APP_` (matches the project's env convention)
- ✅ Use `_DSN` for key=value formats, `_URL` for URI formats — be consistent per resource
- ✅ One var per resource; don't split primary/replica into many fields (use the DSN's own options)

---

## Config Struct

**One `string` field per resource — not a struct of decomposed fields:**

```go
// GOOD — single DSN/URL field
type AppConfig struct {
    Postgres struct {
        DSN string `mapstructure:"dsn" validate:"required"`
    } `mapstructure:"postgres"`
    Redis struct {
        URL string `mapstructure:"url"`
    } `mapstructure:"redis"`
}

// BAD — decomposed fields; loses driver options, more validators, more glue
type AppConfig struct {
    Postgres struct {
        Host     string `validate:"required"`
        Port     int    `validate:"required"`
        User     string `validate:"required"`
        Password string `validate:"required"`
        DBName   string `validate:"required"`
        SSLMode  string
    }
}
```

**Rules:**
- ✅ `validate:"required"` on connection-critical resources (Postgres)
- ✅ Omit `required` for optional resources (Redis cache may be off)
- ❌ Never store host/port/user/password as separate mapped env keys

---

## Constructor Pattern

**Pass the raw string to the driver; let the driver parse it:**

```go
// Postgres via GORM
func NewGormDB(dsn string) (*gorm.DB, func(), error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})  // pgx parses the DSN
    if err != nil {
        return nil, nil, fmt.Errorf("open postgres: %w", err)
    }
    sqlDB, _ := db.DB()
    cleanup := func() { _ = sqlDB.Close() }
    return db, cleanup, nil
}

// Redis via go-redis
func NewRedis(url string) (*redis.Client, error) {
    opts, err := redis.ParseURL(url)                            // client parses the URL
    if err != nil {
        return nil, fmt.Errorf("parse redis url: %w", err)
    }
    return redis.NewClient(opts), nil
}
```

**Wiring at the composition root** (see [dependency-injection.md](dependency-injection.md)):
```go
db, dbCleanup, err := postgres.NewGormDB(cfg.Postgres.DSN)
```

**Rules:**
- ✅ Constructor takes the string (`New(dsn string)` or `New(cfg Config)` where `Config` wraps the DSN)
- ✅ Let the driver parse — don't pre-split the string
- ✅ Return `(T, cleanup, error)` for resources holding connections
- ❌ Never `fmt.Sprintf` a connection string from decomposed fields at the call site

---

## DSN vs URL Formats

**Pick the format the driver expects:**

| Resource | Format | Example |
|----------|--------|---------|
| Postgres (pgx/GORM) | key=value DSN | `host=h port=5432 user=u password=p dbname=d sslmode=disable` |
| Postgres (alt) | URL | `postgres://u:p@h:5432/d?sslmode=disable` |
| Redis | URL | `redis://[:p]@h:6379/0` |
| RabbitMQ | URL | `amqp://u:p@h:5672/vhost` |
| MongoDB | URL | `mongodb://u:p@h:27017/db?options` |
| Kafka | broker list | `h1:9092,h2:9092` (no DSN — use a comma-separated list) |

**Rules:**
- ✅ Match the driver's native format (GORM/pgx → key=value; redis-go → URL)
- ✅ Include `sslmode`/TLS options in the string for prod (`sslmode=require`, `rediss://`)
- ⚠️ Kafka has no DSN — use a broker list; still keep it as one env var (`APP_KAFKA_BROKERS`)

---

## Validation & Fail-Fast

**Validate presence in config; let the driver validate the contents; verify reachability with a Ping:**

```go
// 1. Presence — struct validator (validate:"required")
// 2. Contents — driver parses (gorm.Open / redis.ParseURL)
// 3. Reachability — Ping at startup
func NewGormDB(dsn string) (*gorm.DB, func(), error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, nil, fmt.Errorf("open postgres: %w", err)   // invalid DSN → fail fast
    }
    sqlDB, _ := db.DB()
    if err := sqlDB.Ping(); err != nil {
        return nil, nil, fmt.Errorf("ping postgres: %w", err)   // unreachable → fail fast
    }
    cleanup := func() { _ = sqlDB.Close() }
    return db, cleanup, nil
}
```

**Rules:**
- ✅ Fail fast at startup on a bad/unreachable DSN (return error from the composition root)
- ✅ Wrap errors with context (`open postgres`, `ping postgres`)
- ❌ Don't write custom DSN parsers — the driver already validates

---

## Secrets Handling

**A DSN/URL embeds credentials — treat it as a secret:**

- ✅ Load from an env var / secret manager only
- ✅ `.env.example` uses placeholders (`password=CHANGE_ME`) — never real values
- ✅ Never commit a real `.env`
- ✅ **Mask** the DSN when logging (redact the password):
  ```go
  // log host/db only, never the password
  log.Infow("connected", "host", parsedHost, "db", parsedDB)
  ```
- ❌ Never log the raw DSN/URL string
- ❌ Never include a real DSN in error messages returned to clients
- 🔗 See [security-checklist](../../examples/golang-sample/docs) and [infrastructure-rules.md](infrastructure-rules.md)

---

## Best Practices & Pitfalls

**✅ DO:**
- One connection string per resource
- Let the driver parse it
- Ping at startup; fail fast on error
- Mask credentials in logs

**❌ DON'T:**
- Decompose into host/port/user/password fields
- `fmt.Sprintf` the string at every call site (parse once in the constructor)
- Log the full DSN
- Omit `sslmode`/TLS in production strings

**Pitfalls:**
```bash
# BAD — decomposed config; loses sslmode/timeouts, more env vars
APP_DB_HOST=localhost
APP_DB_PORT=5432
APP_DB_USER=postgres
APP_DB_PASSWORD=x
APP_DB_NAME=app

# BAD — building the string ad-hoc (repeated, error-prone, leaks password in code)
dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", ...) // in every caller

# BAD — logging the raw DSN
log.Infof("connecting with %s", cfg.Postgres.DSN)   # leaks password

# BAD — prod DSN with sslmode=disable
APP_POSTGRES_DSN="... sslmode=disable"              # use sslmode=require in prod
```

---

## Quick Reference

```bash
# .env — one string per resource
APP_POSTGRES_DSN="host=localhost user=postgres password=CHANGE_ME dbname=golang_sample port=5432 sslmode=disable"
APP_REDIS_URL="redis://localhost:6379/0"
```

```go
// config — single string field
Postgres struct {
    DSN string `mapstructure:"dsn" validate:"required"`
} `mapstructure:"postgres"`

// constructor — driver parses the string
db, cleanup, err := postgres.NewGormDB(cfg.Postgres.DSN)
```

| Concern | Rule |
|---------|------|
| Config shape | one `string` field per resource |
| Env name | `APP_<RESOURCE>_DSN` / `_APP_<RESOURCE>_URL` |
| Parsing | let the driver do it (`postgres.Open`, `redis.ParseURL`) |
| Validation | `required` + driver parse + Ping |
| Secrets | env var only; mask in logs; never commit |
| Prod | include TLS/`sslmode=require` |

---

## References

- Config: [`pkg/config/env.go`](../../examples/golang-sample/pkg/config/env.go), [`.env.example`](../../examples/golang-sample/.env.example)
- Postgres constructor: [`pkg/postgres/postgres.go`](../../examples/golang-sample/pkg/postgres/postgres.go)
- [infrastructure-rules.md](infrastructure-rules.md) — env config & validation
- [database-rules.md](database-rules.md) — GORM connection handling
- [dependency-injection.md](dependency-injection.md) — wiring the connection at the composition root
