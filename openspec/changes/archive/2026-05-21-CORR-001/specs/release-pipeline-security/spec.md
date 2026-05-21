# Release Pipeline Security Specification

## Purpose

Environment protection, approval gates, binary verification, cross-platform tests, manual dispatch, post-deploy smoke, CODEOWNERS.

## Requirements

### Requirement: Environment Protection

The release workflow MUST target `environment: release` with required reviewers. GoReleaser SHALL NOT proceed without approval.

#### Scenario: Release gated by environment

- GIVEN release workflow triggered
- WHEN GoReleaser job starts
- THEN GitHub MUST hold deployment pending reviewer approval

### Requirement: Manual Approval Gate

Publishing MUST NOT execute without manual approval, regardless of trigger.

#### Scenario: Tag push blocked

- GIVEN release tag pushed
- WHEN workflow runs
- THEN GoReleaser MUST wait for explicit manual approval

### Requirement: Binary Integrity

SHA-256 checksum of built binary MUST be verified before cosign. Mismatch MUST fail the workflow.

#### Scenario: Bad checksum halts release

- GIVEN build artifact ready, SHA-256 mismatches expected
- WHEN verification runs
- THEN workflow MUST fail with checksum-mismatch error

### Requirement: Cross-Platform Tests

Pre-release tests MUST pass on `ubuntu-latest`, `macos-latest`, `windows-latest` via matrix. Any failure SHALL abort.

#### Scenario: Platform failure stops release

- GIVEN test matrix runs on all three OSes
- WHEN any platform fails
- THEN release SHALL abort

### Requirement: Manual Dispatch

The release workflow SHALL support `workflow_dispatch` for operator-initiated releases.

#### Scenario: UI-triggered release

- GIVEN user with write access dispatches release from Actions UI
- WHEN workflow executes
- THEN identical gates apply as tag-push

### Requirement: Post-Deploy Smoke

After publishing, workflow SHALL verify one artifact is downloadable with matching checksum.

#### Scenario: Smoke confirms binary

- GIVEN GoReleaser completed
- WHEN smoke downloads binary
- THEN download SHALL succeed within 30s, checksum SHALL match

### Requirement: CODEOWNERS

`.github/CODEOWNERS` SHALL cover `release.yml` and `install.ps1`. Changes MUST require owner review.

#### Scenario: Workflow change needs owner

- GIVEN CODEOWNERS covers release-critical paths
- WHEN PR modifies `release.yml`
- THEN GitHub SHALL require CODEOWNER review
