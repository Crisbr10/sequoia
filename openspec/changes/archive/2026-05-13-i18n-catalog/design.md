# Design: i18n Catalog — TUI String Translation Engine

## Technical Approach

Build `internal/i18n/` with `go-i18n/v2` + TOML catalogs. Initialize bundle at app start via `sync.Once` from embedded TOML files. Thread `lang string` through all View functions — each calls `i18n.T(key, lang)`. Migrate strings screen-by-screen with golden file regeneration after each. Re-enable language selector in `configuration.go`; unskip its tests. Wire `opts.Language` into adapter template loading via `RenderTemplateLang(fs, name, lang)`.

## Architecture Decisions

| Decision | Options | Choice | Rationale |
|----------|---------|--------|-----------|
| Message format | TOML / JSON / YAML | **TOML** | Already dependency (`BurntSushi/toml`); go-i18n/v2 native; readable for translators |
| Key convention | dot-notation / camelCase / flat | **dot-notation** (`screen.element`) | Mirrors go-i18n examples; scannable; prevents collisions |
| Fallback on missing key | crash / return key / return empty | **Return key + log warning** | Never crash TUI for a missing translation; key is readable English default |
| Language threading | global var / context / explicit param | **Explicit `lang string` param** | Matches existing signature patterns; testable without global state |
| Template resolution | suffix `.en.tmpl` / subdirectory / separate FS | **Suffix `.en.tmpl` / `.es.tmpl`** | Minimal code change; falls back to bare name (English) if lang file missing |

## Data Flow

```
User selects language → TUIConfig.Language updated (configuration.go:cycleOption)
  → Model.View() reads m.Config.Language, passes to each screen View()
    → Each screen calls i18n.T("key", lang) for every string
  → Pipeline: m.Config.Language → adapter.Install(opts) / adapter.Uninstall(opts)
    → BaseAdapter: RenderTemplateLang(fs, name, lang) picks name.en.tmpl or name.es.tmpl
```

## Public API

```go
// internal/i18n/bundle.go — public API
func Init() error                           // loads embedded TOML catalogs
func T(key, lang string, data ...any) string  // MustLocalize with fallback
func Initialized() bool                     // gate for language selector visibility

// adapters/common/template.go — new variant
func RenderTemplateLang(fs embed.FS, name, lang string, data any) (string, error)
// Resolves: name + "." + lang + ".tmpl" → fallback to name.tmpl if missing
```

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit (i18n) | Bundle init, T() lookup, missing key fallback | `go test ./internal/i18n/...` — table-driven with en+es |
| Unit (screen) | View functions render translated strings | Add `lang` param to existing tests; assert translated labels appear |
| Golden | Regenerate all 14 golden files | `UPDATE_GOLDEN=1 go test ./internal/tui/screens/...` |
| Integration | Full TUI renders in ES | Manual smoke test; automated via Bubbletea test harness |
| Adapter template | `RenderTemplateLang` resolves `.en.tmpl` / `.es.tmpl` | Unit test with embedded test FS |
