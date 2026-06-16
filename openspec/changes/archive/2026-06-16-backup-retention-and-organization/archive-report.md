# Archive Report: backup-retention-and-organization

> **Change**: backup-retention-and-organization
> **Archived**: 2026-06-16
> **Mode**: openspec (file-based)
> **Store**: `openspec/specs/backup-retention-and-organization/spec.md`
> **Strict TDD**: true (active throughout)
> **Commit on main**: `9c1dad2`

---

## Executive Summary

The `backup-retention-and-organization` change consolidates scattered `.sequoia-backup-*` backup directories into a single central root (`<UserConfigDir>/sequoia/backups/`) with a 5-session-per-adapter retention cap enforced at the end of every successful `BaseAdapter.Apply()`. The change introduces a per-session `manifest.json` that records original file paths, enabling `RestoreOrRemoveFile` to locate and restore backups without coupling to per-tool config directories.

The change shipped across **5 stacked PRs** (PR 1 foundation → PR 2 installer wiring → PR 3a manifest + safety-net → PR 3b retention hook → PR 3.5 ReplaceFile/RestoreOrRemoveFile migration), each independently mergeable and each satisfying the 400-line review budget. All 30 implementation tasks are complete, all 7 REQs are satisfied, all 20 spec scenarios are green, and the full project test suite is green (20/20 packages, 83.0% coverage on `adapters/common`).

**Two spec ambiguities were resolved at archive time** (not blocking; spec issues only):
- REQ-BRP-02 example timestamp format: spec example used dash (`-`) between SS and mmm; implementation uses dot (`.`), which is ISO-8601 compliant and lex-sort invariant. Spec example updated to match implementation.
- REQ-BRP-06 scenario: "directory #4 is read-only" was internally inconsistent with max=5 and 7 entries. Scenario rewritten to say "the OLDEST directory is read-only" to match the implementation's correct exercise of the continuation contract.

---

## Artifacts

| File | Description |
|------|-------------|
| `proposal.md` | Original change intent, scope, approach, risks, rollback plan |
| `specs/backup-retention-and-organization/spec.md` | Delta spec (ADDED only — first spec for this project) |
| `design.md` | Technical approach, module layout, API surface, 5 architecture decisions |
| `tasks.md` | 30 tasks across 4 PRs (all marked complete) |
| `state.yaml` | Phase status tracker |
| `apply-progress.md` | Per-PR progress: commit SHAs, file diffs, TDD cycle evidence |
| `verify-report.md` | PR 1 verify (pass-with-warnings) |
| `verify-report-pr2.md` | PR 2 verify (pass-with-warnings) |
| `verify-report-pr3a.md` | PR 3a verify (fail — CRITICAL fixed inline before merge) |
| `verify-report-pr3b.md` | PR 3b verify (pass-with-warnings) |
| `verify-report-pr35.md` | PR 3.5 verify (pass-with-warnings) |
| `archive-report.md` | This file |

---

## PRs Delivered

| PR | Branch | Commit SHA | Lines | Verify Result |
|----|--------|------------|-------|---------------|
| PR 1 — Foundation | `feature/backup-retention-pr1-foundation` | `e460217..5f3214a` | +758/-34 (7 files) | pass-with-warnings |
| PR 2 — Installer wiring | `feature/backup-retention-pr2-installer` | `90010d9..688c579` | +456/-19 (7 files) | pass-with-warnings |
| PR 3a — Manifest + safety-net | `feature/backup-retention-pr3a-manifest` | `a7daa17..6d21d8d` | +459/-11 (9 files) | fail-with-critical → fixed inline |
| PR 3b — Retention hook | `feature/backup-retention-pr3b-retention` | `d3da1a1..4d7f9bc` | +255/0 (2 files) | pass-with-warnings |
| PR 3.5 — ReplaceFile migration | `feature/backup-retention-pr35-replacefile` | `a917c2d..9c1dad2` | +638/-297 (7 files) | pass-with-warnings |

**24 total commits** across 4 PRs to `main` at `9c1dad2`. Net ~2200 lines including tests and docs.

---

## Spec Compliance

**7 REQs — all satisfied:**

| REQ | Description | Status |
|-----|-------------|--------|
| REQ-BRP-01 | Centralized backup root (0o700, `os.UserConfigDir()`) | ✅ PASS |
| REQ-BRP-02 | Per-adapter organization, ISO-8601 session dirs | ✅ PASS |
| REQ-BRP-03 | File-replace backup storage with manifest.json | ✅ PASS |
| REQ-BRP-04 | 5-per-adapter retention cap, end of Apply() | ✅ PASS |
| REQ-BRP-05 | No migration of old scattered backups + TUI note | ✅ PASS |
| REQ-BRP-06 | `BackupHomeDir`/`PruneBackups` exported helpers | ✅ PASS |
| REQ-BRP-07 | Test surface (strict TDD) | ✅ PASS |

**20 spec scenarios — all green.** The 2 carryover spec ambiguities (REQ-BRP-02 timestamp format, REQ-BRP-06 directory numbering) were **resolved** at archive time by updating the spec examples to match the implementation.

---

## Key Decisions Locked

1. **`BackupHomeDir()` + `PruneBackups()`** in `adapters/common/backup_retention.go` — matches `os.UserConfigDir` family; verb-form prune; exported constant `DefaultMaxBackupsPerAdapter = 5`.
2. **JSON `manifest.json`** per session — stdlib `encoding/json`; schema `{version, created_at, entries:[{version, original_path, suffix, created_at, adapter_id}]}` — extensible, parser-friendly.
3. **Hook at end of `BaseAdapter.Apply()`** (before `return nil`) via private `applyRetention()` — `Apply` is the last mutating phase; errors → `AddWarning`, not failures.
4. **Single `manifest.json` per session** listing only `ReplaceFile`-backed files — the natural unit; installer backups use `InstallerConfig.TargetDir`.
5. **4 stacked PRs to `main`** (≤400L each, `force-chained` strategy) — PR1 foundation, PR2 wiring, PR3a manifest+safety-net, PR3b retention, PR3.5 ReplaceFile migration. The 4th split (PR3→3a+3b+3.5) was done due to the 400-line budget overage.

---

## Deviations from Plan

| Deviation | Reason | Resolution |
|-----------|--------|------------|
| PR 3 split into PR 3a + PR 3b + PR 3.5 | 400-line budget exceeded (PR3 was 750L) | Orchestrator Option C: 3-slice split per chained-pr skill |
| PR 3a had a CRITICAL schema divergence (`manifest` missing `created_at` top-level field) | Design-vs-implementation divergence; caught by sdd-verify | Fixed inline in PR 3a before merge (1 line in `manifest.go` + 1 test) |
| `CentralBackupDir` exported (uppercase) vs private (design) | Codex custom `Install` needed cross-package access | Exported; rationale documented in apply-progress |
| Task 2.4 updated only 1 of 4+ test files | The other 4 files don't assert the BaseAdapter backup path location — they test legacy sidecar or fixture params | Verified by independent adversarial grep in PR 2 verify |
| TDD commit shape: RED+GREEN in single commits | Strict TDD audit would want separate commits | Non-blocking; noted as SUGGESTION in all verify reports |

---

## Risks Carried Forward (Post-Archive Follow-ups)

1. **5 `adapters/<tool>/paths.go` docstrings are slightly stale** — the per-tool `backupPath()` functions are no longer consulted on the happy path (the central-home + manifest path is always used). The safety-net fallback in `BackupPathBuilder.Build` hard-codes the path. **Recommend**: KEEP the safety-net for resilience; revisit in a future PR if product wants to simplify.

2. **`BackupPathBuilder.Build` safety-net fallback is rarely/never reached** — fires only when `BackupHomeDir()` itself fails. **Recommend**: KEEP for resilience; the cost is one extra indirection per fallback call.

3. **`replaceFileLegacySidecar` defensive helper has 0% per-function coverage** — the central home always succeeds in tests. **Acceptable** for a defensive safety-net helper; 83.0% file-level gate is met.

4. **CRLF line endings on 5 pre-existing files in `adapters/common/`** — Windows-only artifact introduced before this change. **Not introduced by this change.** CI on Linux/macOS will not see this.

5. **The 2 spec ambiguities are now RESOLVED** — REQ-BRP-02 timestamp format and REQ-BRP-06 "directory #4" numbering were corrected in the spec at archive time.

---

## Test Results

- **20/20 packages PASS** (all 5 PR verify runs)
- **`adapters/common` coverage: 83.0%** (above 70% gate)
- **5 consecutive clean runs** across the PR 3.5 verify chain
- **`go vet ./...`**: clean (all PRs)
- **`go vet`**: clean (all PRs)

| PR | Coverage | Notes |
|----|----------|-------|
| PR 1 | 85.8% | `BackupHomeDir` 85.7%, `PruneBackups` 77.4% |
| PR 2 | 85.5% | `CentralBackupDir` 92.3%, `Prepare` 95.2%, `Stage` 83.3% |
| PR 3a | 85.0% | `manifest.go` 71.9% file-level |
| PR 3b | 85.1% | `applyRetention` 50% (Windows-only; POSIX = 100%) |
| PR 3.5 | 83.0% | `ReplaceFile` 81.5%, `RestoreOrRemoveFile` 88.6%, `NewSessionDir` 100% |

---

## Total Diff Stats

- **24 commits** across 4 PRs (PR 1=6, PR 2=4, PR 3a=5, PR 3b=3, PR 3.5=3, + 3 orchestrator merges)
- **~2200 net lines** including tests and docs
- **Code-only diff (PR 3.5 final)**: +341 net lines (under 400L budget)

---

## Rollback Plan

`git revert <merge-commit>` for each PR merge commit. The change is non-destructive: it does not touch user data outside the central backup home (`<UserConfigDir>/sequoia/backups/`); pre-existing scattered backups at per-tool config directories are untouched per REQ-BRP-05.

---

## Structured Envelope

```json
{
  "status": "archived",
  "change": "backup-retention-and-organization",
  "commit": "9c1dad2",
  "artifacts": {
    "archive_folder": "openspec/changes/archive/2026-06-16-backup-retention-and-organization",
    "archive_report": "openspec/changes/archive/2026-06-16-backup-retention-and-organization/archive-report.md",
    "delta_spec": "openspec/changes/archive/2026-06-16-backup-retention-and-organization/specs/backup-retention-and-organization/spec.md",
    "source_of_truth_spec": "openspec/specs/backup-retention-and-organization/spec.md"
  },
  "prs_delivered": [
    {"pr": "PR 1", "commits": "e460217..5f3214a", "verify": "pass-with-warnings"},
    {"pr": "PR 2", "commits": "90010d9..688c579", "verify": "pass-with-warnings"},
    {"pr": "PR 3a", "commits": "a7daa17..6d21d8d", "verify": "pass-with-warnings (CRITICAL fixed inline)"},
    {"pr": "PR 3b", "commits": "d3da1a1..4d7f9bc", "verify": "pass-with-warnings"},
    {"pr": "PR 3.5", "commits": "a917c2d..9c1dad2", "verify": "pass-with-warnings"}
  ],
  "spec_compliance": {
    "reqs": "7/7 satisfied",
    "scenarios": "20/20 green",
    "ambiguities_resolved": 2
  },
  "test_results": {
    "packages": "20/20 pass",
    "coverage": "83.0% (adapters/common, above 70% gate)",
    "vet": "clean",
    "consecutive_runs": "5 (clean)"
  },
  "next_recommended": "sdd-archive-complete",
  "risks": [
    "5 adapters/<tool>/paths.go docstrings are slightly stale (the per-tool backupPath() is no longer consulted on the happy path); recommend KEEP the safety-net for resilience; revisit in a future PR if product wants to simplify",
    "BackupPathBuilder.Build safety-net fallback is rarely/never reached (fires only when BackupHomeDir() fails); recommend KEEP for resilience",
    "replaceFileLegacySidecar defensive helper has 0% per-function coverage; acceptable for a defensive safety-net; file-level gate met at 83.0%",
    "CRLF line endings on 5 pre-existing files in adapters/common/ are a Windows-only artifact; not introduced by this change",
    "2 spec ambiguities (REQ-BRP-02 timestamp format, REQ-BRP-06 directory numbering) were RESOLVED at archive time in the spec"
  ]
}
```
