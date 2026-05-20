# Tasks: Consolidate go-figure Usage

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~20 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Delivery strategy | single-pr |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

## Phase 1: Core Implementation

- [ ] 1.1 Replace `figure.NewFigure`+`sync.Once` with `const rawLogo` in `internal/tui/styles/logo.go` — use the exact 6-line ASCII art from the design, simplify imports (drop `sync` and `go-figure`), preserve lipgloss per-line styling
- [ ] 1.2 Verify logo unit tests: `go test ./internal/tui/styles/... -count=1` — all 5 tests (NonEmpty, MultiLine, ANSI, idempotency, goroutine safety) must pass with the const-based implementation

## Phase 2: Verification & Cleanup

- [ ] 2.1 Regenerate golden test files: `$env:UPDATE_GOLDEN="1"; go test ./internal/tui/screens/...` — updates `welcome_standard.txt` and `welcome_cursor_status.txt`
- [ ] 2.2 Clean dependencies: `go mod tidy` — removes `go-figure` from go.mod/go.sum
- [ ] 2.3 Full verification: `go vet ./...`, `go build ./...`, `go test -race -count=1 ./...` — all must pass with zero failures
- [ ] 2.4 Visual diff golden files via `git diff internal/tui/screens/testdata/golden/` before committing
