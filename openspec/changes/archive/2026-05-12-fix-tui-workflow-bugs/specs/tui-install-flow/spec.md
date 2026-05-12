# Delta for tui-install-flow

## MODIFIED Requirements

### Requirement: Configuration Screen

The Configuration screen MUST offer language (EN/ES) and persistence backend (Engram/Files/Both) selectors. Engram option SHALL be disabled when MCP is not detected. Tab SHALL switch focus between Language and Persistence fields. Up, Down, Left, and Right arrows SHALL cycle the options within the currently active field.
(Previously: Tab and Up/Down both toggled between fields; only Left/Right cycled options.)

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 7 | Engram available | Engram MCP detected | Config renders | "Engram" option selectable |
| 8 | Engram unavailable | Engram MCP not detected | Config renders | "Engram" greyed out with "(not detected)" note |
| 9 | Proceed | User selects options | Enter pressed | Transitions to Install Progress |
| 10 | Tab switches field | Language field focused | User presses Tab | Persistence field becomes focused |
| 11 | Up/Down cycles options | Language field focused, "English" selected | User presses Down | Selection cycles to "Español" |
| 12 | Left/Right cycles options | Persistence field focused, "Files" selected | User presses Right | Selection cycles to "Both" |

### Requirement: Install Progress Screen

The Install Progress screen MUST show per-tool step-by-step progress with a spinner while running. A goroutine SHALL execute `Install()` or `Uninstall()` and send `ProgressMsg` events through a buffered channel. The UI MUST remain responsive during execution. The title SHALL display "Installing" when `OperationMode` is `"install"` and "Uninstalling" when `OperationMode` is `"uninstall"`. The summary line SHALL display "Installing N of M tools" or "Uninstalling N of M tools" accordingly.
(Previously: title and summary were hardcoded to "Installing" regardless of operation.)

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 10 | Happy path | 1 tool selected, mode=install | Install runs | Steps: `[ ] Skills` → spinner → `[✓] Skills` → `[✓] Commands` → `[✓] System Prompt` |
| 11 | Multi-tool parallel | 2 tools selected, mode=install | Install runs | Independent progress blocks for each tool |
| 12 | Step failure | Install goroutine errors | Step fails | `[✗]` with error message; remaining steps skipped; Error screen |
| 13 | All success | All tools/steps complete | Last step finishes | Transitions to Complete screen |
| 14 | Uninstall progress label | mode=uninstall, 1 tool | Uninstall runs | Title "Uninstalling"; summary "Uninstalling 1 of 1 tools" |
| 15 | Install progress label | mode=install, 2 tools, 1 complete | During install | Title "Installing"; summary "Installing 1 of 2 tools" |

### Requirement: Complete Screen

The Complete screen MUST list succeeded tools and show the first command to try (`/sequoia-init`). `q` exits the TUI. The heading SHALL display "Installation Complete!" when `OperationMode` is `"install"` and "Uninstallation Complete!" when `OperationMode` is `"uninstall"`.
(Previously: heading was hardcoded to "Installation Complete!")

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 14 | All succeeded — install | 2 of 2 installed, mode=install | Install finishes | Heading "Installation Complete!"; both listed with ✅; `/sequoia-init` hint |
| 15 | Partial success | 1 succeeded, 1 failed | Install finishes | Succeeded shown here; failed deferred to Error screen |
| 16 | Uninstall complete | 2 of 2 uninstalled, mode=uninstall | Uninstall finishes | Heading "Uninstallation Complete!"; tools listed as removed |

### Requirement: Error Screen

The Error screen MUST list failed tools with error messages. `r` MUST trigger retry of failed tools only, rebuilding the pipeline. `q` MUST quit while preserving partial state. The heading SHALL display "Installation Failed" when `OperationMode` is `"install"` and "Uninstallation Failed" when `OperationMode` is `"uninstall"`.
(Previously: heading was hardcoded to "Installation Failed")

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 16 | Retry — install mode | 1 tool failed, mode=install | User presses `r` | Pipeline rebuilt for failed tool; returns to Progress with "Installing" |
| 17 | Quit on error | 1 tool failed | User presses `q` | TUI exits; partial install state preserved |
| 18 | Uninstall error heading | 1 tool uninstall failed, mode=uninstall | Error screen shown | Heading "Uninstallation Failed"; failed tool listed |
| 19 | Retry — uninstall mode | 1 tool uninstall failed, mode=uninstall | User presses `r` | Pipeline rebuilt for failed tool; returns to Progress with "Uninstalling" |

## REMOVED Requirements

### Requirement: Up/Down Field Toggle in Configuration

(Reason: Replaced by Tab-only field switching per REQ-FLOW-CONFIG-NAV. Up/Down now cycle options within the active field, matching Left/Right behavior.)

### Requirement: Per-Screen 'q' Key Handlers

(Reason: The global 'q' handler in the root `Update()` preempts screen-specific dispatch. Dead handlers in ToolSelectionUpdate, ConfigurationUpdate, and InstallProgressUpdate are removed. Global quit behavior — tui-core scenario 4 — is unchanged.)
