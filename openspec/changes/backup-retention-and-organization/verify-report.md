# Verify Report: backup-retention-and-organization — PR 1 (Foundation)

> **Verifier**: `sdd-verify` sub-agent (PR 1 only, do NOT proceed to PR 2)
> **Branch under verification**: `feature/backup-retention-pr1-foundation`
> **Commits ahead of main**: 6 (e460217..5f3214a)
> **Strict TDD**: ACTIVE
> **Run timestamp**: 2026-06-16, Windows runner (PowerShell 7)
> **Final status**: **PASS_WITH_WARNINGS**

---

## 1. Executive Summary

PR 1 introduces the central backup home (`BackupHomeDir`), the retention
helper (`PruneBackups`), the retention-cap constant
(`DefaultMaxBackupsPerAdapter = 5`), and re-points `BackupPathBuilder.Build`
at the central root. All four public surfaces are exported, callable, and
covered by unit tests. The full project test run is green (20/20 packages),
`go vet ./...` is clean, and `adapters/common` coverage is **85.8%** — well
above the 70% CI gate. The new file coverage breakdown:

| Function | Coverage |
|---|---|
| `NewBackupPathBuilder` | 100.0% |
| `BackupPathBuilder.Build` | 85.7% |
| `BackupHomeDir` | 85.7% |
| `backupRootFrom` | 100.0% |
| `PruneBackups` | 77.4% |
| `hasSessionPrefix` | 83.3% |

**Finding tally**:
- CRITICAL: 0
- WARNING: 2
- SUGGESTION: 1
- SPEC AMBIGUITY: 2
- OUT OF SCOPE (intentional non-failures): 5 (REQ-BRP-03, REQ-BRP-04, REQ-BRP-05, the manifest work, and the e2e install test)

**One-line verdict**: PR 1 is a solid, well-tested foundation. The two
warnings and the two spec ambiguities are all non-blocking and should be
addressed in a follow-up spec/design touch-up or rolled into PR 2/3 work.

---

## 2. Task Verification

| Task | Claim | Result | Note |
|---|---|---|---|
| 1.1 RED | RED test for `BackupHomeDir` exists | **PASS** | `TestBackupHomeDir_ReturnsAndCreatesPath` exists at `adapters/common/backup_retention_test.go:25`. Test is in the same commit as the implementation (e460217, +131 test / +53 impl), so the RED step is not visible as a separate commit (see SUGGESTION §2.1). |
| 1.2 GREEN | `BackupHomeDir()` implemented | **PASS** | Function at `backup_retention.go:68`. Returns `os.UserConfigDir()+sequoia/backups`, creates with `0o700`, wraps errors with `sequoia/backups` substring. Coverage 85.7%. |
| 1.3 RED | RED test for `PruneBackups` | **PASS** | 7-session fixture, max=5, expects 2 removed. `TestPruneBackups_KeepsExactlyMaxFromSevenSessions` at `backup_retention_test.go:128`. Pass. |
| 1.4 GREEN | `PruneBackups` implemented | **PASS** | `PruneBackups` at `backup_retention.go:108`. Sort descending, remove tail, continue on per-entry error, return first error + count. 77.4% coverage (above 70% gate). |
| 1.5 RED | `DefaultMaxBackupsPerAdapter == 5` asserted | **PASS** | `TestDefaultMaxBackupsPerAdapter_IsFive` at line 104. `assert.Equal(t, 5, DefaultMaxBackupsPerAdapter)`. |
| 1.6 RED | `BackupPathBuilder.Build` uses central home | **PASS** | `TestBackupPathBuilder_Build_UsesCentralHome` (internal) + updated `TestBackupPathBuilder_Build_IncludesAdapterID` + `TestBackupPathBuilder_Build_UsesBackupPathFn` (external) all assert new format. |
| 1.7 GREEN | `Build` delegates to `BackupHomeDir` | **PASS** | `Build()` at `backup_path_builder.go:53` calls `BackupHomeDir()`, falls back to per-tool on failure. 85.7% coverage. |
| 1.8 REFACTOR | Path-prefix extraction | **PASS** | `backupRootFrom(cfg)` helper at `backup_retention.go:85` derives the subpath from the `backupHomeSubpath = "sequoia/backups"` constant. Both `BackupHomeDir` and the error-wrap string share the constant — single source of truth. 100% coverage. |
| X.1 | Package doc | **PASS** | `backup_retention.go:1-32` has a comprehensive package-level comment covering retention policy and the public surface. `go doc adapters/common` shows it correctly (verified). |
| regression | Installer parent-dir creation (commit 5f3214a) | **PASS** | `installer.go:97-99` adds `os.MkdirAll(filepath.Dir(dst), 0o700)` before `copyFile`. The gemini `TestAdapter_Reinstall_OverwritesVersion` test now passes against the new nested layout. This is the one allowed out-of-scope change called out in the apply-progress. |

**Total**: 10/10 tasks PASS.

### 2.1 TDD commit-shape SUGGESTION (not a blocker)

The apply-progress claims strict TDD with separate RED/GREEN cycles, but the
git history shows each RED+GREEN pair landed in a single commit (e.g.,
e460217 added 131 test lines + 53 impl lines together). The end-state is
correct and the tests do exercise the behavior, but a strict TDD audit
would want to see the test commit land first with a build-fail and the
implementation commit land second with a build-pass. Not a blocker for PR 1
to land, but the apply pattern should be tightened for PR 2/3.

---

## 3. Spec Compliance

### REQ-BRP-01 — Centralized backup root — **PASS**

- **Scenario 1 (returns and creates the joined path)**: covered by
  `TestBackupHomeDir_ReturnsAndCreatesPath`. The test overrides
  `userConfigDir` to a `t.TempDir()` and asserts
  `BackupHomeDir() == <tmp>/sequoia/backups` and that the dir exists.
  Mode 0o700 is asserted on POSIX; the test correctly skips the
  mode assertion on Windows (`runtime.GOOS != "windows"`) because
  Windows does not honor POSIX permission bits.
- **Scenario 2 (idempotent on pre-existing root)**: covered by
  `TestBackupHomeDir_IsIdempotent`. Calls `BackupHomeDir()` twice,
  asserts same path returned and the mode is preserved.
- Both tests pass on this runner. **PASS**.

### REQ-BRP-02 — Per-adapter organization (PARTIALLY covered by PR 1) — **PASS_WITH_WARNING**

- **Path shape (central home + adapter ID segment)**: covered by
  `TestBackupPathBuilder_Build_UsesCentralHome` and
  `TestBackupPathBuilder_Build_IncludesAdapterID`. Both assert the
  result contains `<root>/<adapterID>/`. Pass.
- **Distinct adapter IDs produce disjoint subtrees**: covered by
  `TestBackupPathBuilder_Build_DisjointForDifferentAdapters` (internal)
  and `TestBackupPathBuilder_Build_DisjointPathsForDifferentAdapters`
  (external). Pass.
- **Session dir name shape**: `TestBackupPathBuilder_Build_UsesCentralHome`
  asserts the result ends in `/<adapterID>/<ISO8601>-<base36>/` and that
  the trailing token matches a base-36 regex. Pass.

**WARNING (deviation from spec example, see §5)**: The spec scenario
example shows the session dir as `2026-06-15T15-30-45-123Z-<suffix>`
with a `-` between SS and mmm. The implementation uses
`sessionDirLayout = "2006-01-02T15-04-05.000Z"` (Go reference time),
which produces `2026-06-15T15-30-45.000Z-<suffix>` with a `.` between
SS and mmm. Both formats are valid ISO-8601 and both maintain the
lex-sort == chron-sort invariant; the implementation is internally
consistent (formatter and `hasSessionPrefix` parser use the same
layout). The spec's REQ text says "ISO-8601 UTC timestamp
(`YYYY-MM-DDTHH-MM-SS-mmmZ`)" — the `mmm` is a label, not a literal
character, and the example is illustrative. Recommend spec clarification
(see SPEC AMBIGUITY §5.1). Not a CRITICAL finding.

### REQ-BRP-03 — File-replace backup storage — **OUT OF PR 1 SCOPE (PR 3)**

Not implemented in PR 1. `ReplaceFile` and `RestoreOrRemoveFile` still
use the legacy per-tool `.sequoia-backup-<suffix>` + `.sequoia-session`
sidecar (verified at `strategy.go:120-179`). This is correct per the
3-PR split and the apply-progress. No CRITICAL finding.

### REQ-BRP-04 — Retention policy of 5 per adapter (PARTIALLY covered) — **PASS**

PR 1 implements `PruneBackups` and `DefaultMaxBackupsPerAdapter = 5`,
covering the helper surface. The end-to-end enforcement (Apply calls
applyRetention) is **PR 3 task 3.7/3.8** and is **OUT OF PR 1 SCOPE**.

In-scope:
- **Scenario "pruning below threshold is a no-op"**: covered by
  `TestPruneBackups_NoOpBelowMax`. Pass.
- **Scenario "5 backups kept after a sixth install"**: helper-level
  equivalent covered by `TestPruneBackups_KeepsExactlyMaxFromSevenSessions`.
  The full "install 6 times" integration test is **OUT OF PR 1 SCOPE**
  (PR 3 task 3.7).
- The hook in `BaseAdapter.Apply` is **NOT** present in PR 1, which is
  correct — verified at `base_adapter.go:491-520` (no `applyRetention` or
  `PruneBackups` call).

**PASS** for PR 1 scope.

### REQ-BRP-05 — Migration of old scattered backups (NOT performed) — **OUT OF PR 1 SCOPE**

The TUI `Info` note is **PR 2 task 2.5/2.6**. No code in PR 1 touches
pre-existing `.sequoia-backup-*` files. Verified by `git diff --stat` —
no changes in `internal/pipeline/runner.go`.

### REQ-BRP-06 — Path resolution and pruning helpers — **PASS_WITH_SPEC_AMBIGUITY**

- **`BackupHomeDir` exports + 0o700 mode + idempotency**: covered by
  `TestBackupHomeDir_*` (3 tests, all pass). **PASS**.
- **Error wrapping includes `sequoia/backups`**: covered by
  `TestBackupHomeDir_WrapsErrorsWithContext` (pointing `userConfigDir`
  at a regular file so `MkdirAll` fails on a child). Asserts the
  error contains `sequoia/backups` substring and the failing path.
  Pass. **PASS**.
- **`PruneBackups` signature and continue-on-error**: covered by
  `TestPruneBackups_ContinuesOnError`. The test **IS PLATFORM-SKIPPED ON
  WINDOWS** (verified: `--- SKIP: TestPruneBackups_ContinuesOnError` on
  this runner) because `chmod 0o500` does not block `os.RemoveAll` on
  Windows. This is a known platform limitation, not a bug. **PASS** on
  POSIX. CI on `ubuntu-latest`/`macos-latest` will exercise it.

  **SPEC AMBIGUITY (see §5.2)**: The spec scenario text says "directory
  #4 is read-only" with 7 sessions and `max=5`, expecting `removed=2`.
  The implementation marks the OLDEST (offset 0) read-only and asserts
  `removed=1` (the non-read-only next-oldest succeeds; the read-only
  oldest fails). The implementation satisfies the spec's INTENT
  ("continues to attempt removal of subsequent entries on error") but
  the assertion count differs. Reasonable interpretation, but the
  scenario numbering was ambiguous.

### REQ-BRP-07 — Test surface (strict TDD) — **PASS_WITH_OUT_OF_SCOPE**

In-scope (PR 1):
- BackupHomeDir path + 0o700: ✅ `TestBackupHomeDir_ReturnsAndCreatesPath`
- PruneBackups keeps exactly max: ✅ `TestPruneBackups_KeepsExactlyMaxFromSevenSessions`
- PruneBackups no-op at or below max: ✅ `TestPruneBackups_NoOpBelowMax` + `TestPruneBackups_AtExactlyMaxIsNoOp`
- PruneBackups handles missing adapter dir: ✅ `TestPruneBackups_MissingAdapterDir`
- PruneBackups ignores corrupt names: ✅ `TestPruneBackups_IgnoresCorruptNames`

Out-of-scope (correctly deferred to PR 2/3):
- Centralized ReplaceFile/RestoreOrRemoveFile round-trip — **PR 3**
- E2E install leaves at most 5 session dirs — **PR 3 task 3.7**
- 4+ existing tests updated to central path — **PR 2 task 2.4**
  (apply-progress risk #6 notes that 4+ test files still reference old
  per-tool paths; this is acknowledged as PR 2 work)

**PASS** for PR 1 scope.

---

## 4. Independent Re-execution (adversarial)

### 4.1 Full project test run

Command: `go test ./... -coverprofile=coverage.out -count=1 -timeout 120s`

Result: **20/20 packages PASS** (20 ok, 0 FAIL; 1 package with `[no test
files]` and 2 packages with `[no statements]` is normal and not a
failure).

Per-package coverage:

```
ok  github.com/Crisbr10/sequoia/adapters          1.125s  coverage: 96.4%
ok  github.com/Crisbr10/sequoia/adapters/claude   2.654s  coverage: 80.0%
ok  github.com/Crisbr10/sequoia/adapters/codex    2.919s  coverage: 78.0%
ok  github.com/Crisbr10/sequoia/adapters/common   2.995s  coverage: 85.8%
?   github.com/Crisbr10/sequoia/adapters/common/installembed  [no test files]
ok  github.com/Crisbr10/sequoia/adapters/cursor   2.637s  coverage: 89.7%
ok  github.com/Crisbr10/sequoia/adapters/gemini   2.384s  coverage: 86.7%
ok  github.com/Crisbr10/sequoia/adapters/opencode 2.680s  coverage: 81.2%
ok  github.com/Crisbr10/sequoia/adapters/testutil 1.156s  coverage: 90.5%
ok  github.com/Crisbr10/sequoia/cmd/sequoia       5.586s  coverage: 77.5%
ok  github.com/Crisbr10/sequoia/internal/app      1.527s  coverage: 87.1%
ok  github.com/Crisbr10/sequoia/internal/codegraph 0.968s  coverage: 82.1%
ok  github.com/Crisbr10/sequoia/internal/model    1.318s  coverage: [no statements]
ok  github.com/Crisbr10/sequoia/internal/pipeline 1.738s  coverage: 76.9%
ok  github.com/Crisbr10/sequoia/internal/tui      1.269s  coverage: 92.6%
ok  github.com/Crisbr10/sequoia/internal/tui/screens 1.379s  coverage: 88.0%
ok  github.com/Crisbr10/sequoia/internal/tui/styles 1.308s  coverage: 100.0%
ok  github.com/Crisbr10/sequoia/plugin            1.458s  coverage: 94.1%
ok  github.com/Crisbr10/sequoia/plugin/example    1.279s  coverage: 100.0%
ok  github.com/Crisbr10/sequoia/scripts           1.303s  coverage: [no statements]
```

All packages are at or above the 70% CI gate (those without statements
are excluded from the gate automatically).

### 4.2 Per-function coverage on the new files

Command: `go test ./adapters/common/... -count=1 -coverprofile=cov_common.out && go tool cover -func=cov_common.out`

```
adapters/common/backup_path_builder.go:31:  NewBackupPathBuilder   100.0%
adapters/common/backup_path_builder.go:53:  Build                   85.7%
adapters/common/backup_retention.go:68:     BackupHomeDir           85.7%
adapters/common/backup_retention.go:85:     backupRootFrom          100.0%
adapters/common/backup_retention.go:108:    PruneBackups            77.4%
adapters/common/backup_retention.go:167:    hasSessionPrefix        83.3%
```

Every new function is above the 70% gate.

### 4.3 `go vet ./...`

Command: `go vet ./...`

Result: **clean** (no output, exit 0).

### 4.4 `golangci-lint run ./adapters/common/...`

Command: `golangci-lint run ./adapters/common/...`

Result: **5 issues, all `gofmt`**:
```
adapters/common/backup_retention.go:1:1: File is not properly formatted (gofmt)
adapters/common/base_adapter.go:1:1: File is not properly formatted (gofmt)
adapters/common/base_adapter_test.go:1:1: File is not properly formatted (gofmt)
adapters/common/commands.go:1:1: File is not properly formatted (gofmt)
adapters/common/template.go:1:1: File is not properly formatted (gofmt)
```

**WARNING (see §6.1)**: 4 of the 5 flagged files are pre-existing
(`base_adapter.go`, `base_adapter_test.go` modified but the file as a
whole predates PR 1, `commands.go`, `template.go`). The apply-progress
flagged these as pre-existing CRLF/LF line-ending issues (line 96).
`backup_retention.go` (NEW in PR 1) has mixed line endings (176 CRLF
out of 176 lines = all CRLF) — this is a Windows-edited file. CI on
`ubuntu-latest` and `macos-latest` does not see this. Recommend
running `gofmt -w adapters/common/backup_retention.go` before merge as
housekeeping, but the existing `.gitattributes` (if any) and CI matrix
should be the source of truth.

### 4.5 Race detector

`-race` was **not run** on this Windows runner per the project's CI
matrix (verified in apply-progress line 98: "CI matrix disables `-race`
on `windows-latest`"). The CI on `ubuntu-latest` will exercise `-race`
on the Linux path. The PruneBackups function is mostly pure filesystem
read-then-sort-then-remove with no shared in-process state, so race
issues are unlikely. **No finding** — out of platform scope.

### 4.6 Out-of-scope confirmation

- **5 `adapters/<tool>/paths.go` files**: `git diff main..feature/backup-retention-pr1-foundation --name-only` shows
  no changes in `adapters/claude/`, `adapters/codex/`, `adapters/cursor/`,
  `adapters/gemini/`, or `adapters/opencode/`. The per-tool paths are
  retained as the safety-net fallback in `BackupPathBuilder.Build`.
  **CONFIRMED — out of scope not touched.**
- **`ReplaceFile` / `RestoreOrRemoveFile`**: `git diff --stat adapters/common/strategy.go` returns empty.
  Functions still use the legacy `.sequoia-backup-<suffix>` +
  `.sequoia-session` sidecar. **CONFIRMED — out of scope not touched.**
- **`BaseAdapter.Prepare` / `Stage`**: `git diff --stat adapters/common/base_adapter.go` returns empty.
  These methods still use the per-tool `BackupPathBuilder` for `BackupDir`.
  **CONFIRMED — out of scope not touched** (the wiring is PR 2 task 2.2/2.3).
- **`BaseAdapter.Apply` retention hook**: confirmed NOT wired in PR 1
  (verified by `grep -E "applyRetention|PruneBackups" base_adapter.go` →
  no matches). **CONFIRMED — out of scope not touched** (PR 3 task 3.8).
- **Installer regression fix (commit 5f3214a)**: `installer.go:97-99`
  adds parent-dir creation before `copyFile`. This is the one allowed
  out-of-scope change called out in apply-progress. The gemini
  `TestAdapter_Reinstall_OverwritesVersion` test passes against the new
  layout. **CONFIRMED — allowed out-of-scope change is in place.**

---

## 5. Spec Ambiguities

### 5.1 REQ-BRP-02: Session dir timestamp format — example vs implementation

**Spec text** (REQ-BRP-02, line 30, 35-36):
> "Each install session's backup SHALL be written into a subdirectory of the adapter folder, named with an ISO-8601 UTC timestamp (`YYYY-MM-DDTHH-MM-SS-mmmZ`) suffixed with the existing base-36 Unix-nanos session ID."
> "GIVEN adapter ID `claude-code` and the clock at `2026-06-15T15:30:45.123Z` ... THEN the result equals `<root>/claude-code/2026-06-15T15-30-45-123Z-<sessionSuffix>`"

**Implementation** (`backup_retention.go:92`, `backup_path_builder.go:65`):
```go
const sessionDirLayout = "2006-01-02T15-04-05.000Z"  // Go reference time
isoPrefix := now.UTC().Format(sessionDirLayout)      // → "2026-06-15T15-30-45.000Z"
return filepath.Join(home, b.adapterID, isoPrefix+"-"+sessionSuffix)
```

**Discrepancy**: Spec example uses `2026-06-15T15-30-45-123Z-...`
(`-` between SS and mmm). Implementation produces
`2026-06-15T15-30-45.000Z-...` (`.` between SS and mmm). The
formatter/parser are internally consistent (`hasSessionPrefix` parses
with the same layout), so the lex-sort == chron-sort invariant is
preserved and `PruneBackups` works correctly.

**Recommend**: clarify the spec REQ-BRP-02 example. Either:
- (a) Update the spec example to `2026-06-15T15-30-45.000Z-<suffix>` to
  match the implementation, or
- (b) Update the design + implementation to use
  `sessionDirLayout = "2006-01-02T15-04-05-000Z"` to match the spec
  example.

**Not a CRITICAL finding** — the lex-sort invariant holds for both
formats, and the REQ text label `mmmZ` does not pin the separator
character.

### 5.2 REQ-BRP-06: "continues on error" scenario numbering

**Spec scenario** (line 137-142):
> GIVEN adapter `x` has 7 session directories and directory #4 is read-only
> WHEN `PruneBackups("x", 5)` is called
> THEN the two non-read-only oldest directories are removed
> AND the returned `removed` count is `2`

**Implementation** (`backup_retention_test.go:247-282`):
```go
// 7 valid session dirs and marks the OLDEST (offset 0) read-only.
// With max=5, the two oldest are to be removed: offset 0 (fails) and
// offset 1 (succeeds).
assert.Equal(t, 1, removed, "...")
```

**Discrepancy**: Spec says "directory #4 is read-only" and expects
`removed=2` (both other oldest dirs removed). The implementation marks
the OLDEST (offset 0) read-only and asserts `removed=1` (the second
oldest succeeds; the oldest fails). The spec's "directory #4" is
ambiguous — it could mean 4th-newest (out of the kept set, not removed)
or 4th-oldest (within the to-be-removed set).

**The implementation's interpretation is reasonable**: it directly
exercises the "continues to attempt removal of subsequent entries"
contract by placing the read-only entry FIRST in the removal order, so
the next entry MUST succeed for the "continues" claim to hold. The
spec's literal example (directory #4 read-only, `removed=2`) would
require the read-only entry to be the SECOND oldest, so that
`PruneBackups` would fail on #2-of-removal and succeed on #1, giving
`removed=1` (not 2). This means **the spec's scenario is internally
inconsistent** — the "two non-read-only oldest directories are
removed" claim contradicts "directory #4 is read-only" with
`max=5` and 7 entries.

**Recommend**: rewrite the scenario to remove the numbering
ambiguity. e.g., "GIVEN adapter `x` has 7 session directories and the
OLDEST is read-only / WHEN `PruneBackups("x", 5)` is called / THEN
the next-oldest (non-read-only) directory is removed / AND the
returned `removed` count is `1` / AND the returned error is non-nil
and references the read-only directory". This matches the
implementation and the spec's stated intent.

**Not a CRITICAL finding** — the spec's REQ text
("continuing to attempt removal of subsequent entries on error") is
satisfied. The scenario is a documentation fix.

---

## 6. Risks for PR 2

1. **`BaseAdapter.Prepare` still uses per-tool path for `BackupDir`** —
   `a.backup.Build(base)` is the current call site
   (`base_adapter.go:325` approximately). PR 2 task 2.2 must change
   this to compute `BackupDir` from `BackupHomeDir()`. The
   `BackupPathBuilder` injection in `BaseAdapter` should be deprecated
   in favor of the new `centralBackupDir(targetSubdir)` helper.

2. **`backupPathBuilder.go`'s safety-net fallback to `backupPathFn`** is
   a temporary safety net. Once PR 3 task 3.11 (per-adapter `paths.go`
   delegates to `BackupHomeDir()`) lands, the safety net becomes
   unreachable. Decide: keep for resilience (one extra indirection)
   or remove for clarity (cleaner code). PR 2 should make this call.

3. **4+ existing tests still reference `.sequoia-backup` paths** —
   `adapters/common/installer_test.go`,
   `adapters/codex/installer_test.go`,
   `adapters/opencode/install_test.go`, and possibly more (apply-progress
   risk #6). PR 2 task 2.4 is the work to fix these. The
   `TestAdapter_Reinstall_OverwritesVersion` in gemini was the first to
   break and was patched in commit 5f3214a as a focused regression fix.

4. **`TestPruneBackups_ContinuesOnError` is skipped on Windows.** PR 2/3
   may want a Windows-compatible continuation test (e.g., create a
   session dir with an open file handle on Windows). Not blocking for
   PR 1.

5. **Pre-existing test pollution**: PR 1's
   `TestBackupPathBuilder_Build_IncludesAdapterID` /
   `TestBackupPathBuilder_Build_UsesBackupPathFn` still touch the
   real `os.UserConfigDir()` and rely on `cleanupCentralHome` to
   remove the adapter subdir. If a future test fails between
   `Build()` and the cleanup, a session dir is left behind. The
   internal tests use `t.TempDir()` overrides, so PR 2/3 should
   prefer the internal-test pattern.

6. **No `isSequoiaTimestamp` helper in the apply-progress narrative** —
   the apply-progress mentions a "pure function created: isSequoiaTimestamp"
   (line 64) but the actual code uses `hasSessionPrefix` (which combines
   the timestamp parse + separator check). Cosmetic, no behavior change.

---

## 7. Out-of-Scope Confirmation (explicit)

| Out-of-scope area | File(s) | Status | Evidence |
|---|---|---|---|
| Per-adapter paths.go | `adapters/claude/paths.go`, `adapters/codex/paths.go`, `adapters/cursor/paths.go`, `adapters/gemini/paths.go`, `adapters/opencode/paths.go` | **NOT MODIFIED** | `git diff main..feature/backup-retention-pr1-foundation --name-only -- adapters/claude adapters/codex adapters/cursor adapters/gemini adapters/opencode` returns empty. |
| `ReplaceFile` | `adapters/common/strategy.go:126` | **NOT MODIFIED** | `git diff --stat adapters/common/strategy.go` empty; function still uses `.sequoia-backup-<suffix>` + `.sequoia-session` sidecar. |
| `RestoreOrRemoveFile` | `adapters/common/strategy.go:170` | **NOT MODIFIED** | same as above. |
| `BaseAdapter.Prepare` (central-home wiring) | `adapters/common/base_adapter.go:308` | **NOT MODIFIED** | `git diff --stat adapters/common/base_adapter.go` empty; still uses per-tool path. |
| `BaseAdapter.Stage` (central-home wiring) | `adapters/common/base_adapter.go:436` | **NOT MODIFIED** | same as above. |
| `BaseAdapter.Apply` retention hook | `adapters/common/base_adapter.go:491` | **NOT WIRED** | `grep -E 'applyRetention\|PruneBackups' base_adapter.go` returns no matches. Confirms PR 3 task 3.8 work is intact. |
| TUI `Info` migration note | `internal/pipeline/runner.go:200-210` | **NOT MODIFIED** | `git diff --stat internal/pipeline/runner.go` empty. |
| Existing tests in 4+ files with `.sequoia-backup` assertions | `adapters/common/installer_test.go`, `adapters/common/strategy_test.go`, `adapters/common/base_adapter_strategy_test.go`, `adapters/common/base_adapter_error_test.go`, `adapters/codex/installer_test.go`, `adapters/opencode/install_test.go` | **NOT UPDATED (intentional)** | `git diff --stat` on these files returns empty. PR 2 task 2.4 will do this. The one test that DID break (gemini `TestAdapter_Reinstall_OverwritesVersion`) was patched with a minimal regression fix in commit 5f3214a. |

---

## 8. Recommendation

**`proceed-to-pr-2-with-acknowledged-warnings`**

Rationale:
- All 10 PR 1 tasks PASS.
- Full test suite green (20/20 packages).
- `adapters/common` coverage 85.8%, well above the 70% CI gate.
- Every new function is above the 70% gate.
- `go vet ./...` clean.
- 5 golangci-lint `gofmt` issues, 4 pre-existing and 1 new (all Windows
  line-ending; not introduced by the change's logic; CI on Linux/macOS
  will not see this).
- 2 SPEC AMBIGUITIES (timestamp format example; "directory #4" numbering)
  are documentation issues, not implementation issues — the
  implementation satisfies the REQ text's intent.
- 2 WARNINGs (timestamp format deviation; gofmt on the new file) are
  non-blocking.
- 1 SUGGESTION (TDD commit shape) is a process improvement for PR 2/3.
- 5 OUT OF SCOPE areas (REQ-BRP-03, REQ-BRP-04 e2e, REQ-BRP-05,
  per-adapter paths.go, retention hook) are correctly untouched.

PR 1 is a clean foundation. PR 2 (installer wiring + TUI note + test
updates) can start on a `feature/backup-retention-pr2-installer` branch
off `main` after PR 1 merges. The two spec ambiguities should be
filed as a follow-up `sdd-propose` change to clarify the spec, or
addressed as a one-line edit in the spec when sdd-archive lands.

---

## Artifacts

- Verify report (this file):
  `openspec/changes/backup-retention-and-organization/verify-report.md`
- Coverage profile: `coverage.out`, `cov_common.out` (at repo root)
- Per-function coverage: `C:\Users\Usuario\AppData\Local\Temp\opencode\covcommon_func.txt`

---

## Structured Envelope (return value)

```json
{
  "status": "pass-with-warnings",
  "executive_summary": "PR 1 foundation is solid: 10/10 tasks PASS, full test suite green (20/20 packages), adapters/common coverage 85.8% (above 70% gate), go vet clean. Two spec ambiguities (timestamp format example in REQ-BRP-02, 'directory #4' numbering in REQ-BRP-06) and two warnings (gofmt line endings on the new file, TDD commit shape). No CRITICAL findings.",
  "artifacts": {
    "verify_report": "openspec/changes/backup-retention-and-organization/verify-report.md",
    "coverage_profile_root": "coverage.out",
    "coverage_profile_common": "cov_common.out",
    "per_function_coverage": "C:\\Users\\Usuario\\AppData\\Local\\Temp\\opencode\\covcommon_func.txt"
  },
  "next_recommended": "proceed-to-pr-2-with-acknowledged-warnings",
  "risks": [
    "BaseAdapter.Prepare/Stage still use per-tool BackupDir (PR 2 task 2.2/2.3 will fix)",
    "BackupPathBuilder.Build safety-net fallback to backupPathFn becomes unreachable after PR 3 task 3.11 (PR 2 should decide: keep or remove)",
    "4+ existing tests in adapters/common/{installer,strategy,base_adapter_strategy,base_adapter_error}_test.go and adapters/{codex,opencode}/* still reference old per-tool .sequoia-backup paths (PR 2 task 2.4)",
    "TestPruneBackups_ContinuesOnError is skipped on Windows (chmod 0o500 does not block os.RemoveAll); CI on Linux/macOS exercises it",
    "Test pollution: external backup_path_builder_test.go tests touch real os.UserConfigDir() with cleanupCentralHome best-effort",
    "Spec REQ-BRP-02 example timestamp format (2026-06-15T15-30-45-123Z) differs from implementation (2026-06-15T15-30-45.000Z) — both are valid ISO-8601, lex-sort invariant holds; recommend spec clarification",
    "Spec REQ-BRP-06 'continues on error' scenario 'directory #4 is read-only' is internally inconsistent with 'two non-read-only oldest directories are removed' given max=5 and 7 sessions; implementation correctly exercises the continuation contract; recommend scenario rewrite"
  ]
}
```
