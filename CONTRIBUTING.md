# Contributing to Sequoia

Welcome! Sequoia is a comprehensive code audit framework that integrates into
AI-assisted coding tools (Claude Code, OpenCode, Cursor IDE, Gemini CLI, and
OpenAI Codex). This guide explains how to add support for a new tool.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture Overview](#architecture-overview)
3. [Adding a New Adapter](#adding-a-new-adapter)
4. [Testing Checklist](#testing-checklist)
5. [Prompt Strategies](#prompt-strategies)
6. [Pull Request Process](#pull-request-process)
7. [Code Conventions](#code-conventions)

---

## Project Overview

Sequoia provides a unified auditing experience across multiple AI coding
assistants. Each tool gets the same agents, commands, and auditing power
through a shared interface (`ToolAdapter`). The framework includes:

- **CLI installer** (`sequoia install|status|uninstall|version`) via Cobra
- **Interactive TUI** via Bubbletea (install/config/status only)
- **Adapter registry** modelled after `database/sql` self-registration
- **Common installer** with Prepare → Apply → Verify → Rollback lifecycle

Adding a new adapter means Sequoia can install its skills, commands, and
system prompt into a new AI tool's configuration directory.

## Architecture Overview

### ToolAdapter Interface

Every adapter must implement the `ToolAdapter` interface defined in
`adapters/interface.go`:

```
ID()              → machine-readable ID (e.g. "claude-code")
Name()            → human-readable display name
Detect()          → is the tool installed on this machine?
IsInstalled()     → is Sequoia already installed for this tool?
Install()         → install Sequoia files
Uninstall()       → remove Sequoia files
Status()          → current installation state
SkillsPath()      → path to skills directory
CommandsPath()    → path to commands directory
SystemPromptPath()→ path to system prompt file
PromptStrategy()  → injection strategy
```

### Registry

Adapters self-register via `init()` following the `database/sql` pattern.
See `adapters/registry.go` — `Register()` adds adapters, `Get(id)` retrieves
them, `All()` returns everything in registration order.

### Common Installer

The `common.Installer` framework (`adapters/common/installer.go`) provides
a four-phase lifecycle shared by all adapters:

```
Prepare → validate paths, check permissions, back up existing files
Apply   → copy files from staging to target
Verify  → confirm all expected files exist and are readable
Rollback→ restore backups and clean up on any failure
```

### Prompt Strategies

See the [Prompt Strategies](#prompt-strategies) section below for details on
the four strategies and when to use each one.

---

## Adding a New Adapter

Follow these steps in order. Each step references the existing adapters as
examples — the Cursor adapter (`adapters/cursor/`) is the cleanest reference.

### Step 1: Create the Adapter Directory

```bash
mkdir -p adapters/{tool}/templates/commands
```

Replace `{tool}` with a short lowercase identifier (e.g. `claude`, `opencode`,
`cursor`, `gemini`, `codex`).

### Step 2: Implement `adapter.go`

Create `adapters/{tool}/adapter.go`. It must:

1. Implement the full `ToolAdapter` interface
2. Have an `init()` function that calls `adapters.DefaultRegistry.Register()`
3. Accept a `homeDir string` field for testability (use `os.UserHomeDir()`
   when empty)
4. Provide a `NewAdapter(homeDir string)` constructor for tests

```go
package {tool}

import (
    "os"
    "os/exec"
    "path/filepath"
    "sequoia-ai/adapters"
)

type Adapter struct {
    homeDir string
}

func init() {
    adapters.DefaultRegistry.Register(&Adapter{})
}

func NewAdapter(homeDir string) *Adapter {
    return &Adapter{homeDir: homeDir}
}
```

Follow the Cursor adapter (`adapters/cursor/adapter.go`) as your primary
reference. Copy its structure exactly, replacing tool-specific paths and
detection logic.

### Step 3: Implement `paths.go`

Create `adapters/{tool}/paths.go` with OS-specific path helpers.

**Rules:**
- Use `filepath.Join()` for ALL path construction — never `/` or `\\`
- Define a `{tool}Base(homeDir)` function that returns the root config directory
- Define helpers for skills, commands, system prompt, backups, and version file
- Resolve symlinks with `filepath.EvalSymlinks()` before joining

```go
package {tool}

import (
    "os"
    "path/filepath"
)

func {tool}Base(homeDir string) (string, error) {
    if homeDir == "" {
        var err error
        homeDir, err = os.UserHomeDir()
        if err != nil {
            return "", err
        }
    }
    resolved, err := filepath.EvalSymlinks(homeDir)
    if err != nil {
        resolved = homeDir
    }
    return filepath.Join(resolved, ".{tool-config-dir}"), nil
}
```

See `adapters/cursor/paths.go` for the complete reference.

### Step 4: Create Templates

Create the following under `adapters/{tool}/templates/`:

| File | Purpose |
|------|---------|
| `skill.md.tmpl` | The SKILL.md template with agent roster |
| `rules.md.tmpl` or similar | System prompt injection content (named per tool) |
| `commands/sequoia-init.md` | Init command |
| `commands/sequoia-audit.md` | Audit command |
| `commands/sequoia-review.md` | Review command |
| `commands/sequoia-fix.md` | Fix command |
| `commands/sequoia-diff.md` | Diff command |

Templates use `text/template` with a `templateData` struct containing at least
`Version string`. Use `{{.Version}}` inside templates.

Command files use YAML frontmatter matching the tool's conventions. Look at
existing command files in `adapters/cursor/templates/commands/` for reference.

### Step 5: Implement `embed.go`

```go
package {tool}

import "embed"

//go:embed templates
var templateFS embed.FS
```

This embeds all template files into the binary so they ship with the CLI.

### Step 6: Implement `install.go` + `installer.go`

**`install.go`** contains:
- `Version` constant (e.g. `"0.1.0"`)
- `commandFiles` slice listing command filenames
- `templateData` struct for template rendering
- `renderTemplate()` using `text/template`
- `runInstaller()` wrapping `common.Installer.Prepare → Apply → Verify`
- `stageFile()` helper for writing staged content to temp directories

**`installer.go`** contains tool-specific logic:
- If using `StrategyMarkdownSections` or `StrategyConfigMerge`: marker-based
  section injection into existing files (`InjectSection`/`RemoveSection`)
- If using `StrategyFileReplace`: full file replace with backup
  (`GenerateRulesMD`/`RemoveRulesMD`)
- If using `StrategyTOMLMerge`: TOML table merge logic

The `Install()` method on the adapter:
1. Creates a staging temp directory
2. Renders templates to staging
3. Creates target directories
4. Runs `common.Installer` for skills and commands
5. Injects or generates the system prompt file
6. Writes the `.sequoia-version` marker

The `Uninstall()` method reverses this: removes files, restores backups.

Always chain rollbacks: if step N fails, roll back steps 1 through N-1 before
returning the error.

### Step 7: Register in `cmd/sequoia/main.go`

Add a blank import at the top of `cmd/sequoia/main.go`:

```go
import (
    // ... existing imports ...
    _ "sequoia-ai/adapters/{tool}"
)
```

This triggers the `init()` function and registers the adapter. That's it —
the CLI, TUI, and all infrastructure pick it up automatically.

### Step 8: Write Tests

See the [Testing Checklist](#testing-checklist) below. Every adapter must
have tests for paths, detection, install, uninstall, idempotency, and status.

---

## Testing Checklist

When adding a new adapter, every box must be checked before opening a PR.

### Path Tests (`paths_test.go`)

- [ ] `SkillsPath()` returns the correct directory
- [ ] `CommandsPath()` returns the correct directory
- [ ] `SystemPromptPath()` returns the correct file
- [ ] Version file path is inside the skills/commands root
- [ ] Paths use `filepath.Join()` (verify via suffix check with `filepath.ToSlash()`)
- [ ] Home directory resolution works when set explicitly (test constructor)
- [ ] Symlink resolution works (resolve to real path, skip if symlinks unavailable)
- [ ] Missing home directory produces non-empty fallback

### Adapter Interface Tests (`adapter_test.go`)

- [ ] `ID()` returns the expected identifier
- [ ] `Name()` returns the expected display name
- [ ] `PromptStrategy()` returns the correct strategy
- [ ] `Detect()` returns true when the tool's config directory exists
- [ ] `Detect()` returns false when the tool is not installed
- [ ] `IsInstalled()` returns true when Sequoia marker file exists
- [ ] `IsInstalled()` returns false when Sequoia is not installed
- [ ] `Status()` returns installed=false when not installed
- [ ] `Status()` reads version from `.sequoia-version` when present
- [ ] `Status()` returns empty version for legacy installs (no version file)
- [ ] `Status().Path` equals `SkillsPath()`

### Install Tests (`install_test.go`)

- [ ] `Install()` creates all expected files (skill, commands, system prompt)
- [ ] `Install()` writes the version marker file
- [ ] System prompt file contains the version string
- [ ] System prompt file contains Sequoia markers (if applicable)
- [ ] `IsInstalled()` returns true after `Install()`
- [ ] `Install()` is idempotent (running twice produces identical state)
- [ ] `Install()` preserves existing user content (backup created)

### Uninstall Tests (`install_test.go`)

- [ ] `Uninstall()` removes all Sequoia files
- [ ] `Uninstall()` removes the version marker file
- [ ] `IsInstalled()` returns false after `Uninstall()`
- [ ] `Uninstall()` restores backed-up user content
- [ ] `Uninstall()` preserves non-Sequoia content
- [ ] `Uninstall()` is safe when Sequoia was never installed (no-op)

### Template Tests (`templates_test.go`)

- [ ] All templates render without error
- [ ] Rendered templates contain the version string
- [ ] Rendered templates are non-empty
- [ ] Golden file tests (optional but recommended) — write rendered output
  to `testdata/golden/` and assert it matches

### Test Infrastructure Requirements

- [ ] All tests use `t.TempDir()` — never mutate real `~/.claude/`, `~/.cursor/`, etc.
- [ ] Tests are table-driven using `t.Run()` and `testify/assert` + `testify/require`
- [ ] All tests run with `t.Parallel()` (except integration tests)
- [ ] Tests use `NewAdapter(t.TempDir())` to isolate from the real filesystem
- [ ] Race detector enabled: `go test -race ./adapters/{tool}/...`

### Run All Tests

```bash
go test ./adapters/{tool}/...
go test -race ./adapters/{tool}/...
```

---

## Prompt Strategies

Sequoia supports four prompt injection strategies. Choose the right one for
your tool.

### StrategyMarkdownSections

**Used by:** Claude Code (`~/.claude/CLAUDE.md`)

Injects a delimited section into an existing Markdown file using HTML comment
markers:

```markdown
<!-- sequoia:start -->
... Sequoia skill embedding ...
<!-- sequoia:end -->
```

The file is never fully replaced — only the section between markers is managed.
All other content is preserved.

**When to use:** The tool uses a Markdown config file (`.md`) and supports
adding sections without affecting other content.

### StrategyFileReplace

**Used by:** Cursor IDE (`~/.cursor/rules/sequoia-ai.md`), OpenCode (`AGENTS.md`)

Replaces the entire target file with Sequoia content. Before overwriting, a
backup (`.sequoia-backup`) is created if the file contains non-Sequoia content.

On uninstall, the backup is restored. If no backup exists and the file is
Sequoia-managed, it is deleted.

**When to use:** The tool uses a dedicated file for rules/instructions, and
the file is not shared with other extensions or content.

### StrategyConfigMerge

**Used by:** Gemini CLI (`GEMINI.md`)

Like `StrategyMarkdownSections` but adapted for tools where the config file
format is not Markdown but still supports marker-delimited sections.

**When to use:** The tool has a config file that can accept marker-delimited
content but doesn't use Markdown as its native format.

### StrategyTOMLMerge

**Used by:** OpenAI Codex (`~/.codex/config.toml`)

Merges a `[sequoia]` TOML table into an existing TOML config file. All
pre-existing keys and sections are preserved. Only the `[sequoia]` table
is managed.

**When to use:** The tool uses TOML for configuration and supports
arbitrary top-level tables.

---

## Pull Request Process

### Before Opening a PR

1. **Run all tests** — `go test ./...` must pass with zero failures
2. **Run race detector** — `go test -race ./...` must pass
3. **Run linter** — `golangci-lint run` must produce zero issues
4. **Go vet** — `go vet ./...` must produce zero warnings
5. **Check coverage** — new code should maintain or improve coverage
6. **Complete the testing checklist** — every box checked for new adapters

### Commit Format

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(adapters): add {tool} adapter
fix(installer): handle missing backup directory
test(paths): add symlink resolution test
docs(contributing): update testing checklist
refactor(common): extract shared file copy logic
chore: bump go.mod to 1.24
```

**Do NOT** include AI attribution (`Co-Authored-By`, `Generated by`, etc.)
in commit messages.

### Review Expectations

- **New adapters**: Tests are not optional. Every adapter must have the full
  test suite described in the testing checklist.
- **Template changes**: Golden file tests must be updated when templates change.
- **Common code**: Changes to `adapters/common/` must not break any existing
  adapter tests.
- **Cross-platform**: All paths must use `filepath.Join()`. Windows is a
  first-class target.
- **Documentation**: `CONTRIBUTING.md` and adapter READMEs must be updated
  when the adapter development process changes.

### PR Size

Keep PRs focused. If a change exceeds 400 lines (excluding test fixtures and
generated files), split it into chained PRs. Each PR should tell one coherent
story.

---

## Code Conventions

### Godoc

All exported symbols MUST have godoc comments:

```go
// Adapter implements adapters.ToolAdapter for Cursor IDE.
// homeDir overrides os.UserHomeDir() for testing. Leave empty for production.
type Adapter struct { ... }

// cursorBase returns the ~/.cursor/rules/ directory.
// If homeDir is non-empty it is used directly; otherwise os.UserHomeDir()
// is called.
func cursorBase(homeDir string) (string, error) { ... }
```

### Path Construction

**Always use `filepath.Join()`.** Never concatenate paths with `/` or `\\`:

```go
// ✅ Correct
filepath.Join(base, "skills", "SKILL.md")
filepath.Join(resolved, ".cursor", "rules")

// ❌ Wrong
base + "/skills/SKILL.md"
base + "\\skills\\SKILL.md"
fmt.Sprintf("%s/skills/SKILL.md", base)
```

Use `filepath.EvalSymlinks()` to resolve symlinks before path operations.
Always provide a fallback when resolution fails.

### Test Patterns

- Use `t.TempDir()` for all test directories — never touch real config paths
- Use `t.Parallel()` for independent tests
- Use `testify/require` for preconditions, `testify/assert` for checks
- Table-driven tests with `t.Run()` for multiple scenarios
- Temp directories with `NewAdapter(t.TempDir())` for isolation

### Error Handling

- Wrap errors with `fmt.Errorf("context: %w", err)` — use `%w` to preserve
  the error chain
- Return early on errors, don't accumulate
- Chain rollbacks when multi-step operations fail

### Package Organization

- Each adapter is its own package: `adapters/{tool}/`
- Test files use `{tool}_test` package for black-box testing
- Shared infrastructure lives in `adapters/common/`
- Core types (interface, registry, factory) live in `adapters/`

### No AI Attribution

Do not include `Co-Authored-By`, `Generated by AI`, or similar attributions
in commit messages, code comments, or documentation. This is a project
convention enforced in code review.

---

## Getting Help

- Study the Cursor adapter (`adapters/cursor/`) — it's the cleanest and most
  up-to-date reference implementation
- Check existing test files for patterns
- Look at `adapters/interface.go` for the full contract
- Read `adapters/common/installer.go` for the shared install framework
- The `adapters/_template/` directory has copy-paste boilerplate to get
  started quickly
