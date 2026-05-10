# Architecture

This document describes Sequoia's internal design — how the installer, TUI,
CLI, and adapter system fit together.

## High-Level Overview

```
                           sequoia CLI
                          ┌─────────┐
                          │  main() │
                          └────┬────┘
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                ▼
         ┌─────────┐    ┌──────────┐    ┌─────────────┐
         │  Cobra  │    │Bubbletea │    │   Adapter   │
         │  CLI    │    │   TUI    │    │  Registry   │
         └────┬────┘    └────┬─────┘    └──────┬──────┘
              │              │                  │
              └──────────────┼──────────────────┘
                             │
                    ┌────────┴────────┐
                    │  ToolAdapter    │
                    │  (interface)    │
                    └────────┬────────┘
                             │
       ┌─────────┬───────────┼───────────┬──────────┐
       ▼         ▼           ▼           ▼          ▼
    Claude    OpenCode    Cursor      Gemini     Codex
```

## ToolAdapter Pattern

Every AI tool integration implements the `ToolAdapter` interface defined in
`adapters/interface.go`. This is the central abstraction — the CLI, TUI, and
installer never know about specific tools. They only talk to the interface.

```go
type ToolAdapter interface {
    ID()              string         // machine ID: "claude-code"
    Name()            string         // display name: "Claude Code"
    Detect()          bool           // is the tool installed?
    IsInstalled()     bool           // is Sequoia present?
    Install()         error          // install Sequoia files
    Uninstall()       error          // remove Sequoia files
    Status()          AdapterStatus  // current state
    SkillsPath()      string         // where skills live
    CommandsPath()    string         // where commands live
    SystemPromptPath() string        // system prompt file
    PromptStrategy()  PromptStrategy // injection strategy
}
```

### Registry (Self-Registration)

Adapters register themselves at init time, following the `database/sql` pattern:

```go
// adapters/claude/adapter.go
func init() {
    adapters.DefaultRegistry.Register(&Adapter{})
}
```

Then in `cmd/sequoia/main.go`:

```go
import _ "sequoia-ai/adapters/claude" // triggers init()
```

The CLI and TUI call `adapters.DefaultRegistry.All()` to discover every
registered tool. Adding a new adapter is a matter of writing one package
and adding one import — nothing else changes.

### Factory

```go
adapter, err := adapters.NewAdapter("claude-code")
```

`NewAdapter()` returns a fresh instance from the registry. Used by the
headless CLI when the user specifies `--tool=<id>`.

## Prompt Strategies

Sequoia supports four injection strategies, each suited to different tool
conventions:

### StrategyMarkdownSections

**Used by**: Claude Code

Injects a delimited section into an existing Markdown file using HTML comment
markers:

```markdown
<!-- sequoia:start -->
... Sequoia content ...
<!-- sequoia:end -->
```

All content outside the markers is preserved. Running twice produces identical
output (idempotent).

### StrategyFileReplace

**Used by**: OpenCode, Cursor IDE

Replaces the entire target file with Sequoia content. Before overwriting, a
backup (`.sequoia-backup`) is created if the file contains non-Sequoia content.
On uninstall, the backup is restored.

### StrategyConfigMerge

**Used by**: Gemini CLI

Like `StrategyMarkdownSections` but for tools where the config file format is
not standard Markdown. Uses the same marker-based delimitation.

### StrategyTOMLMerge

**Used by**: OpenAI Codex

Merges a `[sequoia]` TOML table into an existing TOML config file. All
pre-existing keys and sections are preserved. Only the `[sequoia]` table
is managed.

## Installer Pipeline

The `adapters/common/installer.go` framework provides a four-phase lifecycle
shared by all adapters:

```
Prepare → Apply → Verify → Rollback
```

| Phase | What Happens |
|-------|-------------|
| **Prepare** | Validates paths, checks write permissions, creates backups |
| **Apply** | Copies files from staging to target directories |
| **Verify** | Confirms all expected files exist and are readable |
| **Rollback** | Restores backups and cleans up on any failure |

Each phase is atomic: if Apply fails, Rollback restores everything Prepare
backed up. If Verify fails, Rollback clears the partial installation.

The pipeline is idempotent — running `Install()` twice produces identical state.

## CLI Commands (Cobra)

The CLI uses [Cobra](https://github.com/spf13/cobra) for command structure:

```
sequoia
├── install   [--tool=<id>] [--no-tui]
├── status
├── uninstall [--tool=<id>] [--all] [--yes]
└── version
```

See the [CLI Reference](cli-reference.md) for detailed usage.

## TUI (Bubbletea)

The interactive TUI uses [Bubbletea](https://github.com/charmbracelet/bubbletea)
with a screen-based architecture:

### Screen State Machine

```
Welcome ──→ ToolSelection ──→ Configuration ──→ InstallProgress ──┬──→ Complete
                                                                    │
                                                                    └──→ Error ──→ (retry) ──→ InstallProgress
                                                                         │
                                                                         └──→ (quit)

Status ←── (from any screen via 's' or 'status' command)
Uninstall ←── (from Status via 'd')
```

### Architecture

| Component | Role |
|-----------|------|
| `internal/app/model.go` | Root Bubbletea `Model` — holds screen state, tool list, config, progress channel |
| `internal/app/update.go` | `Update()` — message dispatch by screen type |
| `internal/app/view.go` | `View()` — renders the current screen |
| `internal/tui/router.go` | `NextScreen()` — state machine transitions |
| `internal/tui/screens/` | Individual screen implementations (welcome, tool-selection, etc.) |
| `internal/tui/styles/` | Lipgloss theme (colors, typography, logo) |
| `internal/model/types.go` | Domain types (`Screen`, `ToolState`, `InstallResult`, `ProgressMsg`, `TUIConfig`) |
| `internal/pipeline/runner.go` | Goroutine-based install/uninstall/status pipeline with buffered channel |

### Pipeline Integration

The TUI launches installation in a goroutine via `internal/pipeline/runner.go`.
Progress messages flow through a buffered channel (capacity 64) that the
Progress screen consumes in its `Update()` function. This keeps the UI
responsive during long-running operations.

**Hard rule**: The TUI handles install / config / status ONLY. Audit features
never appear in the TUI. All audit logic lives in slash commands inside the
AI tools.

## Cross-Platform Design

| Principle | Implementation |
|-----------|---------------|
| Path construction | `filepath.Join()` everywhere — never string concatenation |
| Symlink handling | `filepath.EvalSymlinks()` before path operations |
| OS detection | `runtime.GOOS` for Go code; `uname -s` / `$env:OS` for shell scripts |
| Test isolation | `t.TempDir()` for all tests — never mutate real config directories |
| Golden files | Use `filepath.ToSlash()` for cross-platform golden file comparison |

## Project Structure

```
sequoia/
├── cmd/sequoia/              # Cobra CLI entrypoint
│   └── main.go               # Root command + subcommands
├── adapters/                 # ToolAdapter interface + implementations
│   ├── interface.go          # Contract every adapter satisfies
│   ├── registry.go           # Plugin registry (database/sql pattern)
│   ├── factory.go            # NewAdapter(id) constructor
│   ├── common/               # Shared installer (Prepare → Apply → Verify → Rollback)
│   ├── claude/               # Claude Code adapter
│   ├── opencode/             # OpenCode adapter
│   ├── cursor/               # Cursor IDE adapter
│   ├── gemini/               # Gemini CLI adapter
│   ├── codex/                # OpenAI Codex adapter
│   └── _template/            # Adapter scaffolding reference
├── internal/                 # Private packages
│   ├── app/                  # Bubbletea model, update, view
│   ├── tui/                  # Screen implementations
│   │   ├── screens/          # Welcome, ToolSelection, Config, etc.
│   │   ├── styles/           # Lipgloss theme
│   │   └── router.go         # Screen state machine
│   ├── model/                # Domain types
│   └── pipeline/             # Installation pipeline orchestration
├── scripts/                  # One-line installers (curl | bash, irm | iex)
├── docs/                     # Documentation (you are here)
├── .goreleaser.yaml          # GoReleaser build configuration
└── .golangci.yaml            # Linter configuration
```
