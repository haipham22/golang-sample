---
title: "Interactive Project Generator"
description: "Go-based interactive CLI tool to generate new projects from templates"
status: backlog
priority: P2
effort: 8h
tags: [generator, scaffolding, cli, templates]
created: 2026-06-27
---

# Interactive Project Generator - Backlog

## Status

**Current Status**: BACKLOG - Deferred for detailed design phase

**Reason**: User wants to go deep into implementation details separately. This feature is NOT blocking the main monorepo restructuring.

---

## Overview

Create an interactive Go-based CLI tool that generates new projects from customizable templates, similar to `npm create` or `django-admin startproject`.

**Target Output**: 
```bash
$ generator
? Project name: my-service
? Select template:
  ○ basic (Minimal HTTP server)
  ○ fullstack (HTTP + DB + Auth)
  ○ microservice (Microservice structure)
? Features:
  ✓ PostgreSQL
  ✓ Redis
  ✓ JWT Auth
  ✓ Metrics
? Output directory: ./my-service

✅ Project generated successfully!
$ cd my-service
$ go run cmd/serverd.go
```

---

## Features

### Template System
- **base**: Skeleton structure with clean architecture
- **basic**: Minimal HTTP server with Echo framework
- **fullstack**: Complete API with PostgreSQL, Redis, JWT
- **microservice**: Microservice-ready structure

### Interactive CLI
- Prompt-based project creation (using `promptui`)
- Feature selection UI
- Template preview
- Dependency validation

### Core Functionality
- Template rendering engine (Go templates)
- Govern package integration (auto-import `github.com/haipham22/govern`)
- Module path configuration
- CI/CD file generation
- README generation

---

## Architecture

**Generator Structure**:
```
scripts/generate-project/
├── main.go                 # CLI entry point
├── cli/                    # Interactive prompts
│   └── prompts.go         # User interface
├── generator/              # Generation logic
│   ├── template.go        # Template engine
│   ├── config.go          # Configuration
│   └── render.go          # File rendering
├── templates/              # Template definitions
│   ├── base/              # Base template
│   ├── basic/             # Basic HTTP server
│   ├── fullstack/         # Full application
│   └── microservice/      # Microservice
└── pkg/                    # Shared packages
    ├── files/             # File operations
    └── modules/           # Module handling
```

**Generated Project Structure** (basic template):
```
my-service/
├── cmd/
│   └── serverd.go         # Main entry point
├── internal/
│   ├── handler/           # HTTP layer
│   ├── service/           # Business logic
│   ├── storage/           # Data access
│   └── model/             # Domain models
├── go.mod                 # Module: github.com/user/my-service
├── go.sum
├── Makefile               # Build targets
├── .env.example           # Environment variables
├── README.md              # Project documentation
└── .github/
    └── workflows/          # CI/CD
```

---

## Implementation Plan (When Activated)

### Phase 1: Core Generator (4h)
1. Setup CLI structure with `promptui`
2. Implement interactive prompts
3. Create template rendering engine
4. Implement file generation logic

### Phase 2: Template System (3h)
1. Create base template structure
2. Implement basic template (Echo HTTP server)
3. Add feature selection system
4. Implement template validation

### Phase 3: Integration & Testing (1h)
1. Integrate with govern packages
2. Test all templates
3. Validate generated projects compile
4. Documentation

---

## Success Criteria

**Definition of Done**:
- Interactive CLI works smoothly
- All 4 templates generate valid projects
- Generated projects compile without errors
- All generated projects import govern successfully
- Full documentation

**Validation Methods**:
```bash
# Test generator
cd scripts/generate-project
go build

# Generate test project
./generate-project
# Answer prompts interactively

# Verify generated project
cd <output-dir>
go test ./...
go build ./...
```

---

## Integration with Main Plan

**When to Implement**: AFTER monorepo restructuring + wire removal complete

**Dependencies**:
- Govern library must be at `github.com/haipham22/govern`
- Templates must reference correct import paths
- CLI must run from `scripts/generate-project/`

**Output Directory**: Generator will create projects in user-specified directories (not in govern repository)

---

## Open Questions

1. **Template Format**: Use Go templates or YAML-based config?
2. **CLI Framework**: Use `promptui` or build custom UI?
3. **Template Storage**: Keep in repository or external registry?
4. **Feature Selection**: How complex should the feature selection be?
5. **Govern Integration**: Auto-add all govern packages or let user choose?

---

## Backlog Priority

**Priority**: P2 (Nice to have)

**Blocking**: NO - Does not block main plan implementation

**Estimated Effort**: 8 hours

**Recommendation**: Design deep-dive session before implementation to answer open questions and finalize architecture.

---

**Status**: Ready for detailed design phase when user decides to prioritize.
