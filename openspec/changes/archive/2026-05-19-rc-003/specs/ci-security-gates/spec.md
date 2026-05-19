# ci-security-gates Specification (Delta)

## Purpose

CI/CD pipeline hardening: SBOM generation, vulnerability scanning, ARM64 CI coverage, Dependabot grouping, and pre-release test gating.

## Requirements

### REQ-SG-001: SBOM Generation

The release workflow MUST generate SPDX 2.3 and CycloneDX SBOMs from go.sum and attach to GitHub Releases.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| SBOM attached | Tag push triggers release.yml | GoReleaser completes | SPDX + CycloneDX SBOMs uploaded as release artifacts |
| SBOM failure | go.sum malformed/absent | SBOM action runs | Workflow fails, no artifacts published |

### REQ-SG-002: Vulnerability Scanning

CI MUST run govulncheck ./... on push/PR. gosec MUST be enabled in golangci-lint with exclusion rules for G104 (main.go errcheck) and G304 (path traversal false positives).

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Clean scan | No CVEs in deps | govulncheck runs in CI | Step passes, pipeline continues |
| CVE detected | Known CVE in dep | govulncheck runs | Step fails, blocks merge |
| gosec FP avoided | Safe code triggers G104/G304 | gosec lints file | Exclusion suppresses false positive |

### REQ-SG-003: ARM64 CI Matrix

Matrix MUST include macos-14 and ubuntu-24.04-arm runners. fail-fast: false MUST remain.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| ARM macOS | Push to main | macos-14 job runs | Tests/build/smoke pass on Apple Silicon |
| ARM Linux | Push to main | ubuntu-24.04-arm job runs | Tests/build/smoke pass on ARM64 Linux |
| ARM flaky | ARM runner degraded | ARM job fails | x86_64 jobs succeed independently |

### REQ-SG-004: Dependabot Groups

SHALL group gomod (minor/patch) and github-actions updates. MUST set reviewers: ["Crisbr10"], labels: ["dependencies"].

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Grouped gomod | Multiple minor Go dep updates | Dependabot runs | Single PR groups all gomod, labeled dependencies |
| Grouped actions | Multiple action version updates | Dependabot runs | Single PR groups all actions |
| Reviewer | Any Dependabot PR opens | PR created | Crisbr10 assigned as reviewer |

### REQ-SG-005: Pre-release Test Gate

Release workflow MUST run go test ./... before GoReleaser.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Tests pass | Version tag pushed | Test gate runs | GoReleaser proceeds |
| Tests fail | Version tag, broken tests | Test gate runs | GoReleaser blocked, release fails |
