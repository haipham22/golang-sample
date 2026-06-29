# Go Development Rules (Index)

Portable, self-contained rules for the sample app. Full text in `rules/`; govern-package rules in `../govern/`.

## Critical Rules (always apply)

- Use `mise exec --` for every Go tool invocation.
- `context.Context` is the first parameter of any I/O function; pass it through the call chain.
- Layer direction: `handler → usecase → domain ← repository`; `bootstrap/` wires all layers and owns cleanup.
- Files ≤ 200 lines (rule files ≤ 250). Split when over; long self-documenting names are fine.
- Wrap unexpected errors with `%w`; never discard via `_`.
- Validate at trust boundaries and before persistence (format → business → DB constraint).
- Generate mocks only via `.mockery.yml`; never hand-edit `mock_*.go`.
- Do not log secrets, raw DSNs, passwords, tokens, or PII.

## Full Rule Files (`rules/`)

| File | Covers |
|------|--------|
| [rules/golang-context-concurrency.md](rules/golang-context-concurrency.md) | `ctx` first param, goroutines, channels, fan-out/fan-in, race safety |
| [rules/golang-error-handling.md](rules/golang-error-handling.md) | Error wrapping, custom errors, `errors.Is/As`, per-layer handling, logging |
| [rules/golang-validator.md](rules/golang-validator.md) | go-playground/validator setup, tags, custom validators, 3-layer validation |
| [rules/golang-database.md](rules/golang-database.md) | GORM context/transactions, query optimization, retries, connection pool |
| [rules/database-rules.md](rules/database-rules.md) | GORM repository + PostgreSQL with `DBError`/`AppError` error envelope |
| [rules/golang-testing.md](rules/golang-testing.md) | Table-driven tests, subtests, `t.Helper`, benchmarks, integration tags |
| [rules/golang-swagger.md](rules/golang-swagger.md) | swaggo/swag annotations, workflow, model/security definitions |
| [rules/golang-coding-standards.md](rules/golang-coding-standards.md) | Naming, formatting, imports, file size, struct/function organization |
| [rules/golang-idioms.md](rules/golang-idioms.md) | Interface satisfaction, defer, zero values, channels, type assertions |
| [rules/golang-types-values.md](rules/golang-types-values.md) | Pointers vs values, generics, `var`/`const`, struct/constructor design |
| [rules/docker.md](rules/docker.md) | Docker Compose, multi-stage Dockerfile, healthchecks, container security |
| [rules/web-framework-rules.md](rules/web-framework-rules.md) | Echo handlers, middleware, centralized error handler, request validation |
| [rules/mockery.md](rules/mockery.md) | mockery v3 config, interface design for mockability, `EXPECT()` API |
| [rules/dependency-injection.md](rules/dependency-injection.md) | Manual DI in `bootstrap/`, `(T, cleanup, error)` triple, wiring order |
| [rules/connection-dsn.md](rules/connection-dsn.md) | One DSN/URL per resource, driver parsing, fail-fast Ping, secret masking |
| [rules/cobra-cli.md](rules/cobra-cli.md) | `main.go` + `cmd/` shape, `RunE`, flags, composition root in `RunE` |
| [rules/mise.md](rules/mise.md) | mise toolchain, version pinning, dev workflow, CI integration |
| [rules/infrastructure-rules.md](rules/infrastructure-rules.md) | Zap logging, env config, `ConfigError`/`LoggerError` envelopes |

## Govern Package Rules

One rule file per govern package: [../govern/README.md](../govern/README.md).

## Layer → Rule Quick Map

```
handler/     → web-framework-rules.md, golang-swagger.md
usecase/     → golang-validator.md, golang-error-handling.md
repository/  → golang-database.md, database-rules.md, connection-dsn.md
domain/      → golang-types-values.md, golang-idioms.md
bootstrap/   → dependency-injection.md, cobra-cli.md, infrastructure-rules.md
tooling      → mise.md, mockery.md, docker.md
```

## Notes

- These are full verbatim copies of the project's `.claude/rules/` made portable (project-only paths and plan refs generalized).
- If the host project has newer rule files in `.claude/rules/`, prefer them; these copies are the portable fallback.
