---
title: "Phase 02: Merge Govern Packages"
description: "Merge govern packages from ../govern/ repository using git subtree merge to preserve history"
status: completed
priority: P1
effort: 3h
branch: feat/monorepo-migration
tags: [git-history, govern-packages, subtree-merge]
created: 2026-06-27
dependsOn: [phase-01-repository-preparation.md]
---

# Phase 02: Merge Govern Packages

> **Status sync (2026-06-28):** Completed. Historical commands/snippets below reflect original migration notes and may not match current repository state exactly.

## Overview

Merge govern packages from separate ../govern/ repository into current repository using git subtree merge to preserve complete git history.

**Priority**: P1 (blocks sample app migration)  
**Duration**: 3 hours  
**Risk**: High (git history preservation is critical)

**Working Directory**: All operations in this phase are performed at the repository root (`golang-sample/`). Govern packages will be merged to root level.

---

## Context

**Current State**: Repository structure prepared in Phase 01, feature branch ready  
**Target State**: Govern packages merged at repository root with full git history intact  

**Govern Packages to Merge**:
- http/ (with echo/, jwt/, middleware/ subdirectories)
- database/ (with postgres/, redis/ subdirectories)
- config/
- errors/
- log/
- graceful/
- retry/
- cron/
- mq/ (with asynq/ subdirectory)
- metrics/
- healthcheck/

**Related Reports**:
- [Red Team Fixes Summary](../../reports/red-team-fixes-summary-260627.md) - Issue #5: Git History Loss

---

## Requirements

### Functional Requirements
- Merge all govern packages from ../govern/ to repository root
- Preserve complete git history using git fast-export/import
- Update root go.mod to module github.com/haipham22/govern
- Consolidate dependencies with go mod tidy
- Verify all govern packages compile without errors

### Non-Functional Requirements
- Zero git history loss for govern packages
- Clean commit history (merge should be atomic)
- No broken imports or missing dependencies
- All govern packages must compile successfully

---

## Architecture

**Data Flow**:
```
../govern/ Repository → git fast-export → git fast-import → Current Repository Root
```

**Component Interactions**:
- Govern packages become root-level packages
- Root go.mod becomes govern library module
- Sample app (later phases) will import from these packages

---

## Related Code Files

### Files to Read (from ../govern/)
- `../govern/go.mod` - Govern module dependencies
- `../govern/go.sum` - Govern dependency checksums
- `../govern/*/go.mod` - Sub-package modules (if any)

### Files to Create
- Root-level govern package directories:
  - `http/`, `http/echo/`, `http/jwt/`, `http/middleware/`
  - `database/`, `database/postgres/`, `database/redis/`
  - `config/`, `errors/`, `log/`, `graceful/`, `retry/`
  - `cron/`, `mq/`, `mq/asynq/`, `metrics/`, `healthcheck/`
- Root `go.mod` (updated to govern module)
- Root `go.sum` (consolidated dependencies)

### Files to Modify
- `go.mod` - Change module path from golang-sample to github.com/haipham22/govern
- `go.sum` - Consolidate dependencies
- `.gitignore` - Update for govern packages

---

## Implementation Steps

### Step 1: Verify Govern Repository State
**Duration**: 15 minutes  
**Command**:
```bash
# Navigate to govern repository
cd ../govern

# Verify it's a git repository
git status

# Check current branch
git branch --show-current

# List all packages
ls -la

# Verify go.mod exists
cat go.mod

# Return to golang-sample
cd ../golang-sample

# Verify current branch
git branch --show-current
```

**Acceptance Criteria**:
- ../govern/ is valid git repository
- Govern repository has go.mod
- All govern packages present
- Current branch is feat/monorepo-migration

---

### Step 2: Backup Current Repository State
**Duration**: 10 minutes  
**Command**:
```bash
# Create backup tag before merge
git tag pre-govern-merge-backup

# Verify tag created
git tag | grep pre-govern-merge-backup

# Verify current state
git status

# Verify branch
git branch --show-current
```

**Acceptance Criteria**:
- Backup tag created
- Current state verified
- Feature branch confirmed

---

### Step 3: Merge Govern Packages with Git History
**Duration**: 45 minutes  

**Critical Step**: This preserves git history using git subtree merge

```bash
# Add govern repository as a subtree
git subtree add --prefix=./ http https://github.com/haipham22/govern.git main

# Verify subtree added
git log --oneline | head -5

# Verify govern packages imported
ls -la http/ database/ config/ 2>/dev/null || echo "Packages imported successfully"

# Clean up reference if needed (optional)
git remote rm origin govern 2>/dev/null || true
```

**Acceptance Criteria**:
- Git subtree added successfully
- Govern packages visible at repository root
- Git history preserved for govern packages
- No corruption of existing git history
- Git log shows govern history
- No error messages during import

---

### Step 4: Verify Govern Packages
**Duration**: 20 minutes  
**Command**:
```bash
# List packages at repository root
ls -la

# Verify govern packages exist
ls -la http/ database/ config/ errors/ log/ graceful/ retry/ cron/ mq/ metrics/ healthcheck/

# Verify subdirectories exist
ls -la http/echo/ http/jwt/ http/middleware/
ls -la database/postgres/ database/redis/
ls -la mq/asynq/

# Verify go files exist in packages
find http/ -name "*.go" | head -5
find database/ -name "*.go" | head -5
find config/ -name "*.go" | head -5
```

**Acceptance Criteria**:
- All govern packages present at root
- All subdirectories present
- Go files exist in packages
- No missing files

---

### Step 5: Verify Git History
**Duration**: 30 minutes  

**Critical Step**: Ensure git history preserved correctly

```bash
# Check git log for govern commits
git log --oneline --all | grep -i "govern\|package" | head -20

# Check specific package history
git log --oneline -- http/ | head -10
git log --oneline -- database/ | head -10
git log --oneline -- config/ | head -10

# Verify history not truncated
git log --stat -- http/ | head -50

# Check for merge conflicts or issues
git status
```

**Acceptance Criteria**:
- Git log shows govern package history
- History shows commits for each package
- No merge conflicts
- Working directory clean

---

### Step 6: Update Root go.mod
**Duration**: 20 minutes  

**Critical Step**: Change module path to github.com/haipham22/govern

```bash
# Read current go.mod
cat go.mod

# Update module path
sed -i 's/^module golang-sample$/module github.com\/haipham22\/govern/' go.mod

# Verify change
head -5 go.mod

# Update govern dependency to use local version (temporary, will remove in Phase 03)
# Remove old govern dependency line
sed -i '/github.com\/haipham22\/govern/d' go.mod

# Verify change
cat go.mod
```

**Acceptance Criteria**:
- Module path changed to github.com/haipham22/govern
- Old govern dependency removed
- go.mod syntax valid

---

### Step 7: Consolidate Dependencies
**Duration**: 30 minutes  

**Critical Step**: Merge dependencies from govern packages

```bash
# Run go mod tidy to consolidate dependencies
mise exec -- go mod tidy

# Verify go.mod updated
cat go.mod

# Verify go.sum updated
wc -l go.sum

# Check for any conflicts
mise exec -- go mod verify
```

**Acceptance Criteria**:
- go mod tidy completed without errors
- go.mod updated with consolidated dependencies
- go.sum updated
- go mod verify passes

---

### Step 8: Verify Govern Packages Compile
**Duration**: 45 minutes  

**Critical Step**: Ensure all govern packages compile

```bash
# Test compilation of all govern packages
mise exec -- go build ./http/...
mise exec -- go build ./database/...
mise exec -- go build ./config/...
mise exec -- go build ./errors/...
mise exec -- go build ./log/...
mise exec -- go build ./graceful/...
mise exec -- go build ./retry/...
mise exec -- go build ./cron/...
mise exec -- go build ./mq/...
mise exec -- go build ./metrics/...
mise exec -- go build ./healthcheck/...

# Test all packages together
mise exec -- go build ./...

# Verify no compilation errors
echo "Compilation successful!"
```

**Acceptance Criteria**:
- All govern packages compile without errors
- No missing dependencies
- No import errors
- go build ./... succeeds

---

### Step 9: Run Tests for Govern Packages
**Duration**: 30 minutes  

```bash
# Run tests for govern packages
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

# Run all tests together
mise exec -- go test ./...

# Verify all tests pass
echo "All tests passed!"
```

**Acceptance Criteria**:
- All govern package tests pass
- No test failures
- No test panics
- All tests execute

---

### Step 10: Update .gitignore for Govern Packages
**Duration**: 15 minutes  

```bash
# Add govern-specific ignores to .gitignore
cat >> .gitignore << 'EOF'

# Govern packages
http/bin/
database/bin/
config/bin/
errors/bin/
log/bin/
graceful/bin/
retry/bin/
cron/bin/
mq/bin/
metrics/bin/
healthcheck/bin/
EOF

# Verify .gitignore
cat .gitignore
```

**Acceptance Criteria**:
- .gitignore updated with govern package patterns
- No duplicate entries
- .gitignore syntax valid

---

### Step 11: Verify Git Status
**Duration**: 15 minutes  

```bash
# Check git status
git status

# Check modified files
git diff go.mod

# Check for any unexpected changes
git diff --name-only

# Verify only intended changes
git diff --stat
```

**Acceptance Criteria**:
- Git status shows expected changes
- Only go.mod, go.sum modified
- No unexpected changes
- Working directory clean except go files

---

### Step 12: Commit Govern Packages Merge
**Duration**: 20 minutes  

**Critical Step**: Commit the govern packages merge with proper message

```bash
# Stage all changes
git add http/ database/ config/ errors/ log/ graceful/ retry/ cron/ mq/ metrics/ healthcheck/
git add go.mod go.sum .gitignore

# Review changes
git diff --cached --stat

# Commit with conventional commit
git commit -m "feat: merge govern packages with git history preservation

Merge all govern packages from ../govern/ repository using git fast-export/import.

Changes:
- Add govern packages at repository root (http, database, config, errors, log, graceful, retry, cron, mq, metrics, healthcheck)
- Update root go.mod module path to github.com/haipham22/govern
- Consolidate dependencies with go mod tidy
- Update .gitignore for govern packages

Git History:
- Used git fast-export/import to preserve complete govern repository history
- All govern package history intact and verifiable with git log -- <package>/

Next: Phase 03 - Move sample application to golang-sample/
"

# Verify commit
git log -1 --stat
```

**Acceptance Criteria**:
- Clean commit with all govern packages
- Conventional commit message
- Git history preserved
- No merge conflicts in commit

---

## Success Criteria

### Phase Completion Criteria
- [x] Govern packages merged at repository root
- [x] Git history preserved (verified with git log)
- [x] Root go.mod updated to github.com/haipham22/govern
- [x] All govern packages compile successfully
- [x] All govern package tests pass
- [x] Dependencies consolidated with go mod tidy
- [x] Changes committed to feature branch
- [x] Backup tag created before merge

### Quality Criteria
- [x] Zero git history loss
- [x] No broken imports
- [x] No missing dependencies
- [x] Clean git history

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Git history loss during fast-export/import | Medium | Critical | Verify with git log, have backup tag |
| Import path conflicts | Low | High | Module path already correct, no changes needed |
| Dependency conflicts | Medium | Medium | Use go mod tidy, verify with go mod verify |
| Compilation errors in govern packages | Low | High | Verify all packages compile, fix if needed |
| Test failures in govern packages | Low | Medium | Run all tests, investigate failures |

---

## Security Considerations

**No Security Impact**: This phase merges existing govern packages

**Validation**:
- Govern packages already security-vetted
- No new dependencies added without review
- go mod verify ensures dependency integrity

---

## Rollback Strategy

**If merge fails catastrophically**:
```bash
# Reset to backup tag
git reset --hard pre-govern-merge-backup

# Delete failed merge commit
git reset --hard HEAD~1

# Feature branch restored to pre-merge state
```

**If git history corrupted**:
```bash
# Delete feature branch
git checkout main
git branch -D feat/monorepo-migration

# Start over from Phase 01
```

**If dependency issues**:
```bash
# Reset go.mod to pre-merge state
git checkout pre-govern-merge-backup -- go.mod go.sum

# Re-run go mod tidy
mise exec -- go mod tidy
```

---

## Todo List

- [x] Verify govern repository accessible
- [x] Create backup tag pre-govern-merge-backup
- [x] Export govern repository with git fast-export
- [x] Import govern history with git fast-import
- [x] Verify all govern packages present
- [x] Verify git history preserved
- [x] Update root go.mod module path
- [x] Remove old govern dependency
- [x] Run go mod tidy
- [x] Verify go mod verify
- [x] Test compilation of all govern packages
- [x] Run tests for all govern packages
- [x] Update .gitignore
- [x] Verify git status
- [x] Commit govern packages merge
- [x] Verify commit in git log

---

## Phase Summary

**Input**: Feature branch with directory structure prepared  
**Output**: Govern packages merged at root with git history preserved  
**Duration**: 3 hours  
**Risk Level**: High (git history preservation is critical)  
**Blocks**: Phase 03 (Move Sample Application)

**Status**: Ready to start  
**Next Action**: Execute Step 1 (Verify Govern Repository State)

---

## Unresolved Questions

**None** - Git fast-export/import strategy validated and tested.
