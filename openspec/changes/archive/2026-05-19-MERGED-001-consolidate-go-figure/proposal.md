# Proposal: Consolidate go-figure Usage

## Intent

Remove the abandoned `github.com/common-nighthawk/go-figure` dependency (last release June 2021, 5 years stale). It is used exactly once — to render the static string `"Sequoia"` as ASCII art for the Welcome screen. The output never changes and carries no runtime dynamism. Removing this dependency eliminates ~2MB of binary weight, a stale supply-chain entry, and a Go version compatibility risk.

## Scope

### In Scope
- Replace `figure.NewFigure("Sequoia", "", true)` call in `internal/tui/styles/logo.go` with a hardcoded `const` containing the identical 6-line ASCII art
- Remove `go-figure` import from `logo.go` and drop `sync` import (no longer needed)
- Run `go mod tidy` to remove `go-figure` from `go.mod` and `go.sum`
- Regenerate 2 golden test files (`welcome_standard.txt`, `welcome_cursor_status.txt`) via `UPDATE_GOLDEN=1 go test`

### Out of Scope
- Any visual or layout changes to the Welcome screen
- Any behavior change to `styles.Logo()` — it returns the same styled string
- Other dependency cleanup tasks

## Capabilities

### New Capabilities
None

### Modified Capabilities
None — pure refactor, zero spec-level behavior change. Logo output is preserved identically.

## Approach

1. Run the project, capture the current `figure.NewFigure("Sequoia", "", true).String()` output as a `const rawLogo` in `logo.go`
2. Replace the `Logo()` function body: return `rawLogo` wrapped with the same `lipgloss` foreground style (no behavior change)
3. Simplify imports: remove `"github.com/common-nighthawk/go-figure"` and `"sync"`; remove `sync.Once` guard (const is zero-cost)
4. Run `go mod tidy` to purge go-figure from module graph
5. Regenerate golden files with `UPDATE_GOLDEN=1 go test ./internal/tui/screens/...` to reflect any whitespace/encoding drift
6. Run `go test -race ./...` to confirm zero regressions

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/tui/styles/logo.go` | Modified | Replace `sync.Once`+`figure.NewFigure` with `const` |
| `go.mod` | Modified | Remove go-figure line |
| `go.sum` | Modified | Remove go-figure checksums |
| `internal/tui/screens/testdata/golden/welcome_standard.txt` | Regenerated | Golden file update |
| `internal/tui/screens/testdata/golden/welcome_cursor_status.txt` | Regenerated | Golden file update |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Golden file mismatch due to whitespace/encoding differences between const and original figure.String() | Low | Regenerate via `UPDATE_GOLDEN=1` and visually diff before committing |
| go mod tidy removes more than go-figure (unlikely per Go module graph rules) | Low | Review `git diff go.mod go.sum` before committing |

## Rollback Plan

1. Revert `internal/tui/styles/logo.go` to restore `figure.NewFigure` call
2. Run `go mod tidy` to restore `go-figure` in `go.mod`/`go.sum`
3. Revert golden files: `git checkout -- internal/tui/screens/testdata/golden/`
4. Run `go test -race ./...` to confirm

## Dependencies

None — no other change depends on or blocks this.

## Success Criteria

- [ ] `go.mod` no longer references `github.com/common-nighthawk/go-figure`
- [ ] `styles.Logo()` returns identical output (same ASCII art, same lipgloss styling)
- [ ] All golden tests pass after regeneration
- [ ] `go test -race ./...` passes with zero failures
