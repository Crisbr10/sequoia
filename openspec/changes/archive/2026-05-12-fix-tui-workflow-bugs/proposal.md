# Proposal: Fix TUI Workflow Bugs

## Intent

Fix 11 bugs discovered in the Sequoia TUI installer workflow during sdd-explore analysis. Bugs span install/uninstall labeling, pipeline setup gaps, configuration navigation, uninstall UX, and dead code. All are implementation deviations from spec behavior in `openspec/changes/tui-installer/specs/`.

## Scope

### In Scope
- **Group A** (Bugs 2,4,5): Add `OperationMode` field; make InstallProgress, Complete, Error views mode-aware ("Installing"/"Uninstalling")
- **Group B** (Bugs 3,10): Extract `startPipeline(mode string)` shared method fixing Status→reinstall and Error→retry to rebuild pipeline + ProgressTools
- **Group C** (Bug 1): Remap Configuration Up/Down from `toggleField()` to `cycleOption(direction)`; update footer hints
- **Group D** (Bugs 6,7,8,11): Esc exits confirmation mode; `errorMsg` rendered in UninstallView; `ErrorMsg` set on invalid uninstall; Esc hint in footer; Uninstall "back" respects source screen (Status vs Welcome)
- **Group E** (Bug 9): Remove/comment dead `'q'` handlers from ToolSelectionUpdate, ConfigurationUpdate, InstallProgressUpdate
- Test coverage gaps from exploration: configuration test assertions, reinstall/retry pipeline tests, `startPipeline` unit tests

### Out of Scope
- New features or capability additions
- Changes to pipeline goroutine internals (`pipeline/` package)
- UI redesign beyond spec corrections

## Capabilities

### New Capabilities
None — bugfix-only change.

### Modified Capabilities
- **tui-install-flow**: InstallProgress, Complete, Error views gain `operationMode` parameter; Configuration Up/Down remapped
- **tui-management**: UninstallView gains `errorMsg` rendering; Esc exits confirmation; "back" respects source screen; Status→reinstall builds pipeline
- **tui-core**: `Model.OperationMode` field added; screen state machine tracks previous screen for back-navigation; dead `'q'` handlers removed
- **tui-pipeline**: Shared `startPipeline` guarantees pipeline setup on ALL entry paths to InstallProgress

## Approach

Follow five ordered groups implementing RED→GREEN→REFACTOR per group:

1. **Group C first** (Configuration nav): Smallest, least coupled. Change Up/Down mapping, update tests, update hints.
2. **Group A** (Mode labeling): Add `OperationMode` to Model. Thread through InstallProgressView, CompleteView, ErrorView. Set mode on Configuration confirm and Uninstall confirm.
3. **Group B** (Pipeline consolidation): Extract `startPipeline(mode)` on Model that builds ProgressTools, resets counters, starts pipeline, returns `tea.Batch`. Call from Configuration confirm, Status reinstall, Error retry, Uninstall confirm.
4. **Group D** (Uninstall UX): Add `errorMsg` to UninstallView, Esc in confirmation, Esc hint, source-aware back navigation. Bug 6 fix: add `ScreenStatus` to `TransitionMap[ScreenUninstall]`; use model field to track origin.
5. **Group E** (Dead code): Remove/comment 'q' handlers — no behavioral change, skip tests.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/app/model.go` | Modified | Add `OperationMode` field |
| `internal/app/update.go` | Modified | Extract `startPipeline`, fix reinstall/retry/uninstall, source tracking |
| `internal/tui/screens/configuration.go` | Modified | Up/Down→cycleOption, footer hints, dead 'q' |
| `internal/tui/screens/install-progress.go` | Modified | Mode-aware title/summary |
| `internal/tui/screens/complete.go` | Modified | Mode-aware heading |
| `internal/tui/screens/error.go` | Modified | Mode-aware heading |
| `internal/tui/screens/uninstall.go` | Modified | errorMsg param, Esc hint |
| `internal/tui/screens/tool-selection.go` | Modified | Dead 'q' removal |
| `internal/tui/router.go` | Modified | Add ScreenStatus to Uninstall transitions |
| `internal/tui/screens/configuration_test.go` | Modified | Update assertions |
| `internal/app/integration_test.go` | Modified | Add reinstall/retry pipeline tests |
| `internal/app/model_internal_test.go` | New | `startPipeline` unit tests |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `OperationMode` threading breaks existing install flow | Low | Field defaults to `""` (zero value); only set at entry points; views fall back to "Installing" for unknown modes |
| `startPipeline` extraction regresses Configuration confirm | Low | Keep existing tea.Batch composition; extract as method with identical return; verify with existing integration test |
| Source tracking for Uninstall "back" (Bug 6) adds coupling | Med | Add single `PreviousScreen` field set on NavigateMsg dispatch; reset on explicit navigation; test both paths |
| Test updates miss assertions | Low | Run `go test -race -count=1 ./...` after each group; TDD ensures coverage |
| Coverage threshold not met for new file | Low | `model_internal_test.go` tests `startPipeline` exhaustively with table-driven cases |

## Rollback Plan

1. `git revert` the single commit for this change
2. If partial revert needed: revert Group D (Uninstall) first as it has the most coupling, then Group B (pipeline), then Group C (config), then Group A (mode), then Group E (dead code)
3. `go test -race -count=1 ./...` after each revert to verify

## Dependencies

- `openspec/changes/tui-installer/specs/` (active change specs — not yet archived)
- Existing test suite: `go test -race -count=1 ./...` must pass before starting

## Success Criteria

- [ ] All 11 bugs fixed and verified against source
- [ ] `go test -race -count=1 ./...` passes with coverage ≥80%
- [ ] Configuration Up/Down cycles options within active field
- [ ] InstallProgress shows "Uninstalling" during uninstall flow
- [ ] Complete shows "Uninstallation Complete" after uninstall
- [ ] Status→reinstall rebuilds pipeline and shows correct progress
- [ ] Error→retry rebuilds pipeline and shows correct progress
- [ ] Uninstall confirmation handles Esc (exits confirmation)
- [ ] Uninstall shows error when Enter pressed with no valid tools
- [ ] Uninstall Esc returns to source screen (Status or Welcome)
- [ ] No dead 'q' handlers remain in screen files
