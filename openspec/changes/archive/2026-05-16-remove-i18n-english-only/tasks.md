# Tasks: Remove I18n — English Only

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~6,000 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: Code (Phases 1–4, ~4,000 lines) → main; PR 2: Docs (Phase 5, ~2,000 lines) → main |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |
| 400-line budget risk | High |

## Phase 1: Delete i18n Package & Dependencies

- [x] 1.1 Delete `internal/i18n/` (7 files: bundle, keys, tests, TOML catalogs) — ~815 lines
- [x] 1.2 Delete `adapters/interface_test.go` (44 lines) — tests removed `InstallOpts.Language`
- [x] 1.3 Run `go build ./...`; fix any lingering `i18n` import errors

## Phase 2: Inline English Strings in TUI Screens

- [x] 2.1 `welcome.go` — drop `lang` param, replace 6 `i18n.T()` with English, remove import
- [x] 2.2 `tool-selection.go` — drop `lang`, replace 8 calls, remove import
- [x] 2.3 `configuration.go` — drop `lang`, replace ~15 calls, delete language selector, simplify `toggleField`, remove `i18n.Initialized()` guard, remove import
- [x] 2.4 `install-progress.go` — drop `lang`, replace 4 calls, remove import
- [x] 2.5 `complete.go` — drop `lang`, replace 8 calls, drop warnedCount template data, remove import
- [x] 2.6 `error.go` — drop `lang`, replace 7 calls, remove import
- [x] 2.7 `status.go` — drop `lang`, replace 12 calls, remove import
- [x] 2.8 `uninstall.go` — drop `lang` from `UninstallView` and `RenderConfirmPrompt`, replace 10 calls, remove import
- [x] 2.9 `app/view.go` — drop `lang := string(m.Config.Language)`, remove `lang` from 8 view calls, replace default-case Spanish string, remove import
- [x] 2.10 `app/update.go` — replace 2 `i18n.T(i18n.MsgValidation...)` with English, drop `m.Config.Language` from `startPipeline`, remove import

## Phase 3: Strip Model Language State

- [x] 3.1 `model/types.go` — remove `Language` type, `LangEN`/`LangES` constants, `TUIConfig.Language`
- [x] 3.2 `app/model.go` — `Init()` drops `i18n.Init()`; `NewModel` drops `Language: "en"`; remove import
- [x] 3.3 `pipeline/runner.go` — `RunInstall`/`RunUninstall` drop `lang string`; cascade through `runSteps`/`runInstallSteps`/`runUninstallSteps`; `InstallOpts{Context: ctx}`
- [x] 3.4 `adapters/interface.go` — remove `Language string` from `InstallOpts`; update comments
- [x] 3.5 `adapters/common/template.go` — delete `RenderTemplateLang` (~23 lines)
- [x] 3.6 `adapters/common/base_adapter.go` — replace `RenderTemplateLang(...)` with `RenderTemplate(...)` (×2), drop `lang` extraction, drop `fmt` import if unused
- [x] 3.7 `adapters/codex/adapter.go` — same: drop `lang`, replace `RenderTemplateLang` → `RenderTemplate`
- [x] 3.8 `adapters/_template/adapter.go` — same as codex

## Phase 4: Regenerate Golden Files & Update Tests

- [x] 4.1 Delete `TestConfigurationView_ShowsLanguageOptions` from `configuration_test.go`
- [x] 4.2 Update screen tests — drop `"en"` 3rd arg from all view function calls
- [x] 4.3 Update `app/model_test.go` — drop `Language: "en"` assertion; drop `i18n.Init` test
- [x] 4.4 Regenerate 15 golden files: `UPDATE_GOLDEN=1 go test ./internal/tui/screens/... -run Golden`
- [x] 4.5 Review golden diffs — confirm Spanish→English, no ANSI corruption
- [x] 4.6 Remove `go-i18n/v2` and `x/text` from `go.mod`; run `go mod tidy`

## Phase 5: Rewrite Skill Documentation (PR 2)

- [x] 5.1 Delete `docs/agents/sequoia-i18n.md` and `docs/sequoia/sequoia-phases/07-i18n.md`
- [x] 5.2 Rewrite ~8 agent docs in pure English (remove all Spanish text)
- [x] 5.3 Rewrite ~10 phase and reference docs in English
- [x] 5.4 Rewrite 5 CLI command templates — remove Spanish fragments
- [x] 5.5 Strip P7 from 7 `skill.md.tmpl` templates — remove P7 from roster, phase matrix, health score, delegation sections
- [x] 5.6 Strip P7 from `rules.md.tmpl` and `*-section.md.tmpl` templates (5 files)

## Phase 6: Final Verification

- [x] 6.1 `go mod tidy` — confirm `go-i18n/v2`/`x/text` absent, `BurntSushi/toml` present
- [x] 6.2 `go vet ./...` — zero warnings
- [x] 6.3 `go test -race -count=1 ./...` — zero failures, zero races (core)
- [x] 6.4 `go build -o sequoia.exe .` — binary builds
- [x] 6.5 Review adapter golden files — regenerate if P7 stripping affected them
