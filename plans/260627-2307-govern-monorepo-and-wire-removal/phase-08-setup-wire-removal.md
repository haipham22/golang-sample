---
title: "Phase 08: Setup Wire Removal Environment"
description: "Setup validation environment, establish test baseline, and document current state for wire removal"
status: pending
priority: P1
effort: 2h
dependsOn: [phase-06-validation-testing.md]
---

## Overview

**Priority**: P1 | **Status**: pending | **Effort**: 2h

Establish validation environment, document current Wire dependency graph, establish test coverage baseline, and prepare workspace for wire removal refactoring.

**Working Directory**: All operations in this phase are performed in the `examples/golang-sample/` directory (created by Phase 03). All file paths are relative to `examples/golang-sample/`.

**Prerequisite**: Phase 06 must be complete (monorepo restructuring validated and tested).

## Context Links

- **Parent Plan**: [plan.md](./plan.md)
- **Previous Phase**: [phase-06-validation-testing.md](./phase-06-validation-testing.md)
- **Next Phase**: [phase-09-custom-error-types.md](./phase-09-custom-error-types.md)
- **Related Files**: `examples/golang-sample/go.mod`, `examples/golang-sample/mise.toml`, `examples/golang-sample/cmd/serverd.go`
- **Research**: Go error handling best practices, Wire dependency analysis

## Key Insights

**Current State Analysis**:
- Wire provides compile-time DI with 8 provider functions
- govern/errors used in 2 main files (handler.go, auth controller)
- Test coverage: 83.3% (needs to be maintained)
- Single entry point: `rest.New()` in `cmd/serverd.go`

**Critical Dependencies**:
- Wire generates `wire_gen.go` (must preserve logic during migration)
- govern/errors code mapping: CodeInvalid, CodeNotFound, CodeUnauthorized, CodeForbidden, CodeConflict, CodeInternal
- Error handler in handler.go (210 lines, needs simplification)

## Requirements

### Functional Requirements
1. Document current Wire dependency graph
2. Establish test coverage baseline
3. Create validation scripts
4. Backup current working state
5. Research Go error handling best practices

### Non-Functional Requirements
- No breaking changes to existing tests
- Maintain current code coverage (≥83%)
- Zero runtime errors during validation
- Clear rollback procedure

## Architecture

**Current Wire Dependency Graph**:
```
rest.New(log, port, appConfig)
├── provideAuthConfig(appConfig) → authConfig
├── provideDB(appConfig) → (*gorm.DB, cleanup, error)
├── userRepo.New(log, db) → Storage
├── provideAuthService(log, storage, cfg) → Service
├── authctrl.New(service) → Controller
├── healthctrl.New(db) → HealthController
├── provideDebugFlag(appConfig) → bool
├── provideEnv(appConfig) → string
└── NewHandler(log, echo, authCtrl, healthCtrl, port, debug, env) → Server
```

**Error Flow (Current)**:
```
Controller → governerrors.WrapCode() → customHTTPErrorHandler → HTTP response
```

## Related Code Files

### Files to Read
- `examples/golang-sample/internal/handler/rest/wire.go` - Wire providers
- `examples/golang-sample/internal/handler/rest/wire_gen.go` - Generated DI code
- `examples/golang-sample/internal/handler/rest/handler.go` - Error handler
- `examples/golang-sample/internal/handler/rest/controllers/auth/auth.go` - Error usage

### Files to Create
- `scripts/validate.sh` - Validation script
- `scripts/before-refactor-baseline.sh` - Test baseline script

### Files to Modify
- None in this phase

## Implementation Steps

1. **Document Wire Dependency Graph** (30m)
   ```bash
   # Analyze wire.go and wire_gen.go
   # Create dependency graph visualization
   # Document all provider functions
   # Note cleanup functions and lifecycle
   ```

2. **Establish Test Baseline** (30m)
   ```bash
   # Run full test suite
   go test ./... -v -coverprofile=baseline.out
   # Save coverage report
   go tool cover -html=baseline.out -o baseline-coverage.html
   # Document test count and pass rate
   ```

3. **Create Validation Scripts** (30m)
   ```bash
   # Create validate.sh script
   # - Check compilation
   # - Run all tests
   # - Check coverage threshold
   # - Validate no Wire imports in final state
   ```

4. **Research Go Error Handling** (30m)
   - Review Go 1.25+ error wrapping patterns
   - Research errors.Is/As best practices
   - Document error handling patterns for clean architecture
   - Create research report in `reports/research-errors.md`

## Todo List

- [x] Document Wire dependency graph with visualization
- [x] Run test suite and save baseline coverage report
- [x] Create validation script (scripts/validate.sh)
- [ ] Create baseline test script (scripts/before-refactor-baseline.sh)
- [x] Research Go error handling best practices
- [x] Document current error handling patterns in codebase
- [x] Create rollback branch: `backup/wire-implementation`
- [x] Validate environment compiles cleanly

## Success Criteria

**Definition of Done**:
- Wire dependency graph documented with visualization
- Test baseline established (coverage report saved)
- Validation script created and tested
- Research report completed
- All tests passing (baseline)
- No compilation errors
- Rollback branch created

**Validation Methods**:
```bash
# Run validation script
./scripts/validate.sh

# Expected output:
# ✓ Compilation successful
# ✓ All tests passing (83.3% coverage)
# ✓ Wire dependency graph documented
# ✓ Baseline established
```

## Risk Assessment

**Potential Issues**:
1. **Test Flakiness**: Baseline tests may be flaky
   - Mitigation: Run tests 3 times, take most stable result
2. **Wire Complexity**: Dependency graph more complex than expected
   - Mitigation: Extra time for documentation (30m buffer)

**Low Risk**: This phase is read-only (no code changes)

## Security Considerations

- No security impact (read-only phase)
- Test data may contain credentials (use .env.test)

## Next Steps

**Dependencies**: None (first phase)

**Follow-up Tasks**:
- Phase 02: Create custom error types
- Phase 03: Implement centralized error management
- Phase 04: Manual DI implementation

**Transition Criteria**:
- Baseline tests passing → Start Phase 02
- Validation script working → Use in all subsequent phases
