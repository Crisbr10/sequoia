# pipeline-phase-progress

> Source: CORR-004 — Common Package Split (archived 2026-05-22)
> Domain: NEW — pipeline phase observability

The pipeline MUST expose discrete phases (Prepare, Download, Verify, Stage, Apply) as public functions, each sending typed `ProgressMsg` values through a channel. Consumers SHALL observe, cancel, or recover at phase boundaries.

## Requirements

| # | Requirement | Strength |
|---|------------|----------|
| R1 | Pipeline exposes `Prepare()`, `Download()`, `Verify()`, `Stage()`, `Apply()` as public functions | MUST |
| R2 | Each phase returns a receive-only channel (`<-chan PhaseProgress`) carrying phase-typed progress | MUST |
| R3 | `ProgressMsg.Phase` distinguishes phases (`PhaseDownloading`, `PhaseInstalling`, `PhaseVerifying`) | MUST |
| R4 | Phases accept `context.Context` — cancellation aborts current phase without corrupting state | MUST |
| R5 | Backward-compatible `Run(ctx, adapter, tools)` wrapper preserves existing callers | MUST |
| R6 | Phase errors include phase name, tool ID, and root cause via `%w` wrapping | MUST |

## Scenarios

### Single-tool install through all phases
- GIVEN a tool adapter satisfying `Installer`
- WHEN `Run(ctx, adapter, [tool])` executes
- THEN the progress channel emits at least one message per phase
- AND the final message has `Done=true` and empty `Error`

### Cancellation at phase boundary
- GIVEN a running `Download()` phase
- WHEN the context is cancelled
- THEN the phase returns `ctx.Err()` wrapped with phase context
- AND no subsequent phases execute
- AND downloaded artifacts are cleaned up

### Phase-level error recovery
- GIVEN `Download()` fails for one tool in a multi-tool run
- WHEN the pipeline continues
- THEN `Verify()` and subsequent phases skip the failed tool
- AND the error for the failed tool is reported with phase `PhaseDownloading`

### Individual phase execution
- GIVEN a consumer that only needs to re-verify an installed tool
- WHEN `Verify(ctx, tool, ch)` is called directly
- THEN only verification logic runs
- AND Install/Download phases are not executed
