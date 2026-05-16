# Proposal: Fix Uninstall Orphaned Files

## Intent

Gemini and Codex adapters use wrong path resolution in Install/Uninstall, operating on relative paths (e.g., `./.gemini/sequoia/`) instead of absolute user-home paths (`~/.gemini/sequoia/`). In production, uninstalling leaves orphaned files behind. Additionally, the TUI CompleteView shows "Installed: Skills, Commands, System Prompt" even on clean uninstalls — a contradictory UX.

## Scope

### In Scope
- Fix Gemini adapter Uninstall path resolution (`geminiBase(a.HomeDir())` → `a.base()`)
- Fix Codex adapter Install path resolution (`codexBase(a.HomeDir())` → `a.base()`)
- Fix Codex adapter Uninstall path resolution (`codexBase(a.HomeDir())` → `a.base()`)
- Add `complete.uninstalled_items` i18n key (en + es)
- Fix CompleteView conditional to show correct message for clean uninstalls
- Add tests exercising the production code path (homeDir="")

### Out of Scope
- OpenCode adapter (already correct)
- TUI screens beyond complete.go
- Install/Uninstall pipeline logic changes

## Capabilities

### New Capabilities
None — pure bugfix, no new behaviors.

### Modified Capabilities
None — existing specs already describe correct behavior; these bugs violate them.

## Approach

Replace all `geminiBase(a.HomeDir())` / `codexBase(a.HomeDir())` calls with `a.base()` in the three affected adapter methods. `a.base()` already handles production home dir resolution (`os.UserHomeDir()` + `sync.Once` cache + symlink resolution) — the same method correctly used in `IsInstalled()`, `SkillsPath()`, and OpenCode/Claude adapters.

For TUI: add `uninstalled_items` i18n key, update CompleteView conditional to handle three branches — install, uninstall-without-warnings, uninstall-with-warnings.

Tests: add table-driven cases where `homeDir` is left as `""` (production path), asserting that orphaned files do NOT remain after uninstall.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `adapters/gemini/adapter.go:77` | Modified | Uninstall: `geminiBase(a.HomeDir())` → `a.base()` |
| `adapters/codex/adapter.go:80` | Modified | Install: `codexBase(a.HomeDir())` → `a.base()` |
| `adapters/codex/adapter.go:175` | Modified | Uninstall: `codexBase(a.HomeDir())` → `a.base()` |
| `internal/tui/screens/complete.go:73-78` | Modified | Conditional: add clean-uninstall branch |
| `internal/i18n/translations/en.toml` | Modified | New key `complete.uninstalled_items` |
| `internal/i18n/translations/es.toml` | Modified | New key `complete.uninstalled_items` |
| `adapters/gemini/adapter_test.go` | Modified | Add production-path test cases |
| `adapters/codex/adapter_test.go` | Modified | Add production-path test cases |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `a.base()` introduces symlink resolution that was previously bypassed, causing differing behavior | Low | Symlink resolution is read-only and already used by `IsInstalled()` — no known issues |
| Production code path (homeDir="") is only testable on real user-home filesystems | Med | Use `t.Setenv("HOME", t.TempDir())` or equivalent to test without mutating real home |
| Changing Codex Install() simultaneously with Uninstall() could introduce install regression | Low | Both use the same broken pattern — replacing both with the same correct pattern is symmetric |

## Rollback Plan

Revert the three `xxxBase(a.HomeDir())` → `a.base()` changes in Gemini and Codex adapters. The i18n/TUI changes are additive and do not need rollback.

## Dependencies

None.

## Success Criteria

- [ ] `sequoia uninstall` removes files from `~/.gemini/sequoia/` and `~/.codex/sequoia/` (not `./.gemini/` or `./.codex/`)
- [ ] TUI shows "Uninstalled: Skills, Commands, System Prompt" on clean uninstall
- [ ] All adapter tests pass with both `SetHomeDir(t.TempDir())` and `homeDir=""` paths
- [ ] No regression in `go test -race -count=1 ./...`
