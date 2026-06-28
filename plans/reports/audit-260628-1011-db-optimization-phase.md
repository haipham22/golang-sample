# Audit: Phase 02 — Database Layer Optimization

**Date:** 2026-06-28 · **Auditor:** main · **Target:** [phase-02-database-optimization.md](../260628-0022-upgrades-plan/phase-02-database-optimization.md) vs actual codebase (branch `feat/monorepo-migration`)

## Verdict

**Phase 02 as written is largely obsolete.** The codebase already exceeds what the doc proposes, and the doc is stale (wrong paths, missing the `Product` entity + `repository/postgres/`). The one genuine gap — a **production migration mechanism** — isn't even what the doc frames ("add indexes"); indexes can't be guaranteed because nothing runs AutoMigrate or a migration tool in prod.

**Recommendation:** rewrite the phase around 3 real tasks (below), reject the rest as YAGNI.

---

## Phase doc vs reality (7 steps)

| Step | Doc says | **Actual state** | Verdict |
|------|----------|------------------|---------|
| 1 Query logging | enable slow-query logger | `pkg/postgres.go` hardcodes `Debug: true` → govern logs **all** queries | ⚠️ Done but **hardcoded on in prod** (verbose) |
| 2 Add indexes | add idx on username/email/created_at | User `unique` tags → index **only if AutoMigrate runs**; Product has none; **no prod migration exists** | ❌ Real gap (different from doc) |
| 3 Optimize queries | Select, Preload, transactions | `CheckUniqueness` already CASE-WHEN (1 query); no associations exist; no multi-step ops | ✅ Done / N/A |
| 4 Pool tuning | MaxIdle/MaxOpen + add lifetime | Already `10/100/1h/10min` via govern — **better** than doc's "current" | ✅ Done |
| 5 Benchmarks | add `user_benchmark_test.go` | Exist in `user_test.go` but **mock-based** (`BenchmarkStorage_*_Mock`) — measure mapper, not Postgres | ⚠️ Exists but misses the goal |
| 6 Doc patterns | write patterns.md | n/a (docs hygiene) | ⬜ Optional |
| 7 Perf baseline | profile + document | Not done (mock benchmarks don't give real Postgres numbers) | ⬜ Blocked by step 5 |

---

## What's actually worth doing (YAGNI-filtered)

### 🔴 Real gap — fix this
**1. Production migration mechanism.** No `.sql` files, no AutoMigrate in app startup, no migrate command (`grep` confirms AutoMigrate only in tests). Consequence: the `unique` indexes on `users.username/email` are **not guaranteed in prod** — they only materialize if someone runs AutoMigrate. Options:
   - **AutoMigrate in dev/bootstrap** (`db.AutoMigrate(&orm.User{}, &orm.Product{})` gated on `APP_ENV=development`) — simplest, matches the `database-rules.md` guidance.
   - **Migration tool** (goose/golang-migrate) — proper for prod.
   Recommend the dev-AutoMigrate path now (KISS), plan a real migration tool later.

### 🟡 Hygiene — small, worth it
**2. Make GORM debug logging config-driven.** [postgres.go:27](examples/golang-sample/pkg/postgres/postgres.go) hardcodes `Debug: true`. Wire it to `cfg.Debug` from env (`APP_DEBUG`); `New(Config)` already supports it — `di.go` just calls `NewGormDB()` which ignores config. Prod shouldn't log every query.
**3. Unify error handling.** [repository/user/user.go](examples/golang-sample/internal/repository/user/user.go) uses `github.com/pkg/errors`; [repository/postgres/product.go](examples/golang-sample/internal/repository/postgres/product.go) uses `apperrors` (govern-style codes). Migrate user repo to `apperrors` for consistency + centralized HTTP mapping. (Code quality, not perf.)

### ⬜ Optional
**4. Real Postgres benchmarks.** Existing benchmarks are mock-based. Add SQLite-in-memory or testcontainers-Postgres benchmarks that actually exercise GORM query perf. Only if a baseline is genuinely wanted — otherwise skip.
**5. Fix stale phase doc.** Paths `internal/storage/`→`internal/repository/`, `internal/model/`→`internal/domain/`; add `Product` entity + `repository/postgres/` coverage.

---

## Explicitly REJECT (YAGNI — doc proposes, codebase doesn't need)

| Doc proposal | Why reject |
|--------------|------------|
| `Select` column limiting | User/Product have ~5 cols; micro-opt, hurts readability |
| `Preload` eager loading | **No associations exist** (Product has no User FK) — no N+1 possible |
| Transactions for multi-step | No multi-step DB ops (create user = 1 stmt; product CRUD = 1 stmt each) |
| Index on `created_at` | No query filters/sorts by `created_at` |
| pgx/sqlc migration | Confirmed out of scope by the doc itself; trigger conditions unmet |

---

## Doc discrepancies (phase doc vs reality)

- **Stale paths:** doc says `internal/storage/user/`, `internal/model/user.go` → actual `internal/repository/user/`, `internal/domain/user.go` (clean-arch refactor).
- **Missing entity:** doc only covers `User`; `Product` + `internal/repository/postgres/` + `internal/orm/product.go` exist and are unmentioned.
- **Wrong baseline assumption:** doc's "current" pool (`MaxIdle/MaxOpen` only) understates reality (already has lifetime + idle-time via govern).
- **Numbering chaos:** this is `phase-02-database-optimization.md` but plan.md calls DB optimization **Phase 03** (Phase 02 = Wire removal). Two `phase-02-*` files exist. (Flagged in plan.md reconciliation block.)

---

## Compliance notes (good)

- ✅ `pkg/postgres` correctly uses **govern/database/postgres** (per govern-packages rule) — not raw `gorm.Open`.
- ✅ Pool returns `(db, cleanup, error)` triple; cleanup wired in `di.go`.
- ✅ Repository interfaces defined at consumer (`repository/user/new.go` `Storage`) — bxcodec pattern.
- ✅ DSN passed as single string (connection-dsn rule).
- ✅ `IsExistBy` column whitelist prevents SQL injection.

---

## Recommended next action

Rewrite phase-02-database-optimization.md to a **3-task scope**: (1) migration mechanism, (2) config-driven debug logging, (3) user→apperrors error unification. Drop the perf-micro-opt steps. Effort drops from 6h → ~2-3h, all real value.

**Unresolved:** choice of migration approach — dev-gated AutoMigrate (KISS, recommended) vs. real migration tool (goose/migrate). Depends on how prod schema is managed today (unknown — no migration files in repo).
