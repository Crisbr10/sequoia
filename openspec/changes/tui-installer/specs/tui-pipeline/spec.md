# tui-pipeline Specification

## Purpose
Goroutine-based installer bridging TUI screens to `adapters.ToolAdapter` calls: progress channel, error surfacing, and retry.

## Requirements

### Requirement: Pipeline Execution
The install pipeline MUST run each selected tool's `Install()` in a separate goroutine, sending `ProgressMsg` events through a buffered channel (capacity 64). The channel SHALL be closed when all goroutines complete or on context cancellation. Uninstall MUST follow the same pattern calling `Uninstall()`.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 1 | Single tool install | 1 tool selected | Pipeline started | Goroutine calls `Install()`; progress per step received |
| 2 | Multi-tool concurrent | 3 tools selected | Pipeline started | 3 goroutines; progress interleaved per tool |
| 3 | Channel closure | All goroutines return | Last finishes | Channel closed; final message transitions screen |
| 4 | Context cancellation | `q` pressed during install | Context cancelled | Goroutines stop; channel closed; partial state preserved |

### Requirement: Error Surfacing
Errors from `Install()` or `Uninstall()` MUST be captured per-tool and per-step. `ProgressMsg.Error` SHALL contain the wrapped error. The Error screen SHALL display tool name and error message for each failure.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 5 | Install error | `Adapter.Install()` returns error | Error occurs | `ProgressMsg{Status:"error", Error: wrappedErr}` sent |
| 6 | Multi-error display | 2 of 3 tools fail | Pipeline completes | Error screen lists both failures with distinct messages |

### Requirement: Retry
Retry MUST launch a fresh pipeline for failed tools only, preserving partial state for succeeded tools.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 7 | Retry failed only | Tool A succeeded, Tool B failed | Retry triggered | Only Tool B `Install()` called; Tool A untouched |
| 8 | Retry succeeds | Failed tool retried | Retry completes | All tools ✅; Complete screen shown |
| 9 | Retry fails again | Failed tool retried | Retry fails again | Error screen shown; retry still available |
