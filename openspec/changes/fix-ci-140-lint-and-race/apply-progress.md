# Apply Progress: fix-ci-140-lint-and-race

> **Branch**: `feature/fix-ci-140-lint-and-race`
> **PR scope**: Single PR (per Design Decision 5)
> **Strict TDD**: ACTIVE
> **Commits ahead of main**: 2 (Commits 1+2 done) + apply-progress = 3
> **Status**: ⚠️ **BLOCKED** — package boundary issue on `overrideUserConfigDir` (see §4)

---

## 1. Executive Summary

Commits 1 and 2 (the lint fix) are complete and verified. The race fix
(Commits 3+4) **cannot proceed** because `overrideUserConfigDir` is
unexported and lives in `package common` (the internal test package,
defined at `adapters/common/backup_retention_test.go:349`), while the
test files that need to call it (`base_adapter_test.go`,
`base_adapter_error_test.go`) are in `package common_test` (the external
test package). Unexported symbols cannot be referenced across Go
packages, so the call does not compile.

This was the explicit pre-warning in the orchestrator's prompt:
> If for some reason it's not accessible from the test file, document
> it as a blocker and return BLOCKED.

**Resolution path** (in §4): export the helper (rename to
`OverrideUserConfigDir`) — requires a 1-line change in
`backup_retention_test.go` + 14 caller updates in `package common`
test files. This is a design-level change that the orchestrator must
approve before this apply can continue.

**Commit SHAs (2 work-unit commits done; 3 are blocked)**:

| SHA | Commit | Tasks | Status |
|---|---|---|---|
| `7eecf40` | `adapters/common: delete unused internalFileExists (lint fix)` | 1.1 | ✅ DONE |
| `8e79855` | `chore: gofmt 6 files (CI lint fix)` | 2.1–2.6 | ✅ DONE |
| (not made) | `adapters/common: isolate test central-home in shared helpers` | 3.1–3.3 | ⛔ BLOCKED |
| (not made) | `adapters/common: isolate central-home in 5 direct-build tests` | 4.1–4.5 | ⛔ BLOCKED |
| (this commit) | `sdd: commit fix-ci-140-lint-and-race apply-progress (BLOCKED)` | X.2 | 📝 DONE |

**Files touched (2 done; 0 blocked = 2 actual edits)**:

| File | Action | Lines | Status |
|---|---|---|---|
| `adapters/common/base_adapter_internal_test.go` | deleted `internalFileExists` + removed now-unused `os` import | -8 | ✅ |
| `adapters/opencode/adapter.go` | `gofmt -w` (CRLF→LF + inline-closure reformat) | +3 / -1 | ✅ |
| `adapters/testutil/mock_adapter.go` | `gofmt -w` (CRLF→LF) | line-ending only | ✅ |
| `internal/codegraph/install.go` | `gofmt -w` (CRLF→LF) | line-ending only | ✅ |
| `internal/codegraph/install_test.go` | `gofmt -w` (CRLF→LF) | line-ending only | ✅ |
| `internal/tui/styles/logo.go` | `gofmt -w` (CRLF→LF) | line-ending only | ✅ |
| `internal/tui/styles/styles.go` | `gofmt -w` (CRLF→LF) | line-ending only | ✅ |
| `adapters/common/base_adapter_error_test.go` | +3 lines (one `overrideUserConfigDir` per helper/test) | **⛔ BLOCKED** — package boundary |
| `adapters/common/base_adapter_test.go` | +6 lines (one `overrideUserConfigDir` per helper) | **⛔ BLOCKED** — package boundary |

**Diff total (done so far)**: 7 files changed, 3 insertions(+), 8
deletions(-) for content + 5 line-ending-only changes. Well under the
400-line budget.

---

## 2. Task Status

| Task | Status | Commit SHA | Notes |
|---|---|---|---|
| 1.1 | ✅ DONE | `7eecf40` | `internalFileExists` deleted; `os` import removed (no longer used); `go test ./adapters/common/... -count=1` still passes; `golangci-lint` no longer reports the `unused` issue on `base_adapter_internal_test.go:57` |
| 2.1 | ✅ DONE | `8e79855` | `gofmt -w adapters/testutil/mock_adapter.go` — line-ending normalization |
| 2.2 | ✅ DONE | `8e79855` | `gofmt -w internal/codegraph/install.go` — line-ending normalization |
| 2.3 | ✅ DONE | `8e79855` | `gofmt -w internal/codegraph/install_test.go` — line-ending normalization |
| 2.4 | ✅ DONE | `8e79855` | `gofmt -w internal/tui/styles/logo.go` — line-ending normalization |
| 2.5 | ✅ DONE | `8e79855` | `gofmt -w internal/tui/styles/styles.go` — line-ending normalization |
| 2.6 | ✅ DONE | `8e79855` | `gofmt -w adapters/opencode/adapter.go` — line-ending + inline-closure reformat (3-line content change at L42) |
| 3.1 | ⛔ BLOCKED | — | `overrideUserConfigDir` not accessible from `package common_test`. Resolution: see §4. |
| 3.2 | ⛔ BLOCKED | — | Same as 3.1 |
| 3.3 | ⛔ BLOCKED | — | Same as 3.1 |
| 4.1 | ⛔ BLOCKED | — | Same as 3.1 |
| 4.2 | ⛔ BLOCKED | — | Same as 3.1 |
| 4.3 | ⛔ BLOCKED | — | Same as 3.1 |
| 4.4 | ⛔ BLOCKED | — | Same as 3.1 |
| 4.5 | ⛔ BLOCKED | — | Same as 3.1 |
| X.1 | ⛔ BLOCKED | — | Final verification depends on Commits 3+4; lint and build are green but `go test ./...` cannot be re-run with the race-fix edits until the package boundary is resolved |
| X.2 | 📝 DONE | (this commit) | apply-progress committed per the orchestrator's instructions even though the apply is BLOCKED — this is the explicit fallback path |

---

## 3. Verification (partial — what we CAN verify after Commits 1+2)

**`go build ./...`**: clean (exit 0, no output)
**`go vet ./...`**: clean (exit 0, no output)
**`gofmt -l <6 files>`** (the spec target set): clean (all 6 files now gofmt-formatted)
**`gofmt -l .` (full tree)**: reports 65 files remaining (all in `adapters/common/*` with pre-existing CRLF — see Open Risks §5.1). On the CI ubuntu-latest runner this is reported as 0 because the git index on the runner has the files already in LF form; the local Windows CRLF is a checkout artifact documented in PR 1+2+3a+3b verify reports.
**`golangci-lint run ./adapters/common/...`**: 0 `unused` issues; 5 `gofmt` issues remain (all in `adapters/common/*` pre-existing CRLF, NOT introduced by this PR — see §5.1).
**`go test ./adapters/common/... -count=1 -timeout 60s`**: PASS (post-Commit-1)
**`go test ./adapters/testutil/... ./adapters/opencode/... ./internal/codegraph/... ./internal/tui/styles/... -count=1 -timeout 60s`**: all 4 packages PASS (post-Commit-2)

**`-race`**: NOT RUN on this Windows runner — no CGO available (no `gcc` on PATH). Per the orchestrator's "ask-on-risk" + the established PR 1+2+3a+3b pattern, CI #141 on the 4 non-windows runners is the source of truth for the race fix. With the race fix BLOCKED in this apply, CI #141's `-race` jobs will continue to fail until the package boundary is resolved.

**Coverage**: not re-measured (production code is unchanged in Commits 1+2; test code in Commit 1 was a 1-function deletion that only removed a dead helper, so coverage is unchanged from main). The race fix (Commits 3+4) is the only thing that could affect coverage, and it only adds new test invocations of existing paths.

---

## 4. Open Risks (BLOCKER + advisory for sdd-verify)

### 4.1 ⛔ BLOCKER — `overrideUserConfigDir` is unexported across package boundary

**Symptom**: `go test ./adapters/common/... -count=1` fails with:
```
adapters\common\base_adapter_error_test.go:96:2: undefined: overrideUserConfigDir
adapters\common\base_adapter_test.go:206:2: undefined: overrideUserConfigDir
adapters\common\base_adapter_test.go:331:2: undefined: overrideUserConfigDir
```

**Root cause**: `overrideUserConfigDir` is defined at
`adapters/common/backup_retention_test.go:349` in `package common`
(the **internal** test package). The 8 test files that need to call
it split into two Go packages:
- `package common` (internal): `backup_retention_test.go`,
  `base_adapter_internal_test.go`, `base_adapter_retention_test.go`,
  `backup_path_builder_internal_test.go`, `strategy_central_test.go`,
  `manifest_test.go` — 6 files, ~14 callers, all work fine.
- `package common_test` (external): `base_adapter_test.go`,
  `base_adapter_error_test.go`, `base_adapter_strategy_test.go`,
  `installer_test.go`, `backup_path_builder_test.go`, etc. — these
  cannot reference the unexported helper.

The 8 callers required by this change (Tasks 3.1, 3.2, 3.3, 4.1–4.5)
are all in `package common_test` files. The design assumed the helper
was in `package common_test` (the design's "14 callers" list contains
only `package common` files, which is consistent with the design's
silent assumption that the helper was externally accessible).

**Resolution path (recommended — least invasive)**:

Export the helper. Change the definition in
`adapters/common/backup_retention_test.go:349` from:
```go
func overrideUserConfigDir(t *testing.T, fn func() (string, error)) {
```
to:
```go
// OverrideUserConfigDir is the exported form of overrideUserConfigDir,
// accessible from the package common_test external test package.
// See the doc comment above for usage.
func OverrideUserConfigDir(t *testing.T, fn func() (string, error)) {
```

And update the 14 callers in the 6 `package common` test files to
use the new exported name. The 8 new callers in `package common_test`
files will use `OverrideUserConfigDir` directly.

This is a 16-file change. It does NOT change any production behavior.
The apply cannot proceed without this resolution.

**Alternative resolutions considered and rejected**:
- **Move the helper to a new `package common_test` file**: requires
  the helper to access the `userConfigDir` variable which is
  unexported in `package common` — does not work.
- **Convert the 8 callers' test files to `package common`**: requires
  renaming every type/value/function reference (e.g. `common.BaseAdapter`
  → `BaseAdapter`, `common.NewPathResolver` → `NewPathResolver`); high
  blast radius, NOT minimal.
- **Add a thread-safe override mechanism to production code
  (`adapters/common/backup_retention.go`)**: pollutes production code
  with test-only machinery; rejected on design grounds.

**Orchestrator decision required**: approve export-resolution
(rename + 14 caller updates) and re-launch `sdd-apply` with
expanded scope. The expanded scope is a one-time 16-file rename; it
does not change behavior or design.

### 4.2 ⛔ BLOCKER — `X.1` final verification depends on Commits 3+4

The orchestrator's strict-TDD procedure requires `go test ./... -race -count=1 -timeout 180s`
× 5 consecutive runs to confirm the race fix. With Commits 3+4 blocked,
this cannot be performed. CI #141 on the 4 non-windows runners is the
deferred verification, contingent on the package boundary resolution.

### 4.3 Advisory for sdd-verify — 5 pre-existing `gofmt` issues remain locally

`golangci-lint run ./adapters/common/...` on this Windows runner
reports 5 `gofmt` issues in pre-existing files
(`backup_path_builder.go`, `backup_retention.go`, `base_adapter.go`,
`commands.go`, `installer.go`). These are CRLF line endings on
Windows; on CI's ubuntu-latest runner they are LF in the index and
are not flagged. These 5 files are NOT touched by this PR (per the
design's "Out of scope" section and the prior PR 1+2+3a+3b apply
progress notes). `sdd-verify` should expect the same local-Windows
artifact and confirm the CI lint job is clean on the 4 non-windows
runners.

### 4.4 Advisory — `gofmt -l .` reports 65 files locally (not 6)

Same root cause as 4.3. The 6 files the spec targets are now clean
(post-Commit-2). The 65 remaining files are the broader Windows
checkout CRLF artifact, unchanged from main. CI's `gofmt -l` on
ubuntu-latest does not see this.

---

## 5. Next Batch Hint

**⛔ DO NOT PROCEED to `sdd-verify` until the package boundary is resolved.**

Branch state: `feature/fix-ci-140-lint-and-race` is **2 commits ahead
of main** (`7eecf40`..`8e79855`) + the apply-progress commit. The
lint half of the change is shipped; the race half is blocked.

**Orchestrator's two options**:

1. **Approve export-resolution (recommended)**: re-launch `sdd-apply`
   with expanded scope — add a new commit "export `OverrideUserConfigDir`
   and update 14 existing callers" before Commit 3. This is a 16-file
   rename, ~30 lines of pure diff, no behavior change. Then Commits 3+4
   proceed as designed.

2. **Defer race fix to a follow-up `sdd-propose` change**: ship the
   lint-only Commits 1+2 as PR-1 of a chained chain, and file a new
   `fix-ci-140-race` change for the race fix alone. This keeps the
   lint PR minimal (clean, no design issues) but requires a follow-up
   cycle to address CI #141's race job failures.

**Either way, CI #141 will fail on the 4 non-windows race jobs until
the race fix is in.**

**Relevant Files** (for the orchestrator to review):
- `adapters/common/backup_retention_test.go:349` — helper definition (the only file to edit for export-resolution)
- `adapters/common/backup_retention_test.go:339-348` — doc comment to update
- 6 `package common` test files containing the 14 callers to rename
- `adapters/common/base_adapter_test.go` — needs 2 `OverrideUserConfigDir` calls (Tasks 3.2, 3.3)
- `adapters/common/base_adapter_error_test.go` — needs 6 `OverrideUserConfigDir` calls (Tasks 3.1, 4.1–4.5)
