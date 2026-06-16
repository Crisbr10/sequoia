# Verify Report: backup-retention-and-organization — PR 3a (Manifest + Safety-Net)

> **Verifier**: `sdd-verify` sub-agent (PR 3a only, do NOT proceed to PR 3b or PR 3.5)
> **Branch under verification**: `feature/backup-retention-pr3a-manifest`
> **Commits ahead of main**: 5 (`a7daa17`..`6d21d8d`)
> **Strict TDD**: ACTIVE
> **Run timestamp**: 2026-06-16, Windows runner (PowerShell 7)
> **Final status**: **FAIL** (1 CRITICAL finding — design schema divergence)

---

## 1. Executive Summary

PR 3a introduces the per-session `manifest.json` type system, the
safety-net removal in `BackupPathBuilder.Build`, and the 5 per-adapter
`paths.go` legacy docstrings — all without the `applyRetention` hook
(which is PR 3b) and without the `ReplaceFile`/`RestoreOrRemoveFile`
manifest wiring (which is PR 3.5). The retention cap is **not yet
active**; the central home exists and the manifest format is
established as the foundation for PR 3.5 to consume.

The PR is well-tested (3 consecutive runs green, 20/20 packages pass,
`go vet` clean, all 7 new tests pass, `adapters/common` package
coverage 85.0%, new `manifest.go` 71.9% file-level, modified
`backup_path_builder.go` 100% per-file, per-function on all new
surfaces ≥ 70% except POSIX-only `removeSessionDir` which is correctly
skipped on Windows). The out-of-scope invariants are all respected
(`strategy.go` untouched, no `applyRetention` hook, no
`base_adapter_retention_test.go`).

**One critical issue** was discovered during the schema check:
the implementation's `manifest` struct is **missing the top-level
`created_at` field** that the design's locked schema specifies
(`{version, created_at, entries:[{version, original_path, suffix, created_at, adapter_id}]}`).
The implementation's `manifest` struct is `{version, entries}` — only
2 of the 3 design-specified top-level fields. Per the orchestrator's
explicit instruction ("If the actual schema diverges … classify as
CRITICAL — PR 3.5 will be written against this schema and divergence
now means rework"), this is a CRITICAL finding.

The fix is **trivial** (1 line in `manifest.go` + 1 small test) and
the spec's REQ-BRP-03 scenarios do not actually require a top-level
`created_at` (only `original_path` and `suffix` are named in the
scenarios). The divergence is design-vs-implementation, not
spec-vs-implementation. Two options to resolve:
- (a) **Fix in PR 3a** (1 line + 1 test) — recommended; locks the
  schema in the PR that establishes it; saves PR 3.5 from the rework.
- (b) **Fix in PR 3.5** as part of the manifest wiring — acceptable
  but the divergence will be visible in the diff history.

**Finding tally**:
- CRITICAL: 1 (manifest struct missing top-level `created_at`)
- WARNING: 0
- SUGGESTION: 2 (TDD commit shape — carried from PR 1+2; test
  pollution above cap — carried from PR 1+2)
- SPEC AMBIGUITY: 2 (timestamp format example — carried from PR 1+2;
  "directory #4" numbering — carried from PR 1+2)
- OUT OF SCOPE (intentional non-failures): 5 (`strategy.go`
  untouched, `applyRetention` not added, `base_adapter_retention_test.go`
  not present, retention cap not active, `strategy_central_test.go`
  not present)

**One-line verdict**: Solid, focused slice. CRITICAL schema divergence
is the only blocker; the fix is 1 line + 1 test and can land in
PR 3a (recommended) or PR 3.5 (acceptable). The 184-line overage is
acceptable; the test/prod split is healthy.

---

## 2. Task Verification

PR 3a's apply-progress §PR 3a Task Status lists 4 in-scope task
groups: 3.1, 3.2, 3.11 (safety-net), manifest error-paths (house-
keeping), gofmt (housekeeping). Tasks 3.3-3.10 are deferred to
PR 3.5 or PR 3b.

| Task | Claim | Result | Evidence |
|---|---|---|---|
| **3.1 RED** | `manifestEntry` JSON round-trip test | **PASS** | `TestManifestEntry_JSONRoundTrip` in `adapters/common/manifest_test.go:31-59`. Marshals a fixture entry, unmarshals it, asserts all 5 fields round-trip (Version, OriginalPath, Suffix, CreatedAt, AdapterID). PASS on this runner. |
| **3.2 GREEN** | `manifestEntry` + `manifest` types + helpers | **PASS (with CRITICAL caveat — see §4.8)** | `manifestEntry` struct (5 fields, all matching design) defined at `manifest.go:28-34`. `manifest` struct (2 of 3 design-specified fields — missing top-level `created_at`; see CRITICAL §4.8) defined at `manifest.go:39-42`. Helpers `newEmptyManifest`, `readManifest`, `writeManifest`, `appendManifestEntry`, `removeSessionDir` all present. JSON round-trip works. |
| **3.11 (partial) GREEN** | Safety-net no longer consults `backupPathFn` | **PASS** | `backup_path_builder.go:58-74`: safety-net now hard-codes `<base>/.sequoia-backup/<adapterID>/<suffix>` (was `b.backupPathFn(base) + "-" + b.adapterID + "-" + sessionSuffix` per the diff). Docstring at lines 50-57 explains the rationale. New test `TestBackupPathBuilder_Build_SafetyNetSkipsBackupPathFn` at `backup_path_builder_internal_test.go:139-168` forces `BackupHomeDir()` to fail (blocking file as UserConfigDir parent), asserts the sentinel `backupPathFn` is NOT in the result. PASS. `Build()` coverage rose from 85.7% (PR 1+2) to **100%**. |
| **3.11 (partial) GREEN** | 5 `paths.go` docstrings updated | **PASS** | All 5 files (`adapters/claude/paths.go`, `adapters/codex/paths.go`, `adapters/cursor/paths.go`, `adapters/gemini/paths.go`, `adapters/opencode/paths.go`) have the new 11-line docstring explaining the function is "no longer consulted on the happy path", "retained for backwards compatibility", and "safety-net in BackupPathBuilder.Build uses a hard-coded … shape and does not call this function". Verified by reading all 5. |
| **error-paths** | `TestManifest_ReadCorruptReturnsError`, `TestWriteManifest_MkdirFailureReturnsError`, `TestRemoveSessionDir_FailureReturnsError` | **PASS** | `manifest_test.go:151-208`. Corrupt JSON test: writes `not-valid-json{{`, asserts error contains "manifest: parse" context. Mkdir failure test: creates a regular file as parent, asserts `writeManifest` errors with "manifest: mkdir" context. `removeSessionDir` failure test: same trick, **skipped on Windows** (`runtime.GOOS == "windows"`) because `os.RemoveAll` on a missing path is a no-op there. POSIX CI runners (ubuntu-latest, macos-latest) will exercise the path. All 3 tests pass on this Windows runner. |
| **gofmt** | gofmt alignment fix | **PASS** | Commit `b9f8203` adjusts the struct tag alignment in `manifest.go` (1 ins, 1 del). `gofmt -l` on all 10 PR 3a files returns **empty** (clean). |
| **3.7, 3.8, 3.9** | Retention hook | **DEFERRED to PR 3b (out of scope)** | Not in PR 3a. `git diff main..HEAD -- adapters/common/base_adapter.go | grep applyRetention` is empty. **CONFIRMED.** |
| **3.3, 3.4, 3.5, 3.6** | `ReplaceFile`/`RestoreOrRemoveFile` central-home + manifest | **DEFERRED to PR 3.5 (out of scope)** | Not in PR 3a. `git diff main..HEAD -- adapters/common/strategy.go` is empty. `git ls-files | grep strategy_central_test` returns empty. **CONFIRMED.** |
| **3.10** | Manifest helper consolidation | **DEFERRED to PR 3.5 (out of scope)** | The original `ce64e25` commit depended on `108414c` (PR 3.5 work); per the apply-progress note, this was correctly deferred. **CONFIRMED.** |
| **X.2 (partial)** | apply-progress.md commit | **PASS** | Commit `6d21d8d` adds the PR 3a section to `apply-progress.md` (+143 lines). This is documentation, reviewable separately from the code changes. |

**Total**: 6/6 in-scope task groups PASS (with the CRITICAL caveat on
the manifest struct schema). 5/5 deferred tasks correctly
out-of-scope.

### 2.1 SUGGESTION (non-blocking, carried from PR 1 + PR 2) — TDD commit shape

The apply-progress claims strict TDD with separate RED/GREEN cycles,
but the git history shows each RED+GREEN pair landed in a single
commit (e.g., `a7daa17` adds 132 production lines + 142 test lines
together; `0d91e9f` adds 19 production + 6 production-comment + 40
test lines together). The end-state is correct and the tests do
exercise the behavior, but a strict TDD audit would want to see the
test commit land first with a build-fail and the implementation commit
land second with a build-pass. The PR 1 verify report §2.1 and PR 2
verify report §2.1 noted the same; the apply agent acknowledged the
TDD commit shape in the apply-progress ("End-state is correct and the
tests do exercise the behavior"). Not a blocker for PR 3a; a future
`sdd-apply` should land test-only commits separately.

---

## 3. Spec Compliance

### REQ-BRP-01 — Centralized backup root — **PASS (carried from PR 1)**

PR 3a does not modify `BackupHomeDir()`. Verified by
`git diff main..HEAD -- adapters/common/backup_retention.go` →
empty. The PR 1 helpers continue to satisfy this REQ.

### REQ-BRP-02 — Per-adapter organization — **PASS (carried from PR 1+2)**

PR 3a does not modify `CentralBackupDir` or `BackupPathBuilder.Build`'s
happy path. The new safety-net still uses the
`<base>/.sequoia-backup/<adapterID>/<suffix>` shape that the spec
calls out as the legacy fallback.

The 2 spec ambiguities from PR 1 carry over (still open):
1. Spec example timestamp `2026-06-15T15-30-45-123Z-<suffix>` uses
   `-` between SS and mmm; implementation uses
   `2006-01-02T15-04-05.000Z` (`.` between SS and mmm). Both are valid
   ISO-8601; the formatter/parser are internally consistent. Not a
   code issue. See SPEC AMBIGUITY §6.1.
2. REQ-BRP-06 "directory #4 is read-only" scenario is internally
   inconsistent. Implementation matches the spec's *intent*
   (continues-on-error) with `removed=1`. Not a code issue. See
   SPEC AMBIGUITY §6.2.

### REQ-BRP-03 — File-replace backup storage — **PARTIAL PASS (with CRITICAL)**

The spec REQ-BRP-03 scenarios cover:
1. "ReplaceFile writes the backup to the session directory" (Scenario 1)
2. "RestoreOrRemoveFile reads from the session directory via the manifest" (Scenario 2)

PR 3a establishes the **manifest format** (per-entry types and
helpers) but does NOT yet wire `ReplaceFile`/`RestoreOrRemoveFile` to
read/write the manifest. That is PR 3.5 (deferred correctly).

**In-scope verification (PR 3a)**: the manifest types are defined and
JSON round-trip works. The schema matches the design's intent for the
`manifestEntry` shape (`{version, original_path, suffix, created_at,
adapter_id}` — all 5 fields present).

**CRITICAL finding (see §4.8)**: the `manifest` struct is missing the
top-level `created_at` field that the design specifies. The
implementation's manifest is `{version, entries}` (2 fields); the
design's is `{version, created_at, entries}` (3 fields). This is a
**design-vs-implementation divergence** that the orchestrator's
schema-check rule explicitly classifies as CRITICAL.

The spec REQ-BRP-03 scenarios do NOT mention a top-level `created_at`
(they only assert `original_path` and `suffix`), so the spec is
**silent** on this field. The divergence is design-only.

### REQ-BRP-04 — Retention policy of 5 per adapter — **OUT OF PR 3a SCOPE (PR 3b)**

`applyRetention` is not yet hooked in `BaseAdapter.Apply`. Verified:
- `git diff main..HEAD -- adapters/common/base_adapter.go | grep applyRetention` → empty
- `git diff main..HEAD -- adapters/common/base_adapter.go | grep 'addWarning.*retention'` → empty
- `git ls-files | grep base_adapter_retention_test` → empty

The retention cap is **not active** after PR 3a lands. The helper
(`PruneBackups`, `DefaultMaxBackupsPerAdapter=5`) exists from PR 1
and is fully covered; the wiring is PR 3b work (deferred correctly).

### REQ-BRP-05 — Migration of old scattered backups (NOT performed) — **OUT OF PR 3a SCOPE (PR 2 work)**

PR 3a does not touch the TUI `Info` migration note (PR 2 task 2.5/2.6,
already merged) or the per-tool `paths.go` content (only docstrings
updated; the `filepath.Join(base, ".sequoia-backup")` return value is
unchanged from PR 2). No code path in PR 3a touches pre-existing
`.sequoia-backup-*` files.

### REQ-BRP-06 — Path resolution and pruning helpers — **PASS (carried from PR 1)**

PR 3a does not modify the PR 1 helpers. `BackupHomeDir` and
`PruneBackups` continue to work as before. The new safety-net
hard-coding in `BackupPathBuilder.Build` is a defensive measure (not
a new helper per the spec); it does not change `BackupHomeDir` or
`PruneBackups` behavior.

### REQ-BRP-07 — Test surface (strict TDD) — **PASS**

In-scope for PR 3a:
- `manifestEntry` JSON round-trip: ✅ `TestManifestEntry_JSONRoundTrip`
- `appendManifestEntry` + `readManifest` round-trip: ✅
  `TestManifest_AppendAndRead`
- `appendManifestEntry` preserves existing entries: ✅
  `TestManifest_AppendPreservesExistingEntries`
- `readManifest` on missing file returns empty: ✅
  `TestManifest_ReadMissingReturnsEmpty`
- `readManifest` on corrupt JSON returns error: ✅
  `TestManifest_ReadCorruptReturnsError`
- `writeManifest` mkdir failure returns error: ✅
  `TestWriteManifest_MkdirFailureReturnsError`
- `removeSessionDir` failure returns error: ✅
  `TestRemoveSessionDir_FailureReturnsError` (POSIX-only, skipped on
  Windows; correct platform behavior)
- Safety-net ignores `backupPathFn`: ✅
  `TestBackupPathBuilder_Build_SafetyNetSkipsBackupPathFn`

Out-of-scope (correctly deferred to PR 3.5):
- Centralized `ReplaceFile`/`RestoreOrRemoveFile` round-trip — **PR 3.5**
- E2E install leaves at most 5 session dirs — **PR 3b**
- Existing 4+ tests updated to central path — **PR 2 (done)**

**PASS** for PR 3a scope.

---

## 4. Independent Re-execution (adversarial)

### 4.1 Full project test run (3 consecutive runs)

**Run #1** (with coverage):
```
ok  github.com/Crisbr10/sequoia/adapters          1.275s  coverage: 96.4% of statements
ok  github.com/Crisbr10/sequoia/adapters/claude   3.079s  coverage: 80.0% of statements
ok  github.com/Crisbr10/sequoia/adapters/codex    3.377s  coverage: 78.1% of statements
ok  github.com/Crisbr10/sequoia/adapters/common   3.534s  coverage: 85.0% of statements
ok  github.com/Crisbr10/sequoia/adapters/cursor   2.081s  coverage: 89.7% of statements
ok  github.com/Crisbr10/sequoia/adapters/gemini   3.046s  coverage: 86.7% of statements
ok  github.com/Crisbr10/sequoia/adapters/opencode 2.972s  coverage: 81.2% of statements
ok  github.com/Crisbr10/sequoia/adapters/testutil 1.243s  coverage: 90.5% of statements
ok  github.com/Crisbr10/sequoia/cmd/sequoia       5.893s  coverage: 77.5% of statements
ok  github.com/Crisbr10/sequoia/internal/app      1.358s  coverage: 87.1% of statements
ok  github.com/Crisbr10/sequoia/internal/codegraph 1.186s  coverage: 82.1% of statements
ok  github.com/Crisbr10/sequoia/internal/model    1.214s  coverage: [no statements]
ok  github.com/Crisbr10/sequoia/internal/pipeline 1.752s  coverage: 78.6% of statements
ok  github.com/Crisbr10/sequoia/internal/tui      1.294s  coverage: 92.6% of statements
ok  github.com/Crisbr10/sequoia/internal/tui/screens 1.424s  coverage: 88.0% of statements
ok  github.com/Crisbr10/sequoia/internal/tui/styles 1.529s  coverage: 100.0% of statements
ok  github.com/Crisbr10/sequoia/plugin            1.645s  coverage: 94.1% of statements
ok  github.com/Crisbr10/sequoia/plugin/example    1.354s  coverage: 100.0% of statements
ok  github.com/Crisbr10/sequoia/scripts           1.419s  coverage: [no statements]
```

**Result: 20/20 packages PASS** (1 `[no test files]` and 2
`[no statements]` are normal and not failures).

**Runs #2, #3** (no coverage, faster): all 3 runs green. No FAIL,
no `--- FAIL`, no panic (`Select-String -Pattern 'FAIL|^---|panic'`
across the 3 logs returned 0 matches). The test suite is
**stable** — no flakiness detected.

### 4.2 Per-function coverage on the new surfaces

```
adapters/common/backup_path_builder.go:31:  NewBackupPathBuilder   100.0%
adapters/common/backup_path_builder.go:58:  Build                  100.0%   <-- up from 85.7% (PR 1+2)
adapters/common/manifest.go:48:             newEmptyManifest       100.0%
adapters/common/manifest.go:63:             readManifest            78.6%
adapters/common/manifest.go:89:             writeManifest           77.8%
adapters/common/manifest.go:113:            appendManifestEntry     80.0%
adapters/common/manifest.go:127:            removeSessionDir         0.0%   <-- POSIX-only, test skipped on Windows
```

**Per-file aggregate coverage** (computed from `coverage.out`):

| File | Coverage | Stmts (covered/total) | ≥ 70% gate |
|---|---|---|---|
| `adapters/common/manifest.go` (NEW) | **71.9%** | 23/32 | ✅ |
| `adapters/common/backup_path_builder.go` (MOD) | **100.0%** | 8/8 | ✅ |
| `adapters/common/backup_retention.go` (PR 1, unchanged) | 80.0% | 36/45 | ✅ |
| `adapters/common/base_adapter.go` (PR 2, unchanged) | 85.0% | 204/240 | ✅ |
| `adapters/common/strategy.go` (out of scope, unchanged) | 94.2% | 97/103 | ✅ |
| `adapters/common` (full package) | **85.0%** | 509/599 | ✅ |

**`removeSessionDir` at 0% is correct**: the test
`TestRemoveSessionDir_FailureReturnsError` is platform-skipped on
Windows (`runtime.GOOS == "windows"`) because `os.RemoveAll` on a
missing path is a no-op there. The CI matrix on `ubuntu-latest` and
`macos-latest` will exercise this path. The apply agent documented
this correctly. POSIX CI coverage should hit 100% for this function.

### 4.3 `go vet ./...`

Clean. No output, exit 0.

### 4.4 `gofmt -l` on PR 3a files

Command:
`gofmt -l adapters/common/manifest.go adapters/common/manifest_test.go adapters/common/backup_path_builder.go adapters/common/backup_path_builder_internal_test.go adapters/claude/paths.go adapters/codex/paths.go adapters/cursor/paths.go adapters/gemini/paths.go adapters/opencode/paths.go`

Result: **empty** (all 10 files clean). The apply agent's
gofmt alignment fix (commit `b9f8203`) is effective.

### 4.5 Out-of-scope confirmation (re-verified)

**`strategy.go` diff** (ReplaceFile/RestoreOrRemoveFile territory — must be empty):
```
$ git diff main..HEAD -- adapters/common/strategy.go | Measure-Object -Line
0
```
0 lines of diff. **CONFIRMED** — out of scope not touched.

**`base_adapter.go` retention grep** (applyRetention hook — must be empty):
```
$ git diff main..HEAD -- adapters/common/base_adapter.go | grep -E '(applyRetention|addWarning.*retention)'
(no output)
```
**CONFIRMED** — retention hook not added in PR 3a.

**PR 3b/3.5 test files** (must not exist yet):
```
$ git ls-files | grep -E '(base_adapter_retention_test|strategy_central_test)'
(no output)
```
**CONFIRMED** — no PR 3b/3.5 test files present.

**Production callers of manifest functions** (must be in manifest.go + manifest_test.go only):
```
$ git grep -E 'manifestEntry|newEmptyManifest|readManifest|writeManifest|appendManifestEntry|removeSessionDir' -- 'adapters/**' 'internal/**'
adapters/common/manifest.go: type manifestEntry struct { ... }
adapters/common/manifest.go: ... appendManifestEntry(sessionDir string, entry manifestEntry) error { ... }
adapters/common/manifest.go: ... readManifest / writeManifest / removeSessionDir / newEmptyManifest (all internal)
adapters/common/manifest_test.go: ... test fixtures using manifestEntry / newEmptyManifest / etc.
```

All matches are in `adapters/common/manifest.go` (declarations +
self-references) and `adapters/common/manifest_test.go` (tests).
**CONFIRMED** — no production caller. PR 3.5 will introduce
production callers in `strategy.go`.

**Retention cap inactive in production** (must be no calls to `PruneBackups`):
```
$ git diff main..HEAD -- 'adapters/**' 'internal/**' | Select-String -Pattern 'PruneBackups|applyRetention'
(only one match, in a manifest.go COMMENT explaining the design intent:
 "+// does not accumulate (PruneBackups would clean it up later, but")
```
The single match is a **comment** in `manifest.go:124` (the `removeSessionDir` docstring explaining the spec's intent). It is NOT a call to `PruneBackups` — no production code path invokes the retention helper. **CONFIRMED** — retention cap not active.

### 4.6 Test pollution (carried from PR 1 + PR 2)

Not re-measured on this verification (PR 2 verify report §4.6 already
documented 16 dirs per adapter in `$APPDATA/sequoia/backups/err-test/`
and `opencode/`, above the 5-dir cap). The pollution persists
because the `applyRetention` hook is PR 3b work; PR 3a does not
change the test pollution state. **WARNING carried from PR 2.**

### 4.7 `paths.go` docstring check (5 files)

I read all 5 `adapters/<tool>/paths.go` files end-to-end. Each has
the new 11-line docstring on `backupPath(base string) string`
explaining:
- The function is the **legacy per-tool temp backup dir** used by the
  safety-net fallback.
- The **PR 3 main flow moved to the central backup home** so this
  function is no longer consulted on the happy path.
- It is **retained for backwards compatibility** with any external
  callers and as a **documentary anchor** for the legacy per-tool
  layout.
- The **PR 3 safety-net in BackupPathBuilder.Build uses a hard-coded
  `<base>/.sequoia-backup/<adapterID>/<suffix>` shape** and does not
  call this function — see `backup_path_builder.go` for details.

**All 5 docstrings are accurate, consistent, and complete.** The
docstrings correctly describe the post-PR-3a state.

### 4.8 Manifest JSON schema check (mandatory)

**Design's locked schema** (decision 2, design.md:33):
```
{version, created_at, entries: [{version, original_path, suffix, created_at, adapter_id}]}
```

**Implementation's schema** (extracted from `adapters/common/manifest.go`):

```go
// manifestEntry (lines 28-34):
type manifestEntry struct {
    Version      string    `json:"version"`
    OriginalPath string    `json:"original_path"`
    Suffix       string    `json:"suffix"`
    CreatedAt    time.Time `json:"created_at"`
    AdapterID    string    `json:"adapter_id"`
}

// manifest (lines 39-42):
type manifest struct {
    Version string          `json:"version"`
    Entries []manifestEntry `json:"entries"`
}
```

**JSON tags side-by-side**:

| Design field | Implementation field | Present? |
|---|---|---|
| `manifest.version` | `manifest.Version` (`json:"version"`) | ✅ |
| **`manifest.created_at`** | **(MISSING)** | ❌ **CRITICAL** |
| `manifest.entries` | `manifest.Entries` (`json:"entries"`) | ✅ |
| `manifestEntry.version` | `manifestEntry.Version` (`json:"version"`) | ✅ |
| `manifestEntry.original_path` | `manifestEntry.OriginalPath` (`json:"original_path"`) | ✅ |
| `manifestEntry.suffix` | `manifestEntry.Suffix` (`json:"suffix"`) | ✅ |
| `manifestEntry.created_at` | `manifestEntry.CreatedAt` (`json:"created_at"`) | ✅ |
| `manifestEntry.adapter_id` | `manifestEntry.AdapterID` (`json:"adapter_id"`) | ✅ |

**Divergence summary**: 1 of 8 fields missing — the top-level
`created_at` on the `manifest` struct.

**Spec impact**: The spec REQ-BRP-03 scenarios do NOT mention a
top-level `created_at` (only `original_path` and `suffix` are named
explicitly). The divergence is design-vs-implementation, not
spec-vs-implementation. The spec is silent on this field.

**Why it matters**: The orchestrator's task description explicitly
states: "PR 3.5 will be written against this schema and divergence
now means rework." PR 3.5 will introduce the `ReplaceFile` calls
to `appendManifestEntry` and the `RestoreOrRemoveFile` calls to
`readManifest`. If the schema is finalized at PR 3a with the wrong
shape, the PR 3.5 work has to either (a) update the schema before
consuming it, or (b) write code that produces a schema the
implementation can't read back.

**Classification**: **CRITICAL** per the orchestrator's explicit
schema-check rule ("missing field = CRITICAL — PR 3.5 will be
written against this schema and divergence now means rework").

**Recommended fix** (trivial, 1 line + 1 test):

```go
// In adapters/common/manifest.go, change lines 39-42 from:
type manifest struct {
    Version string          `json:"version"`
    Entries []manifestEntry `json:"entries"`
}

// To:
type manifest struct {
    Version   string          `json:"version"`
    CreatedAt time.Time       `json:"created_at"`
    Entries   []manifestEntry `json:"entries"`
}
```

And update `newEmptyManifest` to initialize `CreatedAt: time.Now().UTC()`
(or omit it, so the zero value is fine; but asserting it is non-zero in
a new round-trip test would be belt-and-suspenders).

Then update `readManifest` to default a zero `CreatedAt` to the
current time if the JSON has an empty/missing value (or to the
session directory's mtime — same idea, defensive).

The new test would be:

```go
func TestManifest_TopLevelCreatedAt_RoundTrip(t *testing.T) {
    t.Parallel()
    dir := t.TempDir()
    now := time.Date(2026, 6, 15, 15, 30, 45, 0, time.UTC)
    m := newEmptyManifest()
    m.CreatedAt = now
    require.NoError(t, writeManifest(dir, m))
    got, err := readManifest(dir)
    require.NoError(t, err)
    assert.True(t, got.CreatedAt.Equal(now))
}
```

This is **2 file changes, ~6 lines, ~25 lines of test**. Easy fix.

**Note**: The fix could alternatively be applied in **PR 3.5** as
part of the manifest wiring (before `ReplaceFile` starts producing
manifests). This would not block PR 3a from merging, but the
divergence would be visible in the diff history. The orchestrator's
recommendation is to fix in PR 3a to "lock the schema in the PR
that establishes it" (saves PR 3.5 from rework).

### 4.9 TDD commit shape SUGGESTION (carried from PR 1 + PR 2)

Same as PR 1 verify §2.1 and PR 2 verify §2.1: each RED+GREEN pair
landed in a single commit. Examples:
- `a7daa17` adds 132 production lines + 142 test lines together
  (claimed to be the RED + GREEN for tasks 3.1 + 3.2 in one commit)
- `0d91e9f` adds 13 production + 6 production-comment + 40 test lines
  together (claimed to be the safety-net + 5 paths.go docstrings +
  test in one commit)

The end-state is correct and the tests do exercise the behavior
(end-to-end). Not a blocker for PR 3a. A future `sdd-apply` should
land test-only commits separately.

### 4.10 Test pollution WARNING (carried from PR 1 + PR 2)

The pollution in `$APPDATA/sequoia/backups/err-test/` and `opencode/`
(16+ session dirs per adapter) persists. The pollution is bounded
only by the 5-backup retention cap, which is NOT active after PR 3a
(the `applyRetention` hook is PR 3b work). PR 3a does not change the
test pollution state. The retention cap will be enforced when PR 3b
merges. **WARNING carried from PR 1 + PR 2** (not a PR 3a-specific
issue).

---

## 5. Spec Ambiguities (carried over from PR 1 + PR 2)

### 5.1 REQ-BRP-02: Session dir timestamp format — example vs implementation

**Status**: Unchanged from PR 1. The implementation uses
`2006-01-02T15-04-05.000Z` (`.` between SS and mmm); the spec
example uses `-` between SS and mmm. Both are valid ISO-8601; the
formatter/parser are internally consistent. The lex-sort ==
chron-sort invariant is preserved.

**PR 3a relevance**: PR 3a does not modify the session dir
formatter. The same carryover applies.

**Recommend**: Spec clarification in `sdd-archive` or a follow-up
`sdd-propose` change. Not a code issue.

### 5.2 REQ-BRP-06: "continues on error" scenario numbering

**Status**: Unchanged from PR 1. The spec's "directory #4 is
read-only, expect removed=2" is internally inconsistent with
max=5 and 7 entries. The implementation correctly tests the
"continues-on-error" contract with `removed=1` (oldest read-only,
next-oldest succeeds).

**PR 3a relevance**: None — PR 3a does not touch `PruneBackups`.

**Recommend**: Scenario rewrite in spec at `sdd-archive` time.

---

## 6. Risks for PR 3b

1. **Manifest struct missing top-level `created_at`** (CRITICAL
   finding) — if the divergence is not fixed in PR 3a, PR 3b (and
   PR 3.5) will need to either fix the schema or work around it.
   **Recommend**: fix in PR 3a (1 line + 1 test).

2. **Test pollution from PR 1 + PR 2 runs persists** — 16+ dirs
   per adapter under `$APPDATA/sequoia/backups/err-test/` and
   `opencode/`. PR 3b task 3.8 (`applyRetention` hook) will bound
   this to 5. PR 3b verification should pre-clean (or expect the
   cap to engage) to get a clean signal from the "6 installs → 5
   dirs" E2E test.

3. **`removeSessionDir` 0% coverage on Windows** — the test is
   correctly platform-skipped (`runtime.GOOS == "windows"`). The
   POSIX CI runners will exercise this path. The 70% file-level
   gate is satisfied (71.9% file-level), so this is not blocking.

4. **`BackupPathBuilder.Build` 100% coverage** (up from 85.7% in
   PR 2) — the safety-net test in PR 3a exercises the fallback
   branch, so coverage rose. After PR 3.5, the safety-net becomes
   dead code (the 5 `paths.go` `backupPath` functions are
   documentary only). Recommend: keep the test, but consider
   removing the safety-net code in PR 3.5 to simplify
   (the docstrings would then need to be removed too).

5. **The 5 `paths.go` docstrings describe the function as "no
   longer consulted"** — this is the post-PR-3a truth. If PR 3.5
   removes the safety-net (decision pending), the docstrings
   would also need to be updated (to "deprecated — see
   `CentralBackupDir`").

6. **TUI `Info` does NOT yet mention retention count** — the spec
   doesn't require it (the Info message is for pre-existing
   scattered backups, not retention). PR 3b may want to add
   "5 most recent kept; N removed" to the Info message after
   `applyRetention` runs — confirm with product before adding.

7. **`SetLastBackupDir` has 0% direct coverage** (carried from
   PR 2) — tooling artifact; the call IS exercised via
   `Codex.Install` in `installer_internal_test.go`. Trivial
   5-line direct test if it ever matters.

8. **CRLF line endings on Windows-edited files** (carried from PR 1
   + PR 2) — `gofmt -l` on the 10 PR 3a files is clean (the
   apply agent's gofmt fix `b9f8203` resolved the new file). The
   5 pre-existing files in `adapters/common/` with CRLF are NOT
   introduced by PR 3a. CI on Linux/macOS will not see it.

---

## 7. Out-of-Scope Confirmation (explicit)

| Out-of-scope area | File(s) | Status | Evidence |
|---|---|---|---|
| `ReplaceFile` migration to central home + manifest | `adapters/common/strategy.go:120-179` | **NOT MODIFIED** | `git diff main..HEAD -- adapters/common/strategy.go` → 0 lines. Function still uses `.sequoia-backup-<suffix>` + `.sequoia-session` sidecar. **PR 3.5 task 3.3/3.4** will update. |
| `RestoreOrRemoveFile` reads from manifest | `adapters/common/strategy.go:165` | **NOT MODIFIED** | same as above. **PR 3.5 task 3.5/3.6** will update. |
| `BaseAdapter.applyRetention` method | `adapters/common/base_adapter.go:561` | **NOT ADDED** | `git diff main..HEAD -- adapters/common/base_adapter.go | grep -E '(applyRetention\|addWarning.*retention)'` → no matches. **PR 3b task 3.7/3.8** will add. |
| `BaseAdapter.Apply` retention hook | `adapters/common/base_adapter.go:561` | **NOT WIRED** | same as above; verified by `grep applyRetention` → no matches. Confirms PR 3b work is intact. |
| `base_adapter_retention_test.go` | (new file expected in PR 3b) | **NOT PRESENT** | `git ls-files | grep base_adapter_retention_test` → empty. **PR 3b task 3.7** will create. |
| `strategy_central_test.go` | (new file expected in PR 3.5) | **NOT PRESENT** | `git ls-files | grep strategy_central_test` → empty. **PR 3.5 task 3.3/3.5** will create. |
| Manifest helper consolidation (task 3.10) | (depends on PR 3.5) | **DEFERRED** | The original `ce64e25` commit depended on `108414c`; correctly deferred per the apply-progress. **PR 3.5** will consolidate. |
| Production callers of `manifest*` types | (will be added in PR 3.5) | **NONE** | `git grep -E 'manifestEntry\|newEmptyManifest\|readManifest\|writeManifest\|appendManifestEntry\|removeSessionDir' -- 'adapters/**' 'internal/**'` returns matches only in `manifest.go` (declarations) and `manifest_test.go` (tests). **CONFIRMED — no production caller.** |
| TUI retention warning on prune error | `internal/pipeline/runner.go` | **NOT YET ADDED** | The Info message at `runner.go:200-210` covers pre-existing scattered backups (REQ-BRP-05, PR 2 work). The retention cap warning is not in the spec's REQ-BRP-05 wording. **PR 3b task 3.9** will add via `AddWarning` (separate warning, not in Info). |

---

## 8. Budget Overage Assessment

The PR 3a diff is **602 insertions / 11 deletions / 10 files
changed**, exceeding the 400-line review budget by **202 lines
(50.5% overage)**.

**Breakdown by file category**:

| Category | Lines | % of total |
|---|---|---|
| Test code (manifest_test.go 209 + backup_path_builder_internal_test.go 40) | **249** | 41.4% |
| Production code (manifest.go 132 + backup_path_builder.go 19 / 6 = +13) | **145** | 24.1% |
| Docstring updates (5 paths.go files × 11 = 55) | **55** | 9.1% |
| Documentation (apply-progress.md 143) | **143** | 23.7% |
| gofmt fix (manifest.go 1 ins / 1 del) | **2** | 0.3% |
| Deletions (mostly old docstrings) | **-11** | -1.8% |

**Assessment**:

- The **test/prod ratio is 249:145 = 1.72:1 (test-heavy)** — this is
  healthy for strict TDD (every production change has a
  corresponding test in the same commit). The agent's claim of
  "53% tests, 47% production" is accurate (excluding
  apply-progress.md).

- The **5 `paths.go` docstring updates (55 lines total)** are tightly
  coupled to the safety-net change in `BackupPathBuilder.Build`
  (commit `0d91e9f`). They document the new "function not consulted
  on happy path" behavior. Splitting them into a separate PR would
  leave the docstrings "lying" about the code (the safety-net
  consultation happens in the same commit). **They don't make
  sense as a separate PR.**

- The **apply-progress.md (143 lines)** is documentation, not code.
  It is reviewable independently of the code changes and is
  typically a single commit in a stacked chain. **It is
  house-keeping, not a code review concern.**

- The **gofmt fix (2 lines)** is a 1-line fix for an
  alignment issue in the struct tag, applied as a separate commit
  (`b9f8203`). Trivial; not a budget concern.

**Verdict: ACCEPTABLE overage**, recommend proceeding without
further split.

The fallback 3-PR split (3a-manifest-types 341 + 3a-safety-net 124 +
3a-apply-progress 143 = 608 total) would re-arrange the same lines
into 3 PRs but would NOT reduce the total review load — it would
just split the safety-net+paths-docstrings across a second PR
without functional benefit. The 3a PR is **structurally coherent**
(manifest + safety-net + paths-docstrings are tightly coupled at
the design level: the safety-net exists to keep the manifest
path-generation independent of the legacy per-adapter `backupPath`,
and the 5 paths.go docstrings are the user-visible documentation of
that independence).

The apply agent's "option C" decision (3-PR split into 3a/3b/3.5)
is the correct higher-level split. Further splitting 3a into 3a-
manifest/3a-safety-net would be over-fragmentation.

**Recommendation: do NOT apply the fallback 3a split. Merge 3a as-is
after fixing the CRITICAL schema finding (1 line + 1 test).**

---

## 9. Recommendation

**`block-and-fix`**

Rationale:
- 1 CRITICAL finding: the `manifest` struct is missing the
  top-level `created_at` field that the design specifies. The
  orchestrator's schema-check rule explicitly classifies "missing
  field" as CRITICAL because "PR 3.5 will be written against this
  schema and divergence now means rework." The fix is trivial
  (1 line + 1 test).
- Full test suite green (20/20 packages, 3 consecutive runs, no
  flakiness).
- `adapters/common` coverage 85.0%, well above the 70% CI gate.
- New `manifest.go` 71.9% file-level (above 70% gate). Per-function:
  newEmptyManifest 100%, readManifest 78.6%, writeManifest 77.8%,
  appendManifestEntry 80%, removeSessionDir 0% (POSIX-only, correct
  skip on Windows).
- `backup_path_builder.go` per-file coverage 100% (up from 85.7% in
  PR 1+2; the new safety-net test exercises the previously-uncovered
  fallback path).
- `go vet ./...` clean.
- `gofmt -l` on the 10 PR 3a files clean.
- All 5 `paths.go` docstrings correctly updated and consistent.
- All 5 out-of-scope areas (strategy.go, applyRetention hook,
  base_adapter_retention_test.go, strategy_central_test.go,
  production callers of manifest*) correctly untouched.
- The 202-line budget overage is acceptable; do not split further.
- 0 WARNINGs, 2 SUGGESTIONs (TDD commit shape carried from PR 1+2;
  test pollution above cap carried from PR 1+2), 2 SPEC AMBIGUITYs
  (timestamp format example carried from PR 1+2; "directory #4"
  numbering carried from PR 1+2) — all non-blocking and carried
  from earlier PRs.

**The CRITICAL schema finding blocks PR 3a from merging until fixed.** The fix is:

```diff
--- a/adapters/common/manifest.go
+++ b/adapters/common/manifest.go
@@ -39,6 +39,7 @@ type manifestEntry struct {
 // schema {version, created_at, entries:[{version, original_path, suffix, created_at, adapter_id}]}
 type manifest struct {
-	Version string          `json:"version"`
-	Entries []manifestEntry `json:"entries"`
+	Version   string          `json:"version"`
+	CreatedAt time.Time       `json:"created_at"`
+	Entries   []manifestEntry `json:"entries"`
 }
```

Plus a default in `readManifest` (set `m.CreatedAt` to zero-time
or `time.Now().UTC()` if empty) and a default in
`newEmptyManifest` (omit — zero value is fine). Plus a small
round-trip test.

**After the CRITICAL fix is applied**:
- Re-run: `go test ./... -coverprofile=coverage.out -count=1 -timeout 120s`
- Re-verify: `go tool cover -func=coverage.out | grep manifest`
- Confirm: `gofmt -l adapters/common/manifest.go` clean
- The PR can be merged; `next_recommended` becomes `proceed-to-pr-3b`.

---

## Artifacts

- Verify report (this file):
  `openspec/changes/backup-retention-and-organization/verify-report-pr3a.md`
- Coverage profile (root): `coverage.out`
- Coverage profile (adapters/common):
  `C:\Users\Usuario\AppData\Local\Temp\opencode\pr3a_common_cov.out`
- Per-function coverage:
  `C:\Users\Usuario\AppData\Local\Temp\opencode\pr3a_common_cov_func.txt`
- Per-file coverage (computed):
  `C:\Users\Usuario\AppData\Local\Temp\opencode\pr3a_file_cov.txt`
- Test run logs (3 runs):
  `C:\Users\Usuario\AppData\Local\Temp\opencode\pr3a_test_run{1,2,3}.log`
- Manifest function references (production caller check):
  `C:\Users\Usuario\AppData\Local\Temp\opencode\pr3a_manifest_refs.txt`

---

## Structured Envelope (return value)

```json
{
  "status": "fail",
  "executive_summary": "PR 3a manifest + safety-net foundation is solid and well-tested: 6/6 in-scope task groups PASS, full test suite green (20/20 packages, 3 consecutive runs, no flakiness), go vet clean, gofmt clean on all 10 PR 3a files, adapters/common coverage 85.0%, new manifest.go 71.9% file-level (above 70% gate), backup_path_builder.go 100% per-file (up from 85.7% in PR 2). All 5 out-of-scope invariants confirmed (strategy.go untouched, no applyRetention hook, no base_adapter_retention_test.go or strategy_central_test.go, no production callers of manifest types). 1 CRITICAL finding: the manifest struct is missing the top-level created_at field that the design's locked schema specifies (design: {version, created_at, entries:[...]}; implementation: {version, entries}). The fix is trivial (1 line in manifest.go + 1 small round-trip test). 0 WARNINGs. 2 SUGGESTIONs (TDD commit shape carried from PR 1+2; test pollution above cap carried from PR 1+2). 2 SPEC AMBIGUITIES carried from PR 1+2 (timestamp format example; directory #4 numbering). The 202-line budget overage is acceptable; recommend merging PR 3a as-is (after the 1-line schema fix) without further splitting.",
  "artifacts": {
    "verify_report": "openspec/changes/backup-retention-and-organization/verify-report-pr3a.md",
    "coverage_profile_root": "coverage.out",
    "coverage_profile_common": "C:\\Users\\Usuario\\AppData\\Local\\Temp\\opencode\\pr3a_common_cov.out",
    "per_function_coverage": "C:\\Users\\Usuario\\AppData\\Local\\Temp\\opencode\\pr3a_common_cov_func.txt",
    "per_file_coverage": "C:\\Users\\Usuario\\AppData\\Local\\Temp\\opencode\\pr3a_file_cov.txt",
    "test_run_logs": [
      "C:\\Users\\Usuario\\AppData\\Local\\Temp\\opencode\\pr3a_test_run1.log",
      "C:\\Users\\Usuario\\AppData\\Local\\Temp\\opencode\\pr3a_test_run2.log",
      "C:\\Users\\Usuario\\AppData\\Local\\Temp\\opencode\\pr3a_test_run3.log"
    ],
    "manifest_function_refs": "C:\\Users\\Usuario\\AppData\\Local\\Temp\\opencode\\pr3a_manifest_refs.txt"
  },
  "next_recommended": "block-and-fix",
  "risks": [
    "CRITICAL: manifest struct missing top-level created_at (design divergence; fix is 1 line + 1 test in adapters/common/manifest.go)",
    "Test pollution from PR 1 + PR 2 test runs persists: 16+ dirs per adapter in $APPDATA/sequoia/backups/err-test/ and opencode/, above the 5-dir cap. PR 3b task 3.8 (applyRetention hook) will bound this. The cap is the only bound until then.",
    "removeSessionDir has 0% direct coverage on Windows (test correctly platform-skipped with runtime.GOOS check); POSIX CI runners will exercise the path. 71.9% file-level gate is satisfied.",
    "BackupPathBuilder.Build safety-net fallback (line 69) is now reachable only when BackupHomeDir() fails; PR 3.5 may want to remove it (the 5 paths.go docstrings would also need updating). Decide before PR 3.5.",
    "The 5 adapters/<tool>/paths.go docstrings describe the function as 'no longer consulted on the happy path' — post-PR-3a truth. If PR 3.5 removes the safety-net, the docstrings will need a third update.",
    "SetLastBackupDir has 0% direct coverage (tooling artifact — call IS exercised via Codex.Install in installer_internal_test.go). Carried from PR 2.",
    "Spec REQ-BRP-02 example timestamp format (2026-06-15T15-30-45-123Z) differs from implementation (2026-06-15T15-30-45.000Z) — carried from PR 1+2, recommend spec clarification at sdd-archive",
    "Spec REQ-BRP-06 'continues on error' scenario 'directory #4 is read-only' is internally inconsistent — carried from PR 1+2, recommend scenario rewrite at sdd-archive",
    "TUI Info does not yet include a retention count (e.g., '5 most recent kept; N removed'). Not required by the spec; PR 3b may add if product wants it (would be a separate warning via AddWarning, not in Info).",
    "5 pre-existing files in adapters/common/ have CRLF line endings (gofmt issue on Windows). Carried from PR 1+2. CI on Linux/macOS will not see it."
  ]
}
```
