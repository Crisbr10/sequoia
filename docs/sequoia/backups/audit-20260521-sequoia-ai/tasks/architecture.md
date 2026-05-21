# Architecture Tasks — sequoia-ai

**Score**: 0/100 (Critical) | **Findings**: 10

---

## 🔴 P3-002 (high): Replace unguarded type assertions in pipeline

**Problem**: `internal/pipeline/runner.go:178-179,185-186` uses unguarded type assertions `a := t.Adapter.(pipelineInstaller)` — will panic if adapter doesn't implement install/uninstall.

**Acceptance Criteria**:
- [ ] Define a `PipelineAdapter` interface combining `ToolAdapter` + `Install()` + `Uninstall()` methods
- [ ] Pipeline functions accept `PipelineAdapter` directly instead of `model.ToolState` + type assertion
- [ ] Or: use guarded assertion `a, ok := t.Adapter.(pipelineInstaller); if !ok { return error }` 
- [ ] Add test that verifies graceful error when adapter lacks Install/Uninstall
- [ ] Consider removing `ToolInfo` wrapper in `internal/model` and passing `ToolAdapter` directly

**Effort**: medium (4-8h) | **Risk**: medium | **Blocks**: CORR-004

---

## P3-003 (medium): Remove unused DefaultRegistry init() pattern

**Problem**: All 5 adapters call `RegisterIn(adapters.DefaultRegistry)` in `init()`, populating a deprecated registry that's never consumed. `cmd/sequoia/main.go:43` creates its own registry.

**Acceptance Criteria**:
- [ ] Remove `DefaultRegistry` variable from `adapters/registry.go`
- [ ] Remove `init()` functions from all 5 adapter packages + _template
- [ ] Verify tests still pass (some tests may depend on DefaultRegistry being populated)
- [ ] Update _template to not include the `RegisterIn(adapters.DefaultRegistry)` line

**Effort**: small (<2h) | **Risk**: low (tests may need updating) | **Blocks**: none

---

## P3-004 (medium): Align _template adapter with real patterns

**Problem**: `adapters/_template/adapter.go` uses eager `Register()` vs lazy `RegisterFactory()`, directly implements ToolAdapter without BaseAdapter, duplicates 250-line Install().

**Acceptance Criteria**:
- [ ] Rewrite `_template/adapter.go` to embed `common.BaseAdapter`
- [ ] Use `RegisterFactory()` with lazy construction
- [ ] Call `SetPaths/SetDetector/SetPromptManager/SetBackup/SetInstallTemplates` in factory
- [ ] Remove duplicated Install() — delegate to BaseAdapter
- [ ] Add context cancellation support

**Effort**: small (<2h) | **Risk**: low | **Blocks**: none

---

## P3-005 (medium): Reduce global mutable state in adapters/common

**Problem**: `CommandFiles` (exported mutable slice), `templateCache` (sync.Map without eviction), `DefaultRegistry` (mutated by init functions).

**Acceptance Criteria**:
- [ ] Make `CommandFiles` unexported with a getter function that returns a copy
- [ ] Add `ResetTemplateCache()` for test isolation (called in TestMain or individual tests)
- [ ] Remove `DefaultRegistry` entirely (see P3-003)

**Effort**: medium (4-8h) | **Risk**: medium | **Blocks**: none

---

## P3-006 (medium): Narrow ToolAdapter interface consumption

**Problem**: `adapters/interface.go:53-59` defines ToolAdapter as a composite of 4 sub-interfaces + Status (11+ methods total). Consumers in `cmd/sequoia/main.go` all accept `[]adapters.ToolAdapter` instead of the narrow interfaces they actually need.

**Acceptance Criteria**:
- [ ] `targetAdapters()` returns `[]adapters.Identifier` instead of `[]adapters.ToolAdapter` (only needs ID/Name)
- [ ] `runStatus()` accepts `[]adapters.Detector` + `[]adapters.InstallStatus` (only needs Detect/IsInstalled/Status)
- [ ] `runInstall()` accepts `[]adapters.Installer` + `[]adapters.Identifier`
- [ ] `runUninstall()` accepts `[]adapters.Installer` + `[]adapters.Identifier`
- [ ] `MockAdapter` split into focused mocks: `MockIdentifier`, `MockDetector`, `MockInstaller`

**Effort**: large (12-24h) | **Risk**: high (touches all consumers) | **Blocks**: CORR-004

---

## P3-007~P3-010 (low/info): Structural cleanup

- **P3-007**: Split `cmd/sequoia/main.go` into `cmd/sequoia/commands/install.go`, `commands/status.go`, etc. (390-line file → ~80 lines each)
- **P3-008**: Define `RegistryInterface` with `Get/All/Register/RegisterFactory` methods; have consumers accept the interface
- **P3-009**: Update `internal/model/types_test.go` to test against local `ToolInfo` interface, not `adapters.ToolAdapter`
- **P3-010**: Document the `adapters/common → adapters` dependency direction; consider extracting types to a shared package

**Effort**: medium (8-12h total) | **Risk**: low | **Blocks**: none

---

## CORR-001 (large): Split adapters/common god package

**Problem**: 7 types, 16+ functions, 6 unrelated concerns in one package. Changes to any area recompile all 5 adapters.

**Acceptance Criteria**:
- [ ] Extract `adapters/install` (Installer, InstallerConfig, copyFile, StageFile, AtomicWriteFile)
- [ ] Extract `adapters/templates` (PromptManager, RenderTemplate, templateCache, embedded FS)
- [ ] Extract `adapters/detect` (Detector, PathResolver, exec.LookPath wrappers)
- [ ] Extract `adapters/backup` (BackupPathBuilder)
- [ ] Extract `adapters/fsutil` (ResolveSymlink, IsSymlink, ResolveHome) — or move to `internal/`
- [ ] `adapters/common` becomes a thin re-export package for backward compatibility (deprecated)
- [ ] All 5 adapters updated to import from new focused packages
- [ ] Tests pass for all adapters after migration

**Effort**: large (16-24h) | **Risk**: high | **Priority**: backlog

---

## Priority Order

1. **P3-002** (type assertions) — safety fix, medium effort
2. **P3-003 + P3-004 + P3-005** (cleanup) — small effort, can batch
3. **P3-007** (split main.go) — medium effort, improves maintainability
4. **P3-006** (interface narrowing) — large effort, high payoff
5. **CORR-001** (split common) — largest effort, backlog
