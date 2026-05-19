# Proposal: RC-003 — Add CI/CD Security Gates

## Intent

sequoia is an AI code audit tool — users MUST trust the binary they install. Five supply-chain gaps leave releases unsigned at the install layer, untested on ARM64, unscanned for CVEs, and without an SBOM. Closing them gives users cryptographic proof of artifact integrity, CI validation on every target architecture, and known-vulnerability detection at every push.

## Scope

### In Scope
- SBOM generation (SPDX + CycloneDX) in release.yml, attached to GitHub Releases
- `govulncheck ./...` and enabled `gosec` in golangci-lint in ci.yml
- `macos-14` (Apple Silicon) and `ubuntu-24.04-arm` runners in the CI test matrix
- Dependabot `groups`, `reviewers`, and `labels` for both gomod + github-actions
- Cosign signature verification in install.sh and install.ps1 with SHA-256 fallback
- Pre-release test gate in release.yml (run tests before GoReleaser)

### Out of Scope
- Windows ARM64 CI (GitHub Actions has no Windows ARM64 runner)
- Docker scanning (no Dockerfile exists)
- SLSA Level 3+ provenance (overkill for Go 1.24 CLI tool maturity)
- npm/pip/submodules Dependabot ecosystems (not applicable)

## Capabilities

### New Capabilities
- `ci-security-gates`: SBOM generation, vulnerability scanning, ARM64 CI matrix, dependabot improvements, pre-release test gate
- `installer-signature-verification`: Cosign verify-blob in install.sh and install.ps1 with graceful fallback when cosign is absent

### Modified Capabilities
- None (pure CI/infrastructure changes — no application behavior changes)

## Approach

| Finding | Strategy |
|---------|----------|
| P6-001 (SBOM) | `anchore/sbom-action@v0` in release.yml after GoReleaser. Generates SPDX 2.3 + CycloneDX from go.sum. Attaches both as release artifacts. |
| P6-003 (SAST/Vuln) | `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` step in ci.yml. Enable `gosec` in `.golangci.yaml` with exclusions for G104 (unchecked errors in main.go) and G304 (path traversal false positives). |
| P6-004 (ARM64 CI) | Add `macos-14` and `ubuntu-24.04-arm` to ci.yml matrix (free for public repos). Keep `windows-latest` (x64 only). `fail-fast: false` already protects against flaky ARM runners. |
| P6-005 (Dependabot) | Add `groups` (gomod grouped by minor/patch, actions grouped), `reviewers: ["Crisbr10"]`, `labels: ["dependencies"]`. |
| P6-008 (Signatures) | In install.sh/install.ps1: after checksum verification, attempt `cosign verify-blob` using `.sig` + `.cert` from the release. Gracefully skip if cosign is not installed. Log a warning recommending cosign installation. |

## Impact

| File | Change |
|------|--------|
| `.github/workflows/ci.yml` | +govulncheck step, +macos-14 and ubuntu-24.04-arm in matrix |
| `.github/workflows/release.yml` | +pre-release test job, +SBOM generation step after GoReleaser |
| `.github/dependabot.yml` | +groups (gomod, github-actions), +reviewers, +labels |
| `.golangci.yaml` | +gosec linter in enabled list, +exclude rules for G104/G304 |
| `scripts/install.sh` | +cosign verification block (~30 lines) after checksum block |
| `scripts/install.ps1` | +cosign verification block (~30 lines) after checksum block |

**CI runtime**: +~2 min (govulncheck ~30s, ARM64 runners comparable to x86). Release workflow: +~1 min (SBOM + pre-release test). No breaking changes to existing workflows or scripts.

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| gosec false positives (G104, G304) | Medium | Exclusion rules in `.golangci.yaml` for known-safe patterns (main.go errcheck, filepath.Join paths) |
| cosign not installed on user machines | Medium | Graceful fallback to SHA-256 only; warn + recommend cosign installation |
| ARM64 runner flakiness (ubuntu-24.04-arm is public beta) | Low | `fail-fast: false` already set; x86_64 matrix remains the quality gate |
| Pre-release test gate blocks release on flaky tests | Low | Test suite already stable; `-race` conditional already handled in CI |

## Rollback Plan

Each change is independently reversible: remove govulncheck step from ci.yml, revert `.golangci.yaml` gosec enable, drop ARM64 matrix entries, remove cosign verification blocks from install scripts. Dependabot groups can be removed without side effects. SBOM generation can be removed from release.yml without affecting GoReleaser output. All rollbacks are single-file reverts.

## Success Criteria

- [x] `govulncheck` runs on every push and PR without errors
- [x] `gosec` passes in CI with zero false-positive-blocking failures
- [x] `macos-14` and `ubuntu-24.04-arm` jobs pass in the test matrix
- [x] Dependabot PRs are grouped, labeled, and assigned to a reviewer
- [x] Release includes both SPDX and CycloneDX SBOM artifacts
- [x] `install.sh` and `install.ps1` verify cosign signatures when cosign is available, fallback gracefully otherwise
- [x] Release workflow runs a test gate before GoReleaser
