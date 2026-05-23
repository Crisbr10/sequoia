# Sequoia Task Index — sequoia-ai

**Audit Date**: 2026-05-21 | **Audit ID**: audit-20260521-sequoia-ai | **Global Score**: 32/100 (Critical)

---

## Global Dependency Graph

```
```
✅ CORR-001 (Release Pipeline Hardening) — COMPLETED
   ├──► P1-001 (binary verification) ✅
   ├──► P1-007 (pin actions) ✅
   ├──► P1-008 (env protection) ✅
   ├──► P6-001 (block continue-on-error) ✅
   ├──► P6-002 (approval gate) ✅
   ├──► P6-003 (cross-platform tests) ✅
   ├──► P6-004 (workflow_dispatch) ✅
   ├──► P6-005 (CODEOWNERS) ✅
   ├──► P6-006 (post-deploy smoke) ✅
   └──► P6-008 (retry in installer) ✅

✅ CORR-002 (Global Mutable State) — COMPLETED    ✅ CORR-003 (TUI Rendering) — COMPLETED
├──► P3-001 (DefaultRegistry) ✅       ├──► P2-002 (precompute logo) ✅
├──► P3-008 (CommandFiles) ✅          ├──► P2-005 (cache styles) ✅
├──► P3-009 (templateCache) ✅         ├──► P2-006 (Grow hints) ✅
└──► P1-004 (init removal) ✅          └──► P2-011 (lazy adapter) ✅

✅ CORR-004 (Common Package Split) — COMPLETED
├──► P3-002 (pipeline refactor) ✅     ├──► P3-003 (ISP fix) ✅
├──► P3-010 (strategy pattern) ✅      ├──► P4-007 (gosec fix) ✅
└──► P4-003 (error context) ✅

Independent Quick Wins:
~~P2-001  Detect caching~~ ✅     ~~P4-001  go.sum cleanup~~ ✅
~~P4-004  Magic number 64~~ ✅    P6-007  Audit command
~~P6-009  Separate tokens~~ ✅    P6-010  Structured logging
```

---

## Priority Tiers

### 🔴 Tier 1 — Blocking (resolve first)

Tasks that unblock other tasks or close multiple findings:

| Task | Phase | Resolves | Effort | Status |
|------|-------|----------|--------|--------|
| CORR-001 | Sec+Ops | 10 findings (P1-001,P1-007,P1-008,P6-001–P6-006,P6-008) | large | ✅ COMPLETED |
| CORR-003 | Perf | 4 findings (P2-002,P2-005,P2-006,P2-011) | small | ✅ COMPLETED |
| ~~P6-001~~ | Ops | ~~1 finding + unblocks CORR-001~~ | small | ✅ (in CORR-001) |
| ~~P6-002~~ | Ops | ~~1 finding + unblocks CORR-001~~ | small | ✅ (in CORR-001) |

### 🟠 Tier 2 — High Leverage

High impact, independent, can be done in parallel:

| Task | Phase | Resolves | Effort |
|------|-------|----------|--------|
| P2-001 | Perf | Uncached PATH scans → reduces CLI latency ~40% | small | ✅ COMPLETED |
| P4-001 | Quality | Typosquat entry in go.sum → supply chain safe | small | ✅ COMPLETED |
| P4-004 | Quality | Magic number constant → consistency across 22+ files | small | ✅ COMPLETED |
| ~~P6-005~~ | Ops | CODEOWNERS → enforces review on release path | small | ✅ (in CORR-001) |
| P6-009 | Ops | Separate tokens → least privilege | small | ✅ COMPLETED |
| CORR-002 | Arch+Sec | 4 findings, removes global mutable state | medium | ✅ COMPLETED |
| CORR-004 | Arch+Quality | 5 findings, splits god package | large | ✅ COMPLETED |

### 🟡 Tier 3 — Backlog

Structural improvements, lower urgency, can schedule after Tier 1+2:

| Task | Phase | Resolves | Effort |
|------|-------|----------|--------|
| P1-002 | Security | File permission hardening | small |
| P1-003 | Security | Probe file permissions | small |
| P1-005 | Security | Sanitize home path in output | small |
| P1-006 | Security | Validate working-directory input | small |
| P1-009 | Quality | Verify go.sum cleanup | small |
| P1-010 | Security | Signature verification for adapters | medium |
| P3-004 | Arch | Fix double error wrapping | small |
| P3-007 | Arch | Align _template adapter | small |
| P4-009 | Quality | Standardize error wrapping | small |
| P6-007 | Ops | Implement audit command output | medium |
| P6-010 | Ops | Structured logging with slog | medium |
| P6-011 | Ops | Enable -race on Windows CI | small |
| P6-013 | Ops | Sync CHANGELOG with releases | small |

---

## Risk Estimate per Area

| Area | Risk Level | Rationale |
|------|-----------|-----------|
| **Release Safety** | **LOW** | ✅ CORR-001 resolved — approval gates, binary verification, CODEOWNERS, smoke tests, and cross-platform testing now in place. |
| **CI Integrity** | **LOW** | ✅ CI phased with blocking vulncheck, coverage gate at 70%, 5-OS matrix (CORR-001). |
| **Supply Chain** | **LOW** | ✅ Actions pinned to commit SHAs (CORR-001). Binary verification in place. Typosquat entry resolved (P4-001: replace directive + go mod verify in CI). |
| **TUI Performance** | **LOW** | ✅ CORR-003 (precompute logo, cache styles, Grow hints, lazy adapter) + P2-001 (PATH scan caching). All fixes applied. |
| **Architecture Debt** | **LOW** | ✅ CORR-002 (global state) + CORR-004 (pipeline/ISP/Strategy/gosec) resolved. Only P3-005 (package split), P3-006 (type assertions), P3-007 (template alignment) remain. |
| **Code Quality** | **LOW** | Strong fundamentals: 65 test files, parallel tests, zero lint suppressions. Issues are small, isolated improvements. |
| **User Security** | **LOW** | Only minor file permission concerns found. No PII handling, no network I/O, no authentication. |

---

## Effort Summary

| Tier | Tasks | Estimated Total Effort |
|------|-------|----------------------|
| 🔴 Tier 1 (Blocking) | ~~4~~ ✅ COMPLETED | ✅ Done |
| 🟠 Tier 2 (High Leverage) | ~~7~~ ✅ COMPLETED | ✅ Done |
| 🟡 Tier 3 (Backlog) | 14 tasks | 20–36 hours |
| **Total** | **~~25~~ 14 remaining** | **~~96–148~~ 20–36 hours** |

---

## Area Task Files

| Area | File | Score | CRITICAL | HIGH | MEDIUM |
|------|------|-------|----------|------|--------|
| 🔒 Security | [security.md](security.md) | 35 | 1 | 0 | 9 |
| ⚡ Performance | [performance.md](performance.md) | 38 | 0 | 2 | 9 |
| 🏗️ Architecture | [architecture.md](architecture.md) | 36 | 0 | 2 | 8 |
| ✅ Quality | [quality.md](quality.md) | 48 | 0 | 0 | 10 |
| 🔧 Operations | [operations.md](operations.md) | 0 | 2 | 4 | 9 |

---

*Generated by Sequoia M2 Reporter — audit-20260521-sequoia-ai | Schema v1.0*
