# Go Development Rules

**Comprehensive rules for Go development in golang-sample project.**

---

## Overview

This directory contains focused, single-responsibility rule files for Go development. Each file covers a specific topic and is kept under 250 lines for maintainability.

**When rules change:** Update this file (`README.md`), NOT the project's main `CLAUDE.md`. This separation keeps project documentation stable while allowing rules to evolve independently.

---

## Rule Files

| Topic | File | Focus | Lines |
|-------|------|-------|-------|
| **Go Rules Hub** | [golang-common-rules.md](golang-common-rules.md) | Quick reference, overview of all rules | ~150 |
| **Types & Values** | [golang-types-values.md](golang-types-values.md) | Pointers, generics, var/const, return values | ~200 |
| **Context & Concurrency** | [golang-context-concurrency.md](golang-context-concurrency.md) | Context usage, goroutines, channels | ~200 |
| **Error Handling** | [golang-error-handling.md](golang-error-handling.md) | Error wrapping, custom errors, logging | ~150 |
| **Validation** | [golang-validator.md](golang-validator.md) | go-playground/validator patterns | ~150 |
| **Database** | [golang-database.md](golang-database.md) | GORM, transactions, query optimization | ~200 |
| **Swagger/OpenAPI** | [golang-swagger.md](golang-swagger.md) | API documentation with swaggo/swag | ~200 |
| **Mise Toolchain** | [mise.md](mise.md) | Toolchain management, version pinning | ~200 |
| **Common Idioms** | [golang-idioms.md](golang-idioms.md) | Interface satisfaction, defer, zero values | ~150 |
| **Testing** | [golang-testing.md](golang-testing.md) | Table-driven tests, mocking, benchmarks | ~200 |
| **Clean Architecture** | [clean-architecture.md](clean-architecture.md) | Project structure, layering, dependencies | ~200 |
| **Dependency Injection** | [dependency-injection.md](dependency-injection.md) | Manual DI, composition root, Wire migration | ~340 |
| **Cobra CLI** | [cobra-cli.md](cobra-cli.md) | `main.go` + `cmd/` structure, RunE, flags | ~360 |
| **Coding Standards** | [golang-coding-standards.md](golang-coding-standards.md) | Naming, formatting, file organization | ~250 |
| **Web Framework** | [web-framework-rules.md](web-framework-rules.md) | Echo handlers, middleware, validation | ~200 |
| **Database Rules** | [database-rules.md](database-rules.md) | GORM patterns, error envelope, PostgreSQL | ~200 |
| **Infrastructure** | [infrastructure-rules.md](infrastructure-rules.md) | Zap logging, environment config | ~200 |
| **Docker** | [docker.md](docker.md) | Docker Compose, multi-stage builds, containers | ~200 |
| **Mockery** | [mockery.md](mockery.md) | Mock generation (v3), config, interface design | ~380 |

---

## Quick Start

**For new developers:**
1. Read [golang-common-rules.md](golang-common-rules.md) for overview
2. Read [clean-architecture.md](clean-architecture.md) for project structure
3. Reference specific files as needed during development

**For specific tasks:**
- **Database operations:** [golang-database.md](golang-database.md), [database-rules.md](database-rules.md)
- **HTTP handlers:** [web-framework-rules.md](web-framework-rules.md)
- **Validation:** [golang-validator.md](golang-validator.md)
- **Testing:** [golang-testing.md](golang-testing.md)
- **Mock generation:** [mockery.md](mockery.md)
- **Dependency injection:** [dependency-injection.md](dependency-injection.md)
- **CLI / commands:** [cobra-cli.md](cobra-cli.md)
- **API docs:** [golang-swagger.md](golang-swagger.md)
- **Docker containers:** [docker.md](docker.md)

---

## Rules Organization

**By layer:**
```
Handler Layer     → web-framework-rules.md
Service Layer     → golang-validator.md, golang-error-handling.md
Repository Layer  → golang-database.md, database-rules.md
Domain Layer      → golang-types-values.md, golang-idioms.md
```

**By concern:**
```
Code Quality      → golang-coding-standards.md, golang-common-rules.md
Concurrency       → golang-context-concurrency.md
Data Validation   → golang-validator.md
Error Handling    → golang-error-handling.md, infrastructure-rules.md
Testing           → golang-testing.md
Tooling           → mise.md, golang-swagger.md, docker.md
```

---

## Key Principles

**All rules follow:**
- **YAGNI** - You Aren't Gonna Need It
- **KISS** - Keep It Simple, Stupid
- **DRY** - Don't Repeat Yourself

**File organization:**
- Maximum 250 lines per file
- Single responsibility per file
- Self-documenting filenames (long is OK)
- Regular refactoring to maintain size limits

---

## Common Patterns

**Function parameter order:**
```
context → dependencies → config → input
```

**Error handling:**
```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

**Database operations:**
```go
err := db.WithContext(ctx).Where("id = ?", id).First(&user).Error
```

**Validation:**
```go
if err := validator.Validate(req); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}
```

---

## When to Update This File

**Update this README when:**
- ✅ New rule file is created
- ✅ Rule file is renamed or moved
- ✅ Rule file is deleted
- ✅ Major reorganization of rules

**Do NOT update CLAUDE.md for rule changes.** Project documentation stays stable.

---

## Related Documentation

**Project-level:**
- [../../CLAUDE.md](../CLAUDE.md) - Project setup and development workflow
- [../../README.md](../../README.md) - Project overview and quick start

**External resources:**
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide)

---

**Last updated:** 2025-01-28
