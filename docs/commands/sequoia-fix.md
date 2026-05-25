---
description: "Generates an actionable task plan from audit findings. Output optimized so another implementing agent can execute without ambiguity. Includes minimum context, files, acceptance criteria."
argument-hint: "<phase|all> [--task=<id>]"
allowed-tools: Read, Glob, Grep
---

# /sequoia fix

> **Nota**: Desde Sequoia v0.2.0, las tareas se generan automáticamente durante `/sequoia audit`. Este comando permanece como fallback para regenerar tareas desde la última auditoría en Engram sin re-ejecutar agentes de fase.

Generates implementable tasks from audit findings. Each task is self-contained: an implementing agent can execute it without re-reading the full audit.

## Precondition

A prior audit must exist in Engram (via `/sequoia audit` or `/sequoia review`). **Validate blocking**: search Engram for the most recent audit. If none found, error: "No prior audit found. Run `/sequoia audit` first." Do not proceed without an audit.

Each finding used to generate a task must have: (a) a valid file location, (b) a description ≥20 characters. Skip insufficient findings with a warning.

## What it does

1. Retrieves findings from the most recent audit
2. Filters by phase (if specified) or takes all
3. Converts each finding into an implementable task
4. Orders by dependencies and priority
5. Generates the task document using the same format as `/sequoia audit` (see task template in SKILL.md)

## Usage

```bash
# Tasks for a specific phase
/sequoia fix security

# Tasks for all phases
/sequoia fix all

# A specific task by ID
/sequoia fix security --task=P1-003
```

## 🔴 CRITICAL: Task ID and Status are NON-NEGOTIABLE

Every generated task, without exception, MUST include:

1. **`[TASK-ID]`** — a unique identifier in the format `{PHASE}-{NNN}` (e.g., `P1-003`, `P3-012`, `M1-001`). This ID is the task's permanent reference for `/sequoia-dev`, `/sequoia fix --task=`, dependency tracking, and status updates.
2. **`**Status**`** — exactly one of `⏳ Pending` or `✅ Resolved`. New tasks ALWAYS default to `⏳ Pending`.

**Tasks missing either field are INVALID.** They cannot be tracked, referenced, or automated. If you encounter a task without `[TASK-ID]` or `**Status**`, regenerate it immediately.

## Per-task format

Each generated task follows this mandatory structure:

```markdown
### [TASK-ID] · [Actionable title]

**Status**: ⏳ Pending | ✅ Resolved
**Priority**: 🔴 Blocking | 🟠 High leverage | 🟡 Backlog
**Source phase**: [P1-P6 | M1-M2]
**Source finding(s)**: [ID(s) of the finding that generates this task]

**Minimum context**:
An explanation of WHAT is wrong and WHY it matters, in 3-5 lines.
Enough to understand the problem without reading the full audit.

**Files involved**:
- `path/to/file.ext` — what role it plays in this task
- `path/to/other.ext` — what needs to be modified

**What to do**:
Concrete step by step. Not "improve X." But:
1. Add function Y in file Z
2. Modify the call in file W to use the new function
3. Update the test in file T

**Expected impact**:
What changes when implementing this. Observable metric if possible.

**Dependencies**:
- Requires [TASK-ID] to be completed first
- Blocked by: [external factor, if applicable]

**Implementation risk**: Low | Medium | High
Reason for the risk.

**Acceptance criteria**:
- [ ] Verifiable condition 1
- [ ] Verifiable condition 2
- [ ] Test that must pass (if applicable)

**Verification**:
How to confirm the task is really done.
Concrete command or manual step.
```

## Principle: self-contained task

A well-generated task meets these rules:

1. **Has a mandatory identifier and status** — every task MUST have a `[TASK-ID]` and a `**Status**` field (`⏳ Pending` or `✅ Resolved`). New tasks default to Pending.
2. **Does not require reading the full audit** — all context is in the task
3. **Is not ambiguous** — a developer (or agent) can implement without questions
4. **Has verifiable acceptance criteria** — not "improve X," but "test Y passes"
5. **Declares explicit dependencies** — knows which tasks must go first
6. **Declares risk honestly** — not everything is "low risk"

## Generation by phase vs all

### By phase (`/sequoia fix security`)
- Takes only findings from the indicated phase
- Orders by severity within the phase
- Generates dependencies only within the phase

### All phases (`/sequoia fix all`)
- Takes findings from all phases
- Uses M1 correlator results to group root causes
- Orders globally: blocking first, then high leverage
- Generates cross-phase dependencies when the root cause is shared

## Implementation order optimization

Tasks are ordered using a deterministic algorithm:

1. **Priority scoring**: `score = severity × findings_resolved × (1/estimated_hours)`
   - severity: critical=4, high=3, medium=2, low=1
2. **Topological sort** by dependencies (task B requiring task A → A before B)
3. **Cycle detection**: if A requires B and B requires A → flag as "dependency cycle" and group together
4. **Tiebreaker**: lower risk first, then lower effort

**Output grouping**:
- 🔴 Blocking (must be done this sprint)
- 🟠 High Leverage (high impact / low effort ratio)
- 🟡 Backlog (important but not urgent)

## Deduplication rule

If multiple findings point to the same root cause (detected by the correlator), ONE task is generated that resolves all related findings. The finding IDs are listed as source.

## Output

Generates task files under `.sequoia/tasks/` using the same format and structure as `/sequoia audit`:

```
.sequoia/tasks/
├── index.md           # Global dependency graph, priority tiers, risk estimate
├── security.md        # Security tasks with full evidence
├── architecture.md    # Architecture tasks with full evidence
├── performance.md     # Performance tasks with full evidence
├── quality.md         # Quality tasks with full evidence
└── operations.md      # Operations tasks with full evidence
```

When filtering by phase (`/sequoia fix security`), only the corresponding area file is generated. When running `all`, all area files + `index.md` are generated.

Each task follows the standard template defined in Sequoia's SKILL.md. See `/sequoia audit` for the primary workflow.

## Example

```bash
# Generate security tasks
/sequoia fix security

# Generate all tasks
/sequoia fix all

# Implement a specific task
/sequoia fix security --task=P1-003
```

## --task Validation

When `--task=<id>` is specified:
1. Validate that the task ID exists in the target phase's findings.
2. If not found, search ALL phases and suggest: "Task P1-003 not found in security. Did you mean --task=P3-003 (found in architecture)?"
3. Only generate the single specified task.

## Fix vs Audit — When to Use Which

| Aspect | `/sequoia fix` | `/sequoia audit` |
|--------|---------------|-----------------|
| Source | Last audit in Engram | Fresh code analysis |
| Speed | Fast (reads from memory) | Slow (re-runs all agents) |
| Freshness | Stale if code changed | Current state |
| Use case | Regenerate tasks from existing findings | Full re-analysis after major changes |
| Output | Task files only | Summary + task files + health score |
| Precondition | Prior audit exists | Init completed |

**Rule of thumb**: Use `fix` when you want to regenerate tasks without re-auditing. Use `audit` when code has changed significantly.

## Manual Edits Preservation

If a task file contains `<!-- MANUAL_EDIT -->` markers, those sections must NOT be overwritten:

```markdown
### P1-003 · Add CSP headers
<!-- MANUAL_EDIT -->
**Files involved**:
- nginx.conf (override the policy set by the CDN)
<!-- /MANUAL_EDIT -->
```

When regenerating tasks, preserve any content between `<!-- MANUAL_EDIT -->` and `<!-- /MANUAL_EDIT -->` markers. Only regenerate the non-marked sections.
