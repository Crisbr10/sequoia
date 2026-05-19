# ci-security-gates Specification

## Purpose

CI/CD pipeline hardening: SBOM generation, vulnerability scanning, ARM64 CI coverage, Dependabot grouping, and pre-release test gating.

## Requirements

### REQ-SG-001: SBOM Generation

The release workflow MUST generate SPDX 2.3 and CycloneDX SBOMs from go.sum and attach them to GitHub Releases as release artifacts.

#### Scenario: SBOMs attached to release

- GIVEN a tag push triggers release.yml
- WHEN GoReleaser completes successfully
- THEN SPDX 2.3 and CycloneDX SBOM artifacts SHALL appear in the GitHub Release

#### Scenario: SBOM generation failure

- GIVEN go.sum is malformed or absent
- WHEN the SBOM generation step executes
- THEN the release workflow MUST fail and publish no artifacts

### REQ-SG-002: Vulnerability Scanning

CI workflow MUST run `govulncheck ./...` on every push and PR. `.golangci.yaml` MUST enable `gosec` linter with exclusion rules for G104 (unchecked errors in main.go) and G304 (path traversal false positives).

#### Scenario: Clean scan passes

- GIVEN no known CVEs exist in direct or transitive dependencies
- WHEN govulncheck runs in ci.yml
- THEN the step SHALL pass and the pipeline SHALL continue

#### Scenario: CVE detected

- GIVEN a dependency has a known vulnerability
- WHEN govulncheck runs
- THEN the step MUST fail and block the merge/PR

#### Scenario: gosec false positive avoided

- GIVEN code triggers G104 (unchecked err) in main.go or G304 (file path) in safe paths
- WHEN gosec lints the file
- THEN the exclusion rule MUST suppress the false positive

### REQ-SG-003: ARM64 CI Matrix

CI test matrix MUST include `macos-14` and `ubuntu-24.04-arm` runners. `fail-fast: false` MUST remain set.

#### Scenario: ARM64 macOS CI passes

- GIVEN a push or PR targets main
- WHEN CI triggers the macos-14 job
- THEN tests, build, and smoke checks SHALL complete on Apple Silicon

#### Scenario: ARM64 Linux CI passes

- GIVEN a push or PR targets main
- WHEN CI triggers the ubuntu-24.04-arm job
- THEN tests, build, and smoke checks SHALL complete on ARM64 Linux

#### Scenario: ARM64 runner flaky

- GIVEN ubuntu-24.04-arm is temporarily degraded
- WHEN ARM64 job fails due to runner infra
- THEN fail-fast=false SHALL allow x86_64 jobs to succeed independently

### REQ-SG-004: Dependabot Groups

Dependabot SHALL use groups: gomod updates grouped by minor/patch, github-actions grouped together. Each group MUST set `reviewers: ["Crisbr10"]` and `labels: ["dependencies"]`.

#### Scenario: Grouped gomod PR

- GIVEN multiple minor/patch Go dependency updates are available
- WHEN Dependabot runs on schedule
- THEN a single PR SHALL group all gomod updates with label "dependencies"

#### Scenario: Grouped actions PR

- GIVEN multiple github-actions version updates are available
- WHEN Dependabot runs
- THEN a single PR SHALL group all action updates

#### Scenario: Reviewer assigned

- GIVEN Dependabot opens any PR
- WHEN the PR is created
- THEN Crisbr10 MUST be assigned as reviewer

### REQ-SG-005: Pre-release Test Gate

Release workflow MUST execute `go test ./...` as a prerequisite before GoReleaser runs.

#### Scenario: Tests pass before release

- GIVEN a version tag (vX.Y.Z) is pushed
- WHEN the release workflow runs the test gate job
- THEN passing tests SHALL allow GoReleaser to proceed

#### Scenario: Tests fail before release

- GIVEN a version tag is pushed with failing tests
- WHEN the test gate job runs
- THEN GoReleaser MUST NOT execute and the release MUST be blocked
