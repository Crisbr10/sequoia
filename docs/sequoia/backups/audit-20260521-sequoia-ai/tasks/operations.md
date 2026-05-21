# Operations Tasks — sequoia-ai

**Score**: 0/100 (Critical) | **Findings**: 15

---

## 🔴 P6-001 (high): Make vulncheck blocking + add to release

**Problem**: CI vulncheck uses `continue-on-error: true`. Release pipeline has no vulncheck at all.

**Acceptance Criteria**:
- [ ] Remove `continue-on-error: true` from vulncheck in `.github/workflows/ci.yml:30-31`
- [ ] Add vulncheck step to `.github/workflows/release.yml` (before GoReleaser)
- [ ] Verify vulncheck exits non-zero on CVE discovery
- [ ] Same as P4-003 — coordinate with Quality tasks

**Effort**: small (<30m) | **Risk**: low | **Blocks**: none (coordinates with P4-003)

---

## 🔴 P6-002 (high): Implement audit command in GitHub Action

**Problem**: `action.yml:113-127` runs `sequoia status` as a placeholder, emits `health_score=N/A`, `findings_count=0`. Action provides no real audit data.

**Acceptance Criteria**:
- [ ] Implement `sequoia audit` command in the Go binary (or document as external dependency)
- [ ] Update action.yml to run `sequoia audit --format=json` instead of `sequoia status`
- [ ] Parse JSON output to populate `health_score` and `findings_count` outputs
- [ ] Replace v0.1.0 placeholder text with real version
- [ ] Update test-action.yml to validate real audit outputs (not just non-empty)

**Effort**: large (16-32h) | **Risk**: high | **Blocks**: CORR-005

---

## 🔴 P6-003 (high): Add post-release smoke test to release workflow

**Problem**: `.github/workflows/release.yml` runs GoReleaser but never verifies published artifacts.

**Acceptance Criteria**:
- [ ] After GoReleaser, add step to download a published binary from GitHub Releases
- [ ] Verify SHA-256 checksum against `checksums.txt`
- [ ] Verify cosign signature on the downloaded binary
- [ ] Run `./sequoia version` and `./sequoia status` on downloaded binary
- [ ] Test on at least 2 platforms (ubuntu + macos or ubuntu + windows)
- [ ] Fail release if any verification step fails

**Effort**: medium (4-8h) | **Risk**: medium | **Blocks**: none

---

## 🔴 P6-004 (high): Document release rollback procedure

**Problem**: No documented recovery procedure if a bad release is published.

**Acceptance Criteria**:
- [ ] Create `docs/RELEASE-ROLLBACK.md` with step-by-step recovery:
  - How to delete a GitHub Release
  - How to revert Homebrew formula to previous version
  - How to revert Scoop manifest to previous version
  - How to communicate to users (GitHub Discussion, CHANGELOG amendment)
- [ ] Link from CONTRIBUTING.md and release checklist
- [ ] Document GoReleaser `draft: true` behavior and manual publish gate

**Effort**: small (1-2h) | **Risk**: low | **Blocks**: none

---

## P6-005 (medium): Enforce coverage threshold in CI

**Problem**: No coverage threshold. CHANGELOG claims 90%+ but it's not enforced.

**Acceptance Criteria**:
- [ ] Add `-coverprofile=coverage.out` to CI test step
- [ ] Add coverage threshold check: `go tool cover -func=coverage.out | grep total | awk '{print $3}'` must be ≥70%
- [ ] Fail CI if coverage drops below threshold
- [ ] Set threshold conservatively (70%) given current unknown state; raise later

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

## P6-006 (medium): Add retry logic to Windows installer

**Problem**: `scripts/install.ps1:182-190` has no retry on download failure. Unix installer (`install.sh:275`) uses `curl --retry 3 --retry-delay 2`.

**Acceptance Criteria**:
- [ ] Add retry loop (3 attempts, 2s delay, exponential backoff) for `Invoke-WebRequest` in install.ps1
- [ ] Retry both binary download AND checksum download
- [ ] Log retry attempts to stderr
- [ ] Update `scripts_test.go:185-188` (currently marks this as known gap)

**Effort**: small (1-2h) | **Risk**: low | **Blocks**: none

---

## P6-007 (medium): Split monolithic CI job into phases

**Problem**: `.github/workflows/ci.yml:42-63` chains test, build, and smoke in a single job. Build runs even if tests fail.

**Acceptance Criteria**:
- [ ] Split into separate jobs: `lint`, `test` (matrix), `build`, `smoke`
- [ ] `build` needs `test`; `smoke` needs `build`
- [ ] Add `if: success()` guard to build step (or use `needs`)
- [ ] Smoke steps removed from test job (only run in smoke job)
- [ ] Share build artifacts between jobs using `actions/upload-artifact`

**Effort**: medium (2-4h) | **Risk**: medium | **Blocks**: CORR-002

---

## P6-008 (medium): Use separate tokens for Homebrew and Scoop

**Problem**: `.goreleaser.yaml:133-165` uses `HOMEBREW_TAP_TOKEN` for both Homebrew and Scoop repos.

**Acceptance Criteria**:
- [ ] Create separate GitHub token with repo-scoped access to `scoop-sequoia`
- [ ] Add `SCOOP_TAP_TOKEN` secret to repository
- [ ] Update `.goreleaser.yaml` scoops section to use `{{ .Env.SCOOP_TAP_TOKEN }}`
- [ ] Verify release workflow still pushes to both repos

**Effort**: small (<30m) | **Risk**: low | **Blocks**: none

---

## P6-009 (medium): Add structured logging with verbosity levels

**Problem**: All output uses bare `fmt.Print`. No `--verbose`, `--debug`, or `--quiet` flags.

**Acceptance Criteria**:
- [ ] Add `--verbose`/`-v` flag (increases output detail)
- [ ] Add `--debug` flag (debug-level output including file paths, timings)
- [ ] Add `--quiet`/`-q` flag (errors only, no progress output)
- [ ] Use Go's `log/slog` (standard library, Go 1.21+) for structured logging
- [ ] CLI mode uses text handler; add `--log-json` for machine-parseable output
- [ ] TUI mode continues using Bubbletea rendering (logging is file-backed in TUI)

**Effort**: medium (4-8h) | **Risk**: medium | **Blocks**: none

---

## P6-010 (medium): Verify release artifacts in release workflow

**Problem**: Release workflow ends at GoReleaser execution — no smoke test of built artifacts.

**Acceptance Criteria**:
- [ ] Add step to download published binary from draft release
- [ ] Verify checksum matches published checksums.txt
- [ ] Run `./sequoia version` on downloaded binary
- [ ] Run `./sequoia status` on downloaded binary
- [ ] Verify the binary exits with correct exit code, not segfault

**Effort**: medium (2-4h) | **Risk**: low | **Blocks**: none

---

## P6-011~P6-015 (low/info): Improvements

- **P6-011**: Enable `-race` on Windows CI (Go race detector works on Windows/amd64)
- **P6-012**: Add minimum cosign version check in `install.sh` (`cosign version` parse)
- **P6-013**: Add CI step to validate CHANGELOG.md vs auto-generated release notes consistency
- **P6-014**: Add integration test that runs `goreleaser release --snapshot --skip-publish` to validate config
- **P6-015**: Defer — OpenTelemetry tracing not needed for CLI tool today

**Effort**: small-medium (2-6h) | **Risk**: low | **Blocks**: none

---

## Priority Order

1. **P6-001** (vulncheck blocking) — immediate, security
2. **P6-005** (coverage threshold) — immediate, quality gate
3. **P6-006** (Windows retry) — quick fix, user-facing reliability
4. **P6-008** (separate tokens) — quick fix, principle of least privilege
5. **P6-007** (split CI jobs) — medium effort, structural improvement
6. **P6-004** (rollback docs) — quick doc, operational readiness
7. **P6-003 + P6-010** (release smoke tests) — medium effort, release safety
8. **P6-009** (structured logging) — medium effort, debugging UX
9. **P6-002** (audit command) — large effort, backlog
