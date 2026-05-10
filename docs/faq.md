# Frequently Asked Questions

## What AI tools are supported?

Sequoia supports five AI coding assistants:

| Tool | Adapter ID | Prompt Strategy |
|------|-----------|----------------|
| Claude Code | `claude-code` | Markdown section injection |
| OpenCode | `opencode` | File replace with backup |
| Cursor IDE | `cursor` | File replace with backup |
| Gemini CLI | `gemini-cli` | Config merge |
| OpenAI Codex | `codex` | TOML merge |

More tools are on the roadmap. If you'd like to add support for one, see
the [Contributing Guide](../CONTRIBUTING.md).

---

## How do I add support for a new tool?

You write a Go package that implements the `ToolAdapter` interface, create
templates, register it via `init()`, and add one import to `main.go`. The
CLI and TUI pick it up automatically.

The [Contributing Guide](../CONTRIBUTING.md) walks through every step with
code examples and a testing checklist. The Cursor adapter
(`adapters/cursor/`) is the cleanest reference implementation.

---

## Does Sequoia modify my existing config files?

It depends on the prompt strategy:

| Strategy | What Happens |
|----------|-------------|
| Markdown sections | A delimited section is injected between `<!-- sequoia:start -->` and `<!-- sequoia:end -->` markers. All existing content is preserved. |
| File replace | A backup (`.sequoia-backup`) is created before overwriting. The backup is restored on uninstall. |
| Config merge | Same as Markdown sections — marker-delimited content. |
| TOML merge | A `[sequoia]` table is merged in. All other tables and keys are untouched. |

In all cases, Sequoia is non-destructive. Your existing configuration is
never lost.

---

## Can I run audits without installing?

No. Sequoia's agents run as slash commands inside your AI tool. The installer
places the skill file, command definitions, and system prompt content that
the tool needs to understand and execute those commands.

However, you can read the skill file manually at
`docs/SKILL.md` to understand what each agent does. The CLI installer just
automates placing these files in the right directories.

---

## What's the difference between Claude Code and OpenCode installation?

| Aspect | Claude Code | OpenCode |
|--------|------------|----------|
| Config file | `~/.claude/CLAUDE.md` | `~/.config/opencode/AGENTS.md` |
| Strategy | Section injection (markers) | Full file replace (with backup) |
| Skills path | `~/.claude/skills/sequoia/` | `~/.config/opencode/skills/sequoia/` |
| Commands path | `~/.claude/commands/` | `~/.config/opencode/commands/` |

The behavioral result is identical — both tools get the same agents, commands,
and audit logic. The difference is only in how files are placed to match
each tool's directory conventions.

---

## How do I update Sequoia?

### CLI update

Run the one-line installer again — it detects your current version and
upgrades if needed:

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.sh | bash

# Windows
irm https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.ps1 | iex
```

### Skill update (after CLI update)

After updating the CLI binary, reinstall into your tools:

```bash
sequoia install --no-tui
```

This places the updated skill and command files. Your configuration is
preserved (the installer is idempotent).

---

## Is my code sent anywhere?

**No.** Sequoia runs entirely on your machine. The audit agents execute
inside your local AI tool. No code, findings, or project metadata leaves
your computer.

The only network calls Sequoia makes:
- The install scripts download the `sequoia` binary from GitHub Releases
- `sequoia install` reads templates embedded in the binary (no network)
- The audit commands run inside your AI tool, which may have its own
  network behavior (check your tool's privacy policy)

Sequoia itself never transmits your code.

---

## What's the Health Score?

The Health Score is a number from 0 to 100 that represents overall project
health. It's calculated as:

```
score = 100 - Σ(severity_weight × scope_multiplier)
```

| Severity | Weight |
|----------|--------|
| Critical | 15 |
| High | 8 |
| Medium | 4 |
| Low | 2 |
| Info | 0 |

The scope multiplier accounts for findings that share a root cause:
- **Isolated** (1.0): Finding affects a single location
- **Shared root cause** (1.5): Multiple findings originate from the same
  underlying issue

The score is broken down by audit phase (security, performance, architecture,
etc.) so you can see which areas need the most attention.

---

## Why the name "Sequoia"?

Sequoias are among the largest and longest-living trees on Earth. They grow
slowly, with deep root systems that intertwine underground — trees that
appear separate above ground are connected below.

The framework is named after them because:
- **Deep roots**: Every finding is traced to a real file and line
- **Connected**: M1 correlates findings across phases to find shared root causes
- **Enduring**: The goal is sustainable code health, not quick fixes

> *"A sequoia doesn't grow in haste. It grows with deep roots."*
