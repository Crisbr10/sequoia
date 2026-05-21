# Proposal: CI Gates Hardening (CORR-002)

## Intent

Consolidate 6 CI findings into a single CI gates improvement: govulncheck is non-blocking (`continue-on-error: true`), coverage is untracked despite CHANGELOG claims, coverage artifacts are committed, and the monolithic CI job masks failures by running build/smoke after test failures.

## Scope

### In Scope
- Remove `continue-on-error: true` from vulncheck in ci.yml; add vulncheck to release.yml
- Add `-coverprofile=coverage.out -covermode=atomic` to CI `go test`
- Enforce ≥70% coverage threshold, failing CI on drop
- Split monolithic `test` job into `lint`, `test` (matrix), `build`, `smoke` with `needs`
- `git rm --cached` stale coverage files; verify .gitignore blocks re-commit

### Out of Scope
- Raising coverage percentage (baseline unknown; 70% is conservative)
- Codecov/coverage service integration
- Race detector on Windows (P6-011, deferred)

## Capabilities

### New Capabilities
- **ci-gates**: CI pipeline with blocking vulncheck, coverage enforcement, and phased job structure

### Modified Capabilities
None — CI is not in `openspec/specs/`.

## Approach

1. **Vulncheck blocking (P6-001/P4-003)**: Change `continue-on-error: true` → `false` (or remove line). Add to release.yml before GoReleaser.
2. **Coverage (P4-005/P6-005)**: Add `-coverprofile=coverage.out -covermode=atomic` to test step. Post-test: parse total % via `go tool cover -func`, fail if <70%. Skip `-covermode=atomic` on Windows.
3. **Cleanup (P4-002)**: `git rm --cached coverage coverage_rc002`. Verify `.gitignore` entries (lines 8-14 already exist: `coverage`, `coverage_*`, `coverage*.out`).
4. **Job split (P6-007)**: Create `lint`, `test` (matrix), `build`, `smoke` jobs. `build` needs `test`; `smoke` needs `build`. Use `actions/upload-artifact`/`download-artifact` for binary.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `.github/workflows/ci.yml` | Modified | Restructured from 1 job → 4 jobs with gates |
| `.github/workflows/release.yml` | Modified | Add vulncheck step before GoReleaser |
| `coverage`, `coverage_rc002` | Removed | `git rm --cached`; covered by existing .gitignore |
| `.gitignore` | Verified | Patterns already present (lines 8-14); no change needed |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| govulncheck surfaces existing CVEs, blocking PRs | Med | Investigate and fix CVEs first; document remediation |
| Coverage threshold fails on first run | High | Set at 70% (conservative); adjust if needed |
| Job split breaks artifact passing | Low | Test with `upload-artifact` v4 + `download-artifact` v4 |
| Race condition: coverage_rc002 not properly ignored | Low | Verify with `git ls-files` after `rm --cached` |

## Rollback Plan

- Revert `ci.yml` and `release.yml` to original state via single commit
- Restore `continue-on-error: true` temporarily if vulncheck CVE blocks critical fix
- No database/migration concerns — pure CI config

## Dependencies

- None. CORR-002 is a root node. Unblocks P6-005, P4-005, and P6-007.

## Success Criteria

- [ ] `govulncheck` exits non-zero and fails CI on CVE discovery
- [ ] Release pipeline runs vulncheck before GoReleaser publishes
- [ ] CI fails when total coverage <70%
- [ ] CI runs lint, test, build, and smoke as independent phases
- [ ] `coverage` and `coverage_rc002` removed from git tracking
- [ ] `go test ./...` passes locally (unchanged)
