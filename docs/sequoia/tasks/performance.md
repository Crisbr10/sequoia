# Performance Tasks — Sequoia Audit

## Context

Sequoia is a Go CLI tool where performance is measured in startup latency, install/uninstall throughput, and memory footprint. The tool initializes all adapters eagerly at import time and performs blocking I/O during installation. Performance findings exclude P2-001 (mapped to Architecture as a correctness bug).

9 findings: 4 medium, 3 low, 2 info. Root cause RC-006 (eager adapter initialization) drives the medium findings P2-002 through P2-005.

## Priority Tiers

### Tier 1 — Immediate (critical + high)

| ID | Task | Effort | Blocks |
|----|------|--------|--------|
| RC-006 | Implement lazy adapter loading — defer construction to first use | medium | P2-002, P2-003, P2-004, P2-005 |

### Tier 2 — Short Term (medium)

| ID | Task | Effort | Blocks |
|----|------|--------|--------|
| P2-002 | Profile and reduce adapter init() overhead | small | — |
| P2-003 | Cache symlink resolution result per adapter instance | small | — |
| P2-004 | Add template rendering output cache (content-addressed) | small | — |
| P2-005 | Batch-read command templates from embed.FS | small | — |

### Tier 3 — Long Term (low + info)

| ID | Task | Effort | Blocks |
|----|------|--------|--------|
| P2-006 | Consolidate redundant os.Stat calls in verification path | small | — |
| P2-007 | Parallelize file copy operations in Installer.Apply() | medium | — |
| P2-008 | Add pprof HTTP endpoint (dev mode only) for profiling | small | — |
| P2-009 | Stream version file reads instead of loading entire file | small | — |
| P2-010 | Add benchmark suite (`go test -bench=.`) to CI | small | — |

## Detailed Tasks

### RC-006 — Eager Adapter Initialization Couples All Adapters to Main
- **Severity**: HIGH (drives 5 child findings)
- **Evidence**: `adapters/registry.go:20` — `DefaultRegistry` is initialized at package level. Each adapter's `init()` function calls `DefaultRegistry.Register()` at import time (e.g., `adapters/claude/adapter.go` init block). All adapters are loaded regardless of which tools the user has installed.
- **Problem**: Startup latency scales linearly with adapter count. Unused adapters consume memory for their embed.FS and template data. `init()` ordering dependencies create non-deterministic behavior when adapters reference shared infrastructure.
- **Fix**:
  1. Create a lazy `AdapterLoader` that stores only metadata (ID, name, detect function) at init time
  2. Defer full adapter construction (BaseAdapter, template loading, path resolution) to first call to `NewAdapter()` or `Registry.Get()`
  3. Remove `init()` functions from adapter packages — replace with explicit `Register()` calls in `cmd/sequoia/main.go`
  4. Add `sync.Once` guards to constructed adapters to ensure thread-safe lazy init
  5. Benchmark startup time before/after: `hyperfine "sequoia --help"`
- **Verification**: Run `hyperfine "sequoia --help"` before and after. Target: <50ms reduction in startup time. Verify that adapters for tools not installed on the system are never fully constructed. Confirm `go test -race ./...` passes (lazy init must be thread-safe).
- **References**: Effective Go — Initialization, Go memory model for sync.Once, Composition Root pattern

### P2-002 — Adapter init() Blocks Exist for All Adapters
- **Severity**: MEDIUM
- **Evidence**: Every adapter package (claude, codex, cursor, gemini, opencode) contains an `init()` function that calls `adapters.DefaultRegistry.Register()`. This runs during program startup before `main()`.
- **Problem**: init() execution is serialized and blocks startup. With 6+ adapters, each initializing templates and path functions, startup can take hundreds of milliseconds unnecessarily. Tools the user doesn't have installed still consume startup time and memory.
- **Fix**: Subsumed by RC-006 (lazy adapter loading). After RC-006 refactoring, verify that init() blocks are removed or reduced to lightweight metadata registration only.
- **Verification**: Add a test that imports the main package and checks `time.Since(start)` for import time. Should be <10ms.
- **References**: Go FAQ — "Why is my program slow to start?"

### P2-003 — Symlink Resolution on Every Base() Call
- **Severity**: MEDIUM
- **Evidence**: `adapters/common/base_adapter.go:204` — `resolved, warning := ResolveSymlink(homeDir)` is called every time `Base()` is invoked, including from `SkillsPath()`, `CommandsPath()`, `SystemPromptPath()`, `IsInstalled()`, `Status()`. User home symlinks don't change during a process lifetime.
- **Problem**: Unnecessary filesystem calls on every adapter status check. Each `Status()` call (called once per adapter in the TUI pipeline) resolves the same symlink. Multiply by 6+ adapters for a noticeable cumulative cost.
- **Fix**:
  1. Cache the resolved home directory in BaseAdapter alongside `cachedHomeDir`
  2. Resolve symlinks once during first `Base()` call, store in `cachedResolvedHome`
  3. Use `sync.Once` to ensure thread-safe caching
- **Verification**: Add a test that calls `Base()` 100 times and asserts `ResolveSymlink` is only called once (inject a counting wrapper). Benchmark shows O(1) not O(n) for repeated calls.
- **References**: Effective Go — Avoiding Repeated Work with sync.Once

### P2-004 — Template Rendering Re-executes on Every Install
- **Severity**: MEDIUM
- **Evidence**: `adapters/common/base_adapter.go:329` — `RenderTemplate(a.templateFS, "templates/skill.md.tmpl", data)` is called on every `Install()`. The template is parsed once (cached in `templateCache`) but executed against data on every call, even when data is identical.
- **Problem**: Template execution allocates bytes.Buffer and walks the template AST every install, even when the output hasn't changed. For repeat installs (update/reinstall flows), this is wasted work.
- **Fix**:
  1. After RC-006 enables lazy construction, cache rendered template output with a content-addressable key (hash of template data)
  2. Invalidate cache when version string or data changes
  3. Use `bytes.Buffer` pooling (`sync.Pool`) for template execution buffers
- **Verification**: Benchmark `Install()` with identical data — second call should be measurably faster. Profile with `-benchmem` to confirm reduced allocations.
- **References**: Go text/template performance, sync.Pool best practices

### P2-005 — Command Files Read from embed.FS Individually
- **Severity**: MEDIUM (recalibrated from LOW)
- **Evidence**: `adapters/common/base_adapter.go:338-346` — loop over `CommandFiles` calls `CommandFS.ReadFile()` for each file individually. Each call creates a new `bytes.Reader` and traverses the embedded filesystem structure.
- **Problem**: embed.FS read performance scales with the number of files when read individually. With 10+ command files, this loop causes repeated filesystem-walk overhead inside the embedded binary.
- **Fix**:
  1. Read all command files into a `map[string][]byte` at adapter construction time (after RC-006 lazy loading)
  2. Use the in-memory map during Install() instead of repeated embed.FS reads
  3. This also eliminates filesystem access during Install, improving reliability
- **Verification**: Profile Install() with `-cpuprofile`. Before fix, expect repeated `embed.(*FS).ReadFile` calls. After fix, expect a single batch read at construction.
- **References**: Go embed.FS performance characteristics, premature optimization vs. hot-path optimization
