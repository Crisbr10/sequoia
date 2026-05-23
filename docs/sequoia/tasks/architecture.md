# Architecture Tasks — sequoia-ai

**Score**: 36/100 (Critical) | **Findings**: 10 (0 CRITICAL, 2 HIGH, 8 MEDIUM) | **Audit ID**: audit-20260521-sequoia-ai

---

## 🔴 HIGH Findings

### P3-001: Remove DefaultRegistry implicit coupling via init()

**Problem**: All 5 adapters call `RegisterIn(adapters.DefaultRegistry)` in `init()`, creating global mutable state at import time. `cmd/sequoia/main.go:43` creates its own registry via `NewRegistry()`, making the init() registration dead data. Tests depend on this global pollution, making them fragile and order-dependent.

**Root Cause**: CORR-002 (Global Mutable State via init() Functions)

**Evidence**:
- `adapters/registry.go` exports `var DefaultRegistry = NewRegistry()`
- Each adapter's `init()` calls `RegisterIn(adapters.DefaultRegistry)`
- `cmd/sequoia/main.go:43` creates `reg := adapters.NewRegistry()` — ignores DefaultRegistry
- Tests in adapter packages import the package and rely on DefaultRegistry being pre-populated

**Fix**: Remove all `init()` functions. Remove `DefaultRegistry`. Make `NewRegistry()` + `RegisterIn()` the only registration path. Update tests to use explicit registry creation.

**Acceptance Criteria**:
- [ ] All `init()` functions removed from 5 adapters + `_template`
- [ ] `DefaultRegistry` variable removed from `adapters/registry.go`
- [ ] Tests updated to create explicit registries: `reg := NewRegistry(); adapter.RegisterIn(reg)`
- [ ] `_template` demonstrates the correct `RegisterFactory()` pattern
- [ ] No test order dependencies remain

**Effort**: medium (4-8h) | **Risk**: medium (tests need updating) | **Blocks**: CORR-002

---

### ✅ P3-002: Refactor pipeline to expose per-phase progress

**Problem**: `internal/pipeline/runner.go` collapses the entire install lifecycle (Prepare → Download → Verify → Stage → Apply) into a single function. Callers can't observe per-phase progress, can't cancel at phase boundaries, and can't recover from partial failures.

**Root Cause**: CORR-004 (Common Package Architectural Bottleneck) — **RESUELTO 2026-05-22**

**Evidence**:
- `pipeline/runner.go` exposes a single `Run(ctx, adapter, tools)` function
- Progress messages (channel-based) only indicate "working" or "done"
- No distinction between "downloading" vs "installing" vs "verifying"
- Error handling can't differentiate between download failure and install failure

**Fix**: Refactor pipeline into exposed phases: `Prepare()`, `Download()`, `Verify()`, `Stage()`, `Apply()`. Each phase returns typed progress. Allow consumers to compose phases or run the full pipeline as a convenience.

**Acceptance Criteria**:
- [x] Pipeline phases exposed as public functions → `Strategy` interface with 6 phased methods + `BaseAdapter` 5-phase decomposition
- [x] Each phase returns typed progress updates via channel → `ProgressMsg.Phase` carries `PhasePrepare`, `PhaseDownload`, etc.
- [x] Phases can be composed or run individually → Pipeline dispatches via switch-based `Strategy` interface
- [x] Backward-compatible `Run()` convenience function preserves existing behavior → `Install()` delegates to phased methods
- [x] Progress channel carries phase-level detail (e.g., `PhaseDownloading`, `PhaseInstalling`) → `ProgressMsg.Phase` + `Step` fields

**Effort**: medium (8-16h) | **Risk**: medium (pipeline consumers need updates) | **Blocks**: CORR-004 | **Completado**: 2026-05-22 (PR 2 de CORR-004)

---

## 🟡 MEDIUM Findings

### ✅ P3-003: Narrow ToolAdapter interface consumption

**Problem**: `ToolAdapter` at `adapters/interface.go:53-59` combines 4 sub-interfaces into an 11-method monolith. Consumers in `cmd/sequoia/main.go` all accept `[]adapters.ToolAdapter` instead of the narrow interfaces they actually need (typically 2-3 methods each).

**Root Cause**: CORR-004 (Common Package Architectural Bottleneck) — **RESUELTO 2026-05-22**

**Fix**: Accept narrow interfaces at consumption points:
- `targetAdapters()` returns `[]adapters.Identifier` (only needs ID/Name)
- `runStatus()` accepts `[]adapters.Detector` + `[]adapters.InstallStatus`
- `runInstall()` accepts `[]adapters.Installer` + `[]adapters.Identifier`
- `runUninstall()` accepts `[]adapters.Uninstaller` + `[]adapters.Identifier`

**Acceptance Criteria**:
- [x] Consumer functions use narrow interfaces instead of `ToolAdapter` → `targetAdapters→[]Identifier`, `runInstall`/`runUninstall`/`runStatus` via role interfaces
- [x] `MockAdapter` split into `MockIdentifier`, `MockDetector`, `MockInstaller` → 5 focused mocks + composite `MockAdapter`
- [x] Tests updated to use focused mocks → 5 compile-time checks + value assertions
- [x] `ToolAdapter` retained as composite for registration convenience → registry uses `ToolAdapter`, consumers use role interfaces

**Effort**: large (12-24h) | **Risk**: high (touches all consumers) | **Blocks**: CORR-004 | **Completado**: 2026-05-22 (PR 3 de CORR-004)

---

### P3-004: Fix double error wrapping in defer functions

**Problem**: Several adapter `Install()` methods wrap errors both in the operation and in the deferred cleanup function. When cleanup fails, the original error is lost, replaced by the cleanup error.

**Evidence**:
- Pattern: `err = fmt.Errorf("install: %w", err)` followed by `defer func() { if err != nil { err = fmt.Errorf("cleanup: %w", err) } }()`
- Original error context is overwritten by defer, making root cause debugging harder
- Go 1.20+ `errors.Join` is available but not used

**Fix**: Use `errors.Join` to combine operation error and cleanup error. Use named return with explicit error collection. Ensure the original error is always preserved.

**Acceptance Criteria**:
- [ ] All deferred error wrapping replaced with `errors.Join`
- [ ] Original error always preserved in combined error
- [ ] Test verifies both operation and cleanup errors are accessible via `errors.Is`/`errors.As`
- [ ] Audit all adapter `Install()` methods for this pattern

**Effort**: small (<2h) | **Risk**: low | **Blocks**: none

---

### P3-005: Split adapters/common god package

**Problem**: `adapters/common` contains 14 files covering 8 unrelated concerns: installer, strategy, path resolution, detection, backup, template rendering, filesystem utilities, and version management. Changes to any file recompile all 5 adapters.

**Evidence**:
- `installer.go` — file copy, staging, atomic writes
- `strategy.go` — installation strategy interface + implementation
- `detector.go` — PATH scanning, file existence checks
- `path_resolver.go` — symlink resolution, home expansion
- `backup.go` — backup path builder
- `templates.go` — template rendering with embedded FS
- `version.go` — version file reading
- `base_adapter.go` — shared adapter struct

**Fix**: Split into focused sub-packages:
- `adapters/install` — Installer, InstallerConfig, file operations
- `adapters/detect` — Detector, PathResolver, exec wrappers
- `adapters/templates` — PromptManager, RenderTemplate, embedded FS
- `adapters/backup` — BackupPathBuilder
- `adapters/fsutil` — ResolveSymlink, IsSymlink, ResolveHome (or `internal/fs`)

**Acceptance Criteria**:
- [ ] Sub-packages extracted with clear boundaries
- [ ] All 5 adapters updated to import from new packages
- [ ] Tests pass for all adapters after migration
- [ ] `adapters/common` becomes thin re-export (deprecated) for backward compatibility
- [ ] No circular dependencies introduced

**Effort**: large (16-24h) | **Risk**: high (touches all adapters) | **Blocks**: none

---

### P3-006: Replace unguarded type assertions in pipeline

**Problem**: `internal/pipeline/runner.go:178-179,185-186` uses unguarded type assertions `a := t.Adapter.(pipelineInstaller)` — will panic if an adapter doesn't implement `Install()`/`Uninstall()`.

**Evidence**:
- Type assertion without `ok` check: `a := t.Adapter.(pipelineInstaller)`
- Pipeline iterates over `[]model.ToolState` which wraps `ToolAdapter`
- If a future adapter omits `Install()`, the pipeline panics at runtime

**Fix**: Use guarded assertion `a, ok := t.Adapter.(pipelineInstaller); if !ok { return error }`. Or define a `PipelineAdapter` interface and have pipeline accept it directly.

**Acceptance Criteria**:
- [ ] Unguarded type assertions replaced with guarded `ok` pattern
- [ ] Graceful error returned instead of panic
- [ ] Test verifies graceful handling of non-installer adapters
- [ ] Consider defining `PipelineAdapter` interface for type safety

**Effort**: small (<2h) | **Risk**: low | **Blocks**: none

---

### P3-007: Align _template adapter with real patterns

**Problem**: `adapters/_template/adapter.go` uses eager `Register()` instead of lazy `RegisterFactory()`, directly implements `ToolAdapter` without embedding `BaseAdapter`, and duplicates the 250-line `Install()` method.

**Fix**: Rewrite `_template/adapter.go` to embed `common.BaseAdapter`, use `RegisterFactory()` with lazy construction, and delegate to `BaseAdapter` for shared behavior.

**Acceptance Criteria**:
- [ ] `_template` embeds `common.BaseAdapter`
- [ ] Uses `RegisterFactory()` pattern
- [ ] Calls `SetPaths/SetDetector/SetPromptManager/SetBackup/SetInstallTemplates` in factory
- [ ] Removes duplicated Install() — delegates to BaseAdapter
- [ ] Adds context cancellation support matching real adapters

**Effort**: small (<2h) | **Risk**: low | **Blocks**: none

---

### P3-008: Reduce global mutable CommandFiles state

**Problem**: `CommandFiles` is an exported mutable slice in `adapters/common`. It's mutated across adapters and lacks synchronization, creating race condition risk in concurrent adapter usage.

**Root Cause**: CORR-002 (Global Mutable State via init() Functions)

**Fix**: Make `CommandFiles` unexported. Provide a getter that returns a copy. Mutations go through controlled functions.

**Acceptance Criteria**:
- [ ] `CommandFiles` renamed to unexported `commandFiles`
- [ ] `GetCommandFiles()` returns a copy of the slice
- [ ] `AddCommandFile()` and `RemoveCommandFile()` for controlled mutations
- [ ] Test verifies getter returns independent copy

**Effort**: small (<2h) | **Risk**: low | **Blocks**: CORR-002

---

### P3-009: Add templateCache eviction and test isolation

**Problem**: `templateCache` is a `sync.Map` at package level in `adapters/common`. It never evicts entries and persists across tests, causing test pollution when tests modify template behavior.

**Root Cause**: CORR-002 (Global Mutable State via init() Functions)

**Fix**: Add `ResetTemplateCache()` for test isolation. Consider whether a process-lifetime cache needs eviction (for a CLI tool, it doesn't — but the test isolation gap is real). Call `ResetTemplateCache()` in `TestMain` or individual tests that modify templates.

**Acceptance Criteria**:
- [ ] `ResetTemplateCache()` function added
- [ ] Called in `TestMain` of adapter test packages
- [ ] Tests no longer depend on template cache state from other tests
- [ ] Document that cache is process-lifetime (no eviction needed for CLI use case)

**Effort**: small (<1h) | **Risk**: low | **Blocks**: CORR-002

---

### ✅ P3-010: Activate Strategy pattern in pipeline

**Problem**: `adapters/common/strategy.go` defines a `Strategy` interface with `Install()` and `Uninstall()` methods, but pipeline code paths bypass it. The strategy pattern exists but is never consumed.

**Root Cause**: CORR-004 (Common Package Architectural Bottleneck) — **RESUELTO 2026-05-22**

**Fix**: Pipeline should consume `Strategy` interface instead of type-asserting to `pipelineInstaller`. Each adapter registers its strategy, and the pipeline calls through the interface. This eliminates the need for type assertions and makes the strategy pattern functional.

**Acceptance Criteria**:
- [x] Pipeline accepts `Strategy` interface instead of type assertions → Pipeline dispatch via guarded `s, ok := adapter.(common.Strategy)`
- [x] Each adapter provides a `Strategy()` method returning its strategy → `BaseAdapter.Strategy()` accessor returns `common.Strategy`
- [x] Type assertions in pipeline replaced with interface method calls → Zero unguarded type assertions in runner.go
- [x] Strategy interface consumed as designed → 5 adapters + BaseAdapter satisfy `common.Strategy` via compile-time checks

**Effort**: medium (4-8h) | **Risk**: medium | **Blocks**: CORR-004 | **Completado**: 2026-05-22 (PR 2 de CORR-004)

---

## Summary

| Priority | Finding | Title | Effort | Blocks |
|----------|---------|-------|--------|--------|
| 🔴 HIGH | P3-001 | Remove DefaultRegistry + init() | medium | CORR-002 |
| 🔴 HIGH | P3-002 | ✅ Refactor pipeline phases | medium | CORR-004 |
| 🟡 MED | P3-003 | ✅ Narrow ToolAdapter interface | large | CORR-004 |
| 🟡 MED | P3-004 | Fix double error wrapping | small | — |
| 🟡 MED | P3-005 | Split common god package | large | — |
| 🟡 MED | P3-006 | Guard type assertions | small | — |
| 🟡 MED | P3-007 | Align _template adapter | small | — |
| 🟡 MED | P3-008 | Reduce CommandFiles mutability | small | CORR-002 |
| 🟡 MED | P3-009 | Fix templateCache test isolation | small | CORR-002 |
| 🟡 MED | P3-010 | ✅ Activate Strategy pattern | medium | CORR-004 |

**Priority Order**: P3-006 (safety) → P3-004 (correctness) → P3-001 + P3-008 + P3-009 (CORR-002 batch) → P3-007 (template) → ~~P3-002 + P3-003 + P3-010 (CORR-004 batch)~~ ✅ → P3-005 (split, backlog)

*Generated by Sequoia M2 Reporter — audit-20260521-sequoia-ai | Schema v1.0*
