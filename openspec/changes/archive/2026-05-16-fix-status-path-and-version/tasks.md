# Tasks: Fix Status Path Display and Welcome Version

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~55 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | auto-chain |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

## Phase 1: RED — Write failing tests

- [x] 1.1 Rename `TestStatusView_ShowsVersionAndPath` → `TestStatusView_ShowsVersion` in `internal/tui/screens/status_test.go`; drop `.claude` assertion (line 51), add `assert.NotContains(t, view, "/home/user/.claude")`. Run: FAILS — path still rendered.
- [x] 1.2 Update `testdata/golden/status_all_installed.txt`: drop `/home/user/.claude` (L3) and `/home/user/.config/opencode` (L4) path strings. Run golden test: FAILS.
- [x] 1.3 Update `testdata/golden/status_mixed.txt`: drop `/home/user/.claude` path (L3), replace `—  —` with single `—` (L4). Run golden test: FAILS.
- [x] 1.4 Add `TestResolveVersion_PassThrough` table-driven test in `cmd/sequoia/main_test.go` with cases: `"1.2.3"` → `"1.2.3"`, `""` → `""`. Run: FAILS — `resolveVersion` undefined.
- [x] 1.5 Add `TestResolveVersion_DevResolves` in `cmd/sequoia/main_test.go`: assert `resolveVersion("0.1.0-dev")` is non-empty, not `"(devel)"`. Run: FAILS.

## Phase 2: GREEN — Implement to make tests pass

- [x] 2.1 Drop `path` variable (L76–79) and path `%s` from `fmt.Sprintf` (L81–87) in `internal/tui/screens/status.go`; change to 4-field format: `prefix + marker + name + version`.
- [x] 2.2 Drop `, and path` from `StatusView` Godoc (L15) in `status.go`.
- [x] 2.3 Add `resolveVersion(raw string) string` to `cmd/sequoia/main.go` after `newVersionCmd` (post-L231): resolves `"0.1.0-dev"` via `debug.ReadBuildInfo()`, pass-through otherwise.
- [x] 2.4 Replace inline resolution in `newVersionCmd()` RunE (L210–224) with single `fmt.Fprintln(cmd.OutOrStdout(), resolveVersion(Version))`.
- [x] 2.5 Change `runTUI()` L404: `Version` → `resolveVersion(Version)`.

## Phase 3: VERIFY — Confirm full suite green

- [x] 3.1 Run `go test ./internal/tui/screens/` — golden tests and `TestStatusView_ShowsVersion` pass.
- [x] 3.2 Run `go test ./cmd/sequoia/` — `TestResolveVersion_*` and existing `TestVersionCmd*` pass.
- [x] 3.3 Run `go test ./...` — full suite green.
