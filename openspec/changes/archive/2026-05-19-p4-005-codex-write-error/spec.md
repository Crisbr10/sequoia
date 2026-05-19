# Delta for Session File Error Propagation

## ADDED Requirements

None

## MODIFIED Requirements

### Requirement: MergeConfig SHALL propagate session file write errors

`MergeConfig` (`adapters/codex/installer.go`) SHALL return an error when `AtomicWriteFile` for the `.sequoia-session` file fails. The error SHALL be wrapped with context using the pattern `fmt.Errorf("merge config: session: %w", err)`.

(Previously: session file write errors were silently discarded in an empty `if err != nil` block at lines 36-39.)

#### Scenario: Session file write fails during config merge

- **GIVEN** `MergeConfig` is called and a backup file exists (non-empty config)
- **WHEN** `AtomicWriteFile` for `.sequoia-session` fails (e.g., disk full, permission denied, read-only FS)
- **THEN** `MergeConfig` SHALL return a non-nil error
- **AND** the error SHALL be wrapped as `"merge config: session: <underlying error>"`

#### Scenario: Session file write succeeds

- **GIVEN** `MergeConfig` is called and a backup file exists
- **WHEN** `AtomicWriteFile` for `.sequoia-session` succeeds
- **THEN** `MergeConfig` SHALL continue normally
- **AND** SHALL write the merged config and return nil (no change from prior behavior)

#### Scenario: Backward compatibility — existing callers compile and behave correctly

- **GIVEN** the modified `MergeConfig` with the error return
- **WHEN** `Adapter.Install()` calls `MergeConfig`
- **THEN** the call site SHALL compile without signature changes
- **AND** `Adapter.Install()` SHALL trigger rollback (skills + commands) on error, matching current behavior at `adapter.go:140-148`
- **AND** `RemoveConfig` SHALL remain unchanged and compile without modification

### Requirement: ReplaceFile SHALL propagate session file write errors

`ReplaceFile` (`adapters/common/strategy.go`) SHALL return an error when `AtomicWriteFile` for the `.sequoia-session` file fails, instead of silently discarding it.

(Previously: session file write errors were silently discarded in an empty `if err != nil` block at lines 132-135.)

#### Scenario: Session file write fails during file replace

- **GIVEN** `ReplaceFile` is called on a non-Sequoia-managed, existing file
- **WHEN** `AtomicWriteFile` for `.sequoia-session` fails
- **THEN** `ReplaceFile` SHALL return a non-nil error
- **AND** SHALL NOT proceed to write the replacement file content

#### Scenario: Session file write succeeds

- **GIVEN** `ReplaceFile` is called on a non-Sequoia-managed, existing file
- **WHEN** `AtomicWriteFile` for `.sequoia-session` succeeds
- **THEN** `ReplaceFile` SHALL write the replacement content and return nil (no change from prior behavior)

#### Scenario: Backward compatibility — RestoreOrRemoveFile and all StrategyFileReplace adapters

- **GIVEN** the modified `ReplaceFile` with the error return
- **WHEN** any `StrategyFileReplace` adapter (Claude, Gemini, OpenCode, Cursor) calls `ReplaceFile`
- **THEN** the call site SHALL compile without signature changes (`ReplaceFile` already returns `error`)
- **AND** `RestoreOrRemoveFile` SHALL remain unchanged and compile without modification

## REMOVED Requirements

None
