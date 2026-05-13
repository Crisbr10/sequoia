# Tasks: i18n-catalog

**Status**: 18/18 complete ✅

## Phase 1: i18n Engine Foundation (PR 1)

- [x] 1.1 Create `internal/i18n/bundle.go` — `Bundle`, `Init()` via `sync.Once`, `T(key,lang,...data)` via `MustLocalize` with English fallback, `Initialized() bool`
- [x] 1.2 Create `internal/i18n/bundle_test.go` — table-driven: TestInit, TestT, TestInitialized, 20-key triangulation
- [x] 1.3 Create `internal/i18n/keys.go` — 65 `Msg*` string constants dot-keyed
- [x] 1.4 Create `internal/i18n/translations/en.toml` + `es.toml` (65 keys), `embed.go`
- [x] 1.5 Create `internal/i18n/keys_test.go` — verify every constant exists in both catalogs
- [x] 1.6 Modify `go.mod` — add `go-i18n/v2` + `x/text`. Modify `model.go` `Init()` — call `i18n.Init()`

## Phase 2: TUI String Migration (PR 2)

- [x] 2.1 Modify `internal/app/view.go` — extract `lang`, pass to all 8 screen View functions
- [x] 2.2 Update all 8 screen View signatures to accept `lang string`. Update test files + create `main_test.go`
- [x] 2.3 Migrate all 8 screen files to `i18n.T(keys.MsgXxx, lang)` — TDD cycle per screen
- [x] 2.4 Migrate `internal/app/update.go` — replace `m.ErrorMsg` with `i18n.T()`
- [x] 2.5 Modify `configuration.go` — uncomment language rendering, gate on `i18n.Initialized()`, remove TODO(i18n)
- [x] 2.6 Unskip `TestConfigurationView_ShowsLanguageOptions` + `TestConfigurationView_RendersLanguageAndPersistence`
- [x] 2.7 Regenerate 14 golden files via `UPDATE_GOLDEN=1 go test`

## Phase 3: Adapter Template Wiring + Specs (PR 3)

- [x] 3.1 Add `RenderTemplateLang()` to `adapters/common/template.go` — lang-resolved template loading
- [x] 3.2 Modify `adapters/common/base_adapter.go` — wire language to template rendering, remove `_ = opts.Language`
- [x] 3.3 Update individual adapters (codex, gemini, _template) — wire language, remove discards
- [x] 3.4 Update `openspec/specs/go-wiring/spec.md` + `openspec/specs/template-wiring/spec.md`
- [x] 3.5 Final verification: `go test ./...` → ALL 18 PACKAGES PASS
