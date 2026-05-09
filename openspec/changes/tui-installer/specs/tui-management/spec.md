# tui-management Specification

## Purpose
Post-install management: Status screen (install state + actions) and Uninstall screen (checkbox + confirmation + progress reuse).

## Requirements

### Requirement: Status Screen
The Status screen MUST show per-tool: name, installed (✅/❌), Sequoia version, installation path. Available actions: `u` update, `r` reinstall, `d` uninstall, `q` quit. Actions SHALL apply to the currently highlighted tool.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 1 | All installed | 2 tools, both `IsInstalled()==true` | Status renders | Both show ✅ with version and path |
| 2 | Mixed state | 1 installed, 1 not | Status renders | Installed: ✅; not-installed: ❌ with "—" |
| 3 | Uninstall action | Tool highlighted | User presses `d` | Transitions to Uninstall for that tool |
| 4 | Reinstall action | Tool highlighted | User presses `r` | Transitions to Install Progress for that tool |
| 5 | No tools detected | 0 adapters registered | Status renders | "No adapters registered" message |

### Requirement: Uninstall Screen
The Uninstall screen MUST show a checkbox list of installed tools, a confirmation step before execution, and reuse the Install Progress screen with uninstall messages. `Enter` on confirmation triggers the uninstall goroutine; `n` cancels.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 6 | No tools installed | 0 tools `IsInstalled()==true` | Uninstall renders | "Nothing to uninstall"; only `q` available |
| 7 | Confirmation prompt | 1 tool selected | Enter pressed | "Remove Sequoia from {name}? [y/N]" shown |
| 8 | Confirm proceed | Confirmation shown | User presses `y` | Uninstall goroutine launched; Progress with "Removing…" |
| 9 | Confirm cancel | Confirmation shown | User presses `n` | Returns to Uninstall selection |
| 10 | Uninstall complete | Goroutine finishes all steps | Done | Transitions to Complete showing removed tools |
