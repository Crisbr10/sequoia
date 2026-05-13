# Proposal: Hide Non-Functional Language Selector

## Intent

The Configuration screen shows an EN/ES language selector that does nothing. Every adapter discards `opts.Language` with `_ = opts.Language`. No i18n catalog, `.po` files, or translation infrastructure exists. Users who select "Español" see no change — the selector misleads users into thinking i18n works.

**Fix**: hide the selector from the UI while preserving all internal plumbing for future i18n.

## Scope

### In Scope
- Comment out language field rendering in `ConfigurationView` (lines 43–62) with `TODO(i18n)` markers
- Regenerate 2 golden test files via `UPDATE_GOLDEN=1 go test ./internal/tui/screens/...`
- Ensure `go test ./...` passes with zero regressions

### Out of Scope
- Wiring language to adapters (requires 800–1200 lines of i18n infrastructure)
- Removing `model.Language`, `TUIConfig.Language`, `InstallOpts.Language`, `languageOptions`, `languageIndex`, or adapter `_ = opts.Language` discards
- Modifying `ConfigurationUpdate` key handling or `cycleOption`
- Changing the pipeline pass-through (`internal/pipeline/runner.go`)

## Capabilities

### New Capabilities
None

### Modified Capabilities
None (view-only change — no spec-level behavior change)

## Approach

1. **Comment out rendering**: lines 43–62 in `configuration.go` (language field label + option list)
2. **Add TODO marker**: `// TODO(i18n): Uncomment language rendering when i18n infrastructure is wired`
3. **Regenerate goldens**: `SET UPDATE_GOLDEN=1 && go test ./internal/tui/screens/...`
4. **Keep all plumbing**: `languageOptions`, `languageIndex`, `cycleOption` (lang branch), internal types, pipeline, and adapter code remain untouched

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/tui/screens/configuration.go` | Modified | Comment out language rendering block |
| `internal/tui/screens/testdata/golden/configuration_standard.txt` | Modified | Regenerated — language field removed |
| `internal/tui/screens/testdata/golden/configuration_engram_unavailable.txt` | Modified | Regenerated — language field removed |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Golden file churn on regeneration | Low | Only 2 files; regenerated via documented command |
| Future developer confusion about commented code | Low | Explicit `TODO(i18n)` markers explain intent |

## Rollback Plan

Uncomment lines 43–62 in `configuration.go` and regenerate golden files. Single-commit revert.

## Dependencies

None.

## Success Criteria

- [ ] Language selector no longer renders on Configuration screen
- [ ] `go test ./...` passes with no regressions
- [ ] Golden files updated and match rendered output
- [ ] All internal types (`Language`, `TUIConfig`, `InstallOpts`) and plumbing untouched
