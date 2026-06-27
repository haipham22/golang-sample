# Echo v4 to v5 Migration Research Report

**Date:** 2026-06-28  
**Project:** golang-sample  
**Current Version:** Echo v4.15.1  
**Target Version:** Echo v5  
**Status:** Research Complete

---

## Executive Summary

Echo v5 is a **major breaking release** with 15+ critical breaking changes. Migration from v4 to v5 requires systematic code updates across handlers, middleware, error handling, and initialization. The current codebase uses Echo v4.15.1 with custom middleware (CORS, security headers, rate limiting) and govern error integration - all require updates.

**Migration Priority:** HIGH (but not urgent - v4 remains stable)  
**Risk Level:** MEDIUM-HIGH (widespread handler signature changes)  
**Estimated Effort:** 4-6 hours for comprehensive migration

---

## Critical Breaking Changes

### 1. Context: Interface → Concrete Struct (🔴 CRITICAL)

**Impact:** Every handler function signature changes

**v4:**
```go
func MyHandler(c echo.Context) error { ... }
```

**v5:**
```go
func MyHandler(c *echo.Context) error { ... }
```

**Current Codebase Impact:**
- ✅ `handler.go:76` - `customHTTPErrorHandler(err error, c echo.Context)` → `c *echo.Context`
- ✅ `middlewares/cors.go:66-82` - All middleware functions using `echo.Context`
- ✅ `middlewares/security.go:38-77` - `SecurityHeadersWithConfig` uses `echo.Context`
- ✅ `middlewares/ratelimit.go:75-149` - `RateLimitWithConfig` uses `echo.Context`
- ✅ `controllers/auth/auth.go:35-96` - All handler methods use `echo.Context`

**Migration Path:**
```bash
# Global replacement (most changes)
find . -type f -name "*.go" -exec sed -i 's/echo\.Context/*echo.Context/g' {} +
```

---

### 2. HTTPErrorHandler: Parameter Swapped (🔴 BREAKING)

**Impact:** Custom error handler signature

**v4:**
```go
type HTTPErrorHandler func(err error, c Context)
e.HTTPErrorHandler = func(err error, c echo.Context) { ... }
```

**v5:**
```go
type HTTPErrorHandler func(c *Context, err error)  // ORDER REVERSED
e.HTTPErrorHandler = func(c *echo.Context, err error) { ... }

// OR use factory
e.HTTPErrorHandler = echo.DefaultHTTPErrorHandler(true)  // exposeError=true
```

**Current Codebase Impact:**
- 🔴 `handler.go:76-184` - `customHTTPErrorHandler` signature must swap parameters

**Migration Required:**
```go
// Before (v4)
func customHTTPErrorHandler(err error, c echo.Context) {
    // err first, c second
    if errCode, ok := governerrors.GetCode(err); ok { ... }
}

// After (v5)
func customHTTPErrorHandler(c *echo.Context, err error) {
    // c first, err second
    if errCode, ok := governerrors.GetCode(err); ok { ... }
}
```

---

### 3. Logger: Custom Interface → slog.Logger (🔴 BREAKING)

**Impact:** All logging code

**v4:**
```go
type Echo struct {
    Logger Logger  // Custom interface
}
func (c Context) Logger() Logger
```

**v5:**
```go
import "log/slog"

type Echo struct {
    Logger *slog.Logger  // Standard library
}
func (c *Context) Logger() *slog.Logger
```

**Current Codebase Impact:**
- 🔴 `handler.go:124-178` - Uses `zap.L().Error()` directly (not affected)
- ✅ Good: Current code uses `zap.L()` directly, not `c.Logger()`
- ⚠️ But `e.Logger` property access would break if added

**Migration:**
- Current code uses `zap.L()` directly → **No immediate changes needed**
- Future `e.Logger` access must use `*slog.Logger`

---

### 4. HTTPError: Message Type Changed (🟡 MODERATE)

**Impact:** Error creation

**v4:**
```go
type HTTPError struct {
    Message interface{}  // Can be any type
}
func NewHTTPError(code int, message ...interface{}) *HTTPError
```

**v5:**
```go
type HTTPError struct {
    Message string  // Now string only
}
func NewHTTPError(code int, message string) *HTTPError  // Single string param
```

**Current Codebase Impact:**
- ⚠️ `handler.go:135-156` - Uses `echo.HTTPError` type assertion
- ⚠️ Must handle `he.Message` as string, not `interface{}`

**Migration:**
```go
// v5 - Message is now string
if he, ok := err.(*echo.HTTPError); ok {
    code = he.Code
    clientMsg := he.Message  // Already string, no formatting needed
}
```

---

### 5. Route Return Types (🟡 MODERATE)

**Impact:** Route registration return values

**v4:**
```go
func (e *Echo) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
func (e *Echo) Routes() []*Route
```

**v5:**
```go
func (e *Echo) GET(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo
func (e *Echo) Routes() Routes  // New collection type
```

**Current Codebase Impact:**
- ⚠️ `routes.go` (if it exists) - Route iteration patterns
- ✅ Current code doesn't store route return values

---

### 6. Response Return Type Changed (🟡 MODERATE)

**Impact:** Response access

**v4:**
```go
func (c Context) Response() *Response
type Response struct {
    Writer http.ResponseWriter
    // ...
}
```

**v5:**
```go
func (c *Context) Response() http.ResponseWriter  // Returns stdlib type
type Response struct {
    http.ResponseWriter  // Embedded
    // ...
}
```

**Current Codebase Impact:**
- ⚠️ `handler.go:181` - `c.Response().Committed` access
- ⚠️ All `c.Response().Header().Set()` calls

**Migration:**
```go
// v5 - Response() returns http.ResponseWriter
c.Response().Header().Set("X-Custom", "value")  // Works on stdlib type

// To get *echo.Response
resp, err := echo.UnwrapResponse(c.Response())
if err == nil {
    _ = resp.Committed
}
```

---

### 7. Middleware Signature Changes (🟡 MODERATE)

**Impact:** Custom middleware

**v4:**
```go
type MiddlewareFunc func(HandlerFunc) HandlerFunc
type HandlerFunc func(Context) error
```

**v5:**
```go
type MiddlewareFunc func(HandlerFunc) HandlerFunc  // Same signature
type HandlerFunc func(*Context) error  // Uses pointer
```

**Current Codebase Impact:**
- ✅ `middlewares/cors.go:66-82` - Returns `echo.MiddlewareFunc` (compatible)
- ✅ `middlewares/security.go:38-77` - Custom middleware pattern works
- ⚠️ But all nested `echo.Context` → `*echo.Context` inside

**Migration:**
```go
// v5 middleware (signature unchanged, internal context changes)
func SecurityHeadersWithConfig(config SecurityHeadersConfig) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c *echo.Context) error {  // ← Pointer added
            c.Response().Header().Set("X-Frame-Options", config.FrameOptions)
            return next(c)
        }
    }
}
```

---

## Specific Codebase Changes Required

### File: `internal/handler/rest/handler.go`

**Lines 76-184:** `customHTTPErrorHandler` signature
```go
// BEFORE (v4)
func customHTTPErrorHandler(err error, c echo.Context) {
    // ...
    if he, ok := err.(*echo.HTTPError); ok {
        clientMsg = fmt.Sprintf("%v", he.Message)
    }
}

// AFTER (v5)
func customHTTPErrorHandler(c *echo.Context, err error) {  // Parameters swapped
    // ...
    if he, ok := err.(*echo.HTTPError); ok {
        clientMsg = he.Message  // Already string, no fmt.Sprintf needed
    }
}
```

**Line 58:** Error handler assignment
```go
// No change needed - signature matches new pattern
e.HTTPErrorHandler = customHTTPErrorHandler
```

**Line 60:** IPExtractor
```go
// No change - API unchanged
e.IPExtractor = echo.ExtractIPFromRealIPHeader()
```

---

### File: `internal/handler/rest/middlewares/cors.go`

**Lines 66-82:** Middleware functions
```go
// BEFORE (v4)
func CORS() echo.MiddlewareFunc {
    return CORSWithConfig(DefaultCORSConfig())
}

// AFTER (v5) - Same signature, internal echo.Context changes
func CORS() echo.MiddlewareFunc {
    return CORSWithConfig(DefaultCORSConfig())
}
```

**Lines 71-82:** CORSWithConfig
```go
// No changes needed - echomiddleware.CORSWithConfig handles Context pointer
func CORSWithConfig(config CORSConfig) echo.MiddlewareFunc {
    corsConfig := echomiddleware.CORSConfig{
        AllowOrigins: config.AllowOrigins,
        // ... rest of config
    }
    return echomiddleware.CORSWithConfig(corsConfig)
}
```

---

### File: `internal/handler/rest/middlewares/security.go`

**Lines 38-77:** SecurityHeadersWithConfig
```go
// BEFORE (v4)
func SecurityHeadersWithConfig(config SecurityHeadersConfig) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {  // ← Change here
            c.Response().Header().Set("X-Frame-Options", config.FrameOptions)
            // ... rest of headers
            return next(c)
        }
    }
}

// AFTER (v5)
func SecurityHeadersWithConfig(config SecurityHeadersConfig) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c *echo.Context) error {  // ← Pointer added
            c.Response().Header().Set("X-Frame-Options", config.FrameOptions)
            // ... rest of headers (no other changes)
            return next(c)
        }
    }
}
```

---

### File: `internal/handler/rest/middlewares/ratelimit.go`

**Lines 75-149:** RateLimitWithConfig
```go
// BEFORE (v4)
func RateLimitWithConfig(ctx context.Context, config RateLimiterConfig) echo.MiddlewareFunc {
    // ...
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {  // ← Change here
            ip := c.RealIP()
            if ip == "" {
                ip = c.Request().RemoteAddr
            }
            // ... rest of logic
            return next(c)
        }
    }
}

// AFTER (v5)
func RateLimitWithConfig(ctx context.Context, config RateLimiterConfig) echo.MiddlewareFunc {
    // ... no changes to setup logic
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c *echo.Context) error {  // ← Pointer added
            ip := c.RealIP()
            if ip == "" {
                ip = c.Request().RemoteAddr
            }
            // ... rest of logic unchanged
            return next(c)
        }
    }
}
```

---

### File: `internal/handler/rest/controllers/auth/auth.go`

**Lines 35-96:** All handler methods
```go
// BEFORE (v4)
func (h *Controller) PostRegister(c echo.Context) error { ... }
func (h *Controller) PostLogin(c echo.Context) error { ... }

// AFTER (v5)
func (h *Controller) PostRegister(c *echo.Context) error { ... }  // Pointer added
func (h *Controller) PostLogin(c *echo.Context) error { ... }    // Pointer added
```

---

## New Features Available in v5

### 1. Type-Safe Parameter Extraction

**v5 Generic Helpers:**
```go
// Instead of manual parsing
idStr := c.Param("id")
id, err := strconv.Atoi(idStr)

// Use type-safe generics
id, err := echo.PathParam[int](c, "id")
page, err := echo.QueryParamOr[int](c, "page", 1)
tags, err := echo.QueryParams[string](c, "tags")
```

**Impact:** Optional enhancement - current code uses `c.Bind()` for JSON

---

### 2. slog.Logger Integration

**v5 Standard Logging:**
```go
import "log/slog"

e.Logger.Info("Server started")
c.Logger().Error("Request failed", "path", c.Path())
```

**Current Code:** Uses `zap.L()` directly - compatible with v5

---

### 3. StartConfig for Server Management

**v5 Server Startup:**
```go
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()

sc := echo.StartConfig{
    Address:         ":8080",
    GracefulTimeout: 10 * time.Second,
}
sc.Start(ctx, e)
```

**Current Code:** Uses `governhttp.NewServer` - abstracts Echo server

---

### 4. Routes Collection with Filtering

**v5 Route Operations:**
```go
routes := e.Router().Routes()
authRoutes := routes.FilterByPath("/api/auth")
route, _ := routes.FindByMethodPath("GET", "/api/users")
```

**Impact:** No current usage of route iteration

---

## Migration Strategy

### Phase 1: Preparation (15 min)
1. Create feature branch: `git checkout -b echo-v5-migration`
2. Update `go.mod`: `github.com/labstack/echo/v5 v5.0.0`
3. Run: `go mod tidy`

### Phase 2: Global Replacements (30 min)
```bash
# Update import paths
find . -type f -name "*.go" -exec sed -i 's/github\.com\/labstack\/echo\/v4/github.com\/labstack\/echo\/v5/g' {} +

# Update handler signatures (echo.Context → *echo.Context)
find . -type f -name "*.go" -exec sed -i 's/echo\.Context/*echo.Context/g' {} +
```

### Phase 3: Manual Updates (2-3 hours)
1. **Fix HTTPErrorHandler parameter swap** (`handler.go:76`)
2. **Fix HTTPError.Message handling** (remove `fmt.Sprintf`)
3. **Fix Response() access** where `Committed` field used
4. **Test all middleware** (CORS, Security, RateLimit)
5. **Test all handlers** (auth, health)

### Phase 4: Validation (1-2 hours)
1. Compile: `mise exec -- go build ./...`
2. Lint: `mise exec -- golangci-lint run`
3. Test: `mise exec -- go test ./... -v`
4. Manual smoke test: Run server, test endpoints

### Phase 5: Optional Enhancements (1 hour)
- Adopt type-safe parameter extraction where applicable
- Use `echotest` package for handler tests
- Implement `StartConfig` if governhttp supports it

---

## Risk Assessment

### High Risk Items
- 🔴 **HTTPErrorHandler parameter swap** - Easy to miss, causes runtime panic
- 🔴 **Handler signatures** - Widespread changes, human error possible

### Medium Risk Items
- 🟡 **HTTPError.Message type** - Subtle bugs if formatting assumed
- 🟡 **Response() return type** - Field access may break

### Low Risk Items
- 🟢 **Import paths** - Automated replacement reliable
- 🟢 **Middleware signatures** - Unchanged, internal updates only

---

## Breaking Changes Summary

| Category | Change Type | Count | Risk Level |
|----------|-------------|-------|------------|
| Handler Signatures | Context → *Context | 15+ | 🔴 HIGH |
| Error Handler | Parameter swap | 1 | 🔴 HIGH |
| Logger | Interface → slog.Logger | 2-3 | 🟡 MEDIUM |
| HTTPError | Message type | 2-3 | 🟡 MEDIUM |
| Route Returns | *Route → RouteInfo | 0-5 | 🟡 MEDIUM |
| Response | Return type change | 5-10 | 🟡 MEDIUM |
| Middleware | Internal Context | 10+ | 🟢 LOW |

---

## Compatibility Notes

### Go Version
- **v4:** Go 1.24.0+
- **v5:** Go 1.25.0+
- **Current:** Go 1.25.0 ✅ (compatible)

### Dependencies
- ✅ `github.com/haipham22/govern` - Must verify Echo v5 compatibility
- ✅ `github.com/labstack/gommon` - Shared utilities (likely compatible)
- ⚠️ `github.com/swaggo/swag` - Swagger integration (test required)

---

## Recommendations

### Immediate Actions
1. **Hold migration** - v4 is stable and production-ready
2. **Monitor Echo v5 adoption** - Wait for wider community testing
3. **Check govern package** - Verify `govern/http/echo` supports Echo v5

### Short-Term (Next 3 months)
1. Create migration plan as documented above
2. Set up staging environment for v5 testing
3. Run comprehensive test suite against v5

### Long-Term (6-12 months)
1. Migrate to v5 once ecosystem stabilizes
2. Adopt v5 type-safe parameter extraction
3. Leverage new slog.Logger integration

---

## Migration Checklist

- [ ] Create feature branch
- [ ] Update go.mod to echo/v5
- [ ] Run `go mod tidy`
- [ ] Global import path replacement
- [ ] Global Context signature replacement
- [ ] Fix HTTPErrorHandler parameter swap
- [ ] Fix HTTPError.Message handling
- [ ] Fix Response() field access
- [ ] Test CORS middleware
- [ ] Test SecurityHeaders middleware
- [ ] Test RateLimit middleware
- [ ] Test auth handlers (PostRegister, PostLogin)
- [ ] Test health handler
- [ ] Run static analysis (`golangci-lint`)
- [ ] Run unit tests (`go test ./...`)
- [ ] Manual integration testing
- [ ] Update CLAUDE.md if new patterns adopted

---

## Unresolved Questions

1. **govern/http/echo compatibility**: Does `github.com/haipham22/govern/http/echo` support Echo v5? Must check with govern package maintainer or test.

2. **Swagger integration**: Does `swaggo/swag` work with Echo v5 handlers using `*echo.Context`? Need testing.

3. **Wire DI compatibility**: Current DI uses Wire - any Echo-specific injection points affected? Likely not, but verify.

4. **Deployment timeline**: When is Echo v5 required for production? Can defer until govern package officially supports v5.

---

## Sources

- [Echo API Changes V5 Official Documentation](https://github.com/labstack/echo/blob/master/API_CHANGES_V5.md) - **Primary source** for all breaking changes
- [Echo Framework GitHub Repository](https://github.com/labstack/echo) - Official releases and development
- [Echo v5 Breaking Changes Discussion (Reddit)](https://www.reddit.com/r/golang/comments/1qayn8i/breaking_changes_in_echo_v5/) - Community feedback
- [Echo Error Handling Guide](https://echo.labstack.com/guide/error-handling/) - Centralized error handling patterns
- [Echo Middleware Package Documentation (v5)](https://pkg.go.dev/github.com/labstack/echo/v5/middleware) - Middleware API changes
- [Echo CHANGELOG.md](https://github.com/labstack/echo/blob/master/CHANGELOG.md) - Release notes and improvements
- [Echo v5 Proposal Discussion](https://github.com/labstack/echo/discussions/2000) - v5 design discussion and community feedback

---

**Report Status:** ✅ COMPLETE  
**Next Review:** 2025-03-01 or when govern package confirms Echo v5 support
