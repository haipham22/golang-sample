# govern/errors Usage Audit

**Date:** 2026-06-27  
**Purpose:** Complete audit of ALL govern/errors usages before migration  
**Finding:** Plan **underestimated scope by 3x** (2 files → 6 files)

---

## 📊 SUMMARY

**Total Files Using govern/errors:** 6 (not 2 as claimed)  
**Total Usages:** 36 occurrences  
**Error Codes Used:** All 6 codes (CodeInvalid, CodeNotFound, CodeUnauthorized, CodeForbidden, CodeConflict, CodeInternal)  
**Pre-built Errors:** ErrUnauthorized (2+ usages in impl.go + tests)

---

## 🚨 CRITICAL FINDINGS

### Files Using govern/errors:

1. **internal/handler/rest/handler.go** - 10 usages
   - Lines: 10 (import), 81, 83, 87, 94, 101, 108, 115, 167
   - CustomHTTPErrorHandler with all 6 error codes
   - ✅ Accounted for in plan

2. **internal/service/auth/impl.go** - 10 usages ⚠️ **MISSING FROM PLAN**
   - Lines: 11 (import), 46, 51, 56, 62, 77, 80, 91, 96, 101, 107
   - Service implementation with error creation/wrapping
   - Uses: WrapCode, NewCode, ErrUnauthorized (pre-built error)
   - ❌ **NOT in Phase 03 scope**

3. **internal/service/auth/service_test.go** - 8 usages ⚠️ **MISSING FROM PLAN**
   - Lines: 9 (import), 56, 68, 80, 92, 132, 133, 196, 206, 242, 244
   - Test assertions using error codes
   - Uses: CodeConflict, CodeInternal, ErrUnauthorized, GetCode
   - ❌ **NOT in Phase 03 scope**

4. **internal/handler/rest/controllers/auth/auth.go** - 2 usages ⚠️ **MISSING FROM PLAN**
   - Lines: 6 (import), 39, 76
   - Controller error wrapping
   - Uses: WrapCode with CodeInvalid
   - ❌ **NOT in Phase 03 scope**

5. **internal/handler/rest/controllers/auth/auth_test.go** - 4 usages ⚠️ **MISSING FROM PLAN**
   - Lines: 11 (import), 101, 116, 118, 157, 170, 172
   - Test mocks and assertions
   - Uses: NewCode, GetCode, ErrUnauthorized
   - ❌ **NOT in Phase 03 scope**

6. **internal/validator/validator.go** - 2 usages ⚠️ **MISSING FROM PLAN**
   - Lines: 9 (import), 31
   - Validation error wrapping
   - Uses: WrapCode with CodeInvalid to wrap ValidationError
   - ❌ **NOT in Phase 03 scope (identified by red team)**

---

## 🔍 DETAILED USAGE ANALYSIS

### Error Code Distribution:

| Error Code | Usages | Files |
|------------|--------|-------|
| CodeInvalid | 4 | handler.go (1), auth.go (2), validator.go (1) |
| CodeNotFound | 1 | handler.go (1) |
| CodeUnauthorized | 5+ | handler.go (1), impl.go (2), tests (2+) |
| CodeForbidden | 1 | handler.go (1) |
| CodeConflict | 7+ | handler.go (2), impl.go (3), tests (2+) |
| CodeInternal | 10+ | handler.go (2), impl.go (5), tests (2+) |

### Function Usage Distribution:

| Function | Usages | Purpose |
|----------|--------|---------|
| governerrors.GetCode() | 8 | Extract error code from wrapped error |
| governerrors.WrapCode() | 8 | Wrap errors with error code |
| governerrors.NewCode() | 5 | Create new error with code and message |
| governerrors.ErrUnauthorized | 4+ | Pre-built unauthorized error constant |

---

## ⚠️ COMPLEX INTEGRATION POINTS

### 1. ValidationError Wrapping (validator.go:31)

```go
return governerrors.WrapCode(governerrors.CodeInvalid,
    &ValidationError{Detail: detail})
```

**Issue:** Custom error system must support wrapping custom types while preserving `errors.As()` functionality.

### 2. Pre-built Error Constants (impl.go:96, 101)

```go
return nil, governerrors.ErrUnauthorized
```

**Issue:** Custom error system must support pre-built error constants, not just dynamic code creation.

### 3. Test Assertions (service_test.go:132-133)

```go
var govErr *governerrors.ErrorWithCode
assert.ErrorAs(t, err, &govErr)
assert.Equal(t, governerrors.CodeInternal, govErr.Code)
```

**Issue:** Tests depend on `*governerrors.ErrorWithCode` type. Custom system must provide equivalent type.

---

## 📋 PHASE 03 SCOPE CORRECTIONS

### Files to Update in Phase 03:

**Current Plan:** 2 files  
**Actual Requirement:** 6 files (+3 test files, +1 validator, +1 controller, +1 service)

**Missing from Phase 03:**
1. `internal/service/auth/impl.go` (10 usages) - HIGHEST PRIORITY
2. `internal/service/auth/service_test.go` (8 usages)
3. `internal/handler/rest/controllers/auth/auth.go` (2 usages)
4. `internal/handler/rest/controllers/auth/auth_test.go` (4 usages)
5. `internal/validator/validator.go` (2 usages)

### Phase 03 Revised Scope:

```
Phase 03: Replace govern/errors in ALL files (not just handler.go)
Effort: 5 hours (not 3 hours)

Files to update:
- internal/handler/rest/handler.go (10 usages) ✅
- internal/service/auth/impl.go (10 usages) ⚠️ MISSING
- internal/service/auth/service_test.go (8 usages) ⚠️ MISSING
- internal/handler/rest/controllers/auth/auth.go (2 usages) ⚠️ MISSING
- internal/handler/rest/controllers/auth/auth_test.go (4 usages) ⚠️ MISSING
- internal/validator/validator.go (2 usages) ⚠️ MISSING
```

---

## 🎯 RECOMMENDED ACTIONS

### Before Implementation:

1. **Add Missing Files to Phase 03**
   - Add all 5 missing files to scope
   - Update Phase 03 effort: 3h → 5h
   - Update total plan effort: 16h → 20h

2. **Handle Special Cases**
   - ValidationError wrapping support
   - Pre-built error constants (ErrUnauthorized)
   - ErrorWithCode type for test assertions

3. **Test Strategy**
   - All 3 test files need updates
   - Test assertions use ErrorWithCode type
   - Must preserve test behavior exactly

### During Implementation:

1. **Replace in Correct Order**
   - Start with service layer (impl.go)
   - Move to controller layer (auth.go)
   - Update handler last (handler.go)
   - Update tests after each layer

2. **Verify Each Replacement**
   - Run tests after each file
   - Check error responses match
   - Verify logging unchanged

---

## 📊 REVISED ESTIMATES

| Phase | Original | Revised | Change |
|-------|----------|---------|--------|
| Phase 01 | 2h | 2h | - |
| Phase 02 | 3h | 3h | - |
| **Phase 03** | **3h** | **5h** | **+2h** |
| Phase 04 | 4h | 4h | - |
| Phase 05 | 2h | 2h | - |
| Phase 06 | 2h | 2h | - |
| **Total** | **16h** | **18h** | **+2h** |

**Note:** This is MINIMUM revision. Red team suggests total should be 26h+ when accounting for complexity.

---

## ❓ UNRESOLVED QUESTIONS

1. **ErrorWithCode Type:** Will custom error system provide equivalent type for test assertions?
2. **Pre-built Errors:** Are we creating constants like `ErrUnauthorized`?
3. **ValidationError:** How to wrap while preserving `errors.As()` functionality?
4. **Test Updates:** Will test assertions need rewriting after error system change?

---

## 🎲 IMPACT ASSESSMENT

**Risk Level:** INCREASED from MEDIUM to HIGH

**Scope Impact:**
- Files to update: +150% (2 → 6)
- Usages to replace: +260% (10 → 36)
- Test files affected: +200% (1 → 3)

**Complexity Impact:**
- ValidationError wrapping requires special handling
- Pre-built errors need constant definitions
- Test assertions depend on ErrorWithCode type

**Timeline Impact:**
- Phase 03: +67% time (3h → 5h)
- Total plan: +13% time (16h → 18h minimum)

---

**Next Steps:**
1. Add all 6 files to Phase 03 scope
2. Design custom error system to handle all 3 special cases
3. Update test strategy for ErrorWithCode assertions
4. Re-verify Phase 03 effort estimate
