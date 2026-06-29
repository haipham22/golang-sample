# Mockery Rules

**Rules for generating and consuming mocks with mockery v3 (`github.com/vektra/mockery/v3`).**

---

## Overview

- **Tool:** [vektra/mockery](https://github.com/vektra/mockery) **v3** — managed by mise (`mockery = "latest"` in `mise.toml`)
- **Template:** `testify` — generates `github.com/stretchr/testify/mock` mocks
- **Config:** `.mockery.yml` at the repository root
- **Mode:** `all: true` — mocks **every** public interface in the packages listed in config

**Core rules:**
- ✅ Regenerate mocks after adding/changing any interface
- ✅ Generate via config (`mockery` with no args), not ad-hoc `--name`
- ✅ Use the type-safe `EXPECT()` expecter API in tests
- ✅ Commit generated mocks (CI verifies they stay in sync)
- ❌ Never hand-edit generated `mock_*.go` files
- ❌ Never mock concrete types — only interfaces

---

## Designing Interfaces for Mockability

Mockery can only mock well-structured interfaces. Design with mocking in mind:

### 1. Accept interfaces, return structs

Inject dependencies **as interfaces** so mocks can substitute them. Constructors take interface params, not concrete structs:

```go
// GOOD — Storage is an interface; mockable in tests
func NewAuthService(log *zap.SugaredLogger, storage Storage, secret string, exp time.Duration) Service {
    return &authService{log: log, storage: storage, ...}
}

// BAD — depends on concrete *repo; cannot inject a mock
func NewAuthService(log *zap.SugaredLogger, storage *repo, ...) Service { ... }
```

### 2. Exported interface + exported methods

Mockery generates mocks only for **exported** interfaces with **exported** methods:

```go
// GOOD — exported interface, exported methods → mockable
type Storage interface {
    FindUserByUsername(ctx context.Context, username string) (*model.User, error)
}

// BAD — unexported interface or methods → mockery skips it
type storage interface { ... }       // unexported, never mocked
```

### 3. All signature types must be exported and importable

Mockery emits an `import` for every type in the signature. Types from other packages must be **exported**:

```go
// GOOD — model.User is exported, defined in an importable package
Register(ctx context.Context, req RegisterRequest) (*model.User, error)

// BAD — leaks an unexported type from another package → mock won't compile
Get(ctx context.Context) (user internalType // unexported, cross-package → generation fails
```

Keep request/response DTOs **exported** and co-located with the interface (the project's `RegisterRequest`, `LoginResponse` pattern).

### 4. No framework types in interface signatures

Don't leak `*gorm.DB`, `echo.Context`, `*http.Request` into interfaces — it ties the mock to the framework and makes mocking painful. Use `context.Context` + domain/DTO types:

```go
// GOOD — context + DTO; mock is framework-free
Register(ctx context.Context, req RegisterRequest) (*model.User, error)

// BAD — couples mock to Echo/GORM, untestable without the framework
Register(c echo.Context) error
Save(db *gorm.DB) error
```

### 5. Context as the first parameter

Consistent `ctx context.Context` first param → uniform mock signatures, `mock.Anything` for the context arg. (See [golang-context-concurrency.md](golang-context-concurrency.md).)

### 6. Define interfaces at the consumer (bxcodec pattern)

The consuming layer owns the interface; the implementer satisfies it. Mockery mocks the interface **the consumer defines**, and that interface lives in a package listed in `.mockery.yml`:

```
service/auth/service.go  → defines Storage interface (consumer)
storage/user/            → *repo implements it
mocks/storage/           → mockery generates MockStorage from the consumer's interface
```

For full placement rules, see the project's clean-architecture rules (or this skill's `../folder-structure.md`) → *Interface Placement Rule*.

### 7. Keep interfaces focused, don't over-interface

- ✅ Interface across architecture layers that you need to mock in tests (service ↔ storage)
- ✅ Narrow interfaces per consumer when feasible (ISP) → smaller mock surface
- ⚠️ The project consolidates per layer (`Storage`, `Service`) — acceptable; just regenerate when methods change
- ❌ Don't interface internal single-implementation code (YAGNI) — adds ceremony with no mock benefit

---

## Toolchain

**Install / update (mise):**
```bash
mise install                  # all tools from mise.toml
mise exec -- mockery version
```

**Install in CI (no mise):**
```bash
go install github.com/vektra/mockery/v3@latest
```

> Pin `mockery/v3` consistently between local (mise) and CI — v2 ≠ v3 output.

---

## Configuration (`.mockery.yml`)

**Project uses `all: true` with layer-based output directories:**
```yaml
# Mockery v3 — mocks ALL interfaces in configured packages
all: true
dir: "internal/mocks"
filename: "mock_{{.InterfaceName}}.go"
force-file-write: true
formatter: goimports
log-level: info
pkgname: "mocks"                       # source-derived to avoid import cycles
structname: "Mock{{.InterfaceName}}"

packages:
  # Storage layer — all storage interfaces
  golang-sample/internal/storage/user:
    config:
      dir: "internal/mocks/storage"

  # Service layer — all service interfaces
  golang-sample/internal/service/auth:
    config:
      dir: "internal/mocks/service"
```

**Config rules:**
- ✅ Add new packages under `packages:` following the layer pattern
- ✅ Group mocks by architecture layer (`mocks/storage/`, `mocks/service/`, `mocks/handler/`)
- ✅ Keep `all: true` — do NOT enumerate interfaces per package
- ✅ Keep `pkgname: "mocks"` and `structname: "Mock{{.InterfaceName}}"`
- ❌ Never mix layers in one mock directory

**Adding a new mockable package:**
```yaml
packages:
  golang-sample/internal/service/email:
    config:
      dir: "internal/mocks/service"   # co-locate by layer
```
Then run `mise exec -- mockery`.

---

## When to Generate

**Regenerate whenever you:** add an interface method, change a signature, add a new interface to a configured package, or rename an interface/package.

```bash
mise exec -- mockery              # regenerate all from .mockery.yml
git diff --exit-code              # CI fails if committed mocks are stale
```

---

## Mock Organization

**Layer-based layout mirrors clean architecture:**
```text
internal/mocks/
├── storage/
│   └── mock_Storage.go           # from internal/storage/user
├── service/
│   └── mock_Service.go           # from internal/service/auth
└── handler/                      # add when mocking controllers
```

- ✅ One file per interface: `mock_{{.InterfaceName}}.go`
- ✅ Import with a descriptive alias: `storageMocks "…/mocks/storage"`
- ❌ Never scatter mock files inside source packages

---

## Consuming Mocks in Tests

**Prefer the type-safe `EXPECT()` expecter API** over string-based `mock.On`:
```go
package auth

import (
    "testing"

    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    storageMocks "golang-sample/internal/mocks/storage"
)

func newTestService(t *testing.T, storage *storageMocks.MockStorage) Service {
    t.Helper()
    return NewAuthService(log, storage, "test-secret", testJWTExpiration)
}

func TestService_Register(t *testing.T) {
    storage := storageMocks.NewMockStorage(t)

    // GOOD — type-safe, refactor-friendly
    storage.EXPECT().
        CheckUniqueness(mock.Anything, "newuser", "new@example.com").
        Return(false, false, nil)

    svc := newTestService(t, storage)
    // ... call svc, assert results
}
```

**Constructor auto-asserts expectations:**
- `storageMocks.NewMockStorage(t)` registers `t.Cleanup` → `AssertExpectations`
- All expected calls verified automatically — no manual `defer ...AssertExpectations`

```go
// GOOD — compile-checked
storage.EXPECT().FindUserByUsername(mock.Anything, "alice").Return(user, nil)

// AVOID — string-based, breaks silently on rename
storage.On("FindUserByUsename", ...).Return(...) // typo → nil return
```

---

## Argument Matching & Returns

```go
// mock.Anything for plumbing args (context, timestamps); concrete values for business inputs
storage.EXPECT().CreateUser(mock.Anything, expectedUser).Return(nil)
storage.EXPECT().FindUserByUsername(mock.Anything, "alice").Return(user, nil)

// Multiple returns — match method signature exactly
storage.EXPECT().CheckUniqueness(...).Return(false, true, nil) // (userExists, emailExists, error)

// Dynamic return via RunAndReturn (return depends on inputs)
storage.EXPECT().CreateUser(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, u *model.User) error {
    u.ID = 99
    return nil
})
```

- ✅ `mock.Anything` for `context.Context` (plumbing, not under test)
- ✅ Concrete values for inputs that affect the assertion
- ❌ Never `mock.Anything` on every arg — defeats the assertion

---

## Table-Driven Tests with Mocks

**Inject per-case mock setup via a `setupMock` closure; create a fresh mock per subtest:**
```go
tests := []struct {
    name      string
    setupMock func(*storageMocks.MockStorage)
    wantErr   error
}{
    {
        name: "username already exists",
        setupMock: func(m *storageMocks.MockStorage) {
            m.EXPECT().CheckUniqueness(mock.Anything, "existinguser", "new@example.com").
                Return(true, false, nil)
        },
        wantErr: ErrConflict,
    },
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        storage := storageMocks.NewMockStorage(t) // fresh mock per case
        tt.setupMock(storage)
        svc := newTestService(t, storage)

        err := svc.Register(context.Background(), ...)
        require.ErrorIs(t, err, tt.wantErr)
    })
}
```

- ✅ `t.Helper()` in setup helpers
- ❌ Never share one mock across unrelated cases (expectation bleed)

---

## Best Practices & Pitfalls

**✅ DO:**
- Mock only the layer directly below the unit under test
- Use `RunAndReturn` when return values depend on inputs
- Verify committed mocks match interfaces in CI

**❌ DON'T:**
- Edit generated `mock_*.go` by hand — regeneration overwrites it
- Mock value types/structs/functions instead of interfaces
- Re-mock an interface already covered by an existing package config
- Commit a stale mock (interface changed, mock not regenerated)
- Skip `EXPECT()` setup and rely on mock zero behavior (silent panics)

**Pitfalls:**
```go
// BAD — editing generated mock directly (overwritten on next run)
func (_m *MockStorage) FindUserByUsername(...) { /* custom */ }

// BAD — mocking a concrete struct (no interface → no mock generated)
type userService struct{}

// BAD — stale committed mock: CI regenerates fresh, committed mock_*.go drifts

// BAD — sharing a mock across cases without reset → duplicate expectations
```

---

## CI Integration

**Matches `.github/workflows/test-sample.yml`:**
```yaml
- name: Install mockery
  run: go install github.com/vektra/mockery/v3@latest

- name: Generate mocks
  run: mockery                      # uses .mockery.yml

- name: Run tests
  run: go test -v -race -coverprofile=coverage.out ./...
```

CI regenerates mocks before tests — fresh mocks guarantee sync with interfaces.

---

## Quick Reference

```bash
mise exec -- mockery            # regenerate all mocks from .mockery.yml
mise exec -- mockery --version  # check version
git diff                        # review generated changes before commit
```

```go
m := storageMocks.NewMockStorage(t)                  // construct (auto-assert on cleanup)
m.EXPECT().Method(ctx, arg).Return(val, nil)         // type-safe expect
m.EXPECT().Method(mock.Anything, arg).Return(...)    // loose context match
m.EXPECT().Method(...).RunAndReturn(func(...) T {…}) // dynamic return
```

| Task | Pattern |
|------|---------|
| Add a package to mock | add entry under `packages:` with `dir: "internal/mocks/<layer>"` |
| Construct mock | `storageMocks.NewMockStorage(t)` |
| Expect call | `m.EXPECT().Method(...).Return(...)` |
| Loose arg | `mock.Anything` |
| Verify expected calls | automatic via `NewMockX(t)` cleanup |

---

## References

- [vektra/mockery](https://github.com/vektra/mockery) — mock generator
- [testify/mock](https://pkg.go.dev/github.com/stretchr/testify/mock) — underlying framework
- [mockery v3 config docs](https://vektra.github.io/mockery/latest/configuration/)
- Project config: `.mockery.yml` at the sample-app root
- Testing rules: [golang-testing.md](golang-testing.md)
