# Delta Spec: i18n-engine — REMOVED

## Action: REMOVED

`internal/i18n/` package, TOML catalogs, go-i18n bundle, message keys, and `i18n.T()` function SHALL be deleted entirely (~815 lines). All TUI screens SHALL use hardcoded English strings instead of `i18n.T(i18n.MsgXxx, lang)`.

| # | Requirement | Scenario |
|---|-------------|----------|
| R1 | i18n package deleted | GIVEN `internal/i18n/` exists → WHEN change applied → THEN 7 files removed; no `i18n` import in any `.go` file |
| R2 | `i18n.Init()` removed | GIVEN app startup → WHEN initialized → THEN no `Init()` call; no TOML catalogs loaded |
| R3 | `i18n.T()` removed | GIVEN any screen view → WHEN rendering → THEN hardcoded English string used; no `T()` call |
