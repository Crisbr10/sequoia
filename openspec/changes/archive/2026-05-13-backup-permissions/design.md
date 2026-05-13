# Design: Restrict Backup File Permissions

## Technical Approach

Four (4) one-line permission constant changes across 3 files, plus one `os.Chmod` call. No new files, no new functions, no API changes. Production file permissions are untouched.

## Changes

### 1. Backup directory permission (`adapters/common/installer.go:84`)

**Before**: `os.MkdirAll(cfg.BackupDir, 0o755)`
**After**: `os.MkdirAll(cfg.BackupDir, 0o700)`

### 2. Backup file permission after copy (`adapters/common/installer.go`, after line 89)

**Before**: `copyFile(src, dst)` — file inherits umask-based permissions (typically 0o644)
**After**: 

```go
if err := copyFile(src, dst); err != nil {
    return fmt.Errorf("prepare: backup %q: %w", name, err)
}
if err := os.Chmod(dst, 0o600); err != nil {
    return fmt.Errorf("prepare: chmod backup %q: %w", name, err)
}
```

Reason: `os.Create` uses process umask, which on most Unix systems results in 0o644. We must explicitly `Chmod` to 0o600 after the copy completes.

### 3. ReplaceFile backup permission (`adapters/common/strategy.go:127`)

**Before**: `os.WriteFile(backup, raw, 0o644)`
**After**: `os.WriteFile(backup, raw, 0o600)`

### 4. Codex MergeConfig backup permission (`adapters/codex/installer.go:30`)

**Before**: `os.WriteFile(backupPath, []byte(existing), 0o644)`
**After**: `os.WriteFile(backupPath, []byte(existing), 0o600)`

## Files Modified

| File | Lines | Change |
|------|-------|--------|
| `adapters/common/installer.go` | ~5 | B1 (MkdirAll 1 line), B2 (Chmod after copyFile ~4 lines) |
| `adapters/common/strategy.go` | 1 | B3 (WriteFile mode) |
| `adapters/codex/installer.go` | 1 | B4 (WriteFile mode) |
| **Total production** | **~7** | |
| **Total with tests** | **~93** | |

## Test Implications

Current tests do not assert backup file/directory permissions explicitly (they verify content and existence only). On Unix, `os.Stat(path).Mode().Perm()` on a backup file would now return `0o600` instead of `0o644`, but no test currently makes this assertion. No test changes are strictly required — however, the task checklist includes running the full test suite (`go test -race -count=1 ./...`) to confirm no regressions.

## Platform Safety

- **Windows**: Unix permission bits (`0o600`, `0o700`) are no-ops on Windows. `os.Chmod` only affects the write bit. No regression risk.
- **Unix/Linux/macOS**: The fix correctly restricts backup files to owner-only access.

## What Stays the Same

- All production file writes (skills, commands, version markers, configs) remain at `0o644`/`0o755`
- Restore logic (Rollback, RestoreOrRemoveFile, RemoveConfig) — already writes to production paths with appropriate permissions
- No API surface changes — `InstallerConfig`, `Installer`, `ReplaceFile`, `MergeConfig` signatures unchanged
