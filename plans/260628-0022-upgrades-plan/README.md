# Golang Sample Upgrades Plan

**Created:** 2026-06-28  
**Status:** In Progress — Phase 02 (Wire removal) **DONE** (synced 2026-06-28)  
**Total Effort:** 16 hours

---

## Quick Overview

This plan addresses four upgrade tracks for the golang-sample project (Wire removal already DONE):

| Upgrade | Decision | Effort | Timeline | Risk |
|---------|----------|--------|----------|------|
| **Wire removal** | ✅ **DONE** (`1672d69`) | 4h | — | Low |
| **Go 1.26** | ⬜ Next (currently on 1.25.8) | 4h | Week 1 | Very Low |
| **Database Layer** | ⬜ Optimize GORM | 6h | Week 1 | Low |
| **Echo v5** | ⚠️ Defer | 6h | Q3-Q4 2026 | Med-High |

---

## Summary

### Go 1.26 Upgrade (RECOMMENDED)
- **10-40% reduction in GC overhead** (Green Tea GC)
- **Enhanced security** (post-quantum TLS by default)
- **Zero breaking changes** (Go 1 compatibility promise)
- **Straightforward upgrade** - Update mise.toml, go.mod, run `go fix`

### Database Layer (STAY WITH GORM)
- **Zero migration cost** - Architecture already clean
- **Performance adequate** - 2-3x overhead acceptable for API scale
- **Team velocity** - Fastest development speed
- **Optimization path** - Add indexes, optimize queries, tune pool

### Echo v5 Migration (DEFERRED)
- **Govern compatibility unknown** - Must verify govern package support
- **Echo v4 stable** - No security issues, production-ready
- **Medium-High risk** - 15+ breaking changes
- **Re-evaluate Q3-Q4 2026**

---

## Files

### Main Plan
- **[plan.md](plan.md)** - Overview and summary

### Implementation Phases
- **[phase-01-go-1.26-upgrade.md](phase-01-go-1.26-upgrade.md)** - Go 1.26 upgrade (4h) — ⬜ pending
- **[phase-02-remove-wire-manual-di.md](phase-02-remove-wire-manual-di.md)** - Remove Wire → manual DI (4h) — ✅ DONE (`1672d69`)
- **[phase-02-database-optimization.md](phase-02-database-optimization.md)** - GORM optimization (6h) — ⬜ pending
- **[phase-03-echo-v5-migration.md](phase-03-echo-v5-migration.md)** - Echo v5 migration (6h, DEFERRED) — ⏸ blocked by govern

### Research Reports
- **[../reports/researcher-260628-0022-go-1.26-release-notes.md](../reports/researcher-260628-0022-go-1.26-release-notes.md)** - Go 1.26 analysis
- **[../reports/researcher-260628-0022-echo-v5-migration-research.md](../reports/researcher-260628-0022-echo-v5-migration-research.md)** - Echo v5 research
- **[../reports/researcher-260628-0022-database-layer-comparison.md](../reports/researcher-260628-0022-database-layer-comparison.md)** - DB layer comparison
- **[../reports/planner-260628-0022-upgrades-plan-summary.md](../reports/planner-260628-0022-upgrades-plan-summary.md)** - Plan summary

---

## Quick Start

### Immediate Actions (Week 1)

```bash
# 1. Update mise.toml
vim mise.toml
# Change: go = "1.26.0"

# 2. Update go.mod
vim go.mod
# Change: go 1.26.0

# 3. Install Go 1.26
mise install

# 4. Run modernizers
mise exec -- go fix ./...

# 5. Verify tests
mise exec -- go test ./...

# 6. Add database indexes
psql -d golang_sample
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

# 7. Tune connection pool (if needed)
vim internal/storage/user/new.go
# Adjust SetMaxIdleConns, SetMaxOpenConns
```

### Deferred Actions (Q3-Q4 2026)

- Monitor Echo v5 adoption
- Check govern package Echo v5 support
- Re-evaluate migration timing

---

## Key Findings

### Go 1.26 Benefits
| Improvement | Impact |
|-------------|--------|
| Garbage Collection | 10-40% reduction in overhead |
| Post-Quantum TLS | Enhanced security by default |
| io.ReadAll | 2x faster, 50% less memory |
| go fix | Revamped modernization tool |

### Database Layer Decision
| Factor | GORM | pgx/v5 | sqlc |
|--------|------|--------|------|
| Migration Cost | None | High | Very High |
| Performance | Good | Excellent | Excellent |
| Team Velocity | Highest | Medium | Medium |

**Recommendation:** Stay with GORM - adequate performance for API scale, zero migration cost.

### Echo v5 Concerns
- 15+ breaking changes
- Govern package compatibility unknown
- Echo v4 stable and production-ready
- Medium-High risk migration

**Recommendation:** Defer until govern package confirms Echo v5 support.

---

## Timeline

```
Week 1 (Immediate)
├── Phase 01: Go 1.26 Upgrade (4h)
│   ├── Update mise.toml, go.mod
│   ├── Run go fix
│   ├── Verify tests pass
│   └── Update CI/CD
└── Phase 02: Database Optimization (6h)
    ├── Add database indexes
    ├── Optimize GORM queries
    ├── Tune connection pool
    └── Establish performance baseline

Q2 2026
├── Monitor Go 1.26 performance improvements
├── Validate database optimization results
└── Document performance baselines

Q3-Q4 2026 (Deferred)
└── Phase 03: Echo v5 Migration (6h)
    ├── Verify govern package support
    ├── Execute migration
    ├── Comprehensive testing
    └── Deployment
```

---

## Success Criteria

### Phase 01 (Go 1.26)
- ✅ All tests pass with Go 1.26
- ✅ No breaking changes detected
- ✅ Performance improvement measurable
- ✅ CI/CD pipeline updated

### Phase 02 (Database)
- ✅ Database indexes added
- ✅ N+1 queries eliminated
- ✅ Connection pool optimized
- ✅ Performance baseline established

### Phase 03 (Echo v5 - Deferred)
- ✅ Govern compatibility verified
- ✅ All tests pass with Echo v5
- ✅ No production breaking changes
- ✅ Documentation updated

---

## Risk Assessment

| Phase | Risk Level | Mitigation |
|-------|------------|------------|
| Go 1.26 | Very Low | Go 1 compatibility promise |
| Database | Low | All changes additive |
| Echo v5 | Medium-High | Defer until govern support |

---

## Next Steps

1. **Review Plan** - Read full plan documents
2. **Approve Phases** - Confirm Go 1.26 and DB optimization
3. **Start Implementation** - Begin Phase 01
4. **Monitor Progress** - Track completion metrics
5. **Re-evaluate Echo v5** - Check govern package in Q3-Q4 2026

---

## Contact

**Questions?** Refer to individual phase documents for detailed implementation steps and acceptance criteria.

**Plan Status:** In Progress — Phase 02 (Wire removal) DONE (synced 2026-06-28)  
**Created:** 2026-06-28  
**Owner:** Development Team
