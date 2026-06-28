---
title: "Phase 01: Go 1.26 Upgrade Implementation"
description: "Upgrade from Go 1.25.0 to Go 1.26 with performance improvements and enhanced security"
status: completed
priority: P1
effort: 4h
branch: main
tags: [go-1.26, upgrade, performance, security]
created: 2026-06-28
---

# Phase 01: Go 1.26 Upgrade Implementation

**Status:** `completed` ✅ — done 2026-06-28. Bumped to **Go 1.26.4** across the monorepo (mise.toml + both `go.mod` + CI + Dockerfile). All tests pass with race detector; build + vet clean. `go fix` applied the `interface{}→any` modernizer. Bonus: `go mod tidy` removed govern's unused `google/wire` dep.  
**Priority:** **P1** (Do First)  
**Risk Level:** Very Low  
**Dependencies:** None  
**Estimated Time:** 4 hours

---

## Overview

Upgrade from Go 1.25.0 to Go 1.26 for:
- **10-40% reduction in garbage collection overhead** (Green Tea GC)
- **Enhanced cryptographic security** (post-quantum TLS by default)
- **Improved developer tooling** (revamped `go fix` command)
- **Zero breaking changes** (Go 1 compatibility promise)

---

## Key Changes Summary

### Performance Improvements
| Area | Improvement | Impact Level |
|------|-------------|--------------|
| Garbage Collection | 10-40% reduction in overhead | **High** (heap-intensive apps) |
| CGO Calls | ~30% faster baseline | Medium (FFI-heavy apps) |
| Stack Allocation | More slices allocated on stack | Low-Medium |
| io.ReadAll | 2x faster, 50% less memory | Medium (large reads) |
| fmt.Errorf | Reduced allocations | Low |

### Security Enhancements
- **Post-Quantum TLS:** `SecP256r1MLKEM768` and `SecP384r1MLKEM1024` key exchanges now default
- **Safer Cryptographic Randomness:** All `crypto/*` packages ignore custom `rand` parameters
- **TLS Security Improvements:** Several legacy TLS options deprecated

### Breaking Changes
**OFFICIAL GOAL:** No breaking changes for almost all Go programs
- JPEG encoder/decoder replacement (bit-for-bit output differences - rare)
- URL parsing strictness (malformed URLs rejected - correct behavior)
- Reflect iterator methods (additive only)

---

## Implementation Steps

### Step 1: Update mise.toml (5 min)

**File:** `mise.toml`

**Before:**
```toml
[tools]
go = "1.25.5"
```

**After:**
```toml
[tools]
go = "1.26.0"
```

**Commands:**
```bash
mise install
mise exec -- go version
```

**Acceptance Criteria:**
- ✅ `go version` outputs `go version go1.26.0`
- ✅ mise reports Go 1.26.0 installed

---

### Step 2: Update go.mod (5 min)

**File:** `go.mod`

**Before:**
```go
go 1.25.0
```

**After:**
```go
go 1.26.0
```

**Commands:**
```bash
mise use go@1.26
mise exec -- go mod tidy
```

**Acceptance Criteria:**
- ✅ `go.mod` updated to `go 1.26.0`
- ✅ `go.sum` updated with new checksums
- ✅ All dependencies compatible with Go 1.26

---

### Step 3: Run Code Modernizers (15 min)

**Commands:**
```bash
mise exec -- go fix ./...
```

**What `go fix` does:**
- Applies modernizers for latest Go idioms
- Updates deprecated syntax
- Source-level inlining via `//go:fix inline` directives
- Dozens of automated fixers

**Acceptance Criteria:**
- ✅ `go fix` completes without errors
- ✅ Review changes with `git diff`
- ✅ Code uses modern Go patterns

---

### Step 4: Verify Dependencies (10 min)

**Commands:**
```bash
mise exec -- go mod verify
mise exec -- go list -m all
```

**What to check:**
- All dependencies verified with checksums
- No incompatible direct dependencies
- Indirect dependencies compatible

**Acceptance Criteria:**
- ✅ `go mod verify` passes (all modules verified)
- ✅ No Go version conflicts in dependency tree
- ✅ Key dependencies verified compatible:
  - `github.com/labstack/echo/v4` ✅
  - `gorm.io/gorm` ✅
  - `github.com/haipham22/govern` ✅

---

### Step 5: Run Test Suite (30 min)

**Commands:**
```bash
# Standard tests
mise exec -- go test ./...

# Race detection
mise exec -- go test -race ./...

# Coverage report
mise exec -- go test -cover ./...

# Verbose output for debugging
mise exec -- go test -v ./...
```

**What to test:**
- HTTP handler tests (GC improvements benefit these)
- Cookie handling tests (scoping change in net/http)
- Cryptographic operations (new randomness behavior)
- Storage layer tests (database operations)

**Acceptance Criteria:**
- ✅ All tests pass (`go test ./...`)
- ✅ No race conditions detected (`-race`)
- ✅ Coverage maintained at 83.3% or higher
- ✅ No new test failures introduced

---

### Step 6: Performance Validation (1h)

**Commands:**
```bash
# Benchmark HTTP handlers
GODEBUG=gctrace=1 mise exec -- go test -bench=. -benchmem ./internal/handler/rest/...

# Compare GC pause times
# Document performance improvements
# Profile before/after if critical
```

**What to measure:**
- HTTP request handling performance
- Memory allocation patterns
- GC pause times (should see improvement)
- JSON marshaling performance

**Acceptance Criteria:**
- ✅ Benchmarks complete successfully
- ✅ GC pause times documented
- ✅ Performance improvement measurable (if available)
- ✅ No performance regressions

---

### Step 7: Update CI/CD (30 min)

**Files to update:**
- `.github/workflows/test.yml`
- `.github/workflows/push.yml`
- Any other GitHub Actions workflows

**Example workflow update:**
```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'  # Updated from 1.25
```

**Acceptance Criteria:**
- ✅ All workflows updated to Go 1.26
- ✅ CI pipeline passes with Go 1.26
- ✅ No workflow failures

---

### Step 8: GODEBUG Migration Planning (30 min)

**Commands:**
```bash
# Check for GODEBUG usage
grep -r "GODEBUG" .
```

**Deprecated GODEBUG settings (removed in Go 1.27):**
- `tlsunsafeekm` - ExportKeyingMaterial requires TLS 1.3+
- `tlsrsakex` - RSA-only key exchanges disabled
- `tls10server` - Minimum TLS version becomes 1.2
- `tls3des` - 3DES cipher suites removed
- `x509keypairleaf` - Certificate.Leaf always populated

**If GODEBUG found:**
1. Document current usage
2. Plan migration before Go 1.27
3. Update to use modern TLS settings
4. Test without GODEBUG overrides

**Acceptance Criteria:**
- ✅ GODEBUG usage documented (if any)
- ✅ Migration plan created for Go 1.27
- ✅ No critical legacy TLS dependencies

---

## Success Criteria

### Required
- ✅ All tests pass with Go 1.26
- ✅ No breaking changes detected
- ✅ CI/CD pipeline updated and passing
- ✅ Development documentation updated

### Desired
- ✅ Performance improvement measurable (GC overhead reduction)
- ✅ Benchmarks documented
- ✅ Zero test failures

---

## Risk Assessment

### Risk Level: Very Low

**Rationale:**
- Go 1 promise of backward compatibility
- Zero breaking changes for typical web applications
- Well-tested release (February 2026)
- Long support timeline

**Potential Issues:**
1. **JPEG encoding differences** (very rare impact)
   - **Mitigation:** Test affected endpoints
   - **Rollback:** Use Go 1.25 if critical

2. **URL parsing strictness** (correct behavior)
   - **Mitigation:** Review URL validation code
   - **Impact:** Should reject malformed URLs anyway

3. **Dependency compatibility** (unlikely)
   - **Mitigation:** All major dependencies tested
   - **Rollback:** Simple version downgrade

### Rollback Strategy

```bash
# Rollback mise.toml
mise install go@1.25.5

# Rollback go.mod
sed -i 's/go 1.26.0/go 1.25.0/' go.mod

# Revert dependencies
mise exec -- go mod tidy
```

---

## Related Code Files

### Files to Modify
- `mise.toml` (update go version)
- `go.mod` (update go directive)
- `.github/workflows/test.yml` (CI/CD update)
- `.github/workflows/push.yml` (CI/CD update)

### Files to Review
- `internal/handler/rest/handler.go` (HTTP handlers)
- `internal/storage/user/user.go` (database operations)
- All test files (verify pass/fail status)

---

## Data Flow

```
Go 1.25 → [Upgrade] → Go 1.26
   ↓
   mise.toml (version update)
   ↓
   go.mod (directive update)
   ↓
   go fix (code modernization)
   ↓
   Tests (validation)
   ↓
   CI/CD (pipeline update)
   ↓
   Production (deployment)
```

---

## Dependency Graph

```
Phase 01 (Go 1.26)
├── mise.toml update
├── go.mod update
├── go fix execution
│   └── depends on: go.mod update
├── Test execution
│   └── depends on: go fix
├── Performance validation
│   └── depends on: Test execution
└── CI/CD update
    └── depends on: Test execution
```

---

## Backwards Compatibility

### Compatibility Strategy
- **No breaking changes** - Go 1 promise maintained
- **Deprecations only** - GODEBUG settings removed in Go 1.27
- **Migration timeline** - 6 months to address deprecations

### Migration Path
1. **Go 1.26** (current upgrade) - All existing code works
2. **Go 1.27** (future) - Must migrate away from deprecated GODEBUG
3. **Go 1.28+** - Continue with modern Go

---

## Testing Strategy

### Unit Tests
- Run full test suite with Go 1.26
- Verify race detector passes
- Check coverage maintained

### Integration Tests
- Test HTTP endpoints (handlers benefit from GC improvements)
- Verify database operations
- Validate cryptographic operations

### Performance Tests
- Benchmark HTTP request handling
- Compare GC pause times
- Document performance improvements

### Regression Tests
- Verify no new test failures
- Check for behavioral changes
- Validate security features

---

## Rollback Plan

### If Upgrade Fails

1. **Immediate rollback:**
   ```bash
   mise install go@1.25.5
   mise exec -- go mod tidy
   ```

2. **Revert code changes:**
   ```bash
   git checkout go.mod
   git checkout go.sum
   git checkout mise.toml
   ```

3. **Investigate failure:**
   - Review error logs
   - Check dependency conflicts
   - Verify test failures

4. **Retry with fixes:**
   - Address dependency issues
   - Fix test failures
   - Retry upgrade

### Rollback Criteria
- Critical production issues
- Test suite failures blocking release
- Dependency incompatibilities

---

## Monitoring & Validation

### Post-Upgrade Monitoring
- GC pause times (should improve)
- Memory usage (may decrease)
- HTTP response times (may improve)
- Error rates (should not increase)

### Validation Steps
1. **Development environment:**
   - All tests pass
   - Benchmarks run successfully
   - No errors in logs

2. **Staging environment:**
   - Smoke tests pass
   - Performance validated
   - No regression detected

3. **Production environment:**
   - Monitor error rates
   - Check performance metrics
   - Validate GC improvements

---

## Next Steps

After Phase 01 complete:
1. Proceed to **Phase 02: Database Layer Optimization**
2. Document performance improvements
3. Update development documentation
4. Plan Go 1.27 migration (deprecated GODEBUG)

---

## References

### Research
- [Go 1.26 Release Notes Research](../reports/researcher-260628-0022-go-1.26-release-notes.md)

### Official Documentation
- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)
- [Go 1 Compatibility Promise](https://go.dev/doc/go1compat)
- [Green Tea GC](https://go.dev/doc/go1.25#gc)

### Project Documentation
- [README.md](../../README.md) - Tech stack reference
- [CLAUDE.md](../../CLAUDE.md) - Development rules

---

**Phase Status:** ✅ Completed (2026-06-28)  
**Completion Target:** Week 1  
**Owner:** Development Team  
**Created:** 2026-06-28
