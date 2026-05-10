# Sequoia Documentation

> Multi-agent code audit framework — install once, audit everywhere.

## Quick Links

| Document | Description |
|----------|-------------|
| [Getting Started](getting-started.md) | 5-minute guide to your first audit |
| [Architecture](architecture.md) | Design overview — adapters, installer, TUI, CLI |
| [CLI Reference](cli-reference.md) | All `sequoia` commands and flags |
| [FAQ](faq.md) | Frequently asked questions |
| [Development Plan](DEVELOPMENT-PLAN.md) | Full task breakdown and project status |
| [Contributing Guide](../CONTRIBUTING.md) | How to add a new AI tool adapter |
| [Release Notes](release-notes/) | Version history |

## Reference Documents

| Document | Description |
|----------|-------------|
| [Scoring Criteria](references/scoring-criteria.md) | Health Score formula and severity weights |
| [Project Map Schema](references/project-map.md) | Project Map YAML reference with chunking |
| [Finding Format](references/finding-format.md) | Audit finding structure and fields |
| [Phase Template](references/phase-template.md) | Template for new audit phases |

## Audit Flows

| Document | Description |
|----------|-------------|
| [Full Audit Flow](flows/full-audit-flow.md) | Complete project audit from init to report |
| [PR Review Flow](flows/pr-review-flow.md) | Diff-based audit for pull requests |
| [Incremental Flow](flows/incremental-flow.md) | Partial re-audit of changed files |
| [Simple Project Flow](flows/simple-project-flow.md) | Lightweight audit for small repos |

## Framework Specification

| Document | Description |
|----------|-------------|
| [SKILL.md](SKILL.md) | The Sequoia skill definition (what the AI loads) |
| [SEQUOIA.md](SEQUOIA.md) | Framework philosophy and design principles |
| [Integration Plan](INTEGRATION-PLAN.md) | How Sequoia integrates into AI tools |

## Contributing

Want to add support for a new AI tool? Start with the
[Contributing Guide](../CONTRIBUTING.md). It walks through the adapter pattern,
prompt strategies, testing checklist, and PR process.
