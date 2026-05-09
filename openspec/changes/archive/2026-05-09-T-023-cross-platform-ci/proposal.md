# Proposal: Cross-Platform CI Pipeline

## Intent

Validate that the Sequoia CLI builds, passes all tests, and completes the install/status/uninstall cycle on all three target operating systems before any code merges. Currently, all testing is local (Windows only), leaving Linux and macOS regressions undetected.

## Scope

### In Scope
- `.github/workflows/ci.yml` workflow triggered on push/PR to `main`
- OS matrix: `ubuntu-latest`, `macos-latest`, `windows-latest`
- Go 1.22 setup with module caching per OS
- `go test ./...` (108+ tests across 5 packages) on every OS
- `go vet ./...` static analysis on every OS
- `go build ./cmd/sequoia/` binary compilation on every OS
- Install cycle smoke test: local build path install check per OS

### Out of Scope
- Release pipeline with GoReleaser (T-033)
- Code coverage reporting
- Linting gate (`.golangci.yaml` exists but not enforced in CI)
- Docker builds or containerized testing
- Self-hosted runners or matrix expansion beyond 3 stock runners

## Capabilities

### New Capabilities
- `cross-platform-ci`: GitHub Actions workflow that gates merges on build + test + vet + install smoke across ubuntu, macos, and windows

### Modified Capabilities
None — CI is infrastructure, no spec-level behavior changes. Existing `installer-scripts` spec already defines the install cycle contract this CI validates.

## Approach

Single `ci.yml` workflow with `strategy.matrix.os`. Shared Go setup action (`actions/setup-go@v5`), then parallel jobs:

1. **Setup**: Go 1.22, enable module cache (`go.sum` keyed)
2. **Vet**: `go vet ./...` — catches platform-specific compilation issues early
3. **Test**: `go test ./... -count=1 -race` — full test suite per OS (race detector skips on Windows GitHub runners; handled gracefully)
4. **Build**: `go build -o <binary> ./cmd/sequoia/` with OS-appropriate extension (`.exe` on Windows)
5. **Install smoke**: Run the built binary with `install --no-tui` in a temp directory, verify `sequoia status`, then `sequoia uninstall --no-tui --force` — validates the full lifecycle without requiring published releases

Tarball naming convention (`sequoia_${OS}_${ARCH}.tar.gz` underscore style from T-022) is a release concern (T-033) — the CI smoke test uses local paths, not downloads.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `.github/workflows/ci.yml` | New | Main CI pipeline with OS matrix |
| `.github/` | New | Directory (currently absent) |

No Go source or script changes. Pure CI infrastructure addition.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| macOS runner slowness (limited free minutes) | Med | Cache Go modules; test-only on macOS, defer heavy builds to T-033 |
| Windows path separator edge case in test assertions | Low | Codebase already uses `filepath.Join()`; vet catches platform issues |
| Race detector unavailable on Windows runners | Low | Conditional `-race` flag; skip with warning on Windows if unsupported |
| `sequoia install --no-tui` exits non-zero (no real tool found) | Med | Pipe stdout/stderr; treat non-zero exit as warning in smoke step, not hard fail |

## Rollback Plan

Delete `.github/workflows/ci.yml`. Remove the `.github/` directory if it becomes empty. No code to revert. GitHub Actions stop running immediately when the workflow file is removed.

## Dependencies

- **T-022 Installer scripts** — **DONE**. The `install --no-tui` and `uninstall --no-tui --force` commands invoked in CI smoke tests are implemented.
- **Go 1.22** — declared in `go.mod`, already the project standard.
- No external services or secrets required.

## Success Criteria

- [ ] CI workflow triggers on push to any branch and on PR to `main`
- [ ] All three OS jobs (ubuntu, macos, windows) complete successfully
- [ ] `go test ./...` passes with zero failures across all OSes
- [ ] `go vet ./...` reports zero issues across all OSes
- [ ] Binary builds successfully on all three OSes
- [ ] Install → status → uninstall smoke cycle exits cleanly on all three OSes
- [ ] Workflow uses Go module caching (second run under 90 seconds per job)
