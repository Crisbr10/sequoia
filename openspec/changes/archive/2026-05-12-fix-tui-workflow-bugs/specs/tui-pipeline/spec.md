# Delta for tui-pipeline

## MODIFIED Requirements

### Requirement: Pipeline Execution

The install pipeline MUST run each selected tool's `Install()` in a separate goroutine, sending `ProgressMsg` events through a buffered channel (capacity 64). The channel SHALL be closed when all goroutines complete or on context cancellation. Uninstall MUST follow the same pattern calling `Uninstall()`. The Model SHALL provide a `startPipeline(mode string)` method that builds ProgressTools, resets progress counters, starts the appropriate pipeline (install or uninstall), starts progress polling, and returns the batched tea commands. This method MUST be called from ALL entry points to the InstallProgress screen.
(Previously: no shared startPipeline method; entry points built pipeline state inconsistently or skipped it entirely.)

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 1 | Single tool install | 1 tool selected | Pipeline started | Goroutine calls `Install()`; progress per step received |
| 2 | Multi-tool concurrent | 3 tools selected | Pipeline started | 3 goroutines; progress interleaved per tool |
| 3 | Channel closure | All goroutines return | Last finishes | Channel closed; final message transitions screen |
| 4 | Context cancellation | `q` pressed during install | Context cancelled | Goroutines stop; channel closed; partial state preserved |
| 5 | startPipeline — install | Selected tools available | `startPipeline("install")` called | ProgressTools built; counters reset to 0/N; RunInstall launched; poll command batched; returned as `tea.Batch` |
| 6 | startPipeline — uninstall | Selected tools available | `startPipeline("uninstall")` called | ProgressTools built; counters reset to 0/N; RunUninstall launched; poll command batched; returned as `tea.Batch` |
