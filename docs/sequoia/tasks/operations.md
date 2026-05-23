# Operations Tasks — sequoia-ai

**Score**: 0/100 (Critical) | **Findings**: 15 (2 CRITICAL, 4 HIGH, 9 MEDIUM) | **Audit ID**: audit-20260521-sequoia-ai

---

## 🔴 CRITICAL Findings

### P6-001: Remove continue-on-error from release workflow

**Problem**: `.github/workflows/release.yml` and `.github/workflows/ci.yml` use `continue-on-error: true` on the govulncheck step. Build failures, test failures, and CVE discoveries are silently masked. The release proceeds regardless, publishing potentially broken or vulnerable binaries to all distribution channels.

**Root Cause**: CORR-001 (Supply Chain Release Pipeline Weakness)

**Evidence**:
- `ci.yml:30-31`: `continue-on-error: true` on govulncheck
- `release.yml`: no vulncheck step exists at all
- No downstream step checks the vulncheck result status
- Combined with no CODEOWNERS or approval gate, this means no human ever sees the failure

**Fix**: Remove `continue-on-error: true`. Add vulncheck to release.yml before GoReleaser. Verify exit code behavior on CVE discovery. Split CI into phased jobs so failures are visible.

**Acceptance Criteria**:
- [x] `continue-on-error: true` removed from all workflow steps
- [x] Vulncheck added to release.yml (runs before GoReleaser)
- [x] CI phased: lint → test → build → smoke (each gates the next)
- [x] Release blocked if any security check fails
- [x] Documented: vulncheck exit behavior and how to handle CVEs

**Effort**: small (<30m, part of CORR-001) | **Risk**: medium (may surface existing CVEs) | **Blocks**: CORR-001
**Status**: ✅ Completed — `0466002` | **SDD cycle**: fast-forward (score=1) | **Pipeline**: vulncheck → test → goreleaser → smoke

---

### P6-002: Add approval gate to release workflow

**Problem**: Any push to a tag triggers automatic publication via GoReleaser to GitHub Releases, Homebrew, and Scoop. No approval, no review, no manual gate. A compromised contributor account or accidental tag push publishes immediately.

**Root Cause**: CORR-001 (Supply Chain Release Pipeline Weakness)

**Evidence**:
- Release workflow trigger: `on: push: tags: ['v*']`
- No `environment` with required reviewers
- GoReleaser `draft: true` exists but is just a cosmetic flag (draft release can still be accessed)
- Homebrew and Scoop publish immediately via GoReleaser — draft doesn't prevent these

**Fix**: Add GitHub Environment `release` with required reviewers (minimum 1). Configure release workflow to reference this environment. Document release approval process.

**Acceptance Criteria**:
- [x] GitHub Environment `release` created with required reviewers
- [x] Release workflow references the protected environment
- [x] At least 1 reviewer approval required before workflow executes
- [x] Documented: who can approve releases, how to request approval
- [x] Tested: tag push triggers approval request, not immediate execution

**Effort**: small (<30m) | **Risk**: low (no breaking change to existing flow) | **Blocks**: CORR-001
**Status**: ✅ Completed — `c203bbf` | **SDD cycle**: fast-forward (score=0) | **Validated**: v1.0.11 published through approval gate

---

## 🔴 HIGH Findings

### P6-003: Add cross-platform pre-release testing

**Problem**: `.github/workflows/release.yml` runs pre-tests only on `ubuntu-latest`. No macOS or Windows validation before publishing. A platform-specific bug in a release goes undetected until users report it.

**Root Cause**: CORR-001 (Supply Chain Release Pipeline Weakness)

**Evidence**:
- Release pre-test job: `runs-on: ubuntu-latest` only
- CI matrix covers windows-latest, macos-latest, ubuntu-latest — but release doesn't
- GoReleaser produces binaries for all 3 platforms but only 1 is tested before publish
- Windows-specific issues (line endings, path handling) are never validated pre-release

**Fix**: Add macOS and Windows to the release pre-test matrix. Or reuse the CI matrix result as a gate (require CI to pass on all platforms before release can run).

**Acceptance Criteria**:
- [ ] Release pre-tests run on ubuntu-latest, macos-latest, windows-latest
- [ ] All platforms must pass before GoReleaser executes
- [ ] Test job uses matrix strategy matching CI configuration
- [ ] Platform-specific test skips documented (e.g., `-race` not on Windows)

**Effort**: medium (2-4h) | **Risk**: medium (may surface platform-specific issues) | **Blocks**: CORR-001

---

### P6-004: Add workflow_dispatch trigger to release

**Problem**: Release workflow has only `push: tags` trigger. Cannot manually re-run a release, cannot re-validate a past release, and cannot test the release workflow without pushing a tag.

**Root Cause**: CORR-001 (Supply Chain Release Pipeline Weakness)

**Fix**: Add `workflow_dispatch` trigger with optional inputs (version tag, skip-publish flag). Enable manual re-runs for debugging and post-mortem validation.

**Acceptance Criteria**:
- [ ] `workflow_dispatch` trigger added to release.yml
- [ ] Input: `version` (tag to release, e.g., `v1.2.3`)
- [ ] Input: `skip_publish` (boolean, for dry-run testing)
- [ ] Dry-run mode runs all steps except GoReleaser publish
- [ ] Documented: how to use workflow_dispatch for testing

**Effort**: small (<1h) | **Risk**: low | **Blocks**: CORR-001

---

### P6-005: Add CODEOWNERS file for release path

**Problem**: No CODEOWNERS file exists. No required code review on `.github/workflows/release.yml`, `.goreleaser.yaml`, `install.sh`, or `install.ps1` — all of which are release-critical files.

**Root Cause**: CORR-001 (Supply Chain Release Pipeline Weakness)

**Evidence**:
- `CODEOWNERS` file does not exist in repository
- Branch protection rules unknown (but CODEOWNERS is the standard enforcement mechanism)
- Release workflow files, installer scripts, and GoReleaser config all unprotected

**Fix**: Create `CODEOWNERS` file requiring review from maintainers on release-critical paths. Configure branch protection to require CODEOWNERS review.

**Acceptance Criteria**:
- [ ] `CODEOWNERS` file created with maintainer team as owners
- [ ] Patterns covering: `.github/workflows/release.yml`, `.goreleaser.yaml`, `install.sh`, `install.ps1`, `go.mod`, `go.sum`
- [ ] Branch protection requires CODEOWNERS review for these paths
- [ ] Documented: who reviews release changes, what requires review

**Effort**: small (<15m) | **Risk**: low | **Blocks**: CORR-001

---

### P6-006: Add post-deploy smoke test after release

**Problem**: Release workflow ends at GoReleaser execution. No verification that published artifacts are functional. A broken binary could be published and remain undetected until user reports.

**Root Cause**: CORR-001 (Supply Chain Release Pipeline Weakness)

**Evidence**:
- Release workflow ends with GoReleaser — no post-publish verification
- No step downloads the published binary and runs it
- No checksum verification against published `checksums.txt`
- No cosign signature verification on published artifacts

**Fix**: After GoReleaser publishes, add steps to: download binary from GitHub Releases, verify SHA-256 checksum, verify cosign signature, run `./sequoia version` and `./sequoia status`.

**Acceptance Criteria**:
- [ ] Post-publish step downloads binary from GitHub Releases
- [ ] SHA-256 checksum verified against published `checksums.txt`
- [ ] Cosign signature verified on downloaded binary
- [ ] `./sequoia version` exits 0 with correct version
- [ ] `./sequoia status` exits 0 and produces output
- [ ] Test on at least 2 platforms (ubuntu + macos recommended)
- [ ] Release marked as failed if any verification fails

**Effort**: medium (4-8h) | **Risk**: medium | **Blocks**: CORR-001

---

## 🟡 MEDIUM Findings

### P6-007: Implement audit command in GitHub Action

**Problem**: `action.yml:113-127` runs `sequoia status` as a placeholder, emitting `health_score=N/A` and `findings_count=0`. The Action provides no real audit data, defeating its purpose.

**Fix**: Implement `sequoia audit --format=json` in the Go binary. Update action.yml to use it and parse JSON output for real health_score and findings_count.

**Acceptance Criteria**:
- [ ] `sequoia audit` command implemented (or documented as external dependency)
- [ ] action.yml runs the audit command with `--format=json`
- [ ] JSON output parsed to populate `health_score` and `findings_count`
- [ ] Placeholder `v0.1.0` text updated to real version
- [ ] test-action.yml validates real audit outputs

**Effort**: large (16-32h) | **Risk**: high (significant feature) | **Blocks**: none

---

### P6-008: Add retry logic to Windows installer

**Problem**: `scripts/install.ps1:182-190` uses `Invoke-WebRequest` without retry. The Unix installer (`install.sh:275`) uses `curl --retry 3 --retry-delay 2`. Windows users with unstable connections get silent failures.

**Root Cause**: CORR-001 (Supply Chain Release Pipeline Weakness)

**Fix**: Add retry loop (3 attempts, 2s delay, exponential backoff) for `Invoke-WebRequest`. Retry both binary download and checksum download. Log retry attempts to stderr.

**Acceptance Criteria**:
- [ ] Retry loop: 3 attempts, 2s initial delay, exponential backoff
- [ ] Both binary and checksum downloads retried
- [ ] Retry attempts logged to stderr
- [ ] Matching `scripts_test.go:185-188` updated (currently marks this as known gap)

**Effort**: small (1-2h) | **Risk**: low | **Blocks**: CORR-001

---

### P6-009: Use separate tokens for Homebrew and Scoop

**Problem**: `.goreleaser.yaml:133-165` uses `HOMEBREW_TAP_TOKEN` for both Homebrew and Scoop repositories. A single compromised token gives access to both distribution channels. Violates principle of least privilege.

**Fix**: Create separate GitHub token with repo-scoped access to `scoop-sequoia`. Add `SCOOP_TAP_TOKEN` secret. Update `.goreleaser.yaml` scoops section.

**Acceptance Criteria**:
- [x] Separate GitHub token created with scope limited to `scoop-sequoia` repo
- [x] `SCOOP_TAP_TOKEN` secret added to repository
- [x] `.goreleaser.yaml` scoops section uses `{{ .Env.SCOOP_TAP_TOKEN }}`
- [x] Homebrew section continues using `HOMEBREW_TAP_TOKEN` (already scoped correctly)
- [x] Release workflow verified: both repos receive updates

**Effort**: small (<30m) | **Risk**: low | **Blocks**: none

---

### P6-010: Add structured logging with verbosity levels

**Problem**: All CLI output uses bare `fmt.Print`/`fmt.Printf`. No `--verbose`, `--debug`, or `--quiet` flags. Error messages, progress output, and debug info are all mixed at the same level.

**Fix**: Add `--verbose`/`-v`, `--debug`, and `--quiet`/`-q` flags. Use Go's `log/slog` (standard library since Go 1.21) for structured logging. CLI mode uses text handler; add `--log-json` for machine-parseable output. TUI mode uses file-backed logging (Bubbletea handles rendering).

**Acceptance Criteria**:
- [ ] `--verbose`/`-v` flag increases output detail
- [ ] `--debug` flag enables debug-level output (file paths, timings, internal state)
- [ ] `--quiet`/`-q` flag suppresses non-error output
- [ ] `log/slog` used as structured logging backend
- [ ] `--log-json` flag for machine-parseable JSON output
- [ ] TUI mode logs to file, not stdout

**Effort**: medium (4-8h) | **Risk**: medium | **Blocks**: none

---

### P6-011: Enable race detector on Windows CI

**Problem**: CI matrix includes `-race` on ubuntu and macos but not on Windows. Go race detector is supported on Windows/amd64. Missing coverage for Windows-specific race conditions (file locking, path handling).

**Fix**: Enable `-race` flag on Windows CI test job. Verify no false positives (Windows file system behavior can differ).

**Acceptance Criteria**:
- [ ] Windows CI test command includes `-race` flag
- [ ] No new race condition failures on Windows
- [ ] If false positives emerge, document and skip specific tests with `-race` annotations
- [ ] CI matrix: all 3 platforms run with race detector

**Effort**: small (<30m) | **Risk**: low | **Blocks**: none

---

### P6-012: Verify cosign version in install scripts

**Problem**: `install.sh` and `install.ps1` check that `cosign` exists but don't verify the version. An outdated cosign with known vulnerabilities could be used to verify signatures, giving false confidence.

**Fix**: Add minimum version check after cosign detection. Parse `cosign version` output. Warn or fail if version is below minimum.

**Acceptance Criteria**:
- [ ] `cosign version` parsed to extract version number
- [ ] Minimum version check (e.g., >= 2.0.0) after detection
- [ ] Warning emitted if version is below minimum (don't block — user may have good reason)
- [ ] Both install.sh and install.ps1 updated

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

### P6-013: Sync CHANGELOG.md with release notes

**Problem**: CHANGELOG.md is manually maintained but GoReleaser auto-generates release notes from commit messages. The two can diverge, causing confusion about what changed in each release.

**Fix**: Add CI step to validate CHANGELOG.md against GoReleaser-generated notes for the latest release. Or configure GoReleaser to use CHANGELOG.md as the source of release notes.

**Acceptance Criteria**:
- [ ] CI step validates CHANGELOG.md entry matches release tag
- [ ] GoReleaser configured to extract release notes from CHANGELOG.md for current version
- [ ] Documented: CHANGELOG.md is the source of truth for release notes
- [ ] Release workflow fails if CHANGELOG doesn't have entry for the tag

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

### P6-014: Validate GoReleaser config in CI

**Problem**: `.goreleaser.yaml` is never validated in CI. Configuration errors are only discovered during an actual release, causing failed releases and manual intervention.

**Fix**: Add CI step that runs `goreleaser release --snapshot --skip-publish --clean` to validate the config. This catches config errors before a release is attempted.

**Acceptance Criteria**:
- [ ] CI step: `goreleaser release --snapshot --skip-publish --clean`
- [ ] Step runs on PRs that modify `.goreleaser.yaml`
- [ ] Fails CI if GoReleaser config is invalid
- [ ] Snapshot artifacts discarded after validation

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

### P6-015: Add observability for CLI operations

**Problem**: No metrics, tracing, or structured logging for CLI operations. Performance regressions, error patterns, and usage patterns are invisible. Hard to debug user-reported issues without reproduction steps.

**Fix**: Not urgent for a CLI tool at this maturity. Consider `log/slog` (P6-010) as the foundation. If the CLI grows to support server mode or long-running operations, add OpenTelemetry tracing.

**Acceptance Criteria**:
- [ ] Deferred — not implementing now
- [ ] P6-010 (structured logging) provides the observability foundation
- [ ] Re-evaluate if CLI grows to support server/daemon mode
- [ ] Users can already use `--debug` + log output for debugging

**Effort**: deferred | **Risk**: low | **Blocks**: none

---

## Summary

| Priority | Finding | Title | Effort | Blocks |
|----------|---------|-------|--------|--------|
| 🔴 CRITICAL | P6-001 | ✅ Remove continue-on-error | small | CORR-001 |
| 🔴 CRITICAL | P6-002 | ✅ Add approval gate | small | CORR-001 |
| 🔴 HIGH | P6-003 | Cross-platform pre-tests | medium | CORR-001 |
| 🔴 HIGH | P6-004 | workflow_dispatch trigger | small | CORR-001 |
| 🔴 HIGH | P6-005 | CODEOWNERS file | small | CORR-001 |
| 🔴 HIGH | P6-006 | Post-deploy smoke test | medium | CORR-001 |
| 🟡 MED | P6-008 | Retry in Windows installer | small | CORR-001 |
| 🟡 MED | P6-009 | Separate Scoop token | small | — |
| 🟡 MED | P6-005 | Coverage threshold | small | — |
| 🟡 MED | P6-011 | Enable -race on Windows | small | — |
| 🟡 MED | P6-012 | Verify cosign version | small | — |
| 🟡 MED | P6-013 | Sync CHANGELOG | small | — |
| 🟡 MED | P6-014 | Validate GoReleaser config | small | — |
| 🟡 MED | P6-010 | Structured logging | medium | — |
| 🟡 MED | P6-007 | Implement audit command | large | — |
| 🟡 MED | P6-015 | Observability | deferred | — |

**Priority Order**: ~~P6-001 + P6-002~~ ✅ → P6-005 (quick gate) → P6-003 + P6-004 + P6-006 (CORR-001 batch) → P6-008 + P6-009 + P6-011 → P6-013 + P6-014 → P6-010 → P6-007 (backlog)

*Generated by Sequoia M2 Reporter — audit-20260521-sequoia-ai | Schema v1.0*
