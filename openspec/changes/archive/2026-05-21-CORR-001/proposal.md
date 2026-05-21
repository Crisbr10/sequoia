# Proposal: Release Pipeline Security Hardening (CORR-001)

## Intent

Close 9 supply-chain weaknesses from a release workflow designed for convenience over security: unverified binaries, floating tags, no environment protection, auto-publish, single-platform tests, no manual dispatch, missing CODEOWNERS, no post-deploy smoke, no installer retries.

## Scope

### In Scope
- SHA-256 verification before cosign (P1-001)
- Pin actions to commit hashes in ci.yml and release.yml (P1-007)
- `environment: release` with required reviewers (P1-008)
- Manual approval before GoReleaser (P6-002)
- Cross-platform pre-release tests (ubuntu, macos, windows) (P6-003)
- `workflow_dispatch` trigger (P6-004)
- `.github/CODEOWNERS` on release-critical paths (P6-005)
- Post-deploy smoke test (P6-006)
- Retry with backoff in install.ps1 (P6-008)

### Out of Scope
- P6-001, P6-007 — completed in CORR-002. No Go code changes. Homebrew tap.

## Capabilities

### New Capabilities
- **release-pipeline-security**: Environment protection, approval gates, binary verification, cross-platform tests, workflow_dispatch, post-deploy smoke, CODEOWNERS (P1-001, P1-008, P6-002–P6-006)
- **action-pinning**: All actions pinned to commit hashes (P1-007)
- **installer-resilience**: Retry with exponential backoff in install.ps1 (P6-008)

### Modified Capabilities
None — no CI/CD specs in `openspec/specs/`.

## Approach

1. **release.yml**: Add `workflow_dispatch`, `environment: release` with reviewers, OS matrix tests, SHA-256 verification before cosign, post-deploy smoke.
2. **ci.yml**: Pin all `@vX` tags to commit SHAs.
3. **CODEOWNERS**: New file assigning release workflow and installer to maintainers.
4. **install.ps1**: Wrap `Invoke-WebRequest` (lines 183, 201) in retry loop (3 attempts, 2s/4s/8s). Document pinning policy in CONTRIBUTING.md.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `.github/workflows/release.yml` | Modified | Environment, dispatch, matrix, verification, smoke |
| `.github/workflows/ci.yml` | Modified | Pin actions to commit hashes |
| `.github/CODEOWNERS` | New | Required review on release paths |
| `scripts/install.ps1` | Modified | Retry on Invoke-WebRequest |
| `docs/CONTRIBUTING.md` | Modified | Release process and pinning policy |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Commit-hash pinning causes dependabot churn | Med | Document pin+update policy |
| Environment blocks emergency hotfix | Low | workflow_dispatch + bypass docs |
| Matrix tests increase release duration | Low | ubuntu-only GoReleaser; others test-only |

## Rollback Plan

Revert to previous state via single commit. Remove CODEOWNERS if it blocks velocity.

## Dependencies

- CORR-002 completed (P6-001, P6-007 preconditions)
- GitHub environment `release` with reviewer protection must exist

## Success Criteria

- [ ] Manual approval required before GoReleaser
- [ ] All actions pinned to commit hashes (`grep '@v' .github/` empty)
- [ ] SHA-256 verified before cosign
- [ ] Tests pass on ubuntu, macos, windows pre-release
- [ ] `workflow_dispatch` functional
- [ ] CODEOWNERS covers release-critical paths
- [ ] Post-deploy smoke verifies published artifacts
- [ ] install.ps1 retries on transient failures
