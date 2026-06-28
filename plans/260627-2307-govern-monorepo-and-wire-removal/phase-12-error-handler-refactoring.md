---
title: "Phase 12: Error Handler Refactoring"
description: "Simplify error handling logic and reduce complexity in customHTTPErrorHandler"
status: pending
priority: P1
effort: 2h
dependsOn: [phase-11-manual-di-implementation.md]
---

## Overview

**Priority**: P1 | **Status**: pending | **Effort**: 2h

Refactor and simplify the `customHTTPErrorHandler` function in handler.go to reduce complexity, improve maintainability, and leverage the new centralized error management system from Phase 11.

**Working Directory**: All operations in this phase are performed in the `examples/golang-sample/` directory. All file paths are relative to `examples/golang-sample/`.

## Context Links

- **Parent Plan**: [plan.md](./plan.md)
- **Previous Phase**: [phase-11-manual-di-implementation.md](./phase-11-manual-di-implementation.md)
- **Next Phase**: [phase-13-wire-removal-testing.md](./phase-13-wire-removal-testing.md)
- **Related Files**: `examples/golang-sample/internal/handler/rest/handler.go`

## Key Insights

**Current handler.go Issues**:
- 210 lines total (at 200-line limit)
- customHTTPErrorHandler is 135+ lines
- Nested if-else statements (4 levels deep)
- Repeated response building logic
- Govern error code checking duplicated

**Refactoring Opportunities**:
- Use centralized error helpers from Phase 03
- Extract response building to separate function
- Simplify error code checking with map lookup
- Reduce nesting with early returns
- Separate validation error handling

## Requirements

### Functional Requirements
1. Simplify customHTTPErrorHandler logic (reduce to <80 lines)
2. Extract response building to helper function
3. Use centralized error helpers
4. Maintain same HTTP response format
5. Preserve all error handling behavior

### Non-Functional Requirements
- Zero breaking changes to API responses
- Improved code readability
- Better maintainability
- File size under 200 lines

## Architecture

**Current Structure** (210 lines):
```
handler.go
├── NewHandler()              # 40 lines
├── customHTTPErrorHandler()  # 135 lines ← COMPLEX
└── buildValidationErrorResponse() # 35 lines
```

**Target Structure** (~150 lines):
```
handler.go
├── NewHandler()              # 40 lines
├── customHTTPErrorHandler()  # 50 lines ← SIMPLIFIED
├── buildErrorResponse()      # 30 lines (extracted)
└── handleValidationError()   # 30 lines (extracted)
```

**Error Response Flow** (Simplified):
```
Error → Get Code → Build Response → Log → Send
```

## Related Code Files

### Files to Modify
- `internal/handler/rest/handler.go` - Refactor error handler

### Files to Create
- None (refactoring existing file)

### Files to Delete
- None

## Implementation Steps

1. **Extract Response Building Function** (30m)
   ```go
   // internal/handler/rest/handler.go

   // buildErrorResponse creates standardized error response
   func buildErrorResponse(
       code int,
       message string,
       path string,
       requestID string,
   ) map[string]interface{} {
       return map[string]interface{}{
           "msg":        message,
           "error":      message,
           "path":       path,
           "request_id": requestID,
       }
   }
   ```

2. **Simplify Error Code Handling** (30m)
   ```go
   // Use map lookup instead of switch statement
   var statusCodeMap = map[errors.ErrorCode]int{
       errors.ErrCodeInvalid:     http.StatusBadRequest,
       errors.ErrCodeNotFound:    http.StatusNotFound,
       errors.ErrCodeUnauthorized: http.StatusUnauthorized,
       errors.ErrCodeForbidden:   http.StatusForbidden,
       errors.ErrCodeConflict:    http.StatusConflict,
       errors.ErrCodeInternal:    http.StatusInternalServerError,
   }

   func getStatusCode(code errors.ErrorCode) int {
       if status, ok := statusCodeMap[code]; ok {
           return status
       }
       return http.StatusInternalServerError
   }
   ```

3. **Extract Validation Error Handler** (20m)
   ```go
   // handleValidationError extracts validation errors
   func handleValidationError(err error, path, requestID string) map[string]interface{} {
       var validationErr *validator.ValidationError
       if errors.As(err, &validationErr) {
           return map[string]interface{}{
               "msg": validationErr.Detail.Msg,
               "error": validationErr.Detail.Msg,
               "errors": []map[string]interface{}{
                   {
                       "property": validationErr.Detail.Property,
                       "msg":      validationErr.Detail.Msg,
                   },
               },
               "path":       path,
               "request_id": requestID,
           }
       }
       return buildErrorResponse(
           http.StatusBadRequest,
           "invalid request parameters",
           path,
           requestID,
       )
   }
   ```

4. **Simplify customHTTPErrorHandler** (40m)
   ```go
   func customHTTPErrorHandler(err error, c echo.Context) {
       // Extract request ID
       requestID := c.Response().Header().Get(echo.HeaderXRequestID)

       // Try custom error types first
       if appErr, ok := err.(*errors.AppError); ok {
           code := getStatusCode(appErr.Code)
           response := buildErrorResponse(
               code,
               getErrorMessage(appErr),
               c.Path(),
               requestID,
           )
           logError(err, c.Path(), code)
           c.JSON(code, response)
           return
       }

       // Try Echo HTTP errors
       if httpErr, ok := err.(*echo.HTTPError); ok {
           handleHTTPError(httpErr, c, requestID)
           return
       }

       // Fallback to internal server error
       handleInternalError(err, c, requestID)
   }
   ```

5. **Add Unit Tests** (20m)
   ```go
   // internal/handler/rest/handler_test.go
   func TestBuildErrorResponse(t *testing.T)
   func TestHandleValidationError(t *testing.T)
   func TestCustomHTTPErrorHandler_AppError(t *testing.T)
   func TestCustomHTTPErrorHandler_HTTPError(t *testing.T)
   func TestCustomHTTPErrorHandler_InternalError(t *testing.T)
   ```

## Todo List

- [x] Extract response building function (buildErrorResponse)
- [x] Create status code map for error codes
- [x] Extract validation error handler (handleValidationError)
- [x] Simplify customHTTPErrorHandler (reduce to 50 lines)
- [x] Add unit tests for new functions
- [x] Test all error paths (App, HTTP, Internal)
- [x] Verify responses match original format
- [x] Check file size is under 200 lines
- [ ] Run integration tests
- [ ] Performance test (no regression)

## Success Criteria

**Definition of Done**:
- customHTTPErrorHandler reduced to <80 lines
- Total handler.go file <200 lines
- All error paths tested
- HTTP responses identical to original
- All tests passing
- Code more readable and maintainable

**Validation Methods**:
```bash
# Run tests
go test ./internal/handler/rest/... -v

# Check file size
wc -l internal/handler/rest/handler.go
# Expected: <200 lines

# Integration test all error types
curl -X POST http://localhost:8080/api/login -d '{}'
curl -X GET http://localhost:8080/api/notfound
curl -X POST http://localhost:8080/api/register -d '{"username":"test"}'

# Compare responses before/after (must be identical)
```

**Complexity Reduction**:
- Cyclomatic complexity reduced
- Nesting depth reduced to ≤2 levels
- Function length reduced to ≤50 lines

## Risk Assessment

**Potential Issues**:
1. **Response Format Changes**: Refactoring may change responses
   - Mitigation: Unit tests verify exact response structure
2. **Missing Error Paths**: Simplification may skip error types
   - Mitigation: Test all error paths, use integration tests
3. **Logging Changes**: Error logging may be affected
   - Mitigation: Verify logging output matches original

**Low-Medium Risk**: Refactoring existing logic, not changing behavior

**Rollback**: Git revert if breaking changes detected

## Security Considerations

- Error messages remain sanitized (no internal details)
- 5xx errors still use generic messages
- Validation errors don't leak sensitive data
- Request ID tracking maintained

## Next Steps

**Dependencies**: Phase 04 must be complete (Manual DI working)

**Follow-up Tasks**:
- Phase 06: Comprehensive testing and validation (final phase)

**Transition Criteria**:
- Handler file <200 lines → Ready for Phase 06
- All tests passing → Final validation
- Error responses verified → Ready for production testing
