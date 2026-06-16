# Spec: ci-green-gate

> **Spec ID**: ci-green-gate
> **Type**: delta (all requirements are NEW, authored from commit `20f352d` CI #140 failing baseline)
> **Authored from**: `openspec/changes/fix-ci-140-lint-and-race/specs/ci-green-gate/spec.md`
> **Archived**: 2026-06-16

## REQ-CIG-01 — Lint-clean (golangci-lint v2.12)

`golangci-lint run ./...` MUST return exit code 0 under golangci-lint v2.12; the `unused` rule and the built-in `gofmt` check MUST be satisfied. The 6 known issues fixed: `adapters/common/base_adapter_internal_test.go:57` (delete unused `internalFileExists`); `adapters/opencode/adapter.go:42`, `adapters/testutil/mock_adapter.go:1`, `internal/codegraph/install.go:1`, `internal/codegraph/install_test.go:1`, `internal/tui/styles/logo.go:1`, `internal/tui/styles/styles.go:1` (gofmt reformat of the whole file).

#### Scenario: golangci-lint run reports zero issues

- GIVEN the 6 lint issues are fixed
- WHEN `golangci-lint run ./...` is executed
- THEN the exit code is 0
- AND no issue is reported at any of the 6 fixed locations

## REQ-CIG-02 — go vet clean

`go vet ./...` MUST return exit code 0.

#### Scenario: go vet ./... passes

- GIVEN the fix is applied
- WHEN `go vet ./...` is executed
- THEN the exit code is 0
- AND no diagnostics are printed

## REQ-CIG-03 — gofmt -l . reports no files

`gofmt -l .` MUST produce no output (all `.go` files already gofmt-formatted).

#### Scenario: gofmt -l . produces no output

- GIVEN the fix is applied
- WHEN `gofmt -l .` is executed
- THEN no filenames are printed
- AND the exit code is 0

## REQ-CIG-04 — Race-free tests under -race

`go test ./... -race -count=1 -timeout 180s` MUST pass on the 4 non-windows CI platforms. The data race MUST be eliminated: parallel `BaseAdapter.Apply()` tests in `adapters/common/base_adapter_error_test.go` share adapterID `"err-test"` and the package-level `userConfigDir` (`adapters/common/backup_retention.go:58`), so the retention hook in one goroutine can call `PruneBackups()` → `os.RemoveAll()` on a session dir a concurrent failed-install rollback in another goroutine is reading from.

#### Scenario: race detector reports no races between parallel central-home tests

- GIVEN 2 parallel tests both call `BaseAdapter.Apply()` (one successful, one failing mid-stage) using adapterID `"err-test"`
- WHEN `go test ./adapters/common/... -race -count=1 -timeout 60s` runs on a non-windows runner
- THEN the race detector reports 0 races
- AND each test resolves `BackupHomeDir()` to its own `t.TempDir()`-backed path
- AND all parallel tests exit successfully

## REQ-CIG-05 — Test isolation via overrideUserConfigDir

Every test in `adapters/common/` that touches the central-home backup flow MUST call `OverrideUserConfigDir(t, func() (string, error) { return t.TempDir(), nil })` at the start of its body, so `BackupHomeDir()` resolves to a per-test isolated path. MUST be updated by this change: `base_adapter_error_test.go` — its `fullInstallTestAdapter` helper is called by 21 `t.Parallel()` tests and MUST call `OverrideUserConfigDir`.

#### Scenario: overrideUserConfigDir is used by every central-home test

- GIVEN the test file inventory in `adapters/common/` after the fix
- WHEN `grep -L OverrideUserConfigDir adapters/common/*_test.go` is run
- THEN it returns only test files that do NOT call `BackupHomeDir()`, `BaseAdapter.Apply()`, or `BaseAdapter.Install()`
- AND `base_adapter_error_test.go` is in the returned list

#### Scenario: fullInstallTestAdapter applies overrideUserConfigDir

- GIVEN `fullInstallTestAdapter(t, home)` is invoked from a parallel test in `base_adapter_error_test.go`
- WHEN the test body runs
- THEN `OverrideUserConfigDir` has been called with a `t.TempDir()`-backed hook
- AND `BackupHomeDir()` returns a path unique to that test invocation

## REQ-CIG-06 — CI workflow unchanged (no -race on Windows)

`.github/workflows/ci.yml` MUST remain unchanged: matrix stays `[ubuntu-latest, macos-latest, macos-14, ubuntu-24.04-arm, windows-latest]`; the Windows branch MUST still drop `-race -covermode=atomic`. This change MUST NOT add `-race` to the Windows branch.

#### Scenario: ci.yml is unchanged from commit 20f352d

- GIVEN the working branch has the fix applied
- WHEN `git diff 20f352d..HEAD -- .github/workflows/ci.yml` is executed
- THEN the diff is empty
- AND the Windows branch still drops `-race -covermode=atomic`

## REQ-CIG-07 — No production behavior change

This change MUST NOT alter any production behavior. Invariants: `applyRetention` still runs at the end of `BaseAdapter.Apply()` (line 595 in `adapters/common/base_adapter.go`) on the happy path; `PruneBackups` still enforces the 5-backup cap; `BackupHomeDir` still returns `<UserConfigDir>/sequoia/backups/`; retention errors still surface via `AddWarning`.

#### Scenario: production behaviors are preserved

- GIVEN a successful `BaseAdapter.Apply()` call against an adapter with 7 pre-seeded session dirs
- WHEN the call returns nil
- THEN `applyRetention` was called exactly once before `return nil`
- AND exactly 5 session directories remain
- AND a prune error (if any) is recorded via `AddWarning` with prefix `"backup retention:"`

## REQ-CIG-08 — No coverage regression

Per-package coverage MUST remain at or above 70% (the CI gate from `openspec/config.yaml` `verify.coverage_threshold`). The fix touches only test files and gofmt-formatted files; production code is not modified.

#### Scenario: per-package coverage still meets the 70% gate

- GIVEN the fix is applied
- WHEN `go test ./... -coverprofile=coverage.out -count=1 -timeout 180s` is executed
- THEN every package reports coverage of 70% or higher
- AND `adapters/common` is still at or above its pre-fix baseline (≈83%)
