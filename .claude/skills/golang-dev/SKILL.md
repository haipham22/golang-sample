---
name: golang-dev
description: Portable Go sample-app development workflow with mise, clean architecture, manual DI, and optional govern package usage. Use when editing sample app Go code, Go modules, tests, mocks, Swagger, or architecture rules.
license: MIT
argument-hint: "[command]"
metadata:
  project: go-sample-app
  go_version: "1.26.4"
---

# Go Sample App Development

Portable skill for a Go sample application. Copy the whole skill directory, not only `SKILL.md`.

```text
.claude/skills/golang-dev/
├── SKILL.md
└── references/
    ├── development-rules.md
    ├── folder-structure.md
    └── govern/
        ├── README.md
        └── <one file per govern package>
```

## Detect Sample Root

Before commands, find app root:

1. If current directory has app `go.mod`, use `.`.
2. Else if `examples/golang-sample/go.mod` exists, use `examples/golang-sample`.
3. Else ask user for app root.

If `go.mod` has local govern replacement, treat govern as external dependency:

```go
replace github.com/haipham22/govern => ../../
```

No `go.work` unless target project explicitly uses one.

## Must Read When Needed

These references live inside this skill folder and travel with it:

- `references/development-rules.md` — condensed rules index (links into the full files below).
- `references/rules/` — full portable copies of all 18 Go development rule files (self-contained; the skill works without the project's `.claude/rules/`).
- `references/folder-structure.md` — portable folder/layer guide.
- `references/govern/README.md` — index of one rule file per govern package.

Optional local project sources if present:

- `.claude/rules/clean-architecture.md`
- `.claude/rules/dependency-injection.md`
- `.claude/rules/govern-packages.md`
- `.claude/rules/mockery.md`
- `.claude/rules/mise.md`

Do not depend on `plans/...` paths inside this skill. Plans are project history, not portable skill runtime.

## Quick Commands

Run from detected sample root.

```bash
mise exec -- go mod tidy
mise exec -- goimports -w .
mise exec -- go test ./...
mise exec -- go test -race ./...
mise exec -- go test -cover ./...
mise exec -- go build ./...
```

If project builds one root binary:

```bash
mise exec -- go build -o bin/serverd .
```

If `.mockery.yml` exists:

```bash
mise exec -- mockery
```

Install tools from directory that owns `mise.toml`:

```bash
mise install
```

## Toolchain

Prefer project `mise.toml`.

Expected tools:

- Go project-pinned version
- `goimports`
- `staticcheck`
- `errcheck`
- `golangci-lint`
- `mockery`
- `swag`
- `pre-commit`

No Wire by default. Use manual DI unless target project explicitly still uses Wire.

## Architecture

Default clean architecture:

```text
handler → usecase → domain ← repository
bootstrap → all layers, for wiring only
```

Rules:

- `handler/` handles delivery: HTTP, gRPC, cron, queue.
- `schemas/` owns HTTP request/response DTOs when present.
- `usecase/` owns business flows and interfaces it consumes.
- `repository/` implements usecase interfaces with DB/cache clients.
- `domain/` stays pure: no Echo, GORM, Redis, Zap, Viper.
- `orm/` is persistence-only; never return ORM structs from handlers.
- `bootstrap/` wires dependencies and owns resource cleanup.

## Folder Map

Portable target shape:

```text
<sample-root>/
├── main.go
├── cmd/                         # commands: server, worker, grpc, etc.
├── internal/
│   ├── bootstrap/               # manual DI composition root
│   ├── domain/                  # pure domain entities
│   ├── schemas/                 # HTTP DTOs, if app has HTTP
│   ├── handler/                 # delivery adapters
│   ├── usecase/                 # business logic + consumed interfaces
│   ├── repository/              # DB/cache adapters
│   ├── orm/                     # persistence entities, if using GORM
│   ├── validator/               # validator wrapper, if needed
│   ├── errors/                  # app custom errors
│   └── mocks/                   # generated mocks
├── pkg/                         # small app support packages only
├── .mockery.yml                 # optional
└── go.mod
```

Full detail: `references/folder-structure.md`.

## Development Rules Summary

See full portable rules: `references/development-rules.md`.

Critical rules:

- Use `mise exec --` for Go tools.
- Pass `context.Context` first for I/O.
- Use `db.WithContext(ctx)` for GORM.
- Validate at boundaries and before persistence.
- Wrap unexpected errors with `%w`.
- Keep app-specific errors in project error package; do not force `govern/errors`.
- Use transactions for multi-step consistency.
- Do not log secrets, raw DSNs, passwords, tokens, or PII.
- Generate mocks through `.mockery.yml` only.
- Keep `SKILL.md` under 500 lines; move detail into `references/*.md`.

## Govern Package Usage

If project depends on `github.com/haipham22/govern`, use govern packages as building blocks instead of rewriting their concerns.

Common choices:

- HTTP server/lifecycle: `github.com/haipham22/govern/http`
- Echo helpers: `github.com/haipham22/govern/http/echo`
- JWT middleware: `github.com/haipham22/govern/http/jwt`
- Common middleware: `github.com/haipham22/govern/http/middleware`
- Logging: `github.com/haipham22/govern/log`
- Config: `github.com/haipham22/govern/config`
- Graceful shutdown: `github.com/haipham22/govern/graceful`
- Retry: `github.com/haipham22/govern/retry`
- Cron: `github.com/haipham22/govern/cron`
- Queue: `github.com/haipham22/govern/mq/asynq`
- Metrics: `github.com/haipham22/govern/metrics`
- Health checks: `github.com/haipham22/govern/healthcheck`
- Postgres/Redis: `github.com/haipham22/govern/database/postgres`, `github.com/haipham22/govern/database/redis`

If target project has app-specific error envelope, follow project rule. Do not blindly force `govern/errors`.

Full detail: `references/govern/README.md` — one rule file per package.

## Manual DI

DI lives in `internal/bootstrap/`. No codegen unless project already chose codegen.

Construction order:

```text
config → logger → database/cache → repositories → usecases → handlers/controllers → server
```

Resource constructors that hold handles return cleanup:

```go
(T, func(), error)
```

Call cleanup in reverse construction order.

## Testing

Run from sample root:

```bash
mise exec -- go test ./...
mise exec -- go test -race ./...
mise exec -- go test -cover ./...
```

Focused packages when present:

```bash
mise exec -- go test ./internal/usecase/...
mise exec -- go test ./internal/repository/...
mise exec -- go test ./internal/handler/...
```

## Mocks

If `.mockery.yml` exists:

```bash
mise exec -- mockery
```

Rules:

- Use config, no ad-hoc flags.
- Do not hand-edit `internal/mocks/**/mock_*.go`.
- Regenerate after interface changes.
- Prefer type-safe `EXPECT()` API.

## Swagger

If project uses `swag`:

```bash
mise exec -- swag fmt
mise exec -- swag init -g main.go
```

Keep annotations near HTTP handlers/controllers. Never put real secrets in examples.

## Troubleshooting

Tools missing:

```bash
mise install
mise list
```

Deps stale:

```bash
mise exec -- go mod tidy
mise exec -- go mod download
mise exec -- go mod verify
```

Interface changed:

```bash
mise exec -- mockery
mise exec -- go test ./...
```

## Scripts

Portable helpers in `scripts/`. Run from anywhere; they auto-detect the sample root.

```bash
# Print detected sample app root (examples/<app> in monorepo, or . standalone).
./scripts/detect-sample-root.sh

# Enforce clean-architecture import rules (domain purity, handler→usecase only, etc.).
./scripts/arch-lint.sh

# Flag deps that duplicate govern packages (backoff, pkg/errors, robfig/cron, raw gorm.Open).
./scripts/forbidden-deps.sh

# One-command quality gate: goimports, vet, lint, staticcheck, errcheck, test -race.
./scripts/dev-check.sh

# Scaffold a new usecase: internal/usecase/<name>/{service,impl,dto}.go.
./scripts/scaffold-usecase.sh product
```

Override detection with `SAMPLE_ROOT=/path ./scripts/...`.

## Scope

This skill handles:

- Go development in the sample app.
- Govern package selection (when app depends on govern).
- Clean architecture layering and import rules.
- Mise-based formatting, testing, building.
- Mockery and Swagger workflows.
- Portable scripts for architecture/dep linting and scaffolding.

Does not handle:

- Frontend work.
- Production deployment.
- Database administration beyond local dev commands.
- Root govern library implementation (use project rules there).
