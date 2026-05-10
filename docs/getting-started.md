# Getting Started

This guide takes you from zero to your first Sequoia audit in 5 minutes.

## Overview

Sequoia is a multi-agent code audit framework. It deploys specialized AI agents
that inspect your project from every angle — security, performance,
architecture, quality, UX, and operations — all in parallel.

Sequoia runs inside your AI coding assistant as slash commands:
- `/sequoia init` — Analyze project context, generate a Project Map
- `/sequoia audit` — Run a full parallel audit
- `/sequoia review` — Review a PR or diff
- `/sequoia fix` — Generate a task list from audit findings
- `/sequoia diff` — Compare against your last audit

## Step 1: Install the CLI

The `sequoia` CLI installs Sequoia's files into your AI tools. Choose your OS:

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.sh | bash
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.ps1 | iex
```

The installer downloads the right binary for your OS and architecture,
verifies its SHA-256 checksum, and runs `sequoia install --no-tui`.

> **Already have the binary?** Run `sequoia install` for the interactive TUI,
> or `sequoia install --no-tui` for headless mode.

### Verify Installation

```bash
sequoia status
```

You should see a table of detected AI tools and their installation status.

## Step 2: Initialize Your Project

Open your project in any supported AI tool and run:

```
/sequoia init
```

This scans your codebase — language, framework, structure, dependencies —
and generates a Project Map. The map tells Sequoia's agents what they're
working with, so their findings are always contextual.

## Step 3: Run Your First Audit

```
/sequoia audit
```

Sequoia activates up to 6 specialized agents in parallel:
1. **P1 — Security**: Auth, injection, secrets, attack surface
2. **P2 — Performance**: Queries, caching, bundle size, Core Web Vitals
3. **P3 — Architecture**: Patterns, coupling, API design, tech debt
4. **P4 — Quality**: Test coverage, naming, dead code, documentation
5. **P5 — Experience**: Accessibility, UX patterns, conversion flows
6. **P6 — Operations**: CI/CD, observability, containerization

Agents that don't apply to your stack are skipped automatically.

## Step 4: Read the Report

The audit produces:
- **Health Score** (0–100) broken down by category
- **Findings** — each traced to a specific file and line
- **Root cause analysis** — when multiple symptoms share one cause
- **Prioritized action plan** — what to fix, in what order, with effort estimates

To address findings, run:

```
/sequoia fix
```

This generates a concrete task list from the audit report. Pick up tasks one
by one in your editor.

## What's Next?

- [Architecture overview](architecture.md) — understand how Sequoia works under the hood
- [CLI reference](cli-reference.md) — all `sequoia` commands and flags
- [FAQ](faq.md) — common questions and answers
- [Contributing guide](../CONTRIBUTING.md) — add support for a new AI tool

## Supported Tools

| Tool | Install Path |
|------|-------------|
| Claude Code | `~/.claude/` |
| OpenCode | `~/.config/opencode/` |
| Cursor IDE | `~/.cursor/rules/` |
| Gemini CLI | `~/.gemini/` |
| OpenAI Codex | `~/.codex/` |
