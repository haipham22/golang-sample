# Common Go Rules

**Best practices for Go development including types, concurrency, error handling, idioms, and testing.**

---

## Overview

This guide is split into focused topic files for better maintainability:

| Topic | File | Lines |
|-------|------|-------|
| **Types & Values** | [golang-types-values.md](golang-types-values.md) | ~200 |
| **Context & Concurrency** | [golang-context-concurrency.md](golang-context-concurrency.md) | ~200 |
| **Error Handling** | [golang-error-handling.md](golang-error-handling.md) | ~150 |
| **Validation** | [golang-validator.md](golang-validator.md) | ~150 |
| **Database & GORM** | [golang-database.md](golang-database.md) | ~200 |
| **Swagger/OpenAPI** | [golang-swagger.md](golang-swagger.md) | ~200 |
| **Mise Toolchain** | [mise.md](mise.md) | ~200 |
| **Common Idioms** | [golang-idioms.md](golang-idioms.md) | ~150 |
| **Testing** | [golang-testing.md](golang-testing.md) | ~200 |
| **Mockery** | [mockery.md](mockery.md) | ~380 |
| **Dependency Injection** | [dependency-injection.md](dependency-injection.md) | ~340 |
| **Cobra CLI** | [cobra-cli.md](cobra-cli.md) | ~360 |
| **Graceful** | [graceful.md](graceful.md) | ~350 |
| **Connection DSN** | [connection-dsn.md](connection-dsn.md) | ~270 |

---

## Quick Reference

### Pointers vs Values

| Scenario | Use |
|----------|-----|
| **Modify data** | `*Type` (pointer) |
| **Read-only** | `Type` (value) |
| **Large structs** | `*Type` |
| **Small structs** | `Type` |
| **Optional values** | `*Type` (can be nil) |

### Context Rules

- ✅ **First parameter** in I/O functions
- ✅ **Pass through call chain** (handler → repository)
- ✅ **Check `ctx.Err()`** in long-running loops
- ❌ **NEVER store** context in struct

### Error Handling

- ✅ **Wrap errors** with context: `fmt.Errorf("failed: %w", err)`
- ✅ **Check errors** immediately
- ✅ **Use `errors.Is()`** for comparison
- ❌ **NEVER ignore** errors: `result, _ = func()`

### Concurrency

- ✅ **Use `errgroup`** for coordinated goroutines
- ✅ **Close channels** from sender side
- ✅ **Iterate with `range`** over channels
- ❌ **NEVER send** to closed channel

---

## Key Patterns

### Function Parameter Order
```
context → dependencies → config → input
```

### Constructor Pattern
```go
func NewUser(name, email string) *User {
    return &User{Name: name, Email: email}
}
```

### Error Handling
```go
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) error {
    user, err := s.repo.FindByEmail(ctx, req.Email)
    if err != nil {
        return fmt.Errorf("failed to find user: %w", err)
    }
    return s.repo.Create(ctx, &user)
}
```

---

## Testing Commands

```bash
# Run tests
go test ./...

# Race detector
go test -race ./...

# Coverage
go test -cover ./...

# Benchmarks
go test -bench=. -benchmem
```

---

## Deep Dive Topics

For detailed rules and examples, see:

1. **[Types and Values](golang-types-values.md)** - Pointers, generics, var/const, return values
2. **[Context and Concurrency](golang-context-concurrency.md)** - Context usage, goroutines, channels
3. **[Error Handling](golang-error-handling.md)** - Error wrapping, custom errors, logging
4. **[Common Idioms](golang-idioms.md)** - Interface satisfaction, defer, zero values
5. **[Testing](golang-testing.md)** - Table-driven tests, mocking, benchmarks

---

## Development Workflow

### Before Making Changes
```bash
mise install  # Install/update tools
```

### After Making Changes
```bash
mise exec -- goimports -w .              # Format code
mise exec -- golangci-lint run           # Lint
mise exec -- go test ./...               # Test
mise exec -- go build ./...              # Verify build
```

### Before Commit
```bash
mise exec -- go test -race ./...         # Check for races
mise exec -- go test -cover ./...        # Check coverage
```

---

## Additional Resources

### Official Documentation
- [Go 1 Compatibility Promise](https://go.dev/doc/go1compat)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Project Documentation
- [CLAUDE.md](../../CLAUDE.md) - Development rules
- [Clean Architecture Rules](clean-architecture.md) - Project structure

---

**File Organization:** Split into focused topic files for maintainability. Each file under 250 lines per development guidelines.
