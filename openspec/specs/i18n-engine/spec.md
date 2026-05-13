# i18n-engine Specification

## Purpose

`internal/i18n/` package providing TOML-catalog message localization via `go-i18n/v2` for the Sequoia TUI.

## Requirements

### Requirement: Bundle Initialization

`internal/i18n/` MUST embed `active.en.toml` and `active.es.toml` from Go embedded FS. `Init()` SHALL parse both into a `go-i18n/v2` bundle via `sync.Once`. Missing/corrupt English catalog MUST be fatal (return error). Missing Spanish catalog SHOULD log warning, not abort.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 1 | Happy path | Both `.toml` files embedded | `Init()` called | Bundle has messages for "en" and "es"; nil error |
| 2 | Missing English fatal | `active.en.toml` not found | `Init()` called | Non-nil error; app SHALL exit |
| 3 | Missing Spanish non-fatal | Only `active.en.toml` present | `Init()` called | Warning logged; bundle initialized with "en" only |

### Requirement: T() Accessor

`T(key, lang)` SHALL return localized message via `MustLocalize`. Missing key in target language MUST fall back to English. Format args MAY be `map[string]interface{}`.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 4 | English match | key="config.title" in en catalog | `T("config.title","en")` | "Configuration" returned |
| 5 | Spanish match | key="config.title" in es catalog | `T("config.title","es")` | "Configuración" returned |
| 6 | Fallback to English | key only in English catalog | `T(key,"es")` | English value returned; no panic |
| 7 | Template args | key="progress.status" with `{Count}` | `T(key,"en",map{"Count":3})` | "3 of 5 tools" returned |

### Requirement: Initialized Guard

Package MUST expose `Initialized() bool`. Configuration screen SHALL only render language selector when true.

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 8 | Initialized true | `Init()` succeeded | `Initialized()` | true |
| 9 | Initialized false | `Init()` not called | `Initialized()` | false |
