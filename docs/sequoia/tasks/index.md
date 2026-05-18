# Global Task Index — Sequoia Audit

## Dependency Graph

```
RC-004 (backup namespace) ─────────────────────────────────────────────────┐
  └─► P3-001 (cross-installer restore)                                     │
  └─► P6-007 (backup cleanup)                                              │
                                                                            │
RC-001 (god-object) ─────────────────────────────────────────────────────┐ │
  └─► P3-002 (22 function fields)                                        │ │
  └─► P3-003 (ISP violation, 12 methods)                                 │ │
  └─► P3-004 (detection vs install coupling)                             │ │
  └─► P3-005 (global mutable state)                                      │ │
  └─► P3-006 (BackupManager missing — DEPENDS ON RC-004 ▲)              │ │
  └─► P3-007 (factory leak)                                              │ │
  └─► P4-001 (error-path tests)                                          │ │
  └─► P4-002 (mock embed.FS tests)                                       │ │
                                                                            │
RC-006 (eager init) ──────────────────────────────────────────────────────┐ │
  └─► P2-002 (init() blocks)                                              │ │
  └─► P2-003 (symlink on every call)                                      │ │
  └─► P2-004 (template re-execution)                                      │ │
  └─► P2-005 (sequential embed reads)                                     │ │
                                                                            │ │
RC-002 (script integrity) ────────────────────────────────────────────┐   │ │
  └─► P1-001 (path sanitization)                                       │   │ │
  └─► P1-002 (PATH injection)                                          │   │ │
  └─► P6-002 (script verification in release)                          │   │ │
                                                                       │   │ │
RC-003 (CI/CD gates) ──────────────────────────────────────────────┐  │   │ │
  └─► P6-001 (SBOM missing)                                         │  │   │ │
  └─► P6-003 (no vuln scanning)                                     │  │   │ │
  └─► P6-004 (no arm64 in matrix)                                   │  │   │ │
  └─► P6-005 (Dependabot only GH Actions)                          │  │   │ │
  └─► P6-008 (no checksum signing)                                  │  │   │ │
                                                                    │  │   │ │
Standalone HIGH (no blockers, fix immediately):                    │  │   │ │
  P2-001 (template cache key)                                       │  │   │ │
  P4-005 (Codex write error swallowed)                             │  │   │ │
                                                                    │  │   │ │
RC-005, RC-007 (medium root causes — no downstream blockers)      │  │   │ │
                                                                    │  │   │ │
MERGED-001 (go-figure consolidation)                               │  │   │ │
MERGED-002 (recover pattern improvement)                           │  │   │ │
```

## Priority Tiers (All Areas)

### Tier 1 — Immediate (Critical + High)
**Risk**: Silent data loss, wrong content installation, potential supply-chain compromise

| # | ID | Area | Task | Effort | Blocks |
|---|-----|------|------|--------|--------|
| 1 | RC-004 | Architecture | Namespace backup directories by adapter ID | small | P3-001, P6-007, P3-006 |
| 2 | P2-001 | Architecture | Fix template cache key | small | — |
| 3 | P4-005 | Quality | Propagate Codex write errors | small | — |
| 4 | RC-001 | Architecture | Decompose BaseAdapter | large | 8 findings |
| 5 | RC-002 | Security | Add install script integrity | medium | P1-001, P1-002 |
| 6 | RC-003 | Operations | Add CI/CD security gates | medium | 8 findings |
| 7 | RC-006 | Performance | Lazy adapter loading | medium | 5 findings |

### Tier 2 — Short Term (Medium)
**Risk**: Degraded reliability, missing security hardening, architectural debt accumulation

| # | ID | Area | Task | Effort |
|---|-----|------|------|--------|
| 8 | MERGED-001 | Quality | Consolidate go-figure usage | small |
| 9 | P6-002 | Operations | Script verification in CI | small |
| 10 | P4-007 | Architecture | Deduplicate shared types | small |
| 11 | P3-002 | Architecture | Encapsulate BaseAdapter fields | medium |
| 12 | P4-001 | Quality | Error-path test coverage | medium |
| 13 | P3-004 | Architecture | Separate detection from installation | medium |

### Tier 3 — Long Term (Low + Info)
**Risk**: Code hygiene, minor hardening, non-critical improvements

## Risk Estimate

### If we fix nothing
- **Blast radius**: Any user who runs Sequoia in headless/TUI mode (P2-001) or has multiple adapters installed (P3-001/RC-004) **will experience data loss**. The silent error swallowing (P4-005) means they won't know their Codex config is corrupt.
- **Supply chain**: A compromised install script (RC-002) would affect 100% of users. With no verification, no SBOM, and no vuln scanning (RC-003), detection would be impossible until post-mortem.
- **Architecture**: The god-object (RC-001) means any bug fix in BaseAdapter risks breaking all 6 adapters. Without refactoring, every new adapter adds coupling.

### If we fix Tier 1 only
- Health Score improves from 3/100 to approximately **65/100**
- Silent data loss eliminated
- Supply-chain integrity established
- Architecture becomes testable and adaptable

### If we fix Tier 1 + Tier 2
- Health Score reaches approximately **82/100**
- All medium-severity issues resolved
- CI/CD includes comprehensive gates
- Test coverage covers error paths

### Time estimate
- Tier 1: 3-4 weeks (RC-001 is the long pole at ~2 weeks)
- Tier 2: 2-3 weeks
- Tier 3: 1-2 weeks spread over maintenance cycles

## Cross-Area Dependencies

| Task | Depends On | Area |
|------|-----------|------|
| P3-006 (BackupManager) | RC-001 (god-object decomposition) + RC-004 (backup namespace) | Architecture |
| P4-001 (error-path tests) | RC-001 (stable adapter interface) | Quality |
| P2-002..P2-005 (perf improvements) | RC-006 (lazy loading foundation) | Performance |
| P1-001, P1-002 (security hardening) | RC-002 (integrity infrastructure first) | Security |
| P6-001..P6-010 (CI/CD improvements) | RC-003 (CI/CD foundation) | Operations |

## Relevant Files

- `adapters/common/base_adapter.go` — Monolithic 482-line god-object (RC-001)
- `adapters/common/template.go` — Template cache with stack-address key (P2-001)
- `adapters/common/installer.go` — Installer lifecycle with shared backup dir (RC-004)
- `adapters/codex/toml_merge.go` — TOML merge with silent write error (P4-005)
- `adapters/interface.go` — 12-method ToolAdapter interface (P3-003)
- `adapters/registry.go` — Global mutable DefaultRegistry (P3-005)
- `.github/workflows/ci.yml` — CI pipeline lacking security gates (RC-003)
- `.golangci.yaml` — Linter config excluding main.go from errcheck (P4-008)
- `install.ps1` — Unverified install script (RC-002)

---

*Generated by M2 Reporter — Sequoia v1.0.7 audit framework*
