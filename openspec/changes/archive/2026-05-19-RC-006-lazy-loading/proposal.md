# Proposal: RC-006 — Lazy Adapter Loading

## Intent

Eliminate double eager construction of adapters. Currently, all 5 adapters (claude, opencode, gemini, cursor, codex) construct their full dependency graph—PathResolver, PromptManager, BackupPathBuilder, Detector, templateFS—twice: once in `init()` → `DefaultRegistry`, then again in `main()` → fresh `Registry`. This wastes CPU, allocates unused memory, and blocks parallelism. Startup time and test isolation suffer.

## Scope

### In Scope
- `Registry.RegisterFactory(id, factory)` — store a lazy constructor, materialize via `sync.Once` on first `Get(id)`
- Migrate each adapter's `RegisterIn` from `Register(concrete)` to `RegisterFactory(id, lazy-fn)`
- Cache `filepath.EvalSymlinks` result in `PathResolver` (runs on every `Base()` call today)
- Migrate all tests from `DefaultRegistry` to fresh registries for isolation
- Deprecate `DefaultRegistry` (already marked deprecated) — stop auto-registration in `init()` paths

### Out of Scope
- Template execution caching (parse is cached, execute is not — future work)
- Lazy `embed.FS` reads during `Install()` — sequential reads remain
- `ToolAdapter` interface changes
- Adapter hot-reload or dynamic discovery

## Capabilities

### New Capabilities
None — pure refactor of construction timing. No new user-visible behavior.

### Modified Capabilities
- **adapter-architecture** — `REQ-DI` (constructor injection): `RegisterFactory` adds a factory variant of registration; `init()` backward compat retained. `REQ-FACTORY` (no convenience wrapper): NOT violated; `RegisterFactory` is registry-internal lazy construction, not a public `NewAdapter(id)` convenience.
- **go-wiring** — Construction timing changes from eager-at-registration to lazy-at-first-use. `main.go` wiring unchanged; all adapters still registered before commands execute.

## Approach

**Factory Registration Pattern** in 4 steps:

1. **Registry**: Add `factory map[string]func() ToolAdapter` + per-key `sync.Once`. `Get(id)` checks factory map first; if factory exists and adapter not constructed, runs factory once. `All()` iterates registered IDs and triggers construction for any unbuilt factories. `Register()` stays eager for backward compat.

2. **Adapters**: Change each `RegisterIn(reg)` — one line change from `reg.Register(newAdapter(""))` to `reg.RegisterFactory("id", func() ToolAdapter { return newAdapter("") })`. The 5 adapters are identical in structure.

3. **PathResolver**: Add `cachedResolved string; cachedResolvedOnce sync.Once` to avoid `filepath.EvalSymlinks` on every `Base()` call. Resolution is immutable per PathResolver instance.

4. **Tests**: Replace `DefaultRegistry` references with `adapters.NewRegistry()` in `cmd/sequoia/main_test.go` (~20 test functions). Add factory-specific tests to `adapters/registry_test.go`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `adapters/registry.go` | Modified | +factory map, +once map, lazy Get/All logic |
| `adapters/{claude,opencode,gemini,cursor,codex}/adapter.go` | Modified | RegisterIn → RegisterFactory (1-line each) |
| `adapters/common/path_resolver.go` | Modified | Cache EvalSymlinks result |
| `cmd/sequoia/main.go` | None | Wiring unchanged |
| `cmd/sequoia/main_test.go` | Modified | DefaultRegistry → fresh registry |
| `adapters/registry_test.go` | Modified | +RegisterFactory tests |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `All()` returns empty before first `Get()` triggers construction | Low | `All()` explicitly triggers pending factories before snapshot |
| `sync.Once` factory panics → adapter permanently unavailable | Low | Factory is `newAdapter("")` — never errors. Add recovery in Get() if needed |
| Parallel tests using `DefaultRegistry` polluted by `init()` adapters | Low | Test migration to fresh registries eliminates this entirely |
| Windows symlink path: `EvalSymlinks` cache may cache wrong value if home changes | Very Low | PathResolver instances are immutable per-adapter; home dir override via `SetHomeDir` clears cache |

## Rollback Plan

Revert each adapter's `RegisterIn` from `RegisterFactory` back to `Register(newAdapter(""))` — one line per adapter. Remove factory map and lazy-Get from Registry. No data migration, no schema changes.

## Dependencies

- **P2-002** through **P2-005**: Performance-focused tasks that should not alter adapter construction paths. Must be completed first to avoid merge conflicts in adapter files.

## Success Criteria

- [ ] `go test -race -count=1 ./...` — full test suite green, zero races
- [ ] Max 5 adapter constructions per CLI invocation (was 10)
- [ ] Adapters not accessed via CLI commands never constructed
- [ ] `Registry.Register()` backward-compatible (eager path preserved)
- [ ] `DefaultRegistry` continues working for external consumers of `adapters` package
