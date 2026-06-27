# Plan Audit - Duplicate Content Analysis

**Date**: 2026-06-27
**Auditor**: Claude
**Plans Audited**:
- Plan 1: 260627-2136-govern-monorepo-restructure (14 phases)
- Plan 2: 260627-2118-wire-removal-and-centralized-error-management (6 phases)

---

## Executive Summary

**Status**: ⚠️ CRITICAL INCONSISTENCIES FOUND

**Findings**:
1. ❌ **CRITICAL**: Phase 03 (monorepo plan) contains obsolete go.work workspace instructions contradicting external import decision
2. ⚠️ **HIGH**: Phase 09-14 files missing from merged plan (only listed in plan.md)
3. ⚠️ **MEDIUM**: Effort estimation discrepancy (line 6 says 16h, should be 32h)
4. ✅ **LOW**: No actual content duplication between plans

---

## Critical Issues

### Issue #1: Obsolete go.work Instructions in Phase 03

**Location**: `plans/260627-2136-govern-monorepo-restructure/phase-03-move-sample-application.md`

**Problem**: Phase 03 Step 5 (lines 301-332) contains instructions to create `go.work` workspace file, but user explicitly decided on **external import approach only**.

**Evidence from User Feedback**:
- "sample là module riêng không liên quan tới root"
- "Import govern như external"
- Decision: External import approach instead of go.work workspace

**Impact**: HIGH - If implemented, would create unnecessary workspace file contradicting architecture decision

**Current Content (Lines 301-332)**:
```markdown
### Step 5: Create Go Workspace File
**Duration**: 20 minutes

**Critical Step**: Configure Go workspace for local development

```bash
# Create go.work at repository root
cat > go.work << 'EOF'
go 1.25

use (
    .                          // govern module (root)
    ./golang-sample   // sample app module
)
EOF
```

**Required Fix**: Remove entire Step 5 from Phase 03, update subsequent steps to remove go.work references.

---

### Issue #2: Missing Phase 09-14 Files

**Location**: `plans/260627-2136-govern-monorepo-restructure/`

**Problem**: plan.md lists Phase 09-14 but these phase files don't exist in the directory.

**Expected Files** (from Plan 2 content):
- `phase-09-setup-wire-removal-environment.md` (from Plan 2 Phase 01)
- `phase-10-custom-error-types.md` (from Plan 2 Phase 02)
- `phase-11-centralized-error-management.md` (from Plan 2 Phase 03)
- `phase-12-manual-di-implementation.md` (from Plan 2 Phase 04)
- `phase-13-error-handler-refactoring.md` (from Plan 2 Phase 05)
- `phase-14-wire-removal-testing.md` (from Plan 2 Phase 06)

**Current State**:
- Only Phase 01-08 files exist
- Phase 09-14 referenced in plan.md but files not created

**Impact**: HIGH - Wire removal phases cannot be executed without these files

**Required Action**: Copy and adapt Plan 2 phase files to Plan 1 as Phase 09-14

---

### Issue #3: Effort Estimation Discrepancy

**Location**: `plans/260627-2136-govern-monorepo-restructure/plan.md:6`

**Problem**: Line 6 shows `effort: 16h` but line 44 correctly states `32 hours (16h monorepo + 16h wire removal)`

**Impact**: MEDIUM - Misleading effort estimation in plan metadata

**Required Fix**: Update line 6 to `effort: 32h`

---

## Content Overlap Analysis

### No Actual Content Duplication ✅

**Finding**: Plan 1 (monorepo) and Plan 2 (wire removal) have **no overlapping content** - they address different concerns:

**Plan 1 Scope** (Phase 01-08):
- Repository structure changes
- Module path updates
- Git history preservation
- Generator implementation
- Documentation migration

**Plan 2 Scope** (Phase 01-06 → Plan 1 Phase 09-14):
- Wire dependency injection removal
- Custom error types
- Centralized error management
- Manual DI implementation
- Error handler refactoring
- Comprehensive testing

**Separation**: Clean separation of concerns - no duplicate work

---

## Phase Mapping: Plan 2 → Plan 1

| Plan 2 Phase | Plan 1 Phase | Status | File Exists |
|--------------|--------------|--------|-------------|
| Phase 01: Setup & Validation | Phase 09: Setup Wire Removal Environment | ❌ Missing | No |
| Phase 02: Custom Error Types | Phase 10: Custom Error Types | ❌ Missing | No |
| Phase 03: Centralized Error Management | Phase 11: Centralized Error Management | ❌ Missing | No |
| Phase 04: Manual DI Implementation | Phase 12: Manual DI Implementation | ❌ Missing | No |
| Phase 05: Error Handler Refactoring | Phase 13: Error Handler Refactoring | ❌ Missing | No |
| Phase 06: Testing & Validation | Phase 14: Wire Removal Testing | ❌ Missing | No |

---

## Recommendations

### Immediate Actions Required

1. **Fix Phase 03** (HIGH PRIORITY):
   - Remove Step 5 (Create Go Workspace File)
   - Update Step 6 (Split Makefile) to remove go.work references
   - Update Step 10 (Verify Go Workspace) → Change to "Verify External Import"
   - Update commit message to remove workspace references
   - Update README to remove workspace references

2. **Create Phase 09-14 Files** (HIGH PRIORITY):
   - Copy Plan 2 Phase 01-06 content to Plan 1 Phase 09-14
   - Update file names and internal references
   - Update context links to point to merged plan
   - Update dependencies to reference previous phases in merged plan

3. **Update Plan Metadata** (MEDIUM PRIORITY):
   - Update `plan.md:6` from `effort: 16h` to `effort: 32h`
   - Verify all effort totals are consistent

4. **Validation** (MEDIUM PRIORITY):
   - After fixes, verify no go.work references remain in merged plan
   - Verify all 14 phase files exist
   - Verify phase dependencies are correct

---

## Detailed Fix: Phase 03 Obsolete Content

### Steps to Remove from Phase 03:

**Remove Step 5 entirely** (lines 301-332):
- "Create Go Workspace File"
- All go.work creation commands
- go work sync commands

**Update Step 6** (lines 336-417):
- Remove go.work from Makefile example
- Update commit message to remove workspace references

**Update Step 10** (lines 593-623):
- Change title from "Verify Go Workspace" to "Verify External Import"
- Replace workspace commands with external import verification:
  ```bash
  # Verify replace directive
  cat golang-sample/go.mod | grep "replace github.com/haipham22/govern"

  # Test external import resolves correctly
  cd golang-sample
  mise exec -- go mod tidy
  mise exec -- go build ./cmd/serverd.go
  ```

**Update Step 13** (commit message, lines 722-749):
- Remove all "Go workspace" references
- Replace with "external import with replace directive for local development"
- Update commit message title

---

## Detailed Fix: Create Phase 09-14

### Process:

1. For each Plan 2 phase file (01-06):
   - Read content
   - Update phase number (+8)
   - Update internal links/references
   - Update context to reflect merged plan structure
   - Update dependencies to reference previous phases in merged plan
   - Save as new file in merged plan directory

2. Example transformation:
   ```
   phase-01-setup-and-validation.md → phase-09-setup-wire-removal-environment.md
   - Update: dependsOn: [phase-03-move-sample-application.md] (was empty)
   - Update: Parent Plan link to [plan.md](../plan.md)
   - Update: Previous Phase link to [phase-08-repository-rename-merge.md]
   ```

---

## Success Criteria

**Audit Complete When**:
- [ ] Phase 03 has no go.work references
- [ ] Phase 03 references external import approach only
- [ ] All 14 phase files exist in merged plan directory
- [ ] Phase 09-14 files properly adapted from Plan 2
- [ ] plan.md effort shows 32h
- [ ] All internal links/references updated correctly
- [ ] Phase dependencies form valid chain (01→02→...→14)

---

## Next Steps

1. Review this audit report
2. Approve fixes for Phase 03
3. Approve creation of Phase 09-14 files
4. Execute fixes
5. Re-audit to verify all issues resolved

---

**Audit Status**: 🔴 CRITICAL ISSUES FOUND - ACTION REQUIRED
**Blocker**: Cannot proceed with implementation until Phase 03 fixed and Phase 09-14 created
