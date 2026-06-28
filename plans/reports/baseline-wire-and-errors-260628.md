# Wire Removal & Error Management Baseline (Phase 08)

**Date**: 2026-06-28
**Branch**: feat/monorepo-migration
**Working dir**: `examples/golang-sample/`
**Plan**: plans/260627-2307-govern-monorepo-and-wire-removal/ (Phases 08-13)

## Test Baseline (pre-refactor)

| Package                                          | Coverage |
| ------------------------------------------------ | -------- |
| internal/handler/rest/controllers/health         | 52.9%    |
| internal/handler/rest/middlewares                | 73.2%    |
| internal/model                                   | 100.0%   |
| internal/service/auth                            | 87.0%    |
| internal/storage/user                            | 62.0%    |
| pkg/config                                       | 83.3%    |
| pkg/utils/password                               | 100.0%   |
| **Total**                                        | **76.6%** |

Target: maintain ≥76% aggregate; do not regress any package. `go test -race ./...` clean.

## Current Wire Dependency Graph

Entry: `rest.New(log, port, appConfig)` in `internal/handler/rest/wire.go` (build tag `wireinject`),
generated into `wire_gen.go` (build tag `!wireinject`).

```
rest.New(log *zap.SugaredLogger, port int64, appConfig *config.EnvConfigMap)
├── echo.New()                                          → *echo.Echo
├── provideDB(appConfig)                                → (*gorm.DB, cleanup, error)
├── userRepo.New(log, db)                               → Storage
├── provideAuthConfig(appConfig)                        → authConfig{jwtSecret}  (panics if empty)
├── provideAuthService(log, storage, cfg)               → Service   (jwtExpiration = 72h)
├── authctrl.New(service)                               → *authctrl.Controller
├── healthctrl.New(db)                                 → *healthctrl.Controller
├── provideDebugFlag(appConfig)                         → bool
├── provideEnv(appConfig)                              → string
└── NewHandler(log, e, authCtrl, healthCtrl, port, debug, env) → governhttp.Server
```

`wire_gen.go:New()` is the exact logic to reproduce as manual DI in Phase 11. The provider helpers
(`provideDB`, `provideAuthService`, `provideAuthConfig`, `provideDebugFlag`, `provideEnv`) and the
`authConfig` struct are duplicated in both files and must be consolidated into one manual `New()`.

## Wire Usage (to remove)

- `internal/handler/rest/wire.go` — `wire.Build(...)`, build tag `wireinject`
- `internal/handler-rest/wire_gen.go` — generated, build tag `!wireinject`
- `google/wire` in `go.mod` — drop after Phase 11
- Makefile `generate-wire` target — remove after Phase 11

## govern/errors Usage (to replace)

**API surface used** (from `github.com/haipham22/govern/errors`):
- Type: `ErrorCode` (string)
- Codes: `CodeInternal`, `CodeInvalid`, `CodeNotFound`, `CodeUnauthorized`, `CodeForbidden`,
  `CodeConflict` (also available but unused: `CodeAlreadyExists`, `CodeRateLimit`)
- Funcs: `WrapCode(code, err)`, `NewCode(code, msg)`, `GetCode(err) (ErrorCode, bool)`, `IsCode`
- Sentinels: `ErrUnauthorized`

**Files using govern/errors** (6):
1. `internal/handler/rest/handler.go` — `customHTTPErrorHandler`: `GetCode` + switch over codes → HTTP status
2. `internal/handler/rest/controllers/auth/auth.go` — `WrapCode(CodeInvalid, err)` on bind failure (×2)
3. `internal/handler/rest/controllers/auth/auth_test.go` — assertions on codes
4. `internal/validator/validator.go` — `WrapCode(CodeInvalid, err)` on validation failure
5. `internal/service/auth/impl.go` — `WrapCode(CodeInternal, …)`, `NewCode(CodeConflict, …)`, `ErrUnauthorized` (×many)
6. `internal/service/auth/service_test.go` — assertions on codes

## Error Flow

```
Storage (gorm errors) ─┐
Validator (bind fail)  ├──► Service (NewCode/WrapCode/ErrUnauthorized)
Controller (WrapCode)  ┘            │
                                    ▼
                  rest.customHTTPErrorHandler
                  (GetCode → switch → HTTP status + sanitized body)
```

## Replacement Plan (Phases 09-13)

### Phase 09 — Custom error types (`internal/errors/`)
Create a small `internal/errors/` package replicating the needed API:
- `type Code string` + constants (Internal, Invalid, NotFound, Unauthorized, Forbidden, Conflict)
- `type Error struct { Code Code; Err error }` with `Error()`, `Unwrap()`, plus `Code() Code`
- `New(code, msg)`, `Wrap(code, err)`, `FromError(err) (Code, bool)`, `IsCode(err, code)`
- Sentinel `ErrUnauthorized` (and others as needed)
Keep the surface minimal — only what the 6 files use.

### Phase 10 — Centralized error management
- Single HTTP-status mapper: `errors.HTTPStatus(code) int` (replaces the switch in `customHTTPErrorHandler`).
- Consistent error response builder (extract `buildValidationErrorResponse` + generic body).
- Tests for the mapper + builder.

### Phase 11 — Manual DI
- Replace `wire.go` + `wire_gen.go` with a single `internal/handler/rest/handlers.go` (or
  `internal/bootstrap/`) `New()` that reproduces the `wire_gen.go` logic exactly.
- Move provider helpers + `authConfig` into the new file (single source).
- Delete `wire.go`, `wire_gen.go`; drop `google/wire` from `go.mod`; remove Makefile `generate-wire`.

### Phase 12 — Error handler refactoring
- Rewrite `customHTTPErrorHandler` to use the Phase 10 mapper/builder (no switch, no govern imports).
- Update the 6 files to import `internal/errors` instead of `govern/errors`.

### Phase 13 — Testing
- Regenerate mocks if interfaces changed.
- `validate.sh` passes (no google/wire, no govern/errors imports).
- Maintain ≥76% coverage; all race tests green.

## Rollback

Backup tag `backup/wire-implementation` created at this baseline. To roll back:
`git reset --hard backup/wire-implementation`.

## Unresolved Questions
- None blocking. Phase 11 may optionally introduce `internal/bootstrap/` per clean-architecture
  rules; decision deferred to Phase 11 (keep simple — a single `New()` in the rest package is fine).
