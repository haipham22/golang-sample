---
title: "Phase 01: Repository Preparation"
description: "Create feature branch, document current state, and prepare directory structure for migration"
status: completed
priority: P1
effort: 2h
branch: feat/monorepo-migration
tags: [preparation, branching, directory-structure]
created: 2026-06-27
---

# Phase 01: Repository Preparation

> **Status sync (2026-06-28):** Completed. Historical commands/snippets below reflect original migration notes and may use pre-migration paths.

## Overview

Prepare repository for monorepo restructuring by creating feature branch, documenting current state, and setting up directory structure.

**Priority**: P1 (blocks all subsequent phases)  
**Duration**: 2 hours  
**Risk**: Low (preparatory work, no destructive changes)

**Working Directory**: All operations in this phase are performed at the repository root (`golang-sample/`). This is before the monorepo structure is created.

---

## Context

**Current State**: Repository named golang-sample with sample application code at root  
**Target State**: Feature branch with prepared directory structure ready for govern packages merge  

**Related Reports**:
- [Govern Monorepo Restructure - Brainstorm Report](../../reports/brainstorm-260627-2136-govern-monorepo-restructure.md)
- [Red Team Review - Critical Issues](../../reports/code-reviewer-260627-2221-govern-monorepo-red-team-review.md)

---

## Requirements

### Functional Requirements
- Create feature branch for migration work
- Document current import paths and dependencies
- Create directory structure for examples/, templates/, scripts/generate-project/
- Verify ../govern/ repository exists and is accessible
- Document current go.mod dependencies

### Non-Functional Requirements
- No changes to main branch
- Clean git history for feature branch
- Documentation must be accurate and complete

---

## Architecture

**Data Flow**:
```
Main Branch → Feature Branch → Directory Setup → Ready for Phase 02
```

**Component Interactions**:
- Feature branch protects main branch from experimental changes
- Directory structure prepares for govern packages import
- Documentation serves as reference during migration

---

## Related Code Files

### Files to Read
- `go.mod` - Current dependencies and module path
- `go.sum` - Current dependency checksums
- `.gitignore` - Current ignore patterns
- `README.md` - Current project documentation

### Files to Create
- `examples/` - Directory for sample applications
- `templates/` - Directory for project templates
- `scripts/generate-project/` - Directory for generator tool
- `plans/reports/pre-migration-state.md` - Documentation of current state

### Files to Modify
- `.gitignore` - Add ignores for examples/, templates/
- `README.md` - Document migration in progress (feature branch only)

---

## Implementation Steps

### Step 1: Verify Prerequisites
**Duration**: 10 minutes  
**Command**:
```bash
# Verify Go version
mise exec -- go version

# Verify current branch
git branch --show-current

# Verify govern repository exists
ls -la ../govern/

# Verify govern repository is git repo
cd ../govern && git status
```

**Acceptance Criteria**:
- Go 1.25+ installed
- Currently on main branch
- ../govern/ directory exists and is a git repository

---

### Step 2: Create Feature Branch
**Duration**: 5 minutes  
**Command**:
```bash
# Ensure clean working directory
git status

# Stash any uncommitted changes if needed
git stash push -m "Pre-migration work"

# Create and checkout feature branch
git checkout -b feat/monorepo-migration

# Verify branch
git branch --show-current
```

**Acceptance Criteria**:
- On feat/monorepo-migration branch
- Working directory clean
- Main branch unchanged

---

### Step 3: Document Current State
**Duration**: 30 minutes  

**Create documentation file**:
```bash
# Create reports directory
mkdir -p plans/reports

# Document current state
cat > plans/reports/pre-migration-state.md << 'EOF'
# Pre-Migration State Documentation

**Date**: 2026-06-27
**Branch**: feat/monorepo-migration
**Repository**: golang-sample

## Current Module Information

**Module Path**: golang-sample
**Go Version**: 1.25.0

## Current Dependencies

[Extracted from go.mod]

## Current Import Paths

- github.com/haipham22/govern/http
- github.com/haipham22/govern/config
- [Full list from codebase]

## Current Directory Structure

[Current structure]

## Git Status

- Current branch: feat/monorepo-migration
- Main branch: main
- Uncommitted changes: None

## Govern Repository State

- Location: ../govern/
- Branch: [main/develop]
- Latest commit: [SHA]
- Packages: [List of packages]
EOF
```

**Document dependencies**:
```bash
# List current imports
grep -r "github.com/haipham22/govern" internal/ cmd/ --include="*.go" > plans/reports/current-imports.txt

# List current dependencies
grep "^require" go.mod >> plans/reports/pre-migration-state.md
```

**Acceptance Criteria**:
- Complete documentation of current state
- All import paths documented
- All dependencies documented
- Govern repository state documented

---

### Step 4: Create Directory Structure
**Duration**: 15 minutes  
**Command**:
```bash
# Create directories
mkdir -p examples/golang-sample
mkdir -p templates/base
mkdir -p templates/basic
mkdir -p templates/fullstack
mkdir -p templates/microservice
mkdir -p scripts/generate-project

# Verify directories
ls -la examples/
ls -la templates/
ls -la scripts/generate-project/

# Create .gitkeep files
touch golang-sample/.gitkeep
touch templates/base/.gitkeep
touch templates/basic/.gitkeep
touch templates/fullstack/.gitkeep
touch templates/microservice/.gitkeep
touch scripts/generate-project/.gitkeep
```

**Acceptance Criteria**:
- golang-sample/ directory exists
- templates/ subdirectories exist
- scripts/generate-project/ directory exists
- .gitkeep files preserve directory structure in git

---

### Step 5: Update .gitignore
**Duration**: 10 minutes  

**Add ignores for new directories**:
```bash
# Add to .gitignore
cat >> .gitignore << 'EOF'

# Monorepo structure
examples/*/bin/
examples/*/tmp/
examples/*/.env
templates/*/bin/
scripts/generate-project/bin/
scripts/generate-project/tmp/

# Generated projects
generated-*/
EOF
```

**Acceptance Criteria**:
- .gitignore updated with new patterns
- No accidental commits of build artifacts
- Generated projects ignored

---

### Step 6: Update README (Feature Branch Only)
**Duration**: 15 minutes  

**Add migration notice**:
```bash
# Add migration notice to README
cat > README_MIGRATION_NOTICE.md << 'EOF'
# 🚧 Migration in Progress

This repository is being restructured into the govern monorepo.

**Current Branch**: feat/monorepo-migration
**Target**: Single repository containing govern library + sample app + generator

**Migration Status**: Phase 01 - Repository Preparation

**What's Happening**:
- Repository will be renamed from golang-sample to govern
- Govern packages will be merged from ../govern/ repository
- Sample app will move to golang-sample/
- Interactive project generator will be added

**No Breaking Changes**: Current imports already use github.com/haipham22/govern

**Migration Plan**: See plans/260627-2136-govern-monorepo-restructure/plan.md

---
EOF

# Prepend to README
cat README_MIGRATION_NOTICE.md README.md > README_TEMP.md
mv README_TEMP.md README.md
rm README_MIGRATION_NOTICE.md
```

**Acceptance Criteria**:
- README updated with migration notice
- Clear communication of ongoing work
- Link to migration plan

---

### Step 7: Verify Preparation
**Duration**: 15 minutes  
**Command**:
```bash
# Verify directory structure
tree -L 2 examples/ templates/ scripts/

# Verify git status
git status

# Verify no unintended changes
git diff main

# Verify documentation exists
ls -la plans/reports/pre-migration-state.md
ls -la plans/reports/current-imports.txt
```

**Acceptance Criteria**:
- Directory structure created correctly
- Git status shows only intended changes
- Documentation files exist
- No unintended modifications

---

### Step 8: Commit Preparation Work
**Duration**: 10 minutes  
**Command**:
```bash
# Stage all changes
git add examples/ templates/ scripts/ .gitignore README.md plans/reports/

# Review changes
git diff --cached

# Commit with conventional commit
git commit -m "feat: prepare repository for monorepo migration

- Create directory structure for examples/, templates/, generator
- Document current state and dependencies
- Update .gitignore for monorepo structure
- Add migration notice to README

Next: Phase 02 - Merge govern packages with git history
"

# Verify commit
git log -1 --stat
```

**Acceptance Criteria**:
- Clean commit with preparation work
- Conventional commit message
- No breaking changes
- Ready for Phase 02

---

## Success Criteria

### Phase Completion Criteria
- [x] Feature branch feat/monorepo-migration created
- [x] Directory structure prepared (examples/, templates/, scripts/generate-project/)
- [x] Current state documented (imports, dependencies, govern repo state)
- [x] .gitignore updated for monorepo structure
- [x] README updated with migration notice
- [x] Preparation work committed to feature branch
- [x] Main branch unchanged

### Quality Criteria
- [x] Documentation accurate and complete
- [x] Git history clean (single commit for preparation)
- [x] No unintended changes to codebase
- [x] All verification steps pass

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Govern repository not accessible | Low | High (blocks Phase 02) | Verify in Step 1 |
| Main branch accidentally modified | Low | Medium | Use feature branch, verify with git diff |
| Directory structure conflicts | Low | Low | Use gitignore, verify with tree command |
| Documentation incomplete | Medium | Medium | Follow checklist, verify all sections |

---

## Security Considerations

**No Security Impact**: This phase only creates directories and documentation

**Validation**:
- No credentials or secrets involved
- No code changes
- No dependencies modified

---

## Next Steps

**Dependencies**:
- Phase 02 must wait for Phase 01 completion
- Govern repository must be accessible

**Follow-up Tasks**:
- Phase 02: Merge Govern Packages (requires directory structure from this phase)
- Phase 03: Move Sample Application (requires govern packages merged)

---

## Rollback Strategy

**If preparation fails**:
```bash
# Delete feature branch
git checkout main
git branch -D feat/monorepo-migration

# Main branch unchanged, no impact
```

**If documentation incorrect**:
- Amend commit before pushing: `git commit --amend`
- Force push to feature branch only (safe for feature branch)

---

## Todo List

- [x] Verify Go 1.25+ installed
- [x] Verify ../govern/ repository exists
- [x] Create feat/monorepo-migration branch
- [x] Document current go.mod dependencies
- [x] Document current import paths
- [x] Document govern repository state
- [x] Create examples/ directory structure
- [x] Create templates/ directory structure
- [x] Create scripts/generate-project/ directory
- [x] Update .gitignore for monorepo
- [x] Update README with migration notice
- [x] Verify all changes with git diff
- [x] Commit preparation work
- [x] Verify commit in git log

---

## Phase Summary

**Input**: Main branch repository  
**Output**: Feature branch with prepared directory structure and documentation  
**Duration**: 2 hours  
**Risk Level**: Low  
**Blocks**: Phase 02 (Merge Govern Packages)

**Status**: Ready to start  
**Next Action**: Execute Step 1 (Verify Prerequisites)
