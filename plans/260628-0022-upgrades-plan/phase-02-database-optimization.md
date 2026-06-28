---
title: "Phase 02: Database Layer Optimization (Revised)"
description: "Rewritten per audit: 3 real tasks (migration, config-debug, apperrors). Drops obsolete micro-opts."
status: completed
priority: P1
effort: 3h
branch: main
tags: [database, gorm, optimization, migration, apperrors]
created: 2026-06-28
updated: 2026-06-28
---

# Phase 02: Database Layer Optimization (Revised)

> **Rewritten 2026-06-28** per [audit report](../reports/audit-260628-1011-db-optimization-phase.md).
> The original doc proposed perf micro-optimizations that are already done or unnecessary
> (YAGNI). This revision narrows scope to the **3 genuine gaps** the audit found.

## Why revised

The codebase **already exceeds** what the original doc proposed:
- Connection pool already tuned `10/100/1h/10min` via govern (original doc assumed defaults).
- Query logging already on (`Debug:true`) — though hardcoded (fix below).
- `CheckUniqueness` already single-query CASE-WHEN.
- **No associations exist** → no N+1 possible → `Preload`/eager-loading is YAGNI.
- **No multi-step DB ops** → transactions are YAGNI.
- Benchmarks already exist (mock-based).

The original doc is also **stale**: references `internal/storage/`+`internal/model/` (actual: `internal/repository/`+`internal/domain/`) and omits the `Product` entity + `internal/repository/postgres/`.

## Decision: Stay with GORM

Confirmed by audit. Migration to pgx/sqlc rejected — trigger conditions (DB overhead >50ms/req, complex queries, JSONB needs) unmet. GORM's 2-3x overhead is fine at this scale.

---

## Tasks (the real work)

### Task 1 — Production migration mechanism 🔴

**Problem:** `grep` confirms `AutoMigrate` runs **only in tests**. No `.sql` files, no migrate
command, no AutoMigrate in app startup. The `unique` indexes on `users.username/email` (defined
via gorm tags) **only materialize if AutoMigrate runs** — so they are not guaranteed in any real
environment.

**Fix (KISS):** dev-gated AutoMigrate in the composition root ([`internal/handler/rest/di.go`](examples/golang-sample/internal/handler/rest/di.go)) right after the DB is created:

```go
import "github.com/haipham22/golang-sample/internal/orm"

// After db, cleanup, err := postgres.NewGormDB(...)
if appConfig.App.Env != "production" {
    if err := db.AutoMigrate(&orm.User{}, &orm.Product{}); err != nil {
        cleanup()
        return nil, nil, fmt.Errorf("auto-migrate: %w", err)
    }
}
```

**Why dev-gated:** matches `database-rules.md` ("AutoMigrate in development only"). Prod schema
stays externally managed (as today); dev/test get schema + indexes automatically.

**Acceptance:** ✅ dev/test runs create `users`+`products` with unique indexes; ✅ prod path skips
migrate; ✅ migrate error tears down the DB (calls cleanup).

---

### Task 2 — Config-driven debug logging 🟡

**Problem:** [`pkg/postgres/postgres.go`](examples/golang-sample/pkg/postgres/postgres.go) `NewGormDB`
hardcodes `Debug: true` → GORM logs **every** query in prod (verbose, noisy). The configurable
`New(Config)` path exists but `di.go` doesn't use it.

**Fix:** thread `appConfig.App.Debug` through. Minimal change — `NewGormDB(dsn)` → `NewGormDB(dsn, debug)`:

```go
func NewGormDB(pgDSN string, debug bool) (*gorm.DB, func(), error) {
    return New(Config{ DSN: pgDSN, Debug: debug, MaxIdleConns: 10, MaxOpenConns: 100,
        MaxLifetime: time.Hour, MaxIdleTime: 10 * time.Minute })
}
```

`di.go`: `postgres.NewGormDB(appConfig.Postgres.DSN, appConfig.App.Debug)`.

**Acceptance:** ✅ query logging follows `APP_DEBUG`; ✅ prod (debug=false) silent; ✅ existing
tests still pass.

---

### Task 3 — Unify user repo errors → apperrors 🟡

**Problem:** [`internal/repository/user/user.go`](examples/golang-sample/internal/repository/user/user.go)
uses `github.com/pkg/errors` + bare `fmt.Errorf`. [`repository/postgres/product.go`](examples/golang-sample/internal/repository/postgres/product.go)
uses `apperrors` (typed codes → centralized HTTP mapping). Inconsistent: a raw `fmt.Errorf("invalid field")`
from the user repo maps to **500** instead of **400** at the handler.

**Fix:** migrate user repo to `apperrors`:
- invalid field name → `apperrors.NewCode(apperrors.CodeInvalid, ...)`
- DB errors → `apperrors.WrapCode(apperrors.CodeInternal, err)`
- drop `github.com/pkg/errors`, use stdlib `errors` for `errors.Is`
- **keep** not-found-as-`nil` contract (service `Login` checks `account == nil`)

**Safety (verified):** the service's duplicate-key race-condition detection
(`errors.Is(err, gorm.ErrDuplicatedKey)` + `strings.Contains("duplicate key value...")`) still
works — `apperrors.Error` implements `Unwrap()` (chain preserved) and `Error()` returns the inner
message when wrapped via `WrapCode`.

**Acceptance:** ✅ no `pkg/errors` in user repo; ✅ all tests pass; ✅ duplicate-key detection intact.

---

## Explicitly REJECTED (YAGNI)

| Proposal | Reason |
|----------|--------|
| `Select` column limiting | User/Product have ~5 cols; micro-opt, hurts readability |
| `Preload` eager loading | No associations exist — no N+1 possible |
| Transactions | No multi-step DB operations |
| Index on `created_at` | No query filters/sorts by it |
| Real Postgres benchmarks | Mock benchmarks suffice for this scale; add only if a baseline is genuinely needed |
| pgx/sqlc migration | Trigger conditions unmet (see Decision above) |

---

## Success Criteria (revised)

- ✅ Dev-gated AutoMigrate creates schema + indexes (Task 1)
- ✅ GORM debug logging follows `APP_DEBUG`, not hardcoded (Task 2)
- ✅ User repo uses `apperrors`, consistent with product repo (Task 3)
- ✅ All tests pass with race detector
- ✅ Duplicate-key race-condition handling intact

---

## References

- [Audit report](../reports/audit-260628-1011-db-optimization-phase.md) — full findings
- [DB layer comparison](../reports/researcher-260628-0022-database-layer-comparison.md) — original research
- [database-rules.md](../../.claude/rules/database-rules.md) — AutoMigrate dev-only guidance
- [golang-database.md](../../.claude/rules/golang-database.md) — GORM patterns

---

**Phase Status:** ✅ Completed (2026-06-28) — all 3 tasks done, full `go test -race ./...` green, build + vet clean. Bonus: dropped the unused `App` wrapper from `bootstrap.New` (returns `server` directly).  
**Owner:** Development Team  
**Created:** 2026-06-28
