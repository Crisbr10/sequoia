# Tasks: Namespace Backup Directories

*Change*: `namespace-backup-dirs` — isolate installer backups by adapter ID and installer type, preventing cross-contamination during Rollback.

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~55 (3 source files + 1 test file) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Delivery strategy | ask-on-risk |
| Suggested split | Single PR |

Decision needed before apply: Yes
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Full backup isolation fix + test | PR 1 | Single PR; all 4 files; well under 400 lines |

## Phase 1: RED — Write Failing Isolation Test

- [x] 1.1 Add `TestInstaller_BackupIsolation_SkillSurvivesCommandFailure` to `adapters/common/installer_test.go`. Create two installers with subdirectory BackupDirs (`skills/`, `commands/` under same parent). Run skill installer fully. Simulate command installer Apply failure (missing source file). Verify skill subdirectory intact and backup restores original. Use `t.TempDir()`, `require.*` for setup, `assert.*` for assertions. Follow table-driven pattern but single case is fine.

## Phase 2: GREEN — Implement Backup Namespacing

- [x] 2.1 **`adapters/common/base_adapter.go:358`** — Insert `a.ID()` into backup path: change `a.backupPathFn(base) + "-" + sessionSuffix` to `a.backupPathFn(base) + "-" + a.ID() + "-" + sessionSuffix`.

- [x] 2.2 **`adapters/common/base_adapter.go:372`** — Change skill installer `BackupDir` from `backupDir` to `filepath.Join(backupDir, "skills")`.

- [x] 2.3 **`adapters/common/base_adapter.go:389`** — Change command installer `BackupDir` from `backupDir` to `filepath.Join(backupDir, "commands")`.

- [x] 2.4 **`adapters/codex/adapter.go:121`** — Apply same pattern: construct `baseBackup := backupPath(base) + "-" + a.ID() + "-" + sessionSuffix`, then set `BackupDir: filepath.Join(baseBackup, "skills")`.

- [x] 2.5 **`adapters/codex/adapter.go:131`** — Set command installer `BackupDir: filepath.Join(baseBackup, "commands")` (reuse `baseBackup` from task 2.4).

- [x] 2.6 **`adapters/_template/adapter.go`** — Add `"strconv"` and `"time"` imports. Generate `sessionSuffix` via `strconv.FormatInt(time.Now().UnixMilli(), 36)`. Build `baseBackup := backupPath(base) + "-" + a.ID() + "-" + sessionSuffix`. Set skill installer `BackupDir: filepath.Join(baseBackup, "skills")` and command installer `BackupDir: filepath.Join(baseBackup, "commands")`.

## Phase 3: REFACTOR/VERIFY — Confirm Isolation

- [x] 3.1 Run `go test ./adapters/common/ -run TestInstaller_BackupIsolation -v`. Confirm new test passes and all existing installer tests still pass.

- [x] 3.2 Run full adapter suite: `go test ./adapters/... -count=1`. Verify no regressions across claude, codex, gemini, opencode, cursor, template.

- [x] 3.3 Run `go vet ./...` to catch any unused imports or shadowed variables.
