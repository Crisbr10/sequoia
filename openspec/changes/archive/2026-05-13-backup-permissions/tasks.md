# Tasks: Restrict Backup File Permissions

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~10 (7 production + ~3 test) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |

## Phase 1: Core Permission Fixes

- [x] 1.1 Fix backup dir permission: `adapters/common/installer.go:84` — change `os.MkdirAll(cfg.BackupDir, 0o755)` to `0o700`
- [x] 1.2 Fix backup file copy permission: `adapters/common/installer.go` after line 89 — add `os.Chmod(dst, 0o600)` with error handling after `copyFile(src, dst)`
- [x] 1.3 Fix ReplaceFile backup permission: `adapters/common/strategy.go:127` — change `0o644` to `0o600` in `os.WriteFile(backup, raw, ...)`
- [x] 1.4 Fix Codex MergeConfig backup permission: `adapters/codex/installer.go:30` — change `0o644` to `0o600` in `os.WriteFile(backupPath, ...)`

## Phase 2: Verification

- [x] 2.1 Run `go test -count=1 ./...`, confirm all tests pass, run `go build ./...` and `go vet ./...`.
