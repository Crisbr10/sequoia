# Design: Global Mutable State Removal (CORR-002)

## Technical Approach

Mechanical refactor removing three sources of global mutable state: `DefaultRegistry` (replaced by explicit `NewRegistry()` DI), `init()` auto-registration (replaced by caller-driven `RegisterIn()`), and mutable `CommandFiles` slice (converted to function returning defensive copy). Strict TDD ordering: tests refactored first, then production code, then docs.

## Architecture Decisions

| Decision | Choice | Tradeoff | Rationale |
|----------|--------|----------|-----------|
| DefaultRegistry removal | Delete variable, keep `NewRegistry()` | Callers must create their own registry | `main.go` already uses explicit DI; `DefaultRegistry` is dead data |
| init() removal | Delete all 6 `init()` blocks, keep `RegisterIn()` | No auto-registration on import | `RegisterIn()` is the DI entry point already consumed by `main.go` |
| CommandFiles immutability | `func CommandFiles() []string` returning `append([]string{}, internal...)` | One allocation per call; negligible (6 strings, called once per install) | Prevents mutation by callers; compile-time enforcement via missing `()` |
| Test migration strategy | Mechanical `s/DefaultRegistry/NewRegistry()/` in model_test.go; explicit registries in registry_test.go | ~28 model_test.go + 3 registry_test.go replacements | No behavioral change — most tests already use explicit registries |
| main.go imports | Keep named imports, update comment only | No code change needed | Imports already named and used by explicit `RegisterIn(reg)` calls |

## Data Flow

```
Before:
  import "adapters/claude" ──→ init() ──→ DefaultRegistry.Register()
  import "adapters/cursor" ──→ init() ──→ DefaultRegistry.Register()
  ...
  main.go: adapters.DefaultRegistry.All()  (dead data — never consumed)

After:
  main.go: reg := adapters.NewRegistry()
           claude.RegisterIn(reg)     ← explicit caller-driven DI
           cursor.RegisterIn(reg)     ← no import-time side effects
           reg.All()                  ← single source of truth
```

## CommandFiles Migration

```
Before: var CommandFiles = []string{"sequoia-init.md", ...}  // mutable export
        range common.CommandFiles       // bare reference
        Files: common.CommandFiles      // passes mutable slice

After:  func CommandFiles() []string { return append([]string{}, files...) }  // defensive copy
        range common.CommandFiles()     // function call
        Files: common.CommandFiles()    // each caller gets own copy
```

Internal backing: `var commandFiles = []string{"sequoia-init.md", ...}` (unexported). 22 call sites across 8 files need mechanical `()` addition.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `adapters/registry.go` | Modify | Delete `DefaultRegistry` variable (line 44) and its doc comment (lines 40-43) |
| `adapters/claude/adapter.go` | Modify | Delete `init()` block (lines 27-29) |
| `adapters/codex/adapter.go` | Modify | Delete `init()` block (lines 30-32) |
| `adapters/cursor/adapter.go` | Modify | Delete `init()` block (lines 25-27) |
| `adapters/gemini/adapter.go` | Modify | Delete `init()` block (lines 31-33) |
| `adapters/opencode/adapter.go` | Modify | Delete `init()` block (lines 27-29) |
| `adapters/_template/adapter.go` | Modify | Delete `init()` block (lines 43-45) |
| `adapters/common/commands.go` | Modify | Convert `var CommandFiles` to `func CommandFiles() []string` with unexported backing slice |
| `adapters/common/base_adapter.go` | Modify | Change 3 `CommandFiles` refs to `CommandFiles()` (lines 323, 386, 466) |
| `adapters/_template/adapter.go` | Modify | Change 3 `common.CommandFiles` refs to `common.CommandFiles()` (lines 189, 226, 271) |
| `adapters/codex/adapter.go` | Modify | Change 3 `common.CommandFiles` refs to `common.CommandFiles()` |
| `adapters/common/command_template_test.go` | Modify | Change 2 `common.CommandFiles` refs to `common.CommandFiles()` |
| `adapters/common/base_adapter_test.go` | Modify | Change 4 `common.CommandFiles` refs to `common.CommandFiles()` |
| `adapters/common/base_adapter_error_test.go` | Modify | Change 8 `common.CommandFiles` refs to `common.CommandFiles()` |
| `adapters/common/base_adapter_mockfs_test.go` | Modify | Change 1 `common.CommandFiles` ref to `common.CommandFiles()` |
| `adapters/common/shared_test.go` | Modify | Change `common.CommandFiles` comparison to `common.CommandFiles()` (line 36) |
| `internal/app/model_test.go` | Modify | Replace 28 `adapters.DefaultRegistry` with `adapters.NewRegistry()` |
| `adapters/registry_test.go` | Modify | Replace 4 `adapters.DefaultRegistry` refs with explicit `adapters.NewRegistry()` + manual registration |
| `cmd/sequoia/main.go` | Modify | Update import comment (line 22): remove "init() +" |
| `CONTRIBUTING.md` | Modify | Step 2: replace `init()` + `DefaultRegistry.Register()` with `RegisterIn()` + `NewRegistry()` pattern; Step 7: remove blank import instruction |

## Interfaces / Contracts

No API signature changes. `RegisterIn(reg *adapters.Registry)` remains the public DI entry point. `ToolAdapter` interface unchanged. `NewRegistry()` remains the sole constructor.

```go
// NEW: CommandFiles() returns a defensive copy of the shared command list.
func CommandFiles() []string {
    return append([]string{}, commandFiles...)
}
```

## Testing Strategy

| Layer | What | How |
|-------|------|-----|
| TDD Phase 1 — test refactor | Replace all `adapters.DefaultRegistry` refs in test files | Mechanical `s/DefaultRegistry/NewRegistry()/` in model_test.go; explicit `NewRegistry()` + `Register()` in registry_test.go; `CommandFiles` → `CommandFiles()` in common tests |
| TDD Phase 2 — production | Remove `DefaultRegistry`, delete `init()`, convert `CommandFiles` | Sequential: registry.go first (no deps), then adapters (depend on registry), then common (independent) |
| Regression | Full test suite | `go test ./...` — verify zero DefaultRegistry references pass; `go vet ./...` — verify no issues |
| Immutability | Defensive copy verification | New unit test: mutate result of `CommandFiles()`, verify next call returns original list |

## Migration / Rollout

**Strict TDD ordering (tests before production):**

1. **model_test.go**: Replace all `adapters.DefaultRegistry` → `adapters.NewRegistry()`. Tests that need adapters already use explicit registries — no behavioral change.
2. **registry_test.go**: Replace `DefaultRegistry` references with caller-owned `NewRegistry()` + manual registration.
3. **shared_test.go + all CommandFiles test refs**: Change `common.CommandFiles` → `common.CommandFiles()` in all test files. Phase 1 tests now fail (production code still has `var CommandFiles`).
4. **commands.go**: Convert `var CommandFiles` → `func CommandFiles()`. Phase 1 tests now pass.
5. **All adapter CommandFiles call sites**: Add `()` to 22 references across 7 production files.
6. **registry.go**: Delete `DefaultRegistry`.
7. **6 adapter packages**: Delete `init()` blocks.
8. **main.go**: Update comment only.
9. **CONTRIBUTING.md**: Update Step 2 and Step 7.
10. **Final verification**: `go test ./...` + `go vet ./...` + `grep -r "DefaultRegistry" --include="*.go"` (expect zero in production, zero in tests).

Rollback: single commit revert restores all removed code. No DB or config migration needed.

## Open Questions

None. All decisions resolved.
