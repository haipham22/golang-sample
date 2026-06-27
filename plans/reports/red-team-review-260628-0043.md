# Red Team Review: Govern Monorepo + Wire Removal

**Review Date**: 2026-06-28
**Reviewer**: code-reviewer (Staff Engineer)
**Plan**: Govern Monorepo Restructure & Wire Removal (260627-2307)
**Overall Assessment**: **CONDITIONAL PASS** - Critical issues require resolution

---

## Executive Summary

This migration plan attempts a complex two-phase transformation: (1) repository restructuring into monorepo with external imports, followed by (2) Wire removal and custom error implementation. While the plan shows attention to detail in several areas, **critical architectural flaws, incorrect assumptions about codebase structure, and dangerous rollback weaknesses** make this plan risky to execute as written.

**Status**: **CONDITIONAL PASS** - Must address all P0 issues before implementation

**Risk Score**: **7/10** (High)

---

## Critical Issues (P0)

### P0-1: Clean Architecture Implementation Mismatch

**Severity**: CRITICAL
**Blocker**: YES

**Issue**: The plan's clean architecture structure doesn't match the existing codebase structure.

**Evidence**:
- Current structure: `internal/{handler,service,storage,model,orm,schemas,validator}/`
- Planned structure: `internal/{usecase,domain,repository,handler,bootstrap}/`
- Plan claims "Direct Migration" but actually requires complete code refactoring

**Problem**: Phase 03 lists files to move but doesn't account for:
1. `internal/storage/` → needs to become `internal/repository/` + `internal/domain/` split
2. `internal/service/` → needs to become `internal/usecase/` + internal restructuring
3. `internal/model/` → needs to become `internal/domain/` + entity extraction
4. `internal/schemas/` → needs to merge into `internal/usecase/*/dto.go`
5. `internal/validator/` → needs to merge into DTOs
6. `internal/orm/` → needs to become `internal/repository/postgres/`

**Impact**: 2h estimate for Phase 03 is off by factor of 10-20x. This is not a "move files" operation, it's a complete architectural refactoring.

**Required Fix**:
```markdown
Create Phase 03-Architecture-Refactoring (16-20h):
- Extract business entities from internal/model/ to internal/domain/
- Move repository interfaces from internal/storage/ to internal/usecase/*/
- Restructure internal/service/ to internal/usecase/*/ (service.go + impl.go + dto.go)
- Move GORM implementations to internal/repository/postgres/
- Merge internal/schemas/ and internal/validator/ into DTOs
- Update all import paths across codebase
```

---

### P0-2: Wire Removal Assumes Wrong Dependency Graph

**Severity**: CRITICAL
**Blocker**: YES

**Issue**: Phase 11's manual DI implementation assumes a dependency graph that doesn't match the actual Wire code.

**Evidence**:
```go
// ACTUAL wire.go (current codebase)
func New(
    log *zap.SugaredLogger,
    port int64,
    appConfig *config.EnvConfigMap,
) (governhttp.Server, func(), error)
```

But Phase 11 shows:
```go
// PLANNED di.go (Phase 11)
func NewApp(cfg *config.Config) (*App, func(), error)
```

**Problems**:
1. Different function signature (current: 3 params, planned: 1 param)
2. Different return type (current: `governhttp.Server`, planned: `*App`)
3. Different config type (current: `*config.EnvConfigMap`, planned: `*config.Config`)
4. Missing `zap.SugaredLogger` dependency
5. cmd/serverd.go doesn't match planned usage

**Impact**: Phase 11's code examples won't compile. The plan is implementing a different system than what exists.

**Required Fix**: Rewrite Phase 11 to match actual Wire implementation:
```go
// MUST match existing wire.go signature
func New(
    log *zap.SugaredLogger,
    port int64,
    appConfig *config.EnvConfigMap,
) (governhttp.Server, func(), error) {
    // Exact same initialization order as wire.go
}
```

---

### P0-3: Git Fast-Export/Import Will Create Repository Corruption

**Severity**: CRITICAL
**Blocker**: YES

**Issue**: Phase 02's git fast-export/import strategy will create corrupted git history.

**Evidence**:
```bash
# Phase 02 Step 3
cd ../govern
git fast-export --all --signed-tags=strip > /tmp/govern-export.fi

cd ../golang-sample
git fast-import --quiet < /tmp/govern-export.fi
```

**Problems**:
1. **Duplicate commits**: fast-import creates new commits with new SHA hashes
2. **Lost commit ancestry**: No merge commit linking govern history to golang-sample history
3. **Broken git blame**: Commits from before "merge" won't show file history correctly
4. **Merge conflicts**: If both repos modified same files (even go.mod), import fails
5. **Branch chaos**: All govern branches imported as-is, creating branch namespace pollution

**What Actually Happens**:
```bash
# After fast-import
git log --oneline --all
# Shows two disconnected commit graphs with NO relationship
# "git merge" won't work because there's no common ancestor
```

**Impact**: Git history becomes corrupted. No way to trace file changes before "merge". Cannot revert individual files.

**Correct Approach**:
```bash
# Use git subtree merge (proven pattern)
git subtree add --prefix=. ../govern main --squash
# OR use git filter-branch with proper merge base
# OR manual copy with proper git commit (loses history but maintains integrity)
```

**Required Fix**: Replace git fast-export/import with proven pattern:
1. **Option A**: `git subtree merge` (preserves history, proper merge)
2. **Option B**: Manual file copy with commit (loses govern history, maintains integrity)
3. **Option C**: `git merge --allow-unrelated-histories` (requires proper setup)

---

### P0-4: External Import Strategy Breaks Local Development

**Severity**: CRITICAL
**Blocker**: YES

**Issue**: Phase 03's external import with replace directive won't work for local development.

**Evidence**:
```go
// Phase 03 planned examples/golang-sample/go.mod
module github.com/haipham22/golang-sample

require github.com/haipham22/govern v1.0.0

replace github.com/haipham22/govern => ../../  // Local development only
```

**Problems**:
1. **Replace directive assumes parent directory structure**: `../../` from `examples/golang-sample/` = repository root
2. **Root module name mismatch**: Repository root will be `module github.com/haipham22/govern` after Phase 02, but replace directive assumes this works
3. **Package import confusion**: Code imports `github.com/haipham22/govern/http` but local module is `github.com/haipham22/golang-sample`
4. **IDE support breaks**: Go modules won't resolve local govern packages correctly
5. **CI/CD complexity**: Need to test both local (replace) and published (no replace) modes

**What Actually Happens**:
```bash
cd examples/golang-sample
go mod tidy
# ERROR: cannot find module providing package github.com/haipham22/govern/http
# Replace directive doesn't work as planned because module paths don't align
```

**Impact**: Local development impossible. Cannot test changes to govern packages without publishing.

**Required Fix**: Either:
1. **Use Go workspace** (recommended for monorepo):
   ```go
   // go.work (repository root)
   go 1.25.0
   
   use (
       ./             # govern module
       ./examples/golang-sample  # sample app module
   )
   ```

2. **Or accept local development requires published version**:
   - Remove replace directive
   - Always use published govern package
   - Test govern changes in separate PR

**Plan Claim vs Reality**:
- Plan: "NO go.work workspace file created"
- Reality: go.work is the standard way to solve this exact problem

---

## High Risks (P1)

### P1-1: Rollback Strategy Has Fatal Gaps

**Severity**: HIGH
**Blocker**: NO (but dangerous)

**Issue**: Rollback strategies in both Part 1 and Part 2 are incomplete.

**Part 1 Rollback Issues**:
```bash
# Phase 02 rollback
git reset --hard pre-govern-merge-backup
```

**Problems**:
1. **No backup of current state**: Only creates tag, doesn't backup working directory
2. **No database rollback consideration**: What if schema changes during migration?
3. **No CI/CD rollback**: GitHub Actions workflows updated but no rollback procedure
4. **No dependency rollback**: go.mod/go.sum changes not easily reverted

**Part 2 Rollback Issues**:
```bash
# Phase 11 rollback
git revert to commit before Phase 09
```

**Problems**:
1. **No testing rollback baseline**: Don't verify Wire version works before migration
2. **No error migration rollback**: Can't easily revert custom error changes
3. **No dependency restoration**: Wire removed from go.mod, no auto-restore

**Required Fix**:
```markdown
Add Pre-Migration Backup Phase:
1. Create backup branch: git branch -c main backup-before-migration
2. Capture test results: go test ./... > baseline-test-results.txt
3. Capture build artifacts: go build -o baseline-binary
4. Verify backup branch compiles and passes tests

Rollback Verification:
1. Test rollback actually works (drill)
2. Document exact commands for each rollback scenario
3. Create rollback runbook with decision tree
4. Add rollback time estimates
```

---

### P1-2: Phase Dependencies Create Unclear Blocking

**Severity**: HIGH
**Blocker**: NO

**Issue**: Phase dependencies create situations where it's unclear what blocks what.

**Evidence**:
- Phase 02 depends on Phase 01 (clear)
- Phase 03 depends on Phase 02 (clear)
- Phase 08 depends on Phase 07 (unclear - Phase 07 is "Repository Rename & Merge", 0h duration)
- Phase 09 depends on Phase 08 (unclear - what does "Setup Wire Removal Environment" mean?)

**Problems**:
1. **Phase 07 has 0h duration** but blocks all of Part 2 - this is actually a "merge to main" operation, not instant
2. **Phase 08 "Setup Wire Removal Environment"** - what does this mean? Create branch? Run tests? Something else?
3. **No validation checkpoint** between Part 1 and Part 2 - what if Part 1 doesn't fully work?

**Required Fix**:
```markdown
Clarify Phase 07:
- Rename this to "Part 1 Completion & Merge to Main"
- Estimate: 2h (not 0h) - includes final testing, PR review, merge
- Add success criteria: Must pass all tests, CI/CD, manual verification

Clarify Phase 08:
- Rename to "Part 2 Preparation & Validation"
- Explicit steps: Create branch, verify Wire version works, capture baseline
- Estimate: 2h (not 2h for unknown "setup")

Add Validation Gate:
- After Phase 07, must run: "Part 1 Validation Checklist"
- Cannot start Part 2 until Part 1 fully validated
```

---

### P1-3: Error Envelope Pattern Doesn't Match Existing Code

**Severity**: HIGH
**Blocker**: NO

**Issue**: Phase 09's error envelope pattern introduces complexity not present in current codebase.

**Evidence**:
```go
// Phase 09 planned envelope pattern
type DBError struct {
    Op       string // Operation
    Table    string // Table
    Err      error  // Underlying error
    Severity string // "transient" or "permanent"
}
```

**Current codebase errors**:
```go
// Current: simple errors
return errors.New("user not found")
// Or govern/errors
return governerrors.WrapCode(governerrors.CodeNotFound, err)
```

**Problems**:
1. **No existing error context**: Current code doesn't track "operation" or "table"
2. **Severity classification**: No existing transient/permanent error handling
3. **Migration complexity**: 44 govern/errors usages need complete rewrite, not simple replacement
4. **Testing gap**: No tests verify error envelope works as expected

**Impact**: Phase 09 estimate of 3h is unrealistic. More like 8-12h to implement envelope pattern + migrate all usages + tests.

**Required Fix**:
```markdown
Simplify error approach:
1. Drop error envelope complexity (severity, operation tracking)
2. Simple AppError type matching govern/errors API
3. Drop-in replacement for govern/errors
4. Migrate usages mechanically (find-replace)

Phase 09 Revised Estimate: 3h → 6h (simple errors, full migration)
```

---

## Medium Concerns (P2)

### P2-1: Time Estimates Are Wildly Optimistic

**Issue**: Most phase estimates don't account for debugging, testing, and unexpected issues.

**Evidence**:
- Phase 02 (git merge): 3h - should be 6-8h (git operations always take longer)
- Phase 03 (move sample app): 2h - should be 20h (complete architecture refactor)
- Phase 11 (manual DI): 4h - should be 8h (need to match exact Wire behavior)

**Recommendation**: Add 50-100% buffer to all estimates. Migration tasks always reveal unexpected issues.

---

### P2-2: Missing Phase for Clean Architecture Migration

**Issue**: Plan assumes existing code already follows clean architecture, but it doesn't.

**Current structure**:
```
internal/
├── handler/    # HTTP handlers
├── service/    # Business logic
├── storage/    # Repository implementations
├── model/      # GORM models
├── orm/        # GORM models
├── schemas/    # DTOs
└── validator/  # Validation
```

**Planned structure**:
```
internal/
├── domain/     # Business entities (flat)
├── usecase/    # Use cases + repository interfaces
├── repository/ # Repository implementations
├── handler/    # HTTP handlers
└── bootstrap/  # Manual DI
```

**Gap**: No phase explains how to transform from current to planned structure.

**Recommendation**: Add explicit "Phase 03-Architecture-Migration" (16-20h) to handle transformation.

---

### P2-3: Module Path Confusion Risk

**Issue**: Plan creates confusing module path situation.

**After Phase 02**: Root module = `github.com/haipham22/govern`
**After Phase 03**: Sample app module = `github.com/haipham22/golang-sample`

**Problem**: Sample app in `examples/golang-sample/` has module path `github.com/haipham22/golang-sample`, but it's actually `github.com/haipham22/govern/examples/golang-sample/`.

**Impact**: When published, users import `github.com/haipham22/golang-sample` expecting standalone app, but it's actually part of govern monorepo.

**Recommendation**: Either:
1. Make sample app module path `github.com/haipham22/govern/examples/golang-sample`
2. Or keep sample app as separate repository (break monorepo pattern)

---

### P2-4: Testing Strategy Insufficient

**Issue**: Testing phases focus on compilation, not behavioral verification.

**Evidence**:
- Phase 06: "Validation & Testing" - 1h - only checks compilation
- Phase 13: "Wire Removal Testing" - 4h - but what to test?

**Missing**:
1. No baseline performance metrics (startup time, memory usage)
2. No API contract testing (verify HTTP responses don't change)
3. No error handling verification (errors still work correctly)
4. No integration testing (full system works end-to-end)
5. No load testing (performance doesn't degrade)

**Recommendation**: Add comprehensive testing strategy phase before Part 2 starts.

---

## Technical Findings

### Architecture Issues

1. **Clean Architecture Layer Violations**:
   - Plan claims clean architecture but current code has violations
   - `internal/model/` mixes entities with GORM concerns
   - `internal/service/` mixes use cases with business logic
   - No clear separation between domain entities and DTOs

2. **Dependency Direction Confusion**:
   - Plan shows `handler → usecase → domain`
   - But current code has `handler → controller → service → storage → model`
   - No mapping between current and planned structure

3. **Interface Placement**:
   - Plan claims interfaces defined by consuming layer (bxcodec pattern)
   - But current code has interfaces in wrong places
   - No phase moves interfaces to correct locations

---

### Implementation Issues

1. **Wire Removal Complexity Underestimated**:
   - Current wire.go has 8 provider functions
   - Plan shows simple `New()` function but doesn't handle all providers
   - No mention of how to handle `provideDebugFlag`, `provideEnv`, etc.

2. **Error Migration Complexity Underestimated**:
   - 44 govern/errors usages across codebase
   - Plan shows simple replacement but doesn't account for:
     - Different import paths
     - Context preservation
     - Error wrapping chain
     - Test compatibility

3. **Git Operations Oversimplified**:
   - Git fast-export/import shown as simple commands
   - No handling for:
     - Branch conflicts
     - Merge conflicts
     - Submodule issues
     - Large file handling

---

### Code Examples Issues

1. **Phase 11 Code Won't Compile**:
   - Shows `bootstrap.NewApp()` but doesn't match existing `cmd/serverd.go`
   - Uses `*config.Config` but existing code uses `*config.EnvConfigMap`
   - Returns `*App` but existing code expects `governhttp.Server`

2. **Phase 09 Code Incomplete**:
   - Shows error envelope pattern but doesn't show how to use it
   - No examples of migrating existing govern/errors calls
   - No test examples

3. **Phase 03 Code Examples Missing**:
   - Shows directory structure but no code examples
   - No examples of how to restructure existing code
   - No import path update examples

---

## Recommendations

### Before Implementation (Mandatory)

1. **Fix P0-1 (Architecture Mismatch)**:
   - Add explicit architecture migration phase
   - Document transformation from current to planned structure
   - Update time estimates (20h instead of 2h)

2. **Fix P0-2 (Wire Dependency Graph)**:
   - Rewrite Phase 11 to match actual wire.go
   - Update code examples to compile
   - Add verification tests

3. **Fix P0-3 (Git Corruption)**:
   - Replace git fast-export/import with git subtree merge
   - Add git history verification steps
   - Test merge on backup repositories first

4. **Fix P0-4 (External Import)**:
   - Accept go.work workspace or redesign external import strategy
   - Document local development workflow
   - Test local development workflow end-to-end

### During Implementation (Recommended)

1. **Add Validation Checkpoints**:
   - After Part 1: Complete validation gate
   - After Phase 11: Verify manual DI works
   - After Phase 13: Full system tests

2. **Improve Rollback Strategy**:
   - Create backup branches before each phase
   - Test rollback procedures
   - Document rollback runbooks

3. **Add Buffer Time**:
   - Increase all estimates by 50%
   - Add explicit debugging time
   - Add contingency phases

### After Implementation (Recommended)

1. **Run Comprehensive Tests**:
   - Performance benchmarks
   - API contract tests
   - Error handling tests
   - Integration tests

2. **Update Documentation**:
   - Document new architecture
   - Update development guides
   - Create troubleshooting guides

---

## Overall Risk Score: 7/10

**Breakdown**:
- **Architecture Risk**: 8/10 (Clean architecture mismatch critical)
- **Implementation Risk**: 7/10 (Wire removal underestimated)
- **Git Operations Risk**: 9/10 (Fast-export/import dangerous)
- **Rollback Risk**: 6/10 (Rollback strategies incomplete)
- **Testing Risk**: 7/10 (Insufficient testing strategy)

**Risk Level**: **HIGH**

---

## Conclusion

This plan shows ambition and attention to detail in some areas, but **critical architectural flaws, incorrect assumptions about the existing codebase, and dangerous technical decisions** make it unsafe to implement as written.

**Must Fix Before Implementation**:
1. P0-1: Add architecture migration phase (20h)
2. P0-2: Fix Phase 11 to match actual Wire code
3. P0-3: Replace git fast-export/import with proven pattern
4. P0-4: Fix external import strategy

**Recommended Approach**:
1. **Part 1 first**: Complete monorepo restructuring, validate thoroughly
2. **Architecture cleanup**: Separate phase for clean architecture migration
3. **Part 2 after Part 1 validated**: Don't start Part 2 until Part 1 proven stable
4. **Incremental validation**: Add validation checkpoints between phases

**Alternative Consideration**: Split into two separate plans:
1. Plan A: Monorepo restructuring (Part 1 only)
2. Plan B: Wire removal + clean architecture (Part 2 only)

Execute sequentially, not as single plan.

---

## Unresolved Questions

1. Why does Phase 07 have 0h duration? What actually happens in this phase?
2. What does "Setup Wire Removal Environment" (Phase 08) actually mean?
3. Has git fast-export/import been tested on backup repositories?
4. Why reject go.work workspace when it solves the exact problem?
5. Where is the mapping between current structure and planned structure?
6. How will existing `internal/{service,storage,model}` code transform to `internal/{usecase,domain,repository}`?
7. What happens if wire.go has more providers than documented?
8. How to verify git history preserved correctly after fast-import?
9. What if govern repository has merge conflicts during fast-export?
10. How to handle GitHub repository rename side effects?

---

**Status**: **CONDITIONAL PASS** - Address all P0 issues before implementation
**Next Actions**: Fix P0 issues, then re-review
**Review Complete**: 2026-06-28
