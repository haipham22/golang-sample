---
title: "Phase 11: Manual DI Implementation"
description: "Replace Wire dependency injection with manual constructors"
status: pending
priority: P1
effort: 4h
dependsOn: [phase-10-centralized-error-management.md]
---

## Overview

**Priority**: P1 | **Status**: pending | **Effort**: 4h

Replace Google Wire dependency injection with explicit manual constructors, maintaining the same dependency graph and lifecycle management while improving code clarity and maintainability.

**Working Directory**: All operations in this phase are performed in the `examples/golang-sample/` directory. All file paths are relative to `examples/golang-sample/`.

## Context Links

- **Parent Plan**: [plan.md](./plan.md)
- **Previous Phase**: [phase-10-centralized-error-management.md](./phase-10-centralized-error-management.md)
- **Next Phase**: [phase-12-error-handler-refactoring.md](./phase-12-error-handler-refactoring.md)
- **Related Files**: `examples/golang-sample/internal/handler/rest/wire.go`, `examples/golang-sample/internal/handler/rest/wire_gen.go`, `examples/golang-sample/cmd/serverd.go`

## Key Insights

**Current Wire Analysis**:
- 8 provider functions in wire.go
- Generated wire_gen.go (44 lines, auto-generated)
- Single entry point: `rest.New(log, port, appConfig)`
- Cleanup function returned for database connection
- No circular dependencies (clean architecture helps)

**Manual DI Benefits**:
- Explicit dependency graph (no code generation)
- Easier debugging and tracing
- Better IDE support and navigation
- Simpler build process (no wire tool)
- More control over initialization order

**Migration Strategy**:
- Preserve exact same initialization order
- Keep cleanup function pattern
- Maintain same constructor signatures
- Layer-by-layer implementation (storage → service → controller → handler)

## Requirements

### Functional Requirements
1. Replace Wire with manual constructors
2. Maintain exact same dependency initialization order
3. Preserve cleanup function lifecycle
4. Create composition root in handler package
5. Remove Wire tool dependencies

### Non-Functional Requirements
- Zero breaking changes to runtime behavior
- Same performance characteristics
- No memory leaks (verify cleanup)
- All tests passing

## Architecture

**Manual DI Structure**:
```
internal/handler/rest/
├── di.go                  # NEW: Manual DI constructors
├── handler.go             # Existing: HTTP server setup
├── routes.go              # Existing: Route registration
└── (wire.go deleted)
   (wire_gen.go deleted)
```

**Dependency Initialization Order** (same as Wire):
```
1. appConfig → authConfig
2. appConfig → db, cleanup
3. db + log → userRepo.Storage
4. storage + authConfig + log → authService
5. authService → authController
6. db → healthController
7. appConfig → debug, env
8. all above → Server
```

**Composition Root** (matches wire.go signature):
```go
// internal/handler/rest/di.go (replaces wire.go)
package rest

import (
    "github.com/haipham22/govern/http"
    "go.uber.org/zap"
    "time"

    "github.com/haipham22/golang-sample/internal/repository/postgres"
    "github.com/haipham22/golang-sample/internal/service/auth"
    "github.com/haipham22/golang-sample/internal/handler/rest/controllers/auth"
    "github.com/haipham22/golang-sample/internal/handler/rest/controllers/health"
    "github.com/haipham22/golang-sample/pkg/config"
)

// New replaces wire.go - manual dependency injection
// Matches exact signature of wire.go for drop-in replacement
func New(
    log *zap.SugaredLogger,
    port int64,
    appConfig *config.EnvConfigMap,
) (governhttp.Server, func(), error) {
    // 1. Extract config values
    jwtSecret := appConfig.API.Secret
    if jwtSecret == "" {
        return nil, nil, errors.New("JWT secret required")
    }

    // 2. Initialize database with cleanup
    dsn := appConfig.Postgres.DSN
    db, cleanup, err := postgres.NewGormDB(dsn)
    if err != nil {
        return nil, nil, err
    }

    // 3. Initialize storage layer
    storage := userRepo.New(log, db)

    // 4. Initialize service layer
    authService := authservice.NewAuthService(
        log,
        storage,
        jwtSecret,
        72*time.Hour,
    )

    // 5. Initialize controllers
    authController := authctrl.New(authService)
    healthController := healthctrl.New(db)

    // 6. Create Echo instance
    e := echo.New()

    // 7. Create HTTP handler
    server := NewHandler(
        log,
        e,
        authController,
        healthController,
        port,
        appConfig.App.Debug,
        appConfig.App.Env,
    )

    return server, cleanup, nil
}
```

**Key differences from Wire**:
- ✅ Explicit dependency construction (no code generation)
- ✅ Clear dependency chain in code
- ✅ Easy to debug and understand  
- ✅ Same function signature as wire.go (drop-in replacement)
- ❌ No compile-time DI verification (manual testing required)

---

**Entry Point** (no changes needed - already uses rest.New):
```go
// cmd/serverd.go (no changes required)
func runCmd(serverd *cobra.Command) error {
    log := zap.S()

    port, err := serverd.Flags().GetInt64("port")
    if err != nil {
        return err
    }

    shutdownTime, err := serverd.Flags().GetInt64("shutdown_time")
    if err != nil {
        return err
    }

    // Load config at composition root
    cfg := config.ENV

    // Use rest.New (same as wire.go, no changes needed)
    handler, cleanup, err := rest.New(log, port, cfg)
    if err != nil {
        return err
    }
    defer cleanup()

    // Create signal context for graceful shutdown
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // Run server with govern graceful runner
    err = govern.Run(ctx, log, time.Duration(shutdownTime)*time.Second, handler)

    return err
}
```

**Graceful Shutdown Sequence** (govern/graceful):
1. Stop accepting new connections
2. Wait for active requests (configurable timeout)
3. Close database connections (via cleanup)
4. Release resources

---

## Bootstrap Directory Structure

**New manual DI structure** (replacing Wire):
```
internal/bootstrap/
├── app.go              # Main DI constructor
├── logger.go          # Logger setup
├── database.go        # Database setup with cleanup
├── http.go            # HTTP server setup
└── worker.go          # Worker setup
```

**Bootstrap Flow:**
```go
// internal/bootstrap/app.go
package bootstrap

import (
    "github.com/haipham22/govern/http"
    "github.com/haipham22/golang-sample/internal/repository"
    "github.com/haipham22/golang-sample/internal/usecase"
    "github.com/haipham22/golang-sample/internal/handler"
    "go.uber.org/zap"
)

type App struct {
    Server governhttp.Server  // ✅ Implements govern/graceful.Runner interface
    Logger *zap.Logger
    DB     *gorm.DB
}

func NewApp(cfg *config.Config) (*App, func(), error) {
    // 1. Logger
    logger, err := NewLogger(cfg.Log)
    if err != nil {
        return nil, nil, err
    }

    // 2. Database with cleanup
    db, cleanup, err := NewDatabase(cfg.Database)
    if err != nil {
        logger.Error("Failed to connect to database", zap.Error(err))
        return nil, nil, err
    }

    // 3. HTTP Server (handler layer handles internal DI: repo → service → controller)
    // All dependency construction happens inside handler/rest package
    server := rest.NewServer(logger, cfg.Server.Port, cfg.Server)

    return &App{
        Server: server,  // governhttp.Server implements govern.Run() interface
        Logger: logger,
        DB:     db,
    }, cleanup, nil
}
```

**Key differences from Wire:**
- ✅ Explicit dependency construction (no code generation)
- ✅ Clear dependency chain in code
- ✅ Easy to debug and understand
- ✅ No external tool dependency
- ❌ No compile-time DI verification (manual testing required)

---

**Entry Point** (replacing Wire):
```go
// cmd/serverd.go
package main

import (
    "context"
    "os/signal"
    "syscall"
    "time"

    govern "github.com/haipham22/govern/graceful"
    "github.com/spf13/cobra"
    "go.uber.org/zap"

    "github.com/haipham22/govern/examples/golang-sample/internal/bootstrap"
    "github.com/haipham22/govern/examples/golang-sample/pkg/config"
)

var serverCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start production API server",
    RunE: func(cmd *cobra.Command, _ []string) error {
        log := zap.S()

        // Load configuration
        cfg, err := config.Load()
        if err != nil {
            return err
        }

        // Bootstrap app (manual DI)
        app, cleanup, err := bootstrap.NewApp(cfg)
        if err != nil {
            return err
        }
        defer cleanup()

        // Create signal context for graceful shutdown
        ctx, stop := signal.NotifyContext(
            context.Background(),
            syscall.SIGINT,
            syscall.SIGTERM,
        )
        defer stop()

        // Run server with govern graceful runner
        shutdownTime, _ := cmd.Flags().GetDuration("shutdown-time")
        err = govern.Run(
            ctx,
            log,
            shutdownTime,
            app.Server,
        )

        return err
    },
}

func init() {
    serverCmd.Flags().Duration("shutdown-time", 10*time.Second, "Graceful shutdown timeout")
}
```

**Graceful Shutdown Sequence** (govern/graceful):
1. Stop accepting new connections
2. Wait for active requests (configurable timeout)
3. Close database connections (via cleanup)
4. Release resources

## Related Code Files

### Files to Create
- `internal/handler/rest/di.go` - Manual DI implementation

### Files to Modify
- `cmd/serverd.go` - Update import from wire.go to di.go
- `go.mod` - Remove Wire dependency (after validation)
- `mise.toml` - Remove wire tool (after validation)

### Files to Delete
- `internal/handler/rest/wire.go` - Wire providers (after validation)
- `internal/handler/rest/wire_gen.go` - Generated code (after validation)

## Implementation Steps

1. **Create Manual DI File** (90m)
   ```go
   // internal/handler/rest/di.go
   package rest

   func New(
       log *zap.SugaredLogger,
       port int64,
       appConfig *config.EnvConfigMap,
   ) (governhttp.Server, func(), error) {
       // 1. Extract config
       jwtSecret := appConfig.API.Secret
       if jwtSecret == "" {
           return nil, nil, errors.New("JWT secret required")
       }

       // 2. Initialize database
       db, cleanup, err := postgres.NewGormDB(appConfig.Postgres.DSN)
       if err != nil {
           return nil, nil, err
       }

       // 3. Initialize storage layer
       storage := userRepo.New(log, db)

       // 4. Initialize service layer
       authService := authservice.NewAuthService(
           log, storage, jwtSecret, 72*time.Hour,
       )

       // 5. Initialize controllers
       authController := authctrl.New(authService)
       healthController := healthctrl.New(db)

       // 6. Create Echo instance
       e := echo.New()

       // 7. Create HTTP handler
       server := NewHandler(
           log, e, authController, healthController,
           port, appConfig.App.Debug, appConfig.App.Env,
       )

       return server, cleanup, nil
   }
   ```

2. **Update cmd/serverd.go** (15m)
   ```go
   // Replace: rest.New(log, port, appConfig)
   // With:    rest.New(log, port, appConfig)
   ```

3. **Add Unit Tests for DI** (45m)
   ```go
   // internal/handler/rest/di_test.go
   func TestNew_Success(t *testing.T)
   func TestNew_MissingJWTSecret(t *testing.T)
   func TestNew_DatabaseError(t *testing.T)
   func TestNew_CleanupFunction(t *testing.T)
   ```

4. **Run Validation Tests** (30m)
   ```bash
   # Run full test suite
   go test ./... -v -cover

   # Integration test: start server
   go run cmd/serverd.go serve

   # Test API endpoints work
   curl http://localhost:8080/health
   ```

5. **Remove Wire Dependencies** (30m)
   ```bash
   # After validation passes:
   rm internal/handler/rest/wire.go
   rm internal/handler/rest/wire_gen.go

   # Update go.mod
   go mod tidy

   # Update mise.toml (remove wire tool)
   ```

6. **Performance Validation** (30m)
   ```bash
   # Benchmark startup time
   hyperfine 'go run cmd/serverd.go serve --timeout 1s'

   # Compare with Wire version (backup branch)
   # Should be similar or better
   ```

## Todo List

- [x] Create manual DI implementation (di.go)
- [x] Update cmd/serverd.go to use New
- [x] Add unit tests for DI construction
- [x] Run validation tests (go test ./...)
- [ ] Integration test server startup
- [ ] Test API endpoints work correctly
- [x] Verify cleanup function works
- [ ] Performance benchmark vs Wire
- [x] Remove wire.go and wire_gen.go
- [x] Remove Wire from go.mod
- [x] Remove wire tool from mise.toml
- [x] Update documentation (remove Wire references)

## Success Criteria

**Definition of Done**:
- Manual DI implemented and tested
- All tests passing with same coverage
- Server starts and runs correctly
- API endpoints work as before
- Cleanup function validated
- Wire code removed from codebase
- Performance equal or better than Wire
- Documentation updated

**Validation Methods**:
```bash
# Compilation
go build ./cmd/serverd.go

# Tests
go test ./... -v -cover

# Integration test
go run cmd/serverd.go serve &
curl http://localhost:8080/health
curl http://localhost:8080/readyz

# Cleanup test
# Verify database connections close properly
```

**Behavioral Verification**:
- Same initialization order as Wire
- Same error handling behavior
- Same cleanup lifecycle
- Same HTTP responses

## Risk Assessment

**Potential Issues**:
1. **Wrong Initialization Order**: Manual order differs from Wire
   - Mitigation: Documented dependency graph from Phase 01, strict order following
2. **Missing Dependencies**: Manual DI may forget a dependency
   - Mitigation: Comprehensive tests, validation against Wire behavior
3. **Cleanup Not Called**: Database cleanup may be missed
   - Mitigation: Explicit cleanup tests, verify connections close
4. **Circular Dependencies**: Manual order may reveal circular deps
   - Mitigation: Clean architecture prevents this, but watch for it

**Medium-High Risk**: This changes core initialization - must be thorough

**Rollback**: Git revert if issues found, Wire code in backup branch

## Security Considerations

- JWT secret validation still enforced
- Database connection cleanup verified
- No new security risks (same initialization, just manual)

## Next Steps

**Dependencies**: Phase 03 must be complete (centralized errors working)

**Follow-up Tasks**:
- Phase 05: Simplify error handler logic (Manual DI working)
- Phase 06: Comprehensive testing and validation

**Transition Criteria**:
- All tests passing → Start Phase 05
- Server runs correctly → Safe to proceed
- Wire removed completely → Ready for final testing
