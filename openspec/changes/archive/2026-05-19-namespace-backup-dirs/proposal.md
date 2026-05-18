# Proposal: Namespace Backup Directories

## Intent

Fix silent data loss during install rollback. Both `skillInstaller` and `cmdInstaller` currently share the same `backupDir`. When `cmdInstaller.Run()` fails and its internal `Rollback()` calls `os.RemoveAll(backupDir)`, it destroys the backups that `skillInstaller.Prepare()` created. The subsequent `skillInstaller.Rollback()` then finds the directory gone — silently losing the user's original skill file. Additionally, backup directories are not namespaced by adapter ID, risking collisions across concurrent adapter installs.

## Scope

### In Scope
- Namespace backup dirs by adapter ID (`a.ID()`) to prevent cross-adapter collisions
- Give `skillInstaller` and `cmdInstaller` independent subdirectories within the backup dir
- Fix `base_adapter.go:358-394` — shared `Install()` used by claude, gemini, cursor, opencode
- Fix `codex/adapter.go:115-137` — custom `Install()` with same shared-backup pattern
- Fix `_template/adapter.go:196-216` — bootstrapping template with same pattern
- Add test verifying skillInstaller rollback survives cmdInstaller failure

### Out of Scope
- CLI-level backup cleanup or retention policies
- Backup dir permissions (already fixed in backup-permissions change)
- Uninstall backup dir in `base_adapter.go:437-445` (already has per-operation timestamp, single installer there)

## Capabilities

### New Capabilities
- `backup-isolation`: Backup directories MUST be namespaced by adapter ID and installer type to prevent cross-contamination during rollback

### Modified Capabilities
- None — this is a bugfix for existing behavior; no spec-level requirement changes

## Approach

**Phase 1 — Add adapter ID to backup path:**
In `base_adapter.go:358` and `codex/adapter.go:117`, change `backupPath + "-" + sessionSuffix` to `backupPath + "-" + a.ID() + "-" + sessionSuffix`. In `_template/adapter.go:199,210`, add a session suffix (currently missing) and adapter ID.

**Phase 2 — Separate subdirectories per installer type:**
Give `skillInstaller` `filepath.Join(backupDir, "skills")` as its `BackupDir`. Give `cmdInstaller` `filepath.Join(backupDir, "commands")` as its `BackupDir`. Each installer's `Rollback()` removes only its own subdirectory — they never touch each other's backups.

**Phase 3 — Regression test:**
Add `TestInstaller_SkillRollbackSurvivesCmdFailure` to `installer_test.go`. Simulate two installers with separate subdirs; make the commands installer fail; verify the skills installer can still restore its backup.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `adapters/common/base_adapter.go:358-394` | Modified | Add `a.ID()` to backup dir name; split into skills/commands subdirs |
| `adapters/codex/adapter.go:115-137` | Modified | Same namespacing fix in codex's custom Install() |
| `adapters/_template/adapter.go:196-216` | Modified | Add session suffix + adapter ID + subdirectory split |
| `adapters/common/installer_test.go` | Extended | New test: `TestInstaller_SkillRollbackSurvivesCmdFailure` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Backup dir name length grows with adapter ID + suffix + subdir | Low | Adapter IDs are short (≤15 chars); OS path limits sufficient |
| `_template` not compiled in CI (`//go:build ignore`) | Low | Template still reviewed manually; fix serves as reference |
| `filepath.Join` doesn't create subdir — `MkdirAll` needed | Low | `Installer.Prepare()` already calls `os.MkdirAll(cfg.BackupDir, 0o700)` |

## Rollback Plan

Revert the backup dir string construction to its original form in all three files. No data migration needed — backup dirs are ephemeral.

## Dependencies

- None

## Success Criteria

- [ ] `go test -race ./adapters/common/` passes with new regression test
- [ ] `go test -race ./adapters/codex/` passes
- [ ] When cmdInstaller fails at Apply, skillInstaller.Rollback() restores the original skill file from its isolated backup
- [ ] Two adapters installing concurrently use non-overlapping backup directories
