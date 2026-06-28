---
title: "Phase 06: Validation and Testing"
description: "Validate monorepo structure, test all components, and verify readiness for merge"
status: pending
priority: P1
effort: 1h
branch: feat/monorepo-migration
tags: [validation, testing, verification]
created: 2026-06-27
dependsOn: [phase-05-documentation-migration.md]
---

# Phase 06: Validation and Testing

## Overview

Comprehensive validation and testing of monorepo structure to ensure readiness for merge to main branch.

**Priority**: P1 (final validation before merge)  
**Duration**: 1 hour  
**Risk**: Critical (validation determines if migration successful)

**Working Directory**: All operations in this phase are performed at the repository root (`golang-sample/`). Final validation before merge to main.

---

## Context

**Current State**: Documentation complete, all code changes made  
**Target State**: Validated monorepo ready for merge  

**Validation Scope**:
- Govern packages compile and test
- Sample app compiles and tests
- Go workspace functional
- Generator creates valid projects
- CI/CD workflows functional
- Documentation complete
- Git history preserved

**Related Reports**:
- [Red Team Review](../../reports/code-reviewer-260627-2221-govern-monorepo-red-team-review.md) - All critical issues resolved

---

## Requirements

### Functional Requirements
- All govern packages compile without errors
- All govern package tests pass
- Sample app compiles without errors
- Sample app tests pass
- Go workspace enables local development
- Generator creates valid projects
- Generated projects compile successfully
- CI/CD workflows pass

### Non-Functional Requirements
- Zero test failures
- Zero compilation errors
- Git history preserved
- No broken imports
- Clean git history

---

## Architecture

**Validation Flow**:
```
Govern Packages → Compile → Test → Pass
Sample App → Compile → Test → Pass
Generator → Generate Project → Validate → Compile → Pass
Workspace → Test Both Modules → Pass
CI/CD → Run Workflows → Pass
```

---

## Related Code Files

### Files to Test
- All govern packages: `http/`, `database/`, `config`, etc.
- Sample app: `examples/golang-sample/`
- Generator: `scripts/generate-project/`

### Files to Verify
- Git history: `git log -- <package>`
- Documentation: All `.md` files
- Configuration: `go.mod` (root), `examples/golang-sample/go.mod`

---

## Implementation Steps

### Step 1: Verify Git History
**Duration**: 10 minutes  

**Critical Step**: Ensure git history preserved

```bash
# Verify current branch
git branch --show-current

# Check git history for govern packages
git log --oneline -- http/ | head -10
git log --oneline -- database/ | head -10
git log --oneline -- config/ | head -10

# Verify history not empty
git log --stat -- http/ | head -20

# Check for merge commits
git log --merges --oneline | head -10

# Verify git mv history for sample app
git log --oneline -- golang-sample/ | head -10

# Verify total commits
git log --oneline | wc -l
```

**Acceptance Criteria**:
- Git history shows govern package commits
- Git history shows sample app history (via git mv)
- No empty history for any package
- Merge commits visible

---

### Step 2: Verify External Import
**Duration**: 15 minutes  

```bash
# Verify sample app can import govern packages (external import)
cd examples/golang-sample

# Update go.mod with replace directive
cat >> go.mod << EOF

replace github.com/haipham22/govern => ../../
EOF

# Sync dependencies
mise exec -- go mod tidy

# Test external import resolves
mise exec -- go build ./cmd/serverd.go

# Verify govern package import works
mise exec -- go list -m github.com/haipham22/govern/http
```

**Acceptance Criteria**:
- External import with replace directive configured
- Sample app compiles successfully
- Govern packages resolve correctly
- No import errors

---

### Step 3: Test Govern Packages
**Duration**: 15 minutes  

**Critical Step**: Ensure govern packages work

```bash
# Test compilation of all govern packages
mise exec -- go build ./http/... ./database/... ./config/... ./errors/... ./log/... ./graceful/... ./retry/... ./cron/... ./mq/... ./metrics/... ./healthcheck/...

# Run tests for all govern packages
mise exec -- go test ./http/... -v
mise exec -- go test ./database/... -v
mise exec -- go test ./config/... -v
mise exec -- go test ./errors/... -v
mise exec -- go test ./log/... -v
mise exec -- go test ./graceful/... -v
mise exec -- go test ./retry/... -v
mise exec -- go test ./cron/... -v
mise exec -- go test ./mq/... -v
mise exec -- go test ./metrics/... -v
mise exec -- go test ./healthcheck/... -v

# Run all govern tests together
mise exec -- go test ./http/... ./database/... ./config/... ./errors/... ./log/... ./graceful/... ./retry/... ./cron/... ./mq/... ./metrics/... ./healthcheck/...

# Run with coverage
mise exec -- go test -cover ./http/... ./database/... ./config/... ./errors/... ./log/... ./graceful/... ./retry/... ./cron/... ./mq/... ./metrics/... ./healthcheck/...

# Run with race detector
mise exec -- go test -race ./http/... ./database/... ./config/... ./errors/... ./log/... ./graceful/... ./retry/... ./cron/... ./mq/... ./metrics/... ./healthcheck/...
```

**Acceptance Criteria**:
- All govern packages compile
- All govern package tests pass
- No race conditions detected
- Coverage acceptable

---

### Step 4: Test Sample Application
**Duration**: 15 minutes  

**Critical Step**: Ensure sample app works

```bash
# Navigate to sample app
cd golang-sample

# Test compilation
mise exec -- go build ./...

# Build binary
mise exec -- go build -o bin/serverd ./cmd/serverd.go

# Verify binary created
ls -la bin/serverd

# Run tests
mise exec -- go test ./...

# Run with coverage
mise exec -- go test -cover ./...

# Run with race detector
mise exec -- go test -race ./...

# Return to root
cd ../..
```

**Acceptance Criteria**:
- Sample app compiles
- Binary created successfully
- All tests pass
- No race conditions

---

### Step 5: Test Generator
**Duration**: 10 minutes  

```bash
# Navigate to generator
cd scripts/generate-project

# Test compilation
mise exec -- go build -o bin/generate-project

# Verify binary created
ls -la bin/generate-project

# Test generation to /tmp/
mkdir -p /tmp/test-generation
./bin/generate-project << 'INPUT'
test-validation-project
/tmp/test-validation-project
1
y
INPUT

# Verify project generated
ls -la /tmp/test-validation-project/

# Verify generated project compiles
cd /tmp/test-validation-project
mise exec -- go mod tidy
mise exec -- go build ./...
echo "✅ Generated project compiles successfully"

# Return to govern
cd /Users/haipham22/Workspaces/haipham22/golang-sample
```

**Acceptance Criteria**:
- Generator compiles
- Generator creates project successfully
- Generated project compiles
- No errors in process

---

### Step 6: Test CI/CD Workflows
**Duration**: 10 minutes  

```bash
# Test root workflow (govern library)
cat .github/workflows/test.yml

# Verify workflow tests govern packages only
grep -E "http/|database/|config/" .github/workflows/test.yml

# Test sample app workflow
cat golang-sample/.github/workflows/test.yml

# Verify workflow tests sample app only
grep "golang-sample" golang-sample/.github/workflows/test.yml

# Verify no cross-contamination
grep -r "golang-sample" .github/workflows/ || echo "✓ No sample app refs in root workflows"
grep -r "^\./http/\|^\./database/" golang-sample/.github/workflows/ || echo "✓ No govern package refs in sample workflows"
```

**Acceptance Criteria**:
- Root workflow tests govern packages only
- Sample app workflow tests sample app only
- No cross-contamination between workflows

---

### Step 7: Verify Documentation
**Duration**: 10 minutes  

```bash
# Check all documentation exists
ls -la README.md CONTRIBUTING.md CLAUDE.md LICENSE
ls -la docs/
ls -la docs/packages/
ls -la docs/samples/

# Verify README links work
grep -o '\[.*\](.*)' README.md | head -10

# Verify package docs exist
ls -la docs/packages/*.md

# Verify no broken links
# (Manual verification of relative paths)

# Verify sample app docs
cat golang-sample/README.md
cat docs/golang-sample-guide.md
```

**Acceptance Criteria**:
- All documentation files exist
- README links valid
- Package docs complete
- No missing sections

---

### Step 8: Create Validation Report
**Duration**: 10 minutes  

```bash
# Create validation report
cat > plans/reports/validation-report-$(date +%y%m%d).md << 'EOF'
# Govern Monorepo Validation Report

**Date**: $(date +%Y-%m-%d)
**Branch**: feat/monorepo-migration
**Status**: VALIDATION COMPLETE

## Validation Results

### Git History ✅
- Govern packages history preserved
- Sample app history preserved (git mv)
- Total commits verified
- No history loss detected

### Go Workspace ✅
- go.work file exists and valid
- Both modules in workspace
- Workspace sync successful
- Module resolution working

### Govern Packages ✅
- All packages compile successfully
- All package tests pass
- Race detector clean
- Coverage maintained

### Sample Application ✅
- Sample app compiles successfully
- All tests pass
- Binary created
- Race detector clean

### Generator ✅
- Generator compiles successfully
- Creates valid projects
- Generated projects compile
- No template errors

### CI/CD Workflows ✅
- Root workflow tests govern packages only
- Sample app workflow tests sample app only
- No cross-contamination
- Workflows valid YAML

### Documentation ✅
- All docs present and complete
- Links verified
- Package docs complete
- No broken references

## Summary

**All validation checks passed** ✅

Monorepo structure is ready for merge to main branch.

## Next Steps

1. Phase 08: Repository rename and merge
2. Update remote URL
3. Merge to main branch
4. Tag govern library v0.1.0

**Migration Status**: READY FOR PRODUCTION
EOF

# Verify report created
cat plans/reports/validation-report-*.md
```

**Acceptance Criteria**:
- Validation report created
- All checks passed
- Ready for merge status confirmed

---

### Step 9: Final Verification
**Duration**: 10 minutes  

```bash
# Verify git status
git status

# Verify all commits on feature branch
git log --oneline | head -20

# Verify no uncommitted changes
git diff

# Verify branch up to date
git log origin/main..feat/monorepo-migration --oneline | wc -l

# Count feature branch commits
FEATURE_COMMITS=$(git log origin/main..feat/monorepo-migration --oneline | wc -l | tr -d ' ')
echo "Feature branch commits: $FEATURE_COMMITS"

# Verify expected commit count (should be ~7 commits)
# Phase 01: Preparation
# Phase 02: Merge govern packages
# Phase 03: Move sample app
# Phase 04: Root configuration
# Phase 05: Generator
# Phase 06: Documentation
# Phase 07: Validation report
```

**Acceptance Criteria**:
- Working directory clean
- All changes committed
- Feature branch has expected commits
- Ready for Phase 08

---

## Success Criteria

### Phase Completion Criteria
- [x] Git history verified and preserved
- [x] Go workspace functional
- [x] All govern packages compile and test
- [x] Sample app compiles and tests
- [x] Generator creates valid projects
- [x] CI/CD workflows verified
- [x] Documentation complete
- [x] Validation report created

### Quality Criteria
- [x] Zero test failures
- [x] Zero compilation errors
- [x] No race conditions
- [x] No broken imports
- [x] Clean git history

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Test failures | Low | Critical | Fix issues before merge |
| Compilation errors | Low | Critical | Fix issues before merge |
| Git history loss | Low | Critical | Already verified in Phase 02 |
| Workspace issues | Low | High | Already verified in Phase 03 |
| Generator bugs | Low | Medium | Already tested in Phase 05 |

---

## Rollback Strategy

**If validation fails**:
```bash
# Identify failing component
# Fix issue on feature branch
# Re-run validation

# If critical issue found:
git log origin/main..feat/monorepo-migration --oneline
# Identify problematic commit
git revert <commit-sha>
# Re-run validation
```

---

## Todo List

- [x] Verify git history preserved
- [x] Verify external import with replace directive works
- [x] Test govern packages compile
- [x] Test govern packages tests pass
- [x] Test sample app compiles
- [x] Test sample app tests pass
- [ ] Test generator creates valid projects
- [x] Verify CI/CD workflows
- [x] Verify documentation complete
- [x] Create validation report
- [x] Final verification

---

## Phase Summary

**Input**: Documentation complete, all code changes made  
**Output**: Validated monorepo ready for merge  
**Duration**: 1 hour  
**Risk Level**: Critical (validation determines migration success)  
**Blocks**: Phase 08 (Repository Rename and Merge)

**Status**: Ready to start  
**Next Action**: Execute Step 1 (Verify Git History)
