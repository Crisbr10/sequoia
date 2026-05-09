# Specs: Status Reporting

## Domain: status-reporting

### Requirement: CLI Status Output Uses AdapterStatus

The `runStatus` function MUST call `a.Status()` once per adapter instead of separately calling `Detect()` and `IsInstalled()`. The output SHALL be a fixed-width column-aligned table with columns: ID, NAME, DETECTED, INSTALLED, VERSION, PATH.

**Scenario: Fully installed tool shown**
- GIVEN Claude Code has Sequoia installed at `~/.claude/skills/sequoia/` with version `0.1.0`
- WHEN `sequoia status` is run
- THEN the output MUST display the tool name, resolved path, `INSTALLED=yes`, and `VERSION=0.1.0`

**Scenario: Detected but not installed**
- GIVEN OpenCode is detected (`Detect()=true`) but `IsInstalled()=false`
- WHEN `sequoia status` is run
- THEN the output MUST show `INSTALLED=no` and `VERSION=""`

**Scenario: Not detected at all**
- GIVEN a registered adapter returns `Detect()=false`
- WHEN `sequoia status` is run
- THEN the output MUST show `INSTALLED=no`, `VERSION=""`, and the adapter's default path

### Requirement: No Adapters Registered Edge Case

- GIVEN `DefaultRegistry.All()` returns an empty slice
- WHEN `runStatus` is called
- THEN it MUST print "No adapters registered." and return nil
