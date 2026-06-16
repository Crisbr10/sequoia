# Proposal: backup-retention-and-organization

## Intent

Consolidate scattered backup directories into a single central root (`~/.config/sequoia/backups/{adapterID}/{timestamp}/`) and enforce a retention policy that keeps only the 5 most recent backups per adapter. Eliminate backup proliferation across tool config directories while preserving transactional safety.

## Scope

### In Scope
- Central backup root via `os.UserConfigDir()` + `sequoia/backups/{adapterID}/{timestamp}/`
- `DefaultMaxBackupsPerAdapter = 5` constant hardcoded in a new `adapters/common/backup_retention.go`
- Retention prune triggered at the end of each successful `sequoia install` (in `BaseAdapter.Install()` success path)
- TUI `Info` message surfacing old scattered backup locations
- Updated `BackupPathBuilder`, `ReplaceFile`/`RestoreOrRemoveFile`, and `Installer.Prepare()` to use the central path

### Out of Scope
- Migration of existing `.sequoia-backup-*` files/dirs
- Config-driven retention parameters
- Separate CLI command for retention
- Retention on rollback

## Capabilities

### New Capabilities
- `backup-retention-and-organization`: Centralized backup root with 5-per-adapter retention, triggered post-install.

### Modified Capabilities
- None at the spec level (existing `openspec/specs/` is empty; this is the first spec).

## Approach

1. New `adapters/common/backup_retention.go` — exports `DefaultMaxBackupsPerAdapter = 5`, `SequoiaBackupHome()` (central root via `os.UserConfigDir()`), and `PruneOldBackups(adapterID string)` to enforce retention.
2. `BackupPathBuilder` updated to build paths under `SequoiaBackupHome()` instead of per-tool config dirs.
3. `ReplaceFile`/`RestoreOrRemoveFile` in `strategy.go` updated to write to the new central path.
4. `Installer.Prepare()` updated to set `BackupDir` to the new central path.
5. `BaseAdapter.Install()` calls `PruneOldBackups()` on success.
6. TUI `Info` message updated to list old scattered backup locations.
7. Existing tests updated; new tests for retention logic and central-home function.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `adapters/common/backup_retention.go` | New | Central backup root + retention pruner |
| `adapters/common/backup_path_builder.go` | Modified | Use central home instead of per-tool path |
| `adapters/common/strategy.go` | Modified | `ReplaceFile`/`RestoreOrRemoveFile` write to central path |
| `adapters/common/installer.go` | Modified | `BackupDir` for `Prepare()` |
| `adapters/common/base_adapter.go` | Modified | Hook retention into `Install()` success path |
| `adapters/<tool>/paths.go` (×5) | Modified | `backupPath()` delegates to central home |
| `internal/pipeline/runner.go` | Modified | TUI `Info` message for old scattered backups |
| Existing tests (4+) | Updated | Path references updated |
| New tests | Added | Retention pruner, central-home function |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Existing tests need path updates | Medium | Update test fixtures; verify CI passes |
| Users with scattered backups won't see them consolidated | Low | Document in TUI `Info` message |
| `ReplaceFile` moves `.sequoia-backup-*` files — manual diffs need new path | Low | TUI info message explaining the new backup location |
| Hardcoded path on Windows (`%APPDATA%`) | Medium | Use `os.UserConfigDir()` which is platform-correct |

## Rollback Plan

Revert the code changes. Old scattered backups remain untouched; new central backups in `~/.config/sequoia/backups/` can be deleted manually. No user data is migrated.

## Dependencies

- Go 1.25.8 standard library: `os.UserConfigDir()`
- No new external dependencies

## Success Criteria

- [ ] All `sequoia install` operations write backups to `~/.config/sequoia/backups/{adapterID}/{timestamp}/`
- [ ] No more than 5 backup directories per adapter at any time
- [ ] Existing tests pass after path updates
- [ ] New tests cover `SequoiaBackupHome()`, `PruneOldBackups()`, and `ReplaceFile` at central path
- [ ] Coverage threshold (70%) maintained
