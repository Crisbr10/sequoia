# Archive Report: T-020 Multi-tool Detection

**Change**: T-020-multi-tool-detection  
**Archived**: 2026-05-09  
**Verdict**: PASS (all warnings resolved during archive)  

---

## What Was Accomplished

Implemented multi-tool detection for the `sequoia` CLI. The `sequoia status` command now displays a 6-column table (ID, NAME, DETECTED, INSTALLED, VERSION, PATH) by calling `a.Status()` on each registered adapter. A new `ScanTools()` function provides structured detection results for all adapters. Each adapter writes a `.sequoia-version` marker file during install and reads it during `Status()`. Symlinked home directories are resolved to real paths via `filepath.EvalSymlinks()`.

## Tasks Completed (6/6)

| Task | Description | Status |
|------|-------------|--------|
| T-020-01 | `versionFilePath()` + `EvalSymlinks` in paths.go | ✅ |
| T-020-02 | `Status()` reads `.sequoia-version` | ✅ |
| T-020-03 | `Install()` writes + `Uninstall()` removes `.sequoia-version` | ✅ |
| T-020-04 | `ScanTools()` + 6-column `runStatus` | ✅ |
| T-020-05 | Integration tests (round-trip, symlink) | ✅ |
| T-020-06 | Edge case tests (legacy, empty registry) | ✅ |

## Files Changed (8 modified, 0 new, 0 deleted)

| File | Description |
|------|-------------|
| `adapters/claude/paths.go` | Added `versionFilePath(base)`, `EvalSymlinks` in `claudeBase()` |
| `adapters/claude/adapter.go` | `Status()` reads version, `Install()` writes, `Uninstall()` removes |
| `adapters/opencode/paths.go` | Mirror of claude changes |
| `adapters/opencode/adapter.go` | Mirror of claude changes |
| `cmd/sequoia/main.go` | `ScanTools()` function, 6-column `runStatus` format |
| `cmd/sequoia/main_test.go` | CLI tests for new output format |
| `adapters/claude/adapter_test.go` | Version file read/write tests |
| `adapters/opencode/adapter_test.go` | Mirror of claude tests |

## Warnings Resolved During Archive

1. **Spec-design Path field mismatch** (WARNING → RESOLVED): Updated `multi-tool-detection/spec.md` from `SystemPromptPath()` to `SkillsPath()` — matching the design decision and implementation. Spec updated in both Engram (#190) and OpenSpec main specs.

2. **tasks.md checkboxes not updated** (WARNING → RESOLVED): All 6 task checkboxes in `tasks.md` updated from `[ ]` to `[x]`. File archived at `openspec/changes/archive/2026-05-09-T-020-multi-tool-detection/tasks.md`.

3. **Symlink tests skip on Windows** (WARNING → ACKNOWLEDGED): `TestAdapter_Base_SymlinkResolved` in both adapters gracefully skips on Windows (admin required for `os.Symlink`). Per design; verified on macOS/Linux CI.

## Test Results

- **Tests**: 108 total, 0 failures across 5 packages
- **Coverage**: Core functions at 100% (Status, runStatus, ScanTools, versionFilePath)
- **Type Check**: `go vet` clean

## Specs Synced to Source of Truth

| Domain | Action | Details |
|--------|--------|---------|
| `multi-tool-detection` | Created | 2 requirements: ScanTools detection results, Cross-platform detection |
| `status-reporting` | Created | 2 requirements: CLI output using AdapterStatus, Empty registry edge case |
| `version-tracking` | Created | 2 requirements: Version marker write/remove, Version read during Status |
| `symlink-handling` | Created | 2 requirements: Symlink resolution in base paths, Graceful fallback |

## Artifact Traceability (Engram)

| Artifact | Engram ID | Topic Key |
|----------|-----------|-----------|
| Proposal | #188 | `sdd/T-020-multi-tool-detection/proposal` |
| Spec (corrected) | #190 | `sdd/T-020-multi-tool-detection/spec` |
| Design | #192 | `sdd/T-020-multi-tool-detection/design` |
| Tasks | #194 | `sdd/T-020-multi-tool-detection/tasks` |
| Verify Report | #196 | `sdd/T-020-multi-tool-detection/verify-report` |
| Archive Report | — | `sdd/T-020-multi-tool-detection/archive-report` |

## Design Decisions Preserved

| Decision | Choice |
|----------|--------|
| Version file location | Per-adapter skills dir (`SkillsPath()/.sequoia-version`) |
| Version file format | Plain text (single line) |
| Missing version file | Silent empty string (backward-compatible) |
| Path field source | `SkillsPath()` (resolved via EvalSymlinks) |
| ScanTools location | `cmd/sequoia/` |
| Symlink resolution | In `base()` functions |
| CLI columns | 6: ID(14), NAME(14), DETECTED(9), INSTALLED(10), VERSION(10), PATH(55) |

## SDD Cycle Complete

The change has been fully planned, implemented, verified with warnings resolved, and archived. Ready for the next change (T-030 TUI status screen).
