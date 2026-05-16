# Tasks: Fix Uninstall Orphaned Files

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~140 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Low

## Phase 1: i18n Foundation

- [x] 1.1 Add `MsgCompleteUninstalledItems = "complete.uninstalled_items"` to `internal/i18n/keys.go:57` (Complete screen block)
- [x] 1.2 Add `uninstalled_items = "Uninstalled: Skills, Commands, System Prompt"` to `internal/i18n/translations/en.toml` under `[complete]`
- [x] 1.3 Add `uninstalled_items = "Desinstalado: Skills, Comandos, System Prompt"` to `internal/i18n/translations/es.toml` under `[complete]`
- [x] 1.4 Add `i18n.MsgCompleteUninstalledItems` to `allKeys` in `internal/i18n/keys_test.go:52` and bump expected count 65→66

## Phase 2: Adapter Path Resolution Fixes

- [x] 2.1 Fix Gemini Uninstall: replace `geminiBase(a.HomeDir())` with `a.Base()` at `adapters/gemini/adapter.go:77`
- [x] 2.2 Fix Codex Install: replace `codexBase(a.HomeDir())` with `a.Base()` at `adapters/codex/adapter.go:80`
- [x] 2.3 Fix Codex Uninstall: replace `codexBase(a.HomeDir())` with `a.Base()` at `adapters/codex/adapter.go:175`

## Phase 3: TUI CompleteView Fix

- [x] 3.1 Add `else if mode == "uninstall"` branch at `internal/tui/screens/complete.go:73` using `MsgCompleteUninstalledItems`

## Phase 4: Testing

- [x] 4.1 Add table-driven production-path test to `adapters/gemini/adapter_test.go` using `t.Setenv("USERPROFILE"`/`"HOME", t.TempDir())` + `NewAdapter("")`, verifying sequoia dir removed after uninstall
- [x] 4.2 Add table-driven production-path test to `adapters/codex/adapter_test.go` using same `t.Setenv` pattern, verifying install creates + uninstall removes files under controlled home
- [x] 4.3 Add `TestCompleteView_UninstallModeShowsUninstalledItems` to `internal/tui/screens/complete_test.go` asserting `"Uninstalled"` in output with `mode="uninstall", warnedCount=0`
- [x] 4.4 Add golden test `TestCompleteView_Golden_UninstallClean` in `complete_test.go`, generate `testdata/golden/complete_uninstall_clean.txt` via `UPDATE_GOLDEN=1`
- [x] 4.5 Run `go test -race -count=1 ./...` to confirm zero regressions
