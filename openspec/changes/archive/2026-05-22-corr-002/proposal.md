# Proposal: Global Mutable State Removal (CORR-002)

## Intent

Remove global mutable state — `DefaultRegistry`, `init()` auto-registration, and mutable `CommandFiles` export. These create implicit coupling, dead data (main.go uses its own registry), and fragile test isolation.

## Scope

### In Scope
- Remove `DefaultRegistry` from `adapters/registry.go`
- Delete `init()` functions from 5 adapters + `_template` (claude, codex, cursor, gemini, opencode)
- Fix `CommandFiles` mutability: exported slice → function returning defensive copy
- Update all tests using `DefaultRegistry` → `NewRegistry()` (registry_test.go, model_test.go, ~15 CommandFiles references)
- Update `CONTRIBUTING.md` adapter creation guide

### Out of Scope
- P3-009 (templateCache) — already resolved (no cache exists; isolation tests present)
- P4-008 (test import pollution) — auto-resolved by DefaultRegistry removal
- P4-009 (error wrapping standardization)
- Changing `RegisterIn()` signatures or registry API

## Capabilities

### New Capabilities
- **adapter-registration**: Explicit DI-based adapter registration via `NewRegistry()` + `RegisterIn()`; no import-time side effects

### Modified Capabilities
None — this is a pure refactor of registration mechanism; no spec-level behavior changes.

## Approach

1. **Remove `DefaultRegistry`**: Delete the `var DefaultRegistry = NewRegistry()` line and its doc comment from `adapters/registry.go`.
2. **Delete `init()` blocks**: Remove `init()` from all 6 adapter packages. Keep `RegisterIn()` functions (already used by main.go).
3. **Fix `CommandFiles`**: Replace `var CommandFiles = []string{...}` with `func CommandFiles() []string` returning a copy. Update all call sites (~15 files).
4. **Update tests**: Replace `adapters.DefaultRegistry` → `adapters.NewRegistry()` in registry_test.go and model_test.go. Add explicit mock registrations where needed.
5. **Update docs**: Rewrite CONTRIBUTING.md Step 2 to show `RegisterIn()` + `NewRegistry()` pattern instead of `init()` + `DefaultRegistry`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `adapters/registry.go` | Modified | Remove DefaultRegistry variable |
| `adapters/{claude,codex,cursor,gemini,opencode,_template}/adapter.go` | Modified | Remove init() from each |
| `adapters/registry_test.go` | Modified | Replace DefaultRegistry with NewRegistry() |
| `internal/app/model_test.go` | Modified | Replace adapters.DefaultRegistry with NewRegistry() + mock reg |
| `adapters/common/commands.go` | Modified | Change CommandFiles from var to func |
| `adapters/common/base_adapter.go` | Modified | Update CommandFiles references |
| ~15 adapter test files | Modified | Update `common.CommandFiles` → `common.CommandFiles()` |
| `CONTRIBUTING.md` | Modified | Update Step 2 documentation |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Test regressions from DefaultRegistry removal | Medium | Full test suite (`go test ./...`) validates; existing tests well-isolated with t.TempDir() |
| Breaking downstream consumers | Low | DefaultRegistry already marked deprecated; main.go uses explicit DI |
| CommandFiles change breaks adapter templates | Low | All adapters use the same embed.FS pattern; change is mechanical |

## Rollback Plan

Single commit revert restores `DefaultRegistry` variable, `init()` functions, and `var CommandFiles`. No database or config migration involved.

## Dependencies

- None. CORR-002 is a root node. Unblocks P3-001, P1-004, P3-008, P4-008.

## Success Criteria

- [ ] `go test ./...` passes with no DefaultRegistry references
- [ ] `grep -r "DefaultRegistry" --include="*.go"` returns zero results (except deprecated docs)
- [ ] `grep -r "func init()" adapters/*/adapter.go` returns zero results
- [ ] `CommandFiles` is a function, not a mutable exported slice
- [ ] CONTRIBUTING.md shows explicit DI pattern
