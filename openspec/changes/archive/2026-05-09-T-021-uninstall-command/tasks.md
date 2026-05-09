# Tasks: Uninstall Command Safety (T-021)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 120–150 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Full change: `--yes` flag + confirmation gate + 8 test cases | PR 1 | Single self-contained change; all tests included |

## Phase 1: RED — Write Failing Tests

- [x] 1.1 Add `isTerminalFn` package var (`var isTerminalFn = isTerminal`) in `cmd/sequoia/main.go`
- [x] 1.2 Write 8 table-driven `TestUninstall*` cases in `cmd/sequoia/main_test.go`:
  - `TestUninstall_YesFlagBypass`: `yes=true` → no prompt, expects no "?"
  - `TestUninstall_ConfirmYes`: `in="y\n"`, term=true → expects prompt with "y/N"
  - `TestUninstall_ConfirmNo`: `in="n\n"`, term=true → expects "aborted", err==nil
  - `TestUninstall_ConfirmEmpty`: `in="\n"`, term=true → expects "aborted"
  - `TestUninstall_PipedStdinError`: term=false, yes=false → expects "--yes" in error
  - `TestUninstall_AllListsTools`: all=true, term=true → output lists tool names before prompt
  - `TestUninstall_InvalidTool`: toolID="no-existe" → expects "unknown adapter"
  - `TestUninstall_YesFlagRegistered`: sets args `["uninstall","--help"]` → output contains "--yes"
  - Use mock `strings.NewReader` for `in`, override `isTerminalFn` with defer restore; do NOT call `t.Parallel()` on tests mutating `isTerminalFn`

## Phase 2: GREEN — Implement Confirmation Gate

- [x] 2.1 Add `yesFlag bool` to `newUninstallCmd()`, register with `cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")`
- [x] 2.2 Change `runUninstall` signature to `runUninstall(toolID string, all bool, yes bool, in io.Reader, out io.Writer) error`
- [x] 2.3 Implement confirmation gate after target resolution and before execute loop:
  - If `yes`: skip prompt, proceed to execute loop
  - If `!isTerminalFn()`: return `fmt.Errorf("stdin is not a terminal; use --yes to skip confirmation")`
  - Prompt format: single tool → `"Remove Sequoia from {Name}? [y/N]: "`; `--all` → `"This will remove Sequoia from:\n  {Name1}\n  {Name2}\nContinue? [y/N]: "`
  - Read with `fmt.Fscanln(in, &response)`; if `response != "y" && response != "Y"` → `fmt.Fprintln(out, "Uninstall aborted.")` + `return nil`
- [x] 2.4 Wire `RunE`: pass `yesFlag` and `cmd.InOrStdin()` to `runUninstall`; update internal call sites if any

## Phase 3: Verify — Run Tests & Polish

- [x] 3.1 Run `go test ./cmd/sequoia/... -run Uninstall -count=1` — all 10 uninstall tests pass (2 existing + 8 new); existing `TestUninstallHelp` and `TestUninstallAllFlag` still pass
- [x] 3.2 Run `go test ./... -count=1 -timeout 120s` — confirmed zero regressions across all 5 packages
- [x] 3.3 Verify godoc comments on `isTerminalFn` and the revised `runUninstall` signature; `fmt` import already present
