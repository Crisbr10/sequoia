# Archive Report: fix-ci-140-lint-and-race

> **Archived**: 2026-06-16
> **Change ID**: fix-ci-140-lint-and-race
> **Created**: (from proposal date)
> **Completed**: 2026-06-16 (CI #142 green on all 5 platforms)
> **Mode**: openspec (file-based)
> **Store**: `openspec/changes/archive/2026-06-16-fix-ci-140-lint-and-race/`
> **Strict TDD**: true (was active throughout)
> **Status**: CLOSED ✅

---

## Executive Summary

The `fix-ci-140-lint-and-race` change fixed 6 lint issues and 2 test assertion failures that caused CI #140 and CI #141 to fail on 4 of 5 CI matrix platforms. The lint fix was `gofmt -w` on 5 CRLF files + deletion of one unused function (`internalFileExists`). The test fix was dependency injection of `removeAllFunc` (a package-level variable defaulting to `os.RemoveAll`) into `PruneBackups`, replacing the broken `chmod 0o500` mechanism that worked on Windows but not on POSIX. CI #142 is green on all 5 platforms (Vulncheck, Lint, 5× Test, 5× Build, Smoke — 13/13 SUCCESS at `https://github.com/Crisbr10/sequoia/actions/runs/27621662064`).

The original proposal diagnosed the test failures as a "data race" between parallel `BaseAdapter.Apply()` tests sharing the package-level `userConfigDir` hook. This diagnosis was refined during the apply phase: CI #141 revealed the actual root cause was silent test bugs — `os.Chmod(dir, 0o500)` on a child directory does not prevent `rmdir(2)` on POSIX, which checks the **parent** directory's permissions, not the child's. The race-fix commits (`OverrideUserConfigDir` export + helpers + 20 `t.Parallel()` removals) were preserved because they address a real, separate class of test-pollution issues and do not hurt; they may help reduce concurrent stress that could mask test bugs in the future.

---

## Artifacts

Archived at: `openspec/changes/archive/2026-06-16-fix-ci-140-lint-and-race/`

| File | Status |
|------|--------|
| `proposal.md` | ✅ |
| `specs/ci-green-gate/spec.md` (delta) | ✅ |
| `design.md` | ✅ |
| `tasks.md` | ✅ (17/17 tasks complete) |
| `apply-progress.md` | ✅ (full history including H.* hotfix section) |
| `verify-report.md` | ✅ |
| `archive-report.md` | ✅ (this file) |

---

## Commits Delivered (10 commits on `main`)

| SHA | Summary |
|-----|---------|
| `7eecf40` | `adapters/common: delete unused internalFileExists (lint fix)` |
| `8e79855` | `chore: gofmt 6 files (CI lint fix)` |
| `46e808e` | `sdd: commit fix-ci-140-lint-and-race apply-progress (BLOCKED)` |
| `666550b` | `adapters/common: export OverrideUserConfigDir helper for cross-package test isolation` |
| `ea4129e` | `adapters/common: isolate test central-home in shared helpers (race fix)` |
| `e4bd2ba` | `adapters/common: isolate central-home in 5 direct-build error tests (race fix)` |
| `0819587` | `sdd: update apply-progress for fix-ci-140-lint-and-race (race fix complete)` |
| `2b45b2d` | `sdd: fix t.Parallel() removal count in apply-progress (16→20)` |
| `bac92d7` | `sdd: commit fix-ci-140-lint-and-race artifacts (proposal, spec, design, tasks, verify-report)` |
| `517cea9` | `adapters/common: inject removeAllFunc for portable POSIX test error injection` |

The apply phase discovered the actual root cause (POSIX `rmdir(2)` parent-permission semantics) differed from the original proposal's hypothesis (race condition). The race-fix commits were preserved because they address a real, separate test-pollution issue and do not hurt. The 10th commit (`517cea9`) added the `removeAllFunc` dependency-injection fix.

---

## Spec Compliance

All 8 REQs from the `ci-green-gate` delta spec are satisfied:

| REQ | Description | Status |
|-----|-------------|--------|
| REQ-CIG-01 | `golangci-lint run ./...` exit 0; 6 lint issues fixed | ✅ PASS |
| REQ-CIG-02 | `go vet ./...` exit 0 | ✅ PASS |
| REQ-CIG-03 | `gofmt -l .` empty (CI; Windows has pre-existing CRLF advisory) | ✅ PASS |
| REQ-CIG-04 | Race-free + correct tests under `-race` | ✅ PASS (corrected — the original "race" diagnosis was refined; tests are now correct on every platform) |
| REQ-CIG-05 | `OverrideUserConfigDir` in every central-home test (31 call sites) | ✅ PASS |
| REQ-CIG-06 | CI workflow unchanged (no `-race` on Windows) | ✅ PASS |
| REQ-CIG-07 | No production behavior change (only `removeAllFunc` injection, defaulting to `os.RemoveAll`) | ✅ PASS |
| REQ-CIG-08 | Coverage ≥ 70% per package; `adapters/common` at 83.7% | ✅ PASS |

---

## Key Decisions

| ID | Decision | Rationale |
|----|----------|-----------|
| **D1** | Lint strategy: `gofmt -w` + delete unused function | `gofmt -w` is atomic and canonical; `internalFileExists` had 0 callers |
| **D2** | Export `OverrideUserConfigDir` from `package common` production file | 23 internal callers in 5 test files needed the exported name; lowercase form was package-internal only |
| **D3** | `removeAllFunc` dependency injection in `PruneBackups` | Replaces `chmod 0o500` trick that doesn't work on POSIX; makes the error path testable on every platform deterministically |
| **D4** | 20 `t.Parallel()` removals in tests sharing `userConfigDir` hook | Required to make the `OverrideUserConfigDir`-based isolation actually fix the race (not strictly necessary for the `removeAllFunc` fix, but preserved as a defensive improvement) |

---

## Deviations from Plan

### Deviation 1: "Race condition" diagnosis was corrected during apply

The original proposal (and design) diagnosed the CI failures as a **data race** between parallel tests sharing `userConfigDir`. CI #141 revealed the **actual** root cause was **assertion failures** — `os.Chmod(dir, 0o500)` on a child directory does not prevent `rmdir(2)` on POSIX, which checks the **parent** directory's permissions. The test silently passed on Windows (where the test was `t.Skip`'d) and silently failed without exercising the error path on all 4 non-Windows runners.

**What changed**: Commit `517cea9` added `removeAllFunc` dependency injection to `PruneBackups`. The previous 9 commits (lint + race-fix pattern) were preserved — they fix a real, separate class of test-pollution issues and do not conflict with the `removeAllFunc` fix.

### Deviation 2: `removeAllFunc` refactor added as 10th commit

The original plan anticipated 6–9 commits covering lint + race-fix. The `removeAllFunc` fix (commit `517cea9`) was added after CI #141 showed the race-fix commits were insufficient to close the test failures.

### Deviation 3: 20 `t.Parallel()` removals (corrected from initial 16)

The apply-progress's §3.3 initially claimed 16 `t.Parallel()` removals. The correct count is **20** (verified in verify-report §8 and corrected in commit `2b45b2d`):
- `base_adapter_test.go`: -4 (14 → 10)
- `base_adapter_error_test.go`: -14 (20 → 6)
- `base_adapter_mockfs_test.go`: -2 (6 → 4)

The `t.Parallel()` removals slow down the test suite by ~1–2s wall-clock. This is documented and intentional.

### Deviation 4: `OverrideUserConfigDir` exported from production file

The design initially suggested using the existing lowercase `overrideUserConfigDir` helper as-is. It was actually defined in a `_test.go` file and could not be called from external test packages. Resolution required exporting it as `OverrideUserConfigDir` in `backup_retention.go` (a non-test production file) — commit `666550b`.

---

## Risks Carried Forward

| Risk | Severity | Note |
|------|----------|------|
| Pre-existing CRLF on `adapters/common/*` | INFO | 65+ files have CRLF on Windows checkout; not seen on CI ubuntu-latest; outside this change's scope |
| `-race` not runnable locally (no CGO) | INFO | CI is source of truth; confirmed green on CI #142 |
| 20 tests lost `t.Parallel()` | LOW | ~1–2s added wall-clock; 9 other parallel tests retain `t.Parallel()` |
| Future test infrastructure changes may need per-goroutine `userConfigDir` | INFO | Separate refactor; not in this change's scope |

**Post-archive follow-ups (optional, none required)**:
- Drop the `chmod 0o500` mechanism in other tests if any are found
- Consider adding `-race` to the Windows CI branch in a future change

---

## Test Results

### CI #142 (all 5 platforms green)

**URL**: `https://github.com/Crisbr10/sequoia/actions/runs/27621662064`

| Job | Platform | Result |
|-----|----------|--------|
| vulncheck | ubuntu-latest | ✅ SUCCESS |
| lint | ubuntu-latest | ✅ SUCCESS |
| test | ubuntu-latest | ✅ SUCCESS |
| test | macos-latest | ✅ SUCCESS |
| test | macos-14 | ✅ SUCCESS |
| test | ubuntu-24.04-arm | ✅ SUCCESS |
| test | windows-latest | ✅ SUCCESS |
| build | ubuntu-latest | ✅ SUCCESS |
| build | macos-latest | ✅ SUCCESS |
| build | macos-14 | ✅ SUCCESS |
| build | ubuntu-24.04-arm | ✅ SUCCESS |
| build | windows-latest | ✅ SUCCESS |
| smoke | ubuntu-latest | ✅ SUCCESS |

**13/13 jobs SUCCESS.**

### Local Verification (Windows runner, no CGO)

| Check | Result |
|-------|--------|
| `go build ./...` | ✅ clean |
| `go vet ./...` | ✅ clean |
| `gofmt -l <6 spec files>` | ✅ clean |
| `go test ./... -count=1 -timeout 180s` | ✅ all 19 packages pass |
| `go test ./adapters/common/... -count=1` | ✅ 5 consecutive runs all deterministic |
| `go test ./adapters/common/... -count=1 -run 'TestPruneBackups_ContinuesOnError\|TestApplyRetention_WarningOnPruneError' -v` | ✅ both tests run and pass (no t.Skip on Windows) |
| `adapters/common` coverage | **83.7%** (above 70% gate) |
| `PruneBackups` function coverage | **87.1%** (up from 77.4% — error branch now exercised on every platform) |

---

## Total Diff Stats

- **10 commits** on `main` (from `20f352d` to `517cea9`)
- **~1700 net lines** (including tests, SDD artifacts, and production code changes)
- **Lint fix**: 6 files formatted, 1 unused function deleted
- **Race/test-fix**: 1 production file changed (`backup_retention.go`), ~20 test files updated

---

## Rollback Plan

Simple: `git revert` the 10 merge commits. The change is non-destructive:
- No production behavior changes except `removeAllFunc` swap, which defaults to `os.RemoveAll` (byte-identical behavior)
- All 10 commits are revertable in sequence
- Rollback restores CI #140's red state — only lint issues and test bugs are reverted

---

## Engram Observations (for traceability)

| Observation | Topic Key | Type |
|-------------|-----------|------|
| `sdd/fix-ci-140-lint-and-race/verify-report` (id: 985) | `sdd/fix-ci-140-lint-and-race/verify-report` | architecture |
| `sdd/fix-ci-140-lint-and-race/archive-report` (this file) | `sdd/fix-ci-140-lint-and-race/archive-report` | architecture |

---

## SDD Cycle Complete

The `fix-ci-140-lint-and-race` change has been fully planned (proposal), specified (delta spec), designed (design.md), implemented (10 commits), verified (CI #142 green), and archived (this report). The change is **closed**.

**Next**: Ready for the next change.
