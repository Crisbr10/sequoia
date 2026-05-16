# Delta for tui-status-display

## MODIFIED Requirements

### Requirement: Status Screen

The Status screen MUST show per-tool: cursor indicator, installed marker (✅/❌), tool name, and Sequoia version. The installation path MUST NOT appear in TUI status rows. Available actions: `u` update, `r` reinstall all installed tools, `d` uninstall, `q` quit. The reinstall action SHALL build ProgressTools from all tools where `IsInstalled()` is true, reset progress counters, start the install pipeline, and begin progress polling.

(Previously: Status screen showed installation path as a fifth column in each row.)

#### Scenario: Installed tool shows checkmark, name, and version only

- GIVEN a tool with `IsInstalled()==true`, version `"v0.1.0"`, and path `"/home/user/.claude"`
- WHEN the Status screen renders
- THEN the row MUST contain `✅`, the tool name, and `"v0.1.0"`
- AND the row MUST NOT contain `"/home/user/.claude"` or any installation path

#### Scenario: Not-installed tool shows cross, name, and dash only

- GIVEN a tool with `IsInstalled()==false`
- WHEN the Status screen renders
- THEN the row MUST contain `❌`, the tool name, and `"—"` for version
- AND the row MUST NOT contain a second `"—"` (no placeholder for missing path)

#### Scenario: Uninstall action (unchanged)

- GIVEN a tool is highlighted on the Status screen
- WHEN the user presses `d`
- THEN the model transitions to the Uninstall screen

#### Scenario: Reinstall action (unchanged)

- GIVEN installed tools are present
- WHEN the user presses `r`
- THEN ProgressTools SHALL be built from all installed tools; counters reset; install pipeline started; polling begins; model navigates to InstallProgress

#### Scenario: No tools detected (unchanged)

- GIVEN zero adapters are registered
- WHEN the Status screen renders
- THEN "No adapters registered" message SHALL be displayed

#### Scenario: Golden files updated to new format

- GIVEN golden files `status_all_installed.txt` and `status_mixed.txt`
- WHEN tests run after the path column is removed from `renderStatusRow`
- THEN `status_all_installed.txt` MUST show `✅ {name} {version}` per row (no path)
- AND `status_mixed.txt` MUST show `✅ {name} {version}` for installed tools and `❌ {name} —` for not-installed (no second `—`)
