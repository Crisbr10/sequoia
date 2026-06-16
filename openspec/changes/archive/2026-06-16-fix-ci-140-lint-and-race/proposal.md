# Proposal: fix-ci-140-lint-and-race

## Intent

CI run #140 failed on commit `20f352d` (v1.0.35 housekeeping). Three failure classes need fixing before the release is trustworthy:

1. **Lint (3 errors)** — `golangci-lint run ./...` reports 6 issues: 5 gofmt formatting violations and 1 unused function (`internalFileExists` in `base_adapter_internal_test.go`).
2. **Race condition (data race)** — Tests fail on Linux/macOS/ARM with `go test -race` (CI drops `-race` on Windows, which is why Windows passes). The race is in the test pollution interaction identified in the PR 3b verify report: parallel tests in `base_adapter_error_test.go` share the central home via the package-level `userConfigDir` variable. When one parallel test's `applyRetention` hook prunes the session dirs, it can remove a dir that a concurrent failed-install rollback is reading from.
3. **CI hygiene** — The release is blocked. Fixing these restores the green CI gate.

## Scope

### In Scope
- Fix 6 lint issues: remove unused `internalFileExists`; run `gofmt -w` on the 5 files reported by `golangci-lint`.
- Fix the test race: isolate parallel tests in `base_adapter_error_test.go` — each test that calls `BackupHomeDir()` or triggers `applyRetention` must use `overrideUserConfigDir(t, func() { return t.TempDir(), nil })` so retention pruning in one test cannot affect another.
- Verify locally: `go test ./... -race -count=1 -timeout 180s` passes (requires CGO; if unavailable, verify via `go test ./... -count=1` and CI confirmation).
- Confirm CI #141 (next run) is green on all 5 platforms.

### Out of Scope
- Adding `-race` to the Windows CI workflow (intentional — Windows race detector is slower and less reliable per project convention).
- Refactoring `applyRetention`, `PruneBackups`, or `BackupHomeDir` beyond race isolation.
- Changing the test infrastructure broadly.
- Reverting the `backup-retention-and-organization` change.

## Capabilities

### New Capabilities
- `ci-green-gate`: CI passes on all 5 platforms. Lint, vet, test (with `-race` on non-windows), build, smoke, action-pinning, vulncheck all green. (New full spec at `openspec/specs/ci-green-gate/spec.md`.)

### Modified Capabilities
- None.

## Approach

1. **Fix lint**: `gofmt -w` on the 5 gofmt-violation files; delete the unused `internalFileExists` function from `base_adapter_internal_test.go`.
2. **Fix race**: In `base_adapter_error_test.go`, every test that calls `BackupHomeDir()` or triggers `BaseAdapter.Apply()` must call `overrideUserConfigDir(t, func() (string, error) { return t.TempDir(), nil })` before the test body runs. This is the same pattern already used by `backup_retention_test.go` and `base_adapter_retention_test.go`.
3. **Verify**: Run `go test ./... -count=1 -timeout 180s` (local; no CGO) → 20/20 pass. Push and let CI confirm the race is resolved.
4. **Push**: Commit the fixes; CI #141 gate runs with `-race` on 4 platforms.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `adapters/common/base_adapter_internal_test.go:57` | Removed | Delete unused `internalFileExists` function |
| `adapters/opencode/adapter.go:42` | Modified | `gofmt -w` — gofmt formatting fix |
| `adapters/testutil/mock_adapter.go:1` | Modified | `gofmt -w` — gofmt formatting fix |
| `internal/codegraph/install.go:1` | Modified | `gofmt -w` — gofmt formatting fix |
| `internal/codegraph/install_test.go:1` | Modified | `gofmt -w` — gofmt formatting fix |
| `internal/tui/styles/logo.go:1` | Modified | `gofmt -w` — gofmt formatting fix |
| `internal/tui/styles/styles.go:1` | Modified | `gofmt -w` — gofmt formatting fix |
| `adapters/common/base_adapter_error_test.go` | Modified | Add `overrideUserConfigDir` per-test isolation to all parallel tests that touch `BackupHomeDir` or `BaseAdapter.Apply` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Race is in production code, not just tests | Low | The PR 3b verify report explicitly identified test-pollution as the cause; `applyRetention` is best-effort and called only after successful `Apply`. Production install flow is single-threaded. |
| `gofmt -w` changes line endings (CRLF on Windows) | Low | `gofmt -w` normalizes to LF; the 5 reported files will have consistent line endings. Pre-existing CRLF files in `adapters/common/` are a known pattern (documented in PR 3b verify report). |
| Cannot verify race locally (no CGO) | Medium | Run `go test ./... -count=1` locally; rely on CI #141 to confirm the race fix. If CI still fails, the apply phase will have the exact race report from CI #140. |

## Rollback Plan

Revert the commit: `git revert <sha>` → push → CI returns to red. No data loss; purely formatting + test-isolation changes.

## Dependencies

- None external.

## Success Criteria

- [ ] `golangci-lint run ./...` reports 0 issues
- [ ] `go test ./... -count=1 -timeout 180s` passes locally (20/20 packages)
- [ ] After push, CI #141 is green on all 5 platforms
- [ ] No regression in the 30 tasks of the `backup-retention-and-organization` change
