# Tasks: RC-001 — Decompose BaseAdapter

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~1,400-1,600 (23 files: 7 new, 15 modified, 1 deleted) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | 6 PRs |

## Phase 1: PathResolver extraction ✅
- [x] 1.1-1.6: PathResolver extracted, 13 tests, 18/18 pass

## Phase 2: Detector extraction ✅
- [x] 2.1-2.5: Detector extracted, 7 tests, 18/18 pass

## Phase 3: PromptManager + BackupPathBuilder extraction ✅
- [x] 3.1-3.5: PromptManager extracted (10 tests), BackupPathBuilder extracted (4 tests), all adapters updated

## Phase 4: Interface segregation ✅
- [x] 4.1-4.4: ToolAdapter segregated into role interfaces, composite preserved

## Phase 5: DI refactor ✅
- [x] 5.1-5.8: NewRegistry(), RegisterIn() on all adapters, *Registry DI, factory.go deleted, 26+ tests updated

## Phase 6: Error-path tests ✅
- [x] 6.1 Context cancellation at 5 checkpoints (6 tests)
- [x] 6.2 Nil function field safety (5 tests)
- [x] 6.3 Staging dir creation failure (1 test)
- [x] 6.4 Template rendering failure (2 tests)
- [x] 6.5 Base() resolution failure (2 tests)
- [x] 6.6 Version file write failure (1 test)
- [x] 6.7 System prompt failure with/without rollback (3 tests)

## Phase 7: Mock FS template tests ✅
- [x] 7.1 Mock FS template rendering (4 tests: installembed.FS)
- [x] 7.2 Per-adapter template tests (9 tests: Claude 4, Codex 5)
- [x] 7.3 Full Install pipeline with mock FS (2 tests)

## Phase 8: Final verification ✅
- [x] 8.1 `go build ./...` — clean, no errors
- [x] 8.2 `go test ./... -count=1` — 18/18 packages pass
- [x] 8.3 Coverage: 84.6% (exceeds 80% threshold)
- [x] 8.4 `go vet ./...` — clean, no warnings
- [x] 8.5 All 8 spec requirements satisfied (15/16 scenarios compliant, 1 partial)

**Result**: 44/44 tasks complete across 8 phases. Verdict: PASS WITH WARNINGS (0 CRITICAL issues).
