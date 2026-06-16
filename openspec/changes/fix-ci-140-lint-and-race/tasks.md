# Tasks: fix-ci-140-lint-and-race

## Task Tracker

### Commit 1 — Lint fix: delete unused function

- [ ] **1.1** Delete unused `internalFileExists` in `adapters/common/base_adapter_internal_test.go` (lines 55–60). Verify: `golangci-lint run ./adapters/common/...` no longer reports unused error; `go test ./adapters/common/... -count=1` still passes.

### Commit 2 — Lint fix: gofmt 6 files

- [ ] **2.1** Run `gofmt -w` on `adapters/testutil/mock_adapter.go`
- [ ] **2.2** Run `gofmt -w` on `internal/codegraph/install.go`
- [ ] **2.3** Run `gofmt -w` on `internal/codegraph/install_test.go`
- [ ] **2.4** Run `gofmt -w` on `internal/tui/styles/logo.go`
- [ ] **2.5** Run `gofmt -w` on `internal/tui/styles/styles.go`
- [ ] **2.6** Run `gofmt -w` on `adapters/opencode/adapter.go` (CRLF + inline-closure reformat)
- **Verify**: `gofmt -l .` reports 0 files; `golangci-lint run ./...` reports 0 gofmt errors; `go build ./...` passes.

### Commit 3 — Test isolation: shared helpers

- [ ] **3.1** Add `overrideUserConfigDir(t, t.TempDir())` to `fullInstallTestAdapter` in `adapters/common/base_adapter_error_test.go` (line 86)
- [ ] **3.2** Add `overrideUserConfigDir(t, t.TempDir())` to `installTestAdapter` in `adapters/common/base_adapter_test.go` (line 198)
- [ ] **3.3** Add `overrideUserConfigDir(t, t.TempDir())` to `warningsTestAdapter` in `adapters/common/base_adapter_test.go` (line 316)
- **Verify**: `go test ./adapters/common/... -count=1` passes. (CGO-dependent race check deferred to CI.)

### Commit 4 — Test isolation: 5 direct-build tests

- [ ] **4.1** Add `overrideUserConfigDir(t, t.TempDir())` to `TestInstall_StagingDirCreationFailure` in `adapters/common/base_adapter_error_test.go`
- [ ] **4.2** Add `overrideUserConfigDir(t, t.TempDir())` to `TestInstall_SkillTemplateNotFound`
- [ ] **4.3** Add `overrideUserConfigDir(t, t.TempDir())` to `TestInstall_SystemPromptTemplateNotFound`
- [ ] **4.4** Add `overrideUserConfigDir(t, t.TempDir())` to `TestInstall_BaseResolutionFailure`
- [ ] **4.5** Add `overrideUserConfigDir(t, t.TempDir())` to `TestInstall_VersionFileWriteFailure`
- **Verify**: `go test ./adapters/common/... -count=1` passes.

### Cross-cutting

- [ ] **X.1** Final verification: `gofmt -l .` empty → `go vet ./...` clean → `golangci-lint run ./...` exit 0 → `go test ./... -race -count=1 -timeout 180s` passes (×5) → `go tool cover -func=coverage.out | grep total` ≥ 70%
- [ ] **X.2** Update apply-progress.md with task status, commit SHAs, verification results, and "Next: sdd-verify" hint.

---

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~50 (5 gofmt files + 1 unused deletion + 19 test additions ≈ 50 net) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR (per Design Decision 5) |
| Delivery strategy | `ask-on-risk` (config), but single PR keeps diff under 400 — auto-proceed |
| Chain strategy | N/A (single PR) |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Delete unused `internalFileExists` | PR 1 | Lint fix, single line deletion |
| 2 | gofmt 6 files | PR 1 | Lint fix, included in same PR |
| 3 | Add `overrideUserConfigDir` to 3 shared helpers | PR 1 | Race fix, covers 14 of 19 tests |
| 4 | Add `overrideUserConfigDir` to 5 direct-build tests | PR 1 | Race fix, remaining 5 tests |
