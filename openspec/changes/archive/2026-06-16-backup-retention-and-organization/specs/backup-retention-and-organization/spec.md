# Delta Spec: backup-retention-and-organization

> Capability: `backup-retention-and-organization` (NEW — first spec for the project).
> Post-archive main spec: `openspec/specs/backup-retention-and-organization/spec.md`.
> Existing main specs: none. This delta contains only `## ADDED Requirements`.

## ADDED Requirements

## REQ-BRP-01 — Centralized backup root

The system SHALL write every backup it creates under a single root directory resolved by joining `os.UserConfigDir()` with `sequoia/backups/`. The root SHALL be created with file mode `0o700` on first use if it does not already exist. Linux/macOS examples resolve to `~/.config/sequoia/backups/`; Windows examples resolve to `%APPDATA%\sequoia\backups\`.

#### Scenario: BackupHomeDir returns and creates the joined path

- GIVEN `os.UserConfigDir()` returns `/home/alice/.config` and the root does not exist
- WHEN `common.BackupHomeDir()` is called
- THEN the returned absolute path equals `/home/alice/.config/sequoia/backups`
- AND the directory exists on disk with permission `0o700`

#### Scenario: BackupHomeDir is idempotent on a pre-existing root

- GIVEN `<root>` already exists with mode `0o700`
- WHEN `common.BackupHomeDir()` is called again
- THEN the same absolute path is returned
- AND no error is returned
- AND the existing mode is preserved

## REQ-BRP-02 — Per-adapter organization

Under the centralized root, the system SHALL create one subdirectory per adapter named exactly by the adapter ID (`claude-code`, `opencode`, `codex`, `gemini-cli`, `cursor`). Each install session's backup SHALL be written into a subdirectory of the adapter folder, named with an ISO-8601 UTC timestamp (`YYYY-MM-DDTHH-MM-SS-mmmZ`) suffixed with the existing base-36 Unix-nanos session ID. The existing `skills/` and `commands/` subdirectory structure MUST be preserved inside every session directory.

#### Scenario: Session directory name combines timestamp and session suffix

- GIVEN adapter ID `claude-code` and the clock at `2026-06-15T15:30:45.123Z`
- WHEN `BackupPathBuilder.Build(base)` is called
- THEN the result equals `<root>/claude-code/2026-06-15T15-30-45-123Z-<sessionSuffix>`
- AND `<sessionSuffix>` is a base-36 Unix-nanos value

#### Scenario: Distinct adapter IDs produce disjoint subtrees

- GIVEN `Build(base)` is called for `claude-code` and for `opencode` with identical timestamp and suffix
- WHEN the two results are compared
- THEN the adapter segment differs (`claude-code` vs `opencode`)
- AND neither path is a prefix of the other

#### Scenario: skills/ and commands/ substructure is preserved inside a session

- GIVEN a freshly built session directory `<root>/claude-code/2026-...-<suffix>/`
- WHEN the install stages a `skills/foo.md` and a `commands/bar.md`
- THEN both files exist at `<root>/.../skills/foo.md` and `<root>/.../commands/bar.md`
- AND no other top-level entries exist in the session directory

## REQ-BRP-03 — File-replace backup storage

The `ReplaceFile` strategy SHALL write backed-up files inside the per-adapter session directory as `<original-relative-filename>.backup` (e.g., `AGENTS.md.backup`). The current `.sequoia-session` sidecar that tracks the suffix SHALL be replaced by a manifest file inside the session directory recording each original path together with the suffix used. `RestoreOrRemoveFile` SHALL read that manifest to locate the correct backup.

#### Scenario: ReplaceFile writes the backup to the session directory

- GIVEN `ReplaceFile` is invoked for `~/.config/opencode/AGENTS.md` with suffix `k2j9m4`
- WHEN the strategy completes
- THEN a byte-equal copy of the original file exists at `<root>/opencode/2026-...-k2j9m4/AGENTS.md.backup`
- AND a manifest inside the session directory lists `original_path=~/.config/opencode/AGENTS.md` and `suffix=k2j9m4`

#### Scenario: RestoreOrRemoveFile reads from the session directory via the manifest

- GIVEN a manifest in `<root>/claude-code/2026-...-<suffix>/manifest.json`
- WHEN `RestoreOrRemoveFile` is called for the same target path
- THEN it locates the backup by parsing the manifest
- AND the original file is restored byte-for-byte
- AND the session directory is removed on success

## REQ-BRP-04 — Retention policy of 5 per adapter

The system SHALL retain at most `common.DefaultMaxBackupsPerAdapter` (exported constant, value `5`) backup session directories per adapter. Retention SHALL run at the end of a successful `BaseAdapter.Install()`, specifically after `Apply()` completes without error. The pruner SHALL sort session directories by their timestamp prefix (lexicographic order matches chronological order for the ISO-8601 prefix), keep the `max` most recent, and remove the rest. Removal errors SHALL be reported as warnings through the existing `AddWarning` mechanism and SHALL NOT cause the install to fail.

#### Scenario: Pruning keeps exactly five backups after a sixth install

- GIVEN adapter `opencode` already has 5 session directories
- WHEN a successful install creates a 6th session directory
- THEN `PruneBackups("opencode", 5)` runs before `Install` returns
- AND the count of session directories equals exactly 5
- AND the just-created (newest) session is preserved

#### Scenario: Pruning below the threshold is a no-op

- GIVEN adapter `cursor` has 3 session directories
- WHEN `PruneBackups("cursor", 5)` is called
- THEN the returned `removed` count is `0`
- AND the 3 directories are untouched

#### Scenario: Removal errors do not fail the install

- GIVEN one of the candidate directories cannot be removed (e.g., permission denied)
- WHEN retention runs
- THEN `Install()` still returns success
- AND a warning is recorded via `AddWarning` describing the failed removal

#### Scenario: Retention is skipped when the install fails

- GIVEN `Apply()` returns an error during `BaseAdapter.Install()`
- WHEN the install path finishes
- THEN `PruneBackups` is NOT called for any adapter
- AND no directory is removed

## REQ-BRP-05 — Migration of old scattered backups (NOT performed)

Pre-existing `.sequoia-backup-*` files and `.sequoia-backup-*-<id>/` directories left in tool config directories (`~/.config/opencode/`, `~/.claude/`, etc.) from earlier sequoia versions SHALL NOT be touched, relocated, or deleted by this change. The TUI `Info` message emitted through the `adapters.BackupDirGetter` path in `internal/pipeline/runner.go` (lines 199-210) SHALL additionally include a one-line note that any pre-existing scattered backups from prior sequoia versions remain in place at their original locations.

#### Scenario: Old scattered backups remain untouched after a new install

- GIVEN a pre-existing `~/.claude/AGENTS.md.sequoia-backup-old123` from a prior sequoia version
- WHEN a new install is run
- THEN that file still exists at the original absolute path with its original mtime and size
- AND no copy of it appears under the new central root

#### Scenario: TUI Info message notes pre-existing scattered backups

- GIVEN an adapter implementing `adapters.BackupDirGetter` returning the new central path
- WHEN the pipeline emits the `Info` message for that adapter
- THEN the message text includes the central backup path
- AND the message text includes a one-line note that pre-existing scattered backups remain at their original locations

## REQ-BRP-06 — Path resolution and pruning helpers

The system SHALL expose two exported helpers in `adapters/common`:
- `BackupHomeDir() (string, error)` — returns the absolute central root, creating it with mode `0o700` when missing; errors SHALL be wrapped with context that includes both the failing path and the `sequoia/backups` suffix.
- `PruneBackups(adapterID string, max int) (removed int, err error)` — removes the oldest session directories for `adapterID`, keeping the `max` most recent by timestamp; returns the number removed and the first error encountered, continuing to attempt removal of subsequent entries on error.

#### Scenario: BackupHomeDir wraps errors with context

- GIVEN `os.UserConfigDir()` succeeds but the joined path cannot be created (e.g., read-only parent)
- WHEN `BackupHomeDir()` is called
- THEN the returned error is non-nil
- AND the error message contains the failing path and the substring `sequoia/backups`

#### Scenario: PruneBackups continues removing after a single failure

- GIVEN adapter `x` has 7 session directories and directory #4 is read-only
- WHEN `PruneBackups("x", 5)` is called
- THEN the two non-read-only oldest directories are removed
- AND the returned `removed` count is `2`
- AND the returned error is non-nil and references directory #4

## REQ-BRP-07 — Test surface (strict TDD)

All new code paths SHALL be covered by unit tests, and the existing 4+ tests that reference the old per-tool `backupPath()` location SHALL be updated to assert against the new centralized location. Required coverage includes: `BackupHomeDir` creates and returns the expected path with mode `0o700`; `PruneBackups` keeps exactly `max` most recent when more exist; `PruneBackups` is a no-op at or below `max`; `PruneBackups` handles a missing adapter directory gracefully; `PruneBackups` ignores corrupt directory names without crashing; centralized `ReplaceFile`/`RestoreOrRemoveFile` round-trip; the end-to-end install flow leaves at most 5 session directories per adapter.

#### Scenario: BackupHomeDir unit test asserts path and mode

- GIVEN a test override for `os.UserConfigDir`
- WHEN the `BackupHomeDir` unit test runs
- THEN it asserts the returned path equals the expected joined string
- AND it asserts the directory was created with mode `0o700`

#### Scenario: PruneBackups handles a missing adapter directory

- GIVEN no directory exists at `<root>/nonexistent-adapter/`
- WHEN `PruneBackups("nonexistent-adapter", 5)` is called
- THEN it returns `(0, nil)` and creates no directories

#### Scenario: PruneBackups ignores corrupt directory names

- GIVEN `<root>/x/` contains a non-timestamp directory `garbage` plus 7 valid session dirs
- WHEN `PruneBackups("x", 5)` is called
- THEN the corrupt directory is left untouched
- AND exactly 2 valid directories are removed
- AND no panic or error is returned

#### Scenario: End-to-end install respects the 5-backup cap

- GIVEN the same adapter is installed 6 times in succession without errors
- WHEN each `Install()` returns
- THEN `<root>/<adapter>/` contains exactly 5 session directories
- AND the oldest session is the one that was removed

#### Scenario: Existing tests are updated to the central path

- GIVEN the 4+ tests that previously asserted against the old per-tool `backupPath()` location
- WHEN the test suite is run after the change
- THEN those tests pass against the new centralized path
- AND no test references the old per-tool `.sequoia-backup/` location
