# Specs: CLI Version Resolution

## Domain: cli-version-resolution

### Requirement: TUI Welcome Screen MUST Show Resolved Build Version

The TUI Welcome screen MUST display the resolved CLI version obtained via `resolveVersion()`, NOT the raw `"0.1.0-dev"` default. The `resolveVersion` function SHALL be the single source of truth for CLI version resolution, shared by `newVersionCmd` and `runTUI`.

#### Scenario: Dev version resolves via debug.ReadBuildInfo

- GIVEN `Version == "0.1.0-dev"` (no ldflags) and `debug.ReadBuildInfo()` returns `Main.Version = "v0.1.0"`
- WHEN `resolveVersion("0.1.0-dev")` is called
- THEN it MUST return `"v0.1.0"`

#### Scenario: ldflags version passes through unchanged

- GIVEN `Version == "1.2.3"` (set via ldflags)
- WHEN `resolveVersion("1.2.3")` is called
- THEN it MUST return `"1.2.3"` unchanged

#### Scenario: runTUI passes resolved version to model

- GIVEN `Version == "0.1.0-dev"` and `debug.ReadBuildInfo()` returns `"v1.0.0-abc12345"`
- WHEN `runTUI("")` is called
- THEN `app.NewModel` MUST receive `"v1.0.0-abc12345"`, NOT `"0.1.0-dev"`

#### Scenario: version command and TUI display same version

- GIVEN the same binary is used
- WHEN the user runs `sequoia version` and launches the TUI Welcome screen
- THEN both MUST display identical version strings

#### Scenario: Devel build with no VCS info returns raw default

- GIVEN `Version == "0.1.0-dev"` and `debug.ReadBuildInfo()` returns `Main.Version = "(devel)"` with no `vcs.revision` setting
- WHEN `resolveVersion("0.1.0-dev")` is called
- THEN it MUST return `"0.1.0-dev"` (raw fallback)
