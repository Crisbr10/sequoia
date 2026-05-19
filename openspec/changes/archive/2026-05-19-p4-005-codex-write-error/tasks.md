# Tasks: Propagate session file write errors in Codex config merge

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~50–70 (4 production + ~50 test + tasks.md) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |
| Chain strategy | not applicable |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: not applicable
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Fix silent error discards + add tests + verify | Single PR | ~60 lines total, well under 400-line budget |

## Phase 1: Production Fix

- [x] 1.1 Fix `MergeConfig` session error propagation — Replace empty `if err != nil` block at `adapters/codex/installer.go:36-39` with `return fmt.Errorf("merge config: session: %w", err)`
- [x] 1.2 Fix `ReplaceFile` session error propagation — Replace empty `if err != nil` block at `adapters/common/strategy.go:132-135` with `return fmt.Errorf("replace file: session: %w", err)`

## Phase 2: Tests (TDD — RED first, then GREEN)

- [x] 2.1 Add `TestMergeConfig_SessionWriteFails` in `adapters/codex/installer_test.go` — Pre-create `.sequoia-session` as a directory (forces `os.Rename` failure on all platforms). Call `MergeConfig`, assert error is non-nil and contains `"merge config: session"`. Reference: spec scenario "Session file write fails during config merge"
- [x] 2.2 Add `TestReplaceFile_SessionWriteFails` in `adapters/common/strategy_test.go` — Pre-create `AGENTS.md.sequoia-session` as a directory. Write user content to `AGENTS.md`. Call `ReplaceFile`, assert error is non-nil and contains `"replace file: session"`. Reference: spec scenario "Session file write fails during file replace"

## Phase 3: Verification

- [x] 3.1 Run full test suite — `go test -race -count=1 ./adapters/...` — confirm all existing tests pass and new tests pass. Also run `go build ./...` to confirm compilation
