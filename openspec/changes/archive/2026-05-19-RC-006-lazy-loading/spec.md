# Delta Spec for RC-006 — Lazy Adapter Loading

## MODIFIED Requirements

### REQ-DI [adapter-architecture]: Factory registration alongside eager registration

Registry SHALL expose `RegisterFactory(id string, factory func() ToolAdapter)`. Factory MUST NOT construct until first `Get(id)`. `Register()` SHALL remain eager (backward-compat). `init() → DefaultRegistry` MUST still compile.

(Previously: only `Register(concrete)` existed.)

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Factory stored, not constructed | Empty Registry | `RegisterFactory("claude", fn)` called | Factory stored; `fn()` not invoked |
| Register still eager | Any Registry | `Register(concrete)` called | Adapter immediately available via `Get(id)` |
| RegisterFactory overwrites | `RegisterFactory("a", fn1)` done | `RegisterFactory("a", fn2)` called | `fn2` replaces `fn1`; neither invoked yet |

### REQ-SYMLINK-RESOLVE [symlink-handling]: Cache EvalSymlinks per PathResolver

`PathResolver.Base()` SHALL call `filepath.EvalSymlinks(homeDir)` at most once per instance via `sync.Once`. `SetHomeDir()` SHALL reset cache.

(Previously: resolved on every `Base()` call.)

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Cache hit on repeat call | PathResolver, homeDir is symlink | `Base()` called twice | `EvalSymlinks` called once; both return same resolved path |
| SetHomeDir invalidates cache | PathResolver with cached resolution | `SetHomeDir("/new")` then `Base()` | Fresh `EvalSymlinks` for new homeDir; old warning not re-emitted |
| Non-symlink path still cached | PathResolver, homeDir NOT symlink | `Base()` called twice | `EvalSymlinks` called once; both return same path |

## ADDED Requirements

### REQ-LAZY-CONST [adapter-architecture]: Lazy construction via sync.Once

`Get(id)` SHALL invoke factory (via `sync.Once`) when factory exists and no concrete adapter cached. Subsequent `Get(id)` MUST return same instance. Untouched adapters MUST NOT be constructed.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| First Get triggers construction | `RegisterFactory("gemini", fn)` | `Get("gemini")` | `fn()` invoked once; result cached |
| Second Get reuses cached | `Get("gemini")` already called | `Get("gemini")` again | Returns cached adapter; `fn()` not re-invoked |
| Never accessed, never built | 5 `RegisterFactory` calls | Only `Get("claude-code")` | Only claude factory invoked; 4 others never run |

### REQ-ALL-PENDING [adapter-architecture]: All() materializes pending factories

`All()` SHALL call `Get(id)` for each registered ID before snapshot, triggering pending factories. Already-materialized adapters MUST NOT be re-triggered. Order MUST match registration order.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| All triggers pending | `RegisterFactory("a", fnA)` + `RegisterFactory("b", fnB)` | `All()` | Both `fnA` and `fnB` invoked; 2 concrete adapters returned in order |
| All skips constructed | `Get("a")` already triggered | `All()` | `fnA` NOT re-invoked; `fnB` invoked |
| All after Get returns unified | Mixed: `Get("a")` + `RegisterFactory("b", fnB)` | `All()` | Both concrete adapters present; no factory stubs in result |

### REQ-TEST-MIGRATE [adapter-architecture]: Fresh registries per test

Tests SHALL use `NewRegistry()` per test case. `cmd/sequoia/main_test.go` SHALL not reference `DefaultRegistry`. `t.Parallel()` MUST be race-free.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Parallel test isolation | Two tests, each with `NewRegistry()` | `t.Parallel()` | No shared state; race detector clean |
| main_test uses fresh registry | `main_test.go` | Any test runs | Zero references to `DefaultRegistry` |

### REQ-BACKWARD-COMPAT [go-wiring]: Eager path preserved

`Register()` SHALL remain eager. `init()` blocks in adapter packages SHALL compile and execute unchanged. `DefaultRegistry` SHALL continue working for external consumers.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| init() still compiles | Each adapter package with `init() → DefaultRegistry` | `go build ./...` | Zero compilation errors |
| DefaultRegistry usable externally | External Go module importing `adapters` | Calls `adapters.DefaultRegistry.Get("claude-code")` | Expected adapter or `ErrUnknownAdapter` returned |

## Cross-References to Blocked Tasks

| Task | Status | Relationship |
|------|--------|-------------|
| **P2-002** (init blocks) | Blocked by RC-006 | init() unchanged — factories defer construction; adapter files should not be modified by P2 tasks before RC-006 lands |
| **P2-003** (symlink cache) | Resolved by RC-006 | REQ-SYMLINK-RESOLVE caches `EvalSymlinks` per PathResolver; P2-003 benefits from this |
| **P2-004** (template re-exec) | Out of scope | Template parsing IS cached; execution caching is future work |
| **P2-005** (sequential embeds) | Out of scope | `embed.FS` reads during `Install()` remain sequential; not addressed here |
