# Tasks: RC-006 — Lazy Adapter Loading

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~235 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR — 4 sequential commits |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

## Phase 1: Registry — Factory Registration (Foundation)

- [ ] 1.1 Test-first: Write `TestRegistry_RegisterFactory_LazyConstruction`, `TestRegistry_RegisterFactory_AllTriggersConstruction`, `TestRegistry_RegisterFactory_ConcurrentGet` in `adapters/registry_test.go`. Verify they FAIL (`go test ./adapters/ -run RegisterFactory -count=1`).
- [ ] 1.2 Implement: Add `factories map[string]func() ToolAdapter` + `onces map[string]*sync.Once` to `Registry` struct in `adapters/registry.go`. Init in `NewRegistry()`. Add `RegisterFactory(id, factory)` method (write-lock). Rewrite `Get(id)` with fast-path (RLock items → return) + factory slow-path (RLock onces → once.Do → Lock items). Rewrite `All()` to trigger pending via `Get(id)` then snapshot. Verify: `go test ./adapters/ -run RegisterFactory -race -count=1` passes.

## Phase 2: Adapter Migration — Register → RegisterFactory

- [ ] 2.1 Change `RegisterIn` in 5 adapter files (`claude`, `opencode`, `gemini`, `cursor`, `codex`): one-line from `reg.Register(newAdapter(""))` to `reg.RegisterFactory("<id>", func() ToolAdapter { return newAdapter("") })`. IDs: `claude-code`, `opencode`, `gemini-cli`, `cursor`, `codex`. Verify: `go test ./adapters/... -count=1` passes (eager `Register()` backward compat via `DefaultRegistry` + `init()`).

## Phase 3: PathResolver — Symlink Cache

- [ ] 3.1 Test-first: Add `TestPathResolver_BaseSymlinkCached`, `TestPathResolver_SetHomeDirResetsSymlinkCache` in `adapters/common/path_resolver_test.go`. Verify they FAIL.
- [ ] 3.2 Implement: Add `cachedResolvedOnce sync.Once`, `cachedResolved string`, `cachedResolvedWarn string` to `PathResolver` struct in `adapters/common/path_resolver.go`. Wrap `ResolveSymlink(homeDir)` in `p.cachedResolvedOnce.Do(...)`. Reset cache in `SetHomeDir` via `p.cachedResolvedOnce = sync.Once{}`. Verify: `go test ./adapters/common/ -race -count=1` passes.

## Phase 4: Test Migration — Fresh Registries

- [ ] 4.1 Add `newPopulatedRegistry(t)` helper in `cmd/sequoia/main_test.go`: creates `adapters.NewRegistry()`, calls `RegisterIn(reg)` for all 5 adapters. Replace all ~17 `adapters.DefaultRegistry` references with `newPopulatedRegistry(t)`. Verify: `go test ./cmd/sequoia/ -race -count=1` passes.

## Phase 5: Final Verification

- [ ] 5.1 `go vet ./...` — zero warnings
- [ ] 5.2 `go build ./...` — clean compilation
- [ ] 5.3 `go test -race -count=1 ./...` — zero races, full suite green
