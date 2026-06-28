# Red Team Review: Error Resolution Refactor — Backward-Compatibility Lens

**Plan:** `plans/260628-0918-error-resolution-refactor/plan.md`
**Reviewer lens:** API contract / backward compat
**Date:** 2026-06-28

---

## Summary

Phase 1 + Phase 2 are internal refactors with no observable wire change — low risk, sound. Phase 3 (drop `Response.Error`) is the load-bearing risk and the plan's mitigation ("sample owns its clients") is **asserted, not verified**. Evidence below shows (a) a second, divergent error producer the plan ignores, (b) a legacy-compat comment that explicitly forbids casual removal, and (c) zero versioning infrastructure to gate a "major bump."

---

## Findings

### [Critical] F1 — `ratelimit.go` emits a *different* error envelope; plan ignores it

`examples/golang-sample/internal/handler/rest/middlewares/ratelimit.go:140-143` returns:
```go
c.JSON(http.StatusTooManyRequests, map[string]string{
    "error": "Too many requests",
    "msg":   "Rate limit exceeded. Please try again later.",
})
```
This bypasses `apperrors.Response` entirely and — crucially — **relies on the `error` JSON key** that Phase 3 proposes to drop from the *canonical* shape. Post-refactor the API will have **two contradicting error envelopes**:
- `apperrors.Response` (no `error`)
- ratelimit map (has `error`, no `request_id`/`path`/`errors`)

The plan's Files-Affected table omits `ratelimit.go`. A client consuming "the `error` field" keeps working *only* on 429 and breaks everywhere else — worst possible inconsistent surface.

**Mitigation:** Phase 2 must fold `ratelimit.go` into `apperrors.Resolve` (it already maps `CodeRateLimit`). Removing the duplicate producer is strictly more valuable than removing one redundant field.

---

### [Major] F2 — Legacy-compat comment explicitly documents the `error` field as intentional

`internal/errors/response.go:9-11`:
> "The shape (msg, error, path) matches the legacy response format for **backward compatibility**; request_id and errors are optional."

This is a deliberate contract-preservation note, not vestigial cruft. The plan treats `Error` as "always equal to Msg, therefore redundant." That's true *today* but the field was reserved as the stable external hook. Removing it contradicts the codebase's own documented intent.

**Mitigation:** If the field is truly dead, first *update the comment* to reflect a deprecation decision; don't silently delete. A reader of git blame will otherwise see "removed the field the prior author flagged as compat-critical."

---

### [Major] F3 — "Sample owns its clients" is unverified; no evidence either way

The plan's Phase 3 recommendation rests entirely on this claim. Repo evidence:
- No `e2e/`, `contract/`, or `integration/` test directories.
- No generated client / SDK / `clients/` directories.
- No git tags or releases (`git tag` lists only backup branches) → **no semver contract to bump**. "Major version bump" is moot; there's no version line.
- No `@Failure` swagger annotations on handlers (`auth.go` has only `@Success`/`@Router`). Swagger does **not** currently publish the error schema.

So the claim is *plausible* but **unprovable from the repo**. External consumers (a frontend, a mobile build, a sibling repo) could exist outside this monorepo. The plan should say "no *internal* consumers found" and require explicit owner sign-off, not hand-wave.

**Mitigation:** Replace "sample owns its clients" with "grep found N internal producers, M internal test assertions, zero external/published contracts." Make the claim falsifiable.

---

### [Minor] F4 — No published OpenAPI error schema = silent contract drift

`auth.go` handlers carry `@Success`/`@Router` but no `@Failure` blocks referencing `apperrors.Response`. Means:
- Dropping `Error` won't break a generated SDK today (none generated).
- But it also means the *real* error shape is undocumented; any client reading `.error` today learned it by observation, not contract. Such clients are the hardest to notify.

**Mitigation:** Phase 3 should *add* `@Failure` annotations pinning the post-drop schema, so the contract becomes explicit going forward. Otherwise the refactor tightens internal cleanliness while leaving the external contract implicit.

---

### [Minor] F5 — Phase 3 ordering risks a flag-day

Plan runs P3 ("drop `Error`") as a hard delete in the same PR series as P1/P2. If any consumer does rely on `error`, there is no grace period — they break on upgrade with no migration path and no deprecation warning.

**Mitigation (if drop proceeds):** Insert a deprecation cycle:
1. Keep `Error` populated but mark the struct field `// Deprecated:` with godoc.
2. Log a one-time warning server-side when the field is emitted (cheap canary).
3. Drop in a later cycle after confirming zero reliance.

Cost: one struct field + one comment. Benefit: reversible.

---

### [Minor] F6 — Cost/benefit favors keeping the field

The `Error` field costs:
- 1 struct field, ~3 assignments, zero runtime overhead, zero new deps.

Removing it saves:
- ~5 lines of code, "redundancy" aesthetic.

Against: external contract risk (F2, F3), divergent producer risk (F1), lost deprecation window (F5). Asymmetric — the upside is cosmetic, the downside is client-visible breakage. YAGNI cuts both ways: you don't *need* to remove it.

---

## Things the plan gets right (for balance)

- Phase 1/2 are pure internal moves; no wire change — safe.
- `FieldErrorProvider` interface keeps `internal/errors` validator-free — correct dependency direction.
- Recognizes Echo handling must stay in the delivery layer.
- Explicitly flags P3 as a decision point rather than hiding it.

---

## VERDICT: **deprecate-first**

Phase 1 and Phase 2: proceed as planned.
Phase 3: **do not hard-drop `Response.Error` in this cycle.** Instead:
1. First, **fold `ratelimit.go` into `apperrors.Resolve`** (fixes F1, the actual inconsistency).
2. Mark `Error` `// Deprecated:` in `response.go`, update the legacy-compat comment to a deprecation note.
3. Add `@Failure` annotations to handlers (fixes F4).
4. Re-evaluate the drop after one release with the deprecation marker visible to any consumer.

If the team insists on dropping now, gate it behind explicit owner confirmation that the field has zero external readers — *not* behind "we own the clients," which the repo cannot prove.

---

## Unresolved questions

1. Does this error envelope escape the monorepo (frontend, mobile, sibling service)? Needs owner input, not grep.
2. Should `ratelimit.go`'s map be considered the "real" legacy contract and `apperrors.Response` the aspirational one? F1 suggests the opposite of the plan's assumption.
3. Is there a roadmap entry to publish OpenAPI error schemas? If yes, P3 should land alongside it, not before.
