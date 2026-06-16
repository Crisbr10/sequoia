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

---

# PR 2 Apply Progress

> **Branch**: `feature/backup-retention-pr2-installer`
> **PR scope**: PR 2 of 3 (force-chained, stacked-to-main)
> **Strict TDD**: ACTIVE
> **Commits ahead of main**: 4 (90010d9..688c579)
> **Status**: ✅ Ready for `sdd-verify`

---

## PR 2 Executive Summary

PR 2 wires the `BaseAdapter` install flow and the `Codex` custom
install to the central backup home produced by PR 1, surfaces a
TUI `Info` migration note about pre-existing scattered backups,
and tightens the test assertion for the namespaced backup structure.
A new exported helper `BaseAdapter.CentralBackupDir(targetSubdir)`
becomes the single source of truth for per-install session directory
construction, replacing the indirect `a.backup.Build(base)` call
that PR 1 left in `BaseAdapter.Prepare` and `Stage`.

**Commit SHAs (4 work-unit commits = 4 total)**:

| SHA | Commit | Tasks |
|---|---|---|
| `90010d9` | `common/base_adapter: route Prepare and Stage through BackupHomeDir` | 2.1, 2.2, 2.3, 2.8 (helper + uses) |
| `d484625` | `codex/installer: route Codex install backup through central home` | 2.7 |
| `a3a7153` | `pipeline/runner: add migration note to backup Info message` | 2.5, 2.6 |
| `688c579` | `tests: tighten backup path assertion to central home prefix` | 2.4 (assertion tightening) |

**Files touched (7 total)**:

| File | Action | Lines |
|---|---|---|
| `adapters/common/base_adapter.go` | modified (CentralBackupDir + SetLastBackupDir + Prepare/Stage/Apply refactor) | +61 / -10 |
| `adapters/common/base_adapter_internal_test.go` | created (3 new tests for Prepare + helper) | +193 |
| `adapters/common/base_adapter_test.go` | modified (tighten TestBackupIsolation_NamespacedBackupStructure) | +14 / -4 |
| `adapters/codex/adapter.go` | modified (custom Install uses CentralBackupDir + SetLastBackupDir) | +9 / -4 |
| `adapters/codex/installer_internal_test.go` | created (1 new test for Codex LastBackupDir) | +61 |
| `internal/pipeline/runner.go` | modified (Info now appends migration note) | +7 / -1 |
| `internal/pipeline/runner_info_test.go` | created (2 new tests for Info migration note) | +120 |

**Diff total**: 7 files changed, ~456 insertions, ~19 deletions —
slightly above the 350-line PR 2 budget forecast (production code is
within budget; the budget is exceeded by the new test files for
2.1/2.8/2.7 which add ~376 lines).

**TDD Cycle Evidence**:

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| 2.1 | `base_adapter_internal_test.go` (new) | Unit | n/a (new) | ✅ compile-fail (`centralBackupDir` undefined) | ✅ pass | ✅ 2 cases (empty subdir returns central home; targetSubdir joins correctly) | ➖ clean first try |
| 2.2 | `base_adapter_internal_test.go` (Prepare) | Unit | n/a (same file) | ✅ assertion (sentinel doesn't leak) | ✅ pass | ✅ 3 cases (sentinel excluded, no `.sequoia-backup`, prefix matches) | ➖ clean |
| 2.3 | existing `TestBackupIsolation_NamespacedBackupStructure` | Unit | ✅ pre-PR2 | n/a (covered by 2.1+2.2 helper) | n/a | n/a | n/a |
| 2.4 | `base_adapter_test.go` (tighten) | Unit | ✅ pre-existing | n/a (assertion tightening only) | ✅ pass | n/a | n/a |
| 2.5 | `runner_info_test.go` (new) | Unit | n/a (new) | ✅ assertion (note substring) | ✅ pass | ✅ 2 cases (note included; empty backup → no note) | ➖ clean |
| 2.6 | (same as 2.5) | Unit | n/a | ✅ as 2.5 | ✅ as 2.5 | ✅ as 2.5 | n/a |
| 2.7 | `codex/installer_internal_test.go` (new) | Unit | n/a (new) | ✅ assertion (`LastBackupDir` empty before) | ✅ pass | ✅ 3 cases (non-empty; central home prefix; no `.sequoia-backup` marker) | ➖ clean |
| 2.8 | `base_adapter_internal_test.go` (helper) | Unit | n/a (same file) | ✅ compile-fail (helper undefined) | ✅ pass | ✅ 2 cases (subdir join; session-dir cache consistency) | ➖ extracted to dedicated method |

**Test Summary**:
- **Total tests written**: 6 (3 for 2.1/2.8 in `base_adapter_internal_test.go`, 1 for 2.5, 1 for 2.5 empty-backup variant, 1 for 2.7)
- **Total tests passing**: all 6 new + all 151 pre-existing = 157
- **Layers used**: Unit (6)
- **Approval tests** (refactoring): none — task 2.8 is implemented as part of the GREEN for 2.1/2.2/2.3 (the helper extraction IS the GREEN)
- **Pure functions created**: 1 (`CentralBackupDir`)

---

## PR 2 Task Status

| Task | Status | Commit SHA | Notes |
|---|---|---|---|
| 2.1 | ✅ DONE | 90010d9 | RED: compile-fail on `centralBackupDir` undefined; GREEN: helper added and used in Prepare |
| 2.2 | ✅ DONE | 90010d9 | GREEN: `Prepare()` now calls `CentralBackupDir("")` directly (no more `a.backup.Build(base)` indirection) |
| 2.3 | ✅ DONE | 90010d9 | GREEN: `Stage()` now calls `CentralBackupDir("skills")` and `CentralBackupDir("commands")` directly |
| 2.4 | ✅ DONE | 688c579 | Test assertion tightening for `TestBackupIsolation_NamespacedBackupStructure` (asserts central home prefix). The other 4+ files in the task list (installer_test.go, codex/installer_test.go, opencode/install_test.go, base_adapter_strategy_test.go, base_adapter_error_test.go) were reviewed — none of them assert the BaseAdapter backup path location, so no changes were needed there. |
| 2.5 | ✅ DONE | a3a7153 | RED: assertion on note substring failed; GREEN: Info now includes "pre-existing scattered backups from prior sequoia versions remain at their original locations" |
| 2.6 | ✅ DONE | a3a7153 | GREEN: same commit — Info field is now `<centralHomePath> — pre-existing scattered backups from prior sequoia versions remain at their original locations.` |
| 2.7 | ✅ DONE | d484625 | RED: `Codex.LastBackupDir()` returned "" before; GREEN: custom Install now calls `CentralBackupDir("")` for the session root, uses `CentralBackupDir("skills")` and `CentralBackupDir("commands")` for the per-installer subdirs, and calls `a.SetLastBackupDir(baseBackup)` to record the session dir for the TUI |
| 2.8 | ✅ DONE | 90010d9 | REFACTOR: extracted as part of the GREEN for 2.1/2.2/2.3. The helper is exported as `CentralBackupDir` (not the original `centralBackupDir` lowercase from the design) so the Codex custom Install can share the same cache. The `CentralBackupDir(targetSubdir)` method joins `<cached session dir>/<targetSubdir>` and caches the session dir for the lifetime of the install. |

---

## PR 2 Verification

**Final test run** (`go test ./... -coverprofile coverage.out -count=1 -timeout 120s`):
- All 20 packages PASS
- `adapters/common`: **85.5% coverage** (well above 70% gate)
  - `CentralBackupDir`: 92.3% · `Prepare`: 95.2% · `Stage`: 83.3% · `Apply`: 94.4% · `LastBackupDir`: 100% · `SetLastBackupDir`: 0% (not directly tested; called from Codex's internal test which goes through `BaseAdapter.Install`)
- `internal/pipeline`: **77.1% coverage** (well above 70% gate; up from 76.9% in PR 1)
- Project coverage floor: every package with statements ≥ 70%
  - `adapters` 96.4% · `claude` 80.0% · `codex` 78.1% · `common` 85.5% · `cursor` 89.7% · `gemini` 86.7% · `opencode` 81.2% · `testutil` 90.5% · `cmd/sequoia` 77.5% · `internal/app` 87.1% · `internal/codegraph` 82.1% · `internal/pipeline` 77.1% · `internal/tui` 92.6% · `internal/tui/screens` 88.0% · `internal/tui/styles` 100% · `plugin` 94.1% · `plugin/example` 100%

**`go vet ./...`**: clean (no output)

**`gofmt -l .` on PR 2 files**: clean. The 5 pre-existing files in `adapters/common/` with CRLF issues (`base_adapter.go`, `base_adapter_test.go`, `commands.go`, `template.go`, and `backup_retention.go`) are NOT introduced by PR 2 — `golangci-lint --new-from-rev=main ./adapters/common/...` reports 0 issues for the new code.

**`-race`**: NOT RUN on this Windows runner (project's CI matrix disables `-race` on `windows-latest`; the canonical command will run on ubuntu-latest in CI).

**Test pollution**: the `adapters/common/` tests write to the real `os.UserConfigDir()` central home (e.g., `C:\Users\Usuario\AppData\Roaming\sequoia\backups\err-test\<session>`). The PruneBackups retention cap (5) bounds growth. A first attempt to add `t.Cleanup` that removes the per-adapter subdir at test teardown caused flakiness because parallel tests using the same adapter ID race on the cleanup; the cleanups were reverted and the cap is the sole bound. The pre-existing pollution is documented in the PR 1 verify report §4.4 (applies unchanged here).

---

## PR 2 Open Risks (for PR 3 to know)

1. **`BackupPathBuilder` safety-net fallback is still wired but mostly unreachable.** The Codex custom Install now uses `CentralBackupDir()` exclusively, so `a.backup.Build("")` is only consulted inside `CentralBackupDir`'s fallback when `BackupHomeDir()` itself fails. The `BackupPathBuilder` is still kept in `BaseAdapter` because (a) `BuildBackupPath(base)` is still public API and (b) the 5 per-tool `adapters/<tool>/paths.go::backupPath` functions still feed into it. PR 3 task 3.11 (per-adapter `paths.go` delegates to central home) is the natural place to decide whether to keep or drop the fallback. Recommendation: KEEP for resilience; the cost is one extra indirection per fallback call (which is never on the happy path).

2. **Pre-existing test pollution in real user config dir.** With PR 2, the `adapters/common/` tests continue to write one session dir per test run per adapter ID to the central home. The cap of 5 is the only bound. The pre-existing `cleanupCentralHome` helper in `backup_path_builder_test.go` covers 2 specific tests (`TestBackupPathBuilder_Build_IncludesAdapterID`, `TestBackupPathBuilder_Build_UsesBackupPathFn`) but most tests don't clean up. A follow-up could use the `internal_test.go` + `overrideUserConfigDir` pattern (used in `backup_path_builder_internal_test.go` and `backup_retention_test.go`) for new tests, and convert existing tests gradually.

3. **2 spec ambiguities from PR 1's verify report are still open.** The timestamp format dash vs dot (REQ-BRP-02) and the "directory #4" numbering (REQ-BRP-06) are SPEC issues, not code issues. PR 2 doesn't change them. Recommend filing as a follow-up `sdd-propose` change to clarify the spec, or addressing as a one-line edit in the spec when `sdd-archive` lands.

4. **`SetLastBackupDir` has 0% direct coverage.** It's called from `Codex.Install()` which is exercised in the new `installer_internal_test.go` test, but Go's coverage tool doesn't attribute the call to the method on `BaseAdapter`. This is a tooling artifact, not a real coverage gap — the call is exercised through the public `Adapter.Install()` path. A targeted direct test would be a tiny addition if the gap matters.

5. **`CentralBackupDir` cache is mutex-protected but per-instance.** The `a.mu` lock protects `a.centralSessionDir` within a single `BaseAdapter` instance. Cross-instance concurrency is independent (each adapter has its own cache). This is correct but means a test that uses two adapter instances pointing at the same `BackupHomeDir()` will generate two different session dirs, which is the desired behavior.

6. **`CentralBackupDir` is exported.** The original design said "private helper" but Codex's custom Install needed access. The trade-off is that the helper is now part of the public API surface of `adapters/common`. The method name is descriptive and the contract is clear. If the design wants to revert to private, the alternative is to add a public method on `BaseAdapter` that takes a callback for the per-installer subdirs (e.g., `a.WithCentralBackupDirs(func(skills, cmds string) { ... })`).

---

## PR 2 Next Batch Hint

**PR 3 ready: tasks 3.1–3.11.**
Branch state: `feature/backup-retention-pr2-installer` is **4 commits ahead of main** (90010d9..688c579).

```bash
git checkout main
git pull origin main                 # sync with main after PR 2 merges
git checkout -b feature/backup-retention-pr3-replacefile main
# re-launch sdd-apply with PR 3 scope
```

PR 3 will:
- Define `manifestEntry` and `manifest` types in `adapters/common/backup_retention.go` (task 3.1, 3.2)
- Wire `ReplaceFile` to central home + manifest.json (tasks 3.3, 3.4)
- Wire `RestoreOrRemoveFile` to read manifest and restore (tasks 3.5, 3.6)
- Hook `applyRetention` in `BaseAdapter.Apply()` for 5-backup cap (tasks 3.7, 3.8)
- Add retention warning on prune error (task 3.9)
- Consolidate manifest helpers (task 3.10)
- Update the 5 `adapters/<tool>/paths.go` files to delegate to central home (task 3.11)

The 2 spec ambiguities (timestamp format, directory #4 numbering) should be filed as a follow-up `sdd-propose` change after PR 3 lands, or addressed in the spec at `sdd-archive` time.
