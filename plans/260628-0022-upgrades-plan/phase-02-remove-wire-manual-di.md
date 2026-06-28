---
title: "Phase 02: Remove Wire - Manual Dependency Injection"
description: "Remove Google Wire and replace with manual dependency injection"
status: completed
priority: P1
effort: 4h
branch: main
tags: [wire, di, refactoring, manual-di]
created: 2026-06-28
---

# Phase 02: Remove Wire - Manual Dependency Injection

**Status:** `completed` ✅ — done in commit `1672d69`. Implementation landed manual DI in `internal/handler/rest/di.go` (the phase doc below references `new.go`/`bootstrap/`, which is stale).  
**Priority:** **P1** (Do After Go 1.26)  
**Risk Level:** Low  
**Dependencies:** Phase 01 (Go 1.26 Upgrade)

---

## Overview

Remove Google Wire compile-time dependency injection and replace with manual DI in `internal/handler/rest/new.go`. This simplifies the build process, reduces code generation overhead, and improves code readability with explicit dependency construction.

### Why Remove Wire?

| Factor | Wire | Manual DI |
|--------|------|-----------|
| **Build Complexity** | Requires code generation | No code generation |
| **Readability** | Generated files obscure flow | Explicit construction in one file |
| **Debugging** | Hard to debug generated code | Easy to trace |
| **Build Time** | +wire generation step | Faster builds |
| **Maintenance** | Regenerate on changes | Direct code editing |

### Current Wire Setup

**Wire-generated function** (`wire_gen.go`):
```go
func New(log *zap.SugaredLogger, port int64, appConfig *config.EnvConfigMap) (http.Server, func(), error)
```

**Wire injector** (`wire.go`):
```go
func New(...) {
    panic(wire.Build(
        wire.NewSet(provideAuthConfig),
        wire.NewSet(provideDB),
        wire.NewSet(userRepo.New),
        wire.NewSet(provideAuthService),
        wire.NewSet(authctrl.New),
        wire.NewSet(healthctrl.New),
        wire.NewSet(NewHandler),
        echo.New,
    ))
}
```

**Will be replaced with:** Manual DI function with same signature

---

## Implementation Steps

### Step 1: Create Manual DI Function (1.5h)

**Replace Wire-generated function in:** `internal/handler/rest/new.go`

```go
package rest

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	governhttp "github.com/haipham22/govern/http"

	authctrl "golang-sample/internal/handler/rest/controllers/auth"
	healthctrl "golang-sample/internal/handler/rest/controllers/health"
	authservice "golang-sample/internal/service/auth"
	userRepo "golang-sample/internal/storage/user"
	"golang-sample/pkg/config"
	"golang-sample/pkg/postgres"
)

// New creates HTTP handler with manual DI (replaces Wire-generated function)
// Returns: server, cleanup function, error
func New(
	log *zap.SugaredLogger,
	port int64,
	appConfig *config.EnvConfigMap,
) (governhttp.Server, func(), error) {
	// 1. Database
	db, cleanup, err := postgres.NewGormDB(appConfig.Postgres.DSN)
	if err != nil {
		return nil, nil, err
	}

	// 2. Repositories
	storage := userRepo.New(log, db)

	// 3. Services
	jwtExpiration := 72 * time.Hour
	service := authservice.NewAuthService(
		log,
		storage,
		appConfig.API.Secret,
		jwtExpiration,
	)

	// 4. Controllers
	authController := authctrl.New(service)
	healthController := healthctrl.New(db)

	// 5. Echo instance
	e := echo.New()

	// 6. HTTP Handler
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

### Step 2: Remove Wire Files (5 min)

**Delete files:**
```bash
rm internal/handler/rest/wire.go
rm internal/handler/rest/wire_gen.go
```

**Remove Wire imports** (if any):
```bash
grep -r "github.com/google/wire" internal/
# Should return empty after Wire removal
```

### Step 3: Update go.mod (5 min)

**Remove Wire dependency:**
```bash
mise exec -- go mod tidy
```

**Verify Wire removed:**
```bash
grep "wire" go.mod
# Should return empty
```

### Step 4: Update Tests (30 min)

**Test files using Wire need updates:**

```go
// Before: Using Wire-generated constructor
handler := rest.New(log, port, cfg)

// After: Using manual constructor
handler := bootstrap.NewHandler(log, port, cfg)
```

**Files to update:**
- `internal/handler/rest/handler_test.go`
- Any integration tests using Wire

### Step 5: Build & Verify (30 min)

**Commands:**
```bash
# Build
mise exec -- go build ./...

# Run tests
mise exec -- go test ./...
mise exec -- go test -race ./...
mise exec -- go test -cover ./...

# Lint
mise exec -- golangci-lint run
```

### Step 6: Update Documentation (30 min)

**Update files:**
- `CLAUDE.md` - Remove Wire references
- `README.md` - Remove Wire from tech stack table
- `docs/system-architecture.md` - Update DI section (if exists)

**Remove from CLAUDE.md:**
- Wire dependency injection references
- Wire build commands
- Wire-related setup instructions

**Add to CLAUDE.md (if not present):**
```markdown
## Dependency Injection

**Manual DI** - Explicit dependency construction in `internal/handler/rest/new.go`

The `New` function manually constructs all dependencies:
```go
func New(log, port, config) (server, cleanup, error)
```

Benefits:
- No code generation needed
- Explicit dependency flow
- Easy to trace and debug
- Faster build times
```

---

## Success Criteria

- ✅ Wire removed from go.mod
- ✅ Wire files deleted (wire.go, wire_gen.go)
- ✅ Manual DI working in `internal/handler/rest/new.go`
- ✅ All tests passing
- ✅ Build succeeds without Wire
- ✅ Documentation updated

---

## Risk Assessment

### Risk Level: Low

**Potential Issues:**
1. **Manual DI errors** - Missing dependencies
   - **Mitigation:** Compile-time checking ensures all deps provided
   - **Testing:** Comprehensive test coverage validates wiring

2. **Runtime vs Compile-time** - Fewer compile-time checks
   - **Mitigation:** Go's type system still provides strong checks
   - **Testing:** Integration tests catch wiring issues

### Rollback Strategy

```bash
# Restore Wire files from git
git checkout internal/handler/rest/wire.go
git checkout internal/handler/rest/wire_gen.go

# Re-add Wire to go.mod
mise exec -- go get github.com/google/wire@latest
mise exec -- go mod tidy
```

---

## Related Code Files

### Files to Create
- `internal/handler/rest/new.go` - Manual DI constructor

### Files to Delete
- `internal/handler/rest/wire.go` - Wire injector definition
- `internal/handler/rest/wire_gen.go` - Generated Wire code

### Files to Update
- `go.mod` - Remove Wire dependency
- `cmd/serverd.go` - Update import path (if needed)
- `CLAUDE.md` - Update DI documentation
- `docs/system-architecture.md` - Update architecture docs
- `README.md` - Remove Wire from tech stack

---

## Dependency Graph

```
Manual DI Construction Order (replaces Wire):
    Config
        ↓
    Database (postgres.NewGormDB)
        ↓
    Repositories (user.New)
        ↓
    Services (authservice.NewAuthService)
        ↓
    Controllers (authctrl.New, healthctrl.New)
        ↓
    Echo Instance (echo.New)
        ↓
    HTTP Handler (NewHandler)
        ↓
    Server (governhttp.Server)
```

**Key Changes from Wire:**
- ✅ No code generation needed
- ✅ Explicit dependency flow
- ✅ Same signature as Wire-generated `New` function
- ✅ Returns cleanup function for graceful shutdown

---

## Testing Strategy

### Unit Tests
- Verify bootstrap constructor creates valid Handler
- Test dependency injection correctness
- Validate cleanup function works

### Integration Tests
- Test full HTTP request flow
- Verify all dependencies wired correctly
- Validate graceful shutdown with cleanup

### Build Tests
- Ensure no Wire generation needed
- Verify clean build without code generation
- Test build time improvement

---

## Performance Impact

### Expected Improvements
- **Build Time:** -5-10% (no wire generation step)
- **Binary Size:** No change (same code generated)
- **Runtime:** No change (identical dependency graph)

### Measurement
```bash
# Before: With Wire
time mise exec -- go build ./...

# After: Without Wire
time mise exec -- go build ./...

# Compare build times
```

---

## Migration Notes

### Wire → Manual DI Mapping

| Wire Concept | Manual DI Equivalent |
|-------------|---------------------|
| `wire.NewSet()` | Direct function calls |
| `wire.Build()` | Sequential construction |
| `//go:generate wire` | No generation needed |
| `wire_gen.go` | Explicit code in bootstrap/ |

### Benefits Gained
1. **Explicit flow** - See exactly how dependencies connect
2. **No magic** - No code generation obscuring logic
3. **Faster builds** - Skip wire generation step
4. **Better debugging** - Trace through actual code
5. **Simpler onboarding** - New devs see explicit DI

---

## Next Steps

After Wire removal complete:
1. Proceed to **Phase 03: Database Layer Optimization**
2. Document manual DI patterns in CLAUDE.md
3. Update architecture documentation
4. Measure build time improvement

---

## References

### Research
- [Wire vs Manual DI Comparison](https://github.com/google/wire)

### Official Documentation
- [Google Wire GitHub](https://github.com/google/wire)
- [Clean Architecture DI Patterns](https://blog肩launclelhy.com/go-dependency-injection)

### Project Documentation
- [CLAUDE.md](../../CLAUDE.md) - Development rules
- [System Architecture](../../docs/system-architecture.md) - Architecture patterns

---

**Phase Status:** ✅ Completed (commit `1672d69`, synced 2026-06-28)  
**Completion Target:** Week 1  
**Owner:** Development Team  
**Created:** 2026-06-28
