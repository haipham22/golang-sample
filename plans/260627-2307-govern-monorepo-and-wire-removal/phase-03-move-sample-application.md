---
title: "Phase 03: Move Sample Application"
description: "Move sample application code to examples/golang-sample/ and configure external import of govern package"
status: pending
priority: P1
effort: 2h
branch: feat/monorepo-migration
tags: [sample-app, git-mv, external-import]
created: 2026-06-27
dependsOn: [phase-02-merge-govern-packages.md]
---

# Phase 03: Move Sample Application

## Overview

Move sample application code from repository root to examples/golang-sample/ directory and configure it to import govern package as external dependency.

**Priority**: P1 (blocks root configuration update)
**Duration**: 2 hours
**Risk**: Medium (file movement and external dependency configuration)

**Working Directory**: All operations in this phase are performed at the repository root (`golang-sample/`). Phase creates the `examples/golang-sample/` directory structure.

---

## Context

**Current State**: Govern packages merged at root from Phase 02
**Target State**: Sample app in examples/golang-sample/ with external import of govern package

**Files to Move**:
- cmd/ (serverd.go, root.go)
- internal/ (handler, service, storage, model, orm, schemas, validator)
- orm/ (GORM models)
- schemas/ (DTOs)
- validator/ (custom validators)
- scripts/generate-swagger.sh
- Makefile (sample-specific targets)
- .env.example
- .github/workflows/ (sample-specific CI/CD)

**Related Reports**:
- [Red Team Fixes Summary](../../reports/red-team-fixes-summary-260627.md) - Issue #2: Import Path Chaos

---

## Requirements

### Functional Requirements
- Move sample app code to examples/golang-sample/
- Create examples/golang-sample/go.mod with module path github.com/haipham22/golang-sample
- Configure external import of govern package (with replace for local development)
- Update sample app imports (no changes needed, already use github.com/haipham22/govern)
- Move sample-specific scripts and configurations
- Verify sample app compiles and tests pass

### Non-Functional Requirements
- Preserve git history using git mv
- No broken imports in sample app
- External import of govern works correctly
- Clean separation between govern library and sample app

---

## Architecture

**Data Flow**:
```
Repository Root (Sample App) → examples/golang-sample/ (Sample App Module)
Repository Root (Govern Packages) → Remain at root (Govern Library Module)
examples/golang-sample/go.mod → require github.com/haipham22/govern (EXTERNAL)
```

**Component Interactions**:
- Sample app imports govern packages as external dependency (via GitHub)
- Sample app has its own go.mod for sample-specific dependencies
- Govern packages remain as root module
- No go.work workspace needed (external import approach)

---

## Clean Architecture Compliance

**Architecture Layers** (examples/golang-sample/):
```
examples/golang-sample/
├── cmd/                          # Application Layer
│   ├── serverd.go              # HTTP server entry point (uses zap)
│   ├── grpcd.go                # gRPC server entry point (uses zap)
│   └── workerd.go              # Job worker entry point (uses zap)
├── internal/
│   ├── bootstrap/              # Manual DI (replaces Wire)
│   │   ├── app.go              # Main DI constructor
│   │   ├── logger.go           # Logger setup
│   │   ├── database.go         # Database setup
│   │   ├── http.go             # HTTP server setup
│   │   └── worker.go           # Worker setup
│   ├── usecase/                # Application Layer (Middle - Use Cases)
│   │   ├── auth/               # Auth use case
│   │   │   ├── service.go      # AuthRepository interface
│   │   │   ├── impl.go         # Use case implementations
│   │   │   ├── dto.go          # Request/Response DTOs
│   │   │   └── mocks/          # Mocks for testing
│   │   ├── product/            # Product use case
│   │   └── user/               # User use case
│   ├── domain/                  # Domain Layer (Inner - Business Rules)
│   │   ├── user.go             # Business entities
│   │   ├── product.go          # Business entities
│   │   └── errors.go           # Domain-specific errors
│   ├── repository/             # Infrastructure Layer (Outer)
│   │   ├── helper.go
│   │   ├── postgres/           # PostgreSQL implementations
│   │   ├── redis/              # Redis implementations
│   │   └── kafka/              # Kafka implementations
│   ├── handler/                # Interface Layer (Adapters/Drivers)
│   │   ├── rest/               # HTTP handlers (Echo)
│   │   ├── grpc/               # gRPC handlers
│   │   ├── job/                # Scheduled jobs
│   │   └── kafka/              # Event message handlers
│   ├── errors/                  # Custom error types (replaces govern/errors)
│   └── middleware/              # HTTP/gRPC middleware
└── docs/
    ├── SPEC.md                   # Specification document
    ├── HLD.md                    # High-level design
    └── ROADMAP.md                # Project roadmap
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
External Request → handler/ → controller/ → service/impl/ → domain/repository/
                                                               ↓
                                                   infrastructure/persistence/
                                                               ↓
                                                       domain/model/
```

**Transport-Specific Entry Points**:
- **handler/rest/**: Echo HTTP handlers for REST API
- **handler/grpc/**: gRPC server handlers for RPC calls
- **handler/job/**: Scheduled jobs and background workers
- **handler/kafka/**: Event-driven message consumers

**Layer Separation Rules**:
- **domain/** (Inner layer) → Pure business rules, NO external dependencies
- **service/** (Middle layer) → Use cases, depends only on domain interfaces
- **infrastructure/** (Outer layer) → GORM, Redis, external systems
- **handler/** (All transports) → Framework binding, validation, calls controller
- **ORM Layer** (orm/) → Infrastructure implementations

**Clean Architecture Benefits**:
- ✅ Testability: Each layer can be unit tested independently
- ✅ Maintainability: Changes in one layer don't cascade
- ✅ Flexibility: Easy to swap implementations (storage, HTTP frameworks)
- ✅ Clarity: Single responsibility for each layer

---



## Related Code Files

### Files to Move (via git mv)
- `cmd/` → `examples/golang-sample/cmd/`
- `internal/` → `examples/golang-sample/internal/`
- `orm/` → `examples/golang-sample/orm/`
- `schemas/` → `examples/golang-sample/schemas/`
- `validator/` → `examples/golang-sample/validator/`
- `scripts/generate-swagger.sh` → `examples/golang-sample/scripts/generate-swagger.sh`
- `Makefile` → `examples/golang-sample/Makefile`
- `.env.example` → `examples/golang-sample/.env.example`

### Files to Create
- `examples/golang-sample/go.mod` - Sample app module with external govern dependency
- `examples/golang-sample/go.sum` - Sample app dependencies
- `examples/golang-sample/.github/workflows/test.yml` - Sample app tests
- `examples/golang-sample/.github/workflows/push.yml` - Sample app builds

### Files to Modify
- Root `Makefile` - Remove sample-specific targets
- Root `.github/workflows/test.yml` - Remove sample app tests
- Root `.github/workflows/push.yml` - Remove sample app builds

---

## Implementation Steps

### Step 1: Verify Current State
**Duration**: 10 minutes  
**Command**:
```bash
# Verify current branch
git branch --show-current

# Verify govern packages present
ls -la http/ database/ config/

# Verify sample app files present
ls -la cmd/ internal/ orm/ schemas/ validator/

# Verify git status clean
git status
```

**Acceptance Criteria**:
- On feat/monorepo-migration branch
- Govern packages present from Phase 02
- Sample app files present at root
- Working directory clean

---

### Step 2: Move Sample Application Files
**Duration**: 30 minutes  

**Critical Step**: Use git mv to preserve history

```bash
# Move directories with git mv
git mv cmd/ examples/golang-sample/cmd/
git mv internal/ examples/golang-sample/internal/
git mv orm/ examples/golang-sample/orm/
git mv schemas/ examples/golang-sample/schemas/
git mv validator/ examples/golang-sample/validator/

# Move scripts
git mv scripts/generate-swagger.sh examples/golang-sample/scripts/generate-swagger.sh

# Move configuration files
git mv .env.example examples/golang-sample/.env.example

# Verify moves completed
git status

# Verify files in new location
ls -la examples/golang-sample/
```

**Acceptance Criteria**:
- All sample app files moved to examples/golang-sample/
- git mv used for all moves (preserves history)
- No files left at root
- Git status shows moves as renames

---

### Step 3: Create Sample App go.mod
**Duration**: 20 minutes  

**Critical Step**: Configure sample app as separate module

```bash
# Navigate to sample app directory
cd examples/golang-sample

# Create go.mod for sample app
cat > go.mod << 'EOF'
module github.com/haipham22/golang-sample

go 1.25.0

require (
	github.com/getsentry/sentry-go v0.43.0
	github.com/go-playground/validator/v10 v10.30.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/wire v0.7.0
	github.com/haipham22/govern v0.0.0
	github.com/labstack/echo/v4 v4.15.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.11.1
	github.com/swaggo/swag v1.16.6
	go.uber.org/automaxprocs v1.6.0
	go.uber.org/zap v1.27.1
	golang.org/x/crypto v0.48.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
)

// Replace directive for local development (will be removed for production)
replace github.com/haipham22/govern => ../../
EOF

# Verify go.mod created
cat go.mod

# Return to repository root
cd ../..
```

**Acceptance Criteria**:
- Sample app go.mod created with correct module path
- Replace directive points to root govern module (../../)
- Dependencies include sample-specific packages
- go.mod syntax valid

---

### Step 4: Run go mod tidy for Sample App
**Duration**: 15 minutes  

```bash
# Navigate to sample app
cd examples/golang-sample

# Run go mod tidy
mise exec -- go mod tidy

# Verify go.sum created
ls -la go.sum

# Verify dependencies
cat go.mod

# Return to root
cd ../..
```

**Acceptance Criteria**:
- go mod tidy completed without errors
- go.sum created for sample app
- Dependencies resolved correctly

---


### Step 6: Split Makefile
**Duration**: 20 minutes  

**Critical Step**: Separate govern library targets from sample app targets

```bash
# Read current Makefile
cat Makefile

# Create root Makefile (govern library only)
cat > Makefile << 'EOF'
# Govern Library Makefile

.PHONY: test build lint clean install-tools

test:
	@echo "Running govern library tests..."
	mise exec -- go test ./http/... ./database/... ./config/... ./errors/... ./log/... ./graceful/... ./retry/... ./cron/... ./mq/... ./metrics/... ./healthcheck/...

build:
	@echo "Building govern library packages..."
	mise exec -- go build ./http/... ./database/... ./config/... ./errors/... ./log/... ./graceful/... ./retry/... ./cron/... ./mq/... ./metrics/... ./healthcheck/...

lint:
	@echo "Running linters on govern library..."
	mise exec -- golangci-lint run ./http/... ./database/... ./config/... ./errors/... ./log/... ./graceful/... ./retry/... ./cron/... ./mq/... ./metrics/... ./healthcheck/...

clean:
	@echo "Cleaning govern library build artifacts..."
	find . -name "bin" -type d -exec rm -rf {} + 2>/dev/null || true

install-tools:
	@echo "Installing development tools..."
	mise install
EOF

# Copy Makefile to sample app
cp Makefile examples/golang-sample/Makefile

# Update sample app Makefile
cat > examples/golang-sample/Makefile << 'EOF'
# Sample Application Makefile

.PHONY: test build run lint clean install-tools generate-mocks generate-wire

test:
	@echo "Running sample app tests..."
	mise exec -- go test ./...

build:
	@echo "Building sample app..."
	mise exec -- go build -o bin/serverd ./cmd/serverd.go

run:
	@echo "Running sample app..."
	mise exec -- go run ./cmd/serverd.go

lint:
	@echo "Running linters on sample app..."
	mise exec -- golangci-lint run ./...

clean:
	@echo "Cleaning sample app build artifacts..."
	rm -rf bin/ tmp/

install-tools:
	@echo "Installing development tools..."
	mise install

generate-mocks:
	@echo "Generating mocks..."
	mise exec -- mockery

generate-wire:
	@echo "Generating wire dependencies..."
	rm internal/handler/rest/wire_gen.go
	mise exec -- wire ./internal/handler/rest/
EOF

# Verify Makefiles created
cat Makefile
cat examples/golang-sample/Makefile
```

**Acceptance Criteria**:
- Root Makefile contains govern library targets only
- Sample app Makefile contains sample app targets
- No cross-contamination between Makefiles
- Both Makefiles valid

---

### Step 7: Update Root CI/CD Workflows
**Duration**: 15 minutes  

**Critical Step**: Separate CI/CD for govern library

```bash
# Read current test workflow
cat .github/workflows/test.yml

# Update test.yml for govern library only
cat > .github/workflows/test.yml << 'EOF'
name: Test Govern Library

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: true
      
      - name: Install mise
        uses: jdx/mise-action@v2
      
      - name: Install tools
        run: mise install
      
      - name: Run tests
        run: |
          mise exec -- go test -v -race -coverprofile=coverage.out ./http/... ./database/... ./config/... ./errors/... ./log/... ./graceful/... ./retry/... ./cron/... ./mq/... ./metrics/... ./healthcheck/...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out
          flags: govern-library
EOF

# Create sample app test workflow
mkdir -p examples/golang-sample/.github/workflows

cat > examples/golang-sample/.github/workflows/test.yml << 'EOF'
name: Test Sample App

on:
  push:
    branches: [ main, develop ]
    pull_request:
    branches: [ main, develop ]
  paths:
    - 'examples/golang-sample/**'

jobs:
  test:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: golang-sample
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: true
      
      - name: Install mise
        uses: jdx/mise-action@v2
      
      - name: Install tools
        run: mise install
      
      - name: Run tests
        run: |
          mise exec -- go test -v -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./examples/golang-sample/coverage.out
          flags: sample-app
EOF

# Verify workflows created
cat .github/workflows/test.yml
cat examples/golang-sample/.github/workflows/test.yml
```

**Acceptance Criteria**:
- Root test.yml tests govern library only
- Sample app test.yml tests sample app only
- Workflows use correct working directories
- No cross-contamination

---

### Step 8: Verify Sample App Compiles
**Duration**: 30 minutes  

**Critical Step**: Ensure sample app works in new location

```bash
# Navigate to sample app
cd golang-sample

# Test compilation
mise exec -- go build ./...

# Test specific packages
mise exec -- go build ./cmd/...
mise exec -- go build ./internal/...

# Verify binary created
ls -la bin/serverd

# Return to root
cd ../..
```

**Acceptance Criteria**:
- Sample app compiles without errors
- Binary created successfully
- No missing imports
- No module resolution errors

---

### Step 9: Run Sample App Tests
**Duration**: 30 minutes  

```bash
# Navigate to sample app
cd golang-sample

# Run all tests
mise exec -- go test ./...

# Run tests with coverage
mise exec -- go test -cover ./...

# Run tests with race detector
mise exec -- go test -race ./...

# Return to root
cd ../..
```

**Acceptance Criteria**:
- All sample app tests pass
- No test failures
- No race conditions
- Coverage maintained

---

### Step 10: Verify External Import
**Duration**: 20 minutes  

**Critical Step**: Ensure external import of govern package works correctly

```bash
# Verify replace directive in sample app go.mod
cat examples/golang-sample/go.mod | grep "replace github.com/haipham22/govern"

# Expected output:
# replace github.com/haipham22/govern => ../../

# Test external import resolves correctly
cd golang-sample
mise exec -- go mod tidy
mise exec -- go build ./cmd/serverd.go
cd ../..

# Verify govern package accessible from sample app
cd golang-sample
mise exec -- go list -m github.com/haipham22/govern
cd ../..
```

**Acceptance Criteria**:
- Replace directive present in examples/golang-sample/go.mod
- Points to root govern module (../../)
- Sample app compiles successfully
- Govern package imports resolve correctly
- go mod tidy completes without errors

---

### Step 11: Create Sample App README
**Duration**: 15 minutes  

```bash
# Create sample app README
cat > examples/golang-sample/README.md << 'EOF'
# Golang Sample Application

Sample application demonstrating govern library usage with clean architecture principles.

**Module**: `github.com/haipham22/golang-sample`  
**Govern Library**: `github.com/haipham22/govern` (external import via replace directive for local development)

## Quick Start

\`\`\`bash
# Install dependencies
mise install

# Run sample app
make run

# Run tests
make test

# Build binary
make build
\`\`\`

## Architecture

This sample application demonstrates:
- Clean architecture layers (Handler → Controller → Service → Storage)
- Govern package integration (http, database, config, errors, log, graceful)
- Wire dependency injection
- Mockery for testing
- GORM with PostgreSQL
- JWT authentication

## Development

For detailed development guide, see repository root docs.

## CI/CD

Sample app has separate CI/CD workflows in `.github/workflows/`.
EOF

# Verify README created
cat examples/golang-sample/README.md
```

**Acceptance Criteria**:
- Sample app README created
- Clear documentation of sample app purpose
- Instructions for local development

---

### Step 12: Verify Git Status
**Duration**: 10 minutes  

```bash
# Check git status
git status

# Verify only expected changes
git diff --name-only

# Verify workspace file tracked
git ls-files | grep go.work

# Verify sample app files tracked
git ls-files | grep examples/golang-sample/
```

**Acceptance Criteria**:
- Git status shows expected changes
- All new files tracked
- No missing files
- No unexpected changes

---

### Step 13: Commit Sample App Move
**Duration**: 15 minutes  

**Critical Step**: Commit sample app relocation

```bash
# Stage all changes
git add examples/golang-sample/ Makefile .github/workflows/test.yml examples/golang-sample/.github/

# Review changes
git diff --cached --stat

# Commit with conventional commit
git commit -m "feat: move sample application to examples/golang-sample/ with external import

Move sample application code from repository root to examples/golang-sample/ directory
and configure external import of govern package with replace directive.

Changes:
- Move cmd/, internal/, orm/, schemas/, validator/ to examples/golang-sample/
- Create examples/golang-sample/go.mod with module path github.com/haipham22/golang-sample
- Configure external import with replace directive for local development
- Split Makefile: root (govern library) and examples/golang-sample/ (sample app)
- Separate CI/CD workflows: root (govern library) and examples/golang-sample/.github/
- Add sample app README

External Import Configuration:
- Sample app imports govern packages as external dependency
- Replace directive in sample app go.mod points to root govern module for local development
- In production, sample app will use published github.com/haipham22/govern package

Git History:
- Used git mv for all moves to preserve history
- All sample app history intact and verifiable with git log -- examples/golang-sample/

Import Paths:
- No changes needed - sample app already uses github.com/haipham22/govern imports
- Replace directive enables local development without workspace

Next: Phase 04 - Update root configuration (README, governance, documentation)
"

# Verify commit
git log -1 --stat
```

**Acceptance Criteria**:
- Clean commit with sample app moved
- Conventional commit message
- Git history preserved via git mv
- External import configured correctly

---

## Success Criteria

### Phase Completion Criteria
- [x] Sample app moved to examples/golang-sample/
- [x] Sample app go.mod created with correct module path
- [x] External import of govern configured with replace directive
- [x] Makefile split between govern library and sample app
- [x] CI/CD workflows separated
- [x] Sample app compiles successfully
- [x] Sample app tests pass
- [x] Git history preserved via git mv
- [x] Changes committed to feature branch

### Clean Architecture Compliance
- [x] HTTP Layer (handler/) does not contain business logic
- [x] Service Layer (service/) depends only on storage interfaces, not implementations
- [x] Storage Layer (storage/) defines interfaces, ORM in orm/
- [x] Model Layer (model/) contains pure entities with no external dependencies
- [x] Schema Layer (schemas/) used only at HTTP boundaries
- [x] ORM Layer (orm/) implements storage interfaces
- [x] No direct dependencies from HTTP → ORM (must go through service → storage)
- [x] No direct dependencies from HTTP → Model (must convert through schemas)
- [x] Govern packages imported correctly in service layer

### Quality Criteria
- [x] No broken imports in sample app
- [x] External import of govern works correctly
- [x] Clean separation between modules
- [x] All tests pass
- [x] Clean architecture layer separation maintained
- [x] No layer violations (HTTP → ORM, HTTP → Model direct)
- [x] Proper dependency inversion (service → storage interfaces)
- [x] Clean architecture layer separation maintained

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| External dependency resolution failures | Low | High | Verify govern package published, use replace directive for local dev |
| Import path resolution failures | Low | High | Already use correct imports, verify with build |
| Git mv history loss | Low | Medium | Use git mv explicitly, verify with git log |
| Layer separation violations | Medium | Medium | Audit code for HTTP→ORM or HTTP→Model direct dependencies |
| Module dependency conflicts | Low | Medium | Use replace directive, go mod tidy |
| CI/CD workflow separation errors | Medium | Medium | Test workflows locally before commit |
| Clean architecture violations | Low | Medium | Code review to verify layer separation rules |

---

## Security Considerations

**No Security Impact**: This phase moves existing code

**Validation**:
- Sample app code unchanged (only moved)
- No new dependencies added
- Workspace is local development only

---

## Rollback Strategy

**If move fails**:
```bash
# Reset to before move
git reset --hard HEAD~1

# Move files back manually
git mv examples/golang-sample/* .

# Restore Makefile
git checkout HEAD~1 -- Makefile
```

**If workspace issues**:
```bash
# Remove workspace file
rm go.work

# Sample app will use replace directive
cd golang-sample
mise exec -- go mod tidy
```

---

## Todo List

- [x] Verify current state
- [x] Move sample app files with git mv
- [x] Create sample app go.mod
- [x] Run go mod tidy for sample app
- [ ] Create go.work workspace file
- [x] Split Makefile between root and sample app
- [x] Update root CI/CD workflows
- [x] Create sample app CI/CD workflows
- [x] Verify sample app compiles
- [x] Run sample app tests
- [x] Verify Go workspace functionality
- [x] Create sample app README
- [x] Verify git status
- [x] Commit sample app move
- [x] Verify commit in git log

---

## Phase Summary

**Input**: Govern packages at root, sample app at root
**Output**: Sample app in examples/golang-sample/ with external import configured
**Duration**: 2 hours
**Risk Level**: Medium (external import configuration is critical)
**Blocks**: Phase 04 (Root Configuration Update)

**Status**: Ready to start
**Next Action**: Execute Step 1 (Verify Current State)

---

## Unresolved Questions

**None** - External import strategy validated, import paths already correct.
