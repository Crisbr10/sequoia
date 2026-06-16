# Design: backup-retention-and-organization

## Technical Approach

Consolidate scattered `.sequoia-backup-*` dirs into a single central root (`<UserConfigDir>/sequoia/backups/{adapterID}/{timestamp-suffix}/`) and enforce a 5-session-per-adapter retention cap at the end of every successful `BaseAdapter.Apply()`. Two helpers (`BackupHomeDir`, `PruneBackups`) live in a new `adapters/common/backup_retention.go`; `ReplaceFile`/`RestoreOrRemoveFile` switch to a single `manifest.json` per session; the TUI `Info` notes pre-existing scattered backups. Delivery: 3-PR stacked chain honoring the 400-line budget (`force-chained` preflight).

## Module layout & call graph

```
adapters/common/
├── backup_retention.go      [NEW]  BackupHomeDir, PruneBackups, DefaultMaxBackupsPerAdapter
├── backup_retention_test.go [NEW]
├── backup_path_builder.go   [MOD]  delegates to BackupHomeDir
├── base_adapter.go          [MOD]  Apply() calls applyRetention() before return nil (line 519)
├── strategy.go              [MOD]  ReplaceFile/RestoreOrRemoveFile → session dir + manifest.json
├── strategy_test.go         [MOD]  update .sequoia-backup- paths
├── installer_test.go        [MOD]  update .sequoia-backup- paths
adapters/<tool>/paths.go (×5)            [MOD] backupPath() → central home
adapters/codex/installer_test.go         [MOD]  adapters/opencode/install_test.go [MOD]
internal/pipeline/runner.go              [MOD] Info (line 200-210) notes legacy backups

BaseAdapter.Install → Apply → applyRetention
                                   → PruneBackups(adapterID, 5)
                                   → readDir + sort + os.RemoveAll (oldest first)
                                   → on err: AddWarning → WarningEmitter → TUI
```

## Architecture Decisions

| # | Choice | Rejected |
|---|--------|----------|
| 1 | `BackupHomeDir() (string,error)`, `PruneBackups(id,max)`, `DefaultMaxBackupsPerAdapter=5` in `adapters/common/backup_retention.go` — matches `os.UserConfigDir` family; no stutter; verb-form prune; const exported | `SequoiaBackupHome()` / `PruneOldBackups()` (stutters; "Old" redundant) |
| 2 | single `manifest.json` per session; `encoding/json`; schema `{version, created_at, entries:[{version, original_path, suffix, created_at, adapter_id}]}` — stdlib only; parser-friendly; extensible | per-file sidecars (the antipattern we fix); plain text (unparseable) |
| 3 | hook at end of `BaseAdapter.Apply()` (before `return nil` line 519) via private `applyRetention()`; best-effort, errors → `AddWarning` — `Apply` is the last phase that mutates user files; `Install()` is a thin wrapper; `Stage`/`Verify` not idempotent | hook in `Install` (couples orchestration); `Stage→Apply` boundary (runs on partial Stage failure) |
| 4 | single `manifest.json` per session listing only `ReplaceFile`-backed files; installer-style backups use `InstallerConfig.TargetDir` — single file is the natural unit; installer backups already carry restore target | per-file sidecars (recreates the problem) |
| 5 | 3 stacked PRs to `main` (≤400L each): PR1 `backup_retention.go` + `BackupPathBuilder`; PR2 installer wiring + 4+ test updates + TUI note; PR3 `ReplaceFile`/manifest/retention hook + per-adapter `paths.go` — matches `force-chained` + 400L budget; each PR independently mergeable | single 1000L PR (exceeds budget); 2 PRs (PR1 too big) |

## API surface & hook

```go
const DefaultMaxBackupsPerAdapter = 5

// BackupHomeDir returns <UserConfigDir>/sequoia/backups, creating it (0o700).
func BackupHomeDir() (string, error)

// PruneBackups keeps the `max` most recent session dirs for adapterID (ISO-8601 sort).
// Returns (count removed, first error); continues on per-entry error. (0, nil) on miss.
func PruneBackups(adapterID string, max int) (removed int, err error)
```

Hook (insertion in `base_adapter.go:Apply`, just before `return nil` line 519):

```go
func (a *BaseAdapter) applyRetention() {
    if _, err := PruneBackups(a.ID(), DefaultMaxBackupsPerAdapter); err != nil {
        a.AddWarning(fmt.Sprintf("backup retention: %v", err))
    }
}
// in Apply(), after AtomicWriteFile(versionFile) succeeds:
a.applyRetention()
return nil
```

## Data shapes

Manifest types use stdlib `encoding/json`; field names + JSON tags per decision 2 schema.

| File | Action | Lines |
|------|--------|-------|
| `adapters/common/backup_retention.go` (+ test) | Create | ~250 |
| `adapters/common/backup_path_builder.go` (+ test) | Modify | +20 |
| `adapters/common/base_adapter.go` | Modify | +12 |
| `adapters/common/strategy.go` (PR3) | Modify | +60 |
| `adapters/common/strategy_test.go` (PR2/PR3) | Modify | +25 |
| `adapters/common/installer_test.go` (PR2) | Modify | +20 |
| `adapters/<tool>/paths.go` (×5, PR3) | Modify | +5 each |
| `adapters/codex/installer_test.go` (PR2) | Modify | +15 |
| `adapters/opencode/install_test.go` (PR2) | Modify | +10 |
| `internal/pipeline/runner.go` (PR2) | Modify | +3 |

## Testing Strategy

**Unit** (`backup_retention_test.go`, override `UserConfigDir`): `BackupHomeDir` returns joined path, creates mode 0o700, idempotent, wraps errors. `PruneBackups`: keeps exactly `max`; no-op below max; missing-adapter → (0,nil); corrupt names ignored; continues on per-entry error (table-driven testify). `BackupPathBuilder.Build` uses central home (updated test). `ReplaceFile` round-trip via `manifest.json` (updated `strategy_test.go`). 4+ existing tests (`strategy_test.go`, `installer_test.go`, `codex/installer_test.go`, `opencode/install_test.go`) updated to assert the central path.
**E2E**: 6 successive installs on one adapter → exactly 5 session dirs (new test in `adapters/common/`); TUI `Info` includes central path + legacy-backup note (`internal/pipeline/runner_test.go`).
**TDD**: every new function gets a RED test first; updated existing tests go RED→GREEN as paths change.

## Migration / Rollout

**No migration** of pre-existing `.sequoia-backup-*` (REQ-BRP-05). TUI `Info` at `runner.go:200-210` gets a one-line note that pre-existing scattered backups remain at their original locations. Rollout: 3 chained PRs to `main`, each independently shippable; no feature flag.

## Open Questions / Risks

- **Corrupt `manifest.json`**: degrade to scanning session dir for `*.backup` (best-effort, warn). Deferred to sdd-apply; safe — each session has ≤1 backup.
- **Windows `%APPDATA%`**: returns Roaming; symlink/redirect handled by `ResolveHome`.
- **Concurrent installs**: cross-process race possible. Acceptable for v1 (single-user, sequential CLI).
- **Adapter-ID drift**: pre-rename sessions orphaned at old ID subdir; not migrated.
