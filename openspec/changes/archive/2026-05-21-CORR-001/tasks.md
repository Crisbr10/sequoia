# Tasks: Release Pipeline Security Hardening

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 150–220 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Low

## Phase 1: Foundation — Pinning + CODEOWNERS

- [x] 1.1 Pin 2 `uses:` to commit SHAs in `.github/workflows/test-action.yml`
- [x] 1.2 Pin all 5 `uses:` to commit SHAs in `.github/workflows/ci.yml`
- [x] 1.3 Add `action-pinning` job to `.github/workflows/ci.yml`: grep scan for `@vX`/`@main`, fail on match
- [x] 1.4 Create `.github/CODEOWNERS` with `* @Crisbr10` + explicit entries for `release.yml`, `install.ps1`, `CODEOWNERS`

## Phase 2: Release Workflow Restructure

- [x] 2.1 Add `workflow_dispatch` trigger to `.github/workflows/release.yml`
- [x] 2.2 Add OS matrix test job (ubuntu, macos, windows): `go test -race -count=1 ./...`
- [x] 2.3 Add SHA-256 verification step in goreleaser job: `sha256sum` binary vs `checksums.txt` before cosign
- [x] 2.4 Add `environment: release` to goreleaser job for manual approval gate
- [x] 2.5 Add post-deploy smoke job: curl published binary, verify checksum against `checksums.txt`
- [x] 2.6 Pin all `uses:` in `.github/workflows/release.yml` to commit SHAs (at least 3 directives)

## Phase 3: Installer Resilience

- [x] 3.1 Add `Invoke-WebRequestWithRetry` function in `scripts/install.ps1`: 3 attempts, delays 2s/4s/8s, exit `$EXIT_NETWORK` on exhaustion
- [x] 3.2 Replace bare `Invoke-WebRequest` at lines 183, 201 with `Invoke-WebRequestWithRetry`

## Phase 4: Testing & Documentation

- [x] 4.1 Extend `scripts_test.go` `TestInstallPs1ChecksumMandatory`: verify retry wrapper with `Start-Sleep` backoff pattern (gap at lines 185–207)
- [x] 4.2 Manually verify pinning check: add fake `@v1` ref → CI fails; remove → CI passes
- [x] 4.3 Document action-pinning policy and release environment setup in `docs/CONTRIBUTING.md`
