# Quality Tasks — sequoia-ai

**Score**: 48/100 (Critical) | **Findings**: 10 (0 CRITICAL, 0 HIGH, 10 MEDIUM) | **Audit ID**: audit-20260521-sequoia-ai

---

## 🟡 MEDIUM Findings

### P4-001: Remove typosquat package from go.sum

**Problem**: `go.sum` contains `go.yaml.in/yaml/v3` — NOT the legitimate `gopkg.in/yaml.v3`. This domain (`yaml.in`) appears to be a typosquatting attempt. While no source code imports it, its presence in go.sum is a supply chain risk.

**Evidence**:
- `go.sum` includes an entry for `go.yaml.in/yaml/v3`
- `grep` confirms zero source code imports of this package
- The entry is likely stale from a transitive dependency that was later removed
- `go mod tidy` should clean it, but hasn't been run recently

**Fix**: Run `go mod tidy`. Verify removal. Add `go mod verify` to CI lint job.

**Acceptance Criteria**:
- [ ] `go.yaml.in/yaml/v3` removed from `go.sum`
- [ ] `go mod tidy` produces no diff (clean state)
- [ ] `go mod verify` added to CI lint job
- [ ] No source code imports the suspicious package

**Effort**: small (<5m) | **Risk**: medium | **Blocks**: none

---

### P4-002: Generate actual coverage data

**Problem**: `coverage` and `coverage_rc002` files in the repo root contain only `mode:` headers — zero coverage data. These are misleading artifacts suggesting coverage exists when it doesn't.

**Evidence**:
- `coverage` file: 1 line, only `mode: set`
- `coverage_rc002` file: 1 line, only `mode: set`
- CI test command omits `-coverprofile` flag
- Coverage has never been collected in CI

**Fix**: Delete the empty coverage files. Add `.gitignore` entry. Add `make coverage` or documented command. Generate initial baseline report.

**Acceptance Criteria**:
- [ ] Empty `coverage` and `coverage_rc002` files deleted
- [ ] `coverage.out` and `coverage.html` added to `.gitignore`
- [ ] Documented command: `go test -coverprofile=coverage.out -covermode=atomic ./...`
- [ ] Initial coverage baseline documented in CHANGELOG or docs
- [ ] CI step to generate and upload coverage artifact

**Effort**: small (<1h) | **Risk**: low | **Blocks**: P4-006 (coverage threshold)

---

### P4-003: Add error context to bare error returns

**Problem**: `adapters/common/strategy.go` returns bare errors without context in several functions. Install failures lack debugging information — callers can't determine what operation failed or why.

**Root Cause**: CORR-004 (Common Package Architectural Bottleneck)

**Evidence**:
- Pattern: `return err` instead of `return fmt.Errorf("installing %s: %w", name, err)`
- `strategy.go` handles file operations (copy, stage, atomic write) — errors need path context
- Without context, error messages are opaque: "permission denied" without saying which file

**Fix**: Wrap all errors in `strategy.go` with operation context. Include file paths, adapter name, and the operation being performed.

**Acceptance Criteria**:
- [ ] All bare `return err` in `strategy.go` replaced with contextual wrapping
- [ ] Error messages include: operation name, file path, adapter identifier
- [ ] `fmt.Errorf` used consistently (not `errors.Wrap` or custom wrappers)
- [ ] Test verifies error messages contain expected context

**Effort**: small (<2h) | **Risk**: low | **Blocks**: CORR-004

---

### P4-004: Extract magic number 64 as named constant

**Problem**: Channel buffer capacity `64` is hardcoded across 22+ files including `internal/app/model.go:48,128`. The value has no semantic name and changing it requires touching every occurrence.

**Evidence**:
- `make(chan model.ProgressMsg, 64)` in 22+ locations (source + test)
- Comment at `model.go:48` explains the purpose but the constant isn't defined
- Tests hardcode the same magic value, making them coupled to the implementation

**Fix**: Define `const ProgressChannelBufferSize = 64` in `internal/model/types.go` (near existing `ProgressMsg` definition). Replace all occurrences.

**Acceptance Criteria**:
- [ ] `ProgressChannelBufferSize` constant defined in `internal/model/types.go`
- [ ] All `make(chan model.ProgressMsg, 64)` references use the constant
- [ ] Test occurrences also use the constant
- [ ] Comment preserved alongside constant definition

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

### P4-005: Make govulncheck blocking in CI

**Problem**: `.github/workflows/ci.yml:30-31` sets `continue-on-error: true` on govulncheck. CVEs never block PRs or releases. The check runs but its result is ignored.

**Evidence**:
- `continue-on-error: true` on the govulncheck step
- No downstream steps check the vulncheck result
- Release workflow (`release.yml`) has no vulncheck step at all

**Fix**: Remove `continue-on-error: true`. Add vulncheck step to `release.yml` (before GoReleaser). Verify exit code behavior.

**Acceptance Criteria**:
- [ ] `continue-on-error: true` removed from ci.yml vulncheck step
- [ ] Vulncheck step added to release.yml (before GoReleaser)
- [ ] Verified that `govulncheck@v1.1.4` exits non-zero on CVEs
- [ ] Documented in CI troubleshooting if baseline CVEs exist

**Effort**: small (<30m) | **Risk**: low (may surface existing CVEs) | **Blocks**: none

---

### P4-006: Enforce coverage threshold in CI

**Problem**: No coverage threshold is enforced. CHANGELOG.md claims 90%+ coverage but CI doesn't verify it. Coverage can silently regress without detection.

**Fix**: Add `-coverprofile=coverage.out -covermode=atomic` to CI test command. Add step that checks coverage against minimum threshold (start at 70%, raise later).

**Acceptance Criteria**:
- [ ] CI test command includes `-coverprofile=coverage.out -covermode=atomic`
- [ ] Coverage threshold check added: fail if < 70%
- [ ] Windows CI uses `-coverprofile` without `-covermode=atomic` (not supported on Windows)
- [ ] Coverage artifact uploaded for review
- [ ] Threshold documented as starting value to be raised

**Effort**: small (<1h) | **Risk**: low | **Blocks**: P4-002 (needs coverage data first)

---

### P4-007: Replace blanket gosec exclusions with line-level nolint

**Problem**: CI gosec configuration uses file-level exclusions that hide real security issues. A new security-relevant code added to an excluded file gets silently ignored.

**Root Cause**: CORR-004 (Common Package Architectural Bottleneck)

**Evidence**:
- gosec configuration excludes entire files or patterns
- Security-relevant code in excluded files is never scanned
- Future code added to excluded files inherits the exclusion

**Fix**: Remove blanket file exclusions. Add line-level `//nolint:gosec` with justification comments only where false positives are confirmed. This ensures new code is always scanned.

**Acceptance Criteria**:
- [ ] File-level gosec exclusions removed
- [ ] Line-level `//nolint:gosec` added with justification comments
- [ ] CI gosec step still passes with line-level nolints
- [ ] New code in previously excluded files is scanned

**Effort**: medium (4-8h) | **Risk**: medium (may surface false positives) | **Blocks**: CORR-004

---

### P4-008: Fix test import pollution from DefaultRegistry

**Problem**: `_test.go` files in adapter packages modify `adapters.DefaultRegistry` during tests. Since `DefaultRegistry` is global mutable state, tests in one package can affect tests in another when run in the same process.

**Fix**: Tests should create their own registries instead of modifying the global one. This is blocked by CORR-002 (init() removal) — once DefaultRegistry is removed, tests naturally use explicit registries.

**Acceptance Criteria**:
- [ ] No test modifies `adapters.DefaultRegistry`
- [ ] All tests create local registries via `NewRegistry()`
- [ ] Test order independence verified (run with `-shuffle=on`)
- [ ] Partially resolved by CORR-002 (DefaultRegistry removal)

**Effort**: small (<2h, after CORR-002) | **Risk**: low | **Blocks**: CORR-002

---

### P4-009: Standardize error wrapping patterns

**Problem**: Codebase uses mixed error wrapping styles: `fmt.Errorf("context: %w", err)` in some places, bare `return err` in others, and `errors.New` for simple messages. No consistent pattern.

**Fix**: Standardize on `fmt.Errorf("operation %s: %w", name, err)` with operation context. Use `errors.New` or `fmt.Errorf` (without `%w`) only for sentinel errors. Use `errors.Is`/`errors.As` for sentinel checking.

**Acceptance Criteria**:
- [ ] Error wrapping style documented in CONTRIBUTING.md
- [ ] Bare `return err` replaced with contextual wrapping
- [ ] Sentinel errors (`var ErrX = errors.New(...)`) used for checkable conditions
- [ ] `errors.Is`/`errors.As` used for sentinel checking in callers

**Effort**: small (<2h) | **Risk**: low | **Blocks**: none

---

### P4-010: Add tests to _template adapter

**Problem**: `adapters/_template` has no tests. It's supposed to be the reference implementation for new adapter authors, but without tests, there's no verification it works correctly.

**Fix**: Add minimal test suite: compilation test, initialization test, registration test. Document expected patterns for adapter authors.

**Acceptance Criteria**:
- [ ] `_template/template_test.go` added with:
  - Test that adapter compiles and initializes without error
  - Test that `GetTemplateInfo()` returns expected values
  - Test that `Install()` succeeds in mock environment
- [ ] Tests serve as documentation: adapter authors can read tests to understand patterns
- [ ] Tests use `t.Parallel()` matching existing conventions

**Effort**: small (<2h) | **Risk**: low | **Blocks**: P3-007 (template alignment)

---

## Summary

| Priority | Finding | Title | Effort | Blocks |
|----------|---------|-------|--------|--------|
| 🟡 MED | P4-001 | Clean typosquat go.sum entry | small | — |
| 🟡 MED | P4-005 | Make vulncheck blocking | small | — |
| 🟡 MED | P4-002 | Generate actual coverage data | small | P4-006 |
| 🟡 MED | P4-004 | Extract magic number constant | small | — |
| 🟡 MED | P4-003 | Add error context in strategy.go | small | CORR-004 |
| 🟡 MED | P4-006 | Enforce coverage threshold | small | P4-002 |
| 🟡 MED | P4-009 | Standardize error wrapping | small | — |
| 🟡 MED | P4-008 | Fix test registry pollution | small | CORR-002 |
| 🟡 MED | P4-010 | Add _template tests | small | P3-007 |
| 🟡 MED | P4-007 | Replace blanket gosec exclusions | medium | CORR-004 |

**Priority Order**: P4-001 (immediate: supply chain) → P4-005 (immediate: CVE bypass) → P4-002 + P4-006 (coverage pipeline) → P4-004 + P4-003 + P4-009 (code quality batch) → P4-008 (post-CORR-002) → P4-007 + P4-010 (medium effort)

*Generated by Sequoia M2 Reporter — audit-20260521-sequoia-ai | Schema v1.0*
