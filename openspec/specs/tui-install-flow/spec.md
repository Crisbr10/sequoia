# Specs: TUI Install Flow

## Domain: tui-install-flow

### Requirement: Configuration Screen

The Configuration screen MUST offer a persistence backend selector (Engram/Files/Both). The language selector (EN/ES) SHALL be rendered when `i18n.Initialized()` is true — the language field label ("Language:") and option lines ("English", "Español") MUST be visible in the view output. When i18n is not initialized, the language selector SHALL NOT be rendered and only the Persistence field SHALL be visible. Engram option SHALL be disabled when MCP is not detected. Tab SHALL switch focus between the language field (field 0) and the Persistence field (field 1). Up and Down arrows SHALL cycle the language options; Left and Right arrows SHALL cycle the persistence options. All `TODO(i18n)` comment markers SHALL be removed from the source file. All field labels and option text SHALL be sourced from the i18n catalog via `i18n.T()`.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 7 | Engram available | Engram MCP detected, i18n initialized | Config renders | "Engram" option selectable; "Language:", "English", "Español" visible |
| 8 | Engram unavailable | Engram MCP not detected, i18n initialized | Config renders | "Engram" greyed out with "(not detected)" note; language labels visible |
| 9 | Proceed | User selects options | Enter pressed | Transitions to Install Progress |
| 10 | Tab switches both fields | Active field 0 (language), i18n initialized | User presses Tab | Active field changes to 1 (persistence); Tab again toggles back to 0 |
| 11 | Up/Down cycles language | Active field 0 (language), config.Language="en" | User presses Down | config.Language changes to "es"; view shows "► Español" |
| 12 | Left/Right cycles persistence | Persistence field focused, "Files" selected | User presses Right | Selection cycles to "Both" |
| 13 | Language selector visible when initialized | Config screen active, i18n initialized, config.Language="en" | ConfigurationView() called | Output contains "Language:", "► English", "Español"; Persistence and footer render normally |
| 14 | Language selector hidden when not initialized | Config screen active, i18n NOT initialized | ConfigurationView() called | Output lacks "Language:"; only persistence field visible |

### Requirement: Golden Files Include Language Selector

Golden test files for the Configuration screen MUST be regenerated to include the language selector. The standard golden file (`configuration_standard.txt`) and engram-unavailable golden file (`configuration_engram_unavailable.txt`) MUST contain "Language:", "English", and "Español" strings. Both files MUST retain the "Persistence:" section with appropriate options and the footer hints (Tab, arrow keys, Enter, Esc).

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 15 | Standard golden includes language | Standard golden file exists | Content is read | Contains "Language:", "English", "Español"; "Persistence:" section with "Engram", "Files", "Both"; footer hints present |
| 16 | Engram-unavailable golden includes language | Engram-unavailable golden file exists | Content is read | Contains "Language:", "English", "Español"; "Persistence:" section with "Engram (not detected)", "Files", "Both"; footer hints present |

### Requirement: Configuration Tests Re-Enabled

The previously skipped test `TestConfigurationView_ShowsLanguageOptions` in `internal/tui/screens/configuration_test.go` SHALL be unskipped (no `t.Skip()` call). The test SHALL assert "English" and "Español" are visible when i18n is initialized. The `TestConfigurationView_RendersLanguageAndPersistence` test in `internal/app/model_test.go` SHALL also be unskipped.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 17 | Unskipped test passes | i18n initialized, standard TUI config | Test runs | English and Español assertions pass; no Skip() call present |

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

The Complete screen MUST list succeeded tools. `q` exits the TUI. The heading SHALL display "Installation Complete!" when `OperationMode` is `"install"` and "Uninstallation Complete!" when `OperationMode` is `"uninstall"`. The summary line SHALL use `complete.installed_items` for install, `complete.uninstalled_items` for clean uninstall (zero warnings), and `complete.warnings_note` when uninstall completed with warnings.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 14 | Install complete — summary | mode=install, all succeeded | Install finishes | Heading "Installation Complete!"; summary "Installed: Skills, Commands, System Prompt" |
| 15 | Partial success | 1 succeeded, 1 failed, mode=install | Install finishes | Succeeded listed with correct summary; failed deferred to Error screen |
| 16 | Clean uninstall — summary | mode=uninstall, all succeeded, warnedCount=0 | Uninstall finishes | Heading "Uninstallation Complete!"; summary "Uninstalled: Skills, Commands, System Prompt" |
| 17 | Uninstall with warnings | mode=uninstall, warnedCount>0 | Uninstall finishes | Heading "Uninstallation Complete! (N with warnings)"; warnings summary displayed — NOT item list |
| 18 | i18n key `complete.uninstalled_items` exists | i18n catalog loaded, lang="en" or "es" | T("complete.uninstalled_items", lang) | Returns localized "Uninstalled: Skills, Commands, System Prompt" (en) / "Desinstalado: Skills, Commands, System Prompt" (es) |

### Requirement: Error Screen

The Error screen MUST list failed tools with error messages. `r` MUST trigger retry of failed tools only, rebuilding the pipeline. `q` MUST quit while preserving partial state. The heading SHALL display "Installation Failed" when `OperationMode` is `"install"` and "Uninstallation Failed" when `OperationMode` is `"uninstall"`.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 16 | Retry — install mode | 1 tool failed, mode=install | User presses `r` | Pipeline rebuilt for failed tool; returns to Progress with "Installing" |
| 17 | Quit on error | 1 tool failed | User presses `q` | TUI exits; partial install state preserved |
| 18 | Uninstall error heading | 1 tool uninstall failed, mode=uninstall | Error screen shown | Heading "Uninstallation Failed"; failed tool listed |
| 19 | Retry — uninstall mode | 1 tool uninstall failed, mode=uninstall | User presses `r` | Pipeline rebuilt for failed tool; returns to Progress with "Uninstalling" |
