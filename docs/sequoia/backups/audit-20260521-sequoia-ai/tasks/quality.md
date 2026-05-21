# Quality Tasks — sequoia-ai

**Score**: 20/100 (Critical) | **Findings**: 12

---

## 🔴 P4-001 (high): Remove typosquat package from go.sum

**Problem**: `go.sum` line 52 contains `go.yaml.in/yaml/v3` — NOT the legitimate `gopkg.in/yaml.v3`. This appears to be a typosquatting domain.

**Acceptance Criteria**:
- [ ] Run `go mod tidy` to clean stale entries from go.sum
- [ ] Verify `go.yaml.in/yaml/v3` is removed from go.sum
- [ ] Verify no source code imports `go.yaml.in/yaml/v3` (confirmed: grep finds zero imports)
- [ ] If entry persists after tidy, manually remove from go.sum and run `go mod verify`
- [ ] Add CI step: `go mod verify` to detect tampered go.sum entries

**Effort**: small (<30m) | **Risk**: low | **Blocks**: none

---

## P4-002 (medium): Generate actual coverage data

**Problem**: `coverage` and `coverage_rc002` files contain only `mode:` headers — zero coverage data. CI never generates coverage.

**Acceptance Criteria**:
- [ ] Delete empty `coverage` and `coverage_rc002` files (they're misleading artifacts)
- [ ] Add `.gitignore` entry for `coverage.out`, `coverage.html`
- [ ] Add `make coverage` or documented command: `go test -coverprofile=coverage.out -covermode=atomic ./...`
- [ ] Generate initial coverage report and document baseline in CHANGELOG or docs

**Effort**: small (<1h) | **Risk**: low | **Blocks**: P4-005, P6-005

---

## P4-003 (medium): Make vulncheck blocking in CI

**Problem**: `.github/workflows/ci.yml:30-31` sets `continue-on-error: true` on govulncheck, so CVEs never block PRs.

**Acceptance Criteria**:
- [ ] Remove `continue-on-error: true` from vulncheck step in ci.yml
- [ ] Add vulncheck step to release.yml (currently absent from release pipeline)
- [ ] Verify that `govulncheck@v1.1.4` returns non-zero exit code on found vulnerabilities
- [ ] Document vulncheck in CI troubleshooting guide if failures occur

**Effort**: small (<30m) | **Risk**: low (may surface existing CVEs) | **Blocks**: none

---

## P4-004 (medium): Extract magic number 64 as named constant

**Problem**: Channel buffer capacity `64` is hardcoded in `internal/app/model.go:48,128` and repeated across 22+ files.

**Acceptance Criteria**:
- [ ] Define `const ProgressChannelBufferSize = 64` in `internal/model/types.go`
- [ ] Replace all `make(chan model.ProgressMsg, 64)` with the constant
- [ ] Replace all test occurrences with the constant
- [ ] Document the constant's purpose (comment already exists at model.go:48 — keep it)

**Effort**: small (<1h) | **Risk**: low | **Blocks**: CORR-004

---

## P4-005 (low): Add -coverprofile to CI test command

**Problem**: `.github/workflows/ci.yml:43-47` runs `go test` without `-coverprofile`, so coverage is never measured.

**Acceptance Criteria**:
- [ ] Add `-coverprofile=coverage.out -covermode=atomic` to `go test` command in ci.yml
- [ ] Add step to check coverage against minimum threshold (e.g., 70%)
- [ ] Upload coverage artifact for review (or integrate with coverage service)
- [ ] Windows CI: use `-coverprofile` without `-covermode=atomic` (Windows doesn't support atomic mode)

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

## P4-006~P4-012 (low/info): Code quality improvements

- **P4-006**: Replace `map[string]interface{}` in Codex TOML merge with typed struct using generics or explicit TOML types
- **P4-007**: Add paths_test.go and strategy_test.go to OpenCode adapter (matching peer adapters)
- **P4-008**: Extract ScreenError handler from 197-line `updateScreenKey` into dedicated method
- **P4-009**: Add golden test for ToolSelection screen (only TUI screen without golden coverage)
- **P4-010**: Add basic tests to `_template` adapter (at minimum verify it compiles and installs correctly)
- **P4-011**: Replace file-level `errcheck` exclusion with line-level `//nolint:errcheck` in `pipeline/runner.go`
- **P4-012**: Run `go mod tidy` to remove stale `golang.org/x/exp` entry from go.sum

**Effort**: medium (8-16h total) | **Risk**: low | **Blocks**: none

---

## Priority Order

1. **P4-001** (typosquat) — immediate, supply chain risk
2. **P4-003** (vulncheck) — immediate, security gate bypass
3. **P4-002 + P4-005** (coverage) — enables quality tracking
4. **P4-004** (magic number) — quick refactor
5. **P4-006~P4-012** (code quality) — incremental improvement
