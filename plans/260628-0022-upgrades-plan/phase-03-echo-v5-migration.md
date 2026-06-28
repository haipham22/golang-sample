---
title: "Phase 03: Echo v4 to v5 Migration"
description: "Migrate from Echo v4 to v5 - DEFERRED until govern package confirms v5 support"
status: pending
priority: P3
effort: 6h
branch: main
tags: [echo-v5, migration, deferred, q3-q4-2026]
created: 2026-06-28
---

# Phase 03: Echo v4 to v5 Migration (DEFERRED)

**Status:** `pending`  
**Priority:** **P3** (DEFER - Q3-Q4 2026)  
**Risk Level:** Medium-High  
**Dependencies:** Govern Echo v5 support  
**Estimated Time:** 6 hours

---

## Overview

**RECOMMENDATION: DEFER migration** until govern package officially supports Echo v5. Echo v4 remains stable and production-ready. Migration has 15+ breaking changes requiring systematic updates across handlers, middleware, and error handling.

---

## Why Defer Migration?

### Critical Reasons

1. **govern/http/echo Compatibility Unknown**
   - Must verify `github.com/haipham22/govern/http/echo` supports Echo v5
   - Current code uses govern package for HTTP server management
   - Breaking changes in govern integration could block migration

2. **Echo v4 is Stable and Production-Ready**
   - No security issues requiring immediate upgrade
   - Actively maintained and supported
   - All current requirements met

3. **Medium-High Risk Migration**
   - 15+ breaking changes across codebase
   - Widespread handler signature changes
   - Error handler parameter swap (easy to miss)
   - Requires comprehensive testing

4. **Ecosystem Maturity**
   - Echo v5 newly released (February 2026)
   - Let community adopt and stabilize
   - Wait for wider v5 adoption
   - Learn from others' migration experiences

---

## Breaking Changes Summary

| Category | Changes | Affected Files | Risk Level |
|----------|---------|----------------|------------|
| **Handler Signatures** | `echo.Context` → `*echo.Context` | 15+ locations | 🔴 HIGH |
| **Error Handler** | Parameter swap `(err, c)` → `(c, err)` | 1 location | 🔴 HIGH |
| **Logger** | Interface → `*slog.Logger` | 2-3 locations | 🟡 MEDIUM |
| **HTTPError** | `Message interface{}` → `Message string` | 2-3 locations | 🟡 MEDIUM |
| **Response** | Return type `*Response` → `http.ResponseWriter` | 5-10 locations | 🟡 MEDIUM |
| **Middleware** | Internal Context changes | 10+ locations | 🟢 LOW |

---

## Current Codebase Impact

### Files Requiring Changes

#### 1. `internal/handler/rest/handler.go`

**Lines 76-184:** `customHTTPErrorHandler` signature

**Before (v4):**
```go
func customHTTPErrorHandler(err error, c echo.Context) {
    if errCode, ok := governerrors.GetCode(err); ok { ... }
    if he, ok := err.(*echo.HTTPError); ok {
        clientMsg = fmt.Sprintf("%v", he.Message)
    }
}
```

**After (v5):**
```go
func customHTTPErrorHandler(c *echo.Context, err error) {
    // Parameters swapped
    if errCode, ok := governerrors.GetCode(err); ok { ... }
    if he, ok := err.(*echo.HTTPError); ok {
        clientMsg = he.Message  // Already string, no fmt.Sprintf
    }
}
```

#### 2. `internal/handler/rest/middlewares/security.go`

**Lines 38-77:** `SecurityHeadersWithConfig`

**Before (v4):**
```go
return func(c echo.Context) error {
    c.Response().Header().Set("X-Frame-Options", config.FrameOptions)
    return next(c)
}
```

**After (v5):**
```go
return func(c *echo.Context) error {
    c.Response().Header().Set("X-Frame-Options", config.FrameOptions)
    return next(c)
}
```

#### 3. `internal/handler/rest/middlewares/ratelimit.go`

**Lines 75-149:** `RateLimitWithConfig`

**Before (v4):**
```go
return func(c echo.Context) error {
    ip := c.RealIP()
    if ip == "" {
        ip = c.Request().RemoteAddr
    }
    return next(c)
}
```

**After (v5):**
```go
return func(c *echo.Context) error {
    ip := c.RealIP()
    if ip == "" {
        ip = c.Request().RemoteAddr
    }
    return next(c)
}
```

#### 4. `internal/handler/rest/controllers/auth/auth.go`

**Lines 35-96:** All handler methods

**Before (v4):**
```go
func (h *Controller) PostRegister(c echo.Context) error { ... }
func (h *Controller) PostLogin(c echo.Context) error { ... }
```

**After (v5):**
```go
func (h *Controller) PostRegister(c *echo.Context) error { ... }
func (h *Controller) PostLogin(c *echo.Context) error { ... }
```

---

## When to Migrate

### Trigger Conditions

✅ **Ready to migrate when:**
- [ ] Govern package confirms Echo v5 support
- [ ] Echo v5 ecosystem stable (3-6 months post-release)
- [ ] Team capacity for migration work available
- [ ] No pressing feature deadlines
- [ ] Migration path documented and tested

❌ **Do not migrate if:**
- [ ] Govern package Echo v5 support unknown
- [ ] Critical production deadlines upcoming
- [ ] Team unfamiliar with Echo v5 patterns
- [ ] Echo v5 ecosystem unstable

---

## Migration Strategy (When Ready)

### Phase 1: Preparation (15 min)

#### Step 1.1: Create Feature Branch

```bash
git checkout -b echo-v5-migration
```

#### Step 1.2: Update Dependencies

```bash
go get github.com/labstack/echo/v5@latest
go mod tidy
```

#### Step 1.3: Verify Govern Compatibility

```bash
# Check govern package
grep -r "govern/http/echo" .

# Test Echo v5 compatibility
# If govern doesn't support v5, STOP and defer
```

**Acceptance Criteria:**
- ✅ Feature branch created
- ✅ Echo v5 dependency added
- ✅ Govern package compatibility verified

---

### Phase 2: Global Replacements (30 min)

#### Step 2.1: Update Import Paths

```bash
# Update import paths
find . -type f -name "*.go" -exec sed -i 's/github\.com\/labstack\/echo\/v4/github.com\/labstack\/echo\/v5/g' {} +
```

**Files affected:**
- `internal/handler/rest/handler.go`
- `internal/handler/rest/routes.go`
- `internal/handler/rest/middlewares/*.go`
- `internal/handler/rest/controllers/auth/auth.go`
- `go.mod`
- `go.sum`

#### Step 2.2: Update Handler Signatures

```bash
# Update handler signatures (echo.Context → *echo.Context)
find . -type f -name "*.go" -exec sed -i 's/echo\.Context/*echo.Context/g' {} +
```

**Acceptance Criteria:**
- ✅ All import paths updated
- ✅ All handler signatures updated
- ✅ No compilation errors from automated replacements

---

### Phase 3: Manual Fixes (3h)

#### Step 3.1: Fix HTTPErrorHandler Parameter Swap

**File:** `internal/handler/rest/handler.go:76`

**Before:**
```go
func customHTTPErrorHandler(err error, c echo.Context) {
```

**After:**
```go
func customHTTPErrorHandler(c *echo.Context, err error) {
```

**Critical:** This is the most dangerous change - parameter order reversed

#### Step 3.2: Fix HTTPError.Message Handling

**File:** `internal/handler/rest/handler.go:135-156`

**Before:**
```go
if he, ok := err.(*echo.HTTPError); ok {
    code = he.Code
    clientMsg = fmt.Sprintf("%v", he.Message)
}
```

**After:**
```go
if he, ok := err.(*echo.HTTPError); ok {
    code = he.Code
    clientMsg = he.Message  // Already string, no formatting needed
}
```

#### Step 3.3: Fix Response() Field Access

**File:** `internal/handler/rest/handler.go:181`

**Before:**
```go
if !c.Response().Committed {
    c.JSON(code, responseBody)
}
```

**After:**
```go
// Response() now returns http.ResponseWriter
// To check committed status, need to unwrap
resp, err := echo.UnwrapResponse(c.Response())
if err == nil && !resp.Committed {
    c.JSON(code, responseBody)
}
```

#### Step 3.4: Test All Middleware

**Files to test:**
- `middlewares/cors.go`
- `middlewares/security.go`
- `middlewares/ratelimit.go`

**Test:**
```bash
go test -v ./internal/handler/rest/middlewares/...
```

#### Step 3.5: Test All Handlers

**Files to test:**
- `controllers/auth/auth.go`
- `controllers/health/health.go`

**Test:**
```bash
go test -v ./internal/handler/rest/controllers/...
```

**Acceptance Criteria:**
- ✅ HTTPErrorHandler parameter swap completed
- ✅ HTTPError.Message handling fixed
- ✅ Response() field access updated
- ✅ All middleware tests pass
- ✅ All handler tests pass

---

### Phase 4: Validation (1h)

#### Step 4.1: Compile

```bash
mise exec -- go build ./...
```

#### Step 4.2: Static Analysis

```bash
mise exec -- golangci-lint run
mise exec -- staticcheck ./...
mise exec -- errcheck -blank ./...
```

#### Step 4.3: Run Tests

```bash
mise exec -- go test ./...
mise exec -- go test -race ./...
mise exec -- go test -cover ./...
```

#### Step 4.4: Manual Integration Testing

```bash
# Start server
mise exec -- go run cmd/serverd.go

# Test endpoints
curl http://localhost:8080/health
curl http://localhost:8080/readyz
curl http://localhost:8080/livez

# Test auth endpoints
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password"}'

curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"password"}'
```

**Acceptance Criteria:**
- ✅ Compilation succeeds
- ✅ Static analysis passes
- ✅ All tests pass
- ✅ Manual testing successful

---

### Phase 5: Documentation (30 min)

#### Step 5.1: Update CLAUDE.md

```markdown
## Tech Stack
| Component | Technology | Version |
|-----------|------------|---------|
| **Framework** | Echo | v5 (upgraded from v4) |
```

#### Step 5.2: Document v5-Specific Changes

```markdown
# Echo v5 Migration Notes

## Key Changes
- Handler signatures: `echo.Context` → `*echo.Context`
- Error handler: parameter swap `(err, c)` → `(c, err)`
- Logger: Custom interface → `*slog.Logger`
- HTTPError: `Message interface{}` → `Message string`

## Migration Date
- Migrated: [DATE]
- Tested: [DATE]
- Deployed: [DATE]
```

#### Step 5.3: Update Examples

```markdown
# Middleware Examples (Echo v5)

## Security Headers Middleware
func SecurityHeadersWithConfig(config SecurityHeadersConfig) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c *echo.Context) error {  // ← Pointer added
            c.Response().Header().Set("X-Frame-Options", config.FrameOptions)
            return next(c)
        }
    }
}
```

**Acceptance Criteria:**
- ✅ CLAUDE.md updated
- ✅ Migration notes documented
- ✅ Examples updated

---

## Risk Assessment

### Risk Level: Medium-High

**High Risk Items:**
- 🔴 **HTTPErrorHandler parameter swap** - Easy to miss, causes runtime panic
- 🔴 **Handler signatures** - Widespread changes, human error possible

**Medium Risk Items:**
- 🟡 **HTTPError.Message type** - Subtle bugs if formatting assumed
- 🟡 **Response() return type** - Field access may break

**Low Risk Items:**
- 🟢 **Import paths** - Automated replacement reliable
- 🟢 **Middleware signatures** - Unchanged, internal updates only

### Mitigation Strategies

1. **Govern Compatibility Check**
   - Verify govern package supports Echo v5 before starting
   - If not, defer migration

2. **Comprehensive Testing**
   - Unit tests for all handlers
   - Integration tests for middleware
   - Manual smoke testing

3. **Incremental Migration**
   - Migrate one handler at a time
   - Test after each change
   - Commit frequently

4. **Rollback Ready**
   - Keep Echo v4 branch ready
   - Simple `go get echo/v4` if issues

### Rollback Strategy

```bash
# If migration fails
git checkout main

# Or revert Echo version
go get github.com/labstack/echo/v4@latest
go mod tidy

# Revert global replacements
git checkout .
```

---

## Monitoring (Until Migration)

### Track Echo v5 Adoption

**Metrics to watch:**
- Echo v5 GitHub stars and issues
- Community adoption rate
- Bug reports and stability
- Third-party package compatibility

### Monitor Govern Package

**What to watch:**
- Govern package releases
- Echo v5 support announcements
- Breaking changes in govern

**Check regularly:**
```bash
# Check govern for Echo v5 updates
go list -m -u github.com/haipham22/govern

# Check govern GitHub
# Watch for Echo v5 support announcements
```

### Review Migration Experiences

**Sources:**
- Echo GitHub discussions
- Reddit r/golang
- Go community blogs
- Stack Overflow questions

**What to learn:**
- Common migration issues
- Workarounds for breaking changes
- Best practices from community
- Performance impact reports

---

## Unresolved Questions

### Critical Questions

1. **Govern Echo v5 Support**
   - When will `github.com/haipham22/govern/http/echo` support Echo v5?
   - Will there be breaking changes in govern integration?
   - Is govern planning Echo v5 support?

2. **Migration Timing**
   - When is Echo v5 required for production?
   - Are there security vulnerabilities in Echo v4?
   - What is the team capacity for migration?

3. **Ecosystem Stability**
   - When will Echo v5 ecosystem stabilize?
   - Are there major bugs in Echo v5?
   - What are common migration issues?

### Questions to Resolve Before Migration

1. Verify govern package Echo v5 support
2. Confirm no pressing deadlines
3. Ensure team availability
4. Validate Echo v5 stability
5. Document migration path

---

## Success Criteria (When Migrating)

### Required
- ✅ All tests pass with Echo v5
- ✅ No breaking changes in production
- ✅ Govern package compatibility verified
- ✅ Documentation updated

### Desired
- ✅ Performance validated
- ✅ No regression detected
- ✅ Migration notes documented

---

## Related Code Files

### Files to Modify (When Migrating)
- `go.mod` (Echo version)
- `internal/handler/rest/handler.go` (error handler)
- `internal/handler/rest/middlewares/*.go` (middleware)
- `internal/handler/rest/controllers/auth/auth.go` (handlers)
- `CLAUDE.md` (documentation)

### Files to Review
- `internal/handler/rest/routes.go`
- `internal/handler/rest/wire.go`
- All test files

---

## Data Flow

```
Current State (Echo v4)
      ↓
[DEFER] Wait for Govern Support
      ↓
[Trigger] Echo v5 Ready
      ↓
Phase 1: Preparation → Branch + Dependencies
      ↓
Phase 2: Global Replacements → Import paths + Signatures
      ↓
Phase 3: Manual Fixes → Error handler + Response access
      ↓
Phase 4: Validation → Compile + Test + Smoke test
      ↓
Phase 5: Documentation → Update docs + Notes
      ↓
Deployment → Monitor + Validate
```

---

## Dependency Graph

```
Phase 03 (Echo v5 Migration) - DEFERRED
├── Preparation
│   └── blocked by: Govern Echo v5 support
├── Global Replacements
│   └── depends on: Preparation
├── Manual Fixes
│   └── depends on: Global Replacements
├── Validation
│   └── depends on: Manual Fixes
└── Documentation
    └── depends on: Validation
```

---

## Timeline & Milestones

### Deferred (Q3-Q4 2026)

**Re-evaluation triggers:**
- [ ] Govern package confirms Echo v5 support
- [ ] Echo v5 ecosystem stable (3-6 months post-release)
- [ ] Team capacity available
- [ ] No critical deadlines

**When ready:**
- Week 1: Migration (Phases 1-5)
- Week 2: Testing and validation
- Week 3: Documentation and deployment

---

## Alternatives If Migration Fails

### Option 1: Stay on Echo v4
- Continue with Echo v4 indefinitely
- Monitor for security vulnerabilities
- Re-evaluate annually

### Option 2: Switch Framework
- Consider alternative frameworks (Gin, Fiber)
- High migration cost
- Last resort option

### Option 3: Fork Govern Package
- Add Echo v5 support ourselves
- High maintenance burden
- Not recommended

---

## Next Steps

### Immediate (Deferred)
- Monitor Echo v5 adoption
- Track govern package updates
- Review community experiences
- Plan migration timeline

### When Ready to Migrate
1. Verify govern package Echo v5 support
2. Create migration branch
3. Execute migration plan (Phases 1-5)
4. Comprehensive testing
5. Deploy to production

---

## References

### Research
- [Echo v5 Migration Research](../reports/researcher-260628-0022-echo-v5-migration-research.md)

### Official Documentation
- [Echo API Changes V5](https://github.com/labstack/echo/blob/master/API_CHANGES_V5.md)
- [Echo Framework GitHub](https://github.com/labstack/echo)
- [Echo v5 CHANGELOG](https://github.com/labstack/echo/blob/master/CHANGELOG.md)

### Project Documentation
- [README.md](../../README.md) - Current tech stack
- [CLAUDE.md](../../CLAUDE.md) - Development rules
- [System Architecture](../../docs/system-architecture.md)

---

## Migration Checklist (When Ready)

- [ ] Govern package Echo v5 support confirmed
- [ ] Feature branch created
- [ ] Echo v5 dependency added
- [ ] Import paths updated
- [ ] Handler signatures updated
- [ ] HTTPErrorHandler parameter swap fixed
- [ ] HTTPError.Message handling fixed
- [ ] Response() field access updated
- [ ] All middleware tests pass
- [ ] All handler tests pass
- [ ] Static analysis passes
- [ ] Manual smoke testing successful
- [ ] Documentation updated
- [ ] Migration notes documented
- [ ] Deployment to staging
- [ ] Production deployment
- [ ] Post-deployment monitoring

---

**Phase Status:** DEFERRED - Q3-Q4 2026  
**Trigger:** Govern package Echo v5 support  
**Completion Target:** 1 week when ready  
**Owner:** Development Team  
**Created:** 2026-06-28
