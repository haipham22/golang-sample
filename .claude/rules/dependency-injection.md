# Dependency Injection Rules (Manual DI)

**Rules for manual (hand-written) dependency injection in the `bootstrap/` composition root — the replacement for Google Wire.**

---

## Overview

- **Pattern:** Constructor injection wired explicitly in a single composition root (`internal/bootstrap/`)
- **Convention:** every component exposes `NewX(deps...)` and resources return `(T, cleanup, error)`
- **Migration target:** replaces Google Wire (`internal/handler/rest/wire.go` + `wire_gen.go`)
- **Folder structure & placement:** see [clean-architecture.md](clean-architecture.md) → *bootstrap/*

**Why manual DI over Wire:**
- ✅ No codegen step — `go build` just works (no `wire generate` before build)
- ✅ Readable stack traces — real calls, not generated glue
- ✅ Explicit wiring — every dependency visible, debuggable with a debugger
- ✅ Zero magic — easier onboarding, no provider-set mental model
- ⚠️ Tradeoff: more boilerplate than Wire (acceptable for this project's size)

**Core rules:**
- ✅ Wire everything in ONE place: the composition root (`bootstrap/app.go`)
- ✅ Inject via constructors (`NewX`), never via globals or setters
- ✅ Return `cleanup` for every resource that holds OS handles (DB, listeners, files)
- ✅ Inject **interfaces** so components are mockable (see [mockery.md](mockery.md))
- ❌ Never use a service locator, `sync.Once` global singleton, or package-level `var` for wiring
- ❌ Never construct a dependency twice — build once, pass it down

---

## Composition Root

**All wiring lives in `internal/bootstrap/`. The entry point (`cmd/serverd.go`) loads config, calls the root, defers cleanup, and runs.**

```go
// cmd/serverd.go — entry point: load config, build app, run
RunE: func(cmd *cobra.Command, _ []string) error {
    cfgFile, _ := cmd.Flags().GetString("config")
    log := bootstrap.NewLogger()                 // construct logger (no global)
    cfg, err := config.LoadConfig(cfgFile, log)  // load config from --config (no global)
    if err != nil {
        return err
    }

    app, cleanup, err := bootstrap.NewApp(cfg, log)
    if err != nil {
        return err
    }
    defer cleanup()                              // teardown in reverse order

    return govern.Run(ctx, log, shutdownTimeout, app.Server)
}
```

**Rules:**
- ✅ `cmd/` is thin: parse flags → load config → call `bootstrap.NewApp` → run
- ✅ One root constructor: `bootstrap.NewApp(cfg, log) (*App, func(), error)`
- ✅ Config and logger are the only things the root receives from `cmd/`
- ❌ Never wire dependencies inside handlers, services, or `cmd/` directly
- ❌ Never call `NewX` for the same component from two places

---

## Constructor Injection

**Every component exposes a `NewX` constructor that takes its dependencies as explicit parameters:**

```go
// Storage implementation — takes raw infra deps
func New(log *zap.SugaredLogger, db *gorm.DB) Storage {
    return &repo{log: log, db: db}
}

// Service — takes the Storage INTERFACE (mockable)
func NewAuthService(log *zap.SugaredLogger, storage Storage, secret string, exp time.Duration) Service {
    return &authService{log: log, storage: storage, secret: secret, exp: exp}
}

// Controller — takes the Service interface
func New(svc authservice.Service) *Controller {
    return &Controller{svc: svc}
}
```

**Rules:**
- ✅ Parameters follow `context → deps → config → input` only at function level; constructors take `deps → config`
- ✅ Accept the **narrowest interface** a component needs (see [mockery.md](mockery.md) §1)
- ✅ Return the concrete type or interface the caller expects
- ❌ Never expose unexported fields via setters — set everything in the constructor
- ❌ Never reach into another package for a dependency — receive it

---

## The `(T, cleanup, error)` Triple

**Any component holding an OS handle (DB pool, network listener, file, message consumer) returns a cleanup function. This is already the project convention (`postgres.NewGormDB`, `restHandler.New`).**

```go
// GOOD — resource returns (value, cleanup, error)
func NewDatabase(cfg DatabaseConfig) (*gorm.DB, func(), error) {
    db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
    if err != nil {
        return nil, nil, fmt.Errorf("connect db: %w", err)
    }
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

    cleanup := func() { _ = sqlDB.Close() }   // capture and close
    return db, cleanup, nil
}

// GOOD — the root aggregates cleanups into ONE function
func NewApp(cfg *config.EnvConfigMap, log *zap.SugaredLogger) (*App, func(), error) {
    var cleanups []func()

    db, dbCleanup, err := NewDatabase(cfg.Postgres)
    if err != nil {
        return nil, nil, err                     // nothing opened yet, no cleanup
    }
    cleanups = append(cleanups, dbCleanup)

    // ... wire services/handlers ...

    cleanup := func() {
        for i := len(cleanups) - 1; i >= 0; i-- { // LIFO — reverse order
            cleanups[i]()
        }
    }
    return &App{Server: server, DB: db}, cleanup, nil
}
```

**Rules:**
- ✅ Cleanup runs in **reverse construction order** (last opened, first closed)
- ✅ Pure-value components (services, controllers) need **no** cleanup
- ✅ On early error, return `nil, nil, err` only if nothing was opened; otherwise clean up what you opened
- ❌ Never return a `*gorm.DB` without a cleanup — leaks the connection pool
- ❌ Never ignore the returned cleanup — `db, _, err := NewDatabase(...)` is a bug

---

## Wiring Order (Topological)

**Construct leaves first, then compose upward. Dependency direction: `domain ← storage ← service ← handler ← server`.**

```go
func NewApp(cfg *config.EnvConfigMap, log *zap.SugaredLogger) (*App, func(), error) {
    // 1. Infrastructure (resources with cleanup)
    db, cleanup, err := NewDatabase(cfg.Postgres)
    // ...

    // 2. Repositories / storage (implement service interfaces)
    userStorage := userrepo.New(log, db)

    // 3. Services (business logic — accept storage interfaces)
    authSvc := auth.NewAuthService(log, userStorage, cfg.API.Secret, jwtExpiration)

    // 4. Controllers / handlers (accept service interfaces)
    authCtrl := authctrl.New(authSvc)
    healthCtrl := healthctrl.New()

    // 5. HTTP server (accepts controllers)
    server, serverCleanup, err := rest.NewHandler(log, cfg.App.Port, authCtrl, healthCtrl)
    // ...

    return &App{Server: server}, aggregateCleanup(cleanup, serverCleanup), nil
}
```

**Rules:**
- ✅ Build in dependency order — a component's deps must exist before it
- ✅ Each layer only knows the interface of the layer below
- ❌ Never reference a layer two levels down (handler must not touch `db` directly)

---

## Config Injection

**Pass typed config values into constructors; validate and extract at the root — not deep in services.**

```go
// GOOD — root extracts the secret once, validates, passes a typed value
func NewApp(cfg *config.EnvConfigMap, log *zap.SugaredLogger) (*App, func(), error) {
    if cfg.API.Secret == "" {
        return nil, nil, errors.New("api.secret is required") // fail fast at startup
    }
    authSvc := auth.NewAuthService(log, userStorage, cfg.API.Secret, 72*time.Hour)
    // ...
}

// GOOD — service takes the typed value, not the whole config bag
func NewAuthService(log *zap.SugaredLogger, storage Storage, secret string, exp time.Duration) Service
```

**Rules:**
- ✅ Fail fast on missing/invalid config at the composition root
- ✅ Pass primitive/typed values (`string`, `time.Duration`, typed `Config` structs) into constructors
- ✅ Load config **once** in the composition root (from the `--config` path), then pass values down
- ❌ Avoid the `config.ENV` global and `zap.ReplaceGlobals` — load logger/config explicitly and inject
- ❌ Never pass the entire `*config.EnvConfigMap` into a service (couples it to all config)
- ❌ Never read env vars or call `viper` inside services or repositories

---

## Accept Interfaces (Mockability Link)

**Inject interfaces so each layer is independently testable with mockery:**

```go
// GOOD — service depends on Storage interface → mockable in service tests
func NewAuthService(log *zap.SugaredLogger, storage Storage, ...) Service

// BAD — service depends on concrete *repo → untestable without a real DB
func NewAuthService(log *zap.SugaredLogger, storage *userrepo.repo, ...) Service
```

This is the bridge between DI and testing — see [mockery.md](mockery.md) → *Designing Interfaces for Mockability*.

---

## Cleanup & Lifecycle

**The root's single `cleanup` runs on shutdown. `govern/graceful.Run` handles the HTTP server lifecycle; the deferred `cleanup()` handles resources.**

```go
// cmd/serverd.go
app, cleanup, err := bootstrap.NewApp(cfg, log)
if err != nil {
    return err
}
defer cleanup()                              // closes DB, listeners (reverse order)

return govern.Run(ctx, log, shutdownTimeout, app.Server)  // graceful HTTP shutdown
```

**Rules:**
- ✅ One `defer cleanup()` in `cmd/` — the root aggregates all sub-cleanups
- ✅ Order: HTTP server drains first (via `govern.Run`), then resources close
- ❌ Never scatter `defer db.Close()` across packages — centralize in the root cleanup

---

## Migrating from Wire

**Translate each Wire provider into an explicit constructor call in `bootstrap/app.go`:**

| Wire construct | Manual DI equivalent |
|----------------|----------------------|
| `wire.NewSet(provideDB)` | `db, cleanup, err := NewDatabase(cfg)` |
| `wire.NewSet(userRepo.New)` | `userStorage := userrepo.New(log, db)` |
| `wire.NewSet(provideAuthService)` | `authSvc := auth.NewAuthService(log, userStorage, secret, exp)` |
| `provideAuthConfig` (extract+validate) | inline at root with fail-fast check |
| `panic(wire.Build(...))` | explicit return of composed `*App` + cleanup |

**Steps:**
1. Create `internal/bootstrap/app.go` with `NewApp(cfg, log) (*App, func(), error)`
2. Translate each provider into an ordered constructor call (§Wiring Order)
3. Aggregate cleanups into one function (§The `(T, cleanup, error)` Triple)
4. Update `cmd/serverd.go` to call `bootstrap.NewApp` instead of `restHandler.New`
5. Delete `wire.go`, `wire_gen.go`; remove `google/wire` from `go.mod`
6. `mise exec -- go build ./...` to verify wiring compiles

---

## Best Practices & Pitfalls

**✅ DO:**
- Keep `bootstrap/` the only package that knows about all layers
- Make wiring order match the dependency DAG (topological)
- Return cleanup for every resource holding an OS handle
- Inject interfaces; validate config at the root

**❌ DON'T:**
- Use a service locator (`GetService("auth")`) or `init()` for wiring
- Hold dependencies in package-level `var`s (hidden globals)
- Construct the same dependency twice (two DB pools)
- Pass `context.Context` into constructors (it belongs in method calls, not construction)
- Swallow the cleanup return value

**Pitfalls:**
```go
// BAD — service locator / global singleton
var DB = MustConnect()           // wired via global; untestable, lifecycle hidden

// BAD — passing the whole config bag
func NewAuthService(cfg *config.EnvConfigMap) // service knows about postgres DSN it never uses

// BAD — ignoring cleanup
db, _, err := NewDatabase(cfg)   // leaks the pool

// BAD — circular wiring
svc := NewService(repo); repo := NewRepo(svc) // compile error or infinite init

// BAD — constructing twice
db1 := NewDatabase(cfg); db2 := NewDatabase(cfg) // two pools, two cleanups, confusion
```

---

## Quick Reference

```go
// Resource with cleanup
func NewDatabase(cfg Cfg) (*gorm.DB, func(), error)

// Pure component (no cleanup)
func NewAuthService(log Logger, storage Storage, secret string) Service

// Composition root
func NewApp(cfg *config.EnvConfigMap, log *zap.SugaredLogger) (*App, func(), error) {
    db, dbCleanup, err := NewDatabase(cfg.Postgres)
    if err != nil { return nil, nil, err }

    userStorage := userrepo.New(log, db)
    authSvc := auth.NewAuthService(log, userStorage, cfg.API.Secret)
    server, srvCleanup, err := rest.NewHandler(log, cfg.App.Port, authCtrl)
    if err != nil { dbCleanup(); return nil, nil, err }

    cleanup := func() { srvCleanup(); dbCleanup() } // reverse order
    return &App{Server: server}, cleanup, nil
}
```

| Concern | Rule |
|---------|------|
| Where to wire | `internal/bootstrap/app.go` only |
| Root signature | `NewApp(cfg, log) (*App, func(), error)` |
| Resource signature | `NewX(cfg) (T, func(), error)` |
| Cleanup order | reverse of construction (LIFO) |
| Inject | interfaces (narrowest needed) |
| Config | typed values, validated at root |
| Entry point | thin `cmd/` → `NewApp` → `defer cleanup` → `govern.Run` |

---

## References

- [clean-architecture.md](clean-architecture.md) → *bootstrap/* folder structure & placement
- [mockery.md](mockery.md) → *Designing Interfaces for Mockability*
- [infrastructure-rules.md](infrastructure-rules.md) → `NewApp` / `NewDatabase` / config loading
- [golang-types-values.md](golang-types-values.md) → constructor patterns
- Migration plan: `plans/260627-2307-govern-monorepo-and-wire-removal/plan.md`
