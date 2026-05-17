# Delta Spec: skill-documentation — ADDED

## Action: ADDED

All ~30 Sequoia `.md` files SHALL be rewritten in native English with no Spanish text.

| # | Scope | Requirement |
|---|-------|-------------|
| N1 | Agent instructions | `docs/agents/*.md` SHALL contain only English |
| N2 | Phase docs | `docs/sequoia/sequoia-phases/*.md` SHALL be English-only (07-i18n.md deleted) |
| N3 | Commands, flows, references | All reference docs SHALL use only English |
| N4 | CLI templates | `adapters/common/templates/commands/*.tmpl` SHALL use only English |

**Scenario**: GIVEN any doc under `docs/` or `adapters/*/templates/` → WHEN inspected → THEN no Spanish text present; all content in English.
