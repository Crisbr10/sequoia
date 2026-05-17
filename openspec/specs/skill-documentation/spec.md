# skill-documentation Specification

## Purpose

All Sequoia `.md` documentation files shall be written in native English with no Spanish text, ensuring maximum AI agent comprehension.

## Requirements

### Requirement: Agent Docs English-Only

`docs/agents/*.md` SHALL contain only English. No Spanish text, characters, or mixed-language content.

#### Scenario: No Spanish in agent docs
- GIVEN any file under `docs/agents/`
- WHEN inspected for Spanish characters (áéíóúñü)
- THEN zero matches found

### Requirement: Phase Docs English-Only

`docs/sequoia/sequoia-phases/*.md` SHALL be English-only. The `07-i18n.md` phase doc SHALL be deleted.

#### Scenario: Phase docs clean
- GIVEN `docs/sequoia/sequoia-phases/`
- WHEN inspected
- THEN no Spanish text present; `07-i18n.md` absent

### Requirement: Reference Docs English-Only

All reference documentation (commands, flows, guides) SHALL use only English.

#### Scenario: Commands docs clean
- GIVEN any reference doc under `docs/`
- WHEN inspected for Spanish characters
- THEN zero matches found

### Requirement: CLI Templates English-Only

`adapters/common/templates/commands/*.tmpl` SHALL use only English.

#### Scenario: CLI templates clean
- GIVEN any command template
- WHEN rendered
- THEN all content is English; no Spanish fragments
