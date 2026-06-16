# Verify Report: backup-retention-and-organization — PR 2 (Installer Wiring)

> **Verifier**: `sdd-verify` sub-agent (PR 2 only, do NOT proceed to PR 3)
> **Branch under verification**: `feature/backup-retention-pr2-installer`
> **Commits ahead of main**: 4 (`90010d9`..`688c579`)
> **Strict TDD**: ACTIVE
> **Run timestamp**: 2026-06-16, Windows runner (PowerShell 7)
> **Final status**: **PASS_WITH_WARNINGS**

---

## 1. Executive Summary

PR 2 wires `BaseAdapter.Prepare` and `BaseAdapter.Stage` to the central
backup home produced by PR 1, threads the same path through `Codex`'s
custom `Install` so its `LastBackupDir` reaches the TUI, adds the
`Info` migration note for pre-existing scattered backups, and tightens
the `TestBackupIsolation_NamespacedBackupStructure` assertion to the
new central-home prefix. The new exported helper
`BaseAdapter.CentralBackupDir(targetSubdir)` is the single source of
truth for per-install session directory construction; it correctly
caches a single session dir for `skills` and `commands` callers.

All 8 in-scope tasks (2.1–2.8) PASS. Full project test run is green
(20/20 packages, all five consecutive runs), `go vet ./...` is clean,
and per-function coverage on the new surfaces is well above the 70%
CI gate (`CentralBackupDir` 92.3%, `Prepare` 95.2%, `Stage` 83.3%,
`Apply` 94.4%, `runInstallSteps` 94.3%, `sendProgress` 100%).

I independently re-grepped the four test files the apply agent claimed
"don't assert the BaseAdapter backup path location". All matches are
**legitimately** out-of-PR-2-scope assertions against the legacy
`ReplaceFile`/`RestoreOrRemoveFile` sidecar (PR 3 territory,
REQ-BRP-03) or against the `BackupDir` *parameter* of
`common.InstallerConfig` (not the BaseAdapter's location). The apply
agent's claim is **verified** — no broken test assertions will land on
main from this deviation.

**Finding tally**:
- CRITICAL: 0
- WARNING: 1
- SUGGESTION: 1
- SPEC AMBIGUITY: 2 (carried over from PR 1, still open)
- OUT OF SCOPE (intentional non-failures): 5

**One-line verdict**: PR 2 is clean, focused, and well-tested. The
single warning is the same test-pollution risk that PR 1 carried, and
the retention cap that will bound it is PR 3's job. Proceed.

---

## 2. Task Verification

| Task | Claim | Result | Evidence |
|---|---|---|---|
| 2.1 RED | RED test for `BaseAdapter.Prepare` central home | **PASS** | `TestBaseAdapter_Prepare_BackupDirUsesCentralHome` in `adapters/common/base_adapter_internal_test.go:77-99` — compiles against `a.Prepare(...)` + `a.LastBackupDir()` and asserts `strings.HasPrefix(backupDir, centralHome+sep)`. Lives in commit `90010d9` alongside the GREEN impl; same TDD-commit-shape caveat as PR 1 (see SUGGESTION §9.1). |
| 2.2 GREEN | `BaseAdapter.Prepare` uses `CentralBackupDir` | **PASS** | `adapters/common/base_adapter.go:395`: `backupDir := a.CentralBackupDir("")` (replaces `a.backup.Build(base)` from PR 1). `a.backup.Build(...)` is no longer called in `Prepare`. The sentinel `backupPathFn` in the new test does NOT leak in. Coverage 95.2%. |
| 2.3 GREEN | `BaseAdapter.Stage` uses `CentralBackupDir` | **PASS** | `adapters/common/base_adapter.go:522,538`: skill and command installers both receive `BackupDir: a.CentralBackupDir("skills")` and `BackupDir: a.CentralBackupDir("commands")` respectively. Both share the same cached session dir (verified by `TestBaseAdapter_CentralBackupDir_CachesSessionDir`). Coverage 83.3%. |
| 2.4 GREEN | Test path assertions updated | **PASS** | **Verified independently** by re-grep of the 4+ files the apply agent claimed didn't assert the BaseAdapter path. See §5.2 for full evidence — all matches are either PR 3 scope (`codex.MergeConfig`/`RemoveConfig` sidecar, `ReplaceFile` strategy) or test-fixture strings (synthesized `BackupDir` *parameters*, not the BaseAdapter's location). Only `adapters/common/base_adapter_test.go:TestBackupIsolation_NamespacedBackupStructure` was actually updated (`/workspace/.../sequoia-ai/sequoia-ai` / `git diff -- adapters/common/base_adapter_test.go` shows the assertion added at lines 612-624: `assert.True(strings.HasPrefix(backupDir, centralHome+sep) ...)` and `assert.NotContains(backupDir, ".sequoia-backup" ...)`). |
| 2.5 RED | TUI `Info` migration note test | **PASS** | `internal/pipeline/runner_info_test.go:26-79`: `TestRunInstall_BackupInfo_MigrationNote` — installs via a stubbed `BackupDirGetter`-implementing adapter, asserts `last.Info` contains the backup path AND one of "pre-existing" / "scattered" / "prior sequoia". Pass. |
| 2.6 GREEN | Pipeline adds migration note | **PASS** | `internal/pipeline/runner.go:200-213`: when `adapter.(adapters.BackupDirGetter)` reports a non-empty `LastBackupDir()`, the emitted `ProgressMsg.Info` is set to `dir + " — pre-existing scattered backups from prior sequoia versions remain at their original locations."`. Coverage: `runInstallSteps` 94.3%, `sendProgress` 100%. |
| 2.7 RED | `Codex` install backup goes to central home | **PASS** | `adapters/codex/installer_internal_test.go:29-60`: `TestAdapter_Install_BackupDirUsesCentralHome` — runs `codex.NewAdapter(tmp).Install(...)` against `t.TempDir()` for the tool home, asserts `a.LastBackupDir()` starts with the real `os.UserConfigDir()/sequoia/backups` prefix and does NOT contain `.sequoia-backup` or `.codex`. Pass. |
| 2.8 REFACTOR | `CentralBackupDir` helper | **PASS** (with deviation — see §5.1) | Helper exists at `adapters/common/base_adapter.go:340-361`. Exported (uppercase `C`) rather than lowercase `centralBackupDir` as the design implied. Caches `a.centralSessionDir` (mutex-protected, cleared in `Prepare` at line 376). Lazily builds `<home>/<adapterID>/<ISO8601>-<suffix>`. Safety-net fallback to `a.backup.Build("")` on `BackupHomeDir()` failure. Coverage 92.3%. |

**Total**: 8/8 tasks PASS.

### 2.1 SUGGESTION (non-blocking) — TDD commit shape

Same observation as PR 1 verify §2.1: each RED+GREEN pair landed in
one commit (e.g., `90010d9` adds 190 lines of test + 76 lines of impl
together; the spec scenarios are present in the test file but the
strict RED→GREEN→REFACTOR cycle is not visible at the commit level).
The end-state is correct and the tests do exercise the behavior. Not
a blocker for PR 2; a future `sdd-apply` should land test-only commits
separately.

---

## 3. Spec Compliance

### REQ-BRP-01 — Centralized backup root — **PASS (carried from PR 1)**

PR 2 does not modify `BackupHomeDir()`. Verified by
`git diff --stat adapters/common/backup_retention.go` → empty. The
PR 1 helpers continue to satisfy this REQ.

### REQ-BRP-02 — Per-adapter organization — **PASS (PR 2 endpoint)**

The PR 2 wiring produces the spec's path shape end-to-end:

- `BaseAdapter.Prepare` → `CentralBackupDir("")` returns
  `<UserConfigDir>/sequoia/backups/<adapterID>/<ISO8601>-<suffix>`
  (`base_adapter.go:355`).
- `BaseAdapter.Stage` → skill + command backups land in
  `<sessionDir>/skills/` and `<sessionDir>/commands/`
  (`base_adapter.go:522,538`).
- `Codex.Install` → same structure
  (`codex/adapter.go:127,132,142`).
- Distinct adapter IDs produce disjoint subtrees — verified by the
  existing `TestBackupPathBuilder_Build_DisjointPathsForDifferentAdapters`
  test (PR 1, still passing).
- `skills/` and `commands/` substructure preserved — verified by
  `TestStrategy_FullSequenceSucceeds` which installs both subdirs
  and asserts their files (PASS).
- The just-tightered `TestBackupIsolation_NamespacedBackupStructure`
  (`base_adapter_test.go:582`) now explicitly asserts the central-home
  prefix, the no-`.sequoia-backup` invariant, AND that the per-tool
  `<home>/err-test` legacy layout is gone.

**Two spec ambiguities from PR 1 carry over (still open):**
1. The spec's example timestamp `2026-06-15T15-30-45-123Z-<suffix>`
   uses `-` between SS and mmm; the implementation uses
   `2006-01-02T15-04-05.000Z` (`.` between SS and mmm). Both are valid
   ISO-8601; the formatter/parser are internally consistent. Not a
   code issue. See SPEC AMBIGUITY §6.1.
2. REQ-BRP-06 "directory #4 is read-only" scenario is internally
   inconsistent with the spec's `removed=2` claim. Implementation
   matches the spec's *intent* (continues-on-error) with `removed=1`.
   Not a code issue. See SPEC AMBIGUITY §6.2.

### REQ-BRP-03 — File-replace backup storage — **OUT OF PR 2 SCOPE (PR 3)**

`ReplaceFile` and `RestoreOrRemoveFile` in `adapters/common/strategy.go`
are untouched. Verified by `git diff main..feature/backup-retention-pr2-installer
-- adapters/common/strategy.go` → empty. The function bodies still
use `.sequoia-backup-<suffix>` + `.sequoia-session` sidecar.

### REQ-BRP-04 — Retention policy of 5 per adapter — **OUT OF PR 2 SCOPE (PR 3)**

`applyRetention` is not yet hooked in `BaseAdapter.Apply`. Verified by
`git diff -- adapters/common/base_adapter.go | grep -E "applyRetention|PruneBackups"`
→ no matches. The helper (`PruneBackups`, `DefaultMaxBackupsPerAdapter=5`)
exists from PR 1 and is fully covered; the wiring is PR 3 task 3.8.

### REQ-BRP-05 — Migration of old scattered backups (NOT performed) — **PARTIAL PASS**

The two scenarios:

1. "Old scattered backups remain untouched after a new install" —
   **OUT OF PR 2 SCOPE** (no code path in PR 2 touches pre-existing
   `.sequoia-backup-*` files; `strategy.go` is untouched).
2. "TUI Info message notes pre-existing scattered backups" —
   **PASS**. The new `Info` string at `runner.go:208` reads:
   `<centralHomePath> — pre-existing scattered backups from prior sequoia versions remain at their original locations.`
   The test `TestRunInstall_BackupInfo_MigrationNote` verifies the
   path AND one of the keywords (pre-existing / scattered / prior
   sequoia) appears. The no-Info-when-empty test
   (`TestRunInstall_BackupInfo_NoMigrationNoteWhenEmpty`) also passes
   (empty `LastBackupDir()` → no Info emitted at all).

**Note**: PR 2 does **NOT** add the retention warning to the Info
message. The spec's REQ-BRP-05 second scenario is satisfied for the
migration note; the retention cap warning is PR 3 (REQ-BRP-04).

### REQ-BRP-06 — Path resolution and pruning helpers — **PASS (PR 1 scope)**

PR 2 does not touch the PR 1 helpers. `BackupHomeDir` and
`PruneBackups` continue to work as before. The PR 2 surfaces
(`BaseAdapter.Prepare`, `Stage`, `Codex.Install`) all use the
PR 1 `BackupHomeDir()` as the root and compose the session dir
identically to `BackupPathBuilder.Build`.

### REQ-BRP-07 — Test surface (strict TDD) — **PASS**

In-scope for PR 2:
- `BaseAdapter.Prepare` central home: ✅ `TestBaseAdapter_Prepare_BackupDirUsesCentralHome`
- `CentralBackupDir` helper: ✅ `TestBaseAdapter_CentralBackupDir_JoinsHomeAndSubdir` + `TestBaseAdapter_CentralBackupDir_CachesSessionDir`
- TUI `Info` migration note: ✅ `TestRunInstall_BackupInfo_MigrationNote` + `TestRunInstall_BackupInfo_NoMigrationNoteWhenEmpty`
- `Codex` custom Install central home: ✅ `TestAdapter_Install_BackupDirUsesCentralHome`
- `TestBackupIsolation_NamespacedBackupStructure` tightened to assert
  central home prefix and no `.sequoia-backup` marker: ✅ in
  `base_adapter_test.go:612-624` (commit `688c579`)

The 4+ test files the task mentioned that needed path updates were
**legitimately out of PR 2 scope** — they test the legacy
`ReplaceFile` sidecar (PR 3 territory). See §5.2 for the full
per-file analysis.

Out-of-scope (correctly deferred to PR 3):
- Centralized `ReplaceFile`/`RestoreOrRemoveFile` round-trip — **PR 3**
- E2E install leaves at most 5 session dirs — **PR 3 task 3.7**
- 5 per-adapter `paths.go` files — **PR 3 task 3.11**

---

## 4. Independent Re-execution (adversarial)

### 4.1 Full project test run (5 consecutive runs)

**Run #1** (with coverage):
```
ok  github.com/Crisbr10/sequoia/adapters          1.393s  coverage: 96.4%
ok  github.com/Crisbr10/sequoia/adapters/claude   3.017s  coverage: 80.0%
ok  github.com/Crisbr10/sequoia/adapters/codex    3.821s  coverage: 78.1%
ok  github.com/Crisbr10/sequoia/adapters/common   3.848s  coverage: 85.5%
?   github.com/Crisbr10/sequoia/adapters/common/installembed  [no test files]
ok  github.com/Crisbr10/sequoia/adapters/cursor   3.015s  coverage: 89.7%
ok  github.com/Crisbr10/sequoia/adapters/gemini   3.263s  coverage: 86.7%
ok  github.com/Crisbr10/sequoia/adapters/opencode 3.286s  coverage: 81.2%
ok  github.com/Crisbr10/sequoia/adapters/testutil 1.447s  coverage: 90.5%
ok  github.com/Crisbr10/sequoia/cmd/sequoia       5.822s  coverage: 77.5%
ok  github.com/Crisbr10/sequoia/internal/app      1.675s  coverage: 87.1%
ok  github.com/Crisbr10/sequoia/internal/codegraph 1.044s coverage: 82.1%
ok  github.com/Crisbr10/sequoia/internal/model    1.278s  coverage: [no statements]
ok  github.com/Crisbr10/sequoia/internal/pipeline 2.030s  coverage: 77.1%
ok  github.com/Crisbr10/sequoia/internal/tui      1.529s  coverage: 92.6%
ok  github.com/Crisbr10/sequoia/internal/tui/screens 1.832s coverage: 88.0%
ok  github.com/Crisbr10/sequoia/internal/tui/styles 1.501s coverage: 100.0%
ok  github.com/Crisbr10/sequoia/plugin            0.902s  coverage: 94.1%
ok  github.com/Crisbr10/sequoia/plugin/example    1.562s  coverage: 100.0%
ok  github.com/Crisbr10/sequoia/scripts           1.590s  coverage: [no statements]
```

**Result: 20/20 packages PASS** (1 `[no test files]` and 2 `[no statements]`
are normal and not failures).

**Runs #2, #3, #4, #5** (no coverage, faster):
All 5 consecutive runs green. No FAIL, no `--- FAIL`, no panic
(`Select-String -Pattern "FAIL|^---|panic"` returned 0 matches in
runs 3, 4, 5). The test suite is **stable** — no flakiness detected.

### 4.2 Per-function coverage on PR 2 surfaces

```
adapters/common/backup_path_builder.go:31:  NewBackupPathBuilder   100.0%
adapters/common/backup_path_builder.go:53:  Build                  85.7%
adapters/common/backup_retention.go:68:     BackupHomeDir          85.7%
adapters/common/backup_retention.go:85:     backupRootFrom         100.0%
adapters/common/backup_retention.go:108:    PruneBackups           77.4%
adapters/common/backup_retention.go:167:    hasSessionPrefix       83.3%
adapters/common/base_adapter.go:191:        LastBackupDir          100.0%
adapters/common/base_adapter.go:201:        SetLastBackupDir       0.0%   <-- see note
adapters/common/base_adapter.go:340:        CentralBackupDir       92.3%
adapters/common/base_adapter.go:369:        Prepare                95.2%
adapters/common/base_adapter.go:506:        Stage                  83.3%
adapters/common/base_adapter.go:561:        Apply                  94.4%
adapters/common/base_adapter.go:629:        Install                94.4%
internal/pipeline/runner.go:107:            runInstallSteps        94.3%
internal/pipeline/runner.go:484:            sendProgress           100.0%
```

`SetLastBackupDir` at 0% is a **tooling artifact**: it is called
from `Codex.Install` (line 166) which is exercised by
`TestAdapter_Install_BackupDirUsesCentralHome` (PASS), but Go's
coverage tool does not attribute the call to the method on
`BaseAdapter`. The functional coverage is real; a targeted direct
test would be a 5-line addition if the gap ever matters.

### 4.3 `go vet ./...`

Clean. No output, exit 0.

### 4.4 Adversarial grep — the 4+ test files the apply agent said "don't assert"

**File: `adapters/common/installer_test.go`**

Grep command: `grep -nE '\.sequoia-backup|BackupDir|LastBackupDir|backupPath'`
(excluding package code matched elsewhere).

**Result**: 5 matches, all in tests that pass `BackupDir` as a
**parameter** to `common.NewInstallerConfig` — not asserting where
the BaseAdapter puts it. Verbatim excerpts:

- `installer_test.go:47`: `backupDir := filepath.Join(t.TempDir(), "backup")` (parameter)
- `installer_test.go:88`: `BackupDir: backupDir,` in `InstallerConfig` (parameter)
- `installer_test.go:158-159`: `backupDir1 := filepath.Join(parent, ".sequoia-backup-"+suffix1)` — **synthesized test fixture** for the FIX-005 collision test
- `installer_test.go:308`: `backupDir := filepath.Join(t.TempDir(), ".sequoia-backup-perm-test")` — **synthesized test fixture** for the permission test
- `installer_test.go:353,361`: `predictableBackup := filepath.Join(parent, ".sequoia-backup")` and `timestampedBackup := filepath.Join(parent, ".sequoia-backup-abc123")` — **synthesized test fixtures** for the FIX-005 collision-isolation test

**None of these assert the BaseAdapter's actual backup path location.**
They test the installer's behavior with constructed backup paths. The
FIX-005 tests are explicitly about NOT colliding with pre-existing
predictable backup dirs — orthogonal to the central-home move.

**Verdict**: Apply agent's claim **VERIFIED CORRECT**.

---

**File: `adapters/codex/installer_test.go`**

Grep command: `grep -nE '\.sequoia-backup|BackupDir|LastBackupDir|backupPath'`

**Result**: 9 matches, all in tests of `codex.MergeConfig` and
`codex.RemoveConfig` — which are the **legacy ReplaceFile sidecar
functions** at `adapters/codex/installer.go:20,60`. Those functions
build `<configPath>.sequoia-backup-<suffix>` + `<configPath>.sequoia-session`
sidecars. This is **exactly** REQ-BRP-03 (PR 3 task 3.3-3.6
territory). The functions are NOT touched by PR 2
(verified: `git diff -- adapters/codex/installer.go` → empty).

Verbatim excerpts (from the grep output):
- `installer_test.go:123`: `if strings.Contains(e.Name(), ".sequoia-backup-")` — `TestMergeConfig_CreatesBackup` (line 102)
- `installer_test.go:134`: `_, err = os.Stat(configPath + ".sequoia-backup")` — same test
- `installer_test.go:195`: `backupPath := configPath + ".sequoia-backup-test123"` — `TestRemoveConfig_RestoresBackup` (line 186) — sets up a fake legacy backup to test the sidecar restore path
- `installer_test.go:208`: `_, err = os.Stat(backupPath)` — same test
- `installer_test.go:308`: `if strings.Contains(e.Name(), ".sequoia-backup-")` — `TestMergeConfig_BackupHasUniqueName` (line 285)
- `installer_test.go:324`: `oldBackup := configPath + ".sequoia-backup-old"` — `TestMergeConfig_ExistingBackupNotOverwritten` (line 318)
- `installer_test.go:345`: `if strings.Contains(e.Name(), ".sequoia-backup-")` — same test
- `installer_test.go:383`: `if strings.Contains(e.Name(), ".sequoia-backup-")` — `TestMergeConfig_BackupPermissions_Restricted` (line 356)
- `installer_test.go:415,437`: same pattern in `TestRemoveConfig_RestoresCorrectBackup` (line 396)

**All 9 matches assert the legacy ReplaceFile sidecar behavior** —
this code path is the strategy.go `ReplaceFile` territory (PR 3).
The new codex test `installer_internal_test.go:TestAdapter_Install_BackupDirUsesCentralHome`
is the test that DOES assert the central home prefix (PASS).

**Verdict**: Apply agent's claim **VERIFIED CORRECT**.

---

**File: `adapters/opencode/install_test.go`**

Grep command: `grep -nE '\.sequoia-backup|BackupDir|LastBackupDir|backupPath'`

**Result**: 1 match.

- `install_test.go:114`: `if strings.Contains(e.Name(), ".sequoia-backup-")` — inside `TestInstall_PreservesExistingAgentsMD` (line ~99). This test installs via `opencode.NewAdapter(tmp).Install(...)`, which goes through `BaseAdapter.Install` → `StrategyFileReplace` strategy → `ReplaceFile` in `strategy.go:120` (the **legacy** ReplaceFile path, untouched by PR 2). The backup at this line is the **system prompt backup** (`AGENTS.md.sequoia-backup-<suffix>`), NOT the `BaseAdapter` skills/commands backup. The skill/command backup (which PR 2 wired) goes to the central home, but this test inspects the strategy.go sidecar at the system prompt path.

**Verdict**: Apply agent's claim **VERIFIED CORRECT**. The single match
exercises the PR 3 territory `ReplaceFile` strategy, not the
BaseAdapter's central-home backup.

---

**File: `adapters/common/base_adapter_strategy_test.go`**

Grep command: `grep -nE '\.sequoia-backup|BackupDir|LastBackupDir|backupPath'`
on this file directly (not via path-truncated broader search).

**Result**: 0 matches for `.sequoia-backup`. The file uses
`backupPath` only in the helper-builder closure
`a.SetBackup(common.NewBackupPathBuilder(func(base string) string
{ return filepath.Join(base, "backup") }, "strategy-test"))` (line
46), where `backup` is just a subdir name in `t.TempDir()` — not
asserting a path location. The tests at lines 67, 85, 106, 122, 136,
152, 166, 193, 222, 243, 262, 286, 307 all test the Strategy phase
lifecycle (Prepare/Download/Verify/Stage/Apply/Rollback/Install) for
correctness of state transitions and error handling — none of them
assert the actual backup path.

**Verdict**: Apply agent's claim **VERIFIED CORRECT**.

---

**File: `adapters/common/base_adapter_error_test.go`**

Grep command: `grep -nE '\.sequoia-backup|BackupDir|LastBackupDir|backupPath'`
on this file directly.

**Result**: 1 match.

- `base_adapter_error_test.go:670`: `func TestInstall_SystemPromptFailure_RollbackBackupDir(...)` — this is a **function name**, not a path assertion. The test body (lines 673-715) sets up a `failingWriteAdapter` (via helper at line 698), installs with rollback=true, and asserts that `skills/SKILL.md` and `commands/...` are restored to their original content. The `BackupDir` configuration is inside the `failingWriteAdapter` helper, but the test never inspects the backup path's location — it only checks the destination file content.

**Verdict**: Apply agent's claim **VERIFIED CORRECT**.

---

**Conclusion on deviation #2**: All 5 files the apply agent reviewed
genuinely do not assert the BaseAdapter backup path location. The
`.sequoia-backup-` patterns in them are either:
- Synthesized test fixtures passed to `common.NewInstaller` (not the
  BaseAdapter's location)
- Assertions against the legacy `ReplaceFile` sidecar (PR 3
  territory, REQ-BRP-03)
- Function names that mention "BackupDir" as a generic concept
  (the test doesn't read the path)

The deviation is **justified**. No broken test assertions will land
on main.

### 4.5 Out-of-scope confirmation (re-verified)

```
$ git diff main..feature/backup-retention-pr2-installer --name-only \
    -- adapters/claude/paths.go adapters/codex/paths.go \
       adapters/cursor/paths.go adapters/gemini/paths.go \
       adapters/opencode/paths.go adapters/common/strategy.go
(empty)
```

The 5 per-adapter `paths.go` files are all still untouched — they
still return the legacy per-tool path
(`filepath.Join(base, ".sequoia-backup")`). This is **PR 3 task 3.11**
scope, correctly preserved here.

`strategy.go` (ReplaceFile + RestoreOrRemoveFile) is also untouched.
The legacy `.sequoia-backup-<suffix>` + `.sequoia-session` sidecar
is still in use for system prompt backups. PR 3 task 3.3-3.6 will
replace it with the central home + manifest.json.

### 4.6 Test pollution in real user config dir

```
$ ls $APPDATA/sequoia/backups/
Name         SubdirCount
----         -----------
err-test     16
opencode     16
```

**16 session dirs per adapter**, above the 5-backup cap. This is
**expected** and **documented** in the apply-progress §PR 2 Open
Risks #2: the cap is a soft bound that the `applyRetention` hook
(PR 3 task 3.7/3.8) will enforce. The pollution is the cumulative
test output from running the suite ~16 times across PR 1 + PR 2
verification cycles. The apply agent mentioned manually pruning via
`Remove-Item` (apply-progress §PR 2 Open Risks #2); I did NOT prune
during verification to confirm the cap is the only bound (it is).

**WARNING**: The pollution persists across CI runs until PR 3 lands
the retention hook. Not a CRITICAL finding for PR 2 (PR 2 is not
in scope to fix it), but a CI/cache housekeeping note for PR 3
verification.

### 4.7 Git status

```
$ git status --porcelain
 M openspec/changes/backup-retention-and-organization/apply-progress.md
```

One file modified, uncommitted: `apply-progress.md`. This is the
orchestrator's expected state (the agent merged PR 1 + PR 2 progress
into the file rather than overwriting; per the orchestrator's
handoff: "the agent merged, did not overwrite"). The single file is
not a code change and not a verification concern.

---

## 5. Deviation Assessment

### 5.1 Deviation #1: `CentralBackupDir` is exported (uppercase) — `accept-with-rationale`

**Design said**: "REFACTOR — `centralBackupDir` helper" (lowercase).

**Implementation**: `CentralBackupDir` (uppercase, exported).

**Rationale** (from apply-progress §PR 2 Open Risks #6): "Codex's
custom `Install` needed access."

**Verification**: I confirmed the rationale by reading
`adapters/codex/adapter.go:127`:
```go
baseBackup := a.CentralBackupDir("")
```
The Codex adapter (a concrete adapter in a different package,
`adapters/codex`) calls `a.CentralBackupDir(...)` from outside the
`adapters/common` package. A lowercase `centralBackupDir` would not
be accessible across packages.

**Code-reviewer impact**: LOW. The helper name is descriptive
(`CentralBackupDir`), the contract is clear in the doc comment
(`base_adapter.go:321-339`), the signature is one parameter and a
return string, and it is mutex-protected internally. The
alternative (a callback-based approach) would add more API surface
for no real benefit. The export is a *deliberate* design decision
in response to the Codex cross-package need.

**Classification**: `accept-with-rationale` — not even a WARNING.
The deviation is well-documented and the cross-package access
pattern is the correct architectural choice.

### 5.2 Deviation #2: Task 2.4 updated only 1 of the 4+ files — `accept-with-rationale`

**Task said**: "Update 4+ test files with old `.sequoia-backup` path
assertions."

**Implementation**: Updated only `adapters/common/base_adapter_test.go`
(`TestBackupIsolation_NamespacedBackupStructure`).

**Rationale** (from apply-progress §PR 2 Task Status #2.4): "The
other 4+ files ... were reviewed and found NOT to assert the
BaseAdapter backup path location."

**Verification**: I re-grepped all 5 files the apply agent listed
(`adapters/common/installer_test.go`, `adapters/codex/installer_test.go`,
`adapters/opencode/install_test.go`, `adapters/common/base_adapter_strategy_test.go`,
`adapters/common/base_adapter_error_test.go`) and found that:

- `adapters/common/installer_test.go`: All 5 matches are **synthesized
  test fixtures** passed as `BackupDir` parameters to
  `common.NewInstallerConfig`. They test the installer's behavior
  with constructed backup paths, not the BaseAdapter's location.
- `adapters/codex/installer_test.go`: All 9 matches assert the
  **legacy ReplaceFile sidecar** at `adapters/codex/installer.go:20,60`
  — this is REQ-BRP-03 / PR 3 task 3.3-3.6 territory.
- `adapters/opencode/install_test.go`: 1 match in
  `TestInstall_PreservesExistingAgentsMD` exercises the
  `StrategyFileReplace` → `ReplaceFile` (strategy.go) path — also
  PR 3 territory.
- `adapters/common/base_adapter_strategy_test.go`: 0 matches for
  `.sequoia-backup`. Tests Strategy phase lifecycle correctness,
  not path location.
- `adapters/common/base_adapter_error_test.go`: 1 match is a
  **function name** (`TestInstall_SystemPromptFailure_RollbackBackupDir`).
  The test body does not inspect the backup path.

Full evidence: §4.4 above. The apply agent's claim is **verified
correct** — these test files legitimately do not need updates in
PR 2.

**Classification**: `accept-with-rationale` — justified.

**Note for PR 3 verification**: The PR 3 apply agent MUST update
`adapters/codex/installer_test.go` (the 9 matches) and
`adapters/opencode/install_test.go` (the 1 match) when REQ-BRP-03
lands the manifest.json-based ReplaceFile. The 4+ files referenced
in the original task spec are correctly PR 3 scope — the task
description was slightly off.

### 5.3 Deviation #3: Test pollution above cap — `WARNING`

**Apply agent mitigation**: "The `PruneBackups` cap of 5 is the only
bound" + "manually `Remove-Item` of the test pollution".

**Verification**: I confirmed 16 dirs per adapter exist (the pollution
is real, the cap is not yet enforced). See §4.6.

**Classification**: **WARNING** — Not a CRITICAL because:
1. The pollution is bounded by the user's `os.UserConfigDir()` size
   (no security/safety issue).
2. The cap is intentionally PR 3 work (applyRetention hook).
3. The PR 2 test files use `t.Cleanup` for the new codex test
   (`installer_internal_test.go:41` removes the codex subdir on
   teardown) and `overrideUserConfigDir` for the base_adapter
   internal tests (`base_adapter_internal_test.go:23,79,115,167`) —
   these do NOT pollute the real user config.

The 16 dirs come from the existing pre-PR-2 tests that touch the
real config dir (e.g., `backup_path_builder_test.go`).

**Risk for PR 3**: The retention hook MUST be exercised with a
real-pollution scenario (16 → 5) to confirm `PruneBackups` works
under the actual condition. The PR 3 task 3.7 E2E test ("6 installs
→ exactly 5 session dirs") will not see this; recommend a PR 3
follow-up test that pre-seeds 7 dirs, calls `applyRetention`, and
asserts ≤ 5 remain.

---

## 6. Spec Ambiguities (carried over from PR 1 verify)

### 6.1 REQ-BRP-02: Session dir timestamp format — example vs implementation

**Status**: Unchanged from PR 1. The implementation uses
`2006-01-02T15-04-05.000Z` (`.` between SS and mmm); the spec
example uses `-` between SS and mmm. Both are valid ISO-8601; the
formatter/parser are internally consistent. The lex-sort ==
chron-sort invariant is preserved.

**PR 2 relevance**: PR 2 uses the same formatter via
`base_adapter.go:353`. No new ambiguity introduced.

**Recommend**: Spec clarification in `sdd-archive` or a follow-up
`sdd-propose` change. Not a code issue.

### 6.2 REQ-BRP-06: "continues on error" scenario numbering

**Status**: Unchanged from PR 1. The spec's "directory #4 is
read-only, expect removed=2" is internally inconsistent with
max=5 and 7 entries. The implementation correctly tests the
"continues-on-error" contract with `removed=1` (oldest read-only,
next-oldest succeeds).

**PR 2 relevance**: None — PR 2 does not touch `PruneBackups`.

**Recommend**: Scenario rewrite in spec at `sdd-archive` time.

---

## 7. Risks for PR 3

1. **Test pollution from PR 1 + PR 2 runs persists** — 16 dirs per
   adapter under `$APPDATA/sequoia/backups/err-test/` and `opencode/`.
   PR 3 task 3.8 (`applyRetention` hook) will bound this to 5. PR 3
   verification should pre-clean (or expect the cap to engage) to
   get a clean signal from the "6 installs → 5 dirs" E2E test.

2. **The 9 `.sequoia-backup-` matches in `adapters/codex/installer_test.go`
   and 1 match in `adapters/opencode/install_test.go` are PR 3 work**
   (REQ-BRP-03 / task 3.3-3.6). They test the legacy `ReplaceFile`
   sidecar. PR 3 must update these tests to assert the central-home
   `<root>/<adapterID>/<session>/<basename>.backup` + `manifest.json`
   format. The task description for 2.4 was misleading; the actual
   PR 3 surface is these 2 files, not 5.

3. **`adapters/<tool>/paths.go` (5 files) still return
   `filepath.Join(base, ".sequoia-backup")`** — PR 3 task 3.11
   must update them to delegate to `common.BackupHomeDir()`. Once
   they do, the `BackupPathBuilder.Build()` safety-net fallback
   becomes unreachable — decide: keep for resilience (one extra
   indirection) or remove for clarity.

4. **`CentralBackupDir` cache is per-instance mutex-protected** —
   correct for the per-install sequential use case. If a future
   feature installs the same adapter concurrently from multiple
   goroutines on the same instance, the cache will create one
   session dir for the whole instance (not per goroutine). PR 3
   shouldn't change this; document if relevant.

5. **`SetLastBackupDir` has 0% direct coverage** (tooling artifact).
   The call IS exercised through `Codex.Install` in the new
   `installer_internal_test.go` test (PASS). If coverage tooling
   ever matters, a 5-line direct test is trivial.

6. **`BackupPathBuilder.Build` 85.7% coverage** — slightly down
   from PR 1 (85.7% same). The fallback path (`Build` returning
   `b.backupPathFn(base)+"-"+b.adapterID+"-"+sessionSuffix` at
   `backup_path_builder.go:62`) is unreachable in PR 2 because
   `Codex.Install` now uses `CentralBackupDir` exclusively.
   PR 3 task 3.11 will either delete the fallback (coverage goes
   up) or keep it for resilience (coverage stays at ~85%).

7. **CRLF line endings on Windows-edited files** — same
   housekeeping issue as PR 1 verify §4.4. `golangci-lint` reports
   gofmt on 5 pre-existing files in `adapters/common/`. Not
   introduced by PR 2; CI on Linux/macOS will not see it.

8. **TUI `Info` does NOT yet mention retention count** — the spec
   doesn't require it (the Info message is for pre-existing
   scattered backups, not retention). PR 3 may want to add
   "5 most recent kept; N removed" to the Info message after
   `applyRetention` runs — confirm with product before adding.

---

## 8. Out-of-Scope Confirmation (explicit)

| Out-of-scope area | File(s) | Status | Evidence |
|---|---|---|---|
| Per-adapter paths.go | `adapters/claude/paths.go`, `adapters/codex/paths.go`, `adapters/cursor/paths.go`, `adapters/gemini/paths.go`, `adapters/opencode/paths.go` | **NOT MODIFIED** | `git diff main..feature/backup-retention-pr2-installer --name-only` on these 5 files returns empty. `codex/paths.go:32` and `opencode/paths.go:32` still return `filepath.Join(base, ".sequoia-backup")` (verified by reading the files). PR 3 task 3.11 will update. |
| `ReplaceFile` | `adapters/common/strategy.go:120` | **NOT MODIFIED** | `git diff main..feature/backup-retention-pr2-installer -- adapters/common/strategy.go` empty. Function still uses `.sequoia-backup-<suffix>` + `.sequoia-session` sidecar. PR 3 task 3.3/3.4 will update. |
| `RestoreOrRemoveFile` | `adapters/common/strategy.go:165` | **NOT MODIFIED** | same as above. PR 3 task 3.5/3.6 will update. |
| `BaseAdapter.Apply` retention hook | `adapters/common/base_adapter.go:561` | **NOT WIRED** | `git diff main..feature/backup-retention-pr2-installer -- adapters/common/base_adapter.go` does not contain `applyRetention` or `PruneBackups` in the diff. Confirmed: `grep -E "applyRetention\|PruneBackups" base_adapter.go` returns no matches (verified via `git diff | grep -E "applyRetention|PruneBackups"` → empty). PR 3 task 3.8 will wire it. |
| `Codex.Installer` (ReplaceFile sidecar) | `adapters/codex/installer.go` | **NOT MODIFIED** | `git diff main..feature/backup-retention-pr2-installer -- adapters/codex/installer.go` empty. The `MergeConfig`/`RemoveConfig` legacy sidecar functions are unchanged. PR 3 will replace with manifest-based restore. |
| TUI retention warning | `internal/pipeline/runner.go` | **NOT YET ADDED** | The new Info message at line 208 covers pre-existing scattered backups (REQ-BRP-05). The retention count warning is not in the spec's REQ-BRP-05 wording. PR 3 may add a separate warning via `AddWarning` (REQ-BRP-04 scenario "Removal errors do not fail the install"). |
| 4+ test files with `.sequoia-backup` assertions | `adapters/common/installer_test.go`, `adapters/codex/installer_test.go`, `adapters/opencode/install_test.go`, `adapters/common/base_adapter_strategy_test.go`, `adapters/common/base_adapter_error_test.go` | **NOT UPDATED (justified)** | See §5.2 and §4.4 for the per-file analysis. The matches in these files are either PR 3 territory (`codex/installer_test.go`, `opencode/install_test.go`) or don't assert the BaseAdapter path (the other 3). |

---

## 9. Recommendation

**`proceed-to-pr-3-with-acknowledged-warnings`**

Rationale:
- All 8 PR 2 tasks PASS.
- Full test suite green (20/20 packages).
- 5 consecutive test runs in a row show no flakiness.
- `adapters/common` coverage 85.5%, `internal/pipeline` coverage
  77.1%, all PR 2 functions above 70% gate.
- `go vet ./...` clean.
- The "only 1 of 4+ test files updated" deviation is **verified
  correct** by independent re-grep — the other files legitimately
  don't assert the BaseAdapter path; the matches are PR 3 scope
  (`ReplaceFile` sidecar) or test-fixture strings.
- The `CentralBackupDir` export deviation is justified by
  cross-package access from `adapters/codex`.
- The test pollution above cap is **expected** (PR 3 will bound it)
  and the WARNING is non-blocking.
- 2 SPEC AMBIGUITIES (timestamp format example; "directory #4"
  numbering) are documentation issues carried over from PR 1.
- 1 SUGGESTION (TDD commit shape) is a process improvement
  carried over from PR 1.
- 5 OUT OF SCOPE areas (REQ-BRP-03, REQ-BRP-04 e2e, per-adapter
  paths.go, retention hook, codex installer legacy) are correctly
  untouched.

PR 2 is clean, focused, and well-tested. PR 3
(`feature/backup-retention-pr3-replacefile`) can start on a branch
off `main` after PR 2 merges.

---

## Artifacts

- Verify report (this file):
  `openspec/changes/backup-retention-and-organization/verify-report-pr2.md`
- Coverage profile: `coverage.out` (at repo root)
- Per-function coverage: `coverage_func.txt` (at repo root)

---

## Structured Envelope (return value)

```json
{
  "status": "pass-with-warnings",
  "executive_summary": "PR 2 installer wiring is clean, focused, and well-tested: 8/8 tasks PASS, full test suite green (20/20 packages, 5 consecutive runs, no flakiness), go vet clean. The 3 known deviations were independently verified: (1) CentralBackupDir export is justified by cross-package access from codex.Adapter, (2) the 4+ test files the apply agent said 'don't assert the BaseAdapter path' genuinely don't — all .sequoia-backup matches are either PR 3 scope (ReplaceFile sidecar) or test-fixture strings, (3) test pollution above the 5-dir cap is expected because the applyRetention hook is PR 3 work. Per-function coverage on all PR 2 surfaces is well above the 70% gate (CentralBackupDir 92.3%, Prepare 95.2%, Stage 83.3%, runInstallSteps 94.3%, sendProgress 100%). No CRITICAL findings. 1 WARNING (test pollution above cap, will be bound by PR 3). 1 SUGGESTION (TDD commit shape, carried from PR 1). 2 SPEC AMBIGUITIES carried from PR 1. 5 out-of-scope areas correctly untouched.",
  "artifacts": {
    "verify_report": "openspec/changes/backup-retention-and-organization/verify-report-pr2.md",
    "coverage_profile_root": "coverage.out",
    "per_function_coverage": "coverage_func.txt"
  },
  "next_recommended": "proceed-to-pr-3-with-acknowledged-warnings",
  "risks": [
    "Test pollution from PR 1 + PR 2 test runs persists: 16 dirs per adapter in $APPDATA/sequoia/backups/err-test/ and opencode/, above the 5-dir cap. PR 3 task 3.8 (applyRetention hook) will bound this. The cap is the only bound until then.",
    "The 9 .sequoia-backup- matches in adapters/codex/installer_test.go and 1 match in adapters/opencode/install_test.go are PR 3 work (REQ-BRP-03 / strategy.go manifest). The original task 2.4 description was misleading about which files needed PR 2 vs PR 3 updates.",
    "The 5 adapters/<tool>/paths.go files still return filepath.Join(base, '.sequoia-backup') (verified by reading codex/paths.go:32 and opencode/paths.go:32). PR 3 task 3.11 must update them.",
    "CentralBackupDir cache is per-instance mutex-protected (a.centralSessionDir cleared in Prepare). Correct for sequential per-install use; would surprise concurrent goroutines on the same instance. Document if relevant.",
    "SetLastBackupDir has 0% direct coverage (tooling artifact — call IS exercised via Codex.Install in the new test, but Go's coverage tool doesn't attribute the call to the BaseAdapter method). Trivial 5-line direct test if it ever matters.",
    "BackupPathBuilder.Build safety-net fallback (line 62) becomes unreachable after PR 3 task 3.11. Decide: keep for resilience (one extra indirection, never on happy path) or remove for clarity (Build's coverage will go to 100%).",
    "Spec REQ-BRP-02 example timestamp format (2026-06-15T15-30-45-123Z) differs from implementation (2026-06-15T15-30-45.000Z) — carried from PR 1, recommend spec clarification at sdd-archive",
    "Spec REQ-BRP-06 'continues on error' scenario 'directory #4 is read-only' is internally inconsistent — carried from PR 1, recommend scenario rewrite at sdd-archive",
    "TUI Info does not yet include a retention count (e.g., '5 most recent kept; N removed'). Not required by the spec; PR 3 may add if product wants it (would be a separate warning via AddWarning, not in Info).",
    "5 pre-existing files in adapters/common/ have CRLF line endings (gofmt issue on Windows). Pre-existing, not introduced by PR 2. CI on Linux/macOS will not see it. Recommend gofmt -w housekeeping at merge time."
  ]
}
```
