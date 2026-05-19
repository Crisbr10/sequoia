# Archive Report — p4-005-codex-write-error

**Date**: 2026-05-19
**Status**: ✅ Archived
**Mode**: hybrid (openspec + engram)

## What Was Done

Fixed a HIGH-severity silent error discard in `MergeConfig` (`adapters/codex/installer.go:36-39`) and `ReplaceFile` (`adapters/common/strategy.go:132-135`). Both functions silently discarded `AtomicWriteFile` errors when writing session-tracking files (`.sequoia-session`). When session writes fail, `RemoveConfig`/`RestoreOrRemoveFile` fall back to a single legacy backup name that may be stale, restoring wrong content during uninstall. The fix returns wrapped errors so installs fail cleanly with proper rollback instead of silently producing broken uninstall state.

## Files Changed

| File | Action | Description |
|------|--------|-------------|
| `adapters/codex/installer.go:36-39` | Modified | Replace empty error block with `return fmt.Errorf("merge config: session: %w", err)` |
| `adapters/common/strategy.go:132-135` | Modified | Replace empty error block with `return fmt.Errorf("replace file: session: %w", err)` |
| `adapters/codex/installer_test.go` | Modified | Added `TestMergeConfig_SessionWriteFails` |
| `adapters/common/strategy_test.go` | Modified | Added `TestReplaceFile_SessionWriteFails` |

## Test Results

| Package | Tests | Result |
|---------|-------|--------|
| `adapters/codex` | 11 (including new) | ✅ PASS |
| `adapters/common` | 22 (including new) | ✅ PASS |
| `adapters` | All | ✅ PASS |
| `adapters/claude` | All | ✅ PASS |
| `adapters/cursor` | All | ✅ PASS |
| `adapters/gemini` | All | ✅ PASS |
| `adapters/opencode` | All | ✅ PASS |
| `adapters/testutil` | All | ✅ PASS |

**Total**: 62 passed, 0 failed, 0 skipped
**Build**: `go build ./...` — clean, no errors

## Verification Outcome

**Verdict**: ✅ PASS

| Category | Result |
|----------|--------|
| Spec compliance | 6/6 scenarios compliant |
| TDD compliance | 6/6 checks passed |
| Design coherence | All 4 design decisions followed |
| Assertion quality | All assertions verify real behavior |
| Caller compatibility | All callers compile and handle errors (no signature changes) |
| Cross-platform | Directory-as-session-path technique proven by existing tests |

## Delta Specs Synced

| Domain | Action | Details |
|--------|--------|---------|
| `go-wiring` | Updated | 2 new requirements added: `MergeConfig Session File Error Propagation` (3 scenarios) and `ReplaceFile Session File Error Propagation` (3 scenarios) |

## Archive Contents

| Artifact | Present |
|----------|---------|
| `proposal.md` | ✅ |
| `exploration.md` | ✅ |
| `spec.md` | ✅ |
| `design.md` | ✅ |
| `tasks.md` | ✅ (5/5 tasks complete) |
| `verify-report.md` | ✅ |
| `archive-report.md` | ✅ |

## Source of Truth Updated

The following spec now reflects the new behavior:
- `openspec/specs/go-wiring/spec.md` — 2 new error propagation requirements appended

## SDD Cycle Complete

The change has been fully planned, implemented, verified, and archived.
Ready for the next change.
