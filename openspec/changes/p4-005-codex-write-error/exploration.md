## Exploration: P4-005 — Propagate session file write errors in Codex config merge

### Current State

#### MergeConfig Flow (step by step)

1. **Read existing config** — `os.ReadFile(path)`. If file doesn't exist, `existing` stays empty. Any non-`IsNotExist` error is returned.
2. **Backup existing content** — if `existing != ""`:
   a. Generate timestamp suffix: `strconv.FormatInt(time.Now().UnixMilli(), 36)`
   b. Write backup: `AtomicWriteFile(backupPath, []byte(existing), 0o600)` — errors returned
   c. Write session file: `AtomicWriteFile(path+".sequoia-session", []byte(suffix), 0o644)` — **errors silently discarded** (lines 36-39)
3. **Merge TOML section** — `MergeSection(existing, table)`
4. **Ensure parent dirs** — `os.MkdirAll(filepath.Dir(path), 0o755)`
5. **Write merged result** — `AtomicWriteFile(path, []byte(result), 0o644)` — errors returned

#### Where the error is swallowed (exact line)

**`adapters/codex/installer.go:36-38`**:
```go
if err := common.AtomicWriteFile(path+".sequoia-session", []byte(suffix), 0o644); err != nil {
    // Best-effort: if session file write fails, the backup exists
    // but RemoveConfig will fall back to scanning for backups.
}
```

The comment is misleading: `RemoveConfig` does NOT "scan for backups". It falls back to the **legacy predictable name** `path+".sequoia-backup"` (line 79), which is a single static name. If that backup is stale or from a different installation, the wrong content gets restored.

#### Same pattern exists in ReplaceFile

**`adapters/common/strategy.go:132-135`** — identical silent discard of session file write error inside `ReplaceFile`. Same consequences for `RestoreOrRemoveFile`.

#### RemoveConfig fallback chain when session file is missing

`RemoveConfig` tries in order:
1. Read `path+".sequoia-session"` → get suffix → find timestamped backup → restore
2. If session file exists but backup doesn't: clean up stale session, continue
3. Fall back to `path+".sequoia-backup"` (single legacy name) → restore
4. If no backup at all: parse and remove `[sequoia]` section via `RemoveSection`

**The risk**: If Install ran before, produced backups, and `RemoveConfig` was called successfully, the legacy backup gets cleaned (line 84: `os.Remove(backupPath)`). But if Install ran, session file write failed, then the user manually deleted the timestamped backup, and a **stale** `.sequoia-backup` exists from a previous installation, it gets restored incorrectly.

#### What errors AtomicWriteFile can return

From `adapters/common/strategy.go:223-232`:
```go
func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
    tmp := path + ".tmp"
    if err := os.WriteFile(tmp, data, perm); err != nil {
        return err  // disk full, permission denied, read-only FS, path too long
    }
    if err := os.Rename(tmp, path); err != nil {
        _ = os.Remove(tmp)
        return err  // cross-device rename, target locked (Windows), permission denied
    }
    return nil
}
```

Common real-world failures:
- **Disk full** — `os.WriteFile` fails with `ENOSPC`
- **Read-only filesystem** — `os.WriteFile` and `os.Rename` both fail
- **Windows file lock** — `os.Rename` fails if target is locked by another process
- **Parent directory doesn't exist** — `os.WriteFile` fails (though `MkdirAll` before config write handles this for the config path itself, the session file path uses the same dir)

#### How MergeConfig is called from Adapter.Install()

From `adapters/codex/adapter.go:140-148`:
```go
if err := MergeConfig(configPath(base), sequoiaTable); err != nil {
    _ = cmdInstaller.Rollback()
    _ = skillInstaller.Rollback()
    return fmt.Errorf("install: merge config: %w", err)
}
```

- If `MergeConfig` returns an error, the entire Install rolls back (skills + commands are restored)
- If `MergeConfig` "succeeds" (backup written, session silently failed), Install continues and writes the version file — leaving a broken state where uninstall can't properly restore
- `MergeConfig` is a package-level function, not a method on `Adapter`, so it has no access to `BaseAdapter.AddWarning()`

#### AddWarning method

- Exists on `BaseAdapter` (line 162-167): thread-safe, appends to `a.warnings` slice
- `Warnings()` returns a copy (line 170-174)
- Warnings are cleared at the start of `Install`/`Uninstall` via `clearWarnings()` (line 307)
- Not directly usable by `MergeConfig` without changing its signature

#### Existing tests for MergeConfig

| Test | What it covers |
|------|---------------|
| `TestMergeConfig_FreshFile` | No existing file |
| `TestMergeConfig_ExistingSequoia_Overwrites` | Overwrite previous [sequoia] section |
| `TestMergeConfig_PreservesOtherContent` | Non-sequoia sections preserved |
| `TestMergeConfig_CreatesBackup` | Backup file created with timestamp |
| `TestMergeConfig_BackupHasUniqueName` | Two calls → two distinct backups |
| `TestMergeConfig_ExistingBackupNotOverwritten` | Old backup untouched |
| `TestMergeConfig_BackupPermissions_Restricted` | Backup uses 0o600 (Unix only) |

**Missing tests (gaps)**:
- No test for MergeConfig when session file write fails
- No test for RemoveConfig when session file is missing but timestamped backup exists
- No test for RemoveConfig with multiple timestamped backups and no session file
- No test verifying the error propagation chain from MergeConfig → Install → user

### Affected Areas

- `adapters/codex/installer.go:36-38` — primary bug site: silent error discard in `MergeConfig`
- `adapters/codex/installer.go:60-102` — `RemoveConfig` fallback chain that gets wrong behavior
- `adapters/codex/adapter.go:140-148` — `Adapter.Install()` caller of `MergeConfig`
- `adapters/common/strategy.go:131-135` — identical bug in `ReplaceFile` (same pattern, same consequences for `RestoreOrRemoveFile`)
- `adapters/codex/installer_test.go` — no test for session write failure
- `adapters/common/strategy_test.go` — no test for `ReplaceFile` session write failure

### Approaches

#### 1. Return the error from MergeConfig (RECOMMENDED)

Remove the silent discard and return the error. The session file is NOT optional — it is the mechanism that guarantees correct backup restoration.

- **Pros**:
  - Simplest fix: delete 2 lines, add 1 line (`return fmt.Errorf(...)`)
  - Install fails cleanly when session can't be written — no broken state
  - Adapter.Install() already handles MergeConfig errors with rollback (lines 144-148)
  - Admin gets a clear error message: `"install: merge config: write session: <os error>"`
  - No signature changes, no API breakage
  - Matches the existing pattern: backup write error at line 32 IS returned
- **Cons**:
  - A transient filesystem issue now fails the entire install (acceptable — partial installs are worse)
  - Need to add error-wrapping to match existing style
- **Effort**: Low (~3 lines changed)

#### 2. AddWarning via function parameter

Add a `warnings *[]string` or callback parameter to `MergeConfig`.

- **Pros**: Non-fatal, install still succeeds
- **Cons**:
  - Changes public API of `MergeConfig` — breaks existing callers
  - Install "succeeds" but produces a broken uninstall state
  - Warnings are not surfaced to the CLI user in the current code unless explicitly checked
  - More code, higher complexity
- **Effort**: Medium

#### 3. Convert MergeConfig to an Adapter method

Make `MergeConfig` a method on `Adapter` so it can use `a.AddWarning()`.

- **Pros**: Clean API, uses existing warning infrastructure
- **Cons**:
  - Large refactor — changes how `MergeConfig` is called everywhere
  - `MergeConfig` is tested as a standalone function — all tests need rewriting
  - Still produces a broken uninstall state (warnings are non-fatal)
- **Effort**: High

### Recommendation

**Approach 1 — return the error.** The session file is the linchpin of the backup/restore mechanism. If it can't be written, the system cannot guarantee correct uninstall. Failing the install cleanly with rollback is the correct behavior. The existing rollback path in `Adapter.Install()` already handles this — no new rollback logic needed.

The same fix should be applied to the identical pattern in `ReplaceFile` (`adapters/common/strategy.go:131-135`), and both should be part of this change since they share the same root cause and consequences.

### Risks

- **Transient disk-full scenarios**: A disk that's temporarily full during session write but has space for the backup file could cause install failures that didn't happen before. Mitigation: the error message is clear, and the admin can retry.
- **ReplaceFile consumers**: All adapters using `StrategyFileReplace` (Claude, Gemini, OpenCode, Cursor) use `ReplaceFile` → `RestoreOrRemoveFile`. Changing ReplaceFile's session behavior could affect all of them. Mitigation: the same logic applies — if session can't be written, the install should fail for ALL adapters. The `ReplaceFile` callers are in `BaseAdapter.Install()` and `BaseAdapter.Uninstall()`, which already have proper rollback and error collection.

### Ready for Proposal

**Yes.** The fix is clear, the scope is narrow, and the root cause is well-understood. Ready to proceed to sdd-propose for P4-005.
