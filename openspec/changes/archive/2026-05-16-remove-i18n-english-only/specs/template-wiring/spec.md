# Delta Spec: template-wiring — MODIFIED

## Action: MODIFIED

Previously: `RenderTemplateLang` resolved `templates/{lang}/` subdirectories; templates had P7 rows in roster tables, Phase 2 matrix, health score, and delegation sections.

| # | Requirement | Scenario |
|---|-------------|----------|
| M15 | `RenderTemplateLang` SHALL be deleted; `RenderTemplate` used instead | GIVEN BaseAdapter renders template → WHEN called → THEN template from single `templates/` directory |
| M16 | `templates/{lang}/` subdirectory pattern SHALL be removed | GIVEN `adapters/common/base_adapter.go` → WHEN resolving templates → THEN path does not include lang segment |
| M17 | P7 references stripped from all adapter templates | GIVEN any adapter `skill.md.tmpl` → WHEN rendered → THEN roster lacks P7 row (P6 → M1); Phase 2 matrix lacks P7; health score lacks i18n category; delegation section lacks P7 |
