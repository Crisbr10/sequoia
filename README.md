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
  <img src="https://img.shields.io/badge/version-0.1.0-blue" alt="Version">
  <a href="https://github.com/Crisbr10/sequoia/actions/workflows/ci.yml"><img src="https://github.com/Crisbr10/sequoia/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
</p>

---

> *"A sequoia doesn't grow in haste. It grows with deep roots."*

Sequoia is a **multi-agent code audit framework** that deploys specialized AI agents to inspect a project from every angle — security, performance, architecture, quality, UX, and operations — in parallel. Every finding is traced to a real file, a real line, or a documented absence. No generic advice. No hallucinated code.

## Quick Start

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.sh | bash

# Windows (PowerShell)
irm https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.ps1 | iex
```

Then open your project in any supported AI tool and run:

```
/sequoia init          # Analyze your project
/sequoia audit         # Full parallel audit
```

[Full getting started guide →](docs/getting-started.md)

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

## Features

- **9 specialized agents** — security, performance, architecture, quality, UX, operations, plus correlation and reporting
- **5 AI tool adapters** — Claude Code, OpenCode, Cursor IDE, Gemini CLI, OpenAI Codex
- **Interactive TUI installer** — multi-select tools, real-time progress, error recovery
- **Headless CLI** — script-friendly mode for CI/CD and automation
- **Cross-platform** — macOS, Linux, Windows (amd64 + arm64)
- **Atomic installation** — Prepare → Apply → Verify → Rollback pipeline; idempotent
- **Four prompt strategies** — adapts to each tool's config format
- **Plugin-ready** — `ToolAdapter` interface with self-registration pattern
- **Strict TDD suite** — 327+ tests, 90%+ coverage

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

## Slash Commands

Sequoia integrates directly into your AI coding assistant:

```bash
/sequoia init          # Analyze project context
/sequoia audit         # Full parallel audit
/sequoia review        # PR / diff review
/sequoia fix           # Generate task list from findings
/sequoia diff          # Compare against last audit
```

## CLI Installer

The Go CLI installs Sequoia into your tools. It handles file placement, template rendering, and prompt injection.

```bash
# Interactive TUI — select tools, configure, watch progress
sequoia install

# Headless — install into all detected tools
sequoia install --no-tui

# Install into a specific tool
sequoia install --tool=claude-code --no-tui

# Check installation status
sequoia status

# Remove Sequoia
sequoia uninstall --all
```

### Supported Tools

| Tool | Adapter ID | Prompt Strategy | Config File |
|------|-----------|----------------|-------------|
| Claude Code | `claude-code` | Markdown section injection | `~/.claude/CLAUDE.md` |
| OpenCode | `opencode` | File replace with backup | `~/.config/opencode/AGENTS.md` |
| Cursor IDE | `cursor` | File replace with backup | `~/.cursor/rules/sequoia-ai.md` |
| Gemini CLI | `gemini-cli` | Config merge | `GEMINI.md` |
| OpenAI Codex | `codex` | TOML merge | `~/.codex/config.toml` |

## Verifying Binaries

All release binaries are signed with [cosign](https://github.com/sigstore/cosign) keyless signing using GitHub Actions OIDC. Download the `.sig` and `.pem` files alongside the binary, then verify:

```bash
cosign verify-blob \
  --signature sequoia-windows-amd64.exe.sig \
  --certificate sequoia-windows-amd64.exe.pem \
  --certificate-identity "https://github.com/Crisbr10/sequoia/.github/workflows/release.yml@refs/tags/v*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  sequoia-windows-amd64.exe
```

This confirms the binary was produced by the official Sequoia CI pipeline on GitHub Actions, not tampered with by a third party.

## Architecture

```
                          sequoia CLI
                         ┌─────────┐
                         │  main() │
                         └────┬────┘
                              │
             ┌────────────────┼────────────────┐
             ▼                ▼                ▼
        ┌─────────┐    ┌──────────┐    ┌─────────────┐
        │  Cobra  │    │Bubbletea │    │   Adapter   │
        │  CLI    │    │   TUI    │    │  Registry   │
        └────┬────┘    └────┬─────┘    └──────┬──────┘
             │              │                  │
             └──────────────┼──────────────────┘
                            │
                   ┌────────┴────────┐
                   │  ToolAdapter    │
                   │  (interface)    │
                   └────────┬────────┘
                            │
      ┌─────────┬───────────┼───────────┬──────────┐
      ▼         ▼           ▼           ▼          ▼
   Claude    OpenCode    Cursor      Gemini     Codex
```

[Full architecture docs →](docs/architecture.md)

## Project Structure

```
sequoia/
├── cmd/sequoia/              # Cobra CLI entrypoint
├── adapters/                 # ToolAdapter interface + implementations
│   ├── interface.go          # Contract every adapter satisfies
│   ├── registry.go           # Plugin registry (database/sql pattern)
│   ├── factory.go            # NewAdapter(id) constructor
│   ├── common/               # Shared installer framework
│   ├── claude/               # Claude Code adapter + templates
│   ├── opencode/             # OpenCode adapter + templates
│   ├── cursor/               # Cursor IDE adapter + templates
│   ├── gemini/               # Gemini CLI adapter + templates
│   ├── codex/                # OpenAI Codex adapter + templates
│   └── _template/            # Adapter scaffolding reference
├── internal/                 # Private packages
│   ├── app/                  # Bubbletea model, update, view
│   ├── tui/screens/          # 8 TUI screens
│   ├── model/                # Domain types
│   └── pipeline/             # Installation pipeline
├── scripts/                  # One-line installers
├── docs/                     # Documentation
├── .goreleaser.yaml          # GoReleaser config
└── .golangci.yaml            # Linter config
```

## Documentation

| Document | Description |
|----------|-------------|
| [Getting Started](docs/getting-started.md) | 5-minute guide to your first audit |
| [Architecture](docs/architecture.md) | Design overview |
| [CLI Reference](docs/cli-reference.md) | All commands and flags |
| [FAQ](docs/faq.md) | Frequently asked questions |
| [Development Plan](docs/DEVELOPMENT-PLAN.md) | Full task breakdown |
| [Contributing Guide](CONTRIBUTING.md) | How to add a new adapter |
| [Release Notes](docs/release-notes/) | Version history |

## Development Status

| Phase | Goal | Status |
|-------|------|--------|
| 1 — Foundation | Specs, Go module, adapter interface, common installer | ✅ Done |
| 2 — Claude Code | Full install pipeline for `~/.claude/` | ✅ Done |
| 3 — OpenCode | Full install pipeline for `~/.config/opencode/` | ✅ Done |
| 4 — CLI Installer | Headless `sequoia` binary with Cobra | ✅ Done |
| 5 — TUI Installer | Interactive Bubbletea interface | ✅ Done |
| 6 — Extensibility & Release | More adapters, docs, GoReleaser, v0.1.0 | ✅ Done |

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
