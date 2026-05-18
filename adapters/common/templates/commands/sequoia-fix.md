---
description: "Generates an actionable task plan from audit findings. Output optimized so another implementing agent can execute without ambiguity. Includes minimum context, files, acceptance criteria."
argument-hint: "<phase|all> [--task=<id>]"
allowed-tools: Read, Glob, Grep
---

# /sequoia fix

> **Nota**: Desde Sequoia v0.2.0, las tareas se generan automáticamente durante `/sequoia audit`. Este comando permanece como fallback para regenerar tareas desde la última auditoría en Engram sin re-ejecutar agentes de fase.

Generates implementable tasks from audit findings. Each task is self-contained: an implementing agent can execute it without re-reading the full audit.

## Precondition

There must be at least one prior audit in Engram (run via `/sequoia audit` or `/sequoia review`).

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

## Per-task format

Each generated task follows this mandatory structure:

```markdown
### [TASK-ID] · [Actionable title]

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

1. **Does not require reading the full audit** — all context is in the task
2. **Is not ambiguous** — a developer (or agent) can implement without questions
3. **Has verifiable acceptance criteria** — not "improve X," but "test Y passes"
4. **Declares explicit dependencies** — knows which tasks must go first
5. **Declares risk honestly** — not everything is "low risk"

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

Tasks are ordered following these criteria:

1. **Production blockers** → first (without exception)
2. **Root causes** → before their symptoms (from correlator)
3. **Technical dependencies** → if task B requires A to be done
4. **High leverage** → maximum impact with minimum change
5. **Implementation risk** → low risk first (quick wins)

## Deduplication rule

If multiple findings point to the same root cause (detected by the correlator), ONE task is generated that resolves all related findings. The finding IDs are listed as source.

## Output

Generates task files under `docs/sequoia/tasks/` using the same format and structure as `/sequoia audit`:

```
docs/sequoia/tasks/
├── index.md           # Global dependency graph, priority tiers, risk estimate
├── security.md        # Security tasks with full evidence
├── architecture.md    # Architecture tasks with full evidence
├── performance.md     # Performance tasks with full evidence
├── quality.md         # Quality tasks with full evidence
├── operations.md      # Operations tasks with full evidence
└── i18n.md            # i18n tasks with full evidence (if applicable)
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
