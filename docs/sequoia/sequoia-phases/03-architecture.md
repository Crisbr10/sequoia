# P3 · Architecture Audit Report

**Target**: `C:\Users\Usuario\Documents\DEMO_APPS\sequoia-ai\sequoia-ai`
**Date**: 2026-05-12
**Agent**: P3 sequoia-architecture
**Stack**: Go 1.24.2, Cobra CLI, Bubbletea TUI, Lipgloss, embed.FS templates

---

## 1. Dependency Map

```
cmd/sequoia/main.go
  ├── adapters (Registry, DefaultRegistry, InstallOpts)
  ├── adapters/common (Version, IsSymlink, ResolveHome)
  ├── internal/app (Model, NewModel)
  └── adapters/claude, adapters/codex, adapters/cursor, adapters/gemini, adapters/opencode (blank import → init() auto-reg)

internal/app/model.go
  ├── adapters (ToolAdapter via DefaultRegistry)
  ├── internal/model (ToolState, Screen, TUIConfig, ProgressMsg)
  └── internal/tui/screens (ProgressTool)  ← via model.go

internal/app/update.go
  ├── internal/model
  ├── internal/pipeline (RunInstall, RunUninstall)
  ├── internal/tui (NavigateMsg)
  └── internal/tui/screens (screen update functions)

internal/app/view.go
  ├── internal/model
  ├── internal/tui/screens (View functions)
  └── internal/tui/styles

internal/model/types.go
  └── adapters (ToolAdapter)  ← INTERNAL DEPENDS ON ADAPTERS

internal/pipeline/runner.go
  ├── adapters (InstallOpts)
  └── internal/model (ToolState, ProgressMsg)

adapters/claude, opencode, cursor, gemini, codex, _template
  ├── adapters (ToolAdapter, PromptStrategy, InstallOpts, AdapterStatus, DefaultRegistry)
  └── adapters/common (Installer, RenderTemplate, StageFile, CommandFiles, Version)

internal/tui/router.go
  └── internal/model (Screen)

plugin/interface.go
  └── (NO IMPORTERS — orphan package)
```

### Key observations from the dependency graph:

| Direction | Description |
|-----------|-------------|
| ✅ `cmd` → `internal` + `adapters` | Correct: entry point depends on internals |
| ✅ `internal/app` → `internal/pipeline`, `internal/tui` | Correct: internal-to-internal |
| ❌ `internal/model` → `adapters` | **Violation**: internal package depends on non-internal |
| ⚠️ 5 adapter packages duplicated | 70-85% code overlap across 5 concrete adapters |
| ❌ `plugin/` | Orphan: defined, never imported |

---

## 2. Findings

### [P3-001] · Massive Adapter Code Duplication — 70-85% Overlap Across 5 Adapters  [🔴 CRÍTICO]
**State**: Confirmed
**Evidence**: Compare `adapters/claude/adapter.go:125-207`, `adapters/opencode/adapter.go:125-209`, `adapters/cursor/adapter.go:120-204`, `adapters/gemini/adapter.go:121-203`, `adapters/codex/adapter.go:126-214` — the `Install()` method follows the exact same 8-step pattern across all 5 adapters: resolve base → create staging dir → render skill template → stage command files → mkdir target dirs → install skills via `common.Installer` → install commands via `common.Installer` → inject/merge system prompt → write version file. The `Status()` method (`claude/adapter.go:105-122`, `opencode/adapter.go:104-122`, etc.) is 100% identical across all 5 adapters. The `Uninstall()` method shares the same file removal prefix pattern everywhere.
**Problem**: 5 concrete adapters contain approximately 150-180 lines of near-identical boilerplate each (~800 lines total). Adding a 6th tool requires copying the entire pattern. Fixing a bug in the staging/rollback logic requires touching all 5+1 files. This is a classic DRY violation that guarantees divergence over time.
**Real Impact**: Every new tool adapter adds ~150 lines of duplicated code. Bug fixes must be replicated in N places. Contributors following the `_template` directory will copy the boilerplate verbatim, perpetuating the problem.
**Minimal High-Leverage Recommendation**: Extract a `BaseAdapter` struct in `adapters/common/` that implements `Install()`, `Uninstall()`, `Status()`, and all path methods. Concrete adapters embed `BaseAdapter` and only override `ID()`, `Name()`, `Detect()`, `IsInstalled()`, `PromptStrategy()`, and the system prompt injection hook (a single method like `WriteSystemPrompt(base string) error`).
**Dependencies/Blockers**: None — `adapters/common` already contains the `Installer` framework used by all adapters.
**Implementation Risk**: Medium — requires adding a new exported type and modifying all 5 concrete adapters, but the interface and factory remain unchanged.
**Acceptance Criteria**:
- [ ] `BaseAdapter` exists in `adapters/common/` with shared `Install()`, `Uninstall()`, `Status()` logic
- [ ] Each concrete adapter embeds `BaseAdapter` and is ≤80 lines (from current ~200-250)
- [ ] All existing tests pass without modification
- [ ] No change to `ToolAdapter` interface

---

### [P3-002] · Strategy Installer Functions Duplicated in Adapter Packages — Should Be in common  [🔴 CRÍTICO]
**State**: Confirmed
**Evidence**: `adapters/claude/installer.go:17-51` (InjectSection) is a **100% byte-identical** copy of `adapters/gemini/installer.go:17-51`. `adapters/claude/installer.go:55-88` (RemoveSection) is a **100% byte-identical** copy of `adapters/gemini/installer.go:55-88`. `adapters/opencode/installer.go:18-45` (GenerateAgentsMD) is a **100% byte-identical** copy of `adapters/cursor/installer.go:18-45` (GenerateRulesMD) and `adapters/_template/installer.go:27-54`. `adapters/opencode/installer.go:51-82` (RemoveAgentsMD) ≡ `adapters/cursor/installer.go:51-82` (RemoveRulesMD) ≡ `adapters/_template/installer.go:62-93`. The `isSequoiaManaged()` helper is duplicated 4 times.
**Problem**: Three distinct injection strategies exist — MarkdownSections (InjectSection/RemoveSection), FileReplace (Generate*/Remove*), and TOMLMerge — but each is duplicated across adapter packages under different function names. The strategy is orthogonal to the tool identity and should be a parameter, not a package-local implementation.
**Real Impact**: A bug in the newline-trimming logic of `InjectSection` (line 37-38 of both files) must be fixed in 2 places. A security fix in the backup logic of `GenerateRulesMD` requires 3 changes. The `markerStart`/`markerEnd` constants are redefined in 4 packages.
**Minimal High-Leverage Recommendation**: Move all three injection strategies into `adapters/common/strategy.go` as exported functions: `InjectMarkdownSection(path, content string) error`, `RemoveMarkdownSection(path string) error`, `ReplaceFile(path, content string) error`, `RestoreOrRemoveFile(path string) error`. Delete the per-adapter `installer.go` files. Adapters call `common.InjectMarkdownSection(...)` directly.
**Dependencies/Blockers**: None — the strategies have zero adapter-specific logic.
**Implementation Risk**: Low — pure refactor; function bodies move unchanged, call sites update import paths.
**Acceptance Criteria**:
- [ ] `adapters/common/strategy.go` contains all 4 strategy functions with tests
- [ ] `adapters/{claude,gemini,cursor,opencode,codex}/installer.go` are deleted
- [ ] All adapter `Install()` methods call `common.*Strategy*` functions
- [ ] All existing tests pass

---

### [P3-003] · internal/model Imports adapters — Violates internal/ Encapsulation  [🟠 RIESGO]
**State**: Confirmed
**Evidence**: `internal/model/types.go:6` — `import "github.com/Crisbr10/sequoia/adapters"`. Line 36: `ToolState.Adapter adapters.ToolAdapter` stores a concrete interface reference from the non-internal `adapters` package.
**Problem**: The Go convention for `internal/` is that packages within it should not expose or depend on non-internal packages — they should define their own abstractions or depend only on other `internal/` packages. `internal/model` currently depends on `adapters`, coupling the shared domain types to the adapter layer. If `ToolAdapter` gains a method, `internal/model` must be re-evaluated even though `ToolState` only uses the interface structurally.
**Real Impact**: Moderate — `internal/model` cannot be reused or tested without importing `adapters`. Structural coupling means adapter interface changes propagate to model types. The `ToolState` type is not truly free of adapter knowledge.
**Minimal High-Leverage Recommendation**: Define a `ToolInfo` interface in `internal/model/` with only the methods `ToolState` actually needs (`ID() string`, `Name() string`, `IsInstalled() bool`). Use this narrower interface instead of `adapters.ToolAdapter`. This follows Interface Segregation and DIP.
**Dependencies/Blockers**: Requires a helper in `internal/app/model.go` to wrap `adapters.ToolAdapter` as `model.ToolInfo` when building `ToolState` from the registry.
**Implementation Risk**: Low — interface subset; `adapters.ToolAdapter` already satisfies all methods.
**Acceptance Criteria**:
- [ ] `internal/model/types.go` no longer imports `adapters`
- [ ] `ToolInfo` interface defined in `internal/model/` with `ID()`, `Name()`, `IsInstalled()`, `Detect()`, `Status()` methods
- [ ] `ToolState.Adapter` changes type to `ToolInfo`
- [ ] All compilation and tests pass

---

### [P3-004] · ScreenRouter Interface and TransitionMap Are Dead Code — Never Used  [🟠 RIESGO]
**State**: Confirmed
**Evidence**: `internal/tui/router.go:14-23` defines `TransitionMap` with all valid screen transitions. `router.go:116-124` defines a `ScreenRouter` interface. `router.go:126-151` implements `router` with `NavigateTo()` and `CurrentScreen()`. `router.go:132-134` provides `NewRouter()`. However, `grep -r "tui.NewRouter\|tui.ScreenRouter\|TransitionMap\|IsValidTransition\|NextScreen" internal/app/` returns **zero matches**. `app.Model` manages screen transitions directly in `update.go:51-67` via inline `NavigateMsg{}` emissions and hardcoded switch-case dispatch per screen.
**Problem**: 151 lines of TUI routing infrastructure (TransitionMap, ScreenRouter interface, router implementation, NavigateMsg, NextScreen, IsValidTransition) are completely unused by the production code. The `app.Model` duplicates transition logic in `update.go` through repeated switch-case blocks (screens.WelcomeUpdate → "install" action → NavigateMsg{ScreenToolSelection}, etc.). This creates two sources of truth for screen transitions.
**Real Impact**: Medium — a developer adding a new screen must update both the (unused) `TransitionMap` and the hardcoded transitions in `update.go` (and `view.go`). If they only update one, inconsistencies arise. The dead code occupies 151 lines of cognitive overhead.
**Minimal High-Leverage Recommendation**: Either (A) wire the `ScreenRouter` into `app.Model` and delete the inline NavigateMsg emission from `update.go`, or (B) delete `ScreenRouter`, `TransitionMap`, `NextScreen`, `IsValidTransition`, and `NewRouter` from `router.go`, keeping only `NavigateMsg` (which IS used). Option B is simpler and less risky.
**Dependencies/Blockers**: `NavigateMsg` is used by `update.go` and `view.go` — must be preserved in either case.
**Implementation Risk**: Low — dead code removal or routing delegation.
**Acceptance Criteria**:
- [ ] Either ScreenRouter is wired into Model, or dead router code is removed
- [ ] No duplicate transition logic remains
- [ ] All TUI navigation tests pass

---

### [P3-005] · app.Model Is an Incipient God Object — 19 Fields Spanning 5 Concerns  [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `internal/app/model.go:20-69` — the `Model` struct has 19 fields: (1) Version, (2) Screen, (3) Tools, (4) Config, (5-6) Width/Height, (7) Cursor, (8) ErrorMsg, (9) OperationMode, (10) PreviousScreen, (11) Progress, (12) ProgressTools, (13) InstallCompleted, (14) InstallFailed, (15) EngramAvailable, (16) UninstallConfirming, (17) Quitting, (18) ctx, (19) cancel. These span 5 distinct concerns: display state (Width, Height, Cursor), configuration (Config, EngramAvailable, Version), pipeline orchestration (Progress, ProgressTools, counters, ctx/cancel), navigation (Screen, PreviousScreen, OperationMode, Quitting, ErrorMsg), and UI interaction (UninstallConfirming). The companion `update.go` is 389 lines with 8 nested screen-specific switch blocks.
**Problem**: The Bubbletea `Model` pattern encourages a flat struct, but 19 fields with mixed concerns creates tight coupling between rendering, pipeline execution, and configuration. The `update.go` dispatch grows linearly with each new screen. Navigation transitions are hardcoded as inline `tea.Cmd` closures rather than dispatched through the ScreenRouter.
**Real Impact**: Adding a 9th screen will push `update.go` past 400 lines. Testing individual concerns requires constructing the full 19-field struct. Pipeline cancellation behavior (ctx/cancel) is entangled with TUI quit keybindings.
**Minimal High-Leverage Recommendation**: Split pipeline state into a `PipelineState` struct: `{Progress chan, ProgressTools, Completed, Failed, ctx, cancel}`. Extract navigation into a `NavigationState` struct: `{Screen, PreviousScreen, OperationMode, Quitting, UninstallConfirming, ErrorMsg}`. The `Model` delegates to these sub-structs. This reduces Model to ~10 fields without architectural changes.
**Dependencies/Blockers**: Requires changes to `update.go` field access patterns (e.g., `m.Screen` → `m.Nav.Screen`).
**Implementation Risk**: Medium — touching 389 lines of update dispatch.
**Acceptance Criteria**:
- [ ] `PipelineState` and `NavigationState` sub-structs exist
- [ ] Model has ≤12 direct fields
- [ ] All TUI tests pass
- [ ] `update.go` unchanged in structure (only field access paths)

---

### [P3-006] · Pipeline Step Granularity Is Cosmetic — 3 Steps Always Complete Atomically  [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `internal/pipeline/runner.go:81-89` sends 3 "running" `ProgressMsg` messages (Skills, Commands, System Prompt) **before** calling `adapter.Install(opts)` on line 92. Lines 108-117 send 3 "done" messages **after** `Install()` returns success. The adapter's `Install()` is a single monolithic call — the pipeline never tracks which internal step failed. On failure, line 96-106 marks only the first step ("Skills") as errored, regardless of where in the adapter the failure occurred. The `validateStepNames()` doc references "Prepare→Apply→Verify→Rollback" which is the `common.Installer` lifecycle, but the pipeline's 3-step names ("Skills", "Commands", "System Prompt") are different.
**Problem**: Users watching the progress screen see all 3 steps flicker from pending → running → done simultaneously (or 1 marked as errored). This provides zero meaningful progress feedback. The pipeline's 3-step model doesn't match the `Installer`'s 4-phase model or the adapter's internal structure. The step names are hardcoded in `defaultStepNames` (runner.go:18) and duplicated in `buildProgressTools` (update.go:302).
**Real Impact**: Users perceive instant completion or instant failure with no intermediate progress. The progress UI is misleading — it suggests granularity that doesn't exist. A slow network-mounted filesystem would still show all 3 steps completing instantly.
**Minimal High-Leverage Recommendation**: Replace the 3-step cosmetic model with a single-step model. The adapter `Install()` returns a single "installing" → "done"/"error" transition. Alternatively, make adapters report their own internal steps via a callback. The single-step approach is simpler and honest about the actual execution model.
**Dependencies/Blockers**: The `screens.ProgressTool` and `screens.InstallProgressView` must support single-step tools (currently they support multi-step).
**Implementation Risk**: Low — reduces complexity, doesn't add it.
**Acceptance Criteria**:
- [ ] Pipeline sends exactly 1 "running" + 1 "done"/"error" per tool (not 6 messages)
- [ ] Progress view correctly displays single-step progress
- [ ] `defaultStepNames` is deleted or reduced to `["Installing"]`
- [ ] Step name constants are defined in a single location (not duplicated in runner.go and update.go)

---

### [P3-007] · Typed Constants (Language, PersistenceBackend) Exist but Raw Strings Used at Consumption Points  [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `internal/model/types.go:79-98` defines `type Language string` with `LangEN`/`LangES` constants and `type PersistenceBackend string` with `PersistenceEngram`/`PersistenceFiles`/`PersistenceBoth` constants. However, `TUIConfig` (line 101-105) uses raw `string` for both fields: `Language string` and `Persistence string`. Similarly, `adapters.InstallOpts.Language` (interface.go:28) is a raw `string`. In `model.go:94`, the default is set with raw `"en"` and `"engram"` instead of `LangEN` and `PersistenceEngram`.
**Problem**: The typed constants provide compile-time safety, IDE autocompletion, and self-documentation — but none of these benefits are realized because the struct fields use `string`. Setting `Config.Language = "fr"` compiles without warning. The type system cannot prevent typos or invalid values.
**Real Impact**: Low — the bug surface is small because language/persistence are set from a constrained TUI menu, not free-form input. However, it's a missed opportunity for defensive design and loses documentation value.
**Minimal High-Leverage Recommendation**: Change `TUIConfig.Language` to `Language`, `TUIConfig.Persistence` to `PersistenceBackend`, and `InstallOpts.Language` to `Language`. Update the default initialization in `model.go:94` to use typed constants. The `Language` type may need to move to a shared location accessible by both `adapters` and `internal/model`.
**Dependencies/Blockers**: The `Language` type must be in a package accessible to both `adapters` (`InstallOpts`) and `internal/model` (`TUIConfig`). Consider moving it to `adapters/common/` (already imported by all relevant packages).
**Implementation Risk**: Low — changing field types; string literals still compile.
**Acceptance Criteria**:
- [ ] `TUIConfig.Language` type is `Language` (not `string`)
- [ ] `TUIConfig.Persistence` type is `PersistenceBackend` (not `string`)
- [ ] `InstallOpts.Language` type is `Language` (not `string`)
- [ ] Default initialization uses typed constants
- [ ] All tests compile and pass

---

### [P3-008] · InstallOpts Used for Both Install and Uninstall — Naming Coupling Smell  [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `adapters/interface.go:56-59` — `Uninstall(opts InstallOpts) error`. All 5 concrete adapters discard `opts.Language` with `_ = opts.Language` at the top of `Uninstall()` (e.g., `claude/adapter.go:211`, `opencode/adapter.go:213`, `gemini/adapter.go:207`). The type name `InstallOpts` suggests it's for installation, but it's also the parameter type for uninstall.
**Problem**: The `Uninstall` method accepts a struct named for its inverse operation. This is confusing for API consumers and suggests that uninstall might someday need install-specific options. If `InstallOpts` grows an install-only field, all `Uninstall` implementations must add a discard line.
**Minimal High-Leverage Recommendation**: Either (A) define a separate `UninstallOpts` struct (currently empty, future-proof), or (B) rename `InstallOpts` to `AdapterOpts` to clarify it's shared between both operations. Option B is simpler since the struct only has `Language` which is relevant to both install and uninstall.
**Dependencies/Blockers**: This touches the `ToolAdapter` interface — all 5 adapters + template need updating.
**Implementation Risk**: Low — rename + recompile.
**Acceptance Criteria**:
- [ ] `InstallOpts` renamed to `AdapterOpts` (or separate `UninstallOpts` type)
- [ ] All adapter `Install()` and `Uninstall()` signatures updated
- [ ] `_ = opts.Language` discard lines removed from `Uninstall()` if Language is genuinely used
- [ ] Pipeline passes the same opts to both operations

---

### [P3-009] · plugin Package Is Orphaned — Defined but Never Imported  [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `plugin/interface.go:7-42` defines a `Plugin` interface and `Agent` struct. A grep for `"github.com/Crisbr10/sequoia/plugin"` across all `.go` files (excluding `plugin/` itself) returns **zero matches**. The `plugin` package is never imported by `cmd/`, `internal/`, `adapters/`, or any test file. There is no loader, no registry, and no init() that references it.
**Problem**: Dead package — 42 lines of interface definitions with no consumers. The package comment describes "file-based plugin loader" and "scanning for .sequoia-plugin.yaml manifest files" but none of this machinery exists. If plugin support is planned for a future release, the stubs are premature and will rot without tests.
**Real Impact**: None currently — it's dead code. However, it occupies mental space in the module tree and may confuse contributors who assume it's functional.
**Minimal High-Leverage Recommendation**: Either (A) delete the `plugin/` directory until plugin loading is implemented, or (B) add a package-level comment: `// Package plugin is a future extension point. Not yet wired. DO NOT USE.` Option A is cleaner.
**Dependencies/Blockers**: None — zero consumers.
**Implementation Risk**: None — dead code removal.
**Acceptance Criteria**:
- [ ] `plugin/` directory is removed from the tree, OR has a clear "not yet implemented" doc comment
- [ ] All compilation succeeds

---

### [P3-010] · Single Sentinel Error for Entire Adapter Layer — No Error Taxonomy  [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `adapters/errors.go:5-7` defines a single sentinel error `ErrUnknownAdapter`. All other errors in the codebase are `fmt.Errorf("install: %w", err)` wrapped ad-hoc strings. There are no sentinel errors for: install failure, uninstall failure, detection failure, pipeline cancellation, permission denied, disk full, template rendering failure, or backup restoration failure.
**Problem**: Callers cannot programmatically distinguish error types. For example, `cmd/sequoia/main.go:226` wraps `a.Install()` errors with `%w`, but the caller in `runInstall` can only check string prefixes or assume all errors are fatal. A "permission denied" error during install gets the same treatment as a "template parse error" — both abort the entire operation and show the raw error string to the user.
**Real Impact**: Low currently — the CLI has only ~4 commands and error handling is simple. Will become painful as the tool gains headless/CI modes where error classification matters for exit codes and retry logic.
**Minimal High-Leverage Recommendation**: Add `ErrInstallFailed`, `ErrUninstallFailed`, and `ErrNotDetected` sentinel errors to `adapters/errors.go`. Wrap adapter errors with `fmt.Errorf("%w: %w", ErrInstallFailed, err)` so callers can use `errors.Is()`. Define a `RetryableError` interface for transient failures (network, permission).
**Dependencies/Blockers**: None — additive change.
**Implementation Risk**: Low — sentinel errors don't break existing code.
**Acceptance Criteria**:
- [ ] `adapters/errors.go` contains at least 3 sentinel errors
- [ ] `Install()` and `Uninstall()` wrap their errors with sentinels
- [ ] CLI error handling uses `errors.Is()` for exit code selection
- [ ] Tests verify error wrapping behavior

---

### [P3-011] · Cursor Adapter SkillsPath/CommandsPath Collide — Both Return Same Directory  [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `adapters/cursor/paths.go:29-32` — both `skillsPath()` and `commandsPath()` return `base` (which is `~/.cursor/rules/`). This means `common.Installer` for skills writes `SKILL.md` directly to `~/.cursor/rules/SKILL.md`, and the command installer writes command files to the same directory. If any command filename collides with an existing file in that directory, the backup/restore logic will interact unexpectedly.
**Problem**: Cursor's adapter treats a shared rules directory as both the skills and commands target. This works because the command files have unique names, but it means the `common.Installer` for commands will back up and potentially overwrite files that are NOT Sequoia-managed but happen to exist in `~/.cursor/rules/`. The `Prepare()` phase backs up any matching file in the target directory regardless of whether Sequoia created it. If the user has a custom `sequoia-init.md` in `~/.cursor/rules/`, the Installer will back it up and replace it.
**Real Impact**: Low — Cursor's rules directory is typically empty except for Cursor-generated files, but the risk is real for users who organize their own rule files there. The backup would be restored on rollback, but the intermediate state could confuse the user.
**Minimal High-Leverage Recommendation**: Add a `sequoia/` subdirectory for skills and commands in Cursor's adapter: `skillsPath(base) → filepath.Join(base, "sequoia", "skills")`, `commandsPath(base) → filepath.Join(base, "sequoia", "commands")`. This isolates Sequoia-managed files from user files and matches the pattern used by Gemini and Codex adapters.
**Dependencies/Blockers**: Requires updating the Cursor template paths and system prompt references.
**Implementation Risk**: Medium — changes the on-disk layout for Cursor users; existing installations would need migration.
**Acceptance Criteria**:
- [ ] `skillsPath()` and `commandsPath()` return distinct, Sequoia-scoped directories
- [ ] User files in `~/.cursor/rules/` are never backed up or overwritten by Sequoia
- [ ] Cursor adapter tests verify the isolated directory structure

---

### [P3-012] · Step Name Constants Duplicated Across runner.go and update.go  [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `internal/pipeline/runner.go:18` defines `defaultStepNames = []string{"Skills", "Commands", "System Prompt"}`. `internal/app/update.go:302` defines the identical list `stepNames := []string{"Skills", "Commands", "System Prompt"}` inside `buildProgressTools()`. `update.go:327` defines it again inside `buildUninstallProgressTools()`. The step count (3) is implicit in slice length — if a step is added, both files must be updated.
**Problem**: The step names are the contract between the pipeline (which sends named ProgressMsg) and the progress screen (which correlates by name). Duplicating the list creates a synchronization risk: if `runner.go` adds a 4th step but `update.go` only lists 3, the progress screen silently ignores the 4th step's messages.
**Real Impact**: Low — the step list changes rarely, but the duplication adds unnecessary maintenance burden and is exactly the kind of silent desync bug that integration tests might miss.
**Minimal High-Leverage Recommendation**: Define step names as an exported constant slice in `internal/pipeline/` (e.g., `var InstallSteps = []string{...}`). Both `runner.go` and `update.go` reference `pipeline.InstallSteps`. This is a single source of truth.
**Dependencies/Blockers**: None — `internal/app` already imports `internal/pipeline`.
**Implementation Risk**: None — rename reference only.
**Acceptance Criteria**:
- [ ] Single exported `InstallSteps` slice in `internal/pipeline/`
- [ ] `update.go` removes duplicated `stepNames` variables
- [ ] `runner.go` references the same constant

---

## 3. Architecture Health Summary

| Dimension | Score | Notes |
|-----------|-------|-------|
| Module Boundaries | 6/10 | `internal/model` leaks by importing `adapters`. Otherwise reasonable. |
| Interface Design | 7/10 | `ToolAdapter` is well-scoped but `InstallOpts` naming hurts. `ScreenRouter` is dead code. |
| Circular Dependencies | 10/10 | No cycles detected in import graph. |
| Code Duplication | 2/10 | Critical: 5 adapters share 70-85% identical code. Strategy functions duplicated 2-4x. |
| Dependency Direction | 7/10 | Flow `cmd → internal → adapters` is correct, but `internal/model → adapters` breaks internal encapsulation. |
| God Objects | 6/10 | `app.Model` (19 fields) is the only incipient god object. Manageable now, will degrade with more screens. |
| Error Handling | 4/10 | Single sentinel error. No error taxonomy. All errors are opaque `fmt.Errorf` wraps. |
| Configuration Cohesion | 7/10 | `TUIConfig` is centralized; version in `adapters/common` is correct. Typed constants underused. |
| Orphan Code | 6/10 | `plugin/` package is dead. `ScreenRouter`/`TransitionMap` are dead. |

### High-Priority Remediation Sequence

1. **P3-002 + P3-001**: Extract `BaseAdapter` and move strategies to `common/` — eliminates ~700 lines of duplication
2. **P3-003**: Break `internal/model → adapters` dependency — restores `internal/` encapsulation
3. **P3-004**: Resolve dead `ScreenRouter` code — either wire it or delete it
4. **P3-005**: Split `Model` into sub-structs — prevents future god object growth
5. **P3-006**: Fix cosmetic pipeline steps — honest UX

---

## 4. Template

```yaml
phase: "03-architecture"
agent: "P3 sequoia-architecture"
target: "C:\Users\Usuario\Documents\DEMO_APPS\sequoia-ai\sequoia-ai"
timestamp: "2026-05-12"
stack: "Go 1.24.2, Cobra, Bubbletea, Lipgloss"
paradigm: "CLI Tool + Adapter Pattern + Plugin Framework"
metrics:
  total_go_files: 55
  total_modules: 10
  adapter_packages: 6
  internal_packages: 4
  largest_file: "internal/app/update.go (389 lines)"
  duplication_ratio: "~70% across adapters"
  circular_deps: 0
  sentinel_errors: 1
  dead_packages: 1 (plugin)
  dead_interfaces: 1 (ScreenRouter)
health_score: 58/100
findings_critical: 2
findings_risk: 2
findings_attention: 8
```
