# Experience Tasks — sequoia-ai (v0.1.0)

**Score**: 0/100 (Critical) | **Findings**: 15 (0 CRITICAL, 4 HIGH, 6 MEDIUM, 4 LOW, 1 INFO) | **Audit ID**: audit-20260524-sequoia-v2

---

## Area Summary

Experience is a new phase in this audit (previously N/A for CLI-only tool). The TUI surface area has grown significantly, adding 6+ screens (ToolSelection, Configuration, InstallProgress, Status, Complete, Error, Uninstall), but UX maturity hasn't kept pace. InstallProgress lacks basic feedback (no spinner, no cancel button, no failure count). Navigation inconsistencies create a confusing user experience—keys work differently on different screens. Headless mode (CLI) is treated as second-class. This phase requires the most attention alongside Architecture (CORR-001) since the screen state machine refactor will enable many of these fixes.

---

## 🔴 HIGH Findings

### EXP-001 (P5-001): Add animated spinner to InstallProgress

**Priority**: 🔴 Blocking

**Problem**: The InstallProgress screen displays "Installing..." as static text with no animation. During long installations (Claude Code can take 30+ seconds to download), users cannot tell if the app is working or frozen. This is the most user-visible screen during the primary workflow.

**Fix**: Add a Bubbletea spinner component (or a simple ASCII animation cycling through `|/-\` characters). The spinner should tick on a `spinner.TickMsg` at ~100ms intervals. When installation completes, transition to the final result screen with the spinner replaced by a checkmark/cross.

**Acceptance Criteria**:
- [ ] InstallProgress shows animated spinner cycling at ~100ms intervals
- [ ] Spinner stops when installation completes (success or failure)
- [ ] Spinner visible during both install and uninstall flows
- [ ] No performance regression: spinner tick doesn't trigger full re-render (dirty flag)
- [ ] Golden test updated to capture spinner states

**Effort**: small (1–2h) | **Risk**: low | **Blocks**: CORR-001 (should be done alongside state machine refactor)

---

### EXP-002 (P5-002): Add cancel/abort mechanism to InstallProgress

**Priority**: 🔴 Blocking

**Problem**: There is no way to cancel an installation once started. If a user accidentally starts installing a large tool (Claude Code ~500MB) or the download hangs, they must kill the entire process. The install pipeline already supports context cancellation—it's just not wired to a UI element.

**Evidence**:
- Install pipeline accepts `context.Context` with cancellation
- No TUI keybinding triggers `context.CancelFunc`
- User must SIGTERM/SIGKILL the process to abort

**Fix**:
1. Add `cancel context.CancelFunc` to the InstallProgress model
2. Bind `ctrl+c` or `esc` key to call `cancel()` during installation
3. Show "Canceling..." state after cancel is triggered
4. Pipeline detects context cancellation and rolls back partial installation
5. Transition to an "Installation Canceled" result screen

**Acceptance Criteria**:
- [ ] `ctrl+c` or `esc` triggers context cancellation during install
- [ ] "Canceling..." message shown while pipeline shuts down
- [ ] Partial installation is rolled back (no leftover files)
- [ ] "Installation Canceled" screen shown with option to return to main menu
- [ ] Test verifies: cancel mid-install, verify no files left behind
- [ ] Cancel keybinding documented and discoverable (shown on screen)

**Effort**: small (2–4h) | **Risk**: low | **Blocks**: CORR-001

---

### EXP-003 (P5-006): Add return-to-main-menu from Complete screen

**Priority**: 🔴 Blocking (CORR-001)

**Problem**: The Complete screen is a near-dead-end. After an installation finishes, the user sees a success/result screen but has no direct way to return to the main menu. They must use `esc` or `q` to navigate back through multiple screens or restart the app. This is the final screen in the install workflow—it should offer a clear "Back to Main Menu" action.

**Root Cause**: CORR-001 — the screen state machine doesn't model a direct Complete→Menu transition. Navigation goes through intermediate screens.

**Fix**: Add "Back to Main Menu" as a primary action on the Complete screen (bound to `enter` or a dedicated key). The transition table must support Complete→MainMenu as a valid transition. This should be resolved as part of CORR-001 when the formal state machine is built.

**Acceptance Criteria**:
- [ ] Complete screen shows "Press Enter to return to main menu" or similar
- [ ] `enter` on Complete screen navigates directly to MainMenu
- [ ] No intermediate screens shown during transition
- [ ] Navigation works from both success and failure Complete screens
- [ ] Transition table entry: `CompleteResult → MainMenu` on `enter`

**Effort**: included in CORR-001 | **Risk**: low | **Blocks**: CORR-001

---

### EXP-004 (P5-007): Implement or remove 'u' (update) action on Status screen

**Priority**: 🔴 Blocking (CORR-001)

**Problem**: The Status screen shows installed tools but the 'u' key (intended for "update") is a dead no-op. It's bound in the key handler but performs no action. This is confusing—users press 'u' expecting an update and nothing happens. This is either unfinished work or leftover from a removed feature.

**Root Cause**: CORR-001 — StatusUpdate (F7, complexity 19) has partially implemented update logic that was abandoned.

**Fix**: One of two paths:
1. **Implement**: Add `sequoia update <tool>` headless command, wire 'u' to trigger it for the selected tool, show progress similar to install
2. **Remove**: Delete the 'u' key binding and any related dead code from StatusUpdate

Path 2 is recommended for v0.1.0—implement update as a v0.2.0 feature. Path 1 requires significant work (download, verify, replace logic) that's out of scope for this remediation.

**Acceptance Criteria (Path 2 — Remove)**:
- [ ] 'u' key binding removed from Status screen
- [ ] Dead update-related code removed from StatusUpdate
- [ ] StatusUpdate complexity reduced (F7: 19 → ≤10)
- [ ] Status help text no longer mentions 'u'
- [ ] No regression in Status screen behavior

**Effort**: small (1–2h, part of CORR-001) | **Risk**: low | **Blocks**: CORR-001

---

## 🟡 MEDIUM Findings

### EXP-005 (P5-003): Fix Status screen flash when navigating from Complete

**Problem**: When navigating from the Complete screen to Status, the Status screen briefly shows an empty state before loading installed tool data. This flash is visible for ~100–200ms and makes the transition feel broken.

**Fix**: Cache the previous Status screen render. When transitioning from Complete→Status, display the cached render until new data is available. Or: pre-load status data before rendering the Status screen (load in Init(), not in first View()).

**Acceptance Criteria**:
- [ ] No visible empty state flash during Complete→Status transition
- [ ] Status screen shows cached data or loading spinner during transition
- [ ] Transition feels instantaneous to the user
- [ ] Golden test captures the transition sequence

**Effort**: small (1–2h) | **Risk**: low | **Blocks**: none

---

### EXP-006 (P5-004): Fix Error screen retry skipping Configuration

**Problem**: When an installation fails and the Error screen offers "Retry", the retry jumps directly to InstallProgress, skipping the Configuration screen. If the failure was caused by incorrect configuration (wrong paths, missing dependencies), retrying without re-configuring will fail again.

**Fix**: Error screen "Retry" should navigate to Configuration, not directly to InstallProgress. The installation should restart from the configuration step. This matches user expectation: "retry from the beginning."

**Acceptance Criteria**:
- [ ] Error screen "Retry" navigates to Configuration screen
- [ ] Configuration preserves previously selected values as defaults
- [ ] User can modify configuration before re-attempting install
- [ ] Second install attempt goes through full flow: Config→Install→Complete

**Effort**: small (1–2h) | **Risk**: low | **Blocks**: none

---

### EXP-007 (P5-005): Accept uppercase 'Y' in Uninstall confirmation

**Problem**: Uninstall confirmation only accepts lowercase `y`. Pressing uppercase `Y` or holding Shift produces no response. This is a UX papercut—case-insensitive confirmation is a standard pattern across CLI/TUI tools.

**Fix**: Accept both `y` and `Y` for confirmation, both `n` and `N` for cancellation. Use `strings.EqualFold` or `strings.ToLower` for case-insensitive comparison.

**Acceptance Criteria**:
- [ ] Both `y` and `Y` accepted for Uninstall confirmation
- [ ] Both `n` and `N` accepted for Uninstall cancellation
- [ ] Case-insensitive behavior documented in help text
- [ ] Same pattern applied to all confirmation dialogs (Install, Reset, etc.)

**Effort**: small (<30m) | **Risk**: low | **Blocks**: none

---

### EXP-008 (P5-008): Add --tools flag to headless install command

**Priority**: 🔴 Blocking (CORR-005)

**Problem**: Headless mode (`sequoia install`) installs all available tools with no way to select a subset. This makes CI/CD automation impractical—a CI pipeline that needs only one tool must install everything. The TUI has tool selection; headless mode should too.

**Root Cause**: CORR-005 (Headless Mode Missing Feature Parity and Test Coverage)

**Fix**: Add `--tools` flag to `sequoia install` (and `sequoia uninstall`):
```
sequoia install --tools=claude,gemma
sequoia install --tools=all       # default behavior
sequoia install --tools=codex     # single tool
```
The flag accepts comma-separated tool IDs. Validate that specified tools are registered. Error on unknown tool IDs.

**Acceptance Criteria**:
- [ ] `--tools` flag added to `install` and `uninstall` commands
- [ ] Comma-separated tool IDs: `--tools=claude,codex,gemma`
- [ ] `--tools=all` installs everything (default behavior preserved)
- [ ] Unknown tool ID produces clear error message listing available tools
- [ ] `--tools=` with empty value installs nothing (exit 0, no-op)
- [ ] Integration test: install single tool, verify only that tool is installed
- [ ] Integration test: install all tools, verify all installed
- [ ] Integration test: install unknown tool, verify error message

**Effort**: medium (part of CORR-005, 4–6h) | **Risk**: low | **Blocks**: CORR-005

---

### EXP-009 (P5-009): Add j/k navigation keys to Configuration screen

**Problem**: The Configuration screen supports arrow keys for navigation but not `j`/`k` (vim-style). Other screens (ToolSelection) support `j`/`k`, creating an inconsistency. Users who prefer keyboard navigation expect `j`/`k` to work everywhere.

**Fix**: Add `j` (down) and `k` (up) keybindings to Configuration screen. Use the same key handler pattern as ToolSelection. Standardize across all list-based screens via the KeyMap contract.

**Acceptance Criteria**:
- [ ] `j` and `k` navigate options on Configuration screen
- [ ] `j`/`k` behavior matches arrow key behavior exactly
- [ ] `j`/`k` supported on all list-based screens (ToolSelection, Configuration, Status)
- [ ] KeyMap contract enforced: every list screen must support `j`/`k`

**Effort**: small (<1h) | **Risk**: low | **Blocks**: TUI-KBD batch

---

### EXP-010 (P5-010): Show failure count during InstallProgress execution

**Problem**: InstallProgress shows a summary of installation status but hides failure counts during execution. The user only sees "Installing tool X..." and learns about failures at the end. A running counter of "5 installed, 2 failed" during execution provides immediate feedback.

**Fix**: Add a live counter to the InstallProgress view:
```
Installing tools...
✓ claude-code (installed)
✓ gemma-cli (installed)
✗ codex-cli (failed: download error)
⏳ opencode (installing...)

Progress: 2 installed, 1 failed, 1 in progress
```
The counter updates as each tool completes (success or failure), sent via the progress channel.

**Acceptance Criteria**:
- [ ] Live counter showing: installed count, failed count, in-progress count
- [ ] Counter updates after each tool completes (not only at the end)
- [ ] Failed tools shown with brief error reason
- [ ] Counter visible during both install and uninstall flows
- [ ] Final summary matches counter values

**Effort**: small (1–2h) | **Risk**: low | **Blocks**: CORR-001 (depends on progress channel visibility)

---

## 🟡 LOW Findings

### EXP-011 (P5-011): Remove dead ErrorUpdate function

**Problem**: `ErrorUpdate` function in `screens/error.go` is dead code—compiles but never called. This was likely a previous version of the Error screen handler that was replaced but not removed.

**Root Cause**: CORR-004 (Suppressed Linter Rules Enabling Dead Code Accumulation)

**Fix**: Remove `ErrorUpdate` and any associated types/tests that exist only to support it. This is part of the dead code cleanup (QUAL-001).

**Acceptance Criteria**:
- [ ] `ErrorUpdate` function removed from codebase
- [ ] Any associated dead types/tests removed
- [ ] Error screen behavior unchanged (uses the active implementation)
- [ ] Linter passes without dead code warnings

**Effort**: small (<30m, part of CORR-004) | **Risk**: low | **Blocks**: CORR-004

---

### EXP-012 (P5-012): Fix Configuration Engram option async morph

**Problem**: The Engram option on the Configuration screen silently changes state after async detection completes. The user selects an option, and while they're configuring other settings, the Engram option morphs from "Detecting..." to "Available/Unavailable" without user interaction. This is disorienting—options shouldn't change state after initial render.

**Fix**: Make Engram detection synchronous during screen initialization, or show a "Pending" state that requires user acknowledgment before changing. Don't silently morph an option the user may have already interacted with.

**Acceptance Criteria**:
- [ ] Engram detection completes before Configuration screen renders (synchronous)
- [ ] Or: Engram shows "Checking..." badge that becomes "Available/Unavailable" on explicit refresh
- [ ] User can trigger re-detection manually (not automatic morph)
- [ ] Configuration screen state is stable after initial render

**Effort**: small (<2h) | **Risk**: low | **Blocks**: none

---

### EXP-013 (P5-013): Warn on empty tool selection

**Problem**: When no tools are selected on the ToolSelection screen and the user presses Enter, the screen navigates back with no warning. The user doesn't know that pressing Enter with no selection means "go back / select nothing." This can lead to accidental no-op installations.

**Fix**: When Enter is pressed with no tools selected, show a brief warning: "No tools selected. Press Enter again to go back, or select at least one tool." On second Enter, navigate back. Alternatively: disable Enter when no tools are selected.

**Acceptance Criteria**:
- [ ] Warning message shown on first Enter with empty selection
- [ ] Second Enter navigates back (confirms intent to select nothing)
- [ ] Or: Enter is disabled when no tools are selected (only `esc`/`q` to go back)
- [ ] Warning message is clearly visible and disappears after selection changes

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

### EXP-014 (P5-014): Disable global 'q' during Uninstall confirmation

**Problem**: Global `q` keybinding kills the entire app during any screen, including Uninstall confirmation. The user might press `q` intending to answer "no" to the confirmation prompt, but instead the app exits. Global quit should be overridden on confirmation screens.

**Fix**: On confirmation screens (Uninstall, Reset, etc.), override the global `q` handler:
- `q` → cancel confirmation (same as `n`/`N`)
- `ctrl+c` → still quits (hard quit is always available)
- `esc` → cancel confirmation

**Acceptance Criteria**:
- [ ] `q` on Uninstall confirmation cancels (same as `n`)
- [ ] `ctrl+c` on any screen still quits (hard quit preserved)
- [ ] `q` on main menu and non-confirmation screens still quits (unchanged)
- [ ] Confirmation behavior documented in help text

**Effort**: small (<30m) | **Risk**: low | **Blocks**: none

---

## ℹ️ INFO Findings

### EXP-015 (P5-015): Add screen reader metadata to TUI

**Problem**: All TUI screens are text-only with no accessibility metadata (ARIA labels, screen reader hints). This makes the TUI unusable for developers relying on screen readers. Bubbletea supports accessibility features but they're not used.

**Fix**: This is deferred for v0.1.0. For future consideration:
- Add `AccessibleLabel` to each screen model
- Use `lipgloss.Style`'s accessibility features
- Test with `orca` or `NVDA` screen reader

**Acceptance Criteria (deferred)**:
- [ ] Deferred to v0.2.0+
- [ ] Each screen model has an `AccessibleLabel` field
- [ ] Screen reader testing performed on at least one platform
- [ ] Accessibility documented in CONTRIBUTING.md

**Effort**: deferred | **Risk**: low | **Blocks**: none

---

## Task Summary

| Priority | Task ID | Finding | Title | Effort | Blocks |
|----------|---------|---------|-------|--------|--------|
| 🔴 HIGH | EXP-001 | P5-001 | Animated spinner on InstallProgress | small (1–2h) | CORR-001 |
| 🔴 HIGH | EXP-002 | P5-002 | Cancel/abort on InstallProgress | small (2–4h) | CORR-001 |
| 🔴 HIGH | EXP-003 | P5-006 | Return to menu from Complete screen | (in CORR-001) | CORR-001 |
| 🔴 HIGH | EXP-004 | P5-007 | Implement or remove 'u' on Status | small (1–2h) | CORR-001 |
| 🟡 MED | EXP-005 | P5-003 | Fix Status screen flash | small (1–2h) | — |
| 🟡 MED | EXP-006 | P5-004 | Fix Error retry routing | small (1–2h) | — |
| 🟡 MED | EXP-007 | P5-005 | Case-insensitive confirmations | small (<30m) | — |
| 🟡 MED | EXP-008 | P5-008 | Add --tools flag to headless | medium (4–6h) | CORR-005 |
| 🟡 MED | EXP-009 | P5-009 | j/k navigation on Configuration | small (<1h) | TUI-KBD |
| 🟡 MED | EXP-010 | P5-010 | Failure counter on InstallProgress | small (1–2h) | CORR-001 |
| 🟡 LOW | EXP-011 | P5-011 | Remove dead ErrorUpdate | small (<30m) | CORR-004 |
| 🟡 LOW | EXP-012 | P5-012 | Fix Engram async morph | small (<2h) | — |
| 🟡 LOW | EXP-013 | P5-013 | Empty tool selection warning | small (<1h) | — |
| 🟡 LOW | EXP-014 | P5-014 | Disable global q during confirm | small (<30m) | TUI-KBD |
| ℹ️ INFO | EXP-015 | P5-015 | Screen reader metadata | deferred | — |

**Priority Order**: EXP-007 + EXP-014 (quick papercuts, independent) → EXP-005 + EXP-006 (navigation fixes, independent) → EXP-001 + EXP-002 + EXP-010 (InstallProgress UX, depends on CORR-001) → EXP-003 + EXP-004 + EXP-009 (CORR-001 batch) → EXP-008 (CORR-005) → EXP-011 (CORR-004) → EXP-012 + EXP-013 (backlog)

---

*Generated by Sequoia v1.0.9 — M2 Reporter | Audit schema v1.0*
