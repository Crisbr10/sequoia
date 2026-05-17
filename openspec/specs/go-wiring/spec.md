# go-wiring Specification

## Purpose

Define the Go wiring between TUI configuration, pipeline runner, and adapter Install/Uninstall calls. The language parameter has been removed; all adapters now use zero-argument signatures and direct template resolution.

## Requirements

### Requirement: InstallOpts Struct Without Language

`InstallOpts` SHALL be defined in `adapters/interface.go` without any `Language` field. The struct SHALL be designed for future extension with non-language options only.

#### Scenario: InstallOpts lacks Language field
- GIVEN the `adapters` package
- WHEN `InstallOpts` is inspected
- THEN it SHALL NOT contain a `Language string` field

### Requirement: Install and Uninstall Zero-Parameter Signatures

The `ToolAdapter` interface methods `Install()` and `Uninstall()` SHALL use zero-parameter signatures: `Install(ctx context.Context) error` and `Uninstall(ctx context.Context) error`. All 5 adapter implementations (opencode, claude, gemini, cursor, codex) and the `_template` adapter SHALL follow this signature.

#### Scenario: Adapter interface uses zero-param signatures
- GIVEN the `ToolAdapter` interface
- WHEN compiled
- THEN `Install()` and `Uninstall()` accept only `context.Context` — no `InstallOpts` parameter

#### Scenario: All adapter implementations compile
- GIVEN all 5 adapters and `_template`
- WHEN `go build ./...` runs
- THEN compilation SHALL succeed with zero errors

### Requirement: Direct Template Resolution Without Language

Adapter `Install()` implementations SHALL render templates via `RenderTemplate(fs embed.FS, name string, data interface{}) (string, error)` without any language resolution. The `RenderTemplateLang` function SHALL be deleted. Templates SHALL be resolved from a single `templates/` directory with no language subdirectory.

#### Scenario: BaseAdapter uses RenderTemplate
- GIVEN `BaseAdapter.Install()` is called
- WHEN the skill and system prompt templates are rendered
- THEN `RenderTemplate(fs, "templates/skill.md.tmpl", data)` SHALL be called
- AND no language parameter is involved

#### Scenario: Codex adapter uses RenderTemplate
- GIVEN `codex.Adapter.Install()` is called
- WHEN the skill template is rendered
- THEN `RenderTemplate(fs, "templates/skill.md.tmpl", data)` SHALL be called

### Requirement: Pipeline Runner Without Language Options

`internal/pipeline/runner.go` SHALL NOT construct `InstallOpts` with a `Language` field. `RunInstall(ctx, tools, ch)` and `RunUninstall(ctx, tools, ch)` SHALL NOT accept a `lang string` parameter. The internal helper functions `runSteps`, `runInstallSteps`, and `runUninstallSteps` SHALL NOT accept `lang`.

#### Scenario: Runner calls Install without language
- GIVEN `RunInstall(ctx, tools, ch)` is called
- WHEN the install goroutine executes
- THEN `adapter.Install(ctx)` SHALL be called with no InstallOpts

#### Scenario: Runner calls Uninstall without language
- GIVEN `RunUninstall(ctx, tools, ch)` is called
- WHEN the uninstall goroutine executes
- THEN `adapter.Uninstall(ctx)` SHALL be called with no InstallOpts

### Requirement: go.mod i18n-Free

`go.mod` SHALL NOT depend on `github.com/nicksnyder/go-i18n/v2`. `golang.org/x/text` SHALL be absent from direct dependencies. `github.com/BurntSushi/toml` SHALL remain as it is used by the Codex adapter for TOML configuration merging.

#### Scenario: Dependencies verified
- GIVEN `go.mod` after `go mod tidy`
- WHEN inspected
- THEN `go-i18n/v2` absent; `BurntSushi/toml` present

### Requirement: Test Suite Green

After all wiring changes, `go test -race -count=1 ./...` MUST pass with zero failures attributable to i18n removal.

#### Scenario: Test suite passes
- GIVEN all interface and implementation changes applied
- WHEN `go test -race -count=1 ./...` runs
- THEN all core package tests SHALL pass
- AND no race conditions SHALL be detected

### Requirement: Backup File and Directory Isolation

The system SHALL ensure that backup files and directories created during install and upgrade operations use owner-only permissions to prevent information leaks on multi-user Unix systems.

#### Scenario: Backup directory permissions
- **GIVEN** an installer with BackupDir configured
- **WHEN** Prepare() backs up existing files
- **THEN** the backup directory SHALL be created with `0o700` (owner rwx only)

#### Scenario: Backup file copy permissions
- **GIVEN** an existing file in TargetDir
- **WHEN** Prepare() backs up that file via copyFile
- **THEN** the backup file SHALL have `0o600` (owner rw only)

#### Scenario: ReplaceFile backup permissions
- **GIVEN** a non-Sequoia-managed file at the target path
- **WHEN** ReplaceFile creates a timestamped backup
- **THEN** the backup file SHALL be written with `0o600` (owner rw only)

#### Scenario: Codex MergeConfig backup permissions
- **GIVEN** an existing config.toml with content
- **WHEN** MergeConfig creates a timestamped backup
- **THEN** the backup file SHALL be written with `0o600` (owner rw only)

#### Scenario: Production permissions unchanged
- **GIVEN** any install or upgrade operation
- **WHEN** production files are written (skills, commands, version markers, configs)
- **THEN** production file permissions SHALL remain at `0o644` (files) and `0o755` (directories)
