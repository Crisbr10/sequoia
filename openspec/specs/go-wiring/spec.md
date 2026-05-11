# go-wiring Specification

## Purpose

Define the Go code changes to pass the user-selected language from the TUI configuration through the pipeline runner to adapter Install/Uninstall calls, replacing the `_ = lang` placeholder.

## Requirements

### Requirement: InstallOpts Struct with Language Field
A new `InstallOpts` struct SHALL be defined in `adapters/interface.go` with a `Language` field of type `string`. The struct MUST be designed for future extension with additional options.

#### Scenario: InstallOpts defined in adapter interface
- GIVEN the `adapters` package
- WHEN `InstallOpts` is defined
- THEN it SHALL export a `Language string` field
- AND the struct SHALL be placed above the `ToolAdapter` interface definition

### Requirement: Install and Uninstall Accept InstallOpts
The `ToolAdapter` interface methods `Install()` and `Uninstall()` MUST be changed to `Install(opts InstallOpts)` and `Uninstall(opts InstallOpts)`. All 5 adapter implementations (opencode, claude, gemini, cursor, codex) and the `_template` adapter SHALL be updated.

#### Scenario: Adapter interface updated
- GIVEN the `ToolAdapter` interface
- WHEN the change is applied
- THEN `Install() error` SHALL become `Install(opts InstallOpts) error`
- AND `Uninstall() error` SHALL become `Uninstall(opts InstallOpts) error`

#### Scenario: All adapter implementations compile
- GIVEN all 5 adapters updated
- WHEN `go build ./...` runs
- THEN compilation SHALL succeed with zero errors

#### Scenario: Unused parameter is explicit
- GIVEN an adapter that does not use language for its operations
- WHEN the `Install(opts InstallOpts)` method is implemented
- THEN `_ = opts.Language` SHALL explicitly mark the parameter as reserved

### Requirement: Pipeline Runner Passes Language
`internal/pipeline/runner.go` MUST construct `InstallOpts{Language: lang}` from the `lang string` parameter and pass it to `adapter.Install()` and `adapter.Uninstall()`. The `_ = lang` placeholder SHALL be removed.

#### Scenario: Runner passes language to Install
- GIVEN `RunInstall(ctx, tools, ch, "es")` is called
- WHEN the install goroutine executes
- THEN `adapter.Install(InstallOpts{Language: "es"})` SHALL be called

#### Scenario: Runner passes language to Uninstall
- GIVEN `RunUninstall(ctx, tools, ch, "en")` is called
- WHEN the uninstall goroutine executes
- THEN `adapter.Uninstall(InstallOpts{Language: "en"})` SHALL be called

### Requirement: All 312+ Tests Remain Green
After all wiring changes, `go test -race -count=1 ./...` MUST pass with zero failures.

#### Scenario: Test suite passes after wiring
- GIVEN all interface and implementation changes applied
- WHEN `go test -race -count=1 ./...` runs
- THEN all existing tests SHALL pass
- AND no race conditions SHALL be detected
