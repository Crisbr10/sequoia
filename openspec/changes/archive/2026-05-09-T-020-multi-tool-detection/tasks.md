# Tasks: Multi-tool Detection (T-020)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 200-250 (50 source, 200 tests) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR — well under 400-line budget |
| Delivery strategy | single-pr |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

## Phase 1: Foundation — Version File + Symlink Paths

- [x] T-020-01 **Add `versionFilePath(base)` to claude + opencode paths.go**. Add `filepath.EvalSymlinks()` fallback to `claudeBase()` and `opencodeBase()`. RED: table-driven test for suffix; symlink test with `os.Symlink` temp dir. GREEN: implement. REFACTOR: godoc. Files: `adapters/claude/paths.go`, `adapters/opencode/paths.go`, both `adapter_test.go`. Effort: M.

## Phase 2: Core — Status, Install, Uninstall

- [x] T-020-02 **Update `Status()` to read `.sequoia-version` in both adapters**. RED: test that Status.Version matches file content; legacy-missing-file → "" . GREEN: read via `os.ReadFile`+`TrimSpace`. REFACTOR: extract `readVersion(base)` helper. Files: both `adapter.go` + `adapter_test.go`. Effort: M.

- [x] T-020-03 **Write `.sequoia-version` in `Install()`, remove in `Uninstall()`**. After staging but before installer runs, write `Version` constant to `versionFilePath(skillsPath(base))`. Uninstall removes it best-effort. RED: round-trip test Install→Status.Version. GREEN: add write/remove calls. REFACTOR: consistent error handling. Files: both `adapter.go` + `adapter_test.go`. Effort: M.

## Phase 3: CLI — Status Columns + ScanTools

- [x] T-020-04 **Add `ScanTools()` and update `runStatus` in `cmd/sequoia/main.go`**. `ScanTools()` iterates `DefaultRegistry.All()`, calls `Status()`, returns `[]AdapterStatus`. `runStatus` replaces per-field calls with single `a.Status()`, uses 6-column format: ID(14) NAME(14) DETECTED(9) INSTALLED(10) VERSION(10) PATH(55). RED: test headers, column alignment, `ScanTools` count. GREEN: implement. REFACTOR: extract `formatStatusRow`. Files: `cmd/sequoia/main.go`, `cmd/sequoia/main_test.go`. Effort: L.

## Phase 4: Verification — Integration + Edge Cases

- [x] T-020-05 **Integration tests: version round-trip and symlink resolution**. RED: Install→Status→Version matches const; symlink dir → `EvalSymlinks` → resolved Path in Status. GREEN: `t.TempDir()`-based setup using public API. REFACTOR: shared test helper `setupAdapterWithVersion(t)`. Files: both `adapter_test.go`, `cmd/sequoia/main_test.go`. Effort: L.

- [x] T-020-06 **Edge case tests**: empty registry (`runStatus` prints "No adapters"), legacy no-version (Status.Version=""), `EvalSymlinks` error fallback, reinstall overwrites version. All `t.TempDir()`, never mutates real home. Files: both `adapter_test.go`, `main_test.go`. Effort: M.
