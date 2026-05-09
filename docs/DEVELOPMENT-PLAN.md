# Sequoia AI — Development Plan

> **Status**: Active  
> **Version**: 1.0  
> **Last updated**: 2026-05-08  
> **Module**: `sequoia-ai`

This is the authoritative task reference for building Sequoia. All implementation decisions are recorded here. Do not start work on a task without reading its dependencies first.

---

## Resolved Architecture Decisions

| Question | Decision | Rationale |
|----------|----------|-----------|
| Persistence backend | Engram if available, openspec as fallback | Engram is cross-session; openspec provides local recovery when Engram MCP is not running |
| Token budget / chunking | Implemented at `/sequoia init` level | Init generates the Project Map — it estimates project complexity there and creates chunked scopes if needed. Audit commands honor chunks. Cleanest insertion point with zero impact on audit logic. |
| Sub-agent delegation | Claude Code `Agent` tool calls from C0 | Claude Code supports formal Agent tool invocations inside skills. C0 (orchestrator) delegates to P1–P6 via Agent calls. No workarounds needed. |
| Go module name | `sequoia-ai` | |
| Primary language | Go 1.22+ | Cross-platform binary, single executable, Bubbletea ecosystem for TUI |
| TUI scope | Install / config / status ONLY | All audit logic lives in editor commands. This is a hard rule — never add audit features to the TUI. |

---

## Tech Stack

| Layer | Choice |
|-------|--------|
| Language | Go 1.22+ |
| TUI | `github.com/charmbracelet/bubbletea` + `lipgloss` + `bubbles` |
| CLI | `github.com/spf13/cobra` |
| Config | `gopkg.in/yaml.v3` |
| Testing | stdlib `testing` + `testify` + `teatest` (golden files) |
| Distribution | GoReleaser |
| Logging | `github.com/charmbracelet/log` |

---

## Project Structure

```
sequoia-ai/
├── cmd/
│   └── sequoia/
│       └── main.go
├── adapters/
│   ├── interface.go
│   ├── registry.go
│   ├── factory.go
│   ├── common/
│   │   └── installer.go
│   ├── claude/
│   │   ├── adapter.go
│   │   ├── paths.go
│   │   ├── installer.go
│   │   └── templates/
│   ├── opencode/
│   │   ├── adapter.go
│   │   ├── paths.go
│   │   ├── installer.go
│   │   └── templates/
│   └── _template/
├── internal/
│   ├── app/
│   ├── tui/
│   │   └── screens/
│   ├── model/
│   └── pipeline/
└── scripts/
    ├── install.sh
    └── install.ps1
```

---

## Adapter Pattern

All tool integrations implement the same interface:

```
ToolAdapter interface
    → registry (All, Register, Get)
    → factory (NewAdapter)
    → common/installer (Prepare → Apply → Verify → Rollback)
    → tool-specific (paths, templates, prompt strategy)
```

Two prompt strategies:
- `StrategyMarkdownSections` — inject section with markers (Claude Code, `CLAUDE.md`)
- `StrategyFileReplace` — full file replace with backup (OpenCode, `AGENTS.md`)

---

## Phases and Tasks

### Phase 1 — Foundation ✅ COMPLETED 2026-05-09

**Goal**: Spec consistency + Go infrastructure before any adapter code.

---

#### T-001 — Unify scoring system ✅

- **Effort**: M | **Priority**: P1 | **Deps**: none
- **Description**: Three competing scoring methodologies exist in `ARCHITECTURE.md`, `references/scoring-criteria.md`, and `SKILL.md`. Reconcile into one system. Emojis are presentation layer only — not core scoring logic.
- **Acceptance criteria**:
  - [x] Single formula: `score = 100 - Σ(severity_weight × scope_multiplier)`
  - [x] Weights: critical=15, high=8, medium=4, low=2, info=0
  - [x] Scope multiplier: 1.0 (isolated) | 1.5 (shared root cause)
  - [x] Emoji mapping: 🔴 critical · 🟠 high · 🟡 medium · 🟢 low · 🔵 info
  - [x] `SKILL.md`, `ARCHITECTURE.md`, `scoring-criteria.md` all reference the same system

---

#### T-002 — Fix README and reference drift ✅

- **Effort**: S | **Priority**: P1 | **Deps**: none
- **Description**: `README.md` references files that do not exist. Update all paths to match the actual file tree.
- **Acceptance criteria**:
  - [x] All paths in `README.md` resolve to real files
  - [x] No references to non-existent files (`flows/re-audit.md`, `flows/quick-check.md`, `references/health-score.md`, `references/project-map.md`, `references/report-template.md`)
  - [x] Agent count matches actual (C0 + P1–P6 + M1–M2 = 9)

---

#### T-003 — Create Project Map reference schema with chunking support ✅

- **Effort**: M | **Priority**: P1 | **Deps**: none
- **Description**: Extract the Project Map YAML schema from `ARCHITECTURE.md` and `sequoia-context.md` into a standalone reference document. Include chunk definitions for token budget management — when `/sequoia init` detects a large project, it writes chunked scopes into the Project Map. Audit commands read these scopes and process chunks sequentially.
- **Acceptance criteria**:
  - [x] New file: `references/project-map.md`
  - [x] Complete schema with all fields documented
  - [x] `chunks:` field added — list of named scopes with file glob patterns
  - [x] `token_budget:` field — estimated complexity, `chunked: true/false`
  - [x] Init command logic documented: estimate size → if > threshold → generate chunks
  - [x] Example Project Maps for 3 project types (frontend, backend, fullstack) including one chunked example

---

#### T-004 — Initialize Go module ✅

- **Effort**: S | **Priority**: P1 | **Deps**: none
- **Description**: Set up Go module, tooling configuration, and directory skeleton.
- **Acceptance criteria**:
  - [x] `go.mod` with module `sequoia-ai`, Go 1.22+
  - [x] `.golangci.yaml` linting config
  - [x] All directories from the project structure above created
  - [x] `go.sum` (empty initially)

---

#### T-005 — Implement ToolAdapter interface, registry, and factory ✅

- **Effort**: L | **Priority**: P1 | **Deps**: T-004
- **Description**: Core adapter abstraction. Every tool integration implements this interface.
- **Acceptance criteria**:
  - [x] `adapters/interface.go` — `ToolAdapter` interface with: `ID()`, `Name()`, `Detect()`, `IsInstalled()`, `Install()`, `Uninstall()`, `Status()`, `SkillsPath()`, `CommandsPath()`, `SystemPromptPath()`, `PromptStrategy()`
  - [x] `adapters/registry.go` — `Registry` with `All()`, `Register()`, `Get(id)` methods
  - [x] `adapters/factory.go` — `NewAdapter(id) (ToolAdapter, error)`
  - [x] Unit tests for registry, factory, and interface contract
  - [x] Godoc on all exported symbols

---

#### T-006 — Common installer framework (Prepare → Apply → Verify → Rollback) ✅

- **Effort**: M | **Priority**: P1 | **Deps**: T-005
- **Description**: Generic installer that all adapters share. Four phases are always executed in order; any failure triggers rollback.
- **Acceptance criteria**:
  - [x] `adapters/common/installer.go` — `Installer` type
  - [x] `Prepare()`: generates files from templates, validates paths, checks write permissions
  - [x] `Apply()`: copies files, injects system prompts, updates configs
  - [x] `Verify()`: confirms all expected files exist and are readable
  - [x] `Rollback()`: restores backup created during Prepare
  - [x] Tests covering happy path, Apply failure, and Verify failure

---

#### T-007 — Create full directory structure ✅

- **Effort**: S | **Priority**: P1 | **Deps**: T-004
- **Description**: Create all required directories with README stubs explaining each adapter's role.
- **Acceptance criteria**:
  - [x] All directories from the project structure section exist
  - [x] `adapters/claude/README.md`, `adapters/opencode/README.md`, `adapters/_template/README.md` each describe their purpose in 2–3 sentences
  - [x] No empty stub `.go` files

---

### Phase 2 — Claude Code Integration

**Goal**: Full installation pipeline for `~/.claude/` working and tested.

---

#### T-008 — Claude Code adapter

- **Effort**: M | **Priority**: P1 | **Deps**: T-005, T-006
- **Description**: Adapter that knows Claude Code's file layout and installation strategy.
- **Acceptance criteria**:
  - [x] `adapters/claude/adapter.go` implements `ToolAdapter`
  - [x] `adapters/claude/paths.go` defines all paths under `~/.claude/`
  - [x] `Detect()` checks for Claude Code binary or config directory
  - [x] `IsInstalled()` checks for existing Sequoia markers in `CLAUDE.md`
  - [x] `PromptStrategy()` returns `StrategyMarkdownSections`
  - [x] Unit tests using temp directories for all path methods

---

#### T-009 — Claude Code templates

- **Effort**: M | **Priority**: P1 | **Deps**: T-008
- **Description**: All Markdown templates for Claude Code: skill, commands, and `CLAUDE.md` section. Commands must include chunking awareness (read `chunks:` from Project Map if present).
- **Acceptance criteria**:
  - [x] `adapters/claude/templates/skill.md.tmpl` — `SKILL.md` with Claude frontmatter
  - [ ] `adapters/claude/templates/commands/sequoia-init.md` — init command with chunking logic
  - [x] `adapters/claude/templates/commands/sequoia-audit.md`
  - [x] `adapters/claude/templates/commands/sequoia-review.md`
  - [x] `adapters/claude/templates/commands/sequoia-fix.md`
  - [x] `adapters/claude/templates/commands/sequoia-diff.md`
  - [x] `adapters/claude/templates/claude-md-section.md.tmpl` — section to inject into `CLAUDE.md`
  - [x] Golden file tests for all templates

---

#### T-010 — CLAUDE.md section injection

- **Effort**: M | **Priority**: P1 | **Deps**: T-009
- **Description**: Safely inject and remove the Sequoia section from `~/.claude/CLAUDE.md` using start/end markers.
- **Acceptance criteria**:
  - [x] `adapters/claude/installer.go` — `InjectSection()` and `RemoveSection()` methods
  - [x] Markers: `<!-- sequoia:start -->` … `<!-- sequoia:end -->`
  - [x] Idempotent: running twice produces identical result
  - [x] Non-destructive: all existing content preserved
  - [x] Handles three cases: file missing, markers absent, markers present
  - [x] Tests for all three cases + idempotency

---

#### T-011 — Claude Code full installation pipeline

- **Effort**: L | **Priority**: P1 | **Deps**: T-010, T-006
- **Description**: Wire the common installer framework to the Claude adapter. Runs the full Prepare → Apply → Verify → Rollback cycle.
- **Acceptance criteria**:
  - [x] `Install()` creates `~/.claude/skills/sequoia/SKILL.md`
  - [x] `Install()` creates `~/.claude/commands/sequoia-*.md` (5 files)
  - [x] `Install()` injects section into `~/.claude/CLAUDE.md`
  - [x] `Verify()` confirms all files readable
  - [x] `Rollback()` restores previous state on any error
  - [x] Full pipeline tested in temp directory

---

#### T-012 — Claude Code end-to-end test

- **Effort**: L | **Priority**: P1 | **Deps**: T-011
- **Description**: Manual validation that Sequoia works in a real Claude Code session. Tests chunking behavior on both small and large repositories.
- **Acceptance criteria**:
  - [ ] Test project set up (two repos: small < 50 files, large > 200 files)
  - [ ] `/sequoia init` generates valid Project Map; large project produces chunked map
  - [ ] `/sequoia audit` activates at least 4 agents and produces findings
  - [ ] Findings reference real files with `file:line` format
  - [ ] Health Score calculated and displayed
  - [ ] Chunked audit processes all chunks and merges findings
  - [ ] All 5 commands accessible and functional
  - [ ] Findings documented in a test report

---

### Phase 3 — OpenCode Integration

**Goal**: Parallel installation pipeline for `~/.config/opencode/`.

---

#### T-013 — OpenCode adapter

- **Effort**: M | **Priority**: P1 | **Deps**: T-005, T-006
- **Description**: Adapter for OpenCode with `StrategyFileReplace` prompt strategy.
- **Acceptance criteria**:
  - [x] `adapters/opencode/adapter.go` implements `ToolAdapter`
  - [x] `adapters/opencode/paths.go` defines all paths under `~/.config/opencode/`
  - [x] `PromptStrategy()` returns `StrategyFileReplace`
  - [x] Unit tests parallel to Claude adapter

---

#### T-014 — OpenCode templates

- **Effort**: M | **Priority**: P1 | **Deps**: T-013
- **Description**: Templates for OpenCode. All 9 agents are inlined in the skill (no sub-files). Commands same as Claude but with OpenCode frontmatter.
- **Acceptance criteria**:
  - [x] `adapters/opencode/templates/skill.md.tmpl` — all 9 agents inline
  - [x] `adapters/opencode/templates/commands/` — 5 command files
  - [x] `adapters/opencode/templates/agents-md-section.md.tmpl` — `AGENTS.md` content
  - [x] Golden file tests

---

#### T-015 — AGENTS.md generation

- **Effort**: M | **Priority**: P1 | **Deps**: T-014
- **Description**: Generate or update `AGENTS.md` for OpenCode. Backs up existing file before overwrite.
- **Acceptance criteria**:
  - [x] `adapters/opencode/installer.go` — `GenerateAgentsMD()` method
  - [x] Backs up existing `AGENTS.md` before writing
  - [x] Three cases handled: file missing, exists with Sequoia section, exists with other content
  - [x] `RemoveSection()` removes only Sequoia content and restores backup on uninstall
  - [x] Tests for all three cases

---

#### T-016 — OpenCode full installation pipeline

- **Effort**: L | **Priority**: P1 | **Deps**: T-015, T-006
- **Description**: Full pipeline for OpenCode. Same structure as T-011.
- **Acceptance criteria**:
  - [x] `Install()` creates `~/.config/opencode/skills/sequoia/`
  - [x] `Install()` creates `~/.config/opencode/commands/`
  - [x] `Install()` generates `AGENTS.md`
  - [x] `Verify()` + `Rollback()` working
  - [x] Tests in temp directory

---

#### T-017 — OpenCode end-to-end test

- **Effort**: L | **Priority**: P1 | **Deps**: T-016
- **Description**: Same test suite as T-012 applied to OpenCode.
- **Acceptance criteria**:
  - [ ] All acceptance criteria from T-012 met on OpenCode
  - [ ] Documented test report

---

#### T-018 — Refactor shared logic

- **Effort**: M | **Priority**: P2 | **Deps**: T-016
- **Description**: Extract duplicated code from Claude and OpenCode installers into `adapters/common/`.
- **Acceptance criteria**:
  - [ ] No duplicated logic between adapters
  - [ ] No behavioral change (all existing tests still pass)

---

### Phase 4 — CLI Installer

**Goal**: Headless `sequoia` binary usable without TUI.

---

#### T-019 — CLI base with Cobra

- **Effort**: M | **Priority**: P1 | **Deps**: T-018
- **Description**: Main entry point. All subcommands wired. TUI launches from `sequoia install` when stdin is a terminal.
- **Acceptance criteria**:
  - [x] `cmd/sequoia/main.go`
  - [x] `sequoia install [--tool=<id>] [--no-tui]`
  - [x] `sequoia status`
  - [x] `sequoia uninstall [--tool=<id>] [--all]`
  - [x] `sequoia version`
  - [x] Help text for all commands and flags

---

#### T-020 — Multi-tool detection

- **Effort**: M | **Priority**: P1 | **Deps**: T-019
- **Description**: Scan home directory for all supported tools and report installation state.
- **Acceptance criteria**:
  - [x] Detects Claude Code, OpenCode on macOS, Linux, and Windows
  - [x] Reports: tool name, installation path, Sequoia installed (yes/no), Sequoia version
  - [ ] Works correctly with symlinked paths
  - [x] Tests per OS using temp directories

---

#### T-021 — Uninstall command

- **Effort**: M | **Priority**: P2 | **Deps**: T-020
- **Description**: Safe removal of Sequoia from one or all tools.
- **Acceptance criteria**:
  - [x] `sequoia uninstall --tool=claude-code` removes skill, commands, and `CLAUDE.md` section
  - [x] `sequoia uninstall --all` loops all installed tools
  - [x] Restores backups where applicable
  - [x] Prompts for confirmation unless `--yes` flag passed
  - [x] Tests for each tool

---

#### T-022 — One-line installer scripts

- **Effort**: S | **Priority**: P2 | **Deps**: T-019
- **Description**: Shell scripts for zero-dependency installation via `curl | bash` and `irm | iex`.
- **Acceptance criteria**:
  - [x] `scripts/install.sh` — detects OS/arch, downloads correct binary, verifies SHA-256, runs `sequoia install`
  - [x] `scripts/install.ps1` — same for Windows PowerShell
  - [x] Both scripts handle existing installations gracefully
  - [x] Both tested on each platform

---

#### T-023 — Cross-platform testing

- **Effort**: L | **Priority**: P1 | **Deps**: T-022
- **Description**: CI matrix that validates install/status/uninstall on all three OS.
- **Acceptance criteria**:
  - [x] GitHub Actions matrix: `ubuntu-latest`, `macos-latest`, `windows-latest`
  - [x] Install → status → uninstall cycle passes on all three
  - [x] Path separators handled correctly everywhere
  - [x] All tests green in CI

---

### Phase 5 — TUI Installer

**Goal**: Interactive Bubbletea TUI for installation. Audit logic stays in the editor — never here.

---

#### T-024 — Bubbletea architecture

- **Effort**: L | **Priority**: P1 | **Deps**: T-019
- **Description**: Central model, update, view, and screen router. Follows the same patterns as Gentle-AI.
- **Acceptance criteria**:
  - [ ] `internal/app/model.go` — root model with screen state
  - [ ] `internal/app/update.go` — message dispatch
  - [ ] `internal/app/view.go` — screen rendering delegation
  - [ ] `internal/tui/router.go` — screen transitions
  - [ ] `internal/model/` — domain types (InstalledTool, InstallResult, etc.)
  - [ ] Skeleton compiles and runs (blank screen is fine)

---

#### T-025 — Welcome screen

- **Effort**: M | **Priority**: P1 | **Deps**: T-024
- **Description**: Entry screen. Shows branding, version, and lists auto-detected tools.
- **Acceptance criteria**:
  - [ ] `internal/tui/screens/welcome.go`
  - [ ] Displays Sequoia name, version, and tagline
  - [ ] Lists detected tools with install status
  - [ ] `Enter` or `→` transitions to Tool Selection
  - [ ] Golden file test

---

#### T-026 — Tool Selection screen

- **Effort**: M | **Priority**: P1 | **Deps**: T-025
- **Description**: Checkbox list of detected tools. Multi-select. At least one must be chosen to proceed.
- **Acceptance criteria**:
  - [ ] `internal/tui/screens/tool-selection.go`
  - [ ] Shows each detected tool with `[x]` / `[ ]` toggle
  - [ ] `Space` toggles selection, `Enter` proceeds
  - [ ] Validation: shows error if zero tools selected
  - [ ] Transitions to Configuration screen
  - [ ] Golden file test

---

#### T-027 — Configuration screen

- **Effort**: M | **Priority**: P2 | **Deps**: T-026
- **Description**: User preferences: language and persistence backend.
- **Acceptance criteria**:
  - [ ] `internal/tui/screens/configuration.go`
  - [ ] Language selector: English / Español
  - [ ] Persistence selector: Engram (if detected) / Files / Both
  - [ ] Engram option disabled with note if MCP not detected
  - [ ] Transitions to Install Progress
  - [ ] Golden file test

---

#### T-028 — Install Progress screen

- **Effort**: L | **Priority**: P1 | **Deps**: T-027
- **Description**: Visual step-by-step progress for each selected tool. Runs actual installation in a goroutine and receives progress messages.
- **Acceptance criteria**:
  - [ ] `internal/tui/screens/install-progress.go`
  - [ ] Per-tool progress: `[ ] Skills` → `[✓] Skills` → `[✓] Commands` → `[✓] System Prompt`
  - [ ] Spinner while step is running
  - [ ] Error state per step with message
  - [ ] Transitions to Complete or Error screen on finish
  - [ ] Tested with a mock installer

---

#### T-029 — Complete and Error screens

- **Effort**: M | **Priority**: P1 | **Deps**: T-028
- **Description**: Post-installation summary.
- **Acceptance criteria**:
  - [ ] `internal/tui/screens/complete.go` — lists what succeeded, shows first command to try
  - [ ] `internal/tui/screens/error.go` — shows what failed, error message, retry option
  - [ ] `r` retries failed tools, `q` exits
  - [ ] Golden file tests

---

#### T-030 — Status screen

- **Effort**: M | **Priority**: P2 | **Deps**: T-026
- **Description**: Shows current installation state for all detected tools.
- **Acceptance criteria**:
  - [ ] `internal/tui/screens/status.go`
  - [ ] Per tool: name, installed (✅/❌), Sequoia version, installation path
  - [ ] Options: `u` update, `r` reinstall, `d` uninstall, `q` quit
  - [ ] Golden file test

---

#### T-031 — Uninstall screen

- **Effort**: M | **Priority**: P2 | **Deps**: T-030
- **Description**: Interactive uninstall with confirmation.
- **Acceptance criteria**:
  - [ ] `internal/tui/screens/uninstall.go`
  - [ ] Checkbox list of installed tools
  - [ ] Confirmation step before executing
  - [ ] Progress screen reused from T-028 (with uninstall messages)
  - [ ] Complete screen shows what was removed
  - [ ] Golden file test

---

#### T-032 — Connect TUI to installation pipeline

- **Effort**: M | **Priority**: P1 | **Deps**: T-024, T-023
- **Description**: Wire screen actions to real adapter calls. Error handling flows back to the UI.
- **Acceptance criteria**:
  - [ ] Progress screen calls real `Install()` methods
  - [ ] Errors from adapters surface in Error screen
  - [ ] Retry triggers a new install pipeline run
  - [ ] Rollback on critical error confirmed in UI
  - [ ] Full user flow tested end-to-end (Welcome → Complete)

---

#### T-033 — GoReleaser configuration

- **Effort**: M | **Priority**: P1 | **Deps**: T-032
- **Description**: Cross-platform binary builds and GitHub release automation.
- **Acceptance criteria**:
  - [ ] `.goreleaser.yaml` builds: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`
  - [ ] SHA-256 checksums file generated
  - [ ] GitHub releases configured with changelog
  - [ ] Homebrew formula template
  - [ ] Scoop manifest template
  - [ ] CI pipeline triggers on `v*` tags

---

#### T-034 — TUI test suite

- **Effort**: L | **Priority**: P1 | **Deps**: T-031
- **Description**: Golden file tests for all screens plus integration flow tests.
- **Acceptance criteria**:
  - [ ] Golden file test for each screen (25×80 and 40×120 terminal sizes)
  - [ ] Interaction tests: keypress sequences for each screen
  - [ ] Happy path flow test: Welcome → Complete
  - [ ] Error recovery test: install fails → Error screen → retry → Complete
  - [ ] Cross-platform path tests
  - [ ] Coverage > 80% on `internal/tui/`

---

### Phase 6 — Extensibility and Release

**Goal**: More adapters, plugin system, documentation, and v0.1.0 release.

---

#### T-035 — Cursor adapter

- **Effort**: M | **Priority**: P2 | **Deps**: T-005, T-006
- **Description**: Adapter for Cursor IDE (`.cursor/rules/`).
- **Acceptance criteria**:
  - [ ] `adapters/cursor/adapter.go` implements `ToolAdapter`
  - [ ] Correct paths for `.cursor/rules/` directory
  - [ ] Installation + templates + tests

---

#### T-036 — Gemini CLI adapter

- **Effort**: M | **Priority**: P2 | **Deps**: T-005, T-006
- **Description**: Adapter for Google Gemini CLI with config merge strategy.
- **Acceptance criteria**:
  - [ ] `adapters/gemini/adapter.go` implements `ToolAdapter`
  - [ ] Config merge strategy implemented
  - [ ] Installation + templates + tests

---

#### T-037 — Codex adapter

- **Effort**: M | **Priority**: P3 | **Deps**: T-005, T-006
- **Description**: Adapter for OpenAI Codex (TOML config strategy).
- **Acceptance criteria**:
  - [ ] `adapters/codex/adapter.go` implements `ToolAdapter`
  - [ ] TOML config merge implemented
  - [ ] Installation + templates + tests

---

#### T-038 — Contributing guide

- **Effort**: M | **Priority**: P2 | **Deps**: T-037
- **Description**: Document how to add a new adapter. Written for someone who has never read the codebase.
- **Acceptance criteria**:
  - [ ] `CONTRIBUTING.md` with step-by-step adapter development guide
  - [ ] Explains interface contract, template structure, prompt strategies
  - [ ] Testing checklist for new adapters
  - [ ] PR process and review expectations

---

#### T-039 — Adapter template

- **Effort**: S | **Priority**: P2 | **Deps**: T-038
- **Description**: Copy-paste boilerplate for new adapters with TODO comments.
- **Acceptance criteria**:
  - [ ] `adapters/_template/adapter.go`
  - [ ] `adapters/_template/paths.go`
  - [ ] `adapters/_template/installer.go`
  - [ ] `adapters/_template/templates/` with example files
  - [ ] All TODOs clearly mark what must be replaced

---

#### T-040 — Plugin system

- **Effort**: L | **Priority**: P3 | **Deps**: T-005
- **Description**: Allow custom audit phases and agents to be loaded at runtime.
- **Acceptance criteria**:
  - [ ] Plugin interface definition
  - [ ] Plugin loader (file-based discovery)
  - [ ] Example plugin with documentation
  - [ ] Tests

---

#### T-041 — GitHub Action

- **Effort**: M | **Priority**: P3 | **Deps**: T-032
- **Description**: Run Sequoia audits in GitHub Actions CI.
- **Acceptance criteria**:
  - [ ] `action.yml` definition
  - [ ] Runs audit on PR, generates report artifact
  - [ ] Posts comment on PR with Health Score and critical findings
  - [ ] Example workflow in docs

---

#### T-042 — Documentation site

- **Effort**: L | **Priority**: P2 | **Deps**: all prior phases
- **Description**: Static site with full project documentation.
- **Acceptance criteria**:
  - [ ] Getting started guide (5 minutes to first audit)
  - [ ] Architecture overview
  - [ ] CLI reference (all commands and flags)
  - [ ] Adapter development guide
  - [ ] FAQ

---

#### T-043 — Release v0.1.0

- **Effort**: M | **Priority**: P1 | **Deps**: T-034, T-017
- **Description**: First public release supporting Claude Code and OpenCode.
- **Acceptance criteria**:
  - [ ] Binaries built via GoReleaser for all platforms
  - [ ] SHA-256 checksums published
  - [ ] Release notes written
  - [ ] `README.md` updated with install instructions and demo
  - [ ] Homebrew formula published

---

## Critical Path

```
T-001  T-002  T-003
  │      │      │         (spec fixes — run in parallel with T-004..T-007)
  └──────┴──────┘

T-004 → T-005 → T-006 → T-007
                  │
          ┌───────┴────────┐
          │                │
     T-008..T-011     T-019 (CLI)
          │                │
        T-012          T-020
          │                │
        T-013..T-016   T-024..T-031
          │                │
        T-017          T-032 → T-033 → T-034
          │
        T-018
          │
    (feeds T-013, T-035, T-036)

T-034 + T-017 → T-043
```

**Parallel tracks after T-006**:
- Claude adapter (T-008–T-012) and CLI foundation (T-019–T-020) can run in parallel
- Additional adapters (T-035–T-037) can run in parallel after T-018
- TUI screens (T-025–T-031) are independent of each other; integrate in T-032

---

## Effort Summary

| Phase | Tasks | Estimated hours |
|-------|-------|-----------------|
| 1 — Foundation | 7 | ~24h |
| 2 — Claude Code | 5 | ~26h |
| 3 — OpenCode | 6 | ~22h |
| 4 — CLI Installer | 5 | ~28h |
| 5 — TUI Installer | 11 | ~68h |
| 6 — Extensibility | 9 | ~48h+ |
| **Total** | **43** | **~216h** |

Estimated calendar time to v0.1.0 (Phases 1–5 + T-043): **26–30 focused development days**.

---

## Risk Register

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Token budget exceeds limits on large projects | High | High | Chunking in Project Map (T-003). Incremental audit per chunk. |
| Claude Code / OpenCode change install paths | Medium | High | CI tests validate paths against real tool versions |
| Sub-agent delegation underperforms on complex codebases | Medium | High | Early manual test in T-012. Iterate prompts before Phase 3. |
| Cross-platform path issues (Windows vs Unix) | High | Medium | `filepath.Join()` everywhere. CI matrix on 3 OS from T-023. |
| Scope creep — audit features added to TUI | High | High | Hard rule enforced in code review. TUI = install/config/status only. |
| Engram not running at install time | Medium | Low | Installer works without Engram. Config saved to filesystem fallback. |
| Test coverage gaps | Medium | Medium | >80% required for TUI (T-034). Golden files for all templates. |

---

## Definition of Done

A task is **done** when:
1. All acceptance criteria checked off
2. Tests written and passing (unit + integration where applicable)
3. No linter errors (`golangci-lint run`)
4. Code reviewed (self-review minimum; peer review for P1 tasks in Phases 2–5)
5. Memory saved to Engram if non-obvious decisions were made during implementation
