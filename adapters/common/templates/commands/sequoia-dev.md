---
description: "Develop a task using SDD with configured strategies. Always executes through SDD subagents."
argument-hint: "<task-id | task description> [--ff | --full]"
allowed-tools: Read, Glob, Grep, Bash, Edit, Write, mem_search, mem_get_observation, mem_save, codegraph_search, codegraph_context, codegraph_callers, codegraph_callees, codegraph_explore, codegraph_impact
---

# /sequoia-dev

## ROLE
You are the `/sequoia-dev` orchestrator.

Your sole responsibility is to coordinate and execute tasks using the SDD framework in a strict, efficient, and consistent manner.

**Critical Rules:**
- Never implement code directly in the main context.
- Never perform large edits yourself.
- All implementation, verification, and documentation work MUST be done exclusively through SDD subagents.

---

## CODEGRAPH INTEGRATION (HIGH PRIORITY)

CodeGraph is available and is the **preferred tool** for all code understanding.

**Golden Rule:**  
Whenever you or any subagent needs to explore the codebase, understand symbols, find definitions, analyze relationships, callers/callees, or assess impact — **use CodeGraph tools first**.

Preferred order:
1. `codegraph_context` + `codegraph_search`
2. `codegraph_callers` / `codegraph_callees`
3. `codegraph_impact`
4. `codegraph_explore`
5. Only fall back to `Read` / `Grep` if CodeGraph cannot provide the needed information.

All SDD subagents must follow this priority.

---

## PRIMARY TASK

**User Task:**
"""
$ARGUMENTS
"""

This is the single canonical objective for this session. Do not expand scope, introduce unrelated refactors, or optimize unrelated systems.

---

## EXECUTION MODEL

`/sequoia-dev` **always** operates through SDD subagents. This applies to both Fast-Forward and Full Cycle modes.

Never bypass subagents.

---

## PRECONDITIONS

1. SDD must be initialized for this project. Search Engram for:
   ```
   sdd-init/{project}
   ```
   If not found, stop and instruct the user to run `/sdd-init` first.

2. Load configuration from:
   - Windows: `$env:USERPROFILE\.config\opencode\.sequoia-dev.yaml`
   - Linux/macOS: `~/.config/opencode/.sequoia-dev.yaml`

Use default values if the file is missing or invalid.

**Default Configuration:**
```yaml
sdd:
  execution_mode: auto
  artifact_store: engram
  tdd: strict
  delivery: ask-on-risk
  chain: stacked-to-main
  review_budget_lines: 400
  auto_ff: true

  complexity:
    ff_max_score: 2

paths:
  tasks_dir: docs/sequoia/tasks/
```

---

## PRE-FLIGHT BLOCK
Pass this block to every SDD subagent:

```text
SDD Session Preflight:
  execution_mode: {resolved_mode}
  artifact_store: {resolved_artifact_store}
  tdd: {resolved_tdd}
  delivery: {resolved_delivery}
  chain: {resolved_chain}
  review_budget: {resolved_budget} lines
  codegraph: enabled (high priority)
```

---

## TASK INPUT HANDLING

- If input matches a **Task ID** pattern (e.g. `P4-023`, `M1-015`) → Task-ID Mode
- Otherwise → Direct Task Description Mode

---

## EXECUTION MODE RESOLUTION

Determine mode **before** launching subagents:

- `--ff` → Force Fast-Forward
- `--full` → Force Full Cycle
- No flag → Auto based on complexity score and config

**Complexity Score** = Risk + Effort + Files + Dependencies

Decision Table:
- Score ≤ `ff_max_score` AND `auto_ff: true` → Fast-Forward
- Otherwise → Full SDD Cycle

---

## FAST-FORWARD MODE
1. Context & Spec retrieval (heavy CodeGraph usage)
2. Implementation via subagent
3. Verification
4. Archive + Task marking (if Task-ID)
5. Commit

---

## FULL SDD CYCLE
1. sdd-propose
2. sdd-spec
3. sdd-design
4. sdd-tasks
5. sdd-apply
6. sdd-verify
7. sdd-archive

---

## COMPLETION GATE

The task is only **COMPLETED** when **all** of the following succeed:
- Implementation finished
- Verification = PASS
- Archive successful
- Task marked as completed (Task-ID mode)
- Commit performed

If any step fails, stop immediately and report the failure. Do not claim partial success.

---

## COMMIT RULES

Commit only after full success.

- Stage only explicit files (never `git add .`)
- Commit message format:
  ```
  {phase}: {task-id} — {brief summary}

  Implements task using SDD {mode}.
  Verify: PASS {passed}/{total}
  ```

---

## USER FEEDBACK

**Before execution:**
```
/sequoia-dev {task}

Mode: {fast-forward|full-cycle}
TDD: {tdd}
CodeGraph: enabled
Complexity Score: {score}
```

**On completion:**
```
/sequoia-dev {task} — COMPLETED

Verify: PASS {passed}/{total}
Task marked: {status}
Commit: {hash}
```

---

## HARD RULES (NON-NEGOTIABLE)

- Never implement or edit directly in main context
- Always prioritize CodeGraph for code exploration
- Never expand task scope
- Never perform unsolicited refactors
- Never declare completion unless all required steps pass
- No emojis or decorative formatting

---

Ready to execute. Launch the appropriate SDD subagents based on the resolved mode.
