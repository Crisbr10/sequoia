# `/sequoia-dev` — Automated SDD Task Development Command

**Date**: 2026-05-21
**Status**: Design complete — ready for implementation
**Author**: Orchestrator + Cris (user validation)

---

## Goal

Eliminate repetitive manual configuration when implementing audit tasks. A single command `/sequoia-dev <TASK-ID>` that automatically launches SDD with the user's preferred strategies, including automatic fast-forward detection for simple tasks.

---

## Architecture Summary

Zero changes to the Go CLI binary. Follows the same pattern as existing Sequoia slash commands (`/sequoia fix`, `/sequoia audit`, etc.): a Markdown template installed to `~/.config/opencode/commands/` that the AI orchestrator reads and executes.

```
~/.config/opencode/
├── commands/
│   ├── sequoia-fix.md          ← existing
│   ├── sequoia-audit.md        ← existing
│   ├── sequoia-dev.md          ← NEW — slash command template
│   └── ...
├── .sequoia-dev.yaml            ← NEW — user configuration (optional, generated from default)
└── ...
```

### Flow

```
User: /sequoia-dev P1-003 [--ff | --full]

1. Orchestrator reads commands/sequoia-dev.md
2. Template instructs orchestrator to:
   a. Read ~/.config/opencode/.sequoia-dev.yaml (merge with defaults)
   b. Locate task P1-003 in docs/sequoia/tasks/{area}.md
   c. Parse task metadata (risk, effort, files, dependencies)
   d. Compute complexity score
   e. Decision:
      - score ≤ 2 AND auto_ff=true (or --ff flag) → sdd-ff
      - score ≥ 3 OR --full flag → sdd-new with full SDD cycle
      - User override --ff / --full always wins
   f. Launch SDD with configured strategies:
      - TDD: strict (from config)
      - Delivery: ask-on-risk (from config)
      - Chain: stacked-to-main (from config)
3. SDD subagents execute: proposal→specs→design→tasks→apply→verify→archive
```

### Complexity Heuristic

```
score = sum of:
  Implementation risk:
    Low    → 0
    Medium → 1
    High   → 3

  Effort:
    small (<2h)   → 0
    medium (2-8h) → 1
    large (>8h)   → 3

  Files involved:
    1-2  → 0
    3-4  → 1
    5+   → 2

  Dependencies:
    None      → 0
    Has deps  → 1

Decision:
  score ≤ 2 → sdd-ff (fast-forward)
  score ≥ 3 → full SDD cycle
```

`--ff` and `--full` flags always override the heuristic.

---

## Task Breakdown

### TASK-1: Create the `/sequoia-dev` command template

**Priority**: 🔴 Blocking
**Source**: New feature
**Implementation risk**: Low
**Effort**: small (<2h)

#### Context

The template is the brain of the command. It's a Markdown file with YAML frontmatter (same format as `sequoia-fix.md`) that tells the orchestrator what to do. It must:
1. Parse the task ID from user input
2. Read config from `.sequoia-dev.yaml`
3. Locate the task in the filesystem
4. Compute complexity and decide `sdd-ff` vs full cycle
5. Launch the appropriate SDD subagents with the right parameters

#### Why

This is the only piece that makes the command function. Without it, there's no behavior. Everything else (config, install) supports this file.

#### Files to create

- `adapters/common/templates/commands/sequoia-dev.md` — NEW

#### Detailed changes

The template must follow the existing convention (see `sequoia-fix.md` for reference):

**YAML frontmatter**:
```yaml
---
description: "Develop a task using SDD with configured strategies. Auto-detects simple tasks for fast-forward mode."
argument-hint: "<task-id> [--ff | --full]"
allowed-tools: Read, Glob, Grep, Bash, mem_search, mem_get_observation, mem_save
---
```

**Template body** must include these sections:

1. **Precondition check**: Verify SDD init has been run (search Engram for `sdd-init/{project}`). If not found, instruct to run `/sdd-init` first.

2. **Config loading**: Read `~/.config/opencode/.sequoia-dev.yaml`. Merge with defaults:
   ```yaml
   sdd:
     tdd: strict
     delivery: ask-on-risk
     chain: stacked-to-main
     auto_ff: true
     complexity:
       ff_max_score: 2
   paths:
     tasks_dir: docs/sequoia/tasks/
   ```

3. **Task location**: Search `{tasks_dir}` for the TASK-ID pattern. Parse the task metadata:
   - `Implementation risk`: Low | Medium | High
   - `Effort`: small | medium | large
   - `Files involved` section → count the `- ` bullet items
   - `Dependencies` section → check if any non-empty value exists

4. **Complexity computation**: Apply the heuristic table above.

5. **SDD launch**:
   - **If sdd-ff**: invoke `sdd-new` with the `--ff` semantics (skip proposal and specs, go directly to tasks+apply)
   - **If full cycle**: invoke `sdd-new` with the task as the change description

   Pass these parameters to ALL SDD subagent launches:
   - TDD mode from config (`strict` or `standard`)
   - Delivery strategy from config
   - Chain strategy from config

6. **Override flags**: `--ff` forces fast-forward regardless of score. `--full` forces full cycle.

7. **Error handling**:
   - Task not found → search all area files, suggest alternatives, stop
   - Config missing → use defaults, continue silently
   - Config syntax error → warn user, use defaults, continue
   - Task missing metadata fields → use safe defaults (risk=Medium, effort=medium, files=3, deps=none)

8. **Feedback**: Always tell the user:
   - Which mode was selected (ff vs full) and why
   - Which strategies are active
   - What's about to happen

---

### TASK-2: Create the default configuration file

**Priority**: 🔴 Blocking
**Source**: New feature
**Implementation risk**: Low
**Effort**: small (<2h)

#### Context

The `.sequoia-dev.yaml.default` file ships with Sequoia and gets copied to `~/.config/opencode/.sequoia-dev.yaml` on first install — only if the user doesn't already have one. This provides sensible defaults while allowing users to customize without fear of being overwritten.

#### Why

Without this file, the command still works (templates have hardcoded defaults). But this enables:
- User customization without touching the template
- Team-wide standardization (different configs for different projects)
- Future extensibility — adding new strategies or parameters without changing the template

#### Files to create

- `adapters/common/templates/.sequoia-dev.yaml.default` — NEW

Alternatively, place it in a `config/` subdirectory:
- `adapters/common/templates/config/.sequoia-dev.yaml.default` — NEW

**Recommendation**: Place at `templates/.sequoia-dev.yaml.default` (top-level of templates). It's a single file — a subdirectory adds nesting without value.

#### Detailed changes

File content:

```yaml
# Sequoia Dev — SDD Strategy Configuration
# https://github.com/Crisbr10/sequoia
#
# This file is read by /sequoia-dev to determine how tasks are developed.
# Edit it to match your preferred workflow.
#
# If this file is deleted, /sequoia-dev falls back to the defaults shown below.

sdd:
  # TDD mode: strict (red-green-refactor) or standard (tests optional)
  tdd: strict

  # Delivery strategy for PR review workflow
  #   ask-on-risk    — Ask before splitting large PRs (>400 lines)
  #   auto-chain     — Auto-split into chained PRs
  #   single-pr      — One PR, require size exception if >400 lines
  #   exception-ok   — One PR, no size limit, maintainer accepts risk
  delivery: ask-on-risk

  # Chain strategy (only used when PRs are split)
  #   stacked-to-main     — Each PR merges to main in order
  #   feature-branch-chain — Tracker branch accumulates, one merge to main
  chain: stacked-to-main

  # Automatic fast-forward for simple tasks
  # When true, tasks scoring ≤ complexity.ff_max_score use sdd-ff
  auto_ff: true

  complexity:
    # Tasks with score ≤ this value use fast-forward (sdd-ff)
    # Tasks with score > this value use full SDD cycle
    ff_max_score: 2
    # Score components:
    #   Implementation risk: Low=0 Medium=1 High=3
    #   Effort: small=0 medium=1 large=3
    #   Files involved: 1-2=0 3-4=1 5+=2
    #   Dependencies: none=0 has_deps=1

# Paths for task file discovery
paths:
  # Directory where Sequoia task files are stored (relative to project root)
  tasks_dir: docs/sequoia/tasks/
```

---

### TASK-3: Register the command in the installer system

**Priority**: 🔴 Blocking (depends on TASK-1)
**Source**: New feature
**Implementation risk**: Low
**Effort**: small (<2h)

#### Context

The Sequoia installer reads `CommandFiles` from `adapters/common/commands.go` to know which `.md` files to copy to `~/.config/opencode/commands/`. Adding `sequoia-dev.md` to this list makes the installer aware of it.

For the config file `.sequoia-dev.yaml.default`, a new mechanism is needed because it goes to `~/.config/opencode/` (base dir), not the `commands/` subdirectory. And it must only copy if the target doesn't already exist.

#### Why

Without this, `sequoia install` won't deploy the new command template or config file. Users would have to manually copy files, defeating the purpose.

#### Files to modify

1. **`adapters/common/commands.go`** — Add `"sequoia-dev.md"` to `CommandFiles` slice
2. **`adapters/common/embed.go`** — Add embed directive for the config default file
3. **`adapters/common/base_adapter.go`** — Add config file installation logic (conditional copy)
4. **`adapters/common/files.go`** — Potentially add a `StageFileIfNotExists` helper

#### Detailed changes

**`adapters/common/commands.go`**:
```go
var CommandFiles = []string{
    "sequoia-init.md",
    "sequoia-audit.md",
    "sequoia-review.md",
    "sequoia-fix.md",
    "sequoia-diff.md",
    "sequoia-dev.md",   // ← ADD THIS
}
```

Add a new variable for config files:
```go
// ConfigFiles lists static configuration files that are copied to the
// adapter's base directory (not the commands/ directory). Files are only
// copied if the target does not already exist, preserving user edits.
var ConfigFiles = []struct {
    Source string // filename relative to templates/ in the embed.FS
    Target string // filename relative to the adapter's base directory
}{
    {Source: ".sequoia-dev.yaml.default", Target: ".sequoia-dev.yaml"},
}
```

**`adapters/common/embed.go`**:
```go
package common

import "embed"

//go:embed templates/commands
var CommandFS embed.FS

//go:embed templates/.sequoia-dev.yaml.default
var ConfigDefaultFS embed.FS
```

**`adapters/common/files.go`** — Add new helper:
```go
// StageFileIfNotExist writes content to filepath.Join(dir, name) only if the
// target file does not already exist. Creates dir and any missing parent
// directories (mode 0o755). Returns nil if the file already exists (not an error).
func StageFileIfNotExist(dir, name string, content []byte) error {
    target := filepath.Join(dir, name)
    if _, err := os.Stat(target); err == nil {
        return nil // file exists, skip
    }
    return StageFile(dir, name, content)
}
```

**`adapters/common/base_adapter.go`** — In the `Install` method, after the section that installs command files (around line 320), add:

```go
// Stage config default file (conditional — only if target doesn't exist).
// Config files go to the base directory, not commands/skills.
configContent, err := ConfigDefaultFS.ReadFile("templates/.sequoia-dev.yaml.default")
if err != nil {
    return fmt.Errorf("install: read config default: %w", err)
}
if err := StageFileIfNotExist(base, ".sequoia-dev.yaml", configContent); err != nil {
    return fmt.Errorf("install: stage config: %w", err)
}
```

Also add config file removal in the `Uninstall` method (line 445 area):
```go
// Remove config files
if err := os.Remove(filepath.Join(base, ".sequoia-dev.yaml")); err != nil && !os.IsNotExist(err) {
    errs = append(errs, fmt.Errorf("remove config .sequoia-dev.yaml: %w", err))
}
```

**Important design note**: The config file is installed to `base` (e.g., `~/.config/opencode/`), NOT to `commands/`. This is intentional — it's user configuration, not a command definition. The copy is conditional: if `.sequoia-dev.yaml` already exists, the default file is skipped entirely. This preserves user customizations across reinstalls.

---

### TASK-4: Add cross-reference in `/sequoia fix` output

**Priority**: 🟠 High leverage
**Source**: New feature
**Implementation risk**: Low
**Effort**: small (<2h)

#### Context

When a user runs `/sequoia fix` and gets a list of tasks, there should be a hint at the bottom telling them they can use `/sequoia-dev <TASK-ID>` to implement each task. This improves discoverability.

#### Why

Without this, users might not know `/sequoia-dev` exists. The integration point is natural — right after task generation, when the user is thinking "ok, now how do I implement this?"

#### Files to modify

- `adapters/common/templates/commands/sequoia-fix.md` — Add a "Next Steps" section at the end

#### Detailed changes

At the very end of `sequoia-fix.md` (after line 190, before EOF), add:

```markdown
## Next Steps: Implementing Tasks

To develop any generated task with SDD workflow, TDD strict mode, and your preferred review strategy:

```bash
/sequoia-dev <TASK-ID>
```

Example:
```bash
/sequoia-dev P1-003
```

The command automatically:
- Reads the task from `docs/sequoia/tasks/`
- Launches SDD with your configured strategies (see `~/.config/opencode/.sequoia-dev.yaml`)
- Uses fast-forward (`sdd-ff`) for simple tasks, full SDD cycle for complex ones
- Optionally force mode: `/sequoia-dev P1-003 --ff` or `/sequoia-dev P1-003 --full`

See `/sequoia-dev` for full documentation.
```

---

### TASK-5: Tests

**Priority**: 🟡 Backlog
**Source**: New feature
**Implementation risk**: Low
**Effort**: medium (2-8h)

#### Context

The project has a test suite in `cmd/sequoia/main_test.go` for CLI integration and `adapters/common/installer_test.go` for installer logic. The new config file installation mechanism and the `StageFileIfNotExist` helper need tests.

The command template itself (`sequoia-dev.md`) can't be unit tested in Go — it's a prompt template evaluated by the AI orchestrator. However, the Go-level changes (config file copy logic, CommandFiles registration) are testable.

#### Why

Prevent regressions: if someone later changes `CommandFiles` or the install flow, tests catch it. The conditional copy logic (`StageFileIfNotExist`) is the most critical to test because a bug there could overwrite user config.

#### Files to modify or create

1. **`adapters/common/files_test.go`** — NEW (or add to existing test file)
2. **`adapters/common/installer_test.go`** — Add config file tests
3. **`adapters/common/commands_test.go`** — NEW — verify CommandFiles includes sequoia-dev.md

#### Detailed changes

**`adapters/common/files_test.go`** — tests for `StageFileIfNotExist`:
- **TestStageFileIfNotExist_createsWhenMissing**: file doesn't exist → creates it
- **TestStageFileIfNotExist_preservesWhenExists**: file already exists → skips, original content preserved
- **TestStageFileIfNotExist_createsParentDirs**: parent directories created if missing

**`adapters/common/installer_test.go`** — add config install tests:
- **TestInstall_writesConfigDefault**: fresh install copies `.sequoia-dev.yaml` to base dir
- **TestInstall_preservesExistingConfig**: existing `.sequoia-dev.yaml` is NOT overwritten on reinstall
- **TestUninstall_removesConfig**: uninstall removes `.sequoia-dev.yaml`

**`adapters/common/commands_test.go`** — simple assertion:
```go
func TestCommandFiles_IncludesSequioaDev(t *testing.T) {
    found := false
    for _, cmd := range CommandFiles {
        if cmd == "sequoia-dev.md" {
            found = true
            break
        }
    }
    assert.True(t, found, "CommandFiles must include sequoia-dev.md")
}
```

Additionally, verify that existing tests still pass after adding `sequoia-dev.md` to `CommandFiles`. The installer test may have hardcoded expectations about file counts or file names — grep for `sequoia-fix.md` in test files and add corresponding `sequoia-dev.md` assertions.

---

### TASK-6: Adapter-specific integration (OpenCode)

**Priority**: 🟡 Backlog
**Source**: New feature
**Implementation risk**: Low
**Effort**: small (<2h)

#### Context

The `adapters/opencode/` package has its own `templates/` directory with `skill.md.tmpl` and `agents-md-section.md.tmpl`. Since the new command template lives in `adapters/common/templates/commands/` (shared across all adapters), no OpenCode-specific changes are needed for the command file itself.

However, the OpenCode adapter's `embed.go` only embeds `templates` (its own directory). The shared `CommandFS` in `adapters/common/embed.go` handles the command templates. These are separate embed FS instances.

#### Why

Verify nothing breaks. The OpenCode adapter has its own test suite (`adapter_test.go`, `install_test.go`, `templates_test.go`) that may have hardcoded expectations.

#### Files to check

1. **`adapters/opencode/adapter_test.go`** — verify no hardcoded command count assertions
2. **`adapters/opencode/install_test.go`** — verify install tests still pass
3. **`adapters/opencode/templates_test.go`** — verify template rendering unchanged

Run the full test suite:
```bash
go test ./adapters/... -v
```

If any test in `adapters/opencode/testdata/` has golden files or snapshots that reference command files, update them to include `sequoia-dev.md`.

---

## Summary — Implementation Order

| Order | Task | Dependency | Risk |
|-------|------|------------|------|
| 1 | TASK-1: Create `sequoia-dev.md` template | None | Low |
| 2 | TASK-2: Create `.sequoia-dev.yaml.default` | None | Low |
| 3 | TASK-3: Register in installer (command + config) | TASK-1, TASK-2 | Low |
| 4 | TASK-4: Cross-reference in `/sequoia fix` | TASK-1 | Low |
| 5 | TASK-5: Tests | TASK-3 | Low |
| 6 | TASK-6: Verify OpenCode adapter compatibility | TASK-3 | Low |

Tasks 1 and 2 can be done in parallel. Task 3 requires both. Tasks 4-6 can be done after 3.

---

## Files Summary

| File | Action | Task |
|------|--------|------|
| `adapters/common/templates/commands/sequoia-dev.md` | CREATE | TASK-1 |
| `adapters/common/templates/.sequoia-dev.yaml.default` | CREATE | TASK-2 |
| `adapters/common/commands.go` | MODIFY | TASK-3 |
| `adapters/common/embed.go` | MODIFY | TASK-3 |
| `adapters/common/base_adapter.go` | MODIFY | TASK-3 |
| `adapters/common/files.go` | MODIFY | TASK-3 |
| `adapters/common/templates/commands/sequoia-fix.md` | MODIFY | TASK-4 |
| `adapters/common/files_test.go` | CREATE | TASK-5 |
| `adapters/common/commands_test.go` | CREATE | TASK-5 |
| `adapters/common/installer_test.go` | MODIFY | TASK-5 |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Config default overwrites user config | Very Low | High | `StageFileIfNotExist` is a no-op when file exists. Verified by test in TASK-5. |
| `CommandFiles` change breaks existing tests | Medium | Low | Run full `go test ./...` after TASK-3. Fix any hardcoded expectations. |
| Template doesn't parse task metadata correctly | Medium | Medium | Template includes fallback defaults for all fields. If `Implementation risk` is missing, defaults to `Medium`. |
| `sdd-ff` vs full cycle heuristic is wrong for some tasks | Medium | Low | `--ff` and `--full` flags let user override. Configurable threshold (`ff_max_score`). |
| OpenCode adapter test expects exact command list | Low | Low | Tests use glob/prefix matching. Verify in TASK-6. |
