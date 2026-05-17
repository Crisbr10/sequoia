# Delta Spec: tui-management — MODIFIED

## Action: MODIFIED

Previously: Screens called `i18n.T()` with `lang` parameter for localized labels.

| # | Screen | Requirement |
|---|--------|-------------|
| M4 | Welcome | Menu options, subtitle, footer SHALL be hardcoded English |
| M5 | Status | Title, empty message, footer hints SHALL be English-only |
| M6 | Complete | Headings, item counts, try-command hint SHALL be English-only |
| M7 | Error | Headings SHALL be English-only |
| M8 | Uninstall | Title, confirm prompt, empty message SHALL be English-only |

**Scenario**: GIVEN any management screen → WHEN rendered → THEN all visible text is hardcoded English; no `i18n.T()` calls.
