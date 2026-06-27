# Pre-Migration State Documentation

**Date**: 2026-06-28
**Branch**: feat/monorepo-migration
**Repository**: golang-sample

## Current Module Information

**Module Path**: golang-sample
**Go Version**: 1.25.0
**Repository**: github.com/haipham22/golang-sample

## Current Dependencies

```
require (
	github.com/getsentry/sentry-go v0.43.0
	github.com/go-playground/validator/v10 v10.30.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/wire v0.7.0
	github.com/haipham22/govern v0.0.0
	github.com/labstack/echo/v4 v4.15.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.11.1
	github.com/swaggo/swag v1.16.6
	go.uber.org/automaxprocs v1.6.0
	go.uber.org/zap v1.27.1
	golang.org/x/crypto v0.48.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
)
```

## Current Import Paths

**Govern packages already imported from github.com/haipham22/govern**:
- github.com/haipham22/govern/http
- github.com/haipham22/govern/errors
- github.com/haipham22/govern/http/echo
- github.com/haipham22/govern/http/middleware
- github.com/haipham22/govern/graceful

**Files using govern imports**:
- internal/handler/rest/wire_gen.go
- internal/handler/rest/handler.go
- internal/handler/rest/wire.go
- internal/handler/rest/controllers/auth/auth.go
- internal/validator/validator.go
- internal/service/auth/impl.go
- cmd/serverd.go

## Current Directory Structure

```
golang-sample/
├── cmd/
│   ├── serverd.go
│   ├── root.go
│   ├── grpcd.go
│   └── workerd.go
├── internal/
│   ├── handler/
│   ├── service/
│   ├── storage/
│   ├── model/
│   ├── orm/
│   ├── schemas/
│   └── validator/
├── orm/
├── schemas/
├── validator/
├── scripts/
├── .github/
├── docs/
└── plans/
```

## Git Status

- **Current branch**: feat/monorepo-migration
- **Base branch**: main
- **Uncommitted changes**: None (clean working directory)
- **Untracked files**: docs/, plans/ (documentation only)

## Govern Repository State

- **Location**: ../govern/
- **Branch**: main
- **Latest commit**: (extracted from govern repo)
- **Status**: Clean working tree

**Govern Packages Available**:
- http/ (with echo/, jwt/, middleware/ subdirectories)
- database/ (with postgres/, redis/ subdirectories)
- config/
- errors/
- log/
- graceful/
- retry/
- cron/
- mq/ (with asynq/ subdirectory)
- metrics/
- healthcheck/

## Migration Readiness Assessment

✅ **Ready for Phase 02**: Govern packages repository accessible
✅ **Import Paths**: Already using github.com/haipham22/govern (no changes needed)
✅ **Go Version**: 1.25.0 matches govern requirements
✅ **Git History**: Clean working directory, ready for feature branch work
✅ **Feature Branch**: Created (feat/monorepo-migration)

## Next Steps

1. Phase 01 Complete: Directory structure setup and documentation
2. Phase 02: Merge govern packages using git subtree/merge
3. Phase 03: Move sample application to examples/golang-sample/
