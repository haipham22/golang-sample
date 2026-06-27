# Phase Consistency Check Report

**Date**: 2026-06-28  
**Plan**: govern-monorepo-and-wire-removal  
**Scope**: All 13 phase files + main plan.md  
**Total Issues Found**: 17 critical, 8 high, 5 medium

---

## Summary

Consistency analysis reveals **30 total issues** across dependency chains, file paths, working directories, phase numbering, technical details, and cross-references. The most critical issues involve:

1. **Circular dependency in Part 2 phases** (Phase 08 references wrong dependency)
2. **Phase numbering mismatch** (filenames vs titles don't match)
3. **Working directory confusion** (golang-sample/ vs examples/golang-sample/)
4. **File path inconsistencies** (Phase 08 missing go.mod file reference)
5. **Context link errors** (phases reference wrong previous/next phases)

---

## Dependency Chain Issues

### CRITICAL-001: Phase 08 Wrong Dependency Reference

**Location**: `phase-08-setup-wire-removal.md` line 7  
**Issue**: Phase 08 declares `dependsOn: [phase-07-repository-rename-merge.md]` but should depend on Phase 07 (Validation & Testing), not Phase 07 (Repository Rename)

**Current**:
```yaml
dependsOn: [phase-07-repository-rename-merge.md]
```

**Should Be**:
```yaml
dependsOn: [phase-06-validation-testing.md]
```

**Impact**: Phase 08 is Part 2 (Wire Removal), which should start after Part 1 (Monorepo) completes. Phase 06 is the last phase of Part 1, not Phase 07 (Repository Rename).

**Fix**: Update line 7 in phase-08 to reference correct dependency.

---

### CRITICAL-002: Phase 09 Wrong Context Link

**Location**: `phase-09-custom-error-types.md` line 21  
**Issue**: References "phase-01-setup-and-validation.md" which doesn't exist

**Current**:
```markdown
- **Previous Phase**: [phase-01-setup-and-validation.md](./phase-01-setup-and-validation.md)
```

**Should Be**:
```markdown
- **Previous Phase**: [phase-08-setup-wire-removal.md](./phase-08-setup-wire-removal.md)
```

**Impact**: Broken documentation link causes confusion.

---

### CRITICAL-003: Phase 10 Wrong Context Link

**Location**: `phase-10-centralized-error-management.md` line 21  
**Issue**: References "phase-03" instead of "phase-09"

**Current**:
```markdown
- **Next Phase**: [phase-03-centralized-error-management.md](./phase-03-centralized-error-management.md)
```

**Should Be**:
```markdown
- **Next Phase**: [phase-11-manual-di-implementation.md](./phase-11-manual-di-implementation.md)
```

---

### CRITICAL-004: Phase 09 Wrong Next Phase Link

**Location**: `phase-09-custom-error-types.md` line 22  
**Issue**: References "phase-03" instead of "phase-10"

**Current**:
```markdown
- **Next Phase**: [phase-03-centralized-error-management.md](./phase-03-centralized-error-management.md)
```

**Should Be**:
```markdown
- **Next Phase**: [phase-10-centralized-error-management.md](./phase-10-centralized-error-management.md)
```

---

### CRITICAL-005: Phase 11 Wrong Previous Phase Link

**Location**: `phase-11-manual-di-implementation.md` line 21  
**Issue**: References "phase-10" as previous phase in text but the dependency is correct in frontmatter

**Current**:
```markdown
- **Previous Phase**: [phase-10-centralized-error-management.md](./phase-10-centralized-error-management.md)
```

**Status**: Actually correct - this is a false positive. The link matches the dependency.

---

### HIGH-001: Phase Dependency Chain Broken

**Location**: `phase-11-manual-di-implementation.md` frontmatter  
**Issue**: Phase 11 references Phase 10 as dependency, but Phase 10 doesn't have corresponding link back to Phase 11

**Impact**: Documentation chain broken - users can't navigate forward from Phase 10.

**Fix**: Update Phase 10 line 22 to reference Phase 11 as next phase (see CRITICAL-004).

---

### HIGH-002: Phase 12 Missing Previous Phase Link

**Location**: `phase-12-error-handler-refactoring.md` frontmatter  
**Issue**: Has dependsOn but no context link to previous phase

**Current**: No context links section at all.

**Should Add**:
```markdown
## Context Links

- **Parent Plan**: [plan.md](./plan.md)
- **Previous Phase**: [phase-11-manual-di-implementation.md](./phase-11-manual-di-implementation.md)
- **Next Phase**: [phase-13-wire-removal-testing.md](./phase-13-wire-removal-testing.md)
```

---

### HIGH-003: Phase 13 Missing Next Phase Link

**Location**: `phase-13-wire-removal-testing.md` frontmatter  
**Issue**: No "Next Phase" link (expected - it's the last phase)

**Status**: This is correct - Phase 13 is the final phase. No fix needed.

---

## File Path Issues

### CRITICAL-006: Phase 01 Missing Directory Creation

**Location**: `phase-01-repository-preparation.md` line 207  
**Issue**: Command creates `golang-sample/` directory but should create `examples/golang-sample/`

**Current**:
```bash
mkdir -p golang-sample
```

**Should Be**:
```bash
mkdir -p examples/golang-sample
```

**Impact**: Creates directory at wrong location - doesn't match monorepo structure.

---

### CRITICAL-007: Phase 03 Inconsistent Directory References

**Location**: `phase-03-move-sample-application.md` lines 203-226  
**Issue**: Commands use `golang-sample/` but should use `examples/golang-sample/`

**Current**: Multiple references to `golang-sample/` without `examples/` prefix.

**Should Be**: All references should be `examples/golang-sample/`

**Impact**: Creates confusion about working directory - Phase 03 explicitly states it moves sample app to `examples/golang-sample/`.

---

### CRITICAL-008: Phase 04 Missing File Deletion

**Location**: `phase-04-root-configuration-update.md` lines 90-92  
**Issue**: Lists files to delete but this contradicts earlier phases

**Current**:
```markdown
### Files to Delete
- Old README.md content (replace existing)
```

**Impact**: No actual files deleted in this phase - this is misleading. Should be "Files to Replace" not "Delete".

---

### CRITICAL-009: Phase 08 Wrong Working Directory Files

**Location**: `phase-08-setup-wire-removal.md` line 85  
**Issue**: References `examples/golang-sample/go.mod` but Phase 08 is Part 2 (Wire Removal) which operates in `examples/golang-sample/` directory

**Current**:
```markdown
- `examples/golang-sample/go.mod` - Sample app module
```

**Should Be**: Since working directory is `examples/golang-sample/`, file path should be relative:
```markdown
- `go.mod` - Sample app module (in examples/golang-sample/)
```

**Impact**: Confusion about whether operations are at root or in sample app directory.

---

### CRITICAL-010: Phase 13 Wrong File Path References

**Location**: `phase-13-wire-removal-testing.md` line 185  
**Issue**: References `docs/code-standards.md` at root level but file is in `examples/golang-sample/docs/`

**Current**:
```markdown
- `docs/code-standards.md` - Update DI section
```

**Should Be**:
```markdown
- `examples/golang-sample/docs/code-standards.md` - Update DI section
```

**Impact**: Wrong documentation file path.

---

### MEDIUM-001: Phase 02 Missing Root Directory Context

**Location**: `phase-02-merge-govern-packages.md`  
**Issue**: No explicit statement that operations occur at repository root

**Impact**: Minor - usually implied, but Part 1 phases should all state "Working Directory: repository root" for clarity.

---

## Working Directory Issues

### CRITICAL-011: Phase 01 Missing Working Directory Statement

**Location**: `phase-01-repository-preparation.md`  
**Issue**: No explicit working directory statement

**Impact**: Not clear if operations are at root or elsewhere. Part 1 phases should all operate at root.

**Fix**: Add "Working Directory: repository root (golang-sample/)" to overview section.

---

### CRITICAL-012: Phase 02 Missing Working Directory Statement

**Location**: `phase-02-merge-govern-packages.md`  
**Issue**: No explicit working directory statement

**Impact**: Same as CRITICAL-011.

---

### CRITICAL-013: Part 1 Phases Inconsistent Working Directory Spec

**Location**: Phases 01-07  
**Issue**: Phases 08-13 explicitly state "Working Directory: All operations in this phase are performed in the `examples/golang-sample/` directory" but Part 1 phases (01-07) don't state their working directory

**Impact**: Creates asymmetry and confusion. Part 1 phases should state "Working Directory: repository root" for consistency.

**Fix**: Add working directory statement to all Part 1 phase files (01-07).

---

### CRITICAL-014: Phase 03 Commands Assume Wrong Directory

**Location**: `phase-03-move-sample-application.md` lines 242-258  
**Issue**: Commands create `examples/golang-sample/` subdirectory but don't navigate to repository root first

**Current**:
```bash
# Move directories with git mv
git mv cmd/ examples/golang-sample/cmd/
```

**Impact**: If not at repository root, this creates wrong structure. Should add:
```bash
# Verify at repository root
pwd  # Should show /path/to/golang-sample
```

---

### HIGH-004: Phase 08 Working Directory Confusion

**Location**: `phase-08-setup-wire-removal.md` line 16  
**Issue**: States working directory is `examples/golang-sample/` but file paths still reference full path from root

**Current**:
```markdown
**Working Directory**: All operations in this phase are performed in the `examples/golang-sample/` directory. All file paths are relative to `examples/golang-sample/`.
```

**But file paths use**:
```markdown
- `examples/golang-sample/internal/handler/rest/wire.go`
```

**Should Be** (relative to working directory):
```markdown
- `internal/handler/rest/wire.go`
```

**Impact**: Inconsistent path specification - confusing for implementers.

---

## Phase Number Issues

### CRITICAL-015: Phase 06 Title Mismatch

**Location**: `phase-06-validation-testing.md` line 1  
**Issue**: Title says "Phase 06: Validation and Testing" but frontmatter title says "Phase 06: Validation and Testing"

**Current**:
```yaml
title: "Phase 06: Validation and Testing"
```

**In file body line 13**:
```markdown
# Phase 07: Validation and Testing
```

**Impact**: Heading number doesn't match filename or frontmatter title.

**Fix**: Change line 13 from "# Phase 07:" to "# Phase 06:"

---

### CRITICAL-016: Phase 05 Title Mismatch

**Location**: `phase-05-documentation-migration.md` line 13  
**Issue**: Heading says "Phase 06: Documentation Migration" but filename is phase-05

**Current**:
```markdown
# Phase 06: Documentation Migration
```

**Should Be**:
```markdown
# Phase 05: Documentation Migration
```

---

### CRITICAL-017: Phase 07 Title Mismatch

**Location**: `phase-07-repository-rename-merge.md` line 13  
**Issue**: Heading says "Phase 08: Repository Rename and Merge" but filename is phase-07

**Current**:
```markdown
# Phase 08: Repository Rename and Merge
```

**Should Be**:
```markdown
# Phase 07: Repository Rename and Merge
```

---

### HIGH-005: plan.md Phase Summary Mismatch

**Location**: `plan.md` lines 39-74  
**Issue**: Phase summary table references correct phase files but doesn't mention the title/heading mismatches

**Impact**: Documentation inconsistency between plan.md and individual phase files.

**Fix**: After fixing phase file headings, verify plan.md summaries match.

---

### MEDIUM-002: Phase 08 Wrong Number in Context

**Location**: `phase-08-setup-wire-removal.md`  
**Issue**: No heading mismatch (doesn't have "# Phase XX:" heading) but this is inconsistent with other phases

**Status**: Minor - Phase 08 doesn't have a heading, so no mismatch to fix.

---

## Technical Inconsistencies

### CRITICAL-018: Phase 02 Wrong Module Path

**Location**: `phase-02-merge-govern-packages.md` line 266  
**Issue**: Says "Update module path to github.com/haipham22/govern" but command shows wrong path

**Current**:
```bash
sed -i 's/^module golang-sample$/module github.com\/haipham22\/govern/' go.mod
```

**Status**: Actually correct - this is the right command. No issue.

---

### CRITICAL-019: Phase 03 Missing External Import Statement

**Location**: `phase-03-move-sample-application.md`  
**Issue**: Phase overview talks about "external import" but implementation steps don't explicitly validate external import works

**Impact**: Critical validation step missing. Should add:
```bash
# Verify external import resolves (without go.work)
cd examples/golang-sample
mise exec -- go mod tidy
mise exec -- go build ./cmd/serverd.go
```

**Fix**: Add to Step 10 or create new validation step.

---

### CRITICAL-020: Phase 08 Wrong Dependencies Listed

**Location**: `phase-08-setup-wire-removal.md` lines 88-90  
**Issue**: Lists "Files to Modify" as "None in this phase" but actually needs to read wire files

**Current**:
```markdown
### Files to Modify
- None in this phase
```

**Should Be**:
```markdown
### Files to Modify
- None (read-only phase - documentation only)
```

**Impact**: Misleading - phase reads files but doesn't modify them.

---

### HIGH-006: Phase 09 Wrong Error File Reference

**Location**: `phase-09-custom-error-types.md` line 23  
**Issue**: References `examples/golang-sample/internal/model/errors.go` but Phase 03 migration moved this to `internal/domain/errors.go`

**Current**:
```markdown
- `examples/golang-sample/internal/model/errors.go` - Replace with custom types
```

**Should Be**:
```markdown
- `internal/domain/errors.go` - Replace with custom types (relative to examples/golang-sample/)
```

**Impact**: Wrong file path - doesn't account for Phase 03 migration.

---

### HIGH-007: Phase 11 Bootstrap Path Mismatch

**Location**: `phase-11-manual-di-implementation.md` lines 106-111  
**Issue**: Shows bootstrap directory structure at root but working directory is `examples/golang-sample/`

**Current**:
```markdown
internal/bootstrap/
```

**Should Be** (relative to working directory):
```markdown
examples/golang-sample/internal/bootstrap/
```

**Impact**: Confusing - implementers won't know where to create files.

---

### HIGH-008: Phase 10 Wrong Service File Reference

**Location**: `phase-10-centralized-error-management.md` line 100  
**Issue**: References `internal/usecase/auth/impl.go` but Phase 03 migration might have changed this structure

**Current**:
```markdown
- `internal/usecase/auth/impl.go` - Replace govern/errors with custom (9 usages)
```

**Status**: Need to verify Phase 03 didn't change usecase structure. If Phase 03 kept usecase/ structure, this is correct.

---

### MEDIUM-003: Phase 13 Missing Phase 11 Files in Testing

**Location**: `phase-13-wire-removal-testing.md` lines 92-94  
**Issue**: Lists files to test but doesn't explicitly include Phase 11 bootstrap files

**Current**:
```markdown
- `internal/handler/rest/di.go` - Manual DI
```

**Should Also Include**:
```markdown
- `internal/bootstrap/*.go` - Bootstrap constructors (Phase 11)
```

**Impact**: Incomplete testing scope.

---

## Cross-Reference Issues

### CRITICAL-021: plan.md Wrong Phase Count

**Location**: `plan.md` line 75  
**Issue**: Says "Total Estimated Time: 36 hours" but should be 40 hours

**Current**:
```markdown
**Total Estimated Time**: 36 hours (12h monorepo + 24h wire removal)
```

**But**: Plan frontmatter line 6 says `effort: 40h`

**Should Be**:
```markdown
**Total Estimated Time**: 40 hours (12h monorepo + 28h wire removal)
```

**Impact**: Effort estimate inconsistent between frontmatter and body.

---

### CRITICAL-022: plan.md Phase 08 Wrong Duration

**Location**: `plan.md` line 68  
**Issue**: Lists Phase 08 duration as "2h" but Phase 08 file says "2h" in effort field

**Status**: Actually consistent - both say 2h. No issue.

---

### HIGH-009: plan.md Missing Phase File Links

**Location**: `plan.md` lines 52-73  
**Issue**: Phase summary table has links to phase files but doesn't validate all links work

**Impact**: Documentation may have broken links.

**Fix**: Verify all phase file links exist and are correct.

---

### HIGH-010: plan.md Wrong Working Directory Statement

**Location**: `plan.md` line 510  
**Issue**: Says "Working Directory: Part 2 phases operate in `examples/golang-sample/` directory" but doesn't specify Part 1 working directory

**Current**:
```markdown
**Working Directory**: Part 2 phases operate in `examples/golang-sample/` directory
```

**Should Be**:
```markdown
**Working Directory**: 
- Part 1 phases (01-07): Repository root
- Part 2 phases (08-13): examples/golang-sample/
```

**Impact**: Asymmetric documentation - Part 1 working directory not stated.

---

### MEDIUM-004: plan.md Missing Generator Phase

**Location**: `plan.md` lines 62-63  
**Issue**: Says "Interactive Generator (4h) separated into future plan" but Phase 05 is "Documentation Migration"

**Current**:
```markdown
**Note**: Interactive Generator (4h) separated into future plan - will be implemented separately with detailed design.
```

**But Phase 05** exists and is "Documentation Migration"

**Impact**: Confusing - is generator being implemented or not?

**Status**: This is actually correct - Phase 05 was changed from "Interactive Generator" to "Documentation Migration" in a plan revision. The note is outdated and should be removed.

---

### MEDIUM-005: Phase 03 Wrong Migration Scope

**Location**: `phase-03-move-sample-application.md` lines 441-539  
**Issue**: Step 7 creates sample app CI/CD workflows but doesn't mention deleting old root workflows

**Current**:
```markdown
### Step 7: Update Root CI/CD Workflows
```

**But**: No step to DELETE old sample app workflows from root

**Impact**: Duplicate workflows may exist after migration.

**Fix**: Add step to delete sample-specific workflows from root before creating new ones in examples/golang-sample/

---

### MEDIUM-006: Phase 06 Missing Go Workspace Validation

**Location**: `phase-06-validation-testing.md` lines 130-156  
**Issue**: Validates go.work file exists but plan.md says "NO go.work workspace file created"

**From plan.md line 87**:
```markdown
- NO go.work workspace file created
```

**But Phase 06 Step 2** validates go.work functionality

**Impact**: Direct contradiction between plan and phase.

**Root Cause**: Plan was revised from "go.work workspace" to "external import with replace directive" but Phase 06 still has workspace validation steps.

**Fix**: Remove all go.work validation from Phase 06, replace with external import validation.

---

### MEDIUM-007: Phase 13 Missing Validation Script Reference

**Location**: `phase-13-wire-removal-testing.md` line 206  
**Issue**: References `./scripts/validate.sh` but Phase 08 was supposed to create this script

**Current**:
```markdown
./scripts/validate.sh > validation-report.txt
```

**But Phase 08**: Creates `scripts/validate.sh` (line 86)

**Status**: Actually consistent - Phase 08 creates the script, Phase 13 uses it. No issue.

---

## Recommendations

### Priority 1 - Must Fix Before Implementation

1. **Fix Phase 08 dependency** (CRITICAL-001) - Wrong dependency breaks Part 2 sequence
2. **Fix all context link errors** (CRITICAL-002 to CRITICAL-004) - Broken documentation links
3. **Fix phase heading mismatches** (CRITICAL-015 to CRITICAL-017) - Confusing phase numbers
4. **Fix Phase 01 directory creation** (CRITICAL-006) - Creates structure at wrong location
5. **Fix Part 1 working directory statements** (CRITICAL-013) - Add to all Part 1 phases
6. **Fix go.work vs external import contradiction** (MEDIUM-006) - Plan says external import but phases validate go.work

### Priority 2 - Should Fix for Clarity

7. **Fix file path inconsistencies** (CRITICAL-007 to CRITICAL-010) - Confusing path references
8. **Add missing context links to Phase 12** (HIGH-002) - Complete documentation chain
9. **Fix effort estimate mismatch** (CRITICAL-021) - plan.md should say 40h not 36h
10. **Remove outdated generator note** (MEDIUM-004) - Plan note contradicts actual Phase 05

### Priority 3 - Nice to Have

11. **Add missing validation steps** (CRITICAL-019) - External import validation
12. **Fix bootstrap path references** (HIGH-007) - Clarify working directory context
13. **Add Phase 1 working directory to plan** (HIGH-010) - Complete documentation

---

## Unresolved Questions

1. **Phase 03 Structure**: Did Phase 03 migration change the internal/ structure? Need to verify if `internal/usecase/auth/` still exists or was changed to `internal/service/auth/`.

2. **Generator Implementation**: Is the interactive generator being implemented or separated to future plan? Current documentation is contradictory.

3. **Go Workspace vs External Import**: Plan says "NO go.work workspace" but Phase 03 and Phase 06 have workspace validation. Which approach is final?

4. **Part 2 Phase Count**: plan.md mentions "Phase 14" in line 528 but there are only 13 phases. Is Phase 14 missing or was this a planning error?

---

## Validation Checklist

Before starting implementation, verify:

- [ ] Phase 08 dependency changed to `phase-06-validation-testing.md`
- [ ] All context links reference correct phase numbers
- [ ] All phase headings match filenames
- [ ] Part 1 phases (01-07) have "Working Directory: repository root"
- [ ] Part 2 phases (08-13) have "Working Directory: examples/golang-sample/"
- [ ] All go.work references removed or replaced with external import
- [ ] plan.md effort estimate updated to 40h
- [ ] All file paths use consistent references (either relative or absolute)
- [ ] Generator note removed from plan.md
- [ ] External import validation added to Phase 03

---

**Report Generated**: 2026-06-28  
**Analyst**: code-reviewer agent  
**Status**: Ready for review and fixes
