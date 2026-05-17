# Delta Spec: tui-pipeline — MODIFIED

## Action: MODIFIED

Previously: `RunInstall(ctx, tools, ch, lang)` and `RunUninstall(ctx, tools, ch, lang)` accepted `lang` parameter.

| # | Requirement | Scenario |
|---|-------------|----------|
| M9 | `RunInstall` / `RunUninstall` SHALL drop `lang` parameter | GIVEN pipeline started → WHEN `RunInstall(ctx, tools, ch)` runs → THEN progress labels English-only |
| M10 | Install/tool-select screens SHALL use English titles, summaries, counts | GIVEN progress or tool selection screen → WHEN rendered → THEN titles and instructions are hardcoded English |
