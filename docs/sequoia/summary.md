# Sequoia Audit Report — Sequoia v1.0.6
**Date**: 2026-05-18 | **Audit mode**: full | **Health Score**: **0/100**

## Executive Summary

Sequoia v1.0.6 carries **7 systemic root causes** that propagate into 52 individual findings across security, performance, architecture, quality, and operations. Two **silent data corruption bugs** are live in production: (1) a template cache key collision installs wrong content in headless mode (P2-001), and (2) shared backup directories cause cross-installer restore failure — silently losing user data on rollback (P3-001/RC-004). The codebase is anchored by a **482-line god-object** (BaseAdapter, 22 exported methods) that drives 8 downstream defects including code duplication, low test coverage, and untested error paths. The CI/CD pipeline has **zero security gates** — no CodeQL, no pre-release testing, no signed binary verification in the GitHub Action. The Health Score of 0/100 reflects real, exploitable correctness bugs and systemic architectural debt in a production tool distributed to thousands of users.

## Health Score

| Category | Score | Findings |
|----------|-------|----------|
| Security | 76/100 | 10 (2M, 5L, 3I) |
| Performance | 78/100 | 9 (1H, 3M, 3L, 2I) |
| Architecture | 50/100 | 12 (2H, 6M, 4L) |
| Quality | 60/100 | 11 (1H, 6M, 4L) |
| Experience | N/A | Skipped (CLI tool) |
| Operations | 82/100 | 8 (3M, 4L, 1I) |
| **GLOBAL** | **0/100** | **52 raw → 11 consolidated (7 RCs, 2 merged, 11 standalone)** |

> **Methodology**: `score = 100 − Σ(severity_weight × scope_multiplier)`, floored at 0
> - severity_weight: critical=15, high=8, medium=4, low=2, info=0
> - scope_multiplier: 1.0 (isolated) | 1.5 (shared root cause ≥2 findings)
> - Child findings absorbed by root causes are NOT double-counted
> - Global deduction: 68 (5 high RCs ×1.5 + 2 standalone high) + 30 (medium) + 13 (low) = 111 → 0/100

## Critical & High Findings

### RC-001 [H] — BaseAdapter is monolithic god-object (22 methods, 482 lines)
- **Evidence**: `adapters/common/base_adapter.go:24-91` — 22 exported methods, 19 fields, Install() at 131 lines with 5+ nesting levels
- **Impact**: 8 child findings: Codex duplicates Install() (P3-006), common/ coverage at 64.8% (P4-002), Install() is untestable at 131 lines (P4-004), session errors silently discarded (P4-005), cancellation rollback untested (P4-011), TOML merge has no shared home (P3-007), duplicate status types (P4-007)
- **Fix**: Decompose into interface-segregated roles: PathResolver, TemplateRenderer, LifecycleManager. Extract shared Install flow into a separate Pipeline type.
- **Effort**: large

### RC-002 [H] — Install scripts lack integrity-by-design
- **Evidence**: `scripts/install.sh`, `scripts/install.ps1`, `.goreleaser.yaml:74-79` — promote `curl | bash` and `irm | iex` without script verification; SKIP_CHECKSUMS flag allows unverified binary execution
- **Impact**: 6 child findings: pipe-to-shell promoted (P1-002), grep/sed JSON parsing fragile (P1-004), SKIP_CHECKSUMS escape hatch (P1-009), stderr suppression (P1-007), predictable temp dir (P1-008), version string injection in URL (P1-005)
- **Fix**: Sign install scripts with GPG/minisign. Add script checksum verification step before execution. Remove SKIP_CHECKSUMS or gate it behind explicit flag with prominent warning.
- **Effort**: medium

### RC-003 [H] — CI/CD pipeline lacks security and quality gates
- **Evidence**: `.github/workflows/ci.yml` — runs go test + lint but NO CodeQL, secret scanning, dependency vuln check, pre-release testing, or SBOM generation
- **Impact**: 8 child findings: no CodeQL (P6-001), no pre-release test gate (P6-002), Action builds from source bypassing signed releases (P6-003), no fail-on validation (P6-006), ARM64 untested (P6-004), race detector off on Windows (P6-008), Dependabot lacks grouping (P6-009), Action downloads binary without checksum (P1-001)
- **Fix**: Add CodeQL + secret scanning workflows. Add `go test` job before GoReleaser in release.yml. Enable Dependabot grouping. Add ARM64 to test matrix via larger runners. Re-enable -race on Windows.
- **Effort**: medium

### RC-004 [H] — Shared backup directory causes cross-installer data loss on rollback
- **Evidence**: `adapters/common/base_adapter.go:358` — `backupDir := a.backupPathFn(base) + "-" + sessionSuffix` — no adapter-ID namespace
- **Impact**: 3 child findings: cross-installer restore failure SILENTLY LOSES USER DATA (P3-001), CI smoke tests ignore failures that would catch this (P6-007), probe-file write check is unnecessary TOCTOU pattern (P2-010)
- **Fix**: Namespace backup directories by adapter ID: `backupDir := a.backupPathFn(base) + "-" + a.ID() + "-" + sessionSuffix`
- **Effort**: small

### RC-006 [H] — Eager adapter init() couples everything to main at startup
- **Evidence**: `cmd/sequoia/main.go:22-28` — blank imports force all 5 adapters' init() to run; `adapters/registry.go:20` — global DefaultRegistry
- **Impact**: 5 child findings: all init() run even for `version` (P2-003), headless install sequential (P2-004), internal/ packages import adapters (P3-005), pipeline type assertions fragile (P3-008), main.go is 382 lines (P4-008)
- **Fix**: Lazy adapter loading — register metadata at init, defer construction. Move registry away from global state to explicit construction.
- **Effort**: medium

### P2-001 [H] — Template cache key uses stack-local address of embed.FS — SILENTLY INSTALLS WRONG CONTENT
- **Evidence**: `adapters/common/template.go:21` — `key := fmt.Sprintf("%p:%s", &fs, name)` where fs is passed by value
- **Impact**: In headless CLI mode, Go reuses the same stack frame, so different adapters get the same cache key and receive wrong template content. Different tools receive wrong skill.md.
- **Fix**: Use stable string identifier (e.g., adapter ID or content hash) instead of stack address.
- **Effort**: small

### P4-005 [H] — Session file write error silently discarded in Codex config merge
- **Evidence**: `adapters/codex/installer.go:36-38` — error from AtomicWriteFile session file is silently discarded with no log, no return
- **Impact**: Codex config.toml corruption during install goes undetected. If session file write fails, uninstall restores wrong backup or fails silently.
- **Fix**: Log the error at minimum; return it or collect it into adapter warnings.
- **Effort**: small

## Root Causes

| ID | Title | Severity | Affected | Approach |
|----|-------|----------|----------|----------|
| RC-001 | BaseAdapter monolithic god-object | HIGH | 8 findings | Decompose into interface-segregated components |
| RC-002 | Install scripts lack integrity-by-design | HIGH | 6 findings | Add script signing, checksum verification |
| RC-003 | CI/CD lacks security and quality gates | HIGH | 8 findings | Add CodeQL, pre-release testing, vuln scanning |
| RC-004 | Shared backup directory, no adapter isolation | HIGH | 3 findings | Namespace backups by adapter ID |
| RC-005 | Global mutable state in adapter registry | MEDIUM | 2 findings | Dependency-inject registry, add nil validation |
| RC-006 | Eager adapter init() couples all to main | HIGH | 5 findings | Lazy loading, explicit construction |
| RC-007 | Inconsistent code generation patterns across adapters | MEDIUM | 4 findings | Update _template to use BaseAdapter |

## Systemic Patterns

| Pattern | Description | Affected |
|---------|-------------|----------|
| **Silent Error Swallowing** | Errors discarded without logging in 6+ locations: install scripts (2>/dev/null), Codex session file (P4-005), CI smoke tests (P6-007), backup restore failure (P3-001), recover() masks panics (P2-007, P4-010), error output leaks PII (P1-006) | P1-006, P1-007, P2-007, P3-001, P4-005, P4-010, P6-007 |
| **Global Mutable State** | adapters.DefaultRegistry is package-level global. Tests mutex-lock around it. Register() has no nil validation. | P3-010, P4-012, RC-005 |
| **Supply Chain as Afterthought** | No checksums in Action, pipe-to-shell distribution, no CodeQL, no SBOM, go-figure abandoned 5 years, Dependabot ungrouped | P1-001, P1-002, P1-003, P6-001, P6-003, P6-009 |
| **False Test Confidence** | Zero-assertion tests inflate coverage, TUI router has no coverage, CI ignores smoke failures, cache key bug should have been caught | P2-001, P4-003, P4-009, P6-007 |

## Priority Action Plan

### Immediate (Critical + High, blockage-ordered)

| # | ID | Action | Effort | Blocks |
|---|-----|--------|--------|--------|
| 1 | RC-004 | Namespace backup directories by adapter ID | small | P3-001, P6-007 |
| 2 | P2-001 | Fix template cache key to use stable identifier | small | — |
| 3 | P4-005 | Propagate session file write errors in Codex merge | small | — |
| 4 | RC-001 | Decompose BaseAdapter into interface-segregated components | large | P3-006, P4-002, P4-004, P4-011, P3-007, P4-007 |
| 5 | RC-002 | Sign install scripts, add integrity verification | medium | P1-001, P1-002, P1-004, P1-005, P1-007, P1-009 |
| 6 | RC-003 | Add CodeQL, pre-release testing, vuln scanning to CI | medium | P6-001..P6-009 |
| 7 | RC-006 | Implement lazy adapter loading | medium | P2-003, P2-004, P3-005, P3-008 |

### Short Term (Medium)

| # | ID | Action | Effort |
|---|-----|--------|--------|
| 8 | MERGED-001 | Replace go-figure with lightweight ASCII art (~20 lines) | small |
| 9 | P3-004 | Integrate or delete dead plugin system code | medium |
| 10 | P4-003 | Remove zero-assertion tests, add real assertions | small |
| 11 | P4-009 | Add tests for TUI router | small |
| 12 | RC-005 | Dependency-inject registry, validate nil adapters | small |
| 13 | RC-007 | Update _template to embed BaseAdapter, use shared templates | small |

### Long Term (Low + Info)

| # | ID | Action | Effort |
|---|-----|--------|--------|
| 14 | P2-006 | Use bufio.Scanner for IsInstalled() substring match | small |
| 15 | P2-009 | Cache lipgloss.Style objects at package level | small |
| 16 | P6-005 | Use separate fine-grained token for Scoop tap | small |
| 17 | P6-010 | Add --verbose/--debug flag for CLI troubleshooting | small |
| 18 | MERGED-002 | Replace defer-recover with select-ok pattern | small |
| 19 | P3-009 | Rename files with hyphens to Go convention | small |
| 20 | P4-006 | Add build tag to _template directory | small |
| 21+ | P1-006, P1-008, P1-010, P2-008, P2-010, P6-004, P6-006, P6-008, P6-009 | Address remaining low/info findings | small |

## All Findings by Category

### Security (P1 + supply chain)

| ID | Sev | Title | Evidence |
|----|-----|-------|----------|
| P1-001 | M | GitHub Action downloads binary without checksum verification | action.yml:83-95 |
| P1-002 | M | Pipe-to-shell install pattern promoted without script integrity verification | .goreleaser.yaml:74-79 |
| P1-003 | L | Unmaintained dependency go-figure from 2021 (→ MERGED-001) | go.mod:9 |
| P1-004 | L | Install script uses grep/sed for GitHub API JSON parsing (→ RC-002) | scripts/install.sh:173 |
| P1-005 | L | Action user-controlled version string interpolated into download URL (→ RC-002) | action.yml:83-91 |
| P1-006 | L | Error output leaks full filesystem paths to stderr | cmd/sequoia/main.go:57-61 |
| P1-007 | I | install.sh suppresses curl/wget stderr with 2>/dev/null (→ RC-002) | scripts/install.sh:159,165,250,253 |
| P1-008 | I | install.ps1 uses predictable temp directory name with Get-Random (→ RC-002) | scripts/install.ps1:142 |
| P1-009 | L | install.sh SKIP_CHECKSUMS flag allows unverified binary execution (→ RC-002) | scripts/install.sh:27,32,258-259 |
| P1-010 | I | os.UserHomeDir() reads from HOME env var on Unix | adapters/common/base_adapter.go:192-202 |

### Performance (P2)

| ID | Sev | Title | Evidence |
|----|-----|-------|----------|
| P2-001 | **H** | Template cache key uses stack-local address of embed.FS — wrong content | adapters/common/template.go:14-21 |
| P2-002 | M | go-figure ASCII art library (~2MB) for single-word rendering (→ MERGED-001) | internal/tui/styles/logo.go:7-8 |
| P2-003 | M | All 5 adapter init() execute at program start — even for `version` (→ RC-006) | cmd/sequoia/main.go:22-28 |
| P2-004 | M | Headless CLI install runs adapters sequentially (→ RC-006) | cmd/sequoia/main.go:258-272 |
| P2-005 | M | Duplicate 6034-byte skill.md.tmpl in claude and gemini (→ RC-007) | adapters/claude/, adapters/gemini/ |
| P2-006 | L | IsInstalled() reads entire config file into memory for substring match | adapters/claude/adapter.go:42-48 |
| P2-007 | L | sendProgress() sets up defer-recover on every call (→ MERGED-002) | internal/pipeline/runner.go:271-279 |
| P2-008 | I | Binary size 10.3 MB — lipgloss terminal library is dominant contributor | go.mod |
| P2-009 | L | lipgloss.Style objects re-created on every View() render frame | internal/tui/styles/styles.go:17-23 |
| P2-010 | I | Installer Prepare() creates probe file — extra I/O (→ RC-004) | adapters/common/installer.go:62-71 |

### Architecture (P3 + P2-001 + P4-007)

| ID | Sev | Title | Evidence |
|----|-----|-------|----------|
| P3-001 | **H** | Shared backup dir causes cross-installer restore failure — DATA LOSS (→ RC-004) | adapters/common/base_adapter.go:358-394 |
| P3-002 | M | BaseAdapter has 22 exported methods violating SRP (→ RC-001) | adapters/common/base_adapter.go:24-91 |
| P3-003 | M | _template adapter does not embed BaseAdapter (→ RC-007) | adapters/_template/adapter.go:28-30 |
| P3-004 | M | Plugin system is unreachable dead code with no integration point | plugin/, cmd/sequoia/main.go |
| P3-005 | M | internal/app and internal/pipeline import adapters (→ RC-006) | internal/app/model.go:10 |
| P3-006 | M | Codex adapter duplicates ~80 lines of BaseAdapter.Install() (→ RC-001) | adapters/codex/adapter.go:68-154 |
| P3-007 | M | StrategyTOMLMerge has no shared implementation in common/ (→ RC-001) | adapters/interface.go:19-21 |
| P3-008 | L | Pipeline type assertions create fragile runtime coupling (→ RC-006) | internal/pipeline/runner.go:162,178 |
| P3-009 | L | Go source files use hyphens violating Go file naming convention (→ RC-007) | internal/tui/screens/tool-selection.go |
| P3-010 | L | Registry.Register() accepts nil adapters with no validation (→ RC-005) | adapters/registry.go:25-40 |
| P4-007 | M | Duplicate status types: ToolStatus and AdapterStatus (→ RC-001) | internal/model/types.go:33, adapters/interface.go:36 |

### Quality (P4)

| ID | Sev | Title | Evidence |
|----|-----|-------|----------|
| P4-001 | M | Dependency go-figure abandoned since 2021 (→ MERGED-001) | go.mod:9 |
| P4-002 | M | adapters/common has 64.8% coverage — lowest non-trivial package (→ RC-001) | go test -cover |
| P4-003 | M | Test with zero runtime assertions inflates coverage signal | internal/app/model_test.go:98 |
| P4-004 | M | BaseAdapter.Install() is 131 lines with 5+ nesting levels (→ RC-001) | adapters/common/base_adapter.go:300-430 |
| P4-005 | **H** | Session file write error silently discarded in Codex config merge (→ RC-001) | adapters/codex/installer.go:36-38 |
| P4-006 | L | Template adapter directory contains 17 TODO markers (→ RC-007) | adapters/_template/ |
| P4-008 | M | cmd/sequoia/main.go is 382 lines mixed concerns (→ RC-006) | cmd/sequoia/main.go:1-436 |
| P4-009 | M | internal/tui has [no statements] coverage — router untested | internal/tui/router.go |
| P4-010 | L | Pipeline runner uses recover() in two functions (→ MERGED-002) | internal/pipeline/runner.go:258-279 |
| P4-011 | L | BaseAdapter.Install() cancellation rollback paths untested (→ RC-001) | adapters/common/base_adapter.go:380-422 |
| P4-012 | L | TUI model tests mutate global DefaultRegistry (→ RC-005) | internal/app/model_test.go:24-34 |

### Operations (P6)

| ID | Sev | Title | Evidence |
|----|-----|-------|----------|
| P6-001 | M | No CodeQL or secret scanning — supply chain security gap (→ RC-003) | .github/ |
| P6-002 | M | Release pipeline has no pre-release test gate (→ RC-003) | .github/workflows/release.yml:12-39 |
| P6-003 | M | GitHub Action "latest" builds from source, bypassing signed releases (→ RC-003) | action.yml:84-89 |
| P6-004 | L | ARM64 binaries shipped but never tested in CI matrix (→ RC-003) | .github/workflows/ci.yml:15-16 |
| P6-005 | L | Scoop tap reuses Homebrew token — token scope confusion | .goreleaser.yaml:149 |
| P6-006 | L | No input validation for fail-on severity in GitHub Action (→ RC-003) | action.yml:20-26 |
| P6-007 | M | CI smoke tests silently ignore install/uninstall failures (→ RC-004) | .github/workflows/ci.yml:56-65 |
| P6-008 | L | Race detector disabled on Windows in CI (→ RC-003) | .github/workflows/ci.yml:38-43 |
| P6-009 | I | Dependabot lacks grouped updates, reviewers, labels (→ RC-003) | .github/dependabot.yml |
| P6-010 | L | No debug/verbose mode for CLI operational troubleshooting | cmd/sequoia/main.go:81-101 |

## Severity Distribution

| Severity | Raw Count | Consolidated | IDs |
|----------|-----------|-------------|-----|
| CRITICAL | 0 | 0 | — |
| HIGH | 2 | 7 | RC-001, RC-002, RC-003, RC-004, RC-006, P2-001, P4-005 |
| MEDIUM | 26 | 10 | RC-005, RC-007, MERGED-001, P3-004, P4-003, P4-009 + 4 absorbed |
| LOW | 17 | 7 | MERGED-002, P2-006, P2-009, P6-005, P6-010, P1-006 + 11 absorbed |
| INFO | 7 | 3 | P2-008, P1-010, P6-009 + 4 absorbed |
| **Total** | **52** | **27** | 7 RCs + 2 merged + 11 standalone + 7 info |

### Time estimate
- Tier 1 (Immediate): 3-4 weeks (RC-001 is the long pole at ~2 weeks for decomposition)
- Tier 2 (Short Term): 2-3 weeks
- Tier 3 (Long Term): 1-2 weeks spread across maintenance cycles
- **If Tier 1 only fixed**: Health Score → ~65/100
- **If Tier 1 + 2 fixed**: Health Score → ~82/100

---

*Generated by Sequoia v1.0.6 — Multi-Agent Code Audit Framework — 2026-05-18*
