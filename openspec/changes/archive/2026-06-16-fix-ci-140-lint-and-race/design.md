# Design: Fix CI #140 lint and race

## Technical Approach

Mechanical CI-gate fix, zero production behavior change. Two issues: (1) 6 lint violations (5 gofmt CRLF + 1 unused function) — `gofmt -w` + one-line deletion; (2) data race in `go test -race` between parallel `BaseAdapter.Apply()` tests sharing the package-level `userConfigDir` (`adapters/common/backup_retention.go:58`) via `applyRetention` → `PruneBackups` → `os.RemoveAll`. Race fix: per-test `overrideUserConfigDir(t, func() (string, error) { return home, nil })` — the established pattern at `backup_retention_test.go:349`, used by 14 callers across `backup_path_builder_internal_test.go`, `base_adapter_internal_test.go`, `base_adapter_retention_test.go`, `strategy_central_test.go`. Maps to REQ-CIG-01..08 of the `ci-green-gate` delta spec.

## Architecture Decisions

| # | Choice | Alternatives (rejected) | Rationale |
|---|--------|-------------------------|-----------|
| 1 | Lint fix = `gofmt -w` + delete `internalFileExists` | Hand-edit each line; manual CRLF→LF | `gofmt -w` is atomic and canonical. `internalFileExists` has 0 callers. `gofmt -d` confirms the 5 files are pure CRLF (56/81/69/74/45/74 CRLF lines, 0 LF). `adapters/opencode/adapter.go` also gets one inline-closure reformat. |
| 2 | Race fix = `overrideUserConfigDir` per test | Per-test unique adapterIDs (doesn't fix shared-state race); drop `-race` (silences real signal); `sync.Mutex` (defeats `t.Parallel()`) | Hook proven in 14 callers. Adding it inside `fullInstallTestAdapter` covers 14 tests in one edit. REQ-CIG-06 forbids touching Windows CI workflow. |
| 3 | Audit call sites (see File Changes) — 3 test files, 20 test functions affected | Skip audit; rely on the helper edit alone | The helper edit misses 5 tests that build BaseAdapter directly (different adapterIDs but same race on the package-level `userConfigDir`). |
| 4 | Use existing helper signature as-is | Export it as `OverrideUserConfigDir` | 14 callers use the lowercase form; no benefit to renaming. Helper already uses `t.Cleanup` to restore. |
| 5 | Single PR, push to main | Chained (lint → race) PRs | Estimated diff: ~50 net lines — well under 400-line budget. Conceptually one deliverable ("CI is green"). `delivery: ask-on-risk` is not triggered. |

## Data Flow

```
  Package-level userConfigDir = os.UserConfigDir
  (shared, mutated by 100+ parallel tests → race)
        |
  overrideUserConfigDir(t, ...) swaps for fn returning t.TempDir()
  (per-test, t.Cleanup-restored)
        |
        v
  BackupHomeDir() ──> <t.TempDir>/sequoia/backups
        |
        v
  a.Install() ──> Apply() ──> applyRetention() ──> PruneBackups(adapterID, 5)
        |
        v
  os.RemoveAll on isolated session dirs (no cross-test contention)
```

Production callers (non-test code) keep using `os.UserConfigDir` — `userConfigDir` is unexported.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `adapters/common/base_adapter_error_test.go` | Modify | Add `overrideUserConfigDir` inside `fullInstallTestAdapter` (line 86); add inline at top of 5 direct-build tests: `StagingDirCreationFailure` (L384), `SkillTemplateNotFound` (L427), `SystemPromptTemplateNotFound` (L469), `BaseResolutionFailure` (L514), `VersionFileWriteFailure` (L575) |
| `adapters/common/base_adapter_test.go` | Modify | Add `overrideUserConfigDir` inside `installTestAdapter` (L198, covers `TestInstall_ReturnsSentinelError`) and `warningsTestAdapter` (L316, covers `TestBaseAdapter_WarningsClearedOnInstall`). `TestBackupIsolation_FreshInstallProducesIdenticalOutput` (L449) and `TestBackupIsolation_NamespacedBackupStructure` (L585) covered by `fullInstallTestAdapter` edit. |
| `adapters/common/base_adapter_mockfs_test.go` | Modify | No edit needed — `TestMockFS_InstallFullPipeline` (L115) and `..._StatusReportsInstalled` (L157) use `fullInstallTestAdapter` |
| `adapters/common/base_adapter_internal_test.go` | Modify | Delete `internalFileExists` (L55–60) |
| `adapters/opencode/adapter.go` | Modify | `gofmt -w` — CRLF→LF + inline-closure reformat |
| `adapters/testutil/mock_adapter.go` | Modify | `gofmt -w` — CRLF→LF |
| `internal/codegraph/install.go` | Modify | `gofmt -w` — CRLF→LF |
| `internal/codegraph/install_test.go` | Modify | `gofmt -w` — CRLF→LF |
| `internal/tui/styles/logo.go` | Modify | `gofmt -w` — CRLF→LF |
| `internal/tui/styles/styles.go` | Modify | `gofmt -w` — CRLF→LF |

The other 9 parallel tests in `adapters/common/` (NilDetector, NilPaths, NilPrompt, HomeDirUnavailable, Uninstall, AddWarning, BaseCachesUserHomeDir, HomeDirOverrideBypassesCache, DetectCached*) do NOT call `Install()`/`Apply()`/`BackupHomeDir()` and need no change.

## Interfaces / Contracts

None new. `overrideUserConfigDir(t *testing.T, fn func() (string, error))` already exists; signature unchanged.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit (race) | Race is gone | `go test ./adapters/common/... -race -count=10 -timeout 300s` (CGO-dependent; CI is source of truth) |
| Integration | All 19 packages pass | `go test ./... -race -count=1 -timeout 180s` × 5 consecutive local runs |
| Lint | Zero issues | `golangci-lint run ./...` exit 0; `gofmt -l .` empty; `go vet ./...` exit 0 |
| E2E | CI green-bar | CI #141+ green on all 5 matrix platforms (Windows drops `-race` per REQ-CIG-06) |

TDD: RED → GREEN → REFACTOR for the race fix. RED = observe race in `go test -race ./adapters/common/... -count=10`. GREEN = add overrides per File Changes. REFACTOR = confirm `fullInstallTestAdapter` is single-point-of-isolation for 14 tests. Lint fix has no RED step (the lint job IS the test).

## Migration / Rollout

No migration. Test-only changes; production code untouched. `userConfigDir` is unexported; production callers go through `BackupHomeDir()` which keeps using `os.UserConfigDir`.

## Open Questions

- [ ] **None.** All 8 REQs have unambiguous paths. CGO availability is a known environment constraint documented in proposal Risk table.
