---
title: "Phase 10: Centralized Error Management System"
description: "Replace govern/errors usage and implement centralized error management"
status: pending
priority: P1
effort: 3h
dependsOn: [phase-09-custom-error-types.md]
---

## Overview

**Priority**: P1 | **Status**: pending | **Effort**: 3h

Replace all govern/errors usage with custom error types and create centralized error management system with standardized error creation, handling patterns, and observability.

**Working Directory**: All operations in this phase are performed in the `examples/golang-sample/` directory. All file paths are relative to `examples/golang-sample/`.

## Context Links

- **Parent Plan**: [plan.md](./plan.md)
- **Previous Phase**: [phase-09-custom-error-types.md](./phase-09-custom-error-types.md)
- **Next Phase**: [phase-11-manual-di-implementation.md](./phase-11-manual-di-implementation.md)
- **Related Files**: `examples/golang-sample/internal/handler/rest/handler.go`, `examples/golang-sample/internal/handler/rest/controllers/auth/auth.go`

## Key Insights

**Current govern/errors Usage**:
- 2 main files: handler.go (error handler), auth.go (controller)
- 4 usages in handler.go (GetCode, error code checks)
- 2 usages in auth.go (WrapCode for validation errors)
- Test file: 1 usage (GetCode for assertions)

**Centralized Management Benefits**:
- Consistent error creation patterns
- Standardized error response format
- Central logging and observability
- Easier maintenance and testing
- Single place for error handling rules

## Requirements

### Functional Requirements
1. Replace all govern/errors imports with custom errors
2. Create centralized error creation helpers
3. Standardize error response format
4. Add structured logging for errors
5. Document error handling patterns

### Non-Functional Requirements
- Zero breaking changes to API responses
- Maintain current error logging behavior
- Add request ID tracking to errors
- No performance regression

## Architecture

**Error Management Structure**:
```
internal/errors/
├── errors.go              # Core error types (from Phase 02)
├── codes.go               # Error code definitions (from Phase 02)
├── wrap.go                # Error wrapping functions (from Phase 02)
├── helpers.go             # NEW: Centralized error creation helpers
├── response.go            # NEW: Standardized error response format
├── logging.go             # NEW: Error logging with observability
└── errors_test.go         # Unit tests
```

**Error Flow (After Refactoring)**:
```
Service/Storage → errors.WrapCode() → Controller → customHTTPErrorHandler → HTTP Response
                                                        ↓
                                                 Error Logging
                                                 Request ID Tracking
```

**Error Response Format**:
```json
{
  "msg": "Error message",
  "error": "error_type",
  "path": "/api/endpoint",
  "request_id": "uuid"
}
```

## Related Code Files

### Files to Modify

**Handler Files** (all transport types):
- `internal/handler/rest/handler.go` - Replace govern/errors with custom (HTTP error handler)
- `internal/handler/rest/controllers/auth/auth.go` - Replace govern/errors imports
- `internal/handler/rest/controllers/auth/auth_test.go` - Update test assertions
- `internal/handler/grpc/*.go` - Replace govern/errors imports (future: gRPC handlers)
- `internal/handler/job/*.go` - Replace govern/errors imports (future: job handlers)
- `internal/handler/kafka/*.go` - Replace govern/errors imports (future: Kafka handlers)

**Service Files:**
- `internal/usecase/auth/impl.go` - Replace govern/errors with custom (9 usages)
- `internal/usecase/auth/dto.go` - Replace govern/errors imports (validation)

**Files to Create**
- `internal/errors/errors.go` - Core error types + envelope types
- `internal/errors/codes.go` - Error code definitions
- `internal/errors/wrap.go` - Error wrapping functions with envelope support
- `internal/errors/envelope/db_error.go` - Database error envelope
- `internal/errors/envelope/config_error.go` - Config error envelope
- `internal/errors/envelope/logger_error.go` - Logger error envelope
- `internal/errors/envelope/http_error.go` - HTTP error envelope
- `internal/errors/helpers.go` - Centralized error creation helpers
- `internal/errors/response.go` - Standardized error response format
- `internal/errors/errors_test.go` - Unit tests

### Files to Delete
- None (govern/errors import removed but package remains in go.mod until Phase 04)

## Implementation Steps

1. **Create Error Creation Helpers** (45m)
   ```go
   // internal/errors/helpers.go
   func InvalidInput(field, message string) *AppError
   func NotFound(resource string) *AppError
   func Unauthorized(message string) *AppError
   func Forbidden(message string) *AppError
   func Conflict(resource string) *AppError
   func Internal(message string) *AppError
   ```

2. **Create Standardized Error Response** (45m)
   ```go
   // internal/errors/response.go
   type ErrorResponse struct {
       Msg       string `json:"msg"`
       Error     string `json:"error"`
       Path      string `json:"path,omitempty"`
       RequestID string `json:"request_id,omitempty"`
   }

   func BuildErrorResponse(err error, path, requestID string) ErrorResponse
   ```

3. **Replace govern/errors in handler.go** (30m)
   ```go
   // Replace: governerrors "github.com/haipham22/govern/errors"
   // With:    apperrors "github.com/haipham22/golang-sample/internal/errors"

   // Replace: governerrors.GetCode(err)
   // With:    apperrors.GetCode(err)

   // Replace: governerrors.CodeInvalid
   // With:    apperrors.ErrCodeInvalid
   ```

4. **Replace govern/errors in auth.go** (15m)
   ```go
   // Replace: governerrors.WrapCode(governerrors.CodeInvalid, err)
   // With:    apperrors.WrapCode(apperrors.ErrCodeInvalid, err)
   ```

5. **Add Error Logging and Observability** (30m)
   ```go
   // internal/errors/logging.go
   func LogError(err error, path, requestID string)
   func LogWarning(err error, path, requestID string)
   ```

6. **Update Tests** (15m)
   ```go
   // Replace governerrors imports
   // Update error code assertions
   // Add request ID tests
   ```

## Todo List

- [x] Create centralized error creation helpers (helpers.go)
- [x] Create standardized error response format (response.go)
- [x] Implement error logging with observability (logging.go)
- [x] Replace govern/errors imports in handler.go
- [x] Replace govern/errors imports in auth.go
- [x] Replace govern/errors imports in auth_test.go
- [x] Update error handler to use new response format
- [x] Add request ID tracking to errors
- [x] Update all test assertions
- [x] Run tests and validate no breaking changes
- [x] Verify error logging works correctly
- [ ] Document error handling patterns (docs/error-handling.md)

## Success Criteria

**Definition of Done**:
- All govern/errors imports replaced with custom errors
- Centralized error helpers implemented and used
- Standardized error response format applied
- Error logging with request ID tracking working
- All tests passing with same coverage
- No breaking changes to API responses
- Error handling patterns documented

**Validation Methods**:
```bash
# Run tests
go test ./... -v

# Check for remaining govern/errors imports
grep -r "governerrors" internal/

# Test error responses match (use curl or integration test)
curl -X POST http://localhost:8080/api/login -d '{"invalid":"data"}'

# Verify error logging
# Check logs for request ID and error details
```

**API Compatibility Check**:
```bash
# Before and after must produce same responses
# Compare HTTP status codes, response structure
# Verify error messages identical
```

## Risk Assessment

**Potential Issues**:
1. **Breaking Error Responses**: Response format may change
   - Mitigation: Unit tests verify exact response structure
2. **Missing Error Context**: Custom errors may lose information
   - Mitigation: Comprehensive test coverage before/after
3. **Logging Performance**: Extra logging may slow requests
   - Mitigation: Use structured logging (Zap), measure performance

**Medium Risk**: This changes core error handling - must be thorough

**Rollback**: Git revert if breaking changes detected

## Security Considerations

- Error messages must not leak internal details
- 5xx errors use generic messages only
- Request IDs don't contain sensitive data
- Error logs sanitized (no passwords/credentials)

## Next Steps

**Dependencies**: Phase 02 must be complete (custom errors available)

**Follow-up Tasks**:
- Phase 04: Manual DI implementation (requires centralized errors)
- Phase 05: Simplify error handler logic (builds on this phase)

**Transition Criteria**:
- All tests passing → Start Phase 04
- No govern/errors imports remaining → Ready to remove from go.mod
- Error responses validated → Safe to proceed
