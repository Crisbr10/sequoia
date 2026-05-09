# tui-core Specification

## Purpose
Bubbletea root model with screen state machine, update dispatch, view delegation, and shared domain types.

## Requirements

### Requirement: Screen State Machine
The TUI root model MUST track current screen via a `Screen` enum and delegate `Update()` dispatch and `View()` rendering to the active screen. Unknown screen keys SHALL be no-ops.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 1 | Initial screen | App launches | Model initialized | `Screen` defaults to `ScreenWelcome` |
| 2 | Screen transition | Model at `ScreenWelcome` | User presses Enter | `Screen` updates to `ScreenToolSelection` |
| 3 | Invalid transition | Model at `ScreenComplete` | Unknown screen key received | Model stays at current `Screen` |
| 4 | Quit from any screen | Model at any screen | User presses `q` or `Ctrl+C` | `tea.Quit` message returned |

### Requirement: Domain Types
The system MUST provide shared Go types: `InstalledTool` (ID, Name, Path, Installed, Version), `InstallResult` (ToolID, Success, Error, Steps), `Screen` (enum), and `ProgressMsg` (ToolID, Step, Status, Error). `ScanTools()` SHALL populate `InstalledTool` from adapter status.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 5 | InstalledTool populated | Adapter returns `AdapterStatus` | `ScanTools()` called | Structs built with all fields |
| 6 | ProgressMsg mid-install | Install goroutine running | Step completes | `ProgressMsg{Status:"done"}` sent to channel |
| 7 | ProgressMsg on error | Install goroutine running | Step fails | `ProgressMsg{Status:"error", Error: err}` sent |
