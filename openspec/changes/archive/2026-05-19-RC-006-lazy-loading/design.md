# Design: Lazy Adapter Loading

## Technical Approach

Replace eager adapter construction with a factory registration pattern. Adapters register a `func() ToolAdapter` at `init()` time; construction defers to first `Get()` or `All()`, guarded by per-key `sync.Once`. The `Register()` method remains for backward compat. PathResolver caches its `EvalSymlinks` call via a second `sync.Once`.

## Architecture Decisions

| Decision | Choice | Rejected | Rationale |
|----------|--------|----------|-----------|
| Factory contract | Factory is pure constructor, does NOT call Register | Factory calls Register internally | Avoids lock-ordering deadlock: Get() must not hold RLock while factory acquires Lock via Register |
| Lazy trigger point | `Get()` and `All()` both trigger construction | Only `Get()` triggers | `All()` must never return empty when factories are registered; commands like `status` and `ScanTools` use `All()` directly |
| `List()` method | NOT added | Add `List()` that returns IDs without constructing | No current consumer; `All()` already serves status/scan. YAGNI |
| Once storage | `map[string]*sync.Once` alongside `factories` | Composite struct per key | Separate maps match existing pattern (items+order); pointer avoids `sync.Once` copy prohibition |
| PathResolver cache reset | `SetHomeDir()` reassigns zero-value `sync.Once{}` | Mutable flag + invalidation | `sync.Once{}` is a struct; assignment creates fresh Once. No new fields needed |
| Test isolation | Fresh `NewRegistry()` + explicit `RegisterIn(reg)` per adapter | Keep `DefaultRegistry` with lazy factories | Eliminates any init()-ordering risk; tests control exactly what adapters exist |

## Data Flow

```
init()                          Get("claude-code")              All()
  │                                  │                            │
  ▼                                  ▼                            ▼
RegisterFactory(id, factory)    RLock items[id]               RLock onces → collect pending
  │                             ┌─found?──▶ return            ▼
  ▼                             │ not found                once.Do(factory) per pending
factories[id] = factory         ▼                               │
onces[id] = &sync.Once{}      RLock onces[id]                 ▼
order = append(id)             ┌─nil?──▶ ErrUnknownAdapter   RLock items + order
                               │ found                       ▼
                               ▼                           return snapshot
                            once.Do():
                              adapter = factory()
                              Lock: items[id] = adapter
                              Unlock
                               │
                               ▼
                            RLock items[id] → return
```

*After `once.Do()`, the adapter is in `items` — subsequent `Get()` hits the fast path (RLock only).*

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `adapters/registry.go` | Modify | Add `factories map[string]func() ToolAdapter`, `onces map[string]*sync.Once`. Add `RegisterFactory` method. Rewrite `Get()` with fast-path + factory slow-path. Rewrite `All()` to trigger pending factories before snapshot |
| `adapters/claude/adapter.go` | Modify | `RegisterIn`: change `reg.Register(newAdapter(""))` → `reg.RegisterFactory("claude-code", func() ToolAdapter { return newAdapter("") })` |
| `adapters/opencode/adapter.go` | Modify | Same 1-line change, ID `"opencode"` |
| `adapters/gemini/adapter.go` | Modify | Same 1-line change, ID `"gemini-cli"` |
| `adapters/cursor/adapter.go` | Modify | Same 1-line change, ID `"cursor"` |
| `adapters/codex/adapter.go` | Modify | Same 1-line change, ID `"codex"` |
| `adapters/common/path_resolver.go` | Modify | Add `cachedResolvedOnce sync.Once`, `cachedResolved string`, `cachedResolvedWarn string` fields. Wrap `ResolveSymlink` call in `cachedResolvedOnce.Do()`. `SetHomeDir` resets the Once |
| `cmd/sequoia/main_test.go` | Modify | Replace `adapters.DefaultRegistry` with `newPopulatedRegistry(t)` helper that calls `RegisterIn` for all 5 adapters into a fresh `NewRegistry()`. ~20 call sites |
| `adapters/registry_test.go` | Modify | Add `TestRegistry_RegisterFactory_LazyConstruction`, `TestRegistry_RegisterFactory_AllTriggersConstruction`, `TestRegistry_RegisterFactory_ConcurrentGet` |

## Interfaces / Contracts

```go
// RegisterFactory stores a lazy constructor. The factory MUST be a pure function:
// it returns a new ToolAdapter without side effects (no Register calls).
// Construction happens once per id, on first Get() or All().
func (r *Registry) RegisterFactory(id string, factory func() ToolAdapter)

// Get triggers lazy construction if a factory exists for id.
// Panics from factory are NOT recovered (factories are newAdapter("") → never panics).
func (r *Registry) Get(id string) (ToolAdapter, error)

// All triggers all pending factories before returning the snapshot.
func (r *Registry) All() []ToolAdapter
```

**Thread safety contract**: `Register()` and `RegisterFactory()` acquire write lock. `Get()` holds read lock for map lookups, releases before calling `once.Do()`. Inside `once.Do()`, write lock acquired only for `items[id] = adapter`. This prevents deadlock: no lock is held during factory execution.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `RegisterFactory` + `Get` lazy construction | `registry_test.go`: register factory, verify `Get` triggers construction exactly once (counter in mock) |
| Unit | `All()` triggers pending factories | Register 3 factories, call `All()`, verify all 3 constructed, second `All()` calls factory 0 times |
| Unit | Concurrent `Get` on same factory | 100 goroutines, same id, factory counter = 1 |
| Unit | `Register()` backward compat (eager path) | Existing tests pass unchanged |
| Unit | PathResolver caches symlink | Call `Base()` twice, verify `EvalSymlinks` called once (mock filesystem or counter) |
| Integration | CLI commands with lazy registry | `main_test.go`: fresh registry + RegisterIn for all adapters, verify `status`, `install --help`, etc. |
| Race | Full suite with `-race` | `go test -race -count=1 ./...` — zero races |

## Migration / Rollout

No feature flags. Backward compat:

1. **init() preserved**: All 5 adapters still call `RegisterIn(DefaultRegistry)` in `init()`. External consumers importing adapter packages see no change.
2. **Eager `Register()` preserved**: Direct `reg.Register(adapter)` calls continue to work for code that constructs adapters manually.
3. **DefaultRegistry**: Still populated with factories via `init()`; adapters construct lazily on first use via DefaultRegistry too.
4. **Rollback**: Revert each adapter's `RegisterIn` from `RegisterFactory` back to `Register(newAdapter(""))`. Remove `factories`/`onces` from Registry, restore original `Get()`/`All()`.

## Open Questions

None — all design decisions have clear rationale and the implementation pattern is well-understood.
