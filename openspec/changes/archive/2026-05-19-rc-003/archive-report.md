# SDD RC-003 Archive Report — Add CI/CD Security Gates

## Change Summary

**Change**: RC-003 "Add CI/CD Security Gates"
**Status**: ARCHIVED — PASS WITH WARNINGS
**Date**: 2026-05-19

### What Was Delivered

CI/CD pipeline hardening across 5 security gates, addressing 5 Sequoia audit findings (P6-001, P6-003, P6-004, P6-005, P6-008). All changes are additive single-file modifications. Zero Go application code changes — pure CI/CD infrastructure.

### Files Changed

| File | Change |
|------|--------|
| `.github/workflows/ci.yml` | +govulncheck step (Linux-only), +macos-14, +ubuntu-24.04-arm in test matrix |
| `.github/workflows/release.yml` | +pre-release test job, +needs:test on goreleaser, +anchore/sbom-action@v0 SBOM step |
| `.github/dependabot.yml` | +groups (gomod minor/patch, github-actions), +reviewers, +labels |
| `.golangci.yaml` | +gosec linter, +exclude-rules for G104/G304 |
| `scripts/install.sh` | +cosign verify-blob block (~47 lines) with graceful fallback |
| `scripts/install.ps1` | +cosign verify-blob block (~58 lines) with PowerShell-native equivalent |

### Capabilities Delivered

| Capability | Status | Key Requirement |
|------------|--------|-----------------|
| `ci-security-gates` | ACTIVE | REQ-SG-001 through REQ-SG-005 (SBOM, vuln scanning, ARM64 CI, Dependabot groups, pre-release test gate) |
| `installer-signature-verification` | ACTIVE | REQ-IV-001 through REQ-IV-003 (cosign for install.sh/install.ps1, graceful fallback) |

### Tasks Completed

8/8 tasks, 4 stacked PRs, ~172 changed lines across 6 files:

| PR | Tasks | Files | Lines |
|----|-------|-------|-------|
| PR1: Foundation | T-001, T-002, T-003 | 3 | ~35 |
| PR2: ARM64 Matrix | T-004 | 1 | ~10 |
| PR3: Release Hardening | T-005, T-006 | 1 | ~22 |
| PR4: Installer Signatures | T-007, T-008 | 2 | ~105 |

## Verification Verdict

**PASS WITH WARNINGS** — 1 procedural warning (non-blocking):

| Check | Result |
|-------|--------|
| Tasks 8/8 complete | PASS |
| Spec coverage 7/7 requirements | PASS |
| Design conformance | PASS |
| YAML validity (4 files) | PASS |
| Go test suite (18/18 packages) | PASS (0 failures) |
| install.ps1 syntax (PSParser) | PASS |
| TDD Cycle Evidence | WARNING (not applicable — infrastructure-only change) |

## Deferred Items

| Item | Reason |
|------|--------|
| Windows ARM64 CI | GitHub Actions has no Windows ARM64 runner |
| Docker scanning | No Dockerfile exists in the project |
| SLSA Level 3+ provenance | Overkill for Go 1.24 CLI tool maturity |

## Open Items (from Design)

| # | Question | Status |
|---|----------|--------|
| 1 | Does `anchore/sbom-action@v0` `upload-release-assets` work with draft releases? | Monitor first release |
| 2 | Is `ubuntu-24.04-arm` runner GA for this repo? | Verify when CI runs on main |

## Rollback Plan

Each component is a single-file revert, independent of others. See design.md for per-component instructions.

## Artifact Traceability

| Artifact | Engram ID | Filesystem Path |
|----------|-----------|-----------------|
| Proposal | #446 | proposal.md |
| Specs (delta) | #447 | specs/ |
| Design | #448 | design.md |
| Tasks | #449 | tasks.md |
| Apply Progress | #450 | (Engram only) |
| Verify Report | #451 | (Engram only) |
| Archive Report | #452 | archive-report.md |

## Specs Synced

| Domain | Action | Details |
|--------|--------|---------|
| ci-security-gates | Created | 5 requirements (REQ-SG-001 through REQ-SG-005), written directly during spec phase |
| installer-signature-verification | Created | 3 requirements (REQ-IV-001 through REQ-IV-003), written directly during spec phase |

Both main specs at `openspec/specs/` already reflect the implemented behavior — no post-hoc merge needed.

## Next Steps

1. Push the 4 stacked PRs to GitHub
2. Verify CI runs (govulncheck passes, ARM64 runners start)
3. Watch for Dependabot grouped PRs
4. Monitor first release (SBOM attachment, cosign signatures)
5. Resolve open questions (ubuntu-24.04-arm GA, draft release compatibility)
