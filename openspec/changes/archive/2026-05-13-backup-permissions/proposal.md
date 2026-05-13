# Proposal: Restrict Backup File Permissions

## Intent

Sequoia writes backup files with world-readable permissions (`0o644`) and backup directories with world-executable permissions (`0o755`). On Unix multi-user systems, other users can read backup files containing personal AI tool configuration. Fix 4 production bugs to restrict backup file permissions to owner-only (`0o600` for files, `0o700` for dirs).

## Scope

### In Scope
- B1: Backup dir `MkdirAll` from `0o755` → `0o700` (`installer.go:84`)
- B2: Backup file `Chmod` to `0o600` after copy (`installer.go` post-line 89)
- B3: Backup file `WriteFile` from `0o644` → `0o600` (`strategy.go:127`)
- B4: Codex backup `WriteFile` from `0o644` → `0o600` (`codex/installer.go:30`)
- T1: Template backup `WriteFile` from `0o644` → `0o600` (`_template/installer.go:52`)
- Update test assertions to match new permission constants

### Out of Scope
- Production file permissions (skill files, command files, version markers, system prompts) — stay at `0o644`/`0o755`
- Restore logic (rollback/uninstall) — already writes to production paths with correct permissions
- `os.OpenFile` migration — no occurrences in codebase

## Capabilities

### New Capabilities
None — pure implementation fix. No new spec-level behavior.

### Modified Capabilities
None — no existing spec requirements change.

## Approach

Fix ~6 permission constants across 4 files. Safe on Windows (Unix permission bits are ignored). Production file writes stay as-is. Add `os.Chmod` call after `copyFile` in B2 to set `0o600` since `os.Create` uses umask. Test updates: change expected permission constants in assertions.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `adapters/common/installer.go` | Modified | Backup dir mode + Chmod after copy (B1, B2) |
| `adapters/common/strategy.go` | Modified | ReplaceFile backup WriteFile mode (B3) |
| `adapters/codex/installer.go` | Modified | MergeConfig backup WriteFile mode (B4) |
| `adapters/_template/installer.go` | Modified | Template boilerplate fix (T1) |
| `adapters/common/installer_test.go` | Modified | Permission assertion updates |
| `adapters/common/strategy_test.go` | Modified | Permission assertion updates |
| `adapters/codex/installer_test.go` | Modified | Permission assertion updates |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Windows regression | Low | Unix perms are no-ops on Windows; `os.Chmod` only affects write bit |
| Restore logic breaks | Low | Restore writes to production paths — no permission change needed |
| Test assertion mismatch | Low | Expected — update tests to match new constants |

## Rollback Plan

Revert the ~6 changed constants back to original `0o644`/`0o755` values. Remove the added `Chmod` call. `git revert` the commit.

## Dependencies

None. No external libraries, DB changes, or config updates required.

## Success Criteria

- [ ] B1-B4 backup files/dirs use owner-only permissions (`0o600`/`0o700`) on Unix
- [ ] All production file/dir permissions remain unchanged (`0o644`/`0o755`)
- [ ] All existing tests pass with updated permission assertions
- [ ] No new test failures on any platform (Linux, macOS, Windows)
- [ ] `go test -race -count=1 ./...` passes
