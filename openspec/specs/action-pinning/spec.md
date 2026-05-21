# Action Pinning Specification

## Purpose

Pin all GitHub Actions to immutable commit SHAs, eliminating floating-tag supply-chain risk.

## Requirements

### Requirement: Commit-SHA Pinning

Every `uses:` directive in `.github/workflows/` MUST reference a full commit SHA. Floating refs (`@vX`, `@main`) are prohibited.

#### Scenario: CI workflow is fully pinned

- GIVEN `.github/workflows/ci.yml`
- WHEN scanned for `uses:` directives
- THEN every entry MUST use `owner/repo@<40-char-sha>`
- AND NO entry SHALL use `@v1`, `@main`, or similar

#### Scenario: Release workflow is fully pinned

- GIVEN `.github/workflows/release.yml`
- WHEN scanned for `uses:` directives
- THEN every entry MUST reference a commit SHA

### Requirement: Pinning Verification

A CI check SHALL verify pinned references. It MAY run on PR affecting workflows or on schedule.

#### Scenario: Floating tag rejected

- GIVEN a PR adds `uses: actions/checkout@v4`
- WHEN the pinning check runs
- THEN the check SHALL fail, identifying the unpinned reference
