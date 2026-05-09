# Proposal: Multi-tool Detection (T-020)

## Intent

`sequoia status` currently ignores `ToolAdapter.Status()` and only shows ID/NAME/DETECTED/SEQUOIA (yes/no). Users cannot see installation paths, Sequoia version, or whether paths are symlink-resolved. This change surfaces all `AdapterStatus` fields and adds version tracking so users know exactly what is installed where.

## Scope

### In Scope
- CLI `status` output: show name, installation path, installed (yes/no), and Sequoia version using `a.Status()`
- Version marker file (`.sequoia-version`) written during install, read during `Status()` in both adapters
- Symlink-safe paths via `filepath.EvalSymlinks()` in adapter `base()` functions
- Dedicated `ScanTools()` function returning structured detection results for all registered adapters
- Cross-platform tests per OS using `t.TempDir()` as mock home directory

### Out of Scope
- New adapters or tool-specific detection changes (`Detect()` already exists)
- TUI status screen (Phase 5, T-030)
- Uninstall-time version cleanup (handled by file removal during `Uninstall()`)

## Capabilities

### New Capabilities
- `multi-tool-detection`: Scan home directory for all supported AI tools, detect Sequoia installation state, return structured status (name, path, installed, version) per tool
- `version-tracking`: Write `.sequoia-version` marker during install; read it during `Status()` so version is always accurate

### Modified Capabilities
None — no existing specs for the sequoia-ai module.

## Approach

1. **Adapter install writes version file**: Each adapter's `Install()` writes a `.sequoia-version` file (containing `Version` constant) into the skills directory. File is removed during `Uninstall()` as part of cleanup.
2. **Adapter Status() reads version file**: Populate `AdapterStatus.Version` by reading `.sequoia-version`. Return `Path` as `SystemPromptPath()` (the root install path). Delegates to `IsInstalled()` for the `Installed` field.
3. **Symlink resolution in base()**: Apply `filepath.EvalSymlinks()` to the resolved home path in `claudeBase()` and `opencodeBase()`, falling back to unresolved path on error.
4. **CLI `runStatus` uses `a.Status()`**: Replace per-field `Detect()`/`IsInstalled()` calls with single `Status()` call. Add PATH and VERSION columns. Use fixed-width format.
5. **New `ScanTools()` function**: In `cmd/sequoia/main.go`, iterate `adapters.DefaultRegistry.All()`, call `Status()`, return `[]AdapterStatus`. Shared by `status` and future TUI.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `adapters/interface.go` | Unchanged | `AdapterStatus` already has all needed fields |
| `adapters/claude/adapter.go` | Modified | `Status()` reads `.sequoia-version`; `Install()` writes it; returns `SystemPromptPath()` as Path |
| `adapters/claude/paths.go` | Modified | Add `versionFilePath(base)`; `claudeBase()` calls `filepath.EvalSymlinks()` |
| `adapters/opencode/adapter.go` | Modified | Same changes as claude adapter |
| `adapters/opencode/paths.go` | Modified | Add `versionFilePath(base)`; `opencodeBase()` calls `filepath.EvalSymlinks()` |
| `cmd/sequoia/main.go` | Modified | `runStatus` uses `a.Status()`; add PATH/VERSION columns; new `ScanTools()` function |
| `cmd/sequoia/main_test.go` | Modified | Tests for `runStatus` output format, `ScanTools()`, temp-dir-based detection |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `.sequoia-version` write failure breaks install | Low | Write failure triggers rollback via existing Prepare/Apply/Verify; file is small (~10 bytes) |
| `filepath.EvalSymlinks()` errors on Windows | Low | Ignore errors gracefully — unresolved path is still correct and usable |
| Version file format mismatch between install/status | Low | Use shared `sequoiaVersionFile` constant; file content is a single-line version string |

## Rollback Plan

1. **Code revert**: `git revert` the feature branch. `go test ./...` passes because AdapterStatus.Version was always empty before.
2. **Orphaned version files**: Existing `.sequoia-version` files are harmless — `Status()` ignores them if the adapter code is reverted. Or they are cleaned out on next `sequoia uninstall`.
3. **CLI output change**: No backward-compatibility concern — `sequoia status` is informational, not machine-parsed.

## Dependencies

- **T-019** (CLI base with Cobra): DONE — provides `cmd/sequoia/main.go` with `runStatus` and adapter registration

## Success Criteria

- [ ] `sequoia status` shows NAME, PATH, INSTALLED, VERSION columns for all registered adapters
- [ ] Version reads correctly after install (matches `Version` constant in adapter package)
- [ ] Symlinked home directories resolve to real paths in status output
- [ ] Tests pass on Windows (current platform) using `t.TempDir()` for mock homes
- [ ] `ScanTools()` returns one `AdapterStatus` per registered adapter with all fields populated

## Effort Estimate

~6 hours (3× code, 2× tests, 1× integration validation)
