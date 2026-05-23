# pipeline-interface-narrowing

> Source: CORR-004 — Common Package Split (archived 2026-05-22)
> Domain: MODIFIED — P3-003 ISP narrowing, P3-010 Strategy activation

Pipeline and CLI consumers SHALL consume narrow role interfaces instead of the 11-method `ToolAdapter`. Unguarded type assertions SHALL be replaced with interface method dispatch.

## Requirements

| # | Requirement | Strength |
|---|------------|----------|
| R7 | `targetAdapters()` returns `[]Identifier` (only `ID()` and `Name()` needed) | MUST |
| R8 | `runInstall()` and `runUninstall()` accept `[]Installer` + `[]Identifier` | MUST |
| R9 | `runStatus()` accepts `[]Detector` + `[]AdapterPaths` + `Status()` | MUST |
| R10 | Pipeline dispatches via `Strategy` interface; zero unguarded type assertions in `runner.go` | MUST |
| R11 | `MockAdapter` split into `MockIdentifier`, `MockDetector`, `MockInstaller` | MUST |
| R12 | `grep 'ToolAdapter' cmd/sequoia/` returns zero results after refactor | MUST |

## Scenarios

### Consumer uses only Identifier methods
- GIVEN `targetAdapters("claude-code", reg)` returns `[]Identifier`
- WHEN caller iterates the result
- THEN only `ID()` and `Name()` are accessible
- AND compilation fails if caller attempts `Detect()` or `Install()`

### Pipeline calls Strategy interface
- GIVEN a tool state wrapping an adapter that implements `Strategy`
- WHEN `runInstallSteps()` dispatches the install
- THEN it calls `adapter.Strategy().Install(opts)` without type assertion
- AND no `panic` can occur from mismatched adapter types

### Graceful handling of non-conforming adapter
- GIVEN a tool state whose adapter does NOT implement `Strategy`
- WHEN pipeline attempts to install
- THEN an error is returned: "adapter %s: does not implement Strategy"
- AND no panic occurs
