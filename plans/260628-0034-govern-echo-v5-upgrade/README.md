# Govern Echo v5 Upgrade Plan

**Created:** 2026-06-28  
**Target:** github.com/haipham22/govern  
**Branch:** feat/monorepo-migration  
**Effort:** 24 hours  
**Status:** ✅ Completed 2026-06-28

---

## Quick Summary

Govern Echo integration was upgraded from Echo v4.15.1 to Echo v5.2.1.

- **Security fix:** CVE-2026-25766 addressed by moving to Echo v5.0.3+.
- **Breaking changes:** Handler signatures, error handler order, HTTPError messages, and middleware APIs updated.
- **Scope:** govern root module first, then downstream `examples/golang-sample/`.
- **Validation:** full build/test/race validation passed during migration.

---

## Priority Matrix

| Phase | Component | Risk | Effort | Status |
|-------|-----------|------|--------|--------|
| 01 | JWT Middleware | HIGH | 8h | ✅ Done |
| 02 | Context Helpers | MED | 5h | ✅ Done |
| 03 | Swagger Integration | LOW | 3h | ✅ Done |
| 04 | TrimStrings Middleware | LOW | 2h | ✅ Done |
| 05 | Tests & Documentation | MED | 6h | ✅ Done |

---

## Plan Structure

```
260628-0034-govern-echo-v5-upgrade/
├── README.md                    # This file
└── plan.md                      # Complete migration record
```

---

## Actual Execution

### Workspace

```bash
cd /Users/haipham22/Workspaces/haipham22/golang-sample
```

Root module is `github.com/haipham22/govern`. Sample module is `examples/golang-sample/` and imports govern via:

```go
require github.com/haipham22/govern v0.0.0
replace github.com/haipham22/govern => ../../
```

No `go.work` is used.

### Outcome

- Govern root `go.mod` uses `github.com/labstack/echo/v5 v5.2.1`.
- Govern `http/echo` handlers/helpers use `*echo.Context`.
- Govern `http/echo/trim.go` delegates to Echo v5 middleware.
- Sample app migrated to Echo v5 after govern compatibility landed.
- Echo v5 `BodyLimit` byte-count API handled in sample middleware.

---

## Phases Overview

### Phase 01: JWT Middleware (8h) — ✅ Done
- Updated context signatures: `echo.Context` → `*echo.Context`.
- Fixed `echo.NewHTTPError` string message usage.
- Updated middleware integration.

### Phase 02: Context Helpers (5h) — ✅ Done
- Updated `GetCurrentUser`, `MustGetCurrentUser` signatures.
- Updated `GetUserID`, `GetUsername` signatures.
- Fixed test helper usage.

### Phase 03: Swagger Integration (3h) — ✅ Done
- Updated route handler signatures.
- Fixed context passing in swagger routes.
- Validated swagger integration.

### Phase 04: TrimStrings Middleware (2h) — ✅ Done
- Swapped custom wrapper to Echo v5-compatible middleware.
- Validated string trimming behavior.

### Phase 05: Tests & Documentation (6h) — ✅ Done
- Updated tests for Echo v5 APIs.
- Synced migration docs.
- Ran validation.

---

## Critical Breaking Changes Handled

### 1. Context Signatures
```go
// Echo v4
func handler(c echo.Context) error

// Echo v5
func handler(c *echo.Context) error
```

### 2. Error Messages
```go
// Echo v4
return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid: %s", field))

// Echo v5
return echo.NewHTTPError(http.StatusBadRequest, "invalid field")
```

### 3. Error Handler Signature
```go
// Echo v4
func customHTTPErrorHandler(err error, c echo.Context)

// Echo v5
func customHTTPErrorHandler(c *echo.Context, err error)
```

### 4. BodyLimit API
```go
// Echo v4-style config
"1M"

// Echo v5-style config
int64(1 << 20)
```

---

## Security

### CVE-2026-25766
- **Issue:** Path traversal vulnerability on Windows.
- **Fixed:** Echo v5.0.3+.
- **Status:** Mitigated by Echo v5.2.1 migration.

---

## References

### Research
- [Echo v5 Govern Compatibility Analysis](../reports/researcher-260628-0030-echo-v5-govern-compatibility.md)

### Official Documentation
- [Echo API Changes V5](https://github.com/labstack/echo/blob/master/API_CHANGES_V5.md)
- [Echo v5 Release Notes](https://github.com/labstack/echo/releases)

---

## Success Criteria

- ✅ All tests pass with Echo v5.
- ✅ No production breaking changes found during validation.
- ✅ CVE-2026-25766 mitigated.
- ✅ Documentation updated.
- ⏳ govern package release/tag still separate release work.

---

## Follow-up

- Release/tag govern package with Echo v5 support when ready.
- Publish downstream migration note if needed.
- Monitor Echo v5 patch releases.

---

**Status:** ✅ Completed 2026-06-28  
**Target:** Achieved  
**Owner:** Development Team
