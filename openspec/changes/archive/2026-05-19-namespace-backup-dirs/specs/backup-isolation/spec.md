# backup-isolation Specification

## Purpose

Ensures installer backup directories are namespaced by adapter identity and installer type, preventing cross-contamination between installers and adapters during rollback.

## Requirements

### Requirement: Installer-type backup isolation

Each installer type (skills, commands) within an adapter MUST use an independent backup subdirectory. One installer's Rollback() MUST NOT affect another installer's backups.

#### Scenario: Skill backup survives command installer failure

- **GIVEN** an install with both skill and command files, skill installer completed successfully
- **WHEN** the command installer fails during Apply
- **AND** the command installer rolls back (calling os.RemoveAll on its own subdirectory)
- **THEN** the skill installer's backup subdirectory MUST remain intact
- **AND** the skill installer's subsequent Rollback() MUST successfully restore the original skill file

#### Scenario: Successful install creates isolated subdirectories

- **GIVEN** a fresh install with no prior files at target
- **WHEN** both installers succeed (Prepare → Apply → Verify)
- **THEN** each installer MUST create its backup under a type-specific subdirectory (e.g., `skills/`, `commands/`)
- **AND** no installer's lifecycle phase interacts with another installer's backup path

### Requirement: Adapter-level backup namespacing

Backup directory paths MUST include the adapter ID to prevent collisions between concurrent adapter installs.

#### Scenario: Concurrent adapters use disjoint backup paths

- **GIVEN** adapter A with ID "claude" and adapter B with ID "opencode"
- **WHEN** both adapters install concurrently
- **THEN** adapter A's backup path MUST contain the segment "claude"
- **AND** adapter B's backup path MUST contain the segment "opencode"
- **AND** the two backup directory trees are entirely disjoint (no shared parent that an os.RemoveAll could destroy)

#### Scenario: Single adapter retains unique session suffix

- **GIVEN** a single adapter installing twice in rapid succession
- **WHEN** both installs complete within the same millisecond (theoretical edge case)
- **THEN** each install's backup path MUST include a session suffix to avoid collision
- **AND** the suffix remains part of the path composition: `{base}-{adapterID}-{sessionSuffix}/{installerType}`

### Requirement: Backward-compatible install path

Single-adapter installs that never trigger the error path MUST produce identical installed files as before.

#### Scenario: Fresh install produces identical output files

- **GIVEN** an adapter using the new namespaced backup paths
- **WHEN** the install succeeds without any rollback
- **THEN** the installed SKILL.md and command files are byte-for-byte identical to pre-fix behavior
- **AND** the system prompt file is unchanged

#### Scenario: Rollback-only path change

- **GIVEN** an install that succeeds completely (no rollback)
- **WHEN** comparing the filesystem state after install
- **THEN** the only observable difference is the deeper backup directory structure (`{base}-{adapterID}-{suffix}/skills` instead of `{base}-{suffix}`)
- **AND** no backup cleanup is performed (cleanup is out of scope)
