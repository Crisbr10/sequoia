# Delta Spec: agent-p7-i18n — REMOVED

## Action: REMOVED

P7 agent removed from roster, phase docs, and agent instructions.

| # | Requirement | Scenario |
|---|-------------|----------|
| R4 | Agent spec deleted | GIVEN `docs/agents/sequoia-i18n.md` → WHEN change applied → THEN file deleted |
| R5 | Phase doc deleted | GIVEN `docs/sequoia/sequoia-phases/07-i18n.md` → WHEN change applied → THEN file deleted |
| R6 | Roster purged | GIVEN any template or doc referencing P7 → WHEN rendered → THEN no P7 agent row exists |
