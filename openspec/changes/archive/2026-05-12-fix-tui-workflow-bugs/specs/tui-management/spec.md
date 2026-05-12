# Delta for tui-management

## ADDED Requirements

### Requirement: Retry Pipeline Rebuild

When the user presses 'r' on the Error screen, the Model SHALL rebuild ProgressTools (reset all steps to pending for failed tools), reset progress counters, start the pipeline matching `OperationMode`, and begin progress polling. Bare screen navigation to InstallProgress without pipeline setup is PROHIBITED.

#### Scenario: Retry from Error — install mode

- GIVEN a tool install failed, Error screen shown with mode=install
- WHEN the user presses 'r'
- THEN ProgressTools rebuilt from failed tools with steps reset to pending
- AND counters reset; install pipeline started; polling begins
- AND model navigates to InstallProgress with "Installing" title

#### Scenario: Retry from Error — uninstall mode

- GIVEN a tool uninstall failed, Error screen shown with mode=uninstall
- WHEN the user presses 'r'
- THEN ProgressTools rebuilt; uninstall pipeline started
- AND model navigates to InstallProgress with "Uninstalling" title

### Requirement: Uninstall Validation Error

When the user presses Enter on the Uninstall screen with no tools that are both selected and installed, the screen SHALL display the error message "Select at least one installed tool to continue". The model SHALL remain on the Uninstall screen.

#### Scenario: No valid tools selected

- GIVEN Uninstall screen with tools but none selected and installed
- WHEN the user presses Enter
- THEN `ErrorMsg` set to "Select at least one installed tool to continue"
- AND screen remains on Uninstall; error rendered in the view

#### Scenario: Valid tools selected

- GIVEN Uninstall screen with at least one tool selected and installed
- WHEN the user presses Enter
- THEN model transitions to confirmation mode; no error displayed

### Requirement: Uninstall Footer Hints

The Uninstall screen footer SHALL display "Esc back" alongside existing navigation hints.

#### Scenario: Footer display

- GIVEN the Uninstall screen is active
- WHEN the view renders the footer
- THEN "Esc back" hint is displayed alongside Space, Enter, and q hints

## MODIFIED Requirements

### Requirement: Status Screen

The Status screen MUST show per-tool: name, installed (✅/❌), Sequoia version, installation path. Available actions: `u` update, `r` reinstall all installed tools, `d` uninstall, `q` quit. The reinstall action SHALL build ProgressTools from all tools where `IsInstalled()` is true, reset progress counters, start the install pipeline, and begin progress polling.
(Previously: actions applied to highlighted tool only; reinstall navigated to InstallProgress without building pipeline state.)

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 1 | All installed | 2 tools, both `IsInstalled()==true` | Status renders | Both show ✅ with version and path |
| 2 | Mixed state | 1 installed, 1 not | Status renders | Installed: ✅; not-installed: ❌ with "—" |
| 3 | Uninstall action | Tool highlighted | User presses `d` | Transitions to Uninstall screen |
| 4 | Reinstall action | Installed tools present | User presses `r` | ProgressTools built from all installed tools; counters reset; install pipeline started; polling begins; navigates to InstallProgress |
| 5 | No tools detected | 0 adapters registered | Status renders | "No adapters registered" message |

### Requirement: Uninstall Screen

The Uninstall screen MUST show a checkbox list of installed tools, a confirmation step before execution, and reuse the Install Progress screen. `Enter` on confirmation triggers the uninstall pipeline; `n` cancels. During confirmation mode, Esc SHALL exit confirmation (same as 'n'). Pressing Esc or Left from Uninstall selection SHALL navigate back to the source screen (Status if arrived via 'd', Welcome if arrived from Welcome menu). Hardcoded return to Welcome is PROHIBITED.
(Previously: confirmation only handled y/n; Esc from Uninstall always returned to Welcome regardless of origin.)

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 6 | No tools installed | 0 tools `IsInstalled()==true` | Uninstall renders | "Nothing to uninstall"; only `q` available |
| 7 | Confirmation prompt | 1 tool selected and installed | Enter pressed | "Remove Sequoia from {name}? [y/N]" shown |
| 8 | Confirm proceed | Confirmation shown | User presses `y` | OperationMode set to "uninstall"; uninstall pipeline built; Progress shown |
| 9 | Confirm cancel with n | Confirmation shown | User presses `n` | Returns to Uninstall selection |
| 10 | Confirm cancel with Esc | Confirmation shown | User presses Esc | Returns to Uninstall selection (same as `n`) |
| 11 | Uninstall complete | Goroutine finishes all steps | Done | Transitions to Complete showing removed tools |
| 12 | Back to Status | Arrived via 'd' on Status screen | User presses Esc or Left | Navigates to Status screen |
| 13 | Back to Welcome | Arrived via Welcome menu | User presses Esc or Left | Navigates to Welcome screen |
| 14 | Validation error | 0 tools selected and installed | Enter pressed | Error "Select at least one installed tool to continue"; stays on Uninstall |
