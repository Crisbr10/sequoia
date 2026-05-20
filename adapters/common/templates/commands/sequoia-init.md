---
description: "Initializes Sequoia in the project. Delegates to sequoia-context for detection, persists the Project Map, and reports applicable agents. Mandatory first step before any audit."
allowed-tools: Read, Glob, Grep, Bash
---

# /sequoia init

Initializes Sequoia in the current project. Delegates ALL detection work to the context agent and persists the result.

## What it does

1. **Delegates** to `sequoia-context` agent — the single source of truth for stack, paradigm, size, maturity, and agent applicability.
2. **Persists** the Project Map in Engram with topic_key `sequoia/{project-name}/project-map`.
3. **Reports** back: detected stack, applicable agents, estimated audit duration.

## Precondition

None. This is always the first command.

## Post-condition

The Project Map is persisted in Engram. All subsequent commands (`audit`, `review`, `diff`, `fix`) consume it automatically.

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
3. If a previous Project Map exists, compare and report significant changes.
4. Report to user:
   - Detected stack
   - Maturity level
   - Applicable agents (with reasons for excluded ones)
   - Estimated audit duration

## Template files

- `docs/agents/sequoia-context.md` — the context agent definition
