# Design: Release Pipeline Security Hardening (CORR-001)

## Technical Approach

Harden the release supply chain through three coordinated configuration changes: (1) restructure `release.yml` with environments, OS matrix, SHA-256 verification, and post-deploy smoke; (2) pin all actions to commit SHAs across both workflow files and add a CI enforcement check; (3) wrap `Invoke-WebRequest` calls in `install.ps1` with exponential-backoff retry. Zero Go code — pure YAML and PowerShell config.

## Architecture Decisions

| # | Decision | Option A | Option B | Choice | Rationale |
|---|----------|----------|----------|--------|-----------|
| 1 | Approval gate mechanism | `environment: release` with required reviewers | Branch protection on tag patterns | **A** | Tag pushes bypass branch protection rules. GitHub Environments provide built-in deployment gates, reviewer requirements, and audit history. Already depends on pre-existing `release` environment. |
| 2 | Cross-platform matrix | Independent 3-OS matrix in release.yml (ubuntu, macos, windows) | Reuse CI's 5-OS matrix | **A** | Release needs only representative platform coverage, not the full CI matrix (which includes arm64 and macos-14). Simpler, faster, and avoids coupling release to CI config changes. |
| 3 | Smoke test source | Download from published GitHub Release URL | Reuse local build artifact | **A** | Spec requires verifying the *published* artifact is downloadable with matching checksum. Local artifact smoke proves the build works; release smoke proves the *distribution* works. |
| 4 | Binary verification timing | SHA-256 of GoReleaser output → compare against `checksums.txt` → cosign | Post-build SHA compare only | **A** | Spec mandates "verify before cosign." GoReleaser generates `checksums.txt`; we compute SHA of the built binary against it, then cosign signs. Catches corruption before signing. |
| 5 | Retry strategy | Exponential backoff: 2s → 4s → 8s (3 attempts) | Fixed delay (5s, 3 attempts) | **A** | Exponential backoff is the spec requirement (SHOULD). Standard pattern for transient network recovery. PowerShell loop with `Start-Sleep` is simple to implement. |
| 6 | Pinning enforcement | CI check job running grep/shell scan | Dependabot-assisted pinning | **A** | Simpler, immediate PR feedback, no external service dependency. Dependabot already handles *updates*; the CI check prevents regressions. |
| 7 | CODEOWNERS scope | `.github/workflows/release.yml`, `scripts/install.ps1`, `.github/CODEOWNERS` | All `.github/` paths | **A** | Narrow scope aligns with spec ("release-critical paths"). Self-referencing CODEOWNERS prevents tampering with the review gate itself. Maintainer: `@Crisbr10` (from existing dependabot.yml). |

## Data Flow

```
Tag push / workflow_dispatch
        │
        ▼
  Pre-release Tests (matrix: ubuntu, macos, windows)
        │  go test -race -count=1 ./...
        ▼
  Build + Verify (steps in goreleaser job)
        │
        ├─► GoReleaser builds binaries → outputs dist/
        ├─► SHA-256 check: sha256sum dist/sequoia-* vs checksums.txt
        └─► Cosign sign (keyless, OIDC)
        │
        ▼
  [environment: release — manual approval gate]
        │
        ▼
  GoReleaser publish (upload assets + checksums.txt + signatures)
        │
        ▼
  Post-Deploy Smoke
        │  curl -L release-url → sha256sum → compare
        ▼
      Done
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `.github/workflows/release.yml` | Modify | Add `workflow_dispatch`, OS matrix test job, SHA-256 verification step in goreleaser job, `environment: release`, post-deploy smoke job. Pin all actions to commit SHAs. |
| `.github/workflows/ci.yml` | Modify | Pin all 5 `uses:` directives to commit SHAs. Add `action-pinning` job: grep-based scan for floating refs, fail on match. |
| `.github/workflows/test-action.yml` | Modify | Pin 2 `uses:` directives to commit SHAs. |
| `.github/CODEOWNERS` | Create | Assign `* @Crisbr10` global + explicit entries for `release.yml`, `install.ps1`, `CODEOWNERS`. |
| `scripts/install.ps1` | Modify | Replace bare `Invoke-WebRequest` (lines 183, 201) with retry function: 3 attempts, delays 2s/4s/8s, exit `$EXIT_NETWORK` on exhaustion. |
| `docs/CONTRIBUTING.md` | Modify | Document action-pinning policy and release process. |

## Key Patterns

**Retry wrapper (PowerShell 5.1 compatible)** — non-obvious pattern:

```powershell
function Invoke-WebRequestWithRetry {
    param([string]$Uri, [string]$OutFile)
    $delays = @(2, 4, 8)
    for ($i = 0; $i -lt 3; $i++) {
        try {
            Invoke-WebRequest -Uri $Uri -OutFile $OutFile -UseBasicParsing -ErrorAction Stop
            return
        } catch {
            if ($i -eq 2) { throw }
            Write-Warn "Attempt $($i+1) failed, retrying in $($delays[$i])s..."
            Start-Sleep -Seconds $delays[$i]
        }
    }
}
```

**Pinning check script** (inline in ci.yml pinning job):

```bash
if grep -nP 'uses:\s+\S+@(?![\da-f]{40})' .github/workflows/*.yml; then
  echo "ERROR: Unpinned action references found above."
  exit 1
fi
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| CI validation | Workflow runs correctly | Push to branch, observe workflow execution end-to-end |
| CI validation | Pinning check passes/fails | Add fake `@v1` ref in PR → verify check fails; remove → verify passes |
| Unit (Go) | Retry function logic | Extend `scripts_test.go` `TestInstallPs1ChecksumMandatory`: verify retry wrapper exists and contains `Start-Sleep` with backoff pattern (already has a test gap marker at line 185–207) |
| CI validation | Smoke test verifies binary | Post-release: confirm smoke job downloads artifact and checksum matches |
| Manual | CODEOWNERS enforcement | Open PR modifying release.yml → confirm review requirement is enforced |

## Migration / Rollout

No migration required. All changes are additive configuration. Rollback is a single `git revert` commit. The `release` GitHub Environment must be created before the workflow runs (documented in CONTRIBUTING.md setup steps).

## Open Questions

None — all decisions resolved in the architecture table above.
