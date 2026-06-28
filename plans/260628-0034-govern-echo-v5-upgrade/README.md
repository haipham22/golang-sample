# Govern Echo v5 Upgrade Plan

**Created:** 2026-06-28  
**Target:** github.com/haipham22/govern  
**Branch:** echo-v5-migration  
**Effort:** 24 hours

---

## 📋 Quick Summary

Upgrade govern package Echo integration from v4.15.1 to v5.2.1 with:
- **Security fix:** CVE-2026-25766 (path traversal on Windows)
- **Breaking changes:** 15+ API changes requiring systematic updates
- **5 files affected:** jwt.go, middleware.go, swagger.go, trim.go, context_test.go

---

## 🎯 Priority Matrix

| Phase | Component | Risk | Effort | Status |
|-------|-----------|------|--------|--------|
| 01 | JWT Middleware | HIGH | 8h | Pending |
| 02 | Context Helpers | MED | 5h | Pending |
| 03 | Swagger Integration | LOW | 3h | Pending |
| 04 | TrimStrings Middleware | LOW | 2h | Pending |
| 05 | Tests & Documentation | MED | 6h | Pending |

---

## 📁 Plan Structure

```
260628-0034-govern-echo-v5-upgrade/
├── README.md                    # This file
└── plan.md                      # Complete migration plan
```

---

## 🚀 Getting Started

### Prerequisites
1. Clone govern repository: `/Users/haipham22/Workspaces/haipham22/govern`
2. Create feature branch: `echo-v5-migration`
3. Read [plan.md](plan.md) for detailed steps

### Quick Start
```bash
# Navigate to govern directory
cd /Users/haipham22/Workspaces/haipham22/govern

# Create branch
git checkout -b echo-v5-migration

# Update Echo version
go get github.com/labstack/echo/v5@latest
go mod tidy

# Start with Phase 01: JWT Middleware
# See plan.md for detailed steps
```

---

## 📊 Phases Overview

### Phase 01: JWT Middleware (8h) - CRITICAL
- Update context signatures: `echo.Context` → `*echo.Context`
- Fix error handling: Remove `fmt.Sprintf` from errors
- Update middleware integration

### Phase 02: Context Helpers (5h)
- Update `GetCurrentUser`, `MustGetCurrentUser` signatures
- Update `GetUserID`, `GetUsername` signatures
- Fix test helper functions

### Phase 03: Swagger Integration (3h)
- Update route handler signatures
- Fix context passing in swagger routes
- Validate swagger UI serving

### Phase 04: TrimStrings Middleware (2h)
- Update middleware signature
- Fix context dereferencing
- Validate string trimming logic

### Phase 05: Tests & Documentation (6h)
- Update all test files
- Update README with Echo v5 examples
- Run full test suite

---

## ⚠️ Critical Breaking Changes

### 1. Context Signatures (Affects ALL handlers)
```go
// ❌ Echo v4
func handler(c echo.Context) error

// ✅ Echo v5
func handler(c *echo.Context) error  // ← pointer required
```

### 2. Error Messages (String only)
```go
// ❌ Echo v4
echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid: %s", field))

// ✅ Echo v5
echo.NewHTTPError(http.StatusBadRequest, "invalid field")  // string only
```

### 3. Error Handler Signature
```go
// ❌ Echo v4
func customHTTPErrorHandler(err error, c echo.Context)

// ✅ Echo v5
func customHTTPErrorHandler(c *echo.Context, err error)  // ← parameter swap
```

---

## 🔒 Security

### CVE-2026-25766
- **Issue:** Path traversal vulnerability on Windows
- **Fixed:** Echo v5.0.3+
- **Urgency:** HIGH for Windows deployments, LOW for Linux

---

## 📚 References

### Research
- [Echo v5 Govern Compatibility Analysis](../reports/researcher-260628-0030-echo-v5-govern-compatibility.md)

### Official Documentation
- [Echo API Changes V5](https://github.com/labstack/echo/blob/master/API_CHANGES_V5.md)
- [Echo v5 Release Notes](https://github.com/labstack/echo/releases)

---

## 🎯 Success Criteria

- ✅ All tests pass with Echo v5
- ✅ No breaking changes in production
- ✅ CVE-2026-25766 fixed
- ✅ Documentation updated
- ✅ govern package released with Echo v5 support

---

## 📞 Questions?

See [plan.md](plan.md) for:
- Detailed implementation steps
- Risk assessment
- Rollback strategy
- Timeline and milestones

---

**Status:** Ready for implementation  
**Target:** Week 1  
**Owner:** Development Team
