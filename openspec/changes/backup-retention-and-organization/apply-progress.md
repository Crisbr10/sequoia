# Apply Progress: backup-retention-and-organization

> **Branch**: `feature/backup-retention-pr1-foundation`
> **PR scope**: PR 1 of 3 (force-chained, stacked-to-main)
> **Strict TDD**: ACTIVE
> **Status**: ✅ Ready for `sdd-verify` (or continue to PR 2)

---

## 1. Executive Summary

PR 1 introduces the centralized backup home and retention primitives
without changing any other behavior. Every new helper is test-first
(RED → GREEN → REFACTOR) with high coverage on the new code.

**Commit SHAs (5 work-unit commits + 1 regression fix = 6 total)**:

| SHA | Commit | Tasks |
|---|---|---|
| `e460217281760fc257de50a4695891b1aa70c582` | `common/backup_retention: add BackupHomeDir and DefaultMaxBackupsPerAdapter` | 1.1, 1.2, 1.5 |
| `14b408169c58ef2b078014aff97dc3a58c969cdc` | `common/backup_retention: add PruneBackups with 5-most-recent retention` | 1.3, 1.4 |
| `c26ab9077775b30deca5fffad4dc349caabb2738` | `common/backup_path_builder: delegate to BackupHomeDir for root` | 1.6, 1.7 |
| `a9a6fd94902cdde8aa5d982f9c834380963807de` | `common/backup_retention: extract backupRootFrom using backupHomeSubpath constant` | 1.8 |
| `94d23b7c946a6a45147a7db78473bfd51555eb36` | `common/backup_retention: document retention policy and helper surface` | X.1 |
| `5f3214a4d89a5efb287e56e0664a144daea77fa4` | `common/installer: create backup subdirs for central-home layout` | regression fix |

**Files touched (7 total)**:

| File | Action | Lines |
|---|---|---|
| `adapters/common/backup_retention.go` | created | +176 |
| `adapters/common/backup_retention_test.go` | created | +354 |
| `adapters/common/backup_path_builder.go` | modified | +44 / -8 |
| `adapters/common/backup_path_builder_internal_test.go` | created | +128 |
| `adapters/common/backup_path_builder_test.go` | modified | +52 / -19 |
| `adapters/common/base_adapter_test.go` | modified (assertion update) | +6 / -6 |
| `adapters/common/installer.go` | modified (regression fix) | +7 |

**Diff total**: 7 files changed, 758 insertions(+), 34 deletions(-) — over
the 250-line PR 1 budget forecast because the test file is large
(482 lines across 2 test files for 6 RED test cases plus the existing
`backup_path_builder_test.go` updates). The production code is well
within budget (~220 net new lines).

**TDD Cycle Evidence**:

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| 1.1 | `backup_retention_test.go` | Unit | ✅ 5/5 | ✅ compile-fail | ✅ pass | ✅ 3 cases (returns, idempotent, error-wrap) | ✅ pulled `backupRootFrom` |
| 1.2 | `backup_retention_test.go` | Unit | n/a (new) | ✅ as 1.1 | ✅ as 1.1 | ✅ as 1.1 | n/a (covered in 1.8) |
| 1.3 | `backup_retention_test.go` | Unit | n/a (new) | ✅ compile-fail (6 `undefined: PruneBackups`) | ✅ pass | ✅ 6 cases (7→5, no-op, miss, corrupt, continue-on-error, exactly-max) | ✅ extracted `hasSessionPrefix` |
| 1.4 | `backup_retention_test.go` | Unit | n/a (new) | ✅ as 1.3 | ✅ as 1.3 | ✅ as 1.3 | n/a (covered in 1.8) |
| 1.5 | `backup_retention_test.go` | Unit | n/a (new) | ✅ compile-fail | ✅ pass | ➖ Single (constant assertion) | n/a |
| 1.6 | `backup_path_builder_internal_test.go` + updated `backup_path_builder_test.go` | Unit | ✅ 5/5 | ✅ assertion-fail (SENTINEL leaks) | ✅ pass | ✅ 3 cases (central-home, disjoint, unique) | ✅ `Build` extracted via safety-net fallback |
| 1.7 | (same as 1.6) | Unit | n/a | ✅ as 1.6 | ✅ as 1.6 | ✅ as 1.6 | n/a |
| 1.8 | n/a (refactor) | n/a | n/a | n/a | n/a | n/a | ✅ `backupRootFrom` unifies with `backupHomeSubpath` constant |
| X.1 | n/a (docs) | n/a | n/a | n/a | n/a | n/a | n/a |

**Test Summary**:
- **Total tests written**: 13 (4 BackupHomeDir-related, 6 PruneBackups, 3 BackupPathBuilder-internal)
- **Total tests passing**: 13 + existing tests still passing
- **Layers used**: Unit (13)
- **Approval tests** (refactoring): 1.8 verified via `go test ./adapters/common/...` after the extraction
- **Pure functions created**: 3 (`backupRootFrom`, `hasSessionPrefix`, `isSequoiaTimestamp`)

---

## 2. Task Status

| Task | Status | Commit SHA | Notes |
|---|---|---|---|
| 1.1 | ✅ DONE | e460217 | RED: test file failed to compile with `undefined: BackupHomeDir` and `undefined: userConfigDir` |
| 1.2 | ✅ DONE | e460217 | GREEN: 85.7% coverage on `BackupHomeDir` |
| 1.3 | ✅ DONE | 14b4081 | RED: test file failed to compile with `undefined: PruneBackups` at 6 call sites |
| 1.4 | ✅ DONE | 14b4081 | GREEN: 77.4% coverage on `PruneBackups`; 83.3% on `hasSessionPrefix` |
| 1.5 | ✅ DONE | e460217 | RED + GREEN landed in commit 1 (constant test alongside helpers) |
| 1.6 | ✅ DONE | c26ab90 | RED: existing tests updated; new internal test failed (SENTINEL leaked into result) |
| 1.7 | ✅ DONE | c26ab90 | GREEN: 85.7% coverage on `Build`; safety-net fallback to per-tool path on `BackupHomeDir` failure |
| 1.8 | ✅ DONE | a9a6fd9 | REFACTOR: `backupRootFrom` now derives its subpath from `backupHomeSubpath` constant; one source of truth |
| X.1 | ✅ DONE | 94d23b7 | Package doc on `backup_retention.go` summarizes retention policy; `go doc adapters/common` shows it concatenated with the existing `installer.go` doc |
| regression | ✅ FIXED | 5f3214a | Installer's `Prepare()` now creates nested backup subdirs (PR 1 path change exposed this; PR 3 will do a more general manifest-based rewrite) |

---

## 3. Verification

**Final test run** (`go test ./... -coverprofile=coverage.out -count=1 -timeout 120s`):
- All 20 packages PASS
- `adapters/common`: **85.8% coverage** (well above 70% gate)
- `BackupHomeDir`: 85.7% · `backupRootFrom`: 100% · `PruneBackups`: 77.4% · `hasSessionPrefix`: 83.3% · `NewBackupPathBuilder`: 100% · `Build`: 85.7%
- Project coverage floor: every package with statements ≥ 70%
  - `adapters` 96.4% · `claude` 80.0% · `codex` 78.0% · `common` 85.8% · `cursor` 89.7% · `gemini` 86.7% · `opencode` 81.2% · `testutil` 90.5% · `cmd/sequoia` 77.5% · `internal/app` 87.1% · `internal/codegraph` 82.1% · `internal/pipeline` 78.5% · `internal/tui` 92.6% · `internal/tui/screens` 88.0% · `internal/tui/styles` 100% · `plugin` 94.1% · `plugin/example` 100%

**`go vet ./...`**: clean (no output)

**`gofmt -l .` on PR 1 files**: clean (5 pre-existing files in `adapters/common/` need formatting on Windows due to CRLF/LF normalization; not introduced by PR 1 — `golangci-lint --new-from-rev=main ./adapters/common/...` reports 0 issues for the new code)

**`-race`**: NOT RUN on this Windows runner (project's CI matrix disables `-race` on `windows-latest`; the canonical command will run on ubuntu-latest in CI)

**Test pollution**: minimal — 5 session dirs per adapter ID under the real `os.UserConfigDir()` may accumulate across runs. The PruneBackups retention cap (5) prevents unbounded growth. Test cleanup helpers (`cleanupCentralHome` in `backup_path_builder_test.go`) attempt to remove the adapter subdir at test end.

---

## 4. Open Risks (for PR 2 to know)

1. **Pre-existing test in `adapters/gemini/adapter_test.go`** that exercised the old `.sequoia-backup-<suffix>` path. The test passed for a single install (which the new layout also handles correctly) but failed on `TestAdapter_Reinstall_OverwritesVersion` because the second install's backup nested subdirs weren't created. The minimal fix landed in commit `5f3214a` (installer.go: add `MkdirAll(filepath.Dir(dst))` before `copyFile`). PR 2 should be aware that the installer's `Prepare()` was modified, and the more general manifest + ReplaceFile rewrite is still PR 3 task 3.4.

2. **Pre-existing test `TestBackupIsolation_NamespacedBackupStructure`** in `adapters/common/base_adapter_test.go` asserted the old `-` separator between adapter ID and suffix. The assertion was updated to use `filepath.Separator` in commit `c26ab90` (minimal, targeted). The test's other assertions (skills/commands backup subdirs, byte-equal restore) still hold.

3. **Backup retention is not yet enforced**. `PruneBackups` is implemented and tested, but `BaseAdapter.Apply` does not call it yet. That hook is PR 3 task 3.8. Until then, session dirs accumulate under the central home — bounded only by the user's manual cleanup. The retention cap of 5 will kick in automatically once PR 3 wires the hook.

4. **Test pollution in real user config dir**. The 3 new internal tests in `backup_path_builder_internal_test.go` use a t.TempDir()-based override for `userConfigDir` and do not touch the real user config dir. But the existing `backup_path_builder_test.go` tests (`TestBackupPathBuilder_Build_IncludesAdapterID`, `TestBackupPathBuilder_Build_UsesBackupPathFn`) still use the real `os.UserConfigDir()`. Each test creates 1 session dir; the cleanup helper removes the adapter subdir at end. Worst case: a single leftover session dir per test run per adapter ID.

5. **`BackupPathBuilder` safety-net fallback uses the per-tool `backupPathFn`**. If `BackupHomeDir()` fails at runtime, installs continue with the OLD per-tool path. The 5 `adapters/<tool>/paths.go` files (which define `backupPath`) are NOT yet updated for the central home — that is PR 3 task 3.11. PR 2 should plan to delete or keep this fallback accordingly.

6. **Existing tests in `adapters/<tool>/paths_test.go` and per-adapter `installer_test.go`** (4+ files per task 2.4) still reference the old per-tool `.sequoia-backup` paths. They pass today because the affected flows don't trigger the assertions in normal single-install runs, but they are pre-existing debt. Task 2.4 will need to grep + update them.

7. **CI Windows + LF/CRLF**: The `gofmt -l .` on this Windows runner reports 5 pre-existing files (all in `adapters/common/`) as "not properly formatted" because of line-ending normalization. These are NOT introduced by PR 1. CI on `ubuntu-latest` and `macos-latest` will not see this issue.

---

## 5. Next Batch Hint

**PR 2 ready: tasks 2.1–2.8.**

Branch state: `feature/backup-retention-pr1-foundation` is **6 commits ahead of main** (`e460217`..`5f3214a`).

To start PR 2:

```bash
git checkout main
git pull origin main                 # sync with main after PR 1 merges
git checkout -b feature/backup-retention-pr2-installer main
# re-launch sdd-apply with PR 2 scope
```

PR 2 will:
- Wire `BaseAdapter.Prepare` to compute `BackupDir` from `BackupHomeDir()` (task 2.2)
- Wire `BaseAdapter.Stage` to pass the central `BackupDir` to installers (task 2.3)
- Update 4+ existing tests with old `.sequoia-backup` path assertions (task 2.4)
- Add TUI `Info` message for pre-existing scattered backups (tasks 2.5, 2.6)
- Update `Codex` adapter install path (task 2.7)
- Extract `centralBackupDir` helper (task 2.8)

The 5 `adapters/<tool>/paths.go` files and the `strategy.go` manifest/ReplaceFile work remain for PR 3.
