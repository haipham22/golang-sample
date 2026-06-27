# Govern Monorepo Restructure - Red Team Review

**Date**: 2026-06-27  
**Type**: Adversarial Architecture Review  
**Status**: CRITICAL ISSUES FOUND

---

## Executive Summary

The proposed monorepo restructuring contains **5 critical flaws**, **8 high-severity concerns**, and **12 operational risks** that must be addressed before implementation. The design demonstrates incomplete thinking about Go module mechanics, CI/CD complexity, and template maintainability.

**Recommendation**: DO NOT PROCEED with current design. Requires significant revision.

---

## CRITICAL ISSUES (Must Fix Before Proceeding)

### 1. **Go Module Path Contradiction**
**Severity**: CRITICAL  
**Impact**: Breaking change, confusion, potential build failures

**Problem**: Design claims govern packages use `github.com/haipham22/govern` module but repository is named `golang-sample`. This is fundamentally incompatible with Go's module system.

```go
// Design states:
github.com/haipham22/govern  // Root module
// But repository is:
github.com/haipham22/golang-sample  // Repository URL
```

**Why This Breaks**:
- Go modules must match repository paths
- `go get github.com/haipham22/govern` will fail if repo is `golang-sample`
- Requires either: (a) rename repository, (b) use subdomain, or (c) accept mismatched paths

**Actual Required Structure**:
```
Option A: Rename repo → govern/
Option B: Use module subdomain → govern.golang-sample.com
Option C: Accept broken imports
```

**Evidence**: From design doc lines 86-89 claim this works, but Go module system documentation contradicts this.

---

### 2. **Import Path Chaos During Migration**
**Severity**: CRITICAL  
**Impact**: All imports break, non-functional intermediate states

**Problem**: Big bang migration moves govern packages to root but doesn't account for existing import references in current codebase.

**Current Imports**:
```go
import "github.com/haipham22/govern/http"
import "github.com/haipham22/govern/config"
import "github.com/haipham22/govern/database/postgres"
```

**After Migration (Design Claim)**:
```go
// Still uses external import?
import "github.com/haipham22/govern/http"
// Or local?
import "golang-sample/http"
```

**Missing Steps**:
- No import path migration strategy
- No compatibility layer
- No phased transition plan
- Assumes `replace` directive solves everything (it doesn't for CI/CD)

---

### 3. **CI/CD Workflow Assumptions Invalid**
**Severity**: CRITICAL  
**Impact**: Broken pipelines, failed deployments, blocked releases

**Problem**: Design assumes CI/CD can test "govern packages + samples" separately in same workflow, but current workflows are designed for single application.

**Current Test Workflow**:
```yaml
go test -v -race -coverprofile=coverage.out ./...
```

**Design Assumes**:
```yaml
# Test library packages
go test ./http/... ./database/... ./config/...
# Test samples  
go test ./samples/sample-app/...
```

**Critical Flaws**:
1. No coverage aggregation strategy (two separate coverage reports)
2. Mockery generation broken by nested module structure
3. Pre-commit hooks configured for single module (`-local golang-sample`)
4. Docker images: which module gets published?
5. Go caching in CI broken by module changes

---

### 4. **Template Generator Fundamental Flaw**
**Severity**: CRITICAL  
**Impact**: Generated projects won't compile, users stuck with broken code

**Problem**: Generator designed to create projects importing `github.com/haipham22/govern` but that module path doesn't match repository URL.

**Generator Creates**:
```go
// Generated project go.mod:
module github.com/user/my-api

require (
    github.com/haipham22/govern v0.1.0  // This path doesn't exist!
)
```

**But Repository Is**:
```
github.com/haipham22/golang-sample  // Wrong path
```

**Three Impossible Scenarios**:
1. If repo renamed to `govern`: breaks all existing golang-sample references
2. If repo stays `golang-sample`: generator creates broken imports
3. If subdomain used: requires DNS + registry setup

**Missing Validation**:
- No template testing strategy
- No integration test for generated projects
- No verification that `go get` would work

---

### 5. **Git History Loss Guaranteed**
**Severity**: CRITICAL  
**Impact**: Loss of attribution, broken blame, lost context

**Problem**: Design claims "git mv for moves, preserve commits" but then contradicts itself with "copy all packages from ../govern/" in Phase 2, Step 4.

**Design Says** (Line 107):
```
Copy all packages from ../govern/ to repository root
```

**Then Claims** (Line 381):
```
Git history preserved for govern packages
```

**Reality Check**:
- `git cp` (copy) does NOT preserve history
- `git mv` (move) requires same repo
- Cross-repo history merge requires `git filter-branch` or `git fast-export`
- No mention of history merge strategy

**Actually Required**:
```bash
# Not mentioned in design
git filter-branch --tree-filter ... # OR
git fast-export ../govern/ | git fast-import
```

---

## HIGH SEVERITY CONCERNS (Should Address)

### 6. **Replace Directive Production Anti-Pattern**
**Severity**: HIGH  
**Impact**: Production deployment failures

**Problem**: Design uses `replace` directive for sample app development but no strategy for removing it in production.

```go
replace github.com/haipham22/govern => ../
```

**Issues**:
- `replace` directives are local-only (not committed to go.sum)
- Can't use `replace` with published modules
- No documentation on when/how to remove replace
- CI/CD will fail with replace directive

---

### 7. **Module Dependency Hell**
**Severity**: HIGH  
**Impact**: Unresolvable imports, circular dependencies

**Problem**: Root module depends on sample app dependencies, but sample app depends on root module.

**Root go.mod** (govern):
```go
module github.com/haipham22/govern

require (
    github.com/labstack/echo/v4 v4.15.1  // From sample app?
    gorm.io/gorm v1.31.1                 // From sample app?
)
```

**Sample go.mod**:
```go
module github.com/haipham22/golang-sample

require (
    github.com/haipham22/govern v0.0.0  // Depends on root
)
```

**Missing Strategy**:
- Which dependencies belong where?
- How to prevent dependency leakage?
- What if govern and sample need different versions?

---

### 8. **Template Maintenance Nightmare**
**Severity**: HIGH  
**Impact**: Templates drift from best practices, generation creates outdated code

**Problem**: No strategy for keeping templates synchronized with govern package changes.

**Scenarios Not Addressed**:
- Govern API change → template still uses old API
- New govern package → templates don't include it
- Security fix in govern → generated projects use vulnerable version
- Best practice evolves → templates stuck in past

**Missing**:
- Template testing automation
- Template versioning strategy
- Template update mechanism
- Template-to-govern integration tests

---

### 9. **Pre-commit Hooks Configuration Broken**
**Severity**: HIGH  
**Impact**: Broken development workflow, inconsistent code quality

**Problem**: Current pre-commit hooks configured for single module (`-local golang-sample`) but design creates multi-module repo.

**Current Configuration**:
```yaml
- id: go-imports-local
  entry: goimports -local golang-sample  # Only works for one module
```

**After Migration**:
```
Multiple modules need different local imports:
- govern packages: -local govern  
- sample app: -local golang-sample
- generated projects: -local <varies>
```

**No Update Strategy**:
- Per-module pre-commit configs?
- Global hooks? (doesn't exist in Go)
- Remove hooks from templates?

---

### 10. **Wire Generation Broken by Module Structure**
**Severity**: HIGH  
**Impact**: Dependency injection failures, unbuildable code

**Problem**: Wire generates code using package paths. Multi-module structure breaks wire's package resolution.

**Current Wire Usage**:
```go
// internal/handler/rest/wire.go
+wirebuild
```

**After Migration**:
```
Root module (govern): no wire
Sample module: wire needs to import govern packages
Generated projects: wire needs to import external govern
```

**Missing**:
- Wire configuration for each module
- Wire gen file updates
- Cross-module wire strategy

---

### 11. **Mockery Generation Path Failures**
**Severity**: HIGH  
**Impact**: Test failures, broken mocks

**Problem**: Mockery generates mocks based on directory structure. Nested modules break mockery's path resolution.

**Current Mockery Config**:
```yaml
# .mockery.yml
packages:
  internal/storage:
    interfaces: Interface
  internal/service:
    interfaces: Interface
```

**After Migration**:
```
Root: no interfaces to mock
samples/sample-app/internal/: paths change, mockery breaks
Templates: need mockery config for each template
```

**CI/CD Impact**:
```yaml
- name: Install mockery
  run: go install github.com/vektra/mockery/v3@latest  # Only works once
  
- name: Generate mocks
  run: mockery  # Fails in nested modules
```

---

### 12. **Docker Multi-Module Madness**
**Severity**: HIGH  
**Impact**: Broken builds, oversized images, deployment failures

**Problem**: Single repository with multiple modules but unclear Docker strategy.

**Design Claims**:
```yaml
# push.yml
- Build library Docker image (minimal, for library usage)
- Optionally: Build sample app image
```

**Reality Check**:
- What is a "library Docker image"? Libraries don't run in containers.
- Which Dockerfile gets built? Root? Sample app?
- How to build sample app image from nested directory?
- What about generated project images?

**Missing Docker Strategy**:
- Multi-stage builds for each module
- Docker context configuration (can't build from nested dir easily)
- Image naming/publishing strategy
- Development vs production images

---

### 13. **Documentation Discoverability Disaster**
**Severity**: HIGH  
**Impact**: Users can't find relevant docs, abandoned project

**Problem**: Design scatters docs across multiple locations with no clear navigation.

**Current Structure**:
```
docs/
├── project-overview-pdr.md
├── code-standards.md
├── system-architecture.md
```

**After Migration**:
```
docs/                          # Govern docs?
├── quickstart.md             # Govern quickstart?
├── packages/                 # Govern package docs?
└── samples/
    └── sample-app-guide.md  # Sample app docs?

samples/sample-app/README.md  # Or here?
CLAUDE.md                      # Project instructions (govern or sample?)
```

**Missing**:
- Root README strategy (govern or sample?)
- Documentation index/navigation
- CLI vs library documentation separation
- Version-specific documentation

---

## OPERATIONAL ISSUS (Nice to Know)

### 14. **Migration Rollback Strategy Missing**
**Impact**: No way to undo if migration fails catastrophically

**Problem**: Big bang migration has no rollback plan.

**Missing**:
- Pre-migration checkpoint strategy
- Rollback verification steps
- Data migration rollback (if any)
- Team coordination plan

---

### 15. **No Validation of "No External Consumers"**
**Impact**: Potential breaking changes for unknown users

**Problem**: Design claims "no external projects import govern" but provides no evidence.

**Should Verify**:
- GitHub dependents check
- Go package index search
- Internal tool audit
- Private repo scan

---

### 16. **Generator CLI Distribution Unclear**
**Impact**: Users can't install/use generator easily

**Problem**: Design shows two incompatible installation methods.

```bash
# Method 1: Local build
go build -o bin/generate-project ./scripts/generate-project

# Method 2: Global install
go install github.com/haipham22/golang-sample/scripts/generate-project@latest
```

**Issues**:
- Method 2 requires published Go module (which is broken by issue #1)
- Method 1 requires cloning entire repo
- No standalone binary distribution
- No brew/snap/docker distribution

---

### 17. **Template Feature Selection Complexity**
**Impact**: Analysis paralysis for users, maintenance burden

**Problem**: Template features are hardcoded in generator, adding new features requires code changes.

**Current Template Features**:
```go
[*] PostgreSQL (GORM)
[*] Redis (caching)  
[ ] JWT authentication
[*] Prometheus metrics
[*] Health checks
[ ] Asynq task queue
[ ] Cron scheduler
```

**Missing**:
- Feature versioning (what if Prometheus API changes?)
- Feature compatibility matrix (can't combine X with Y)
- Feature testing strategy
- Feature documentation generation

---

### 18. **No Performance Testing Strategy**
**Impact**: Unknown performance regression

**Problem**: Migration introduces module boundaries that could affect performance.

**Not Tested**:
- Import path resolution overhead
- Module dependency resolution latency
- Build time impact (nested modules)
- Runtime performance (should be same, but not verified)

---

### 19. **Security Scanning Configuration Broken**
**Severity**: MEDIUM  
**Impact**: Missed vulnerabilities, false sense of security

**Problem**: Current security scanning assumes single module.

**Affected Tools**:
- `gosec` (source analysis)
- `govulncheck` (vulnerability scanner)
- Dependabot (dependency updates)
- Snyk/CodeQL (if used)

**Multi-Module Issues**:
- Scan each module separately or aggregate?
- How to handle transitive dependencies through modules?
- False positives from replace directives?

---

### 20. **No Developer Migration Guide**
**Impact**: Team confusion, lost productivity

**Problem**: Design documents technical steps but not developer workflow.

**Missing**:
- How to switch between working on govern vs sample app
- How to test changes to govern in sample app context
- How to release govern vs sample app
- How to onboard new developers to complex structure

---

### 21. **Go Workspaces Compatibility Unknown**
**Impact**: Future Go tooling compatibility

**Problem**: Design doesn't address Go workspace compatibility (Go 1.18+).

**Questions**:
- Does this work with `go work`?
- Should we use workspaces instead of replace directives?
- Workspace file location and management?

---

### 25. **License and Attribution Confusion**
**Impact**: Legal issues, unclear licensing

**Problem**: Combining two repositories with potentially different licenses.

**Need to Verify**:
- govern repository license
- golang-sample repository license
- License compatibility for combined repo
- Attribution requirements for copied code

---

## MISSING CONSIDERATIONS

### 22. **Local Development Experience**
**Impact**: Poor developer productivity

**Not Addressed**:
- How to run both govern and sample app in development?
- Hot reload across module boundaries?
- IDE configuration for multi-module repo (VS Code, GoLand)
- Local debugging across modules

---

### 23. **Release Versioning Strategy**
**Impact**: Confusing version numbers, breaking changes

**Not Addressed**:
- Govern library versioning (semantic versioning?)
- Sample app versioning (separate from govern?)
- Template versioning (which govern version do templates target?)
- Generated project versioning (pinned to govern version?)

---

### 24. **Integration Testing Strategy**
**Impact**: Unknown quality assurance

**Not Addressed**:
- End-to-end tests across module boundaries
- Integration tests for generator
- Template validation tests
- Sample app as integration test for govern?

---

### 25. **Telemetry and Usage Analytics**
**Impact**: No insight into adoption

**Not Addressed**:
- How to track govern usage?
- How to track generator usage?
- Which templates are most popular?
- Which govern packages are most used?

---

## POSITIVE OBSERVATIONS

### What Works Well

1. **Template System Concept**: Good idea to provide starter templates, just needs better execution
2. **Interactive Generator**: Go-based generator is right choice (better than bash)
3. **Feature Selection**: Multi-select template features is user-friendly approach
4. **Govern Package Organization**: Logical grouping of packages by concern

---

## UNRESOLVED QUESTIONS

1. **Repository Naming**: Will repository be renamed or not? Design contradicts itself.
2. **Module Path Strategy**: What is the actual Go module path for govern packages?
3. **Import Path Transition**: How do existing imports transition during migration?
4. **External Consumers**: Has anyone verified there are truly zero external consumers?
5. **CI/CD Consolidation**: How will single workflow test two separate modules properly?
6. **Docker Strategy**: What actually gets built and published?
7. **Documentation Ownership**: Who maintains which documentation?
8. **Template Testing**: How are templates validated against govern API changes?
9. **Performance Impact**: Has build/runtime performance been tested with multi-module structure?
10. **Rollback Plan**: What if migration fails catastrophically?

---

## RECOMMENDATIONS

### Do Not Proceed Until

1. **Resolve Module Path Contradiction**: Choose repository name OR module path, can't have both
2. **Validate "No External Consumers"**: Run actual dependency analysis
3. **Design Multi-Module CI/CD**: Test how workflows actually work with nested modules
4. **Create Migration Rollback Plan**: Document how to undo if things break
5. **Test Generator End-to-End**: Verify generated projects actually compile and run
6. **Address Wire/Mockery Breaking**: Plan for dependency injection tool updates

### Alternative Approaches to Consider

1. **Separate Repositories**: Keep govern and golang-sample separate, link via git submodule
2. **Go Workspaces**: Use Go 1.18+ workspaces instead of physical monorepo
3. **Monorepo Tools**: Use Bazel or similar for true monorepo management
4. **Template as Separate Repo**: Templates in separate repository, versioned independently

---

## CONCLUSION

**Current Design Status**: NOT READY FOR IMPLEMENTATION

The design demonstrates insufficient understanding of Go module mechanics, CI/CD complexity, and operational reality. The critical flaws around module paths and import management make the current approach unworkable.

**Estimated Fix Effort**: 2-3 weeks of additional design and prototyping

**Risk if Proceeded As-Is**: HIGH - Likely to fail during implementation or require complete restart

**Recommended Next Step**: Address critical issues #1-5 before any implementation begins. Conduct proof-of-concept testing of multi-module CI/CD and generator before committing to approach.

---

**Review Status**: REQUIRES REDESIGN  
**Next Action**: Revise design addressing critical issues  
**Re-review Date**: After critical issues resolved