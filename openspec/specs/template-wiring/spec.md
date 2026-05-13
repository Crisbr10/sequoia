# template-wiring Specification

## Purpose

Define the template updates required to wire P7 (i18n), P4 deep dependency scanning, and P3 resilience patterns into all 5 adapter template files and the AGENTS.md section template. Also define the language-aware template resolution mechanism (`RenderTemplateLang`) that routes to language-specific `.tmpl` variants.

## Requirements

### Requirement: Language-Aware Template Resolution
`adapters/common/template.go` MUST provide `RenderTemplateLang(fs embed.FS, name, lang string, data interface{}) (string, error)`. The function SHALL resolve `{name}.{lang}.tmpl` first, falling back to `{name}.tmpl` when the language-specific template file does not exist in the embedded FS. Adapter `Install()` methods SHALL call `RenderTemplateLang` instead of `RenderTemplate` for skill and system prompt templates.

#### Scenario: Language-specific template exists
- GIVEN `name = "skill.md"`, `lang = "en"`, and `skill.md.en.tmpl` in the FS
- WHEN `RenderTemplateLang(fs, "skill.md", "en", data)` is called
- THEN `skill.md.en.tmpl` SHALL be loaded and rendered

#### Scenario: Language file missing — fallback
- GIVEN `name = "skill.md"`, `lang = "es"`, and NO `skill.md.es.tmpl` in the FS
- WHEN `RenderTemplateLang(fs, "skill.md", "es", data)` is called
- THEN `skill.md.tmpl` SHALL be loaded as fallback
- AND no error SHALL occur

#### Scenario: Existing templates unaffected
- GIVEN existing adapters with only `skill.md.tmpl` and no language-specific variants
- WHEN `RenderTemplateLang(fs, "skill.md", "en", data)` is called
- THEN `skill.md.tmpl` SHALL be used (backward compatible)
- AND existing behavior SHALL be preserved

### Requirement: P7 Agent Roster Entry in All Templates
All 5 adapter `skill.md.tmpl` files and the OpenCode `agents-md-section.md.tmpl` MUST include P7 in their agent roster tables. The P7 entry SHALL follow the existing row format: `| P7 i18n | Hardcoded strings, locale formatting, RTL, translation keys | Always |`.

#### Scenario: P7 present in all roster tables
- GIVEN the 5 adapter template files and the AGENTS.md section template
- WHEN templates are rendered with `{{.Version}}`
- THEN each roster table SHALL contain exactly one P7 row
- AND the P7 row SHALL be placed after P6 Operations and before M1 Correlator

### Requirement: Phase 2 Selection Matrix Includes P7
The Phase 2 agent selection table in opencode, gemini, and claude templates MUST include a P7 row. P7 SHALL be triggered for all project types (`**Siempre**`).

#### Scenario: P7 triggers on all project types
- GIVEN the Phase 2 selection table in the opencode template
- WHEN the table is rendered
- THEN row `P7 (i18n)` SHALL map to `**Siempre**`

### Requirement: Health Score Includes i18n Category
All templates that define the Health Score categories MUST include `i18n` in the category list. The `health_score.categories` block SHALL have an `i18n: [0-100]` entry.

#### Scenario: Health score renders i18n category
- GIVEN the health_score block in opencode and gemini templates
- WHEN the block is rendered
- THEN `i18n: [0-100]` SHALL appear between `operations` and any trailing category

### Requirement: Template Description Mentions i18n
The YAML frontmatter `description` field SHALL be updated to mention i18n among the specialized angles. The phrase SHALL change from "security, performance, architecture, quality, UX, and operations" to include "internationalization".

#### Scenario: Description includes i18n
- GIVEN any adapter template frontmatter
- WHEN the `description` field is rendered
- THEN it SHALL contain the word "internationalization"

### Requirement: P7 Section in Agent Delegation Templates
OpenCode and gemini templates (full-length variants) SHALL include a P7 delegation section following the existing agent section pattern. The section SHALL define misión, methodology, and calibración for the i18n domain.

#### Scenario: P7 section renders in full template
- GIVEN the opencode template (1931 lines)
- WHEN the complete skill document is rendered
- THEN a `# Sequoia i18n — P7` section SHALL exist
- AND it SHALL contain `## Misión` and `## Calibración de Libertad`
