<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/icons8-árbol-conífero-64.png">
    <img src="docs/icons8-árbol-conífero-64.png" alt="Sequoia" width="80">
  </picture>
</p>

<h1 align="center">Sequoia</h1>
<p align="center"><strong>Multi-Agent Code Audit Framework</strong></p>

<p align="center">
  <img src="https://img.shields.io/badge/go-1.22+-00ADD8?style=flat&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="MIT">
  <a href="https://github.com/features/actions"><img src="https://img.shields.io/badge/CI-GitHub_Actions-2088FF?style=flat&logo=githubactions" alt="CI"></a>
</p>

---

> *"A sequoia doesn't grow in haste. It grows with deep roots."*

Sequoia is a **multi-agent code audit framework** that deploys specialized AI agents to inspect a project from every angle — security, performance, architecture, quality, UX, and operations — in parallel. Every finding is traced to a real file, a real line, or a documented absence. No generic advice. No hallucinated code.

## How It Works

```
┌──────────────────────────┐
│   C0 · Orchestrator      │
│   Detects context,        │
│   coordinates agents,     │
│   synthesizes results     │
└──────┬───────────────────┘
       │
 ┌─────┼─────┬─────────────┐
 ▼     ▼     ▼             ▼
P1    P2    P3 ... P6     M1 · M2
```

1. **C0 scans your project** — detects the stack, size, and conventions.
2. **Phase agents (P1–P6) audit in parallel** — each owns a single domain. Security doesn't guess about performance. Architecture doesn't guess about UX.
3. **M1 correlates findings** — when five symptoms share one root cause, you get one fix, not five tickets.
4. **M2 produces the report** — a Health Score (0–100), a prioritized action plan, and actionable tasks.

Sequoia integrates directly into your AI coding assistant — **Claude Code** and **OpenCode** — so auditing is a slash command away:

```bash
/sequoia init          # Analyze project context
/sequoia audit         # Full parallel audit
/sequoia review        # PR / diff review
/sequoia fix           # Generate task list from findings
/sequoia diff          # Compare against last audit
```

## Agents

| ID | Agent | Domain | Runs when |
|----|-------|--------|-----------|
| C0 | Orchestrator | Project map, coordination, synthesis | Always |
| P1 | Security | Auth, injection, secrets, attack surface | Always |
| P2 | Performance | Bundle size, queries, caching, Core Web Vitals | Frontend / backend / fullstack |
| P3 | Architecture | Patterns, coupling, API design, tech debt | Always |
| P4 | Quality | Test coverage, naming, dead code, documentation | Always |
| P5 | Experience | Accessibility, UX patterns, conversion flows | Frontend / fullstack / mobile |
| P6 | Operations | CI/CD, containerization, observability, IaC | Backend / fullstack / infra |
| M1 | Correlator | Cross-phase deduplication and root cause analysis | Always |
| M2 | Reporter | Health Score, deliverables | Always |

## CLI Installer

This repository contains the Go CLI that installs Sequoia into your tools. It handles the file placement, template rendering, and prompt injection so you don't have to.

```bash
# Install Sequoia into all detected AI tools (interactive TUI)
sequoia install

# Headless install into a specific tool
sequoia install --tool=claude-code --no-tui

# Check installation status
sequoia status

# Remove Sequoia (with confirmation)
sequoia uninstall --all
```

### Supported Tools

| Tool | Prompt Strategy | Config File |
|------|----------------|-------------|
| Claude Code | Section injection (markers) | `~/.claude/CLAUDE.md` |
| OpenCode | File replace with backup | `~/.config/opencode/AGENTS.md` |

More adapters (Gemini CLI, Continue, Cursor) are on the roadmap.

## Project Structure

```
sequoia-ai/
├── cmd/sequoia/              # Cobra CLI entrypoint
├── adapters/                 # ToolAdapter interface + implementations
│   ├── interface.go          # Contract every adapter satisfies
│   ├── registry.go           # Plugin registry (database/sql pattern)
│   ├── factory.go            # NewAdapter(id) constructor
│   ├── common/               # Shared installer (Prepare → Apply → Verify → Rollback)
│   ├── claude/                # Claude Code adapter
│   │   └── templates/        # SKILL.md, commands, CLAUDE.md section
│   ├── opencode/              # OpenCode adapter
│   │   └── templates/        # SKILL.md, commands, AGENTS.md section
│   └── _template/            # Adapter scaffolding reference
├── internal/                 # Private packages
│   ├── app/                  # Bubbletea model, update, view
│   ├── tui/screens/          # Welcome, Tool Selection, Install Progress, etc.
│   ├── model/                # Domain types
│   └── pipeline/             # Installation pipeline orchestration
├── scripts/                  # One-line installers (curl | bash, irm | iex)
├── docs/                     # Framework specification and design docs
└── openspec/                 # Artifacts (proposals, specs, verify reports)
```

## Development Status

| Phase | Goal | Status |
|-------|------|--------|
| 1 — Foundation | Specs, Go module, adapter interface, common installer | ✅ Done |
| 2 — Claude Code | Full install pipeline for `~/.claude/` | 🚧 In progress |
| 3 — OpenCode | Full install pipeline for `~/.config/opencode/` | 🚧 In progress |
| 4 — CLI Installer | Headless `sequoia` binary with Cobra | ✅ Done |
| 5 — TUI Installer | Interactive Bubbletea interface | 📋 Planned |
| 6 — Distribution | GoReleaser, Homebrew, CI/CD | 📋 Planned |

Full details in [`docs/DEVELOPMENT-PLAN.md`](docs/DEVELOPMENT-PLAN.md).

## Philosophy

| Principle | Meaning |
|-----------|---------|
| **Evidence over opinion** | Every finding cites specific code. Not "this looks wrong" — "this pattern in `auth/middleware.ts:42` allows validation bypass." |
| **Context over dogma** | Rules adapt to the detected stack. A monolith is not judged by microservice standards. |
| **Root cause over symptom** | When five findings share one cause, you get one fix — not five issues. |
| **Actionable, always** | If a finding doesn't include *what* to change, *where*, and *why*, it isn't emitted. |
| **Prioritizable debt** | Every finding is scored by severity × impact × effort. The team decides with data. |

## Rules

1. Every finding requires evidence. No file + line reference → no finding.
2. Never emit generic advice. "Consider using HTTPS" is not acceptable.
3. Root causes are correlated across phases.
4. The Health Score is mandatory — 0 to 100, broken down by category.
5. No change is suggested without trade-off analysis.
6. Agents don't invent code that doesn't exist.
7. Severity is calibrated to the project — a typo in a deploy script is not the same as a typo in production auth.
8. Findings are actionable or they are removed before the report.
9. Findings are never duplicated between agents. M1 deduplicates.
10. The final report includes a prioritized action plan with effort estimates.
11. Dependencies are audited by real risk, not age.
12. Sequoia analyzes and reports. It never modifies code.

## License

MIT — see [LICENSE](LICENSE).

---

<p align="center"><em>Deep roots. Precise findings. Concrete action.</em></p>
