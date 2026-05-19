# Proposal: Propagate session file write errors in Codex config merge

## Intent

Fix a HIGH-severity silent error discard in `MergeConfig` (`adapters/codex/installer.go:36-39`) and `ReplaceFile` (`adapters/common/strategy.go:132-135`). Both functions silently discard `AtomicWriteFile` errors when writing session-tracking files. When session writes fail, `RemoveConfig`/`RestoreOrRemoveFile` fall back to a single legacy backup name that may be stale, restoring wrong content during uninstall. The fix returns these errors so installs fail cleanly with proper rollback instead of silently producing broken uninstall state.

## Scope

### In Scope
- Return wrapped error from `MergeConfig` when session file write fails
- Return error from `ReplaceFile` when session file write fails
- Add test `TestMergeConfig_SessionWriteFails` in `adapters/codex/installer_test.go`
- Add test `TestReplaceFile_SessionWriteFails` in `adapters/common/strategy_test.go`

### Out of Scope
- Changing the backup mechanism or session file format
- Refactoring `MergeConfig` into an `Adapter` method
- Modifying the `RemoveConfig`/`RestoreOrRemoveFile` fallback chain
- Adding `AddWarning`-based recovery path

## Capabilities

### New Capabilities
None

### Modified Capabilities
None

## Approach

**Approach 1 from exploration — return the error.** Replace the empty error block at both locations with a wrapped error return.

**`installer.go:36-39`** → `return fmt.Errorf("merge config: session: %w", err)`

**`strategy.go:132-135`** → `return err` (caller context is sufficient; `ReplaceFile` is not wrapping errors today)

**Why this is correct**: The session file is the linchpin of the backup/restore mechanism. Without it, `RemoveConfig` cannot identify the correct timestamped backup and falls back to a single legacy name that can be stale. A clean failed install with rollback is strictly better than an install that "succeeds" with silently broken uninstall. `Adapter.Install()` already handles `MergeConfig` errors with rollback — no new rollback logic needed.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `adapters/codex/installer.go:36-39` | Modified | Return error instead of silent discard |
| `adapters/common/strategy.go:132-135` | Modified | Return error instead of silent discard |
| `adapters/codex/installer_test.go` | New test | `TestMergeConfig_SessionWriteFails` |
| `adapters/common/strategy_test.go` | New test | `TestReplaceFile_SessionWriteFails` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Transient disk-full now fails installs that previously "succeeded" broken | Low | Clear error message with retry path; cleaner than silent data corruption |
| ReplaceFile change affects all StrategyFileReplace adapters | Low | Same logic applies universally; all callers have rollback |

## Rollback Plan

Revert the two changed lines to empty blocks. No schema changes, no config changes, no API changes.

## Dependencies

None

## Success Criteria

- [ ] `TestMergeConfig_SessionWriteFails` passes — confirms error propagates when session write fails
- [ ] `TestReplaceFile_SessionWriteFails` passes — confirms error propagates when session write fails
- [ ] All existing tests pass (`go test -race -count=1 ./adapters/...`)
- [ ] `go build ./...` succeeds
