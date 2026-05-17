# template-wiring Specification

## Purpose

Define the template resolution mechanism and adapter template structure. All language-aware template resolution has been removed. P7 (i18n) references have been stripped from all adapter templates.

## Requirements

### Requirement: Direct Template Resolution

`adapters/common/template.go` MUST provide `RenderTemplate(fs embed.FS, name string, data interface{}) (string, error)`. The function SHALL resolve templates from a single `templates/` directory with no language subdirectory resolution. `RenderTemplateLang` SHALL be deleted.

#### Scenario: Template resolved directly
- GIVEN `name = "templates/skill.md.tmpl"` exists in the embedded FS
- WHEN `RenderTemplate(fs, "templates/skill.md.tmpl", data)` is called
- THEN the template SHALL be loaded and rendered from the single directory

#### Scenario: No language subdirectory used
- GIVEN `adapters/common/base_adapter.go`
- WHEN resolving templates
- THEN the path SHALL NOT include a language segment

### Requirement: No P7 Agent References in Templates

All adapter template files (`skill.md.tmpl`, `rules.md.tmpl`, `*-section.md.tmpl`) SHALL NOT contain P7 (i18n) agent references. The agent roster, Phase 2 selection matrix, health score categories, delegation sections, and frontmatter descriptions SHALL all be stripped of P7/i18n entries.

#### Scenario: Roster lacks P7 row
- GIVEN any adapter `skill.md.tmpl`
- WHEN rendered
- THEN the agent roster table SHALL NOT contain a P7 row
- AND P6 Operations SHALL be followed directly by M1 Correlator

#### Scenario: Phase 2 matrix lacks P7
- GIVEN the Phase 2 selection table in any adapter template
- WHEN rendered
- THEN no P7 (i18n) row exists

#### Scenario: Health score lacks i18n category
- GIVEN the health score section in any template
- WHEN rendered
- THEN no `i18n` category exists

#### Scenario: Frontmatter lacks i18n mention
- GIVEN any adapter template frontmatter
- WHEN the `description` field is rendered
- THEN it SHALL NOT contain "internationalization" or "i18n"

### Requirement: No P7 Delegation Section

Templates SHALL NOT include a P7 (i18n) delegation section. The full agent delegation section for i18n SHALL be removed from all templates.

#### Scenario: No P7 delegation section
- GIVEN the opencode template
- WHEN the complete skill document is rendered
- THEN no `# Sequoia i18n — P7` section exists
