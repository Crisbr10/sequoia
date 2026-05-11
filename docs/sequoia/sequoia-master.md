# Sequoia Audit — Master Report

**Project**: sequoia-ai v0.1.0  
**Date**: 2026-05-11  
**Mode**: Full  
**Health Score**: 🟢 **78/100** (Grade B)

---

## Executive Summary

Sequoia is a **well-structured Go CLI tool** with a clean plugin architecture and strong test culture. The codebase is fundamentally sound — no circular dependencies, no hardcoded secrets, no injection vectors. The audit identified 44 findings across 4 domains, with the most actionable improvements concentrated in **code deduplication across adapters** and **interface design for evolvability**.

The single highest-ROI remediation is to **fully utilize `adapters/common/`** — extracting the shared install workflow, template embedding, and path resolution into common helpers. This eliminates 8 findings at once, reduces ~500 lines of duplicated code, and makes future adapter additions a 10-line exercise instead of an 80-line copy-paste.

---

## Health Score Dashboard

| Category | Score | Grade | Findings |
|----------|-------|-------|----------|
| Security | 85 | B+ | 7 (0C, 0H, 2M, 4L, 1I) |
| Performance | 82 | B | 9 (0C, 0H, 3M, 3L, 3I) |
| Architecture | 65 | C+ | 13 (0C, 4H, 7M, 2L) |
| Quality | 79 | B | 15 (0C, 2H, 6M, 5L, 2I) |
| **Global** | **78** | **B** | **44 total — 0 critical, 6 high** |

---

## Top Findings by Severity

### 🔴 High (6 findings)

| ID | Category | Title |
|----|----------|-------|
| P3-001 | Architecture | 507+ duplicated lines across all 6 adapters |
| P3-002 | Architecture | All xxxBase() share identical symlink resolution code |
| P3-004 | Architecture | 25 command template files duplicated identically across adapters |
| P3-007 | Architecture | Adding any method to ToolAdapter breaks 12+ files |
| P4-001 | Quality | interface{} used instead of `any` in Go 1.24 codebase (20+ occurrences) |
| P4-002 | Quality | 5 adapter Install() methods nearly identical copy-paste (~250 lines) |

### 🟡 Medium (18 findings)

Security: P1-001 (pipe-to-shell), P1-002 (checksum skip)  
Performance: P2-001 (duplicate templates), P2-002 (duplicate InjectSection), P2-003 (Lipgloss style caching)  
Architecture: P3-003, P3-005, P3-006, P3-008, P3-010, P3-012  
Quality: P4-003, P4-004, P4-005, P4-006, P4-007, P4-008

---

## Root Causes (Cross-Phase Correlation)

### RC1 🔴 — adapters/common/ underutilized *(8 findings across 3 agents)*

The `adapters/common/` package provides `Installer`, `RenderTemplate`, `StageFile`, and `CommandFiles` — but it doesn't go far enough. The identical `xxxBase()` functions, identical install workflows, identical template structs, and identical embedded command templates remain copy-pasted across 5 adapters. **Extracting these into common would eliminate the largest source of maintenance risk and binary bloat in the codebase.**

### RC3 🔴 — Rigid ToolAdapter interface *(3 findings)*

The interface has 11 methods with no default implementations. Adding any method breaks 6 concrete adapters, 1 mock, 1 template, and all adapter-specific test mocks — ~12 files must be updated atomically. **Splitting into smaller interfaces (Installable, Detectable, Statusable) would enable incremental evolution.**

### RC2 🟡 — Immature error system *(4 findings across 2 agents)*

The installer's 4-phase lifecycle produces errors that are all generic `fmt.Errorf` wrappers. There's no way to programmatically distinguish a Prepare failure from a Verify failure. **Typed errors would enable the TUI to offer phase-specific recovery options.**

### RC5 🟡 — Weak supply chain verification *(3 findings)*

The pipe-to-shell install pattern and skipped checksum verification in the GitHub Action are the most impactful security findings. **Strengthening verification does not require code changes — it's primarily documentation and CI configuration.**

---

## Remediation Roadmap

### Phase 1: Structural (7–11h) — fixes 50% of findings
- Extract `common.BaseResolver()` to eliminate 5 identical path functions
- Create `common.InstallSkills()` helper to eliminate 5 identical Install() methods
- Centralize command template embedding in `adapters/common/`
- Split `ToolAdapter` into smaller interfaces

### Phase 2: Reliability (8–12h) — fixes 35% of findings
- Define typed errors for install lifecycle phases
- Strengthen supply chain verification (documentation + CI config)
- Fill test coverage gaps in `cmd/sequoia/`
- Decompose `main.go` into per-command files

### Phase 3: Polish (3–5h) — fixes 15% of findings
- Replace `interface{}` with `any` globally
- Cache Lipgloss styles at package level
- Fix file naming (hyphens → underscores)
- Micro-optimizations (string builder, byte operations)

---

## What Went Well

- ✅ **Test quality** is commendably high — behavioral tests with real filesystem I/O, error paths covered
- ✅ **Dependency direction** is clean — no circular dependencies, correct layering
- ✅ **Security baseline** is strong — no secrets, correct permissions, proper input validation
- ✅ **Plugin architecture** is appropriate — database/sql registration pattern, clean separation between plugin/ and adapters/
- ✅ **CI/CD** is comprehensive — 3-OS matrix, GoReleaser, linting integrated
- ✅ **Documentation** is extensive — README, architecture docs, FAQ, CLI reference

---

## Deliverables Generated

```
docs/sequoia/
├── sequoia-master.md          ← This file
├── sequoia-score.md           ← Health scorecard with breakdown
├── sequoia-tasks.md           ← 49 actionable tasks with effort estimates
└── sequoia-phases/
    ├── 01-security.md         ← P1: 7 findings
    ├── 02-performance.md      ← P2: 9 findings
    ├── 03-architecture.md     ← P3: 13 findings
    ├── 04-quality.md          ← P4: 15 findings
    ├── 05-experience.md       ← Skipped (CLI tool)
    └── 06-operations.md       ← Partial (CI observations)
```
