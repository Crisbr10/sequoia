# Proposal: i18n Catalog — TUI String Translation Engine

## Intent

The Configuration screen has a language selector that was hidden (FIX-014) because no i18n infrastructure exists. ~136 user-facing strings are hardcoded in English across 15+ files. Users who select "Español" see zero effect. Build the i18n engine so the TUI renders in the selected language, and re-enable the selector.

## Scope

### In Scope
- Add `github.com/nicksnyder/go-i18n/v2` + `golang.org/x/text/language` dependencies
- Create `internal/i18n/` package: `Bundle`, `T()`, TOML message files (`en.toml`, `es.toml`), `MustLocalize` with template data
- Migrate ~90 TUI strings (8 screen files + `app/view.go` + `app/update.go`) to `i18n.T(key, lang)` calls
- Re-enable language selector rendering in `configuration.go` (uncomment lines 43–62, remove `TODO(i18n)` markers)
- Unskip configuration tests and regenerate golden files with language labels visible
- Wire `Config.Language` into adapter template loading: `renderTemplate(name, lang)` loads `name.en.tmpl` or `name.es.tmpl`

### Out of Scope
- CLI cobra strings (~35) — deferred to Phase 2
- Adapter template content translation (Spanish .tmpl files) — deferred; Phase 1 picks language-resolved files only
- Human review of machine translations
- RTL or locale-aware date/number formatting

## Capabilities

### New Capabilities
- **i18n-engine**: `internal/i18n/` package with `go-i18n/v2` bundle, TOML catalogs (en+es), and `T(key, lang)` accessor. Bootstrap from embedded filesystem at app init.

### Modified Capabilities
- **tui-install-flow**: Configuration screen MUST render language selector when i18n catalogs exist. Golden files MUST include "Language:", "English", "Español". `TODO(i18n)` markers removed.
- **go-wiring**: Adapters MUST use `opts.Language` to select language-resolved template files (`renderTemplate(name, lang)` instead of discarding `_ = opts.Language`).
- **tui-core**: All user-facing strings in `model.go`, `view.go`, `update.go` replaced by `i18n.T()` calls.
- **tui-pipeline**: Progress labels, step names, completion messages use `i18n.T()`.
- **tui-management**: Error messages, footer hints, status labels use `i18n.T()`.

## Approach

1. Create `internal/i18n/i18n.go`: `Bundle` wrapped around `go-i18n/v2`, `T(key, lang)` calling `MustLocalize`, `Init()` loading embedded TOML files
2. Create `internal/i18n/active.en.toml` + `active.es.toml`: keyed by domain (`config.*`, `progress.*`, `error.*`, etc.)
3. Migrate strings screen-by-screen with regression test passes after each file
4. Uncomment language rendering in `configuration.go`; condition on `i18n.Initialized()`
5. Regenerate golden files with `UPDATE_GOLDEN=1 go test ./internal/tui/screens/...`
6. Update `adapters/common/template.go`: `RenderTemplate(name, lang string)` appends `.en.tmpl`/`.es.tmpl` suffix

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/i18n/` | **New** | Bundle, T(), TOML catalogs, embedded FS |
| `go.mod` | Modified | Add `go-i18n/v2` + `x/text/language` |
| `internal/tui/screens/configuration.go` | Modified | Uncomment language rendering |
| `internal/tui/screens/*.go` (7 files) | Modified | Migrate hardcoded strings to `i18n.T()` |
| `internal/app/view.go`, `update.go` | Modified | Migrate strings |
| `adapters/common/template.go` | Modified | Language-resolved template loading |
| `adapters/*.go` (5 adapters) | Modified | Pass language to `RenderTemplate` |
| `internal/tui/screens/testdata/golden/` | Modified | Regenerate `configuration_*.txt` |
| `openspec/specs/tui-install-flow/spec.md` | Modified | Reverse hide-language-selector requirements |
| `openspec/specs/go-wiring/spec.md` | Modified | Adapters must use (not discard) language |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Missed strings during migration | Medium | Grep for `/".*"/` in TUI files post-migration; reviewer checklist |
| Test regression from golden file changes | Medium | Regenerate goldens per documented command; run full suite |
| Template breakage from language-aware loading | Low | Fallback to `.en.tmpl` if lang-resolved file missing |
| go-i18n/v2 incompatibility | Low | `golang.org/x/text` already in `go.mod`; `go-i18n/v2` depends on same |

## Rollback Plan

Revert single commit: comment out language rendering, revert template loading to `_ = lang`, delete `internal/i18n/`, regenerate golden files. No data migration involved.

## Dependencies

- `github.com/nicksnyder/go-i18n/v2` (new — MIT license, 2.6k stars, direct dep of `golang.org/x/text`)
- `golang.org/x/text/language` (new direct dep; already indirect in `go.mod`)

## Success Criteria

- [ ] `go test -race -count=1 ./...` passes with zero failures
- [ ] TUI renders in English when `Config.Language = "en"`
- [ ] TUI renders in Spanish when `Config.Language = "es"`
- [ ] Language selector visible on Configuration screen (EN/ES cycleable)
- [ ] Configuration golden tests pass with visible language labels
- [ ] Adapter template loading respects language (`.en.tmpl` vs `.es.tmpl`)
- [ ] Skipped configuration tests re-enabled and passing
