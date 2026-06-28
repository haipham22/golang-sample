# Plan: Move Error Resolution into the Sample (`internal/errors`)

**Date:** 2026-06-28 | **Branch:** feat/monorepo-migration | **Status:** Revised (post red-team)
**Scope:** `examples/golang-sample/internal/errors`, `internal/handler/rest` (+ `middlewares/ratelimit.go`), `internal/validator`, `internal/schemas`

---

## Context

The HTTP error path is inconsistent and scattered:

- `apperrors.Resolve` (`internal/errors`) already maps typed error → status+body and falls back to generic 500 — but `resolveError` (handler) re-does this work (double `GetCode`, redundant unknown→500 case).
- `enrichValidation` lives in the handler and reaches into `validator.ValidationError` — a validator concern in the delivery layer.
- **`ratelimit.go` (~140)** returns a raw `map[string]string{"error":..., "msg":...}` — a **third** error shape that bypasses `apperrors.Response` entirely. This is the real inconsistency.
- **`schemas.ErrResponseBody`** (response.go:36) is **dead code** (0 usages).
- `Response.Msg` and `Response.Error` are always identical (legacy compat) — flagged removable, but **deferred** (see Decisions).

**Goal:** `internal/errors.Resolve` becomes the single source for the error envelope; the handler is delivery-only; ratelimit joins the same envelope; dead schema deleted.

---

## Decisions (made after red-team review)

| # | Decision | Why |
|---|----------|-----|
| D1 | **Constructor, not interface** | Carry field errors on `*apperrors.Error` via a `Validation(...)` constructor; `Resolve` reads `e.Errors` directly. No reflection-based `errors.As` probe, no one-customer interface (YAGNI). |
| D2 | **Fold `ratelimit.go` into the envelope** (new Phase 0) | It's the actual inconsistency; dropping `Error` elsewhere without this makes 429 the lone outlier. |
| D3 | **Delete dead `schemas.ErrResponseBody`** | 0 usages. |
| D4 | **Defer dropping `Response.Error`** | Red team: deprecate-first. Not worth a flag-day break this cycle. Revisit after consolidation. |
| D5 | **Add `apperrors.NewBody(msg,path,requestID)` helper** | Centralize `Response{}` construction (currently 3 places) so a future `Error`-field drop is 1 line. |
| D6 | **Echo `HTTPError` stays in the handler** | `internal/errors` must not import Echo (verified: no `WrapCode(_, echo.HTTPError)` overlap exists, so echo-first ordering is safe). |

---

## Target Design

### 1. `*apperrors.Error` carries field errors (D1)

```go
// internal/errors/error.go
type Error struct {
    Code    Code
    Err     error
    message string
    Errors  []FieldError // populated for validation errors
}

// internal/errors/helpers.go
// Validation builds a CodeInvalid error with field-level detail.
func Validation(property, msg string) *Error {
    return &Error{
        Code:    CodeInvalid,
        message: msg,
        Errors:  []FieldError{{Property: property, Msg: msg}},
    }
}
```

### 2. Validator emits the typed error directly

`CustomValidator.Validate` returns `apperrors.Validation(property, msg)` instead of `*validator.ValidationError`. Controllers stop wrapping validation errors (`auth.go:39,76` `WrapCode(CodeInvalid, err)` → `return err`). `validator.ValidationError` type is removed (its tests updated).

### 3. `Resolve` reads `Errors` + uses `NewBody` (D5)

```go
func Resolve(err error, path, requestID string) (int, Response) {
    code, ok := GetCode(err)
    if !ok {
        return 500, NewBody("Internal Server Error", path, requestID)
    }
    body := NewBody(code.ClientMessage(), path, requestID)
    if code == CodeInvalid {
        var e *Error
        if errors.As(err, &e) && len(e.Errors) > 0 {
            body.Msg = e.Errors[0].Msg
            body.Errors = e.Errors
        }
    }
    return code.HTTPStatus(), body
}
```

### 4. Handler is delivery-only (D6)

```go
func resolveError(err error, path, requestID string) (int, apperrors.Response) {
    if he, ok := err.(*echo.HTTPError); ok {
        return resolveEchoError(he, path, requestID) // sanitize 5xx, build via NewBody
    }
    return apperrors.Resolve(err, path, requestID)   // typed + validation + unknown→500
}
```

Delete `enrichValidation`, the double `GetCode`, and the redundant unknown case.

### 5. Ratelimit joins the envelope (D2)

`ratelimit.go` returns an `apperrors`-typed error (`CodeRateLimit`) instead of the raw `map` — so it flows through `Resolve` and produces a consistent body.

---

## Phases

### Phase 0 — Consolidate producers (D2, D3)
- [ ] `ratelimit.go`: return `apperrors.New(CodeRateLimit, ...)` (or a `RateLimit(...)` helper) instead of raw `c.JSON(map{...})`
- [ ] Delete `schemas.ErrResponseBody` (verify 0 usages first)
- [ ] Add `NewBody(msg, path, requestID)` helper; route `Resolve` + handler echo/unknown branches through it
- [ ] `mise exec -- go test ./...`

### Phase 1 — Constructor-based enrichment (D1) — no behavior change
- [ ] Add `Errors []FieldError` to `*apperrors.Error`; add `Validation(property, msg)` constructor
- [ ] `CustomValidator.Validate` returns `apperrors.Validation(...)`; remove `validator.ValidationError`
- [ ] Controllers (`auth.go`): drop `WrapCode(CodeInvalid, err)` for validation → `return err`
- [ ] `Resolve`: read `e.Errors` on `CodeInvalid` (with fallback to generic msg when absent)
- [ ] **Run `./internal/handler/rest/...` too** (proves no behavior change)
- [ ] `mise exec -- go test ./internal/errors/... ./internal/validator/... ./internal/handler/rest/...`

### Phase 2 — Slim the handler (D6)
- [ ] Rewrite `resolveError` (echo-only + delegate); extract `resolveEchoError`
- [ ] Delete `enrichValidation`; clean unused imports (`errors`, `apiValidator`)
- [ ] Migrate `handler_test.go`: typed+validation assertions → `internal/errors`; echo + unknown stay in handler tests
- [ ] `mise exec -- go build ./... && mise exec -- go test -race ./...`

### Phase 3 — Drop `Response.Error` *(DEFERRED — D4)*
- [ ] Not this cycle. Revisit after Phase 0–2 land. When done: grep for `error` JSON-key reliance **before** dropping; update legacy-compat comment; consider deprecation alias.

---

## Tests (must-have cases — red-team flagged gaps)

- [ ] **5xx sanitization:** `apperrors.New(CodeInternal, "db pass xyz")` → `body.Msg` must NOT contain the secret
- [ ] **`CodeInvalid` w/o validation cause** → `Msg == "invalid request parameters"`, `Errors` empty
- [ ] **`CodeInvalid` with validation** → `Msg == field msg`, `Errors[0].Property` correct
- [ ] **`nil` err** guard (echo branch must not panic)
- [ ] **Echo 4xx pass-through** + **5xx sanitize** (existing, preserve)
- [ ] **Unknown → 500** (existing, preserve; also cover in `internal/errors`)
- [ ] **(defensive)** echo+apperrors overlap — unreachable today, pin a test
- [ ] **Ratelimit (429)** now flows through `Resolve` → consistent body

---

## Files Affected

| File | Change |
|------|--------|
| `internal/errors/error.go` | `+ Errors []FieldError` on `*Error` |
| `internal/errors/helpers.go` | `+ Validation(property, msg)` constructor |
| `internal/errors/response.go` | `Resolve` reads `Errors`; `+ NewBody` helper |
| `internal/validator/validator.go` | `Validate` returns `apperrors.Validation`; remove `ValidationError` |
| `internal/handler/rest/handler.go` | slim `resolveError`; `+ resolveEchoError`; delete `enrichValidation` |
| `internal/handler/rest/handler_test.go` | split/migrate assertions |
| `internal/handler/rest/middlewares/ratelimit.go` | return typed `CodeRateLimit` error |
| `internal/handler/rest/controllers/auth/auth.go` | drop validation `WrapCode` |
| `internal/schemas/response.go` | delete `ErrResponseBody` |

---

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Validator return-type change ripples to tests | Phase 1 updates `validator.ValidationError` tests in same step |
| Ratelimit body shape changes (was raw map) | Phase 0 — `error`/`msg` keys preserved via `Response`; covered by test |
| Echo+apperrors overlap (order flip) | Verified unreachable via grep; pin defensive test |
| Behavior drift in `Resolve` fallback | Explicit `CodeInvalid`-without-validation test |
| Dropping `Error` field later breaks clients | Deferred (D4); grep-before-drop when revisited |

---

## Success Criteria

- `resolveError` is delivery-only (echo + delegate); no validator knowledge
- `internal/errors.Resolve` is the single source for the **typed-error** envelope (status + body + field detail)
- `GetCode` invoked at most once per error
- Ratelimit (429) uses the same envelope as every other error
- Dead `ErrResponseBody` gone; all red-team test gaps covered
- `mise exec -- go test -race ./...` green; no new deps from `internal/errors`

---

## Red-Team Audit Trail (2026-06-28)

Three reviewers (architecture / backward-compat / correctness): verdicts **ship-with-changes · deprecate-first · behavior-risky**. Key overrides applied to this plan:
- 🔴 `ratelimit.go` raw-map producer missed → **Phase 0** (D2)
- 🔴 `schemas.ErrResponseBody` dead → delete (D3)
- 🟡 interface → constructor (D1)
- 🟡 drop `Error` field → defer (D4)
- 🟡 order-of-checks flip → verified safe (D6) + defensive test
- 🟡 test gaps (5xx sanitize, CodeInvalid fallback, nil) → Tests section

---

## Open Questions

1. Ratelimit message: keep current wording ("Rate limit exceeded…") verbatim, or align to `ClientMessage()` ("Too many requests")? → keep current wording via constructor arg
2. `Validation(...)` supports only single field (parity with today); multi-error left as future contract change → documented on the constructor
