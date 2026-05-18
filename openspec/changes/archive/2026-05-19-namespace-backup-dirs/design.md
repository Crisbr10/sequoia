# Design: Namespace Backup Directories

## Technical Approach

Inject `{adapterID}` into the backup path suffix and split backup dirs by installer type (`skills/`, `commands/`). The Installer lifecycle is unchanged; only the `BackupDir` values passed to `InstallerConfig` change at the call sites in each adapter's `Install()` method.

## Architecture Decisions

### Decision: Three-segment backup path

| Option | Tradeoff | Decision |
|--------|----------|----------|
| `{base}-{adapterID}-{sessionSuffix}/{installerType}` | Longer paths, but full isolation | **Chosen** |
| Single flat dir with file prefixes | Shorter path, but shared dir still destroyed by os.RemoveAll | Rejected |
| UUID-only (no adapter ID) | Unique but loses adapter traceability | Rejected |

**Choice**: `{base}-{adapterID}-{sessionSuffix}/{installerType}`
**Rationale**: Adapter ID prevents cross-adapter collision; installer-type subdirectory prevents cross-installer rollback contamination; session suffix prevents temporal collisions.

### Decision: No changes to installer.go

**Choice**: Each installer receives a unique `BackupDir` subdirectory — installer's existing `os.RemoveAll(cfg.BackupDir)` in Rollback() targets only its own subdirectory.
**Alternatives**: Add a "namespace" field to InstallerConfig and guard os.RemoveAll against sibling paths.
**Rationale**: The root cause is shared BackupDir, not Rollback logic. Unique subdirectories isolate os.RemoveAll without touching the installer contract. Minimal diff, maximum safety.

## Data Flow

```
backupPathFn(base) → "{base}-sequoia-backup"
       +
    "-" + a.ID() + "-" + sessionSuffix
       │
       ├── filepath.Join(result, "skills")  → skillInstaller.BackupDir
       │
       └── filepath.Join(result, "commands") → cmdInstaller.BackupDir
```

cmdInstaller.Rollback() → `os.RemoveAll("{...}/commands")` → skills subdirectory untouched.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `adapters/common/base_adapter.go:358` | Modify | `backupDir := a.backupPathFn(base) + "-" + a.ID() + "-" + sessionSuffix` |
| `adapters/common/base_adapter.go:372` | Modify | `BackupDir: filepath.Join(backupDir, "skills"),` |
| `adapters/common/base_adapter.go:389` | Modify | `BackupDir: filepath.Join(backupDir, "commands"),` |
| `adapters/codex/adapter.go:121` | Modify | Same pattern: `a.ID()` + subdirs in BackupDir |
| `adapters/codex/adapter.go:131` | Modify | Same as above for commands installer |
| `adapters/_template/adapter.go:196-210` | Modify | Add session suffix generation + adapter ID + subdirs; add `strconv`/`time` imports |
| `adapters/common/installer_test.go` | Modify | Add `TestInstaller_BackupIsolation_SkillSurvivesCommandFailure` |

## Call-site contract

Every adapter's Install() MUST construct BackupDir as:

```go
sessionSuffix := strconv.FormatInt(time.Now().UnixMilli(), 36)
baseBackup := a.backupPathFn(base) + "-" + a.ID() + "-" + sessionSuffix

skillInstaller := NewInstaller(InstallerConfig{
    BackupDir: filepath.Join(baseBackup, "skills"),
    ...
})
cmdInstaller := NewInstaller(InstallerConfig{
    BackupDir: filepath.Join(baseBackup, "commands"),
    ...
})
```

Adapters inheriting BaseAdapter.Install() (claude, gemini, opencode, cursor) get this automatically. Codex and future custom adapters must follow the same pattern. The `_template` adapter is updated to serve as a reference.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Skill backup survives command rollback | `t.TempDir()` + two installers with different subdirectory BackupDirs; run skill installer, simulate cmd install failure, verify `skills/` subdir intact, restore from backup |
| Unit | Fresh install produces identical output | Compare installed file content with pre-fix golden files in `testdata/golden/` |
| Integration | Concurrent adapter installs | Two adapters with distinct IDs using `t.TempDir()`, verify disjoint backup trees |

New test: `TestInstaller_BackupIsolation_SkillSurvivesCommandFailure` in `installer_test.go`. Table-driven pattern matching existing tests. Uses `t.Run(tc.name, ...)`, `require.*` for setup, `assert.*` for verification.

## Migration / Rollout

No migration required. Backup dirs are ephemeral (created per install, cleaned by rollback). Old backup dirs without subdirectories are not touched — they remain on disk until manual cleanup.

## Open Questions

None.
