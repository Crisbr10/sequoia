# Design: RC-001 — Decompose BaseAdapter

## Technical Approach

Extract 4 cohesive structs from the 482-line god-object `BaseAdapter`, segregate the 12-method `ToolAdapter` interface into 4 role interfaces, replace the `DefaultRegistry` global with constructor DI, and remove `factory.go`. Behavior-preserving refactor; no user-visible changes. Each phase independently committable and revertable.

## Architecture Decisions

### Decision: Composition over embedding for extracted structs

**Choice**: BaseAdapter holds named pointer fields (`a.paths *PathResolver`, `a.detector *Detector`, `a.prompt *PromptManager`, `a.backup *BackupPathBuilder`), NOT embedded structs.

**Alternatives considered**: Embedded structs (promoted methods). Rejected: embedding would expose internal methods on BaseAdapter, making it an even larger interface. Named fields provide clear boundary — BaseAdapter orchestrates, delegates to sub-structs.

**Rationale**: Named fields make the dependency graph explicit: `Install()` calls `a.paths.Base()`, `a.detector.IsInstalled()`, `a.prompt.Write()`, `a.backup.Generate()`. A reader can trace every call without wondering which embedded struct provides it.

### Decision: Detector depends on a `baseFn` function, not `*PathResolver` directly

**Choice**: `Detector` stores `isInstalledFn func(base string) bool`, `detectFn func() bool`, and `baseFn func() (string, error)`. Constructor: `NewDetector(baseFn, isInstalledFn, detectFn)`.

**Rationale**: `IsInstalled()` requires the base directory (current code calls `a.Base()`). Giving Detector a `*PathResolver` would couple detection to path resolution. A function field is more testable and respects the spec's "Detector has zero install dependencies" requirement.

### Decision: PathResolver owns shared warnings via callback

**Choice**: PathResolver uses `warnFn func(string)` callback, BaseAdapter passes `a.AddWarning` during construction.

**Rationale**: The current `Base()` method calls `a.AddWarning(msg)` when symlink resolution emits a warning. A callback pattern avoids shared mutable state while keeping warnings flowing to BaseAdapter.

### Decision: Keep `DefaultRegistry` as deprecated compat shim, add `NewRegistry()`

**Choice**: `var DefaultRegistry = NewRegistry()` remains as a top-level variable. `NewRegistry() *Registry` is the constructor. Adapter `init()` functions continue to use `DefaultRegistry.Register()`. `NewModel()`, `runStatus()`, `ScanTools()`, `runUninstall()` accept `*Registry` via parameter.

**Rationale**: Removing `DefaultRegistry` instantly breaks 5 adapter `init()` functions + 20+ tests. Backward-compat shim lets us migrate consumers incrementally. Tests that used global pointer swap now use constructor DI, enabling `t.Parallel()` without mutex.

## Data Flow

### Install() flow after decomposition

```
Install(opts)
  ├─ checkContext(ctx)                  [BaseAdapter]
  ├─ a.paths.Base()                     [PathResolver] → (home, err)
  ├─ a.makeTemplateData()               [BaseAdapter, user-provided]
  ├─ MkdirTemp + RenderTemplate + Stage [BaseAdapter, uses a.templateFS]
  ├─ checkContext(ctx)                  [BaseAdapter]
  ├─ a.paths.SkillsPath(base)           [PathResolver]
  ├─ a.paths.CommandsPath(base)         [PathResolver]
  ├─ a.backup.Generate(base, a.id)      [BackupPathBuilder] → backupDir
  ├─ a.setLastBackupDir(backupDir)      [BaseAdapter, mu-protected]
  ├─ MkdirAll + skillInstaller.Run()    [BaseAdapter]
  ├─ checkContext → skill rollback      [BaseAdapter]
  ├─ cmdInstaller.Run()                 [BaseAdapter]
  ├─ checkContext → both rollback       [BaseAdapter]
  ├─ RenderTemplate(systemPrompt)       [BaseAdapter]
  ├─ a.prompt.Write(base, content)      [PromptManager]
  ├─ checkContext → both rollback       [BaseAdapter]
  └─ AtomicWriteFile(version)           [BaseAdapter]
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `adapters/common/path_resolver.go` | Create | PathResolver struct (152 LOC): Base(), SkillsPath(), CommandsPath(), SystemPromptPath(), HomeDir() |
| `adapters/common/detector.go` | Create | Detector struct (59 LOC): Detect(), IsInstalled(), baseFn injection |
| `adapters/common/prompt_manager.go` | Create | PromptManager struct (70 LOC): PromptStrategy(), Write(), Remove(), ShouldRollback() |
| `adapters/common/backup_path_builder.go` | Create | BackupPathBuilder struct (37 LOC): Generate(base, adapterID) string |
| `adapters/common/base_adapter.go` | Modify | Removed extracted fields. Added 4 named struct pointers. 398 LOC (down from 482) |
| `adapters/interface.go` | Modify | Split into Identifier, Detector, Installer, AdapterPaths. Composite ToolAdapter. |
| `adapters/registry.go` | Modify | Added `NewRegistry() *Registry`. Deprecated DefaultRegistry. |
| `adapters/factory.go` | Delete | Removed `NewAdapter()` — consumers use `registry.Get(id)` directly |
| `adapters/claude/adapter.go` | Modify | Wire PathResolver, Detector, PromptManager, BackupPathBuilder + RegisterIn() |
| `adapters/codex/adapter.go` | Modify | Wire extracted structs + RegisterIn() |
| `adapters/cursor/adapter.go` | Modify | Wire extracted structs + RegisterIn() |
| `adapters/gemini/adapter.go` | Modify | Wire extracted structs + RegisterIn() |
| `adapters/opencode/adapter.go` | Modify | Wire extracted structs + RegisterIn() |
| `adapters/common/base_adapter_error_test.go` | Create | 20 error-path tests |
| `adapters/common/base_adapter_mockfs_test.go` | Create | 6 mock FS tests |
| `adapters/claude/template_test.go` | Create | 4 Claude template rendering tests |
| `adapters/codex/template_test.go` | Create | 5 Codex template rendering tests |
| `internal/app/model.go` | Modify | `NewModel(toolID, version string, reg *Registry)` |
| `cmd/sequoia/main.go` | Modify | Thread registry through commands |

## Interfaces / Contracts

### Segregated role interfaces

```go
type Identifier interface {
    ID() string
    Name() string
}

type Detector interface {
    Detect() bool
    IsInstalled() bool
}

type Installer interface {
    Install(InstallOpts) error
    Uninstall(InstallOpts) error
}

type AdapterPaths interface {
    SkillsPath() string
    CommandsPath() string
    SystemPromptPath() string
    PromptStrategy() PromptStrategy
}

type ToolAdapter interface {
    Identifier
    Detector
    Installer
    AdapterPaths
    Status() AdapterStatus
}
```
