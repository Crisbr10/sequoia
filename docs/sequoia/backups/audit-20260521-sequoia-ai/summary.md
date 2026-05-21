# Sequoia Audit Report — sequoia-ai

**Date**: 2026-05-21
**Stack**: Go 1.24 (Cobra + Bubbletea CLI/TUI)
**Size**: Medium (133 files, 19 packages, 27 deps)
**Maturity**: Growth
**Audit ID**: audit-20260521-sequoia-ai
**Schema Version**: 1.0

---

## Global Health Score: 23/100 — Critical

> **Note on scoring**: The deterministic formula penalizes finding count heavily. Many findings are small-effort fixes (<1 hour each). The project has strong fundamentals: 65 test files, 5-OS CI matrix, clean zero-lint-suppression codebase, and well-respected architecture boundaries. The low score reflects accumulation of many small gaps rather than fundamental flaws.

### Phase Scores

| Phase | Score | Classification | Findings |
|-------|-------|----------------|----------|
| 🔒 Security | **70**/100 | Fair | 4 (0 crit, 0 high, 2 med, 2 low) |
| ⚡ Performance | **0**/100 | Critical | 12 (0 crit, 2 high, 4 med, 6 low) |
| 🏗️ Architecture | **0**/100 | Critical | 10 (0 crit, 2 high, 4 med, 4 low) |
| ✅ Quality | **20**/100 | Critical | 12 (0 crit, 1 high, 3 med, 8 low) |
| 🔧 Operations | **0**/100 | Critical | 15 (0 crit, 4 high, 6 med, 5 low) |

> 🎨 Experience: N/A (CLI/TUI tool excluded)

---

## Root Causes Identified

### CORR-001: adapters/common God Package (Architecture + Security)
**Confidence**: 0.85 | **Symptoms**: 4 | **Risk**: HIGH

The `adapters/common` package concentrates 7 exported types, 16+ functions across 6 unrelated concerns including installer file operations. File permission issues (P1-001, P1-002) are direct consequences of installer code living in a package without clear security boundaries. Global mutable state (P3-005) compounds the problem.

**Fix**: Split into `adapters/install`, `adapters/templates`, `adapters/detect` packages. Resolves P3-001, P1-001, P1-002, P3-005. **Effort**: large (12-24h)

### CORR-002: CI Pipeline Quality Gate Absence (Operations + Quality)
**Confidence**: 0.90 | **Symptoms**: 6 | **Risk**: HIGH

The CI pipeline is a single monolithic job without quality gates. `govulncheck` uses `continue-on-error: true` (CVEs won't block PRs), coverage is never collected (`-coverprofile` missing), and build runs even after test failures. Both P4 and P6 independently identified the same CI weaknesses.

**Fix**: Split CI into separate jobs with `needs`, remove `continue-on-error`, add `-coverprofile` with threshold. Resolves P6-001, P4-003, P6-005, P4-005, P4-002, P6-007. **Effort**: small (2-4h)

### CORR-003: Uncached I/O Pattern Across Adapter Infrastructure (Performance + Architecture)
**Confidence**: 0.75 | **Symptoms**: 4 | **Risk**: HIGH

`Detect()`, `IsInstalled()`, and `Status()` in BaseAdapter/Detector independently perform I/O without memoization. The `templateCache` pattern exists but wasn't extended. TUI renders call these at 60fps, causing jank. CLI commands like `sequoia status` make 5+ PATH scans + 10 file reads per invocation.

**Fix**: Add internal cache to Detector/BaseAdapter memoizing Detect/IsInstalled/Status for process lifetime. Resolves P2-001, P2-002, P2-003, P2-006, P2-008, P2-011, P2-012. **Effort**: small (2-4h)

### CORR-004: Broad Interface → Pipeline Type Safety Gap (Architecture + Quality)
**Confidence**: 0.70 | **Symptoms**: 4 | **Risk**: MEDIUM

`ToolAdapter` has 11 methods. The pipeline strips to `ToolInfo` (5 methods) then uses unguarded type assertions to recover `Install`/`Uninstall`. The magic number `64` spreads across 22+ files without an interface-defined contract.

**Fix**: Define narrower role interfaces for pipeline consumption, replace type assertions with interface checks. Extract progress channel buffer as named constant. Resolves P3-002, P3-006, P4-004. **Effort**: medium (8-16h)

### CORR-005: Release Infrastructure Immaturity (Operations + Quality)
**Confidence**: 0.78 | **Symptoms**: 5 | **Risk**: HIGH

The GitHub Action delivers placeholder output (`health_score=N/A`). Releases lack smoke tests, rollback isn't documented, and code coverage isn't tracked. These are symptoms of prioritizing feature velocity over operational maturity.

**Fix**: Implement audit command, add post-release verification to release.yml, document rollback procedure. Resolves P6-002, P6-003, P6-004, P6-010. **Effort**: large (16-32h)

---

## Prioritized Roadmap

### 🔴 Immediate (Critical + High, Ordered by Dependencies)

| ID | Action | Resolves | Effort | Blocks |
|----|--------|----------|--------|--------|
| CORR-002 | **Fix CI quality gates** — split jobs, remove continue-on-error, add coverage | 6 findings | small (2-4h) | — |
| CORR-003 | **Add detector/adapter caching** — memoize Detect/IsInstalled/Status | 7 findings | small (2-4h) | — |
| P2-004 | **Pre-build Lipgloss styles** as package vars instead of per-frame creation | 1 finding | small (1h) | — |
| P2-005 | **Pre-render Logo** once at init time | 1 finding | small (30m) | — |

### 🟠 Short-term (Medium, High Leverage)

| ID | Action | Resolves | Effort |
|----|--------|----------|--------|
| CORR-004 | **Fix pipeline type assertions** — use proper interfaces, extract constant | 3 findings | medium (8-16h) |
| P4-001 | **Clean go.sum** of typosquat `go.yaml.in/yaml/v3` entry | 1 finding | small (30m) |
| P6-006 | **Add retry loop** to Windows installer (install.ps1) | 1 finding | small (1h) |
| P6-008 | **Separate Homebrew/Scoop tokens** with repo-scoped credentials | 1 finding | small (30m) |
| P6-009 | **Add structured logging** with `--verbose`/`--debug` flags | 1 finding | medium (4-8h) |

### 🟡 Backlog (Low, Incremental Improvement)

| ID | Action | Effort |
|----|--------|--------|
| CORR-001 | Split `adapters/common` into focused packages | large |
| CORR-005 | Implement audit command, post-release smoke tests | large |
| P3-003 | Remove unused `DefaultRegistry` init() pattern | small |
| P3-004 | Update `_template` to match real adapter patterns | small |
| P4-009 | Add golden test for ToolSelection screen | small |
| P6-011 | Enable `-race` on Windows CI | small |
| P6-013 | Align CHANGELOG.md with auto-generated release notes | small |

---

## Quick Wins (< 1 hour, high impact)

1. **P2-004 + P2-005**: Convert TUI styles to package-level vars + pre-render Logo → **30 min**, eliminates ~1.3 MB/s GC pressure at 60fps
2. **P4-001**: Run `go mod tidy` to clean typosquat entry → **5 min**, closes supply chain risk
3. **P6-001**: Remove `continue-on-error: true` from vulncheck + add to release.yml → **15 min**, closes CVE bypass
4. **P4-005 + P6-005**: Add `-coverprofile` to CI test command → **15 min**, enables coverage tracking
5. **P6-008**: Create separate token for Scoop → **10 min**, closes token sharing risk

---

## Findings by Phase

### 🔒 Security (70/100 — 4 findings)
- P1-001 (med): `copyFile` creates files with umask-dependent permissions
- P1-002 (med): Write-probe file created with default permissions
- P1-003 (low): Home directory path leaked in status output
- P1-004 (low): GitHub Action working-directory input unsanitized

**Strengths**: No command injection, no hardcoded secrets, no HTTP calls, no token exposure in config files.

### ⚡ Performance (0/100 — 12 findings)
- P2-001 (high): `Detect()` calls exec.LookPath without caching
- P2-002 (high): `IsInstalled()` uncached file reads in TUI at frame rate
- P2-003 (med): `Status()` reads version file without caching
- P2-004 (med): Lipgloss styles recreated every View() render
- P2-005 (med): Logo recomputed every render
- P2-006 (med): UninstallView calls IsInstalled() twice per frame
- P2-007~P2-012 (low/info): Redundant calls, wasteful allocations

**Strengths**: Lean dependency tree, embed.FS for templates, sync.WaitGroup + context cancellation for goroutines.

### 🏗️ Architecture (0/100 — 10 findings)
- P3-001 (high): `adapters/common` god package — 6+ unrelated concerns
- P3-002 (high): Unguarded type assertions in pipeline
- P3-003 (med): init() self-registration creates dead DefaultRegistry data
- P3-004 (med): `_template` adapter diverges from real patterns
- P3-005 (med): Global mutable state (CommandFiles, templateCache, DefaultRegistry)
- P3-006 (med): ToolAdapter too broad (11 methods)
- P3-007~P3-010 (low/info): File size, interface gaps, coupling notes

**Strengths**: No circular dependencies, `internal/` boundary well-respected, sub-role interfaces exist alongside ToolAdapter, adapter packages don't import each other.

### ✅ Quality (20/100 — 12 findings)
- P4-001 (high): Suspicious typosquat `go.yaml.in/yaml/v3` in go.sum
- P4-002 (med): Empty coverage files — no data collected
- P4-003 (med): govulncheck `continue-on-error: true`
- P4-004 (med): Magic number `64` repeated in 22+ files
- P4-005~P4-012 (low/info): Coverage gaps, complexity, lint exclusions, stale deps

**Strengths**: 594 `t.Parallel()` calls (nearly all tests parallel-safe), zero `//nolint` comments, 25/25 exported types have godoc, 15 golden test files, comprehensive error-path testing (705-line error test), Dependabot configured.

### 🔧 Operations (0/100 — 15 findings)
- P6-001 (high): Vulncheck non-blocking in CI, absent from release
- P6-002 (high): GitHub Action produces non-functional placeholder
- P6-003 (high): No post-release smoke test or artifact verification
- P6-004 (high): No documented rollback procedure
- P6-005~P6-010 (med): Coverage threshold, Windows retry, monolithic CI, shared tokens, no structured logging, no artifact verification
- P6-011~P6-015 (low/info): Windows race detector, cosign version check, CHANGELOG sync, release integration tests, observability

**Strengths**: Atomic file writes with rollback (Prepare→Apply→Verify→Rollback), SHA-256 + Cosign dual integrity verification, draft release review gate, proper secrets references, context-based cancellation, sentinel error types, dependabot for both Go and Actions.

---

## Score Trends

*No previous audit for comparison. This is the baseline.*

---

*Report generated by Sequoia v1.0.6 | Audit schema v1.0*
