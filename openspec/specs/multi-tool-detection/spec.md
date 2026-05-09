# Specs: Multi-Tool Detection

## Domain: multi-tool-detection

### Requirement: ScanTools Returns Structured Detection Results

The system MUST provide a `ScanTools()` function that iterates all registered adapters via `DefaultRegistry.All()`, calls `Status()` on each, and returns a `[]AdapterStatus` slice.

**Scenario: All adapters detected on macOS**
- GIVEN Claude Code and OpenCode are installed on macOS
- WHEN `ScanTools()` is called
- THEN the result MUST contain two `AdapterStatus` entries with `Installed=true`
- AND each entry's `Path` field MUST be the resolved `SkillsPath()`

**Scenario: Only Claude Code detected on Windows**
- GIVEN only Claude Code is installed on Windows
- WHEN `ScanTools()` is called
- THEN the result MUST contain two entries (Claude Code and OpenCode)
- AND Claude Code's `Detect()=true` but OpenCode's `Detect()=false`

**Scenario: No tools detected**
- GIVEN no supported AI tools are installed on the machine
- WHEN `ScanTools()` is called
- THEN each entry MUST report `Installed=false` and `Version=""`

### Requirement: Cross-Platform Detection

Each adapter's `Detect()` method MUST check for the tool's home directory AND its binary in PATH, operating identically on macOS, Linux, and Windows.

**Scenario: OpenCode detection on all platforms**
- GIVEN the OpenCode adapter is registered
- WHEN `Detect()` is called
- THEN it MUST return `true` if `~/.config/opencode/` exists OR `opencode` is in PATH
- AND it MUST return `false` only when both checks fail
