# Design: Propagate session file write errors in Codex config merge

## Technical Approach

Two identical one-line changes: replace silent `AtomicWriteFile` error discards with wrapped error returns in `MergeConfig` (`installer.go:36-39`) and `ReplaceFile` (`strategy.go:132-135`). Both callers — `Adapter.Install()` (codex) and `BaseAdapter.Install()` (shared) — already have rollback paths that clean up partial state when these functions return errors. No new rollback logic, no signature changes, no API breakage. The session file is the linchpin of backup/restore; without it, `RemoveConfig`/`RestoreOrRemoveFile` fall back to a single legacy backup name that may be stale, potentially restoring wrong content during uninstall.

## Architecture Decisions

### Decision: Return error vs. warn for session write failure

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Return wrapped error | Install fails cleanly with rollback; admin gets clear message | ✅ Chosen |
| AddWarning via parameter | Install "succeeds" but produces broken uninstall state | ❌ Rejected — worse outcome |
| Convert MergeConfig to Adapter method | Large refactor, same broken uninstall state | ❌ Rejected |

**Rationale**: The session file is NOT advisory. It is the mechanism that guarantees correct backup restoration during uninstall. A clean failed install with rollback is strictly better than an install that "succeeds" with silently corrupted uninstall.

### Decision: Error wrapping style for ReplaceFile

| Option | Tradeoff | Decision |
|--------|----------|----------|
| `fmt.Errorf("replace file: session: %w", err)` | Minor inconsistency with ReplaceFile's bare error returns, but includes function name for log search | ✅ Chosen |
| Bare `return err` | Consistent with existing ReplaceFile pattern, less context for debugging | ❌ |

**Rationale**: Match the wrapping style used in `MergeConfig` (`"merge config: session: %w"`) and `RemoveConfig`'s restore path (`"remove config: restore backup: %w"`). The session write is a distinct sub-operation warranting its own label for error traceability.

## Impact on Callers

| Caller | File | Already handles error? | Rollback? |
|--------|------|----------------------|-----------|
| `Adapter.Install()` | `adapter.go:144-148` | ✅ Yes — rolls back skills + commands, wraps with `ErrInstallFailed` | ✅ |
| `BaseAdapter.Install()` via `writeSystemPrompt` | `base_adapter.go:409-414` | ✅ Yes — rolls back when `rollbackOnSystemPromptError` is true | ✅ |
| `BaseAdapter.Uninstall()` via `removeSystemPrompt` | `base_adapter.go:477` | ✅ Yes — appends to `errs`, wraps with `ErrUninstallFailed` | N/A (uninstall) |

No caller changes needed. All callers already handle error returns from these functions.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `adapters/codex/installer.go:36-39` | Modify | Replace empty block with `return fmt.Errorf("merge config: session: %w", err)` |
| `adapters/common/strategy.go:132-135` | Modify | Replace empty block with `return fmt.Errorf("replace file: session: %w", err)` |
| `adapters/codex/installer_test.go` | Modify | Add `TestMergeConfig_SessionWriteFails` |
| `adapters/common/strategy_test.go` | Modify | Add `TestReplaceFile_SessionWriteFails` |

No files deleted. No new files.

## Testing Strategy

| Layer | Test | Technique |
|-------|------|-----------|
| Unit | `TestMergeConfig_SessionWriteFails` | Create `config.toml.sequoia-session` as a **directory** (causes `os.Rename` to fail on all platforms — cannot replace directory with file). Call `MergeConfig`, assert error is non-nil and contains `"merge config: session"`. |
| Unit | `TestReplaceFile_SessionWriteFails` | Same technique: pre-create `AGENTS.md.sequoia-session` as a directory. Write user content to `AGENTS.md`. Call `ReplaceFile`, assert error is non-nil and contains `"replace file: session"`. |
| Regression | All existing tests | `go test -race -count=1 ./adapters/...` must pass |

**Cross-platform**: The directory-as-session-path technique triggers `os.Rename` failure on every platform. This pattern is already validated by `TestAtomicWriteFile_FailedRenameCleansTemp` in the same test file. No platform-specific skip needed.

## Rollback Plan

Revert two changed lines to their original empty blocks (`// Best-effort: ...`). No schema changes, no config changes, no API changes.

## Open Questions

None. Callers verified, cross-platform test strategy proven by existing test patterns, both error sites share the same root cause.
