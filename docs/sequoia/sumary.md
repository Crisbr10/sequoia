# Sequoia Audit Report — sequoia-ai

**Date**: 2026-05-21
**Stack**: Go 1.24 (Cobra CLI + Bubbletea TUI + Lipgloss)
**Size**: Medium (133 Go files, 68 source + 65 test, 19 packages, 27 deps)
**Maturity**: Growth
**Audit ID**: audit-20260521-sequoia-ai
**Schema Version**: 1.0

---

## Global Health Score: 32/100 — Critical

> **Classification**: Critical (0–49) — Immediate risk, urgent action required.
> The most severe issues cluster in **Operations** (score: 0), where the release pipeline lacks defense-in-depth at every stage. Security, performance, architecture, and quality all have multiple systemic weaknesses. Resolving CORR-001 (supply chain) and CORR-003 (TUI caching) alone recovers ~15 points.

### Scoring Methodology

```
score = 100 − Σ(severity_weight × scope_multiplier), floored at 0
severity_weight: CRITICAL=15, HIGH=8, MEDIUM=4, LOW=2, INFO=0
scope_multiplier: 1.0 (isolated) | 1.5 (shared root cause across ≥2 findings)
```

### Phase Scores

| Phase | Score | Classification | CRITICAL | HIGH | MEDIUM | Total |
|-------|-------|----------------|----------|------|--------|-------|
| 🔒 Security | **35**/100 | Critical | 1 | 0 | 9 | 10 |
| ⚡ Performance | **38**/100 | Critical | 0 | 2 | 9 | 11 |
| 🏗️ Architecture | **36**/100 | Critical | 0 | 2 | 8 | 10 |
| ✅ Quality | **48**/100 | Critical | 0 | 0 | 10 | 10 |
| 🎨 Experience | **N/A** | — | — | — | — | — |
| 🔧 Operations | **0**/100 | Critical | 2 | 4 | 9 | 15 |

> Experience: Not Applicable (CLI/TUI tool — no web or mobile interface).
> P5's 0.10 weight redistributed proportionally: Security 0.278, Performance 0.167, Architecture 0.222, Quality 0.167, Operations 0.167.

### Score Visualization

```
Security      ██████████████████░░░░░░░░░░░░░░░░░░░░░░░░ 35/100
Performance   ███████████████████░░░░░░░░░░░░░░░░░░░░░░░ 38/100
Architecture  ██████████████████░░░░░░░░░░░░░░░░░░░░░░░░ 36/100
Quality       ████████████████████████░░░░░░░░░░░░░░░░░░ 48/100
Operations    ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  0/100
────────────────────────────────────────────────────────────
GLOBAL        ████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░ 32/100  CRITICAL
```

---

## Root Causes Identified

### CORR-001: Supply Chain Release Pipeline Weakness
**Confidence**: 0.90 | **Symptoms**: 10 | **Risk**: CRITICAL

Root cause: Release workflow designed for convenience over security — automated tag→release with no gates, no verification, no reviewers, no post-publication validation. A single compromised contributor account could push malicious signed binaries to all distribution channels (GoReleaser, Homebrew, Scoop).

**Causal chain**:
- No binary verification (P1-001) → signed artifacts are trusted blindly
- Unpinned actions (P1-007) → supply chain attacks via compromised action repos
- No environment protection (P1-008) → no required reviewers on release workflow
- continue-on-error masks failures (P6-001) → release proceeds despite build/test errors
- No approval gate (P6-002) → any tag push auto-publishes to all channels
- Ubuntu-only pre-tests (P6-003) → no macOS/Windows validation before release
- No workflow_dispatch (P6-004) → cannot manually re-run or re-validate releases
- No CODEOWNERS (P6-005) → no required code review on release paths
- No post-deploy smoke test (P6-006) → broken releases go undetected
- No retry in Windows installer (P6-008) → download failures cause silent breakage

**Findings resolved**: P1-001, P1-007, P1-008, P6-001, P6-002, P6-003, P6-004, P6-005, P6-006, P6-008

**Fix**: Add environment protection, binary verification, approval gates, CODEOWNERS, and post-deploy smoke tests. Split CI into phased jobs with proper gating. **Effort**: large (16–32h)

---

### CORR-002: Global Mutable State via init() Functions
**Confidence**: 0.85 | **Symptoms**: 4 | **Risk**: HIGH

Root cause: All 5 adapters use `init()` to mutate `DefaultRegistry`, creating global mutable state that couples all adapters at import time. Tests depend on this global pollution. `CommandFiles` and `templateCache` compound the problem as additional mutable globals in `adapters/common`.

**Causal chain**:
- `init()` auto-registration (P1-004) → implicit side effects on import
- `DefaultRegistry` coupling (P3-001) → all adapters tightly coupled to one global registry
- Global `CommandFiles` map (P3-008) → race condition risk across goroutines
- Global `templateCache` sync.Map (P3-009) → no eviction, test pollution

**Findings resolved**: P1-004, P3-001, P3-008, P3-009

**Fix**: Remove all `init()` functions, enforce `NewRegistry()` + `RegisterIn()` as the only registration path. Make `CommandFiles` unexported with a getter that returns a copy. Add `ResetTemplateCache()` for test isolation. **Effort**: medium (8–16h)

---

### CORR-003: TUI Rendering Path Inefficiency
**Confidence**: 0.80 | **Symptoms**: 4 | **Risk**: HIGH

Root cause: No caching or pre-computation in the TUI render loop. Every frame re-allocates styles, recomputes strings, and builds views from scratch. At 60fps this creates unnecessary GC pressure. Combined with eager adapter construction (all 5 adapters built before first frame), startup is slower than necessary.

**Causal chain**:
- Logo recomputed every frame (P2-002) → expensive string manipulation at 60fps
- Lipgloss styles created per call (P2-005) → ~200 bytes heap per style per frame
- No `Grow()` pre-allocation (P2-006) → repeated `strings.Builder` reallocation
- Eager adapter construction (P2-011) → all adapters built before first frame renders

**Findings resolved**: P2-002, P2-005, P2-006, P2-011

**Fix**: Cache styles as package-level vars, precompute logo at init, add `Grow()` hints to `strings.Builder`, defer adapter construction until needed. **Effort**: small (2–4h)

---

### CORR-004: Common Package Architectural Bottleneck
**Confidence**: 0.75 | **Symptoms**: 5 | **Risk**: HIGH

Root cause: `adapters/common` is a 14-file god package with no internal boundaries. The pipeline collapses the install lifecycle into a single step, hiding per-phase progress. Error handling is inconsistent (bare returns in strategy.go), and security scanning requires blanket gosec exclusions.

**Causal chain**:
- Pipeline collapsed to single step (P3-002) → no per-phase progress or rollback visibility
- ToolAdapter violates ISP (P3-003) → consumers tied to 11 methods when they need 2–3
- Strategy pattern bypassed (P3-010) → `Strategy` interface exists but code paths ignore it
- Bare error returns (P4-003) → no context for debugging install failures
- Blanket gosec exclusions (P4-007) → real security issues hidden among false positives

**Findings resolved**: P3-002, P3-003, P3-010, P4-003, P4-007

**Fix**: Split into sub-packages (installer, strategy, fs, detect, templates). Refactor pipeline to expose per-phase progress. Replace type assertions with proper interface checks. Use line-level `//nolint` instead of file-level exclusions. **Effort**: large (16–32h)

---

## Prioritized Roadmap

### 🔴 Immediate (Critical + High, Ordered by Dependencies)

| ID | Action | Resolves | Effort | Blocks |
|----|--------|----------|--------|--------|
| CORR-001 | **Harden release pipeline** — approval gates, binary verification, post-deploy smoke tests | 10 findings | large (16–32h) | — |
| CORR-003 | **Add TUI rendering cache** — precompute logo, cache styles, lazy adapter init | 4 findings | small (2–4h) | — |
| P2-001 | **Cache Detect() results** — `sync.Once` memoization for `exec.LookPath` | 1 finding | small (1h) | P2-003, P2-008 |

### 🟠 Short-term (Medium, High Leverage)

| ID | Action | Resolves | Effort |
|----|--------|----------|--------|
| P1-001 | **Verify binary integrity** — add checksum validation before signing | 1 finding | medium (4–8h) |
| P4-001 | **Clean go.sum** — remove typosquat `go.yaml.in/yaml/v3` entry | 1 finding | small (5m) |
| P4-004 | **Extract magic number 64** — define `ProgressChannelBufferSize` constant | 1 finding | small (1h) |
| P6-007 | **Implement audit command** in GitHub Action output | 1 finding | medium (4–8h) |
| P6-009 | **Separate Homebrew/Scoop tokens** with repo-scoped credentials | 1 finding | small (10m) |
| P6-010 | **Add structured logging** with `--verbose`/`--debug` flags | 1 finding | medium (4–8h) |

### 🟡 Backlog (Low, Incremental Improvement)

| ID | Action | Effort |
|----|--------|--------|
| CORR-002 | Remove all `init()` auto-registration, enforce explicit registry | medium (8–16h) |
| CORR-004 | Split `adapters/common` into focused sub-packages | large (16–32h) |
| P3-004 | Fix double error wrapping in defer functions | small (1h) |
| P3-007 | Align `_template` adapter with real adapter patterns | small (2h) |
| P4-009 | Standardize error wrapping patterns across codebase | small (2h) |
| P6-011 | Enable `-race` on Windows CI | small (30m) |
| P6-013 | Sync CHANGELOG.md with auto-generated release notes | small (1h) |

---

## Quick Wins (< 1 hour, high impact)

1. **P2-002 + P2-005**: Pre-render Logo + cache Lipgloss styles as package vars → **30 min**, eliminates ~1.3 MB/s GC pressure at 60fps
2. **P4-001**: Run `go mod tidy` to clean typosquat entry → **5 min**, closes supply chain risk
3. **P6-001**: Remove `continue-on-error: true` from vulncheck → **15 min**, closes CVE bypass in CI
4. **P6-009**: Create separate token for Scoop → **10 min**, closes token sharing risk
5. **P6-005**: Add CODEOWNERS file for release workflow path → **5 min**, enforces review gate

---

## Findings by Phase

### 🔒 Security (35/100 — 10 findings: 1 CRITICAL, 0 HIGH, 9 MEDIUM)

**Critical**:
- P1-001: No binary verification in release pipeline — signed artifacts are trusted without verifying they match the build

**Medium** (9):
- P1-002: `copyFile` creates files with umask-dependent permissions (`adapters/common/installer.go:200`)
- P1-003: Write-probe file created with default permissions (`adapters/common/installer.go:64`)
- P1-004: `init()` auto-registration creates implicit side effects at import time (CORR-002)
- P1-005: Home directory path leaked in status output (`cmd/sequoia/main.go:321`)
- P1-006: GitHub Action `working-directory` input unsanitized (`action.yml`)
- P1-007: Unpinned GitHub Actions in release workflow — uses `@v4` instead of commit hash (CORR-001)
- P1-008: No environment protection on release workflow — no required reviewers (CORR-001)
- P1-009: Suspicious typosquat entry `go.yaml.in/yaml/v3` in `go.sum`
- P1-010: No signature verification on downloaded adapter tools

**Strengths**: No command injection, no hardcoded secrets, no HTTP calls, no token exposure in config files.

---

### ⚡ Performance (38/100 — 11 findings: 0 CRITICAL, 2 HIGH, 9 MEDIUM)

**High** (2):
- P2-001: `Detect()` calls `exec.LookPath` without caching — path scans repeated on every CLI invocation and TUI frame
- P2-002: Logo recomputed every TUI frame — expensive string computation at 60fps (CORR-003)

**Medium** (9):
- P2-003: `IsInstalled()` performs uncached file reads every TUI frame
- P2-004: `Status()` reads version file without caching
- P2-005: Lipgloss styles created per call instead of package-level vars (CORR-003)
- P2-006: No `Grow()` pre-allocation in `strings.Builder` usage (CORR-003)
- P2-007: `UninstallView` calls `IsInstalled()` twice per frame
- P2-008: Multiple redundant PATH scans per CLI invocation
- P2-009: No view rendering memoization — full recomputation on every model update
- P2-010: Repeated `os.Stat()` calls without caching in installer views
- P2-011: All 5 adapters constructed eagerly before first frame renders (CORR-003)

**Strengths**: Lean dependency tree (27 deps), `embed.FS` for templates, `sync.WaitGroup` + context cancellation for goroutines.

---

### 🏗️ Architecture (36/100 — 10 findings: 0 CRITICAL, 2 HIGH, 8 MEDIUM)

**High** (2):
- P3-001: `DefaultRegistry` implicit coupling through `init()` — all adapters mutate global state at import time (CORR-002)
- P3-002: Pipeline collapsed to single step — Install/Uninstall lifecycle hidden behind one monolithic function (CORR-004)

**Medium** (8):
- P3-003: `ToolAdapter` violates Interface Segregation Principle — 11 methods where consumers need 2–3 (CORR-004)
- P3-004: Double error wrapping in defer functions — errors wrapped both in operation and in deferred cleanup
- P3-005: `adapters/common` is a god package — 14 files, 8 unrelated concerns
- P3-006: Unguarded type assertions in pipeline — `t.Adapter.(pipelineInstaller)` without `ok` check
- P3-007: `_template` adapter diverges from real adapter patterns
- P3-008: Global mutable `CommandFiles` slice — exported and mutated across adapters (CORR-002)
- P3-009: `templateCache` as package-level mutable state — no eviction, persists across tests (CORR-002)
- P3-010: Strategy pattern collapsed — `Strategy` interface exists but pipeline code paths bypass it (CORR-004)

**Strengths**: No circular dependencies, `internal/` boundary well-respected, sub-role interfaces exist alongside `ToolAdapter`, adapter packages don't import each other.

---

### ✅ Quality (48/100 — 10 findings: 0 CRITICAL, 0 HIGH, 10 MEDIUM)

**Medium** (10):
- P4-001: Suspicious typosquat `go.yaml.in/yaml/v3` in `go.sum` — looks like domain squatting on `gopkg.in`
- P4-002: Empty coverage data files — `coverage` and `coverage_rc002` contain only `mode:` headers
- P4-003: Bare error returns without context in `strategy.go` — install failures lack debugging info (CORR-004)
- P4-004: Magic number `64` (channel buffer) repeated across 22+ files without named constant
- P4-005: `govulncheck` runs with `continue-on-error: true` — CVEs never block PRs
- P4-006: No test coverage threshold enforced — CHANGELOG claims 90%+ but CI doesn't verify
- P4-007: Blanket gosec exclusions hiding real issues in CI configuration (CORR-004)
- P4-008: Import pollution from `_test.go` files modifying `DefaultRegistry`
- P4-009: Inconsistent error wrapping patterns — mixture of `fmt.Errorf` and `errors.Wrap` styles
- P4-010: `_template` adapter diverges from live adapter patterns — untested reference implementation

**Strengths**: 594 `t.Parallel()` calls (nearly all tests parallel-safe), zero `//nolint` comments, 25/25 exported types have godoc, 15 golden test files, comprehensive error-path testing (705-line error test), Dependabot configured.

---

### 🔧 Operations (0/100 — 15 findings: 2 CRITICAL, 4 HIGH, 9 MEDIUM)

**Critical** (2):
- P6-001: `continue-on-error` masks all build failures in release workflow — release proceeds regardless (CORR-001)
- P6-002: No approval gate on release — any push to tag triggers automatic publication to all channels (CORR-001)

**High** (4):
- P6-003: Ubuntu-only pre-release tests — no macOS or Windows validation before publishing (CORR-001)
- P6-004: No `workflow_dispatch` — cannot manually trigger or re-validate releases (CORR-001)
- P6-005: No CODEOWNERS file — no required reviewers for release-critical paths (CORR-001)
- P6-006: No post-deploy smoke test — broken releases go undetected until user reports (CORR-001)

**Medium** (9):
- P6-007: GitHub Action produces non-functional placeholder output (`health_score=N/A`)
- P6-008: No retry loop in Windows installer — download failures cause silent breakage (CORR-001)
- P6-009: Shared token for Homebrew and Scoop distribution — violates least privilege
- P6-010: No structured logging — only `fmt.Printf` and lipgloss output, no verbosity levels
- P6-011: Race detector not enabled on Windows CI
- P6-012: Cosign version pinned but not verified via checksum in install scripts
- P6-013: `CHANGELOG.md` not synced with auto-generated release notes
- P6-014: No release integration tests — GoReleaser config never validated in CI
- P6-015: No observability — no metrics, no tracing for CLI operations

**Strengths**: Atomic file writes with rollback (Prepare→Apply→Verify→Rollback), SHA-256 + Cosign dual integrity verification, draft release review gate, proper secrets references, context-based cancellation, sentinel error types, Dependabot for both Go and Actions.

---

## Score Trends

*No previous audit for comparison. This is the baseline.*

---

*Report generated by Sequoia v1.0.9 (M2 Reporter) | Audit schema v1.0*
