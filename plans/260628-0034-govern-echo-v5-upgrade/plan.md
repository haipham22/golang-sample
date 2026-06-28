---
title: "Govern Package Echo v4 → v5 Migration Plan"
status: completed
completed: 2026-06-28
description: "Upgrade govern package Echo integration from v4 to v5 with breaking changes and security fixes"
status: pending
priority: P2
effort: 24h
branch: echo-v5-migration
tags: [govern, echo-v5, migration, breaking-changes]
created: 2026-06-28
---

# Govern Package Echo v5 Migration Plan

**Project:** github.com/haipham22/govern  
**Current State:** Echo v4.15.1  
**Target State:** Echo v5.2.1  
**Location:** /Users/haipham22/Workspaces/haipham22/govern

**Executive Summary:** Upgrade govern package Echo integration from v4 to v5. Echo v5 includes critical security fix (CVE-2026-25766) and performance improvements. **Breaking changes affect 5 files** requiring systematic updates. Migration effort: 24 hours.

---

## Priority Matrix

| Component | Risk | Value | Effort | Timeline | Status |
|-----------|------|-------|--------|----------|--------|
| **JWT Middleware** | HIGH | HIGH | 8h | Week 1 | ✅ Required |
| **Context Helpers** | MED | HIGH | 5h | Week 1 | ✅ Required |
| **Swagger Integration** | LOW | MED | 3h | Week 1 | ✅ Required |
| **TrimStrings Middleware** | LOW | LOW | 2h | Week 1 | ✅ Required |
| **Tests & Documentation** | MED | MED | 6h | Week 1 | ✅ Required |

---

## Phase 01: JWT Middleware Migration (8 hours)

**Status:** `pending`  
**Priority:** **P1** (Critical Path)  
**Risk Level:** High  
**Dependencies:** None

### Overview
JWT authentication middleware requires extensive updates due to Echo v5 breaking changes.

### Breaking Changes Impact

**Handler Signatures:**
```go
// ❌ Echo v4 (current)
func JWTMiddleware(config *jwt.MiddlewareConfig) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // JWT validation logic
        }
    }
}

// ✅ Echo v5 (required)
func JWTMiddleware(config *jwt.MiddlewareConfig) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c *echo.Context) error {  // ← pointer required
            // JWT validation logic
        }
    }
}
```

**Error Handling:**
```go
// ❌ Echo v4
return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("invalid token: %s", err))

// ✅ Echo v5  
return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")  // string only
```

### Implementation Steps

#### Step 1: Update Context Signatures (3h)
**File:** `http/echo/jwt.go`

**Changes:**
- Replace `echo.Context` with `*echo.Context` in all handler functions
- Update middleware function signatures
- Update context helper functions

**Example:**
```go
// Before
func GetCurrentUser(c echo.Context) (*jwt.Claims, bool)

// After
func GetCurrentUser(c *echo.Context) (*jwt.Claims, bool)
```

#### Step 2: Fix Error Handling (2h)
**File:** `http/echo/jwt.go`

**Changes:**
- Remove `fmt.Sprintf` from `echo.NewHTTPError` calls
- Use string constants or simple string concatenation
- Update error message patterns

**Example:**
```go
// Before
return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("Bearer token required: %s", msg))

// After
return echo.NewHTTPError(http.StatusUnauthorized, "Bearer token required")
```

#### Step 3: Update JWT Middleware Integration (2h)
**File:** `http/echo/jwt.go`

**Changes:**
- Update middleware registration pattern
- Fix context passing to handlers
- Validate JWT extraction logic

#### Step 4: Update Context Helper Functions (1h)
**File:** `http/echo/context_test.go`

**Changes:**
- Update `GetCurrentUser`, `MustGetCurrentUser` signatures
- Update `GetUserID`, `GetUsername` signatures
- Fix test helper functions

### Success Criteria
- ✅ All handler signatures use `*echo.Context`
- ✅ Error messages use strings only
- ✅ Middleware compiles without errors
- ✅ JWT authentication works correctly

### Risk Assessment
- **Risk Level:** High
- **Mitigation:** Comprehensive test coverage validates changes
- **Rollback:** Simple `go get echo/v4` if critical issues

---

## Phase 02: Context Helpers Migration (5 hours)

**Status:** `pending`  
**Priority:** **P1** (Critical Path)  
**Risk Level:** Medium  
**Dependencies:** Phase 01

### Overview
Update context helper functions used throughout govern Echo integration.

### Breaking Changes

**Context Helper Functions:**
```go
// ❌ Echo v4
func GetCurrentUser(c echo.Context) (*jwt.Claims, bool)
func MustGetCurrentUser(c echo.Context) *jwt.Claims
func GetUserID(c echo.Context) (string, bool)
func GetUsername(c echo.Context) (string, bool)

// ✅ Echo v5
func GetCurrentUser(c *echo.Context) (*jwt.Claims, bool)
func MustGetCurrentUser(c *echo.Context) *jwt.Claims
func GetUserID(c *echo.Context) (string, bool)
func GetUsername(c *echo.Context) (string, bool)
```

### Implementation Steps

#### Step 1: Update Context Signatures (2h)
**Files:** `http/echo/context_test.go`, `http/echo/jwt.go`

**Changes:**
- Update all context helper function signatures
- Fix context parameter types
- Update return value handling

#### Step 2: Update Test Helpers (2h)
**File:** `http/echo/context_test.go`

**Changes:**
- Update test helper functions
- Fix context creation in tests
- Update assertions

#### Step 3: Validate Integration (1h)
**Commands:**
```bash
go test ./http/echo/...
go test -race ./http/echo/...
```

### Success Criteria
- ✅ All context helpers use `*echo.Context`
- ✅ All tests pass
- ✅ No race conditions detected

---

## Phase 03: Swagger Integration Migration (3 hours)

**Status:** `pending`  
**Priority:** **P2** (Required)  
**Risk Level:** Low  
**Dependencies:** Phase 01

### Overview
Swagger integration requires minimal updates for Echo v5 compatibility.

### Breaking Changes

**Swagger Setup:**
```go
// ❌ Echo v4
func WithEchoSwagger(e *echo.Echo, opts ...SwaggerOption) error

// ✅ Echo v5 (same signature, minimal changes)
func WithEchoSwagger(e *echo.Echo, opts ...SwaggerOption) error
```

### Implementation Steps

#### Step 1: Update Route Handlers (1.5h)
**File:** `http/echo/swagger.go`

**Changes:**
- Update route handler signatures to use `*echo.Context`
- Fix context passing in swagger routes
- Validate swagger UI serving

#### Step 2: Update Swagger Authentication (1h)
**File:** `http/echo/swagger_auth_example.go`

**Changes:**
- Update authentication handlers
- Fix context usage in auth examples

#### Step 3: Test Swagger UI (30min)
**Commands:**
```bash
# Run integration test
go test -v ./http/echo/swagger_test.go

# Manual verification
# Start server and access /swagger/
```

### Success Criteria
- ✅ Swagger UI accessible
- ✅ Route handlers use `*echo.Context`
- ✅ Authentication works correctly

---

## Phase 04: TrimStrings Middleware Migration (2 hours)

**Status:** `pending`  
**Priority:** **P2** (Required)  
**Risk Level:** Low  
**Dependencies:** Phase 01

### Overview
TrimStrings middleware requires minimal updates for Echo v5 compatibility.

### Breaking Changes

**Middleware Signature:**
```go
// ❌ Echo v4
func TrimStrings(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Trim strings logic
    }
}

// ✅ Echo v5
func TrimStrings(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) error {  // ← pointer required
        // Trim strings logic
    }
}
```

### Implementation Steps

#### Step 1: Update Middleware Signature (1h)
**File:** `http/echo/trim.go`

**Changes:**
- Update handler signature to use `*echo.Context`
- Fix context dereferencing
- Validate string trimming logic

#### Step 2: Update Tests (1h)
**File:** `http/echo/trim_test.go`

**Changes:**
- Update test context creation
- Fix assertions
- Validate trimming behavior

### Success Criteria
- ✅ Middleware uses `*echo.Context`
- ✅ String trimming works correctly
- ✅ All tests pass

---

## Phase 05: Tests & Documentation Update (6 hours)

**Status:** `pending`  
**Priority:** **P1** (Required)  
**Risk Level:** Medium  
**Dependencies:** Phases 01-04

### Overview
Update all tests and documentation to reflect Echo v5 changes.

### Implementation Steps

#### Step 1: Update Unit Tests (2h)
**Files:** `http/echo/*_test.go`

**Changes:**
- Update context creation in all tests
- Fix assertions for new signatures
- Add Echo v5 specific tests

#### Step 2: Update Integration Tests (2h)
**Files:** `http/echo/*_test.go`

**Changes:**
- Update integration test scenarios
- Test JWT authentication flow
- Test swagger UI integration
- Test TrimStrings middleware

#### Step 3: Update Documentation (1h)
**Files:** `http/echo/README.md`

**Changes:**
- Update code examples to Echo v5
- Update installation instructions
- Document breaking changes
- Update migration guide

#### Step 4: Update go.mod (30min)
**File:** `go.mod`

**Changes:**
```bash
# Update Echo version
go get github.com/labstack/echo/v5@latest
go mod tidy
```

#### Step 5: Run Full Test Suite (30min)
**Commands:**
```bash
# Standard tests
go test ./http/echo/...

# Race detector
go test -race ./http/echo/...

# Coverage
go test -cover ./http/echo/...

# Build verification
go build ./...
```

### Success Criteria
- ✅ All tests pass
- ✅ No race conditions
- ✅ Coverage maintained
- ✅ Documentation updated
- ✅ go.mod updated to Echo v5

---

## Timeline & Milestones

### Week 1
- [ ] Phase 01: JWT Middleware Migration (8h)
- [ ] Phase 02: Context Helpers Migration (5h)
- [ ] Phase 03: Swagger Integration Migration (3h)
- [ ] Phase 04: TrimStrings Middleware Migration (2h)
- [ ] Phase 05: Tests & Documentation Update (6h)

### Week 2
- [ ] Release govern v0.x.x with Echo v5 support
- [ ] Update golang-sample to use new govern version
- [ ] Monitor for issues and bug fixes

---

## Rollback Strategy

### If Migration Fails

```bash
# Revert Echo version
go get github.com/labstack/echo/v4@latest
go mod tidy

# Revert code changes
git checkout HEAD~1 -- http/echo/
```

### Rollback Criteria
- Critical security vulnerabilities in Echo v5
- Breaking changes prevent basic functionality
- Testing reveals fundamental incompatibilities

---

## Security Considerations

### CVE-2026-25766 (Path Traversal Flaw)

**Impact:** Current Echo v4.15.1 vulnerable on Windows deployments

**Mitigation in Echo v5:**
- Fixed in Echo v5.0.3+
- Upgrade required for Windows deployments

**Urgency:** HIGH if deploying on Windows, LOW for Linux deployments

---

## Unresolved Questions

1. **Govern Maintenance:** Is govern actively maintained? What are Echo v5 plans?

2. **Backward Compatibility:** Should govern maintain Echo v4 support alongside v5?

3. **Release Timeline:** When to release govern Echo v5 version?

4. **Migration Support:** Provide migration guide for govern users?

---

## Dependencies & Blocking

### External Dependencies
- **Echo v5 stability:** Monitor for Echo v5 bug reports
- **Echo ecosystem:** Wait for middleware compatibility

### Internal Dependencies
- Phase 02 depends on Phase 01 (context helpers need JWT)
- Phase 03 depends on Phase 01 (swagger uses JWT auth)
- Phase 05 depends on Phases 01-04 (test everything)

---

## Resources & References

### Research Reports
- [Echo v5 Govern Compatibility Analysis](../reports/researcher-260628-0030-echo-v5-govern-compatibility.md)

### Official Documentation
- [Echo API Changes V5](https://github.com/labstack/echo/blob/master/API_CHANGES_V5.md)
- [Echo v5 Release Notes](https://github.com/labstack/echo/releases)
- [Echo Security Advisories](https://echo.labstack.com/docs/security)

### Project Documentation
- [Govern README](/Users/haipham22/Workspaces/haipham22/govern/README.md)
- [Govern Echo Integration](/Users/haipham22/Workspaces/haipham22/govern/http/echo/README.md)

---

## Next Steps

After govern Echo v5 migration complete:
1. Release govern package with Echo v5 support
2. Update golang-sample to use new govern version
3. Monitor for issues and bug fixes
4. Plan govern monorepo integration into golang-sample

---

**Plan Status:** ✅ Completed 2026-06-28 — govern `http/` migrated to Echo v5.2.1; full `go test ./...` green. Sample migrated in the same pass (plan 260628-0022 Phase 04).  
**Target Completion:** Week 1  
**Owner:** Development Team  
**Created:** 2026-06-28
