# Govern Monorepo Restructuring - Design & Planning Session

**Date**: 2026-06-27 21:32
**Severity**: [Critical]
**Component**: [Repository architecture / Governance structure]
**Status**: [Resolved - Planning Complete]

## What Happened

Successfully completed design and planning for govern monorepo restructuring. Session included scout exploration of golang-sample and govern packages, collaborative brainstorming, red team review that identified 5 critical issues, resolution of all blockers, and creation of detailed 8-phase implementation plan (16 hours effort).

## The Brutal Truth

This restructuring is absolutely necessary because the current golang-sample repository mixes library code with sample applications, creating confusion about what's production-ready vs. demo code. The frustrating part is that we've tolerated this ambiguity for months - developers can't distinguish between govern packages (actual library) and sample implementations (reference examples). This feels like a massive cleanup that should have happened earlier, but better late than never.

## Technical Details

**Repository Structure Decision:**
- New repository: `github.com/haipham22/govern`
- Go workspaces for multi-module development (golang.work)
- Module path standardized: `github.com/haipham22/govern/{module}`
- Samples isolated in `/samples` directory

**History Preservation Strategy:**
- git fast-export/import for history migration
- Preserves commit metadata, authors, and branching
- Critical for maintaining CI/CD pipeline continuity

**CI/CD Architecture:**
- Separate workflows for govern and samples
- Govern triggers on all modules: `['**/go.mod']`
- Samples triggers on sample-specific files: `['samples/**', 'cmd/**', '.github/workflows/samples.yml']`

**Interactive Generator:**
- Go-based CLI tool for project template creation
- Replaces complex bash scripts with programmatic approach
- Supports module selection, dependency injection, and configuration generation

## What We Tried

Initial brainstorming explored several approaches:
1. **Single repository with subdirectories** - Rejected due to import path confusion
2. **Multiple repositories** - Rejected due to maintenance overhead
3. **Monorepo with separate samples** - **Selected** for clear separation while maintaining unified governance

## Root Cause Analysis

The fundamental issue was unclear project boundaries. The name "golang-sample" implied everything was demo code, when in fact govern packages were production-ready libraries. This caused:
- Confusion about stability guarantees
- Unclear versioning strategy
- Mixed CI/CD requirements (library vs. application)
- Difficulty in onboarding new developers

## Lessons Learned

1. **Repository names matter** - "sample" in the repo name created false assumptions about code stability
2. **Library vs. sample separation is critical** - Clear boundaries prevent confusion about what's production-ready
3. **Go workspaces enable clean monorepos** - Modern tooling solves historical module path issues
4. **History preservation is non-negotiable** - Fast-export/import ensures we don't lose valuable context
5. **Red team reviews save time** - Catching 5 critical issues during planning prevented expensive rework

## Next Steps

**Immediate Actions:**
1. Execute phase-01 (repository initialization) - Create github.com/haipham22/govern
2. Execute phase-02 (go.work setup) - Configure workspace with govern and samples
3. Execute phase-03 (module migration) - Migrate govern packages to new structure
4. Execute phase-04 (history migration) - Fast-export/import git history

**Dependency Chain:**
- Restructure (phases 01-04) MUST complete before wire removal (phase-06)
- Wire removal depends on clean module boundaries established by restructuring
- Total effort: 16 hours across 8 phases

**Critical Success Factors:**
- Preserve all git history during migration
- Maintain backward compatibility for existing govern imports
- Ensure CI/CD pipelines function correctly in new structure
- Validate go.work resolution before publishing new modules

**Timeline:** Ready to begin immediate execution of phase-01
