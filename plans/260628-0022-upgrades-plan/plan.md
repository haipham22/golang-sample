---
title: "Golang Sample Upgrades: Go 1.26, Echo v5, Database Layer"
description: "Comprehensive upgrade plan for Go 1.26, Echo v4→v5 migration, and database layer evaluation"
status: completed
priority: P2
effort: 16h
branch: main
tags: [upgrade, dependencies, go-1.26, echo-v5, database]
created: 2026-06-28
---

# Golang Sample Upgrades Plan

**Project:** golang-sample  
**Current State:** Go 1.26.4, Echo v5.2.1, GORM v1.31.1, manual DI, dev-gated DB migration  
**Target State:** Achieved 2026-06-28  

**Executive Summary:** Completed all upgrade tracks. Go 1.26.4 landed across monorepo, Wire was removed from sample app, database plan was narrowed to 3 real fixes and completed, and Echo v5 unblocked early after govern migrated first.

---

> ## 🔁 Status Reconciliation (synced 2026-06-28)
>
> Reconciled against the actual codebase (branch `feat/monorepo-migration`):
>
> | Phase | Plan status | **Actual** | Evidence |
> |-------|-------------|------------|----------|
> | 01 Go 1.26 | pending | ✅ **DONE** | bumped to `1.26.4` (mise.toml + both `go.mod` + CI workflows + Dockerfile); all tests race-clean; `go fix` ran; bonus: unused `wire` dep removed from govern |
> | 02 Wire removal | pending | ✅ **DONE** | commit `1672d69`; sample `go.mod` clean of wire; no `wire*.go`; manual DI in `internal/handler/rest/di.go` |
> | 03 DB optimization | pending | ✅ **DONE** | dev-gated AutoMigrate (User+Product); GORM debug→`cfg.App.Debug`; user repo→`apperrors` (pkg/errors removed); full `go test -race` green |
> | 04 Echo v5 | DEFERRED | ✅ **DONE** | both modules migrated to `echo/v5 v5.2.1` (govern first, then sample); full `go test -race` green |
>
> **Remaining doc drift (accepted):**
> - Historical implementation snippets in phase files still show original paths (`new.go`, `bootstrap/handler.go`, Echo v4 examples). Status headers now carry actual completion state.
> - Phase-file numbering diverges from this `plan.md`: two `phase-02-*` files exist (`database-optimization` + `remove-wire`); Echo v5 is `phase-03` in files vs `phase-04` here.

---

## Priority Matrix

| Upgrade | Risk | Value | Effort | Timeline | Status |
|---------|------|-------|--------|----------|--------|
| **Go 1.26** | LOW | HIGH | 4h | Immediate | ✅ **DONE** (1.26.4) |
| **Remove Wire** | LOW | MED | 4h | Immediate | ✅ **DONE** (commit `1672d69`) |
| **DB Layer** | LOW | MED | 6h | Immediate | ✅ **DONE** (3-task audit scope) |
| **Echo v5** | MED-HIGH | MED | 6h | Q3-Q4 2026 | ✅ **DONE** (v5.2.1, unblocked early) |

---

## Phase 01: Go 1.26 Upgrade (4 hours)

**Status:** `completed` ✅ — Go 1.26.4 across monorepo; tests/build/race clean.  
**Priority:** **P1** (Do First)  
**Risk Level:** Very Low  
**Dependencies:** None  

### Overview
Upgrade from Go 1.25.0 to Go 1.26 for performance improvements (10-40% GC reduction), enhanced security (post-quantum TLS), and improved tooling. Zero breaking changes for typical web applications.

### Key Changes
- **Green Tea GC:** 10-40% reduction in garbage collection overhead
- **Crypto Security:** Post-quantum TLS enabled by default
- **Tooling:** Revamped `go fix` command for code modernization
- **No Breaking Changes:** Full backward compatibility maintained

### Implementation Steps

#### Step 1: Update mise.toml (5 min)
```toml
[tools]
go = "1.26.0"
```

#### Step 2: Update go.mod (5 min)
```bash
mise use go@1.26
mise exec -- go mod tidy
```

#### Step 3: Run Code Modernizers (15 min)
```bash
mise exec -- go fix ./...
```

#### Step 4: Verify Dependencies (10 min)
```bash
mise exec -- go mod verify
mise exec -- go list -m all
```

#### Step 5: Run Test Suite (30 min)
```bash
mise exec -- go test ./...
mise exec -- go test -race ./...
mise exec -- go test -cover ./...
```

#### Step 6: Performance Validation (1h)
```bash
# Benchmark HTTP handlers
GODEBUG=gctrace=1 mise exec -- go test -bench=. ./...

# Compare GC pause times before/after
# Document performance improvements
```

#### Step 7: Update CI/CD (30 min)
- Update GitHub Actions Go version to 1.26
- Verify all workflows pass
- Update development documentation

#### Step 8: GODEBUG Migration Planning (30 min)
- Check for deprecated GODEBUG usage: `grep -r "GODEBUG" .`
- Document any legacy TLS settings
- Plan migration before Go 1.27 (6 months)

### Success Criteria
- ✅ All tests pass with Go 1.26
- ✅ No breaking changes detected
- ✅ Performance improvement measurable (GC overhead reduction)
- ✅ CI/CD pipeline updated and passing
- ✅ Development documentation updated

### Risk Assessment
- **Risk Level:** Very Low
- **Mitigation:** Go 1 promise of backward compatibility
- **Rollback:** Simple `mise install go@1.25.5` if issues arise

### Related Code Files
- `mise.toml` (update go version)
- `go.mod` (update go directive)
- `.github/workflows/*.yml` (CI/CD updates)

### Next Steps
After Go 1.26 upgrade complete → Proceed to Phase 02 (Database Optimization)

---

## Phase 02: Remove Wire - Manual Dependency Injection (4 hours)

**Status:** `completed` ✅ — done in commit `1672d69`. Manual DI lives in `internal/handler/rest/di.go` (not `new.go`/`bootstrap/` as described below).  
**Priority:** **P1** (Do After Go 1.26)  
**Risk Level:** Low  
**Dependencies:** Phase 01 (Go 1.26 Upgrade)

### Overview
Remove Google Wire compile-time dependency injection and replace with manual DI. This simplifies the build process, reduces code generation overhead, and improves code readability with explicit dependency construction.

### Why Remove Wire?

| Factor | Wire | Manual DI |
|--------|------|-----------|
| **Build Complexity** | Requires code generation | No code generation |
| **Readability** | Generated files obscure flow | Explicit construction |
| **Debugging** | Hard to debug generated code | Easy to trace |
| **Build Time** | +wire generation step | Faster builds |
| **Flexibility** | Compile-time only | Runtime + compile-time |

### Implementation Steps

#### Step 1: Create Manual Bootstrap (1.5h)
Create `internal/bootstrap/handler.go` with manual DI construction replacing Wire providers.

#### Step 2: Remove Wire Files (5 min)
Delete `wire.go` and `wire_gen.go` files.

#### Step 3: Update go.mod (5 min)
Remove Wire dependency with `go mod tidy`.

#### Step 4: Update Tests (30 min)
Update test files using Wire to use manual constructor.

#### Step 5: Build & Verify (30 min)
Build, test, and validate manual DI works correctly.

#### Step 6: Update Documentation (30 min)
Remove Wire references from CLAUDE.md, README.md, architecture docs.

### Success Criteria
- ✅ Wire removed from go.mod
- ✅ Manual DI working in `internal/bootstrap/`
- ✅ All tests passing
- ✅ Build succeeds without Wire
- ✅ Documentation updated

### Risk Assessment
- **Risk Level:** Low
- **Mitigation:** Go's type system provides compile-time checks
- **Rollback:** Restore Wire files from git if needed

### Related Code Files
- `internal/bootstrap/handler.go` (create)
- `internal/handler/rest/wire.go` (delete)
- `internal/handler/rest/wire_gen.go` (delete)
- `go.mod` (remove Wire)
- `CLAUDE.md` (update DI documentation)

### Next Steps
After Wire removal complete → Proceed to Phase 03 (Database Optimization)

---

## Phase 03: Database Layer Optimization (6 hours)

**Status:** `completed` ✅ — revised audit scope done: dev AutoMigrate, debug-driven GORM logging, user repo `apperrors`.  
**Priority:** **P1** (Do After Wire Removal)  
**Risk Level:** Low  
**Dependencies:** Phase 02 (Wire Removal)

### Overview
**RECOMMENDATION: Stay with GORM with optimizations.** Based on comprehensive analysis, GORM remains best fit for current project scale. Migration to pgx/sqlc has high cost (2-4 weeks) with marginal benefits for small-to-medium API.

### Analysis Summary

| Factor | GORM | pgx/v5 | sqlc |
|--------|------|--------|------|
| **Migration Cost** | None | High (2-3 weeks) | Very High (3-4 weeks) |
| **Performance** | Good (2-3x slower) | Excellent | Excellent |
| **Type Safety** | Good | Very Good | Best |
| **Team Productivity** | Highest | Medium | Medium |

### Decision Rationale
1. **Zero migration cost** - Architecture already clean
2. **Performance adequate** - 2-3x overhead acceptable for API scale  
3. **Team velocity** - Fastest development speed
4. **Clean architecture fit** - Already properly layered

### Implementation Steps

#### Step 1: Database Performance Analysis (1h)
```bash
# Enable query logging
gorm:logger:gorm_logger := logger.New(
  log.New(os.Stdout, "\r\n", 0),
  logger.Config{
    SlowThreshold: time.Second,
    LogLevel: logger.Info,
  },
)

# Analyze slow queries
# Check for N+1 queries
# Review connection pool settings
```

#### Step 2: Add Database Indexes (30 min)
```sql
-- Add indexes for common query fields
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Verify index usage
EXPLAIN ANALYZE SELECT * FROM users WHERE username = 'test';
```

#### Step 3: Optimize GORM Queries (2h)
```go
// Before: N+1 query risk
users, _ := db.Find(&users)
for _, user := range users {
    db.Model(&user).Association("Posts").Find(&posts)
}

// After: Eager loading
db.Preload("Posts").Find(&users)

// Select specific columns
db.Select("id, username, email").Find(&users)

// Use transactions for multi-step
db.Transaction(func(tx *gorm.DB) error {
    // ... operations
    return nil
})
```

#### Step 4: Connection Pool Tuning (30 min)
```go
// Current settings in internal/storage/user/new.go
db.DB().SetMaxIdleConns(10)
db.DB().SetMaxOpenConns(100)

// Tune based on workload
// Monitor connection usage
// Adjust pool sizes
```

#### Step 5: Add Query Performance Tests (1h)
```go
func BenchmarkUserQueries(b *testing.B) {
    repo := setupTestRepo()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        repo.FindUserByUsername(context.Background(), "testuser")
    }
}
```

#### Step 6: Document Patterns (30 min)
- Create query patterns guide
- Document GORM best practices
- Update architecture documentation
- Note migration path to pgx/sqlc if needed later

#### Step 7: Performance Baseline (1h)
```bash
# Run benchmarks
mise exec -- go test -bench=. -benchmem ./...

# Document baseline metrics
# Track query performance
# Set up monitoring
```

### Success Criteria
- ✅ Database indexes added for common queries
- ✅ N+1 queries eliminated
- ✅ Connection pool optimized
- ✅ Performance baseline established
- ✅ Query patterns documented
- ✅ No breaking changes to existing queries

### Risk Assessment
- **Risk Level:** Low
- **Mitigation:** All changes additive, no schema migrations
- **Rollback:** Simple git revert if issues

### Related Code Files
- `internal/storage/user/user.go` (optimize queries)
- `internal/storage/user/new.go` (connection pool)
- Database schema (add indexes)

### Next Steps
After database optimization complete → Proceed to Phase 04 (Echo v5 - DEFERRED)

---

## Phase 04: Echo v4 → v5 Migration (6 hours)

**Status:** `completed` ✅ — govern and sample migrated to Echo v5.2.1; full validation green.  
**Priority:** **P3** (Originally deferred; unblocked early)  
**Risk Level:** Medium-High  
**Dependencies:** Govern Echo v5 support  

### Overview
**RECOMMENDATION: DEFER migration** until govern package officially supports Echo v5. Echo v4 remains stable and production-ready. Migration has 15+ breaking changes requiring systematic updates across handlers, middleware, error handling.

### Why Defer?
1. **govern/http/echo compatibility unknown** - Must verify govern package supports Echo v5
2. **Echo v4 stable** - No security issues, production-ready
3. **Medium-High risk** - Widespread handler signature changes
4. **Ecosystem maturity** - Wait for wider v5 adoption

### Breaking Changes Summary

| Category | Changes | Risk |
|----------|---------|------|
| Handler Signatures | `echo.Context` → `*echo.Context` | HIGH |
| Error Handler | Parameter swap `(err, c)` → `(c, err)` | HIGH |
| Logger | Custom interface → `*slog.Logger` | MED |
| HTTPError | `Message interface{}` → `Message string` | MED |
| Response | Return type `*Response` → `http.ResponseWriter` | MED |

### When to Migrate
- ✅ Govern package confirms Echo v5 support
- ✅ Echo v5 ecosystem stable (3-6 months post-release)
- ✅ Team capacity for migration work available
- ✅ No pressing feature deadlines

### Migration Strategy (When Ready)

#### Step 1: Verify Dependencies (30 min)
```bash
# Check govern package
grep -r "govern/http/echo" .

# Test Echo v5 compatibility
go get github.com/labstack/echo/v5@latest
go mod tidy
```

#### Step 2: Global Replacements (30 min)
```bash
# Update import paths
find . -type f -name "*.go" -exec sed -i 's/github\.com\/labstack\/echo\/v4/github.com\/labstack\/echo\/v5/g' {} +

# Update handler signatures
find . -type f -name "*.go" -exec sed -i 's/echo\.Context/*echo.Context/g' {} +
```

#### Step 3: Manual Fixes (3h)
- Fix `customHTTPErrorHandler` parameter swap (`handler.go:76`)
- Fix `HTTPError.Message` handling (remove `fmt.Sprintf`)
- Fix `Response()` field access patterns
- Test all middleware (CORS, Security, RateLimit)
- Test all handlers (auth, health)

#### Step 4: Validation (1h)
```bash
mise exec -- go build ./...
mise exec -- golangci-lint run
mise exec -- go test ./...
```

#### Step 5: Documentation (30 min)
- Update CLAUDE.md Echo patterns
- Document v5-specific changes
- Update middleware examples

### Success Criteria (When Migrating)
- ✅ All tests pass with Echo v5
- ✅ No breaking changes in production
- ✅ Govern package compatibility verified
- ✅ Documentation updated
- ✅ Performance validated

### Risk Assessment
- **Risk Level:** Medium-High
- **Mitigation:** Defer until govern support confirmed
- **Rollback:** Simple `go get echo/v4` if issues

### Related Code Files
- `internal/handler/rest/handler.go` (error handler)
- `internal/handler/rest/middlewares/*.go` (middleware)
- `internal/handler/rest/controllers/auth/auth.go` (handlers)
- `go.mod` (Echo version)

### Monitoring (Until Migration)
- Track Echo v5 adoption
- Monitor govern package updates
- Watch Echo v5 stability reports
- Review v5 migration experiences

### Next Steps
**DEFERRED** - Re-evaluate in Q3-Q4 2026 or when govern package confirms Echo v5 support

---

## Timeline & Milestones

### Immediate (Week 1)
- [x] Phase 01: Go 1.26 upgrade (4h) — **DONE** (1.26.4, 2026-06-28)
- [x] Phase 02: Wire removal (4h) — **DONE** (commit `1672d69`)
- [x] Phase 03: Database optimization — **DONE** (revised 3-task scope, 2026-06-28)

### Short-Term (Q2 2026)
- [x] Complete Go 1.26 upgrade
- [x] Complete Wire removal
- [x] Complete database optimization
- [x] Validate all improvements

### Medium-Term (Q3-Q4 2026)
- [ ] Monitor manual DI stability
- [ ] Measure build time improvement (Wire removal)
- [x] Re-evaluate Echo v5 migration timing
- [x] Check govern package Echo v5 support
- [x] Plan and complete Echo v5 migration

### Long-Term (2027+)
- [ ] Plan Go 1.27 upgrade (deprecated GODEBUG settings)
- [ ] Evaluate database layer migration if performance critical
- [ ] Consider sqlc/pgx if project scales significantly

---

## Rollback Strategy

### Go 1.26 Rollback
```bash
mise use go@1.25.5
mise exec -- go mod tidy
```

### Wire Removal Rollback
```bash
# Restore Wire files from git
git checkout internal/handler/rest/wire.go
git checkout internal/handler/rest/wire_gen.go

# Re-add Wire to go.mod
mise exec -- go get github.com/google/wire@latest
mise exec -- go mod tidy
```

### Database Optimization Rollback
- All changes additive - revert individual optimizations
- Database indexes can be dropped: `DROP INDEX IF EXISTS`
- Query changes revert via git

### Echo v5 Rollback (If Migrated)
```bash
go get github.com/labstack/echo/v4@latest
go mod tidy
# Revert global replacements
```

---

## Unresolved Questions

1. **Go 1.26 GC Impact:** Need production-like benchmark before claiming app-specific GC gains.

2. **Database Performance:** Need explicit latency targets before considering pgx/sqlc or more query work.

3. **Manual DI Impact:** Optional build-time/runtime comparison remains undocumented; no functional blocker.

---

## Dependencies & Blocking

### External Dependencies
- **Go 1.27 release:** GODEBUG migration planning remains future work.
- ~~Govern package: Echo v5 migration blocked until govern compatibility confirmed~~ → resolved 2026-06-28.

### Internal Dependencies
- Phase 02 (Wire removal) depends on Phase 01 (Go 1.26) — done.
- Phase 03 (Database) depends on Phase 02 (Wire removal) — done.
- Phase 04 (Echo v5) unblocked by govern Echo v5 migration (plan 260628-0034) — done. Both modules on `echo/v5 v5.2.1`, full `go test -race ./...` green.

### Team Coordination
- No coordination required for Go 1.26, Wire removal, and database optimization
- Echo v5 migration requires team availability and testing capacity

---

## Resources & References

### Research Reports
- [Go 1.26 Release Notes Analysis](../reports/researcher-260628-0022-go-1.26-release-notes.md)
- [Echo v5 Migration Research](../reports/researcher-260628-0022-echo-v5-migration-research.md)
- [Database Layer Comparison](../reports/researcher-260628-0022-database-layer-comparison.md)

### Official Documentation
- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)
- [Echo API Changes V5](https://github.com/labstack/echo/blob/master/API_CHANGES_V5.md)
- [GORM Documentation](https://gorm.io/docs/)
- [pgx PostgreSQL Driver](https://github.com/jackc/pgx)
- [sqlc Documentation](https://sqlc.dev/)

### Project Documentation
- [README.md](../../README.md) - Current tech stack
- [CLAUDE.md](../../CLAUDE.md) - Development rules
- [System Architecture](../../docs/system-architecture.md) - Architecture patterns

---

**Plan Status:** ✅ Completed (synced 2026-06-28)  
**Next Review:** Before Go 1.27 upgrade or if latency targets require DB migration review  
**Owner:** Development Team  
**Created:** 2026-06-28
