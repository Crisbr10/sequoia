# Tasks: backup-retention-and-organization

## Task Tracker

> Checkboxes consumed by the gentle-ai dispatcher and by `sdd-apply` per-PR progress. The detailed task descriptions live in the section headings below (`### 1.1`–`### X.3`).

### PR 1 — Foundation (`backup_retention.go` + `BackupPathBuilder`)

- [x] **1.1** RED — `BackupHomeDir` path and 0o700 mode test
- [x] **1.2** GREEN — `BackupHomeDir` implementation
- [x] **1.3** RED — `PruneBackups` with 7-session fixture
- [x] **1.4** GREEN — `PruneBackups` implementation
- [x] **1.5** RED — `DefaultMaxBackupsPerAdapter == 5` constant test
- [x] **1.6** RED — `BackupPathBuilder.Build` uses central home
- [x] **1.7** GREEN — `BackupPathBuilder.Build` delegates to `BackupHomeDir`
- [x] **1.8** REFACTOR — path-prefix helper extraction

### PR 2 — Installer wiring (`BaseAdapter` + TUI note + test updates)

- [x] **2.1** RED — `BaseAdapter.Prepare` writes to central home
- [x] **2.2** GREEN — `BaseAdapter.Prepare` uses `BackupHomeDir()`
- [x] **2.3** GREEN — `BaseAdapter.Stage` passes central `BackupDir`
- [x] **2.4** GREEN — Update existing tests with old `.sequoia-backup` path assertions
- [x] **2.5** RED — TUI `Info` message notes pre-existing scattered backups
- [x] **2.6** GREEN — TUI `Info` adds migration note
- [x] **2.7** RED — `Codex` adapter `Install` backup goes to central home
- [x] **2.8** REFACTOR — `centralBackupDir` helper

### PR 3 — ReplaceFile + manifest + retention hook

- [x] **3.1** RED — `manifestEntry` JSON round-trip
- [x] **3.2** GREEN — `manifestEntry` and `manifest` types
- [x] **3.3** RED — `ReplaceFile` writes to central home with manifest
- [x] **3.4** GREEN — `ReplaceFile` writes to central home + manifest
- [x] **3.5** RED — `RestoreOrRemoveFile` reads from central home via manifest
- [x] **3.6** GREEN — `RestoreOrRemoveFile` reads manifest + restores
- [x] **3.7** RED — Retention hook: 6 installs → exactly 5 session dirs
- [x] **3.8** GREEN — `applyRetention` private method hooked in `Apply()`
- [x] **3.9** RED — Retention warning path on prune error
- [x] **3.10** REFACTOR — Consolidate manifest helpers
- [x] **3.11** GREEN — Per-adapter `paths.go` delegates to central home

### Cross-PR

- [x] **X.1** Package doc for `backup_retention.go`
- [x] **X.2** CHANGELOG entry
- [x] **X.3** `openspec/config.yaml` rules check

---

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~950 (PR1=250 + PR2=350 + PR3=400, cross-PR tasks ~50) |
| 400-line budget risk | PR3 is at the 400-line ceiling; all others within budget |
| Chained PRs recommended | Yes — per `force-chained` strategy and design's 3-PR stacked-to-main split |
| Suggested split | PR1=250 / PR2=350 / PR3=400 / cross=50 |
| Delivery strategy | `force-chained` (preflight cached) — auto-proceed with stacked-to-main |
| Chain strategy | stacked-to-main — each PR merges to main in order |

Decision needed before apply: No — `force-chained` is the user's cached choice; each PR is within the 400-line budget.
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: Medium (PR3 is right at 400; if any task estimate grows, split into PR3.5)

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Foundation: `BackupHomeDir` + `PruneBackups` + `BackupPathBuilder` wiring | PR 1 | Base branch = main; new files `backup_retention.go` + test; modify `backup_path_builder.go` |
| 2 | Installer wiring: `BaseAdapter.Prepare/Stage` central path + TUI note + 4+ test updates | PR 2 | Immediate parent = PR 1; modify `base_adapter.go`, `runner.go`; update 4 test files |
| 3 | ReplaceFile/manifest/retention hook + per-adapter `paths.go` updates | PR 3 | Immediate parent = PR 2; modify `strategy.go`, `base_adapter.go` (applyRetention); all 5 `paths.go` files |
| 4 | Cross-PR docs: package doc, CHANGELOG, config rules | PR 3 or cross | Lightweight; can land with PR 3 or as separate |

---

## Phase 1: PR 1 — Foundation

**Goal**: New `backup_retention.go` with `BackupHomeDir` + `PruneBackups` + `DefaultMaxBackupsPerAdapter`; `BackupPathBuilder` delegates to `BackupHomeDir`.

### 1.1 · RED — `BackupHomeDir` path and mode test

- **File**: `adapters/common/backup_retention_test.go` (new)
- **What fails**: `BackupHomeDir()` not defined
- **Verify plan**: `go test ./adapters/common/... -run TestBackupHomeDir -v -count=1` fails compile
- **Commit hint**: `common/backup_retention: add RED test for BackupHomeDir path and 0o700 mode`

### 1.2 · GREEN — `BackupHomeDir` implementation

- **File**: `adapters/common/backup_retention.go` (new)
- **What**: `const DefaultMaxBackupsPerAdapter = 5`; `func BackupHomeDir() (string, error)` — joins `os.UserConfigDir()` + `sequoia/backups/`, creates dir with `0o700`, wraps errors with path + suffix context
- **Verify**: `go test ./adapters/common/... -run TestBackupHomeDir -v -count=1` passes
- **Commit hint**: `common/backup_retention: implement BackupHomeDir with 0o700 creation`

### 1.3 · RED — `PruneBackups` with 7-session fixture

- **File**: `adapters/common/backup_retention_test.go` (append)
- **What fails**: `PruneBackups` not defined
- **Test**: fixture 7 session dirs with ISO-8601 timestamps; call `PruneBackups(adapterID, 5)`; assert exactly 2 removed, 5 kept (most recent by lexicographic sort)
- **Verify plan**: `go test ./adapters/common/... -run TestPruneBackups -v -count=1` fails compile
- **Commit hint**: `common/backup_retention: add RED test for PruneBackups 5-most-recent logic`

### 1.4 · GREEN — `PruneBackups` implementation

- **File**: `adapters/common/backup_retention.go` (append)
- **What**: `func PruneBackups(adapterID string, max int) (removed int, err error)` — reads `<root>/<adapterID>/`, filters to valid ISO-8601 timestamp dirs, sorts lexicographically descending, removes oldest until count ≤ max; continues on per-entry error; returns `(0, nil)` on miss
- **Verify**: `go test ./adapters/common/... -run TestPruneBackups -v -count=1` passes; coverage ≥70% on `backup_retention.go`
- **Commit hint**: `common/backup_retention: implement PruneBackups with best-effort removal`

### 1.5 · RED — `DefaultMaxBackupsPerAdapter == 5` constant test

- **File**: `adapters/common/backup_retention_test.go` (append)
- **What fails**: constant not defined
- **Test**: `require.Equal(t, 5, DefaultMaxBackupsPerAdapter)`
- **Verify plan**: `go test ./adapters/common/... -run TestDefaultMaxBackups -v -count=1` fails compile
- **Commit hint**: `common/backup_retention: add RED test for DefaultMaxBackupsPerAdapter constant`

### 1.6 · RED — `BackupPathBuilder.Build` uses central home

- **File**: `adapters/common/backup_path_builder_test.go` (update existing)
- **What fails**: `BackupPathBuilder.Build` still returns per-tool path
- **Test update**: assert the result path starts with `BackupHomeDir()` result and contains `<adapterID>/<timestamp>-<suffix>/`
- **Verify plan**: `go test ./adapters/common/... -run TestBackupPathBuilder -v -count=1` fails assertion
- **Commit hint**: `common/backup_path_builder: add RED test asserting central home prefix`

### 1.7 · GREEN — `BackupPathBuilder.Build` delegates to `BackupHomeDir`

- **File**: `adapters/common/backup_path_builder.go` (modify)
- **What**: `BackupPathBuilder.Build` calls `BackupHomeDir()` first, then appends `/{adapterID}/{timestamp}-{suffix}/`
- **Verify**: `go test ./adapters/common/... -run TestBackupPathBuilder -v -count=1` passes
- **Commit hint**: `common/backup_path_builder: delegate to BackupHomeDir for root`

### 1.8 · REFACTOR — path-prefix helper extraction

- **File**: `adapters/common/backup_retention.go` (optional helper)
- **What**: extract any duplicated path-prefix logic into a private `backupRoot()` helper; lint clean (`golangci-lint run ./adapters/common/...`)
- **Verify**: `golangci-lint run ./adapters/common/...` passes; `go vet ./adapters/common/...` clean
- **Commit hint**: `common/backup_retention: extract backupRoot helper; lint clean`

**PR 1 Verify**: `go test ./adapters/common/... -run 'Backup|Retention' -v -count=1 -coverprofile=pr1.out; go tool cover -func=pr1.out | grep backup_retention` — coverage ≥70% on new files.

---

## Phase 2: PR 2 — Installer Wiring

**Goal**: `BaseAdapter.Prepare/Stage` use `BackupHomeDir()` for `BackupDir`; TUI `Info` notes legacy backups; 4+ test files updated.

### 2.1 · RED — `BaseAdapter.Prepare` writes to central home

- **File**: `adapters/common/base_adapter_test.go` (new test or update)
- **What fails**: `BackupDir` still per-tool
- **Test**: assert the `BackupDir` argument passed to `NewInstaller` starts with `BackupHomeDir()` result
- **Verify plan**: `go test ./adapters/common/... -run TestPrepareBackupDir -v -count=1` fails assertion
- **Commit hint**: `common/base_adapter: add RED test for Prepare BackupDir central home`

### 2.2 · GREEN — `BaseAdapter.Prepare` uses `BackupHomeDir()`

- **File**: `adapters/common/base_adapter.go` (modify, ~line 325)
- **What**: `BaseAdapter.Prepare()` computes `BackupDir` from `BackupHomeDir()` instead of `a.backup.Build(base)`; pass to `NewInstaller`
- **Verify**: `go test ./adapters/common/... -run TestPrepareBackupDir -v -count=1` passes
- **Commit hint**: `common/base_adapter: compute BackupDir from BackupHomeDir in Prepare`

### 2.3 · GREEN — `BaseAdapter.Stage` passes central `BackupDir`

- **File**: `adapters/common/base_adapter.go` (modify, ~lines 452, 468)
- **What**: `Stage()` passes the same central-home `BackupDir` to skill and command installers
- **Verify**: `go test ./adapters/common/... -run TestStage -v -count=1` passes
- **Commit hint**: `common/base_adapter: pass central BackupDir to installers in Stage`

### 2.4 · GREEN — Update existing tests with old `.sequoia-backup` path assertions

- **Files** (scan + update each):
  - `adapters/common/base_adapter_test.go`
  - `adapters/common/installer_test.go`
  - `adapters/codex/installer_test.go`
  - `adapters/opencode/install_test.go`
  - `adapters/common/base_adapter_strategy_test.go`
  - `adapters/common/base_adapter_error_test.go`
- **What**: grep for `.sequoia-backup` patterns; replace with central-home assertions; ensure each test passes against new path
- **Verify**: `go test ./adapters/... -count=1` all pass
- **Commit hint**: `tests: update path assertions from per-tool .sequoia-backup to central home`

### 2.5 · RED — TUI `Info` message notes pre-existing scattered backups

- **File**: `internal/pipeline/runner_test.go` (new test or update)
- **What fails**: `Info` message does not mention legacy backups
- **Test**: when adapter implements `BackupDirGetter`, assert the `Info` string includes a one-line note about pre-existing scattered backups
- **Verify plan**: `go test ./internal/pipeline/... -run TestInfoMigrationNote -v -count=1` fails
- **Commit hint**: `pipeline/runner: add RED test for scattered-backup migration note in Info`

### 2.6 · GREEN — TUI `Info` adds migration note

- **File**: `internal/pipeline/runner.go` (modify, ~lines 199-210)
- **What**: in `sendProgress` block, when `getter.LastBackupDir() != ""`, append one-line note: "Note: pre-existing scattered backups from prior sequoia versions remain at their original locations."
- **Verify**: `go test ./internal/pipeline/... -run TestInfoMigrationNote -v -count=1` passes
- **Commit hint**: `pipeline/runner: surface legacy backup note in Info message`

### 2.7 · RED — `Codex` adapter `Install` backup goes to central home

- **File**: `adapters/codex/installer_test.go` (update)
- **What fails**: `Codex.Install` backup still per-tool
- **Test**: assert backup path starts with `BackupHomeDir()` result
- **Verify plan**: `go test ./adapters/codex/... -run TestCodexInstall -v -count=1` fails
- **Commit hint**: `codex/installer: add RED test for central home backup path`

### 2.8 · REFACTOR — `centralBackupDir` helper

- **File**: `adapters/common/base_adapter.go` (modify)
- **What**: extract `BackupDir` construction into `func (a *BaseAdapter) centralBackupDir(targetSubdir string) string` private helper; update `Prepare()` and `Stage()` call sites; lint clean
- **Verify**: `go test ./adapters/... ./internal/pipeline/... -count=1` all pass; `golangci-lint run ./adapters/common/...`
- **Commit hint**: `common/base_adapter: extract centralBackupDir helper; lint clean`

**PR 2 Verify**: `go test ./adapters/... ./internal/pipeline/... -count=1` — full project test run passes.

---

## Phase 3: PR 3 — ReplaceFile + Manifest + Retention Hook

**Goal**: `ReplaceFile`/`RestoreOrRemoveFile` use central home + `manifest.json`; `applyRetention` hooked in `BaseAdapter.Apply()`; all 5 per-adapter `paths.go` updated.

### 3.1 · RED — `manifestEntry` JSON round-trip

- **File**: `adapters/common/backup_retention_test.go` (append) or new `adapters/common/manifest_test.go`
- **What fails**: `manifestEntry` not defined
- **Test**: `encoding/json` round-trip; assert fields `Version`, `OriginalPath`, `Suffix`, `CreatedAt`, `AdapterID` serialize/deserialize correctly
- **Verify plan**: `go test ./adapters/common/... -run TestManifestEntry -v -count=1` fails compile
- **Commit hint**: `common/manifest: add RED test for manifestEntry JSON round-trip`

### 3.2 · GREEN — `manifestEntry` and `manifest` types

- **File**: `adapters/common/backup_retention.go` (append) or new `adapters/common/manifest.go`
- **What**: define `type manifestEntry struct { ... }` and `type manifest struct { Version string; Entries []manifestEntry }` with `encoding/json` tags; add `appendManifestEntry`, `readManifest`, `writeManifest` private helpers
- **Verify**: `go test ./adapters/common/... -run TestManifestEntry -v -count=1` passes
- **Commit hint**: `common/manifest: define manifestEntry/manifest types and helpers`

### 3.3 · RED — `ReplaceFile` writes to central home with manifest

- **File**: `adapters/common/strategy_test.go` (update)
- **What fails**: `ReplaceFile` still writes to per-tool path
- **Test**: assert backup at `{BackupHomeDir}/{adapterID}/{sessionDir}/<basename>.backup` and `manifest.json` exists with correct `original_path`
- **Verify plan**: `go test ./adapters/common/... -run TestReplaceFileCentral -v -count=1` fails
- **Commit hint**: `common/strategy: add RED test for ReplaceFile central home + manifest`

### 3.4 · GREEN — `ReplaceFile` writes to central home + manifest

- **File**: `adapters/common/strategy.go` (modify)
- **What**: `ReplaceFile` writes to `<root>/<adapterID>/<sessionDir>/<basename>.backup`; creates/updates `manifest.json` in the session dir; uses `appendManifestEntry`
- **Verify**: `go test ./adapters/common/... -run TestReplaceFileCentral -v -count=1` passes
- **Commit hint**: `common/strategy: wire ReplaceFile to central home and manifest.json`

### 3.5 · RED — `RestoreOrRemoveFile` reads from central home via manifest

- **File**: `adapters/common/strategy_test.go` (append)
- **What fails**: `RestoreOrRemoveFile` still reads from per-tool path
- **Test**: set up session dir with manifest; call `RestoreOrRemoveFile`; assert original file restored byte-for-byte and session dir removed
- **Verify plan**: `go test ./adapters/common/... -run TestRestoreOrRemoveCentral -v -count=1` fails
- **Commit hint**: `common/strategy: add RED test for RestoreOrRemoveFile manifest-based restore`

### 3.6 · GREEN — `RestoreOrRemoveFile` reads manifest + restores

- **File**: `adapters/common/strategy.go` (modify)
- **What**: `RestoreOrRemoveFile` reads `manifest.json` from session dir; locates backup by `original_path`; restores file; removes session dir on success
- **Verify**: `go test ./adapters/common/... -run TestRestoreOrRemoveCentral -v -count=1` passes
- **Commit hint**: `common/strategy: implement manifest-based RestoreOrRemoveFile`

### 3.7 · RED — Retention hook: 6 installs → exactly 5 session dirs

- **File**: `adapters/common/base_adapter_test.go` (new test)
- **What fails**: `applyRetention` not defined; retention not hooked
- **Test**: pre-seed 7 session dirs in `<root>/<adapter>/`; call `BaseAdapter.Apply()` (or a test wrapper that calls `applyRetention`); assert ≤ 5 session dirs remain; the 2 oldest removed
- **Verify plan**: `go test ./adapters/common/... -run TestRetentionEnforcement -v -count=1` fails
- **Commit hint**: `common/base_adapter: add RED test for 5-backup retention enforcement`

### 3.8 · GREEN — `applyRetention` private method hooked in `Apply()`

- **File**: `adapters/common/base_adapter.go` (modify, ~line 519)
- **What**: add `func (a *BaseAdapter) applyRetention()` calling `PruneBackups(a.ID(), DefaultMaxBackupsPerAdapter)`; on error call `a.AddWarning(fmt.Sprintf("backup retention: %v", err))`; call `applyRetention()` just before `return nil` in `Apply()`
- **Verify**: `go test ./adapters/common/... -run TestRetentionEnforcement -v -count=1` passes
- **Commit hint**: `common/base_adapter: hook applyRetention before Apply return nil`

### 3.9 · RED — Retention warning path on prune error

- **File**: `adapters/common/base_adapter_test.go` (append)
- **What fails**: warnings not emitted on prune failure
- **Test**: make a session dir read-only; run retention; assert `a.Warnings()` includes a "backup retention:" message
- **Verify plan**: `go test ./adapters/common/... -run TestRetentionWarning -v -count=1` fails
- **Commit hint**: `common/base_adapter: add RED test for retention warning on prune error`

### 3.10 · REFACTOR — Consolidate manifest helpers

- **File**: `adapters/common/backup_retention.go` or `manifest.go`
- **What**: move all manifest read/write helpers to a single file; ensure `encoding/json` error wrapping is consistent; lint clean
- **Verify**: `golangci-lint run ./adapters/common/...` passes; `go vet ./adapters/common/...` clean
- **Commit hint**: `common/manifest: consolidate helpers; lint clean`

### 3.11 · GREEN — Per-adapter `paths.go` delegates to central home

- **Files** (5 adapters):
  - `adapters/claude/paths.go`
  - `adapters/codex/paths.go`
  - `adapters/cursor/paths.go`
  - `adapters/gemini/paths.go`
  - `adapters/opencode/paths.go`
- **What**: each `backupPath()` method delegates to `BackupHomeDir()` (via `common.BackupHomeDir()` or by importing the common package) — but note: design says these files are modified in PR3; confirm whether they import `common` or need a different approach (e.g., the `BackupPathBuilder` already handles the routing, so these may just need test updates, not code changes — verify by grepping the existing `backupPath()` implementations first)
- **Verify**: `go test ./adapters/... -count=1` all pass
- **Commit hint**: `adapters/<tool>: update backupPath to delegate to BackupHomeDir`

**PR 3 Verify**: `go test ./... -coverprofile=coverage.out -count=1 -timeout 120s` — coverage ≥70% project-wide (CI gate).

---

## Phase 4: Cross-PR Documentation

### X.1 · Package doc for `backup_retention.go`

- **File**: `adapters/common/backup_retention.go` (or `manifest.go`)
- **What**: add/modify package doc-comment summarizing the retention policy (5 per adapter, central home) and the manifest schema
- **Verify**: `go doc adapters/common` shows updated doc
- **Commit hint**: `common/backup_retention: document retention policy and manifest schema`

### X.2 · CHANGELOG entry

- **File**: `CHANGELOG.md` (or `CHANGELOG.md` equivalent in project root)
- **What**: add entry noting new backup location (`~/.config/sequoia/backups/`), 5-backup retention per adapter, and TUI migration note for pre-existing scattered backups
- **Verify**: file updated; no format check needed
- **Commit hint**: `docs: log new central backup location and 5-backup retention`

### X.3 · `openspec/config.yaml` rules check

- **File**: `openspec/config.yaml`
- **What**: confirm `rules.tasks` and `testing.strict_tdd: true` are still correct; no changes needed unless project rules have drifted
- **Verify**: config.yaml re-read; no modifications needed
- **Commit hint**: `none: config.yaml rules review — no changes needed`

---

## Implementation Order

1. **PR 1 first** — `backup_retention.go` + `BackupPathBuilder` wiring is the foundation everything else depends on. Tests and code are fully isolated to `adapters/common/`.
2. **PR 2 second** — depends on PR 1 (`BackupHomeDir` must exist). `BaseAdapter` wiring and TUI note are the integration layer.
3. **PR 3 third** — depends on PR 2. `ReplaceFile`/manifest/retention hook is the behavioral core; all 5 `paths.go` files updated here.
4. **Cross-PR docs** — can land with PR 3 or as a lightweight follow-up; no code dependencies.

---

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| PR3 at 400-line ceiling — any task overage forces a 4th PR | Medium | Monitor diff size during apply; if over, split manifest helpers as PR3.5 |
| Windows path handling in `BackupHomeDir` | Medium | `os.UserConfigDir()` is platform-correct; test on Windows CI matrix |
| Corrupt `manifest.json` during restore | Low | Best-effort scan for `*.backup` files; warn via `AddWarning` — deferred to sdd-apply |
| Concurrent installs race on retention | Low | Acceptable for v1 single-user sequential CLI |
| 4+ existing tests need path assertion updates | High | T2.4 is explicit; grep confirmed target files |
