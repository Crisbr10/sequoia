# Specs: TUI Install Flow

## Domain: tui-install-flow

### Requirement: Configuration Screen

The Configuration screen MUST offer a persistence backend selector (Engram/Files/Both). The language selector (EN/ES) SHALL NOT be rendered in the view — the language field label and option lines MUST be absent from the visual output. However, internal language state cycling, Tab field switching, and pipeline propagation to `adapters.InstallOpts.Language` MUST remain fully functional for future i18n wiring. Engram option SHALL be disabled when MCP is not detected. Tab SHALL switch focus between the hidden language field and the visible Persistence field. Up, Down, Left, and Right arrows SHALL cycle the options within the currently active field. The commented-out language rendering block MUST bear `// TODO(i18n):` markers referencing "translation catalog" or "i18n infrastructure".

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 7 | Engram available | Engram MCP detected | Config renders | "Engram" option selectable; no "Language:", "English", or "Español" visible |
| 8 | Engram unavailable | Engram MCP not detected | Config renders | "Engram" greyed out with "(not detected)" note; no language labels visible |
| 9 | Proceed | User selects options | Enter pressed | Transitions to Install Progress |
| 10 | Tab switches field | Active field 0 (language, hidden) | User presses Tab | Active field changes to 1 (persistence); config unchanged |
| 11 | Up/Down cycles language internally | Active field 0 (language, hidden), config.Language="en" | User presses Down | config.Language changes to "es"; view does NOT show "Language:", "English", or "Español" |
| 12 | Left/Right cycles persistence | Persistence field focused, "Files" selected | User presses Right | Selection cycles to "Both" |
| 13 | View hides language section | Config screen active, config.Language="en" | ConfigurationView() called | Output does NOT contain "Language:", "English", or "Español"; Persistence and footer render normally |
| 14 | Commented code has TODO(i18n) marker | Source file configuration.go | Language rendering block inspected | Block preceded by "TODO(i18n)" comment referencing translation catalog or i18n infrastructure |

### Requirement: Golden Files Reflect Hidden Language Selector

Golden test files for the Configuration screen MUST be regenerated to reflect the hidden language selector. The standard golden file (`configuration_standard.txt`) and engram-unavailable golden file (`configuration_engram_unavailable.txt`) MUST NOT contain "Language:", "English", or "Español" strings. Both files MUST retain the "Persistence:" section with appropriate options and the footer hints (Tab, arrow keys, Enter, Esc).

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 15 | Standard golden lacks language | Standard golden file exists | Content is read | No "Language:", "English", or "Español"; "Persistence:" section with "Engram", "Files", "Both"; footer hints present |
| 16 | Engram-unavailable golden lacks language | Engram-unavailable golden file exists | Content is read | No "Language:", "English", or "Español"; "Persistence:" section with "Engram (not detected)", "Files", "Both"; footer hints present |

### Requirement: Install Progress Screen

The Install Progress screen MUST show per-tool step-by-step progress with a spinner while running. A goroutine SHALL execute `Install()` or `Uninstall()` and send `ProgressMsg` events through a buffered channel. The UI MUST remain responsive during execution. The title SHALL display "Installing" when `OperationMode` is `"install"` and "Uninstalling" when `OperationMode` is `"uninstall"`. The summary line SHALL display "Installing N of M tools" or "Uninstalling N of M tools" accordingly.

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

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 14 | All succeeded — install | 2 of 2 installed, mode=install | Install finishes | Heading "Installation Complete!"; both listed with ✅; `/sequoia-init` hint |
| 15 | Partial success | 1 succeeded, 1 failed | Install finishes | Succeeded shown here; failed deferred to Error screen |
| 16 | Uninstall complete | 2 of 2 uninstalled, mode=uninstall | Uninstall finishes | Heading "Uninstallation Complete!"; tools listed as removed |

### Requirement: Error Screen

The Error screen MUST list failed tools with error messages. `r` MUST trigger retry of failed tools only, rebuilding the pipeline. `q` MUST quit while preserving partial state. The heading SHALL display "Installation Failed" when `OperationMode` is `"install"` and "Uninstallation Failed" when `OperationMode` is `"uninstall"`.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 16 | Retry — install mode | 1 tool failed, mode=install | User presses `r` | Pipeline rebuilt for failed tool; returns to Progress with "Installing" |
| 17 | Quit on error | 1 tool failed | User presses `q` | TUI exits; partial install state preserved |
| 18 | Uninstall error heading | 1 tool uninstall failed, mode=uninstall | Error screen shown | Heading "Uninstallation Failed"; failed tool listed |
| 19 | Retry — uninstall mode | 1 tool uninstall failed, mode=uninstall | User presses `r` | Pipeline rebuilt for failed tool; returns to Progress with "Uninstalling" |
