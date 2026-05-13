# Delta for go-wiring: Backup Permissions

Reference: `openspec/specs/go-wiring/spec.md` (existing)

## ADDED Requirements

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

## Scope
This is a narrowing of existing backup behavior — no new capabilities, no API changes.
