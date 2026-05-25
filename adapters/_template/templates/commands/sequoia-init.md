---
description: "Initializes Sequoia in the project. Delegates to sequoia-context for detection, persists the Project Map, and reports applicable agents. Mandatory first step before any audit."
allowed-tools: Read, Glob, Grep, Bash
---

# /sequoia init

Initializes Sequoia in the current project. Delegates ALL detection work to the context agent and persists the result.

## What it does

1. **Delegates** to `sequoia-context` agent — the single source of truth for stack, paradigm, size, maturity, and agent applicability.
2. **Persists** the Project Map in Engram with topic_key `sequoia/{project-name}/project-map`.
3. **Saves** the Project Map to `.sequoia/project-map.md` in the audited project root (filesystem persistence).
4. **Reports** back: detected stack, applicable agents, estimated audit duration.

## Precondition

None. This is always the first command.

## Post-condition

The Project Map is persisted in Engram and saved to `.sequoia/project-map.md` in the project root. All subsequent commands (`audit`, `review`, `diff`, `fix`) consume it automatically.

## Delegate prompt

```
You are sequoia-context. Build the complete Project Map for this project.

Output the full Project Map YAML as defined in your schema.
After completion, I will persist it in Engram.
```

## After delegation

1. Receive the Project Map from context agent.
2. Save to Engram:
   - **title**: "Sequoia Project Map — {project-name}"
   - **topic_key**: `sequoia/{project-name}/project-map`
   - **type**: `architecture`
3. Write to filesystem — save Project Map to `.sequoia/project-map.md` in the
   audited project root:
   - If `.sequoia/` directory does not exist, create it.
   - Render the Project Map as a structured Markdown document (one section per
     YAML key: identity, dimensions, infrastructure, testing, agents).
   - If write fails (permissions), warn user and continue — Engram save is the
     canonical copy.
4. If a previous Project Map exists, compare and report significant changes.
5. Report to user:
   - Detected stack
   - Maturity level
   - Applicable agents (with reasons for excluded ones)
   - Estimated audit duration

## Template files

- `docs/agents/sequoia-context.md` — the context agent definition

<!-- @see adapters/common/templates/commands/sequoia-init.md — keep in sync -->
