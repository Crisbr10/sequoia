# CLI Reference

The `sequoia` CLI installs, manages, and removes Sequoia from AI coding tools.

## Global Behavior

- All commands return exit code 0 on success, non-zero on error
- Help for any command: `sequoia help [command]` or `sequoia [command] --help`
- Version: `sequoia version`

---

## `sequoia install`

Install Sequoia into one or more supported AI tools.

```bash
sequoia install [--tool=<id>] [--no-tui]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tool` | `string` | `""` (all) | Target a specific adapter by ID |
| `--no-tui` | `bool` | `false` | Skip interactive TUI and install directly |

### Adapter IDs

| ID | Tool |
|----|------|
| `claude-code` | Claude Code |
| `opencode` | OpenCode |
| `cursor` | Cursor IDE |
| `gemini-cli` | Gemini CLI |
| `codex` | OpenAI Codex |

### Examples

```bash
# Interactive TUI — select tools, configure language/persistence, watch progress
sequoia install

# Headless — install into all detected tools without TUI
sequoia install --no-tui

# Headless — install into a specific tool only
sequoia install --tool=claude-code --no-tui

# Install into multiple specific tools (run once per tool)
sequoia install --tool=opencode --no-tui
```

### Behavior

- **TUI mode** (default, when stdin is a terminal): Launches the Bubbletea
  interactive interface. You select tools, pick language and persistence
  backend, and watch real-time progress.
- **Headless mode** (`--no-tui` or piped input): Installs into all
  detected tools (or the one specified by `--tool`). Progress is printed
  to stdout.

### What Gets Installed

For each selected tool:

1. **Skills**: A `SKILL.md` file in the tool's skills directory (containing
   all 9 agent definitions)
2. **Commands**: Five slash commands in the tool's commands directory
   (`sequoia-init`, `sequoia-audit`, `sequoia-review`, `sequoia-fix`,
   `sequoia-diff`)
3. **System prompt**: Section injection or file generation into the tool's
   configuration file (e.g. `CLAUDE.md`, `AGENTS.md`)
4. **Version marker**: `.sequoia-version` file for tracking

---

## `sequoia status`

Show installation status for all detected tools.

```bash
sequoia status
```

### Flags

None.

### Output

```
Tool            Detected    Sequoia     Version    Path
────            ────────    ───────     ───────    ────
Claude Code     ✓           ✓           0.1.0      ~/.claude/skills/sequoia
OpenCode        ✓           ✗           —          ~/.config/opencode/skills/sequoia
Cursor IDE      ✓           ✓           0.1.0      ~/.cursor/rules/sequoia-ai
Gemini CLI      ✗           ✗           —          —
OpenAI Codex    ✓           ✗           —          —
```

### Columns

| Column | Meaning |
|--------|---------|
| Tool | Human-readable tool name |
| Detected | `✓` if the tool is installed on your machine |
| Sequoia | `✓` if Sequoia has been installed for this tool |
| Version | Installed Sequoia version (or `—`) |
| Path | Installation root path (or `—`) |

---

## `sequoia uninstall`

Remove Sequoia from one or all tools.

```bash
sequoia uninstall [--tool=<id>] [--all] [--yes]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tool` | `string` | `""` | Target a specific adapter by ID |
| `--all` | `bool` | `false` | Remove from all installed tools |
| `--yes` / `-y` | `bool` | `false` | Skip confirmation prompt |

### Examples

```bash
# Remove from Claude Code only (prompts for confirmation)
sequoia uninstall --tool=claude-code

# Remove from all installed tools (prompts for confirmation)
sequoia uninstall --all

# Remove from all tools without prompting
sequoia uninstall --all --yes

# Remove from OpenCode via piped confirmation
echo "y" | sequoia uninstall --tool=opencode
```

### Behavior

- Backups created during installation are restored where applicable
  (e.g., `AGENTS.md` for OpenCode, `CLAUDE.md` section for Claude Code)
- If stdin is not a terminal and `--yes` is not passed, the command fails
- `--tool` and `--all` are mutually exclusive — exactly one must be specified
- Uninstalling a tool that never had Sequoia is a no-op

---

## `sequoia version`

Print the CLI version.

```bash
sequoia version
```

### Flags

None.

### Output

```
0.1.0
```

The version is embedded at build time via `-ldflags`. Development builds
show `0.1.0-dev`.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (invalid flags, adapter not found, permission denied) |
| 2 | Checksum verification failed (installer script) |
| 3 | Network error (installer script) |
