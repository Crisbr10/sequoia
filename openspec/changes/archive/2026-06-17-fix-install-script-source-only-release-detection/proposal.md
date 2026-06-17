# Proposal: Fix Install Script Source-Only Release Detection

## Intent

When a GitHub release exists but contains only source code (no precompiled binaries), the installers emit a generic "Download failed — Error: Not Found". This does not distinguish a missing tag from a source-only release, causing user confusion. The fix adds a specific error message for the source-only release case.

## Scope

### In Scope
- Improve diagnostics in `scripts/install.ps1` (lines ~198–207) to detect and report a source-only release with actionable guidance.
- Apply the same improvement to `scripts/install.sh` (lines ~274–290).
- Add regression tests in `scripts/install_scripts_test.go` asserting the new error pattern is present in both scripts.

### Out of Scope
- Build-from-source fallback in the installer.
- Changes to `.github/workflows/release.yml` or `.goreleaser.yaml`.
- Go code changes in `adapters/`, `internal/`, or `cmd/`.
- Re-publish of v1.0.36.
- Changes to unrelated installer code (cosign verification, PATH manipulation).

## Capabilities

### New Capabilities
- None — diagnostics-only change; no new capability introduced.

### Modified Capabilities
- None — error message text is not a capability-level requirement.

## Approach

On HTTP 404 during binary download, query the GitHub Releases API for the tag. If the tag exists but the binary asset is missing, emit: "Release $ResolvedVersion exists on GitHub but no precompiled binary asset was found. This usually means the release was published without GoReleaser. Try installing a previous version or report this at https://github.com/Crisbr10/sequoia/issues." Fall back to the generic error if the API is unreachable.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `scripts/install.ps1` | Modified | Source-only release detection in download error path |
| `scripts/install.sh` | Modified | Source-only release detection in download error path |
| `scripts/install_scripts_test.go` | Modified | Regression tests for new error pattern |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Release API rate limit during error check | Low | Fall back to generic error message |
| Test fragility (string-based grep) | Low | Use a unique, stable error marker string |

## Rollback Plan

Revert `scripts/install.ps1`, `scripts/install.sh`, and `scripts/install_scripts_test.go` to prior state. No schema or data migration required.

## Dependencies

- GitHub Releases API (`https://api.github.com/repos/{owner}/{repo}/releases/tags/{tag}`) must be reachable from the install environment.

## Success Criteria

- [ ] PowerShell installer with source-only release tag produces the specific source-only error message.
- [ ] Bash installer with source-only release tag produces equivalent specific error message.
- [ ] `go test ./scripts/... -count=1` passes with no regression.
