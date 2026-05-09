# Design: Phase 5 — TUI Installer

## Technical Approach

Bubbletea root Model lives in `internal/app/`. Screens are pure functions in `internal/tui/screens/` dispatching via a `Screen` enum switch in `Update()`/`View()`. Installation runs in a `tea.Cmd` goroutine sending `ProgressMsg` over a buffered channel (64) — never blocking the UI loop. Adapter calls go through `adapters.DefaultRegistry` (existing), no new abstractions.

## Architecture Decisions

| Decision | Options | Choice | Rationale |
|----------|---------|--------|-----------|
| Screen dispatch pattern | Interface-per-screen vs enum switch | Enum switch in `Update()`/`View()` | Simpler, fewer allocations, no screen knows about others — each screen is a standalone `(model) → string` function |
| Progress communication | Callbacks vs channel vs shared state | Buffered chan (64) with `tea.Cmd` polling | Channels are idiomatic Go for goroutine comms; `tea.Cmd` integrates naturally with Bubbletea event loop |
| Config persistence | Screen-only state vs write to disk | Screen-only state, passed to pipeline | TUI session config is ephemeral; actual persistence choice is per-adapter work deferred to Phase 6 |
| Styling library | lipgloss only vs bubbles components | lipgloss for base styling + bubbles spinner/progressbar | lipgloss for consistent theming; bubbles for well-tested interactive widgets (checkbox list, spinner) |

## Data Flow

```
cmd/sequoia/main.go
  │ tea.NewProgram(app.NewModel(toolID), tea.WithAltScreen())
  ▼
internal/app/model.go      ──►  internal/pipeline/runner.go
  │ Model{Screen, Tools, ...}      goroutine: for each tool:
  │ Update(msg) ── per-Screen        adapter.Install()
  │ View(model) ── screen.Render()   send ProgressMsg → chan (64)
  ▼                                  defer close(chan)
internal/tui/screens/
  welcome.go tool-selection.go
  configuration.go install-progress.go
  complete.go error.go status.go uninstall.go
  │
  ▼
adapters.DefaultRegistry.All() → ToolAdapter.Install()/Uninstall()
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/model/types.go` | Create | `Screen` enum, `InstalledTool`, `InstallResult`, `ProgressMsg`, `TUIConfig`, `ToolState` |
| `internal/app/model.go` | Create | Root Bubbletea `Model` struct, `NewModel(toolID string)`, `Init()` |
| `internal/app/update.go` | Create | `Update(msg tea.Msg)` — screen dispatch switch, key handling |
| `internal/app/view.go` | Create | `View(m Model) string` — screen dispatch to renderers |
| `internal/tui/router.go` | Create | Screen transition map `map[Screen][]Screen`, `NextScreen(current, action)` |
| `internal/tui/screens/welcome.go` | Create | Branding header, detected tools table, "Press Enter" footer |
| `internal/tui/screens/tool-selection.go` | Create | Checkbox list via `bubbles/list`, Space toggle, ≥1 validation |
| `internal/tui/screens/configuration.go` | Create | Language picker (EN/ES), persistence backend (Engram/Files/Both) |
| `internal/tui/screens/install-progress.go` | Create | Per-tool spinner + step status, `tea.Batch` for parallel installs |
| `internal/tui/screens/complete.go` | Create | Success summary with tool names, paths, version, next steps |
| `internal/tui/screens/error.go` | Create | Failed tools table, retry/quit prompt |
| `internal/tui/screens/status.go` | Create | All-tools install status table, actions (uninstall/quit) |
| `internal/tui/screens/uninstall.go` | Create | Checkbox multi-select + confirmation, reuse InstallProgress |
| `internal/pipeline/runner.go` | Create | `RunInstall`/`RunUninstall` — goroutine per tool, context cancel, chan(64) |
| `internal/app/styles.go` | Create | lipgloss theme: colors, borders, padding constants |
| `cmd/sequoia/main.go` | Modify | Replace `runTUI()` stub: `tea.NewProgram(app.NewModel(toolID)).Run()` |
| `go.mod` | Modify | Add `bubbletea v1.x`, `lipgloss`, `bubbles`; `teatest` for dev |
| `internal/tui/screens/*_test.go` | Create | Golden file + interaction tests per screen |
| `internal/app/model_test.go` | Create | Model initialization, screen transitions, init message |
| `internal/pipeline/runner_test.go` | Create | Happy path, error, cancel, channel closure |

## Interfaces / Contracts

```go
// internal/model/types.go

type Screen int
const (
    ScreenWelcome Screen = iota
    ScreenToolSelection
    ScreenConfiguration
    ScreenInstallProgress
    ScreenComplete
    ScreenError
    ScreenStatus
    ScreenUninstall
)

type ToolState struct {
    Adapter    adapters.ToolAdapter
    Selected   bool
    Result     *InstallResult
}

type InstallResult struct {
    ToolID  string
    Success bool
    Error   string
    Steps   []StepResult
}

type ProgressMsg struct {
    ToolID string
    Step   string // "prepare" | "apply" | "verify"
    Done   bool
    Error  string
}

type TUIConfig struct {
    Language    string // "en" | "es"
    Persistence string // "engram" | "files" | "both"
}
```

```go
// internal/app/model.go

type Model struct {
    Screen   model.Screen
    Tools    []model.ToolState
    Config   model.TUIConfig
    Width    int
    Height   int
    Progress chan model.ProgressMsg // buffered 64
    Quitting bool
}
```

## Screen State Machine

```
Welcome ──Enter──► ToolSelection ──Enter──► Configuration ──Enter──► InstallProgress
                                                                           │
                                                    ┌──────────────────────┤
                                                    ▼                      ▼
                                               Complete (all ok)      Error (any fail)
                                                    │                      │
                                                    ▼                      ▼
                                               Status ◄─────────── (retry → ToolSelection)
                                                    │
                                                    ▼
                                               Uninstall ──► InstallProgress (reuse)
```

Global: `q` | `ctrl+c` → quit from any screen.

## Keybinding Map

| Key | Screen | Action |
|-----|--------|--------|
| `q` / `ctrl+c` | All | Quit |
| `Enter` | Welcome → ToolSelection → Configuration | Advance |
| `Space` | ToolSelection, Uninstall | Toggle checkbox |
| `↑` / `↓` / `j` / `k` | ToolSelection, Configuration, Error, Uninstall | Navigate |
| `r` | Error | Retry failed tools |
| `tab` | Configuration | Switch field (language ↔ persistence) |
| `enter` | Status | Select action (uninstall / quit) |

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit — screens | Render output per screen | Golden files at `testdata/golden/{screen}_{w}x{h}.golden` (25×80, 40×120); `filepath.ToSlash()` normalizes paths |
| Unit — pipeline | Install/uninstall goroutine | Mock adapter via `testAdapter` implementing `ToolAdapter`; verify channel messages, error surfacing, context cancel |
| Integration — TUI | Full flow | `teatest.NewTestModel(app.NewModel(""))` with interaction script; verify screen sequence and final state |
| Integration — CLI | `runTUI()` launch | Override `isTerminalFn`; verify `tea.NewProgram` called with correct model and options |

Golden file test pattern:
```go
func TestWelcomeScreen(t *testing.T) {
    tc := []struct { name string; width, height int }{
        {"25x80", 80, 25},
        {"40x120", 120, 40},
    }
    for _, c := range tc {
        t.Run(c.name, func(t *testing.T) {
            m := app.NewModel("")
            m.Width, m.Height = c.width, c.height
            got := app.View(m)
            golden.Assert(t, got, fmt.Sprintf("welcome_%s.golden", c.name))
        })
    }
}
```

## Dependency Graph

New Go modules:
- `github.com/charmbracelet/bubbletea` v1.x — TUI framework
- `github.com/charmbracelet/lipgloss` — terminal styling
- `github.com/charmbracelet/bubbles` — spinner, progressbar, textinput
- `github.com/charmbracelet/teatest` (dev) — TUI test harness

No indirect dependencies removed; no existing packages deleted.

## Open Questions

- [ ] Should config (language/persistence) persist across sessions in a `.sequoia-tui.yaml` file, or remain ephemeral per-run? Design assumes ephemeral; Phase 6 can add persistence.
- [ ] Progress pipeline: sequential (one tool at a time) vs parallel (all tools concurrently)? Design uses sequential with per-step progress — simpler error recovery, clearer UX.
