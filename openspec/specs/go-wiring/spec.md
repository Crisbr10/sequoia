# go-wiring Specification

## Purpose

Define the Go code changes to pass the user-selected language from the TUI configuration through the pipeline runner to adapter Install/Uninstall calls, including language-aware template resolution via `RenderTemplateLang`.

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

### Requirement: Adapters Use Language for Template Selection
All adapter `Install()` implementations that render templates (BaseAdapter, Codex, `_template`) MUST use `opts.Language` to select language-specific templates via `RenderTemplateLang`. The `_ = opts.Language` discard pattern SHALL be removed from these adapters. Uninstall methods and adapters that do NOT render templates MAY keep the language parameter but MUST NOT discard it with `_ = opts.Language`.

#### Scenario: BaseAdapter passes language to template rendering
- GIVEN `BaseAdapter.Install()` with `opts.Language = "es"`
- WHEN the skill and system prompt templates are rendered
- THEN `RenderTemplateLang(fs, name, "es", data)` SHALL be called
- AND the function SHALL attempt to load `{name}.es.tmpl` before falling back to `{name}.tmpl`

#### Scenario: Codex adapter uses language for skill template
- GIVEN `codex.Adapter.Install()` with `opts.Language = "en"`
- WHEN the skill template is rendered
- THEN `RenderTemplateLang(templateFS, "templates/skill.md", "en", data)` SHALL be called

#### Scenario: Default to English when language is empty
- GIVEN `opts.Language = ""` (empty string)
- WHEN any adapter renders templates
- THEN the language SHALL default to `"en"`

### Requirement: Language-Aware Template Resolution
`adapters/common/template.go` MUST provide a `RenderTemplateLang(fs embed.FS, name, lang string, data interface{}) (string, error)` function. The function SHALL first attempt to load `{name}.{lang}.tmpl` from the embedded FS. If the language-specific file does not exist, it SHALL fall back to `{name}.tmpl` for backward compatibility.

#### Scenario: Lang-resolved template found
- GIVEN `name = "skill.md"`, `lang = "en"`, and `skill.md.en.tmpl` exists in the FS
- WHEN `RenderTemplateLang(fs, "skill.md", "en", data)` is called
- THEN `skill.md.en.tmpl` SHALL be loaded and rendered

#### Scenario: Fallback when lang file missing
- GIVEN `name = "skill.md"`, `lang = "es"`, and `skill.md.es.tmpl` does NOT exist
- WHEN `RenderTemplateLang(fs, "skill.md", "es", data)` is called
- THEN `skill.md.tmpl` SHALL be loaded as fallback
- AND no error SHALL be returned

#### Scenario: Unknown language falls back
- GIVEN `name = "skill.md"`, `lang = "zh"`, and `skill.md.zh.tmpl` does NOT exist
- WHEN `RenderTemplateLang(fs, "skill.md", "zh", data)` is called
- THEN `skill.md.tmpl` SHALL be loaded as fallback
- AND existing behavior SHALL be preserved

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
