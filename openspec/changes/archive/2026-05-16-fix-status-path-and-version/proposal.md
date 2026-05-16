# Proposal: Fix Status Path Display and Welcome Version

## Intent

Remove installation paths from the TUI status screen (visual clutter) and fix the Welcome screen showing `"0.1.0-dev"` on local builds instead of the resolved build version.

## Scope

### In Scope
- Remove path column from TUI `renderStatusRow()` format string
- Update status screen tests and golden files for the removed path
- Extract `resolveVersion()` from `newVersionCmd` and call it in `runTUI()` before model creation

### Out of Scope
- CLI `runStatus` PATH column (headless mode — scripting/CI users benefit from full paths)
- Adapter `Path` field removal (used by adapter tests and internal logic)
- Unifying `adapters/common/Version` (content version) with CLI binary version
- Changing `WelcomeView` or `NewModel` (they render/store whatever version given)

## Capabilities

### New Capabilities
None

### Modified Capabilities
None — pure display change and version resolution refactor. No spec-level requirements change.

## Approach

**Issue 1**: Remove `path` variable and its `%s` placeholder from `renderStatusRow()` in `status.go:81-87`. Drop the path assertion from `TestStatusView_ShowsVersionAndPath` (rename to `TestStatusView_ShowsVersion`). Update two golden files to omit paths.

**Issue 2**: Lift the `debug.ReadBuildInfo()` resolution logic from `newVersionCmd` (lines 210-222) into a standalone `resolveVersion(raw string) string` function. Call it in `runTUI()` before `app.NewModel(toolID, Version)` becomes `app.NewModel(toolID, resolveVersion(Version))`. `newVersionCmd` delegates to the same function.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/tui/screens/status.go:71-87` | Modified | Drop path from `renderStatusRow()` format |
| `internal/tui/screens/status_test.go:42-52` | Modified | Rename test, remove path assertion |
| `internal/tui/screens/testdata/golden/status_all_installed.txt` | Modified | Remove path lines |
| `internal/tui/screens/testdata/golden/status_mixed.txt` | Modified | Remove path placeholders |
| `cmd/sequoia/main.go:203-231` | Modified | Extract `resolveVersion()` from `newVersionCmd` |
| `cmd/sequoia/main.go:402-412` | Modified | Call `resolveVersion()` in `runTUI()` |
| `cmd/sequoia/main_test.go` | Modified | Add `TestRunTUI_ResolvesDevVersion` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Golden file mismatch after path removal | High | Run `UPDATE_GOLDEN=1 go test ./internal/tui/screens/` after code change |
| `debug.ReadBuildInfo()` returns `(devel)` with no VCS revision | Low | Existing fallback returns raw version — `newVersionCmd` already handles this |

## Rollback Plan

Revert the commit. No schema migrations, no data persisted. Golden files restored from git.

## Dependencies

None

## Success Criteria

- [ ] TUI status screen shows tool name and version only (no paths)
- [ ] `sequoia status` (headless) still prints PATH column unchanged
- [ ] Welcome screen shows resolved build version (e.g., `v0.1.0-abc12345`) on local builds
- [ ] `sequoia version` output unchanged
- [ ] All existing tests pass; new test covers version resolution in TUI path
