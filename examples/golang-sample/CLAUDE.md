# golang-sample - Go Development Guide

Sample application demonstrating the [govern](../..) library with clean architecture principles
(Echo + GORM + JWT). Module: `github.com/haipham22/golang-sample`.

## External Import

This module imports govern as an **external dependency** — there is no `go.work`. The `replace`
directive in [go.mod](go.mod) points at the root govern module for local development:

```
require github.com/haipham22/govern v0.0.0
replace github.com/haipham22/govern => ../../
```

For production, drop the `replace` and use the published govern version.

## Project Setup

```bash
mise install                  # Install Go + tools
mise exec -- go mod tidy      # Resolve deps (incl. local govern via replace)
mise exec -- go mod download
```

## Development Workflow

```bash
mise exec -- goimports -w .               # Format
mise exec -- golangci-lint run            # Lint
mise exec -- staticcheck ./...            # Static analysis
mise exec -- errcheck -blank ./...        # Error check
mise exec -- go test ./...                # Test
mise exec -- go test -race ./...          # Test (race)
mise exec -- go test -cover ./...         # Coverage
mise exec -- go build -o bin/serverd .    # Build binary
mise exec -- mockery                      # Regenerate mocks (if interfaces changed)
```

Or via Makefile: `make test | build | run | lint | generate-mocks`.

## Architecture

Clean architecture based on the bxcodec/go-clean-arch pattern.

```
HTTP Request → Handler → Controller → Service → Domain Model ← Storage
```

### Layers

| Layer        | Location                                 | Responsibility                          |
| ------------ | ---------------------------------------- | --------------------------------------- |
| Entry point  | `main.go` + `cmd/` (Cobra: serverd/grpcd/workerd) | Bootstrap, signal handling       |
| Bootstrap    | `internal/bootstrap/`                    | Composition root (manual DI)            |
| HTTP Handler | `internal/handler/rest/`                 | Echo binding, routing, response mapping |
| Controller   | `internal/handler/rest/controllers/`     | Request orchestration, error mapping    |
| gRPC/Job/MQ  | `internal/handler/{grpc,job,message}/`   | gRPC server, cron jobs, asynq consumer  |
| Usecase      | `internal/usecase/{auth,product,user}/`  | Business logic + repository interfaces  |
| Repository   | `internal/repository/{user,postgres,redis}/` | Data access (GORM/Redis) + interfaces |
| Domain       | `internal/domain/`                       | Pure domain entities (no externals)     |
| ORM          | `internal/orm/`                          | GORM entities                           |
| Schemas      | `internal/schemas/`                      | Request/response DTOs (HTTP boundary)   |
| Validator    | `internal/validator/`                    | go-playground/validator wrapper         |
| Errors       | `internal/errors/` (pkg `apperrors`)     | Typed app errors, HTTP status mapping   |
| Config       | `pkg/config/`, `pkg/postgres/`           | Environment config, DB connection       |

**Dependency Rule:** `handler → usecase → domain ← repository`. No handler→ORM or handler→domain
direct dependencies — always convert through schemas at the boundary.

### Dependency Injection

Manual dependency injection (no code generation) in `internal/bootstrap/` (`app.go`).
`bootstrap.New(cfg)` wires the graph — logger → DB → repositories → usecases → controllers →
HTTP server — and returns `(governhttp.Server, cleanup, error)`. `cmd/serverd.go` calls it.
The `New(log, port, appConfig)` constructor wires the graph explicitly:
`appConfig → db → storage → service → controllers → echo → NewHandler → server`.
Errors propagate (e.g. `ErrMissingJWTSecret`); the DB cleanup function is returned
and must be called on shutdown.

## Development Rules

The full Go ruleset is at [`../../.claude/rules/`](../../.claude/rules/) — see
[`.claude/rules/README.md`](../../.claude/rules/README.md). Key rules:

- ✅ Always use `mise exec --` for Go commands
- ✅ Pass `context.Context` as first parameter in I/O functions
- ✅ Use transactions for multi-step database operations
- ✅ Validate input before database operations
- ✅ Add Swagger annotations to HTTP handlers
- ✅ Follow clean architecture layering
- ✅ Never expose password fields (`json:"-"`)
- ✅ Wrap errors with context: `fmt.Errorf("...: %w", err)`

## Configuration

Copy `.env.example` to `.env`. Key variables: `APP_ENV`, `APP_DEBUG`, `APP_POSTGRES_DSN`,
`APP_API_SECRET` (32+ chars). Never commit real credentials.

## Testing

- SQLite in-memory for storage tests (no external DB required)
- Mockery-generated mocks in `internal/mocks/{service,storage}/`
- Test fixtures: `.test-env`, `config.test.yaml` (resolved via `getProjectRoot()`)
- Always run with `-race`: `mise exec -- go test -race ./...`

## File Naming

- **Go files:** `snake_case` (`user_service.go`)
- **Test files:** `source_test.go` (`user_service_test.go`)
- **Packages:** `snake_case`, singular, lowercase (`package auth`)
