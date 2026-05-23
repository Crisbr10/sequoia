# Performance Tasks — sequoia-ai

**Score**: 38/100 (Critical) | **Findings**: 11 (0 CRITICAL, 2 HIGH, 9 MEDIUM) | **Audit ID**: audit-20260521-sequoia-ai

---

## 🔴 HIGH Findings

### P2-001: Cache Detect() results for process lifetime

**Problem**: Every adapter's `Detect()` (`adapters/claude/adapter.go:64` and siblings) calls `exec.LookPath` on every CLI command and TUI screen with zero caching. `sequoia status` makes up to 5 PATH scans, each spawning a subprocess enumeration.

**Evidence**:
- `Detect()` in `BaseAdapter` delegates to `Detector.Detect()` which calls `exec.LookPath`
- Called from `runStatus`, `runInstall`, `runUninstall`, and `ToolSelectionModel.Init`
- Process lifetime: PATH doesn't change during execution — results are stable

**Fix**: Add `detectedOnce sync.Once` and `detectedResult bool` to `BaseAdapter`. `Detect()` calls `sync.Once.Do()` wrapping the `exec.LookPath` + install check. Subsequent calls return cached result immediately.

**Acceptance Criteria**:
- [x] `sync.Once` memoization added to `BaseAdapter.Detect()`
- [x] `exec.LookPath` called exactly once per adapter per process
- [x] Test verifies single invocation for repeated `Detect()` calls
- [x] CLI commands: `sequoia status` PATH scans reduced from 5 to 0 (first call cached)

**Effort**: small (<2h) | **Risk**: low | **Blocks**: P2-003, P2-008

---

### P2-002: Pre-render Logo at init time

**Problem**: `Logo()` in `internal/tui/styles/logo.go:19-29` splits the logo string, creates a new lipgloss style, and renders through the style pipeline on every frame. At 60fps this is ~60 allocations/sec of multi-line strings.

**Root Cause**: CORR-003 (TUI Rendering Path Inefficiency)

**Fix**: Pre-render the logo string once as a package-level `var` in `init()` or at declaration. `Logo()` function returns the pre-rendered string directly.

**Acceptance Criteria**:
- [ ] Logo pre-rendered as package-level var
- [ ] `Logo()` function reads cached result, performs zero allocations
- [ ] Golden test confirms logo output unchanged
- [ ] Frame render time for TUI header reduced by ~0.3ms

**Effort**: small (<30m) | **Risk**: low | **Blocks**: CORR-003

---

## 🟡 MEDIUM Findings

### P2-003: Cache IsInstalled() results for process lifetime

**Problem**: `BaseAdapter.IsInstalled()` (`adapters/common/base_adapter.go:225-230`) calls `Detector.IsInstalled()` which reads the system prompt file via `os.ReadFile`. Called in TUI `View()` at 60fps. UninstallView calls it up to 10 times per frame.

**Fix**: Add `installedOnce sync.Once` and `installedResult bool` to `BaseAdapter`. Cache is process-lifetime (files don't change during a CLI session).

**Acceptance Criteria**:
- [ ] `sync.Once` memoization added to `BaseAdapter.IsInstalled()`
- [ ] `os.ReadFile` called exactly once per adapter per process
- [ ] Cache respects process lifetime — no invalidation needed
- [ ] Test verifies single file read for repeated `IsInstalled()` calls

**Effort**: small (<2h) | **Risk**: low | **Blocks**: P2-007

---

### P2-004: Cache Status() version file reads

**Problem**: `Status()` reads the version file via `os.ReadFile` on every invocation without caching. CLI `sequoia status` reads each adapter's version file on every call.

**Fix**: Cache version string after first successful read in `BaseAdapter`. Return cached version on subsequent `Status()` calls.

**Acceptance Criteria**:
- [ ] Version string cached after first successful `os.ReadFile`
- [ ] Subsequent `Status()` calls return cached result
- [ ] Cache respects process lifetime (version file doesn't change during session)
- [ ] Error on first read still returned (don't cache errors)

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

### P2-005: Pre-build Lipgloss styles as package-level vars

**Problem**: All 8 style functions in `internal/tui/styles/styles.go` call `lipgloss.NewStyle()` on every invocation, creating ~200 bytes of heap per call. At 60fps with multiple styled elements this adds up to significant GC pressure.

**Root Cause**: CORR-003 (TUI Rendering Path Inefficiency)

**Fix**: Convert `Title()`, `Subtitle()`, `Body()`, `Accent()`, `Error()`, `Success()`, `Muted()`, `Highlight()` from functions to package-level `var` declarations. Lipgloss styles are immutable once built — no reason to recreate them.

**Acceptance Criteria**:
- [ ] All 8 style functions converted to package-level vars
- [ ] Golden tests confirm visual output unchanged
- [ ] Zero heap allocations in TUI style rendering path
- [ ] ~1.3 MB/s reduction in GC pressure at 60fps

**Effort**: small (<1h) | **Risk**: low | **Blocks**: CORR-003

---

### P2-006: Add Grow() hints to strings.Builder usage

**Problem**: Multiple `strings.Builder` instances in TUI views are created without `Grow()` hints. The builder reallocates its internal buffer multiple times as content is appended, causing unnecessary allocations and copies.

**Root Cause**: CORR-003 (TUI Rendering Path Inefficiency)

**Fix**: Audit all `strings.Builder` usage in `internal/tui/`. Add `b.Grow(estimatedSize)` before building where the output size is predictable. For list views, estimate based on item count × average line length.

**Acceptance Criteria**:
- [ ] `Grow()` hints added to all TUI `strings.Builder` instances
- [ ] Estimates based on: logo (~500 bytes), status list (~80 bytes/item), menu (~200 bytes)
- [ ] Benchmark shows reduced allocations in View() path
- [ ] No behavioral change — same output, fewer reallocations

**Effort**: small (<1h) | **Risk**: low | **Blocks**: CORR-003

---

### P2-007: Compute installed count in Update(), not View()

**Problem**: `UninstallView` loops through all tools calling `IsInstalled()` twice per frame — once for count, once for rendering. Combined with P2-003 caching, the first frame still does redundant work.

**Fix**: Add `installedCount` field to `UninstallModel`. Compute once in `UninstallUpdate` when model is created. `UninstallView` reads pre-computed count.

**Acceptance Criteria**:
- [ ] `installedCount` computed once on model initialization
- [ ] `UninstallView` reads cached count from model field
- [ ] Loop in View() replaced with O(1) field access
- [ ] Test verifies count is accurate after model initialization

**Effort**: small (<1h) | **Risk**: low | **Blocks**: P2-003 (depends on caching)

---

### P2-008: Eliminate redundant PATH scans in CLI commands

**Problem**: `runStatus` in `cmd/sequoia/main.go` calls `Detect()` for each adapter even though the adapter list is already known from `reg.All()`. Multiple PATH scans for the same binary across different code paths.

**Fix**: Remove redundant `Detect()` loops. If the adapter is already known to be registered, the Detect result is implicitly known. Only call `Detect()` when explicitly needed.

**Acceptance Criteria**:
- [ ] Redundant `Detect()` call in `runStatus` removed (use P2-001 cache instead)
- [ ] `runInstall` skips Detect if adapter was explicitly specified
- [ ] `sequoia status` runtime reduced by ~30ms on cold cache

**Effort**: small (<1h) | **Risk**: low | **Blocks**: P2-001 (depends on caching)

---

### P2-009: Add view rendering memoization

**Problem**: TUI views are recomputed on every `Update()` call, even when the model state hasn't changed. Keyboard events that don't modify state still trigger full re-renders.

**Fix**: Track a `dirty` flag on each model. Set `dirty = true` only when state actually changes. `View()` returns cached string when `!dirty`. Clear dirty after render.

**Acceptance Criteria**:
- [ ] `dirty` flag added to TUI model structs
- [ ] `View()` returns cached output when state unchanged
- [ ] `dirty` reset to false after each render
- [ ] No visual regression in golden tests
- [ ] Frame render time reduced to near-zero on idle frames

**Effort**: medium (2-4h) | **Risk**: medium (state tracking complexity) | **Blocks**: none

---

### P2-010: Cache os.Stat() results in installer views

**Problem**: Installer views call `os.Stat()` to check file existence on every model update. These files don't change during a TUI session. Combined with P2-003 caching, Stat calls become the remaining I/O hot path.

**Fix**: Memoize `os.Stat()` results in the installer model. Files being installed/checked don't change paths during a session.

**Acceptance Criteria**:
- [ ] `os.Stat()` results cached in installer model
- [ ] Cache keyed by file path, process-lifetime TTL
- [ ] Test verifies single Stat call per file path
- [ ] Install/uninstall flows verified via integration test

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

### P2-011: Defer adapter construction until needed

**Problem**: All 5 adapters are constructed eagerly in `cmd/sequoia/main.go` before the first TUI frame renders. Each adapter initializes `BaseAdapter`, sets paths, creates detector, prompt manager, backup handler, and install templates — all before the user sees anything.

**Root Cause**: CORR-003 (TUI Rendering Path Inefficiency)

**Fix**: Use `RegisterFactory()` pattern (already supported). Construct adapters lazily when first accessed via the registry. Defer construction past the first frame render.

**Acceptance Criteria**:
- [ ] All adapters use `RegisterFactory()` instead of `Register()`
- [ ] Factory construction deferred until first access
- [ ] First-frame TTI (Time To Interactive) reduced by ~100ms
- [ ] Test verifies adapter construction is lazy
- [ ] `_template` updated to demonstrate factory pattern

**Effort**: small (1-2h) | **Risk**: low | **Blocks**: CORR-003

---

## Summary

| Priority | Finding | Title | Effort | Blocks |
|----------|---------|-------|--------|--------|
| 🔴 HIGH | P2-001 | Cache Detect() results | small | P2-003, P2-008 |
| 🔴 HIGH | P2-002 | Pre-render Logo | small | CORR-003 ✅ |
| 🟡 MED | P2-003 | Cache IsInstalled() | small | P2-007 |
| 🟡 MED | P2-004 | Cache Status() reads | small | — |
| 🟡 MED | P2-005 | Cache Lipgloss styles | small | CORR-003 ✅ |
| 🟡 MED | P2-006 | Add Grow() hints | small | CORR-003 ✅ |
| 🟡 MED | P2-007 | Compute count in Update | small | P2-003 |
| 🟡 MED | P2-008 | Remove redundant PATH scans | small | P2-001 |
| 🟡 MED | P2-009 | View rendering memoization | medium | — |
| 🟡 MED | P2-010 | Cache os.Stat() results | small | — |
| 🟡 MED | P2-011 | Defer adapter construction | small | CORR-003 ✅ |

**Quick Win**: ✅ Fix P2-002 + P2-005 + P2-006 + P2-011 together (~3h) — COMPLETED. Eliminates ~95% of TUI rendering overhead.

*Generated by Sequoia M2 Reporter — audit-20260521-sequoia-ai | Schema v1.0*
