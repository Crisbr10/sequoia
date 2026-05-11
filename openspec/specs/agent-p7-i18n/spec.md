# agent-p7-i18n Specification

## Purpose

Define the P7 internationalization (i18n) audit agent — a new phase agent that inspects projects for i18n gaps: hardcoded strings, missing locale support, RTL readiness, and translation key consistency. P7 follows the existing agent spec pattern (YAML frontmatter → misión → methodology → checklists → calibración).

## Requirements

### Requirement: P7 Agent Document Structure
The P7 agent spec document (`docs/agents/sequoia-i18n.md`) SHALL follow the canonical agent pattern with YAML frontmatter, Spanish-language mission section, methodology decision trees, inspection checklists, and Calibración de Libertad. The frontmatter MUST define `name: sequoia-i18n`, a multiline description listing trigger keywords, and `tools: Read, Grep, Glob`.

#### Scenario: Agent spec validates against existing pattern
- GIVEN the existing `docs/agents/sequoia-quality.md` as reference
- WHEN `docs/agents/sequoia-i18n.md` is created
- THEN it SHALL contain a YAML frontmatter block with `name`, `description`, and `tools`
- AND it SHALL contain sections: `## Misión`, at least one methodology decision tree, `## Calibración de Libertad`

### Requirement: Hardcoded String Detection
P7 MUST detect hardcoded user-facing strings in source code that lack i18n framework calls. The agent SHALL distinguish between developer-facing strings (log messages, comments) and user-facing strings (UI labels, error messages shown to users).

#### Scenario: Detects hardcoded strings in JSX
- GIVEN a React project with `<button>Submit</button>` and no i18n library usage
- WHEN P7 audits the codebase
- THEN a finding SHALL be produced for the hardcoded "Submit" string
- AND the finding SHALL cite the file path and line number

#### Scenario: Ignores developer-facing strings
- GIVEN a Go project with `log.Printf("processing request")`
- WHEN P7 audits the codebase
- THEN no i18n finding SHALL be raised for the log message
- AND developer-facing strings SHALL be classified as out-of-scope

### Requirement: Locale Formatting Verification
P7 MUST verify that dates, numbers, and currencies use locale-aware formatting. The agent SHALL detect uses of `toLocaleDateString()`, `Intl.NumberFormat`, and equivalent locale-aware APIs versus hardcoded format strings.

#### Scenario: Detects missing locale in date formatting
- GIVEN code using `new Date().toLocaleDateString()` without a locale argument
- WHEN P7 scans the codebase
- THEN a low-severity finding SHALL note the implicit locale dependency

### Requirement: RTL Support Detection
P7 SHALL detect whether the project supports right-to-left (RTL) languages. The agent MUST check for CSS logical properties, `dir` attributes, and RTL-specific style overrides.

#### Scenario: Reports absence of RTL support
- GIVEN a web project with no `dir` attributes and no CSS logical properties
- WHEN P7 audits for i18n
- THEN an informational finding SHALL note RTL support gaps
- AND the finding SHALL recommend CSS logical properties (`margin-inline-start` over `margin-left`)

### Requirement: Translation Key Consistency
P7 SHALL verify that translation keys in i18n resource files are consistent across all locale files. Keys present in one locale file but missing in another MUST be reported.

#### Scenario: Detects missing translation key
- GIVEN `en.json` contains `"greeting": "Hello"` but `es.json` lacks the `"greeting"` key
- WHEN P7 cross-references locale files
- THEN a finding SHALL report the missing key in `es.json`
