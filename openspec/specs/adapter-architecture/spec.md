# Delta Specs for RC-001 — Decompose BaseAdapter

## Added Requirements

### REQ-COMPOSITION: BaseAdapter SHALL use named composition over embedding

BaseAdapter SHALL hold extracted sub-structs as named pointer fields, NOT embedded structs. This makes the dependency graph explicit and prevents method promotion that would inflate the adapter interface.

| Extracted Struct | Field Name | Responsibility |
|-----------------|------------|----------------|
| `PathResolver` | `paths` | Path resolution and home directory caching |
| `Detector` | `detector` | Tool presence detection with injected baseFn |
| `PromptManager` | `prompt` | System prompt write/remove lifecycle |
| `BackupPathBuilder` | `backup` | Unique backup directory path generation |

#### Scenario: Install orchestrates through named delegation
- GIVEN a BaseAdapter with all 4 sub-structs configured
- WHEN Install() is called
- THEN `a.paths.Base()` resolves the base directory
- AND `a.detector.IsInstalled()` checks installation status
- AND `a.prompt.Write(base, content)` writes the system prompt
- AND `a.backup.Generate(base, a.id)` creates the backup path
- AND no embedded struct methods appear on BaseAdapter's public API

#### Scenario: Sub-structs are independently testable
- GIVEN PathResolver, Detector, PromptManager, and BackupPathBuilder have their own constructors
- WHEN each is tested in isolation
- THEN no BaseAdapter construction is required
- AND test coverage for each extracted struct exceeds 80%

---

### REQ-ISP: ToolAdapter interface SHALL be segregated into role interfaces

The 12-method `ToolAdapter` interface SHALL be split into 4 role interfaces:

| Interface | Methods | Consumers |
|-----------|---------|-----------|
| `Identifier` | ID(), Name() | Logging, display |
| `Detector` | Detect(), IsInstalled() | Pre-flight checks |
| `Installer` | Install(InstallOpts), Uninstall(InstallOpts) | CLI commands |
| `AdapterPaths` | SkillsPath(), CommandsPath(), SystemPromptPath(), PromptStrategy() | Path queries |

A composite `ToolAdapter` interface SHALL embed all four role interfaces for backward compatibility with `Registry.Register()` and `Registry.Get()`.

#### Scenario: Consumer uses narrow Detector interface
- GIVEN a function signature `func preflightCheck(d Detector) bool`
- WHEN called with any adapter
- THEN only Detect() and IsInstalled() are accessible
- AND Install/Uninstall/SkillsPath are not in scope

#### Scenario: Registry stores composite ToolAdapter
- GIVEN a concrete adapter implementing Identifier + Detector + Installer + AdapterPaths
- WHEN `registry.Register(adapter)` is called
- THEN the adapter is stored and retrievable via `registry.Get(id)` as ToolAdapter
- AND all 4 role interfaces are accessible through the composite

---

### REQ-DI: Registry SHALL be injected via constructor, not global variable

`DefaultRegistry` SHALL remain as a deprecated backward-compat shim. All new code paths SHALL receive `*Registry` via constructor parameters. A `NewRegistry() *Registry` constructor SHALL be provided for tests and DI consumers.

#### Scenario: Test constructs local registry via constructor
- GIVEN a test function
- WHEN `reg := adapters.NewRegistry(); reg.Register(mock)` is called
- THEN no reference to `adapters.DefaultRegistry` exists
- AND `t.Parallel()` can be called without race conditions

#### Scenario: Main function constructs and injects registry
- GIVEN the `main()` function
- WHEN the application starts
- THEN `reg := adapters.NewRegistry()` is constructed
- AND adapter packages register via `RegisterIn(reg)` instead of `init() → DefaultRegistry`
- AND `reg` is threaded through `runStatus()`, `ScanTools()`, `runInstall()`, `runUninstall()`

---

### REQ-FACTORY: No convenience factory SHALL wrap registry access

The `NewAdapter(id)` factory function SHALL be removed. `factory.go` SHALL be deleted. Consumers that need an adapter by ID SHALL call `registry.Get(id)` directly.

#### Scenario: Consumer uses registry.Get directly
- GIVEN a CLI command that needs to find a specific adapter
- WHEN `adapter, err := registry.Get("claude-code")` is called
- THEN the adapter is retrieved without referencing factory.go or NewAdapter
- AND ErrUnknownAdapter is returned for unknown IDs

---

### REQ-TEST-ERRORS: Error-path tests SHALL cover Install failure modes

Tests SHALL cover at minimum: context cancellation, nil function fields, staging directory creation failure, template rendering failure, base resolution failure, version file write failure, and system prompt write failure with and without rollback.

#### Scenario: Context cancellation before any work
- GIVEN a BaseAdapter with valid config and a cancelled context
- WHEN Install(InstallOpts{Context: cancelledCtx}) is called
- THEN a context error is returned
- AND no directories or files are created

#### Scenario: Nil function fields handled gracefully
- GIVEN a BaseAdapter with nil detector, nil paths, or nil prompt manager
- WHEN the affected method is called
- THEN a clear error message is returned
- AND no panic propagates to the caller

---

### REQ-TEST-MOCKFS: Mock filesystem tests SHALL exercise template rendering

Tests SHALL use mock embed.FS instances to verify per-adapter template rendering within the Install pipeline. Each adapter's template SHALL have its own test suite proving template content isolation.

#### Scenario: Different adapters produce different template content
- GIVEN a Claude adapter with Claude-specific templates
- AND a Codex adapter with Codex-specific templates
- WHEN each adapter's Install is tested
- THEN each produces content matching its own template, not a shared template

## Coverage Summary

| Category | Requirements | Scenarios |
|----------|-------------|-----------|
| Composition pattern | 1 | 2 |
| Interface segregation | 1 | 2 |
| Dependency injection | 1 | 2 |
| Factory removal | 1 | 1 |
| Error-path tests | 1 | 2 |
| Mock FS tests | 1 | 1 |
| **Total** | **6** | **10** |
