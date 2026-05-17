# Delta Spec: tui-core — MODIFIED

## Action: MODIFIED

Previously: Model had `Language` field; views accepted `lang string` parameter; Configuration screen had language selector.

| # | Requirement | Scenario |
|---|-------------|----------|
| M1 | Model SHALL NOT carry `Language` field | GIVEN `model/types.go` and `app/model.go` → WHEN compiled → THEN no `Language string` field in any struct |
| M2 | Views SHALL render English-only strings; no `lang` parameter | GIVEN any screen view function → WHEN called → THEN all labels hardcoded English; no `lang` param in signature |
| M3 | Configuration screen SHALL have only persistence backend (1 field, was 2) | GIVEN Configuration screen → WHEN rendered → THEN no language dropdown exists; only backend selector shown |
