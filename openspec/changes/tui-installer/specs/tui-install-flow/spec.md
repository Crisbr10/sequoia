# tui-install-flow Specification

## Purpose
Guided install flow: Welcome → Tool Selection → Configuration → Install Progress → Complete/Error.

## Requirements

### Requirement: Welcome Screen
The Welcome screen MUST display branding (name, version, tagline) and list auto-detected tools with install status. `Enter` or `→` MUST transition to Tool Selection.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 1 | No tools detected | 0 adapters return `Detect()==true` | Welcome renders | "No tools detected" shown; Enter still advances |
| 2 | Tools detected | 2 adapters detected | Welcome renders | Both listed with `[installed]` / `[not installed]` |
| 3 | Transition | User at Welcome | Presses Enter | Screen advances to Tool Selection |

### Requirement: Tool Selection Screen
The Tool Selection screen MUST show a checkbox list of detected tools. `Space` toggles selection. At least one tool MUST be selected to proceed; zero selection SHALL show a validation error.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 4 | Multi-select toggle | 3 tools shown, 0 selected | Space on tool 1, then tool 2 | Both `[x]`, third `[ ]` |
| 5 | Zero selected validation | 0 tools selected | Enter pressed | Error "Select at least one tool"; screen stays |
| 6 | Proceed with selection | ≥1 tool selected | Enter pressed | Transitions to Configuration |

### Requirement: Configuration Screen
The Configuration screen MUST offer language (EN/ES) and persistence backend (Engram/Files/Both) selectors. Engram option SHALL be disabled when MCP is not detected.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 7 | Engram available | Engram MCP detected | Config renders | "Engram" option selectable |
| 8 | Engram unavailable | Engram MCP not detected | Config renders | "Engram" greyed out with "(not detected)" note |
| 9 | Proceed | User selects options | Enter pressed | Transitions to Install Progress |

### Requirement: Install Progress Screen
The Install Progress screen MUST show per-tool step-by-step progress with a spinner while running. A goroutine SHALL execute `Install()` and send `ProgressMsg` events through a buffered channel. The UI MUST remain responsive during installation.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 10 | Happy path | 1 tool selected | Install runs | Steps: `[ ] Skills` → spinner → `[✓] Skills` → `[✓] Commands` → `[✓] System Prompt` |
| 11 | Multi-tool parallel | 2 tools selected | Install runs | Independent progress blocks for each tool |
| 12 | Step failure | Install goroutine errors | Step fails | `[✗]` with error message; remaining steps skipped; Error screen |
| 13 | All success | All tools/steps complete | Last step finishes | Transitions to Complete screen |

### Requirement: Complete Screen
The Complete screen MUST list succeeded tools and show the first command to try (`/sequoia-init`). `q` exits the TUI.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 14 | All succeeded | 2 of 2 installed | Install finishes | Both listed with ✅; `/sequoia-init` hint shown |
| 15 | Partial success | 1 succeeded, 1 failed | Install finishes | Succeeded shown here; failed deferred to Error screen |

### Requirement: Error Screen
The Error screen MUST list failed tools with error messages. `r` MUST trigger retry of failed tools only. `q` MUST quit while preserving partial state.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 16 | Retry failed tools | 1 tool failed | User presses `r` | Fresh pipeline for failed tool; returns to Progress |
| 17 | Quit on error | 1 tool failed | User presses `q` | TUI exits; partial install state preserved |
