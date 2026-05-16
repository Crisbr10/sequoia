# Specs: Symlink Handling

## Domain: symlink-handling

### Requirement: Symlink Resolution in Base Paths

Each adapter's `base()` function MUST apply `filepath.EvalSymlinks()` to the resolved home directory before constructing the tool config path. The status output SHALL display the resolved (real) path, not the symlink path.

**Scenario: macOS home directory is a symlink**
- GIVEN `/Users/jane` is a symlink to `/Volumes/External/Users/jane`
- WHEN `claudeBase(homeDir)` resolves the home directory
- THEN the returned path MUST be `/Volumes/External/Users/jane/.claude`
- AND `sequoia status` MUST display the real path

**Scenario: No symlinks — normal path unchanged**
- GIVEN the home directory is not a symlink
- WHEN `base()` resolves the path
- THEN the returned path MUST be identical to `filepath.Join(homeDir, ".claude")` (or `.config/opencode`)

### Requirement: Centralized Symlink Resolution

The system SHALL use a single shared symlink resolution function for all adapters. No adapter SHALL duplicate `EvalSymlinks` or fallback logic.

**Scenario: Shared function used by all adapters**
- GIVEN multiple AI tool adapters
- WHEN any adapter resolves its home directory
- THEN a shared resolution function MUST be used
- AND no adapter SHALL contain its own EvalSymlinks fallback code

**Scenario: Adapter path functions are minimal**
- GIVEN resolution is centralized
- WHEN an adapter constructs its config path
- THEN it SHALL only join the resolved home with the tool directory name
- AND it SHALL NOT duplicate symlink logic

### Requirement: Graceful Fallback on Resolution Failure

If `filepath.EvalSymlinks()` returns an error, the adapter MUST fall back to the unresolved path and SHALL NOT propagate the error. The system MUST detect whether the unresolved path is a symlink via `os.Lstat`. When the path IS a symlink, a warning containing the unresolved path MUST be emitted to the user. When the path is NOT a symlink, the fallback is legitimate and no warning SHALL be emitted.

**Scenario: EvalSymlinks fails — path is a symlink**
- GIVEN EvalSymlinks fails and os.Lstat confirms the path is a symlink
- WHEN base() is called
- THEN the unresolved path MUST be returned without error
- AND a warning with the unresolved path MUST be emitted

**Scenario: EvalSymlinks fails — path is NOT a symlink**
- GIVEN EvalSymlinks fails and os.Lstat confirms the path is NOT a symlink
- WHEN base() is called
- THEN the unresolved path MUST be returned without error
- AND no warning SHALL be emitted

**Scenario: EvalSymlinks succeeds**
- GIVEN EvalSymlinks succeeds
- WHEN base() is called
- THEN the resolved path MUST be used
- AND no warning SHALL be emitted

**Scenario: Warning in TUI**
- GIVEN a symlink warning was emitted
- WHEN the TUI renderer processes the ProgressMsg
- THEN the affected tool SHALL show a warning marker
- AND the warning message MUST include the unresolved path

**Scenario: Warning in CLI**
- GIVEN a symlink warning was emitted
- WHEN running in CLI headless mode
- THEN the warning MUST be printed to stderr
- AND MUST include the unresolved path

**Scenario: Warning clears on re-resolution**
- GIVEN a previous install emitted a symlink warning
- WHEN the symlink becomes resolvable and install is re-run
- THEN no warning SHALL be emitted
- AND the resolved path MUST be used

### Requirement: All Adapter Install/Uninstall MUST Use Centralized Path Resolution

Gemini and Codex adapter Install and Uninstall methods SHALL use `a.base()` for path resolution — the same centralized method used by `IsInstalled()`, `SkillsPath()`, and BaseAdapter. No adapter SHALL bypass `base()` by calling a tool-specific path function directly with `a.HomeDir()`.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 1 | Gemini Uninstall — production path | Gemini adapter, homeDir="" | Uninstall() called | Path resolved via `a.base()`; files removed from `~/.gemini/sequoia/` |
| 2 | Codex Uninstall — production path | Codex adapter, homeDir="" | Uninstall() called | Path resolved via `a.base()`; files removed from `~/.codex/sequoia/` |
| 3 | Codex Install — production path | Codex adapter, homeDir="" | Install() called | Path resolved via `a.base()`; files placed in `~/.codex/sequoia/` |
| 4 | No orphaned files after uninstall | Sequoia installed for Gemini/Codex via homeDir="" | Uninstall completes | Zero sequoia files remain in tool config directory |
| 5 | Uninstall when files already absent | Adapter never had Sequoia installed | Uninstall runs | Reports success without error; no filesystem mutation |
| 6 | Partial uninstall — some files pre-deleted | Only Skills file exists; Commands/Prompt already absent | Uninstall runs | Removes Skills; reports success without error on pre-absent files |
