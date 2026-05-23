# Security Tasks — sequoia-ai

**Score**: 35/100 (Critical) | **Findings**: 10 (1 CRITICAL, 0 HIGH, 9 MEDIUM) | **Audit ID**: audit-20260521-sequoia-ai

---

## 🔴 CRITICAL — P1-001: Verify binary integrity in release pipeline

**Impact**: Signed artifacts are trusted without verifying they match the built binaries. A compromised release workflow could sign attacker-controlled binaries, distributing malware through all channels (GoReleaser, Homebrew, Scoop).

**Evidence**:
- `release.yml` pushes to GoReleaser which signs with cosign, but no step verifies the binary before signing
- No SHA-256 comparison between build output and the artifact being signed
- Supply chain trust is transitive: if build step is compromised, the signed output is also compromised

**Root Cause**: CORR-001 (Supply Chain Release Pipeline Weakness)

**Acceptance Criteria**:
- [ ] Add verification step after build: compute SHA-256 of built binary and compare against expected value
- [ ] Run cosign sign only after checksum verification passes
- [ ] Add CI step that rebuilds and verifies checksum matches published one (reproducible build check)
- [ ] Document the binary verification chain in CONTRIBUTING.md

**Effort**: medium (4-8h) | **Risk**: high | **Blocks**: CORR-001

---

## 🟡 MEDIUM Findings

### P1-002: Explicit file permissions on installed files

**Problem**: `copyFile` at `adapters/common/installer.go:200` uses `os.Create(dst)` which creates files with umask-dependent permissions. Installed skill/command files could be world-readable on misconfigured systems.

**Fix**: Use `os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0o644)` instead of `os.Create`. Add `os.Chmod` fallback after `os.Create` if using that path. Verify `StageFile` at `installer.go:109` already uses `os.WriteFile` with `0o644`.

**Acceptance Criteria**:
- [ ] `copyFile` creates destination file with explicit `0o644` permissions
- [ ] Test verifies installed files are not world-writable
- [ ] Audit `StageFile` and `AtomicWriteFile` for consistent permission handling

**Effort**: small (<2h) | **Risk**: low | **Blocks**: none

---

### P1-003: Explicit permissions on probe file

**Problem**: `Prepare()` creates `.sequoia-probe` with default umask permissions at `adapters/common/installer.go:64`. Race window exists where an empty file sits on disk with uncontrolled permissions.

**Fix**: Replace `os.Create(probe)` with `os.OpenFile(probe, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)`. Clean up probe file in defer immediately after creation check passes.

**Acceptance Criteria**:
- [ ] Probe file created with `0o600` (owner read/write only)
- [ ] Probe file cleaned up in defer immediately after successful write check
- [ ] Test verifies probe file has restricted permissions

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

### P1-004: Remove init() auto-registration side effects

**Problem**: All 5 adapters call `RegisterIn(adapters.DefaultRegistry)` in `init()`, creating implicit side effects at import time. Tests depend on this global pollution. `cmd/sequoia/main.go:43` creates its own registry, making the init() registration dead code.

**Root Cause**: CORR-002 (Global Mutable State via init() Functions)

**Fix**: Remove all `init()` functions from adapter packages. Make `NewRegistry()` + `RegisterIn()` the only registration path. Update tests that depend on pre-populated `DefaultRegistry`.

**Acceptance Criteria**:
- [ ] All `init()` functions removed from 5 adapters + `_template`
- [ ] `DefaultRegistry` removed from `adapters/registry.go`
- [ ] All tests pass without relying on global state
- [ ] `_template` updated to demonstrate `RegisterFactory()` pattern

**Effort**: medium (4-8h) | **Risk**: medium (tests may need updating) | **Blocks**: CORR-002

---

### P1-005: Sanitize home directory path in status output

**Problem**: `runStatus` at `cmd/sequoia/main.go:321` prints the user's full home directory path (including resolved symlink target) to stdout, which may leak OS usernames in CI logs or shared terminals.

**Fix**: Display `~` shorthand or `filepath.Base(home)` instead of full absolute path. Show symlink resolution relative to home directory.

**Acceptance Criteria**:
- [ ] Home directory path displayed as `~` or basename, not full path
- [ ] Resolved symlink target shown relative to home
- [ ] Output still clearly indicates symlink presence

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

### P1-006: Validate working-directory input in GitHub Action

**Problem**: `action.yml` uses `${{ inputs.working-directory }}` unsanitized in shell expressions and JS interpolation. Path traversal possible if workflow author passes `..` segments.

**Fix**: Add validation step that checks `inputs.working-directory` doesn't contain `..` segments and resolves within `${{ github.workspace }}`. Fail action early with clear error message.

**Acceptance Criteria**:
- [ ] Validation step rejects paths containing `..` segments
- [ ] Validation verifies path resolves within `${{ github.workspace }}`
- [ ] Action fails early with descriptive error on invalid input
- [ ] Input description in `action.yml` documents path constraints

**Effort**: small (<2h) | **Risk**: low | **Blocks**: none

---

### P1-007: Pin GitHub Actions to commit hash

**Problem**: Release workflow uses floating version tags (`@v4`, `@v6`) for GitHub Actions. A compromised action repository could inject malicious code into the release pipeline.

**Root Cause**: CORR-001 (Supply Chain Release Pipeline Weakness)

**Fix**: Replace all `uses: action@v4` with `uses: action@<full-commit-hash>`. Document the pinning policy in CONTRIBUTING.md.

**Acceptance Criteria**:
- [ ] All GitHub Actions in `ci.yml` and `release.yml` pinned to commit hashes
- [ ] Dependabot configured to update pinned hashes (already configured for actions)
- [ ] CONTRIBUTING.md documents the pinning requirement

**Effort**: small (<1h) | **Risk**: low | **Blocks**: CORR-001

---

### P1-008: Add environment protection to release workflow

**Problem**: Release workflow has no required reviewers or environment protection rules. Any contributor with push access can publish a release.

**Root Cause**: CORR-001 (Supply Chain Release Pipeline Weakness)

**Fix**: Configure GitHub Environment with required reviewers for the release workflow. Require at least 1 reviewer approval before release runs.

**Acceptance Criteria**:
- [ ] GitHub Environment `release` created with required reviewers
- [ ] Release workflow references the protected environment
- [ ] At least 1 reviewer required before release workflow executes
- [ ] Document release approval process in CONTRIBUTING.md

**Effort**: small (<30m) | **Risk**: low | **Blocks**: CORR-001

---

### ✅ P1-009: Remove typosquat package from go.sum

**Problem**: `go.sum` line 52 contains `go.yaml.in/yaml/v3` — NOT the legitimate `gopkg.in/yaml.v3`. This appears to be a domain squatting on `yaml.in`. Cobra v1.10.2 tiene dependencia activa del typosquat, por lo que `go mod tidy` solo no bastaba.

**Fix**: Se aplicó `replace go.yaml.in/yaml/v3 v3.0.4 => gopkg.in/yaml.v3 v3.0.0` en `go.mod`. Se agregó `go mod verify` al job de lint en CI.

**Acceptance Criteria**:
- [x] `go.yaml.in/yaml/v3` removed from `go.sum`
- [x] `grep` confirms zero source imports of the suspicious package
- [x] `go mod verify` added to CI lint job
- [x] No `go mod tidy` diff after cleanup

**Effort**: small (<5m) | **Risk**: medium | **Blocks**: none | **Resuelto**: 2026-05-22 (SDD fast-forward, verify PASS 4/4)

---

### P1-010: Verify signatures on downloaded adapter tools

**Problem**: Adapters download and install external tools (Claude Code, Codex CLI, Gemma CLI) without verifying cryptographic signatures. A compromised download server could distribute backdoored binaries.

**Fix**: For each adapter that downloads external binaries, verify GPG/signify signatures or checksum against a known-good manifest. Document which tools support signature verification.

**Acceptance Criteria**:
- [ ] Each adapter documents whether the external tool supports signature verification
- [ ] For tools that support it, implement signature verification before installation
- [ ] For tools that don't, document the risk and recommend manual verification
- [ ] Install failure if signature verification fails (with clear error message)

**Effort**: medium (4-8h) | **Risk**: medium | **Blocks**: none

---

## Summary

| Priority | Finding | Title | Effort | Blocks |
|----------|---------|-------|--------|--------|
| 🔴 CRITICAL | P1-001 | Binary verification in release | medium | CORR-001 |
| ✅ | P1-009 | Clean typosquat go.sum entry | small | — |
| 🟡 | P1-002 | File permissions on install | small | — |
| 🟡 | P1-003 | Probe file permissions | small | — |
| 🟡 | P1-004 | Remove init() auto-registration | medium | CORR-002 |
| 🟡 | P1-005 | Sanitize home path output | small | — |
| 🟡 | P1-006 | Validate GH Action input | small | — |
| 🟡 | P1-007 | Pin Actions to commit hash | small | CORR-001 |
| 🟡 | P1-008 | Environment protection | small | CORR-001 |
| 🟡 | P1-010 | Adapter tool signature verification | medium | — |

*Generated by Sequoia M2 Reporter — audit-20260521-sequoia-ai | Schema v1.0*
