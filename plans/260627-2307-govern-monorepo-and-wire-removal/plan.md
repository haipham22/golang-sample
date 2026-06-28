---
title: "Govern Monorepo Restructure & Wire Removal"
description: "Complete migration: Restructure to monorepo with external imports, then remove Wire DI and replace govern/errors with custom error types"
status: done (Phase 07 rename deferred to merge)
priority: P1
effort: 40h
branch: main
tags: [monorepo, migration, govern, wire-removal, error-handling, clean-architecture]
created: 2026-06-27
blocks: []
---

# Govern Monorepo Restructure & Wire Removal - Complete Migration Plan

## Overview

Complete two-phase migration: (1) Restructure golang-sample repository into govern monorepo with library packages at root and sample app in examples/ directory using external imports, then (2) Remove Google Wire dependency injection and replace govern/errors with custom error types.

**Target State**: 
- Govern library at root (`github.com/haipham22/govern`)
- Sample app in examples/ directory (`examples/golang-sample/`)
- External import approach (NO go.work workspace)
- Manual dependency injection (NO Wire)
- Custom error types (NO govern/errors)

**Approach**: Sequential execution - Part 1 (monorepo) must complete before Part 2 (wire removal) begins.

---

## Migration Summary

**Part 1: Monorepo Restructuring** (Phases 01-07, 12h)
- Repository rename: golang-sample → govern
- Govern packages at root level
- Sample app at root as independent module
- External import with replace directive for local dev
- NO go.work workspace

**Part 2: Wire Removal & Error Management** (Phases 09-14, 24h)
- Remove Google Wire DI
- Implement manual dependency injection
- Replace govern/errors with custom error types
- Centralized error management system
- Comprehensive testing

---

## Phase Summary

### Part 1: Monorepo Restructuring (12h)

| Phase | Status | Duration | Priority | Dependencies |
|-------|--------|----------|----------|--------------|
| [Phase 01: Repository Preparation](phase-01-repository-preparation.md) | done | 2h | P1 | None |
| [Phase 02: Merge Govern Packages](phase-02-merge-govern-packages.md) | done | 3h | P1 | Phase 01 |
| [Phase 03: Move Sample Application](phase-03-move-sample-application.md) | done | 2h | P1 | Phase 02 |
| [Phase 04: Root Configuration Update](phase-04-root-configuration-update.md) | done | 2h | P1 | Phase 03 |
| [Phase 05: Documentation Migration](phase-05-documentation-migration.md) | done | 2h | P1 | Phase 04 |
| [Phase 06: Validation & Testing](phase-06-validation-testing.md) | done | 1h | P1 | Phase 05 |
| [Phase 07: Repository Rename & Merge](phase-07-repository-rename-merge.md) | deferred | 0h | P1 | Phase 06 |

**Note**: Interactive Generator (4h) separated into future plan - will be implemented separately with detailed design.

### Part 2: Wire Removal & Error Management (24h)

| Phase | Status | Duration | Priority | Dependencies |
|-------|--------|----------|----------|--------------|
| [Phase 08: Setup Wire Removal Environment](phase-08-setup-wire-removal.md) | done | 2h | P1 | Phase 07 |
| [Phase 09: Custom Error Types](phase-09-custom-error-types.md) | done | 4h | P1 | Phase 08 |
| [Phase 10: Centralized Error Management](phase-10-centralized-error-management.md) | done | 6h | P1 | Phase 09 |
| [Phase 11: Manual DI Implementation](phase-11-manual-di-implementation.md) | done | 5h | P1 | Phase 10 |
| [Phase 12: Error Handler Refactoring](phase-12-error-handler-refactoring.md) | done | 3h | P1 | Phase 11 |
| [Phase 13: Wire Removal Testing](phase-13-wire-removal-testing.md) | done | 4h | P1 | Phase 12 |

**Total Estimated Time**: 40 hours (12h monorepo + 28h wire removal)

---

## Critical Success Criteria

### Part 1 Must Have (P0)
- [ ] All govern packages compile without errors
- [ ] Sample application compiles and all tests pass
- [ ] Sample app imports govern as external dependency successfully
- [ ] Git history preserved for govern packages (git fast-export/import)
- [ ] Git history preserved for sample app (git mv)
- [ ] NO go.work workspace file created
- [ ] External import configured with replace directive
- [ ] CI/CD workflows pass for both govern library and sample app

### Part 2 Must Have (P0)
- [ ] All Wire code removed from codebase
- [ ] No govern/errors imports remaining
- [ ] Manual DI constructors implemented
- [ ] Centralized error management operational
- [ ] All tests passing (≥80% coverage)
- [ ] All 7 production files migrated (handler, controller, service, validator)
- [ ] No compilation errors
- [ ] No runtime errors

---

## Architecture Decisions

### External Import Strategy (CRITICAL)

**Decision**: Use external import approach, NOT Go workspace

**Rationale**:
- Sample app is independent module unrelated to root
- Clear separation between library and sample
- Production uses published govern package
- Local development uses replace directive

**Implementation**:
```go
// examples/golang-sample/go.mod
module github.com/haipham22/golang-sample

require github.com/haipham22/govern v1.0.0

replace github.com/haipham22/govern => ../../  // Local development only
```

### Module Paths

**Repository**: `github.com/haipham22/govern` (renamed from golang-sample)

**Module Paths**:
- Govern library: `github.com/haipham22/govern` (root module)
- Sample app: `github.com/haipham22/golang-sample` (in examples/ directory)
- Generator: `github.com/haipham22/govern/scripts/generate-project` (CLI tool)

**Directory Structure**:
```
govern/                              # Repository root
├── go.mod                           # Govern module
├── http/, config/, database/, etc.  # Govern packages
├── examples/                         # Samples directory
│   └── golang-sample/               # Sample application
│       ├── cmd/
│       │   ├── serverd.go           # HTTP server (zap logger)
│       │   ├── grpcd.go             # gRPC server
│       │   └── workerd.go           # Job worker
│       ├── internal/
│       │   ├── bootstrap/            # Manual DI wiring (thay Wire)
│       │   │   ├── app.go            # Main DI constructor
│       │   │   ├── logger.go         # Logger setup
│       │   │   ├── database.go       # Database setup
│       │   │   ├── http.go           # HTTP server setup
│       │   │   └── worker.go         # Worker setup
│       │   ├── usecase/             # Application Layer (Middle - Use Cases)
│       │   │   ├── auth/             # Auth use case
│       │   │   │   ├── service.go    # AuthRepository interface + impl
│       │   │   │   ├── impl.go       # Use case implementations
│       │   │   │   ├── dto.go        # Request/Response DTOs
│       │   │   │   └── mocks/        # Mocks for testing
│       │   │   ├── product/          # Product use case
│       │   │   │   ├── service.go    # ProductRepository interface
│       │   │   │   ├── impl.go
│       │   │   │   ├── dto.go
│       │   │   │   └── mocks/
│       │   │   └── user/             # User use case
│       │   │       ├── service.go    # UserRepository interface
│       │   │       ├── impl.go
│       │   │       ├── dto.go
│       │   │       └── mocks/
│       │   ├── domain/              # Domain Layer (Inner - Business Rules)
│       │   │   ├── user.go           # Business entities
│       │   │   ├── product.go        # Business entities
│       │   │   └── errors.go        # Domain-specific errors
│       │   ├── repository/          # Infrastructure Layer (Outer - Implementations)
│       │   │   ├── helper.go
│       │   │   ├── postgres/        # GORM PostgreSQL implementations
│       │   │   ├── redis/           # Redis implementations
│       │   │   └── kafka/           # Kafka implementations
│       │   ├── handler/             # Delivery Layer (Outer)
│       │   │   ├── rest/            # HTTP handlers
│       │   │   ├── grpc/            # gRPC handlers
│       │   │   ├── job/             # Job workers
│       │   │   └── kafka/           # Event consumers
│       │   ├── errors/              # Custom error types
│       │   └── middleware/          # HTTP/gRPC middleware
│       ├── docs/
│       │   ├── SPEC.md              # Specification
│       │   ├── HLD.md               # High-level design
│       │   └── ROADMAP.md           # Project roadmap
│       └── go.mod                   # Sample app module
└── scripts/, templates/             # Generator and templates
```

**Clean Architecture Layers** (Dependency Rule: Dependencies point inward):
```
┌──────────────────────────────────────────────────────────────┐
│  OUTER LAYER (Frameworks & Drivers)                            │
│  ├── handler/{rest,grpc,job,kafka}/  ← External interfaces   │
│  ├── repository/{postgres,redis,kafka}/ ← GORM, Redis, etc.   │
│  └── bootstrap/                        ← Manual DI wiring       │
└──────────────────────────────────────────────────────────────┘
                           ↓ (depends on)
┌──────────────────────────────────────────────────────────────┐
│  MIDDLE LAYER (Application Business Rules)                   │
│  ├── usecase/{auth,product,user}/     ← Use case logic       │
│  │   ├── service.go                 ← Repository interfaces  │
│  │   ├── impl.go                    ← Use case impls       │
│  │   └── dto.go                     ← Request/Response    │
└──────────────────────────────────────────────────────────────┘
                           ↓ (depends on)
┌──────────────────────────────────────────────────────────────┐
│  INNER LAYER (Enterprise Business Rules)                      │
│  ├── domain/{user.go,product.go}     ← Business entities    │
│  └── domain/errors.go               ← Domain errors        │
└──────────────────────────────────────────────────────────────┘
```

**Data Flow** (Unidirectional):
```
External Request → handler/ → usecase/ → domain/ (interfaces)
                                               ↓
                                         repository/ (implementations)
                                               ↓
                                         domain/ (entities)
```

**Key Design Decisions**:
- **cmd/**: Uses zap for structured logging (not govern/logger)
- **handler/**: Organized by transport type (REST, gRPC, Jobs, Events)
- **domain/**: Inner layer - pure business rules, no external dependencies
- **usecase/**: Middle layer - use case implementations, depends on domain interfaces
- **repository/**: Outer layer - GORM, Redis, external systems
- **errors/**: Custom error types replacing govern/errors
- **docs/**: SPEC.md, HLD.md, ROADMAP.md following affiliate-tracking approach

### Folder Structure Rules (Critical - Must Follow)

**Based on bxcodec/go-clean-arch + Clean Architecture principles:**

#### **domain/** (INNER LAYER - Enterprise Business Rules)
**Contains:**
- Pure business entities (structs with business logic methods)
- Domain-specific errors
- Value objects

**MUST NOT:**
- ❌ Define interfaces (interfaces defined by consuming layer)
- ❌ Import external packages (no frameworks, no databases)
- ❌ Depend on usecase/, repository/, handler/

**Rules:**
- Flat structure (NO subdirectories like domain/model/)
- Files use snake_case: user.go, product.go, errors.go
- Pure Go - no framework dependencies

**Example:**
```go
// internal/domain/user.go
package domain

type User struct {
    ID       int64
    Name     string
    Email    string
}

func (u *User) Validate() error { ... }
```

---

#### **usecase/** (MIDDLE LAYER - Application Business Rules)
**Contains:**
- Use case implementations
- Repository interfaces (DEFINED by usecase, NOT domain)
- DTOs (Data Transfer Objects)
- Mocks for testing

**MUST NOT:**
- ❌ Depend on handler/, repository/ implementations
- ❌ Know about frameworks (Echo, gRPC, GORM)

**Rules:**
- Organized by business capability: usecase/auth/, usecase/product/, usecase/user/
- Each usecase folder: service.go (interface + impl), impl.go (use cases), dto.go, mocks/
- Interfaces defined HERE (service.go), not in domain/

**Example:**
```go
// internal/usecase/auth/service.go
package auth

import "github.com/haipham22/golang-sample/internal/domain"

// AuthRepository interface defined by auth usecase
type AuthRepository interface {
    FindByEmail(email string) (domain.User, error)
    Create(user *domain.User) error
}

type Service struct {
    authRepo AuthRepository
}

func NewService(repo AuthRepository) *Service { ... }
```

---

#### **repository/** (OUTER LAYER - Frameworks & Drivers)
**Contains:**
- Database implementations (PostgreSQL, MongoDB, Redis)
- External service clients (Kafka, S3, HTTP clients)
- Infrastructure code

**MUST:**
- ✅ Implement interfaces from usecase/
- ✅ Use frameworks (GORM, Redis client, etc.)

**MUST NOT:**
- ❌ Define business logic
- ❌ Be imported by domain/

**Rules:**
- Nested by technology: repository/postgres/, repository/redis/, repository/kafka/
- Each implements interfaces from usecase/

**Example:**
```go
// internal/repository/postgres/auth.go
package postgres

import (
    "github.com/haipham22/golang-sample/internal/usecase/auth"
    "gorm.io/gorm"
)

type authRepository struct {
    db *gorm.DB
}

// Implements auth.AuthRepository interface
func (r *authRepository) FindByEmail(email string) (domain.User, error) { ... }
```

---

#### **handler/** (OUTER LAYER - Delivery Mechanisms)
**Contains:**
- HTTP handlers (Echo framework)
- gRPC handlers
- Job workers
- Kafka consumers
- Middleware

**MUST:**
- ✅ Use usecase/ services
- ✅ Handle framework-specific concerns (binding, validation, routing)

**MUST NOT:**
- ❌ Contain business logic
- ❌ Access database directly (use repository/)

**Rules:**
- Organized by transport: handler/rest/, handler/grpc/, handler/job/, handler/kafka/
- Thin layer - delegate to usecase/ immediately

**Example:**
```go
// internal/handler/rest/auth.go
package rest

import "github.com/haipham22/golang-sample/internal/usecase/auth"

type authHandler struct {
    authService *auth.Service
}

func (h *authHandler) Login(c echo.Context) error {
    var req dto.LoginRequest
    if err := c.Bind(&req); err != nil { ... }
    
    // Delegate to usecase
    token, err := h.authService.Login(c.Request().Context(), req)
    if err != nil { ... }
    
    return c.JSON(200, resp)
}
```

---

#### **bootstrap/** (Manual DI - Replacing Wire)
**Contains:**
- Dependency injection constructors
- Setup functions (logger, database, HTTP server)

**MUST:**
- ✅ Wire all dependencies explicitly
- ✅ Return cleanup functions for resources

**Rules:**
- app.go: Main constructor
- database.go: Database setup with cleanup
- http.go: HTTP server setup
- logger.go: Logger setup

**Example:**
```go
// internal/bootstrap/app.go
package bootstrap

func NewApp(cfg *config.Config) (*App, func(), error) {
    // 1. Logger
    logger := zapLogger()
    
    // 2. Database with cleanup
    db, cleanup, err := postgresDB(cfg.Database)
    
    // 3. Repositories
    authRepo := postgres.NewAuthRepository(db)
    
    // 4. Use cases
    authService := auth.NewService(authRepo)
    
    // 5. Handlers
    authHandler := rest.NewAuthHandler(authService)
    
    // 6. HTTP Server
    server := rest.NewServer(authHandler, cfg.HTTP)
    
    return &App{server}, cleanup, nil
}
```

---

### Dependency Rules (CRITICAL)

**Allowed dependencies (arrow means "can import"):**

```
handler/     → usecase/
usecase/      → domain/
repository/   → domain/ + usecase/ interfaces
bootstrap/    → everything
domain/       → NOTHING (no external dependencies)
```

**Forbidden dependencies:**

```
domain/    ✗ usecase/
domain/    ✗ repository/
domain/    ✗ handler/
usecase/   ✗ handler/
usecase/   ✗ repository/ implementations
handler/   ✗ repository/
```

---

### Interface Placement Rule (bxcodec pattern)

**Interfaces defined by CONSUMING layer:**

```
usecase/auth/service.go    → Defines AuthRepository interface
repository/postgres/       → Implements AuthRepository interface
```

**NOT:**

```
domain/repository.go       → ❌ WRONG: Don't define interfaces in domain
```

**Rationale:**
- Interface defined by who USES it (Dependency Inversion Principle)
- Domain doesn't know about repositories
- Use case defines what it needs

### Wire Removal Scope

**Production Files Affected** (with new clean architecture):
1. `examples/golang-sample/internal/handler/rest/handler.go` (9 governerrors usages)
2. `examples/golang-sample/internal/handler/rest/controllers/auth/auth.go` (2 usages)
3. `examples/golang-sample/internal/usecase/auth/impl.go` (9 usages) - Previously: service/auth/impl.go
4. `examples/golang-sample/internal/usecase/auth/dto.go` (2 usages) - Validation moved here
5. Test files (auth_test.go, service_test.go, usecase_test.go, etc.)

**Directory Migration** (Phase 03):
```
Current                            → Target
internal/model/                   → internal/domain/ (flat: user.go, product.go, errors.go)
internal/storage/                 → internal/usecase/ (repository interfaces defined here)
internal/orm/                     → internal/repository/postgres/ (implementations)
internal/service/auth/           → internal/usecase/auth/ (service.go + impl.go + dto.go)
internal/schemas/                → internal/usecase/auth/dto.go (validation)
internal/validator/              → internal/usecase/auth/dto.go (merged)
```

**Total governerrors usages**: 44 (25 excluding tests)

**Bootstrap Process** (Manual DI replacing Wire):
- Phase 11: Creates bootstrap constructors in internal/bootstrap/
- Eliminates Wire code generation step
- Explicit dependency construction in code
- Files: `bootstrap/app.go`, `bootstrap/database.go`, `bootstrap/http.go`
- Removes: `wire.go`, `wire_gen.go`

**Working Directory**: Part 2 phases operate in `examples/golang-sample/` directory

---

## Key Dependencies

### Part 1 Dependencies
- Phase 02 must wait for Phase 01 (branch creation, directory setup)
- Phase 03 must wait for Phase 02 (govern packages merged first)
- Phase 04 must wait for Phase 03 (examples/golang-sample/ moved)

### Part 2 Dependencies  
- Phase 08 must wait for Phase 07 (monorepo restructuring complete)
- Phase 09 must wait for Phase 08 (baseline established)
- Phase 10 must wait for Phase 09 (custom errors available)
- Phase 11 must wait for Phase 10 (error management operational)
- Phase 12 must wait for Phase 11 (bootstrap/manual DI implemented)
- Phase 13 must wait for Phase 12 (refactoring complete)
- Phase 14 must wait for Phase 13 (refactoring complete)

**Critical**: Part 2 cannot start until Part 1 fully complete

---

## Risk Assessment

### Part 1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Git history loss during merge | Medium | High | Use git fast-export/import, verify history |
| Import path conflicts | Low | High | Current imports already use github.com/haipham22/govern |
| External dependency resolution | Low | Medium | Use replace directive for local dev |
| Repository rename side effects | Low | Low | Update remote URL after GitHub rename |

### Part 2 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Circular dependencies in manual DI | Medium | High | Layer-by-layer migration from storage |
| Error handling changes break API | Medium | High | Comprehensive tests before/after |
| Missing service layer migration | High | High | Updated scope to include all 7 files |
| Runtime errors from missing deps | Medium | Medium | Validation phase + comprehensive testing |

---

## File Ownership Matrix

### Part 1: Monorepo Restructuring

| Phase | Files Modified | Files Created | Files Deleted | Owner |
|-------|----------------|---------------|---------------|-------|
| Phase 01 | .gitignore | examples/, templates/, scripts/generate-project/ | None | planner |
| Phase 02 | go.mod, go.sum | govern/* (15+ packages) | None | planner |
| Phase 03 | Multiple | examples/golang-sample/* | cmd/, internal/, orm/, schemas/, validator/ | planner |
| Phase 04 | .github/workflows/*, Makefile, README.md | .github/workflows/test.yml | scripts/generate-swagger.sh | planner |
| Phase 05 | scripts/generate-project/* | templates/*, scripts/generate-project/* | None | planner |
| Phase 06 | docs/*, CLAUDE.md | docs/quickstart.md, docs/packages/*, docs/examples/* | None | planner |
| Phase 07 | Remote repository URL | Git tag v0.1.0 | None | planner |

### Part 2: Wire Removal & Error Management

| Phase | Files Modified | Files Created | Files Deleted | Owner |
|-------|----------------|---------------|---------------|-------|
| Phase 08 | examples/golang-sample/go.mod, mise.toml | validation scripts, baseline reports | None | planner |
| Phase 09 | examples/golang-sample/internal/model/errors.go | examples/golang-sample/internal/errors/* | None | planner |
| Phase 10 | examples/golang-sample/internal/handler/rest/*.go | examples/golang-sample/internal/errors/{helpers,response,logging}.go | None | planner |
| Phase 11 | examples/golang-sample/internal/handler/rest/*.go | examples/golang-sample/internal/bootstrap/* | examples/golang-sample/internal/handler/rest/wire.go, wire_gen.go | planner |
| Phase 12 | examples/golang-sample/cmd/serverd.go | None | None | planner |
| Phase 13 | All test files | Test reports | None | tester |

**Working Directory**: Part 2 phases operate in `examples/golang-sample/` directory

---

## Testing Strategy

### Part 1 Testing

**Unit Testing**:
- Govern packages: `go test ./http/... ./database/... ./config/...`
- Sample app: `cd golang-sample && go test ./...`

**Integration Testing**:
- External import: Test sample app imports govern as external dependency
- Generator: Test generation to /tmp/, verify compiles

**End-to-End Testing**:
- Sample app startup: `cd golang-sample && go run cmd/serverd.go`
- Generated project: Generate, build, run, verify imports govern

### Part 2 Testing

**Unit Testing**:
- Custom error types: `go test ./internal/errors/...`
- Manual DI: `go test ./internal/handler/rest/...`
- Error management: `go test ./internal/...`

**Integration Testing**:
- API error responses: Compare before/after HTTP responses
- Error logging: Verify request ID tracking
- DI construction: Test all dependency combinations

**Regression Testing**:
- Full test suite: `go test ./... -v -cover`
- Race detector: `go test -race ./...`
- Performance benchmarks: Compare startup times

---

## Success Metrics

### Part 1 Success

**Functional**:
- All tests pass (govern + sample app)
- Generator creates valid projects
- CI/CD workflows pass without errors
- No broken imports or missing dependencies

**Migration**:
- Git history intact (verified with git log)
- No merge conflicts in feature branch
- Clean commit history (single atomic migration)

### Part 2 Success

**Functional**:
- All tests passing with same coverage
- Server starts and runs correctly
- API endpoints work as before
- Cleanup function validated

**Quality**:
- No Wire code remaining
- No govern/errors imports
- Manual DI working correctly
- Error handling centralized

---

## Rollback Strategy

### Part 1 Rollback

**If migration fails during implementation**:
1. Delete feature branch: `git branch -D feat/monorepo-migration`
2. Checkout main: `git checkout main`
3. No changes to main branch (all work in feature branch)

**If migration fails after merge**:
1. Revert merge commit: `git revert -m 1 <merge-sha>`
2. Force push revert (if already pushed)
3. Restore repository name on GitHub (if renamed)

### Part 2 Rollback

**If wire removal fails**:
1. Git revert to commit before Phase 09
2. Restore Wire code from backup branch
3. Re-run tests to verify

**If error migration fails**:
1. Revert specific phase commit
2. Fix issues in custom error types
3. Retry phase

**Data Loss Risk**: None (no database migrations, only code restructuring)

---

## Next Steps

### Immediate Actions

1. **Start Part 1**: Begin with Phase 01 (Repository Preparation)
2. **Execute Part 1 Phases 01-08**: Complete monorepo restructuring
3. **Validate Part 1**: Ensure all success criteria met
4. **Start Part 2**: Begin with Phase 09 (Setup Wire Removal)
5. **Execute Part 2 Phases 09-14**: Complete wire removal and error migration
6. **Final Validation**: Comprehensive testing of entire system

### Implementation Sequence

```
Part 1: Monorepo (16h)
├── Phase 01: Repository Preparation (2h)
├── Phase 02: Merge Govern Packages (3h)
├── Phase 03: Move Sample Application (2h)
├── Phase 04: Root Configuration Update (2h)
├── Phase 05: Interactive Generator (4h)
├── Phase 06: Documentation Migration (2h)
├── Phase 07: Validation & Testing (1h)
└── Phase 08: Repository Rename & Merge (0h)

Part 2: Wire Removal (24h) [BLOCKED until Part 1 complete]
├── Phase 09: Setup Wire Removal Environment (2h)
├── Phase 10: Custom Error Types (4h)
├── Phase 11: Centralized Error Management (6h)
├── Phase 12: Manual DI Implementation (5h)
├── Phase 13: Error Handler Refactoring (3h)
└── Phase 14: Wire Removal Testing (4h)
```

---

## Unresolved Questions

**None** - All critical issues resolved in audit and consolidation:

✅ External import strategy confirmed (no go.work)
✅ Wire removal scope updated (7 files, not 2)
✅ Module paths clarified (govern at root, golang-sample at root)
✅ Working directory specified (golang-sample/ for Part 2)
✅ Effort estimates realistic (40h total)
✅ Phase dependencies correct (Part 2 blocked by Part 1)

---

**Plan Status**: Ready for implementation
**Estimated Completion**: 40 hours
**Risk Level**: Medium (mitigated with rollback strategies)
**Red Team Review**: PENDING
