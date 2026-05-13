# FIX-008 — Apply Progress

**Change**: sequoia-fix-FIX-008 — Mover detección de Engram a async en TUI startup
**Mode**: Strict TDD

## Completed Tasks

- [x] Add `EngramDetectedMsg` message type
- [x] Create `detectEngram()` async detection function
- [x] Modify `NewModel()` — remove `exec.LookPath`, init `EngramAvailable: false`
- [x] Modify `Init()` — return `tea.Batch(detectEngram)`
- [x] Add handler in `Update()` for `EngramDetectedMsg`
- [x] Update existing test `TestNewModel_InitReturnsCmd` for new behavior
- [x] Write `TestNewModel_EngramAvailableDefaultsFalse` test
- [x] Write `TestDetectEngram_ReturnsMsg` test (white-box)
- [x] Write `TestEngramDetectedMsg_UpdatesModel` test (white-box)
- [x] Write `TestNewModel_NoExecLookPath` test

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/app/model.go` | Modified | Added `EngramDetectedMsg` type, `detectEngram()` function; removed sync `exec.LookPath` from `NewModel()`; set `EngramAvailable: false`; modified `Init()` to batch `detectEngram` |
| `internal/app/update.go` | Modified | Added `EngramDetectedMsg` handler in `Update()` switch |
| `internal/app/model_test.go` | Modified | Removed `os/exec` import; renamed+updated `TestNewModel_EngramDetection` → `TestNewModel_EngramAvailableDefaultsFalse`; updated `TestNewModel_InitReturnsCmd`; added `TestNewModel_NoExecLookPath` |
| `internal/app/model_internal_test.go` | Modified | Added `TestDetectEngram_ReturnsMsg` and `TestEngramDetectedMsg_UpdatesModel` |

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| EngramDetectedMsg type | `model_internal_test.go` | Unit | ✅ 91/91 | ✅ Written (compile fail) | ✅ Passed | ➖ Single | ✅ Clean |
| detectEngram() function | `model_internal_test.go` | Unit | ✅ 91/91 | ✅ Written (compile fail) | ✅ Passed | ➖ Single path | ✅ Clean |
| NewModel: remove sync call | `model_test.go` | Unit | ✅ 91/91 | ✅ Written (assert false) | ✅ Passed | ✅ 2 tests: default + timing | ✅ Clean |
| Init: async batch | `model_test.go` | Unit | ✅ 91/91 | ✅ Written | ✅ Passed | ✅ Updated existing test | ✅ Clean |
| Update: EngramDetectedMsg handler | `model_internal_test.go` | Unit | ✅ 91/91 | ✅ Written (compile fail) | ✅ Passed | ✅ true + false cases | ✅ Clean |

## Test Summary
- **Total tests written**: 4 new + 1 updated
- **Total tests passing**: 92/92
- **Layers used**: Unit (92)
- **Build**: `go build -o sequoia.exe ./cmd/sequoia` — SUCCESS

## Deviations from Design
None — implementation matches design.

## Issues Found
None.

## Status
All tasks complete. Ready for verify.
