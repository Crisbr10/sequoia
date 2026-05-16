# Delta for symlink-handling

## ADDED Requirements

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
