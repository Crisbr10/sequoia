# Architecture Tasks — Sequoia Audit

## Context
Sequoia's architecture centers on the `ToolAdapter` interface with a plugin-registry pattern (`database/sql`-style self-registration). The codebase has 5 adapter implementations (claude-code, opencode, cursor, gemini-cli, codex) plus a `_template` scaffold, a shared `common/` framework, and a TUI pipeline. The dominant architectural issue is **BaseAdapter as god-object** (RC-001) which drives 8 child findings. The audit identified 12 architecture findings: 2 HIGH, 6 MEDIUM, 4 LOW.

**Key files**: `adapters/common/base_adapter.go` (482 lines), `adapters/interface.go`, `adapters/registry.go`, `adapters/_template/adapter.go`, `internal/pipeline/runner.go`

## Priority Tiers

### Tier 1 — Immediate (HIGH)

| ID | Task | Effort | Blocks |
|----|------|--------|--------|
| RC-004 | Namespace backup directories by adapter ID | small | P3-001 |
| RC-001 | Decompose BaseAdapter into interface-segregated components | large | P3-006, P3-007, P4-007 |
| RC-006 | Implement lazy adapter loading, remove init() coupling | medium | P3-005, P2-004 |

### Tier 2 — Short Term (MEDIUM)

| ID | Task | Effort |
|----|------|--------|
| P3-004 | Integrate or delete dead plugin system code | medium |
| RC-005 | Dependency-inject registry, validate nil adapters | small |
| RC-007 | Update _template to embed BaseAdapter, use shared templates | small |

### Tier 3 — Long Term (LOW)

| ID | Task | Effort |
|----|------|--------|
| P3-008 | Replace pipeline type assertions with interface methods | small |
| P3-009 | Rename Go files from hyphens to underscores | small |

---

## Detailed Tasks

### RC-004 — Namespace backup directories by adapter ID
- **Severity**: HIGH
- **Evidence**: `adapters/common/base_adapter.go:358` — `backupDir := a.backupPathFn(base) + "-" + sessionSuffix`
- **Problem**: Both skillInstaller and cmdInstaller receive the same `backupDir`. When cmdInstaller.Run() fails and rolls back (removing backupDir via os.RemoveAll), the subsequent skillInstaller.Rollback() finds the directory gone — silently losing user's original skill file.
- **Fix**:
  1. In `base_adapter.go:358`, change to: `backupDir := filepath.Join(a.backupPathFn(base) + "-" + a.ID() + "-" + sessionSuffix)`
  2. Verify each adapter's `backupPathFn` produces adapter-specific paths
  3. Add test: simulate cmdInstaller failure, verify skillInstaller can still rollback independently
- **Verification**: `go test ./adapters/common/ -run TestBackupIsolation -v -count=1`
- **References**: Go: os.RemoveAll is destructive; shared mutable state between lifecycle phases

### RC-001 — Decompose BaseAdapter into interface-segregated components
- **Severity**: HIGH (affects 8 child findings)
- **Evidence**: `adapters/common/base_adapter.go:24-91` — 22 exported methods, 19 fields, Install() 131 lines
- **Problem**: BaseAdapter mixes three concerns: (a) configuration via 9 setter methods, (b) runtime state (warnings, lastBackupDir), and (c) complex lifecycle methods (Install/Uninstall orchestrating staging, rendering, backup, rollback). Codex duplicates Install() because extracting the shared flow from the monolith is too risky. Common/ coverage is 64.8% because the class is too large to test effectively.
- **Fix**:
  1. Extract `AdapterConfig` struct to hold immutable configuration (ID, Name, paths, strategy, templates)
  2. Create `LifecycleManager` that accepts `AdapterConfig` + a `PromptWriter` interface
  3. Split `PromptWriter` into strategy-specific implementations (MarkdownInjector, FileReplacer, TOMLMerger)
  4. Move `TOMLMerge` from `adapters/codex/` to `adapters/common/toml_strategy.go`
  5. Refactor `Install()`: extract staging, backup, copy, and rollback into a `Pipeline` type
  6. Update all 5 adapters to compose `LifecycleManager` with their specific `PromptWriter`
- **Verification**: All existing adapter tests must pass. Common/ coverage should reach >80%. New adapter authoring guide: <50 lines per new adapter.
- **References**: Go proverb "Accept interfaces, return structs"; Builder pattern; Template Method pattern

### RC-006 — Implement lazy adapter loading
- **Severity**: HIGH (affects 5 child findings)
- **Evidence**: `cmd/sequoia/main.go:22-28` — blank imports for all 5 adapters; each `init()` calls `newAdapter("")` + `DefaultRegistry.Register()`
- **Problem**: Every command including `sequoia version` pays the full init cost of all adapters. internal/app and internal/pipeline must import adapters/ because init() auto-registers everything. Headless install can't parallelize because adapters are loaded sequentially at init.
- **Fix**:
  1. Change adapter `init()` to only register metadata (ID, Name, DetectFn factory), not construct full BaseAdapter
  2. Create `LazyAdapter` type that constructs `BaseAdapter` on first method call
  3. Separate `ToolAdapter` interface into `ToolDetector` (Detect, ID, Name) and `ToolInstaller` (Install, Uninstall, Status)
  4. Remove blank imports from main.go; use `adapters.All()` which returns pre-registered metadata
  5. In headless install loop, construct only needed adapters via `NewAdapter(id)`, then launch goroutines
- **Verification**: `sequoia version` time should drop by ~5ms. `go test -race ./...` must pass. Headless install of 5 tools should be parallel.
- **References**: Lazy initialization pattern; database/sql driver registration (metadata-only at init)

### P3-004 — Integrate or delete dead plugin system code
- **Severity**: MEDIUM
- **Evidence**: `plugin/interface.go`, `plugin/loader.go` — no production code imports this package
- **Problem**: The plugin subsystem (Plugin interface, YAML manifest loader, example plugin) has zero callers outside its own package. It increases binary size and creates a public API surface that may need breaking changes when actual integration begins.
- **Fix**:
  1. Option A (integrate): Add plugin discovery to main.go startup. On `sequoia install`, scan plugin dir, load manifests, register as additional adapters.
  2. Option B (delete): Remove `plugin/` directory entirely. Re-add when v0.2 audit command needs plugin extensibility.
  3. Decision: Option B recommended for v1.0.7 (YAGNI). Plugin system has no tests and no integration path.
- **Verification**: Binary size should decrease. `go build ./...` must succeed. No import of plugin/ from any package.
- **References**: YAGNI principle; Go packages with no importers are dead code

### RC-005 — Dependency-inject registry, validate nil adapters
- **Severity**: MEDIUM
- **Evidence**: `adapters/registry.go:25-40` — Register() accepts nil with no validation, panics at a.ID()
- **Problem**: Global `DefaultRegistry` is shared mutable state. Tests manually mutex-lock around it. A nil registration in any adapter's init() crashes the entire binary at import time — all commands become unreachable.
- **Fix**:
  1. Add nil check in Register(): `if a == nil { panic("adapters: Register(nil)") }` or return error
  2. Replace `var DefaultRegistry = &Registry{}` with `func DefaultRegistry() *Registry` that returns a singleton via sync.Once
  3. Add `NewRegistry() *Registry` constructor for tests to use their own registry
  4. Update all adapter init() and test files
- **Verification**: Tests no longer need `registryMu.Lock()`. Test with nil adapter must not crash.
- **References**: Defensive programming; Go nil interface values are not nil

### RC-007 — Update _template to embed BaseAdapter
- **Severity**: MEDIUM
- **Evidence**: `adapters/_template/adapter.go` — 267 lines, no BaseAdapter embedding; all 5 real adapters embed BaseAdapter and average ~60 lines
- **Problem**: New contributors who copy the template produce 200+ line adapters that duplicate install/uninstall orchestration, diverging from the established pattern. The template has 17 TODO markers and no test file.
- **Fix**:
  1. Rewrite `_template/adapter.go` to embed `common.BaseAdapter` (like real adapters)
  2. Remove duplicated staging, template rendering, installer lifecycle code
  3. Add `//go:build ignore` directive to prevent compilation
  4. Remove 17 TODO markers; replace with concise doc comments
  5. Add `_template/adapter_test.go` showing how to test a new adapter
  6. Move shared `skill.md.tmpl` from claude+gemini to `adapters/common/templates/`
- **Verification**: Template adapter compiles under build tag. New adapter author can copy template and have <70 lines of code.
- **References**: DRY principle; Template Method pattern

### P3-008 — Replace pipeline type assertions with interface methods
- **Severity**: LOW
- **Evidence**: `internal/pipeline/runner.go:162,178` — `a := t.Adapter.(pipelineInstaller)` (unchecked, would panic)
- **Problem**: Pipeline uses 3 type assertions on `model.ToolInfo` to recover the full `ToolAdapter`. The `pipelineInstaller` assertion is unchecked — a wrapper change causes runtime panic. `BackupDirGetter` and `WarningEmitter` assertions silently degrade if wrapping changes.
- **Fix**:
  1. Add `Install(InstallOpts) error` and `Uninstall(UninstallOpts) error` to `model.ToolInfo` interface
  2. Or: create `PipelineAdapter` interface with exactly the methods pipeline needs
  3. Remove all type assertions from runner.go
- **Verification**: Compile-time safety. Wrapping changes caught at build time, not runtime.
- **References**: Interface segregation; Go type assertions are runtime-only checks

### P3-009 — Rename Go files from hyphens to underscores
- **Severity**: LOW
- **Evidence**: `internal/tui/screens/tool-selection.go`, `internal/tui/screens/install-progress.go`
- **Problem**: Go convention uses lowercase without separators or underscores only in `_test.go`. Hyphens in source filenames are not idiomatic. Test files use underscores correctly (`tool_selection_test.go`), creating inconsistency.
- **Fix**:
  1. Rename: `tool-selection.go` → `tool_selection.go`
  2. Rename: `install-progress.go` → `install_progress.go`
  3. Update any internal references (should be none — Go doesn't reference files by name)
- **Verification**: `go build ./...` succeeds. `go test ./...` succeeds. Golden file tests pass.
- **References**: Go source file naming conventions (golang.org/doc)
