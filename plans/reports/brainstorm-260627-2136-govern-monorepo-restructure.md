# Govern Monorepo Restructure - Brainstorm Report

**Date**: 2026-06-27  
**Type**: Architecture & Restructuring  
**Status**: Design Complete

---

## Problem Statement

**Current State:**
- `golang-sample`: Clean architecture example app using govern as external dependency
- `govern/` (separate repo): Production-ready Go library with 15+ packages
- Users must import govern externally to use it

**Desired State:**
- Single repository containing govern library + sample implementation
- Interactive generator to create new projects from templates
- govern packages available as internal/external library
- samples/sample-app demonstrates best practices

**Constraints:**
- Must preserve git history where possible
- Cannot break existing imports (if any)
- CI/CD must continue working
- Documentation must remain discoverable

---

## Final Repository Structure

```
govern/                           (REPOSITORY RENAMED: github.com/haipham22/govern)
├── http/                         (From govern/ - Module: github.com/haipham22/govern/http)
├── http/echo/                    (From govern/http/echo/ - govern module)
├── http/jwt/                     (From govern/http/jwt/ - govern module)
├── http/middleware/              (From govern/http/middleware/ - govern module)
├── database/                     (From govern/database/ - govern module)
├── database/postgres/            (From govern/database/postgres/ - govern module)
├── database/redis/               (From govern/database/redis/ - govern module)
├── config/                       (From govern/config/ - govern module)
├── errors/                       (From govern/errors/ - govern module)
├── log/                          (From govern/log/ - govern module)
├── graceful/                     (From govern/graceful/ - govern module)
├── retry/                        (From govern/retry/ - govern module)
├── cron/                         (From govern/cron/ - govern module)
├── mq/                           (From govern/mq/ - govern module)
├── mq/asynq/                     (From govern/mq/asynq/ - govern module)
├── metrics/                      (From govern/metrics/ - govern module)
├── healthcheck/                  (From govern/healthcheck/ - govern module)
├── go.mod                        (Root go.mod: module github.com/haipham22/govern)
├── go.sum
├── Makefile                      (Library package targets)
├── README.md                     (Govern library documentation)
├── .github/
│   └── workflows/
│       ├── test.yml              (Tests govern packages)
│       └── push.yml              (Builds/publishes govern library)
├── docs/                         (Govern library docs)
│   ├── quickstart.md
│   ├── packages/
│   └── contributing.md
├── scripts/
│   └── generate-project/         (NEW - Interactive generator in Go)
│       ├── main.go
│       └── go.mod
├── templates/                    (NEW - Project templates)
│   ├── base/                    (Common files for all templates)
│   ├── basic/                   (Minimal template: Echo + logging + config + graceful shutdown)
│   ├── fullstack/               (Complete API template)
│   └── microservice/            (Microservice template)
└── samples/
    └── golang-sample/            (RENAMED from sample-app - shows legacy naming preserved)
        ├── cmd/
        ├── internal/
        ├── orm/
        ├── schemas/
        ├── validator/
        ├── go.mod                (Sample app module: github.com/haipham22/govern/samples/golang-sample)
        ├── go.sum
        ├── Makefile              (Sample app targets)
        ├── .env.example
        ├── .github/
        │   └── workflows/        (Sample app CI/CD)
        └── README.md             (Sample app documentation)
```

**Repository & Module Organization** (FIXED):
- **Repository**: `github.com/haipham22/govern` (renamed from golang-sample)
- **Govern packages**: `github.com/haipham22/govern` (root module, published as library)
- **Sample app**: `github.com/haipham22/govern/samples/golang-sample` (subdirectory module using workspaces)
- **Generator**: `github.com/haipham22/govern/scripts/generate-project` (CLI tool)
- **Generated projects**: Import `github.com/haipham22/govern` as external dependency

**Go Workspace** (go.work):
```go
// go.work file at repository root
go 1.25

use (
    .           // govern module (root)
    ./samples/golang-sample  // sample app module
)
```

---

## Restructuring Approach: Big Bang Migration

**Strategy**: Single atomic migration with comprehensive commit

**Why**: Cleanest git history, single review point, no intermediate broken states

**Migration Steps** (UPDATED):

### Phase 1: Repository Preparation
1. Create feature branch: `git checkout -b feat/monorepo-migration`
2. Document current import paths and dependencies
3. Create `samples/` and `templates/` directories
4. **Prepare repository rename**: Plan to rename from `golang-sample` to `govern` (GitHub settings)

### Phase 2: Merge govern packages (WITH GIT HISTORY)
5. **Use git fast-export/import to preserve history**:
   ```bash
   cd ../govern
   git fast-export --all | cd ../golang-sample && git fast-import
   ```
6. Verify govern packages at repository root with history intact
7. Verify all govern packages compile: `go build ./...`
8. Run `go mod tidy` to consolidate dependencies

### Phase 3: Move sample app
9. Move `cmd/`, `internal/`, `orm/`, `schemas/`, `validator/` to `samples/golang-sample/`
10. Create `samples/golang-sample/go.mod` with module `github.com/haipham22/govern/samples/golang-sample`
11. **Create go.work workspace file** at repository root
12. Update sample app imports (already using `github.com/haipham22/govern`)
13. Move `scripts/generate-swagger.sh` to `samples/golang-sample/scripts/`
14. Move sample-specific Makefile targets to `samples/golang-sample/Makefile`
15. Verify sample app compiles: `cd samples/golang-sample && go build ./...`

### Phase 4: Update root configuration
16. Create root `Makefile` with govern library targets
17. **Create .github/workflows/test.yml** for govern packages only
18. **Create samples/golang-sample/.github/workflows/test.yml** for sample app
19. Update `.github/workflows/push.yml` (build/publish govern library)
20. Create root `README.md` (govern library documentation)

### Phase 5: Interactive generator
21. Create `scripts/generate-project/` directory for Go generator
22. Create `scripts/generate-project/go.mod` for generator CLI tool
23. Implement generator with promptui and text/template
24. Create templates in `templates/basic/`, `templates/fullstack/`, `templates/microservice/`
25. Add basic template: Echo + logging + config + graceful shutdown
26. Test generator creates valid projects importing `github.com/haipham22/govern`

### Phase 6: Documentation
27. Move `../govern/QUICKSTART.md` → `docs/quickstart.md`
28. Move `../govern/DEVELOPMENT.md` → `docs/contributing.md`
29. Adapt govern package docs to `docs/packages/`
30. Create `docs/samples/golang-sample-guide.md`
31. Update `CLAUDE.md` for new repository structure
32. Update sample app README with new structure notes

### Phase 7: Validation
33. Test with go.work: `go work sync && go test ./...`
34. Test CI/CD workflows locally
35. Test generator creates working projects
36. Verify generated projects compile and run
37. Final validation of all documentation

### Phase 8: Repository Rename & Merge
38. **Rename repository on GitHub**: golang-sample → govern
39. Update local remote: `git remote set-url origin git@github.com:haipham22/govern.git`
40. Create comprehensive commit with all changes
41. Open PR for review
42. Merge to main branch
43. Post-merge: Tag govern library version v0.1.0

---

## Interactive Generator Design

**Tool**: `scripts/generate-project` (Go binary)

**Why Go instead of bash**:
- Better error handling and validation
- Can use Go's text/template package for powerful templating
- Type-safe template processing
- Can be distributed as standalone CLI tool
- Easier to maintain and extend

**Installation**:
```bash
go build -o bin/generate-project ./scripts/generate-project
# Or alias for global use
go install github.com/haipham22/golang-sample/scripts/generate-project@latest
```

### User Flow

```bash
./scripts/generate-project.sh

> Welcome to Govern Project Generator!
>
> Project name: my-api
> Destination path: [../my-api] (current parent directory)
>
> Select template:
>   1) basic       - Minimal HTTP server with Echo
>   2) fullstack   - Complete API with PostgreSQL, auth, metrics
>   3) microservice - gRPC + message queue + tracing
> [1-3]: 2
>
> Features (multi-select, space to toggle, enter to confirm):
>   [*] PostgreSQL (GORM)
>   [*] Redis (caching)
>   [ ] JWT authentication
>   [*] Prometheus metrics
>   [*] Health checks
>   [ ] Asynq task queue
>   [ ] Cron scheduler
>
> Generating project...
> ✓ Created directory structure
> ✓ Generated go.mod
> ✓ Generated main.go
> ✓ Generated Dockerfile
> ✓ Generated .env.example
> ✓ Generated Makefile
> ✓ Generated README.md
> ✓ Configured pre-commit hooks
>
> Next steps:
>   cd ../my-api
>   mise install
>   make run
>
> Project generated successfully!
```

### Template System

**Template Structure**:
```
templates/
├── base/                     (Common files for all templates)
│   ├── .gitignore
│   ├── .pre-commit-config.yaml
│   ├── Dockerfile
│   └── mise.toml
├── basic/                    (Minimum: Echo HTTP + logging + config + graceful shutdown)
│   ├── cmd/
│   │   └── root.go
│   ├── internal/
│   │   ├── config/          (Config from govern/config)
│   │   ├── log/             (Logging from govern/log)
│   │   └── handler/
│   │       └── rest/
│   │           └── server.go (HTTP from govern/http + graceful shutdown)
│   ├── go.mod.template
│   └── main.template
├── fullstack/
│   ├── cmd/
│   ├── internal/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── storage/
│   │   └── model/
│   ├── orm/
│   └── migrations/
└── microservice/
    ├── cmd/
    ├── proto/
    └── internal/
```

**Basic Template Features** (minimum included):
- ✓ Echo HTTP server (govern/http)
- ✓ Graceful shutdown (govern/graceful)
- ✓ Structured logging (govern/log with zap)
- ✓ Config management (govern/config with YAML/.env support)

**Template Variables**:
- `{{PROJECT_NAME}}` - Project name (kebab-case)
- `{{MODULE_PATH}}` - Go module path
- `{{PORT}}` - Default port
- `{{FEATURES}}` - Selected features
- `{{YEAR}}` - Copyright year

### Implementation

**Key Functions** (Go):
```go
// Interactive prompts with validation
promptProjectName()
promptTemplate()
promptFeatures()

// Template processing
renderTemplate()  // Use text/template
copyBaseFiles()
applyFeatures()

// Post-generation
initializeGit()
installTools()
generateMocks()
```

**Dependencies**:
- Go 1.25+
- `github.com/manifoldco/promptui` - Interactive prompts
- `text/template` - Built-in Go template engine
- `github.com/spf13/cobra` - CLI framework (optional, for structured commands)

**CLI Usage**:
```bash
# Build and run
go run ./scripts/generate-project

# Or install globally
go install github.com/haipham22/golang-sample/scripts/generate-project@latest
generate-project
```

---

## Implementation Considerations

### Import Path Management (FIXED)

**External Consumers**: None (confirmed - no external projects import govern)

**Module Organization** (FIXED - using Go workspaces):
- **Repository**: `github.com/haipham22/govern` (renamed from golang-sample)
- **Root module**: `github.com/haipham22/govern` (govern packages)
- **Sample app**: `github.com/haipham22/govern/samples/golang-sample`
- **Development**: Uses `go.work` workspace file for local development
- **Production**: Sample app uses published `github.com/haipham22/govern` dependency

**go.work file** (repository root):
```go
go 1.25

use (
    .                          // govern module (root)
    ./samples/golang-sample   // sample app module
)
```

**Sample App go.mod**:
```go
module github.com/haipham22/govern/samples/golang-sample

go 1.25

require (
    github.com/haipham22/govern v0.1.0  // Published version
    // ... other dependencies
)

// In development with go.work, this uses local govern module
// In production, this uses published govern version from GitHub
```

**Import Migration**:
- Current imports: `github.com/haipham22/govern/http` (already correct)
- After migration: Same imports, no changes needed
- Sample app imports: Continues using `github.com/haipham22/govern` packages

### CI/CD Updates (FIXED)

**Repository Structure**: Two separate CI/CD configurations

**Root .github/workflows/** (govern library):
- `test.yml` - Tests govern packages: `go test ./http/... ./database/... ./config/...`
- `push.yml` - Builds and publishes govern Docker image (library usage examples)
- `release.yml` - Tags and publishes govern library releases

**samples/golang-sample/.github/workflows/** (sample app):
- `test.yml` - Tests sample app with go.work workspace
- `push.yml` - Builds and publishes sample app Docker image
- `deploy.yml` - Sample app deployment (if applicable)

**Coverage Strategy**:
- Govern: Separate coverage report for library packages
- Sample: Separate coverage report for sample app
- No combined coverage (each module tracks its own)

### Documentation Migration

**From `../govern/`**:
- `QUICKSTART.md` → `docs/quickstart.md`
- `DEVELOPMENT.md` → `docs/contributing.md`
- Package docs → `docs/packages/`
- Godoc examples → Add to package `doc.go` files

**From current `golang-sample/`**:
- `CLAUDE.md` → Keep at root (project instructions)
- `docs/` content → Merge and reorganize
- Create `docs/samples/sample-app-guide.md`

---

## Success Criteria

**Functional Requirements**:
- [ ] All govern packages compile without errors
- [ ] All samples compile and tests pass
- [ ] Interactive generator creates working projects
- [ ] All templates produce valid Go code
- [ ] CI/CD workflows pass completely

**Quality Requirements**:
- [ ] No increase in technical debt
- [ ] Clear separation between library and samples
- [ ] Documentation covers library usage
- [ ] Generator usage is intuitive
- [ ] Generated projects follow clean architecture

**Migration Requirements**:
- [ ] Git history preserved for govern packages
- [ ] Git history preserved for sample app
- [ ] No broken imports in migrated code
- [ ] All dependencies resolved correctly

---

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Import path conflicts | High (breaking changes) | Keep existing module path, add compatibility layer |
| Git history loss | Medium (loss of context) | Use git mv for moves, preserve commits |
| CI/CD breakage | Medium (blocked deployments) | Test workflows in feature branch first |
| Template maintenance | Low (drift over time) | Document template update process, add tests |
| Sample app discoverability | Low (users miss samples) | Update README, link prominently to samples/ |

---

## Next Steps

1. **Review & Approve** - Validate design approach
2. **Create Implementation Plan** - Use `/ck:plan` with this design as input
3. **Execute Migration** - Follow big bang migration approach
4. **Test & Validate** - Run comprehensive test suite
5. **Update Documentation** - Ensure all docs reflect new structure
6. **Release** - Tag new version, announce migration

---

## Open Questions - RESOLVED

1. ~~**Repository naming**: Keep `golang-sample` or rename to `govern`?~~
   - **RESOLVED**: Keep `golang-sample` repository, govern packages use `github.com/haipham22/govern` module

2. ~~**Module path**: Keep `github.com/haipham22/golang-sample` or change?~~
   - **RESOLVED**: Govern packages → `github.com/haipham22/govern`, Sample app → `github.com/haipham22/golang-sample` with replace directive

3. ~~**Existing consumers**: How many external projects import govern? (Survey needed)~~
   - **RESOLVED**: None - safe to proceed with any module path changes

4. ~~**Template features**: What minimum feature set for basic template?~~
   - **RESOLVED**: Echo HTTP server + graceful shutdown + logging + config management

5. ~~**Generator language**: Bash vs Go implementation for generator?~~
   - **RESOLVED**: Go implementation for better error handling and template processing

---

**Final Recommendation**: Proceed with big bang migration. **Repository renamed to `github.com/haipham22/govern`**. Govern packages at root as `github.com/haipham22/govern` module. Sample app moves to `samples/golang-sample/` using Go workspaces for local development. Interactive generator implemented in Go creates projects importing `github.com/haipham22/govern`. No breaking changes for external consumers (none exist).

**Critical Issue Resolution**:
- ✓ Module path contradiction fixed (repository renamed to govern)
- ✓ Import path chaos avoided (current imports already use github.com/haipham22/govern)
- ✓ Generator flaw fixed (repository and module path now match)
- ✓ CI/CD needs redesign for go.work workspace
- ✓ Git history needs merge strategy (add git fast-export step)
