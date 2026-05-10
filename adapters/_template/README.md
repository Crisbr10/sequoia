# Adapter Template

Copy-paste boilerplate for adding a new tool integration to Sequoia.

## Quick Start

1. **Copy** this entire directory:
   ```bash
   cp -r adapters/_template adapters/my-tool
   ```

2. **Replace** every `TODO` comment across all `.go` files:
   - `adapter.go` — adapter ID, name, detection logic, prompt strategy
   - `paths.go` — config directory path, system prompt filename
   - `installer.go` — system prompt injection logic
   - `install.go` — version constant, command file list

3. **Create** the remaining command templates:
   ```bash
   cp templates/commands/sequoia-init.md templates/commands/sequoia-audit.md
   cp templates/commands/sequoia-init.md templates/commands/sequoia-review.md
   cp templates/commands/sequoia-init.md templates/commands/sequoia-fix.md
   cp templates/commands/sequoia-init.md templates/commands/sequoia-diff.md
   ```
   Then edit each command file with tool-specific content.

4. **Register** the adapter in `cmd/sequoia/main.go`:
   ```go
   import (
       // ... existing imports ...
       _ "sequoia-ai/adapters/my-tool"
   )
   ```

5. **Write tests** using the checklist in `CONTRIBUTING.md`.

## File Reference

| File | Purpose | What to Replace |
|------|---------|-----------------|
| `adapter.go` | ToolAdapter implementation | ID, Name, Detect logic, PromptStrategy |
| `paths.go` | OS-specific path helpers | Config directory, system prompt filename |
| `installer.go` | System prompt injection/removal | Strategy (marker section, file replace, etc.) |
| `install.go` | Install pipeline + helpers | Version, command file list, template data |
| `embed.go` | Embeds templates into binary | Nothing — works as-is |
| `templates/skill.md.tmpl` | SKILL.md template | Agent roster content |
| `templates/rules.md.tmpl` | System prompt content | Tool-specific rules/sections |
| `templates/commands/` | Command files (5 total) | Tool-specific command descriptions |

## Testing Checklist

See `CONTRIBUTING.md` for the complete testing checklist. Every adapter must
have tests for:

- Paths (skills, commands, system prompt, version file)
- Interface (ID, name, strategy, detect, status)
- Install (creates files, idempotent, preserves user content)
- Uninstall (removes files, restores backups, safe on missing)

## Reference Implementations

- **Cursor** (`adapters/cursor/`) — cleanest, most up-to-date reference
- **Claude Code** (`adapters/claude/`) — StrategyMarkdownSections example
- **OpenCode** (`adapters/opencode/`) — StrategyFileReplace example
- **Gemini CLI** (`adapters/gemini/`) — StrategyConfigMerge example
- **Codex** (`adapters/codex/`) — StrategyTOMLMerge example

## Need Help?

Read `CONTRIBUTING.md` at the repo root for the full step-by-step adapter
development guide.
