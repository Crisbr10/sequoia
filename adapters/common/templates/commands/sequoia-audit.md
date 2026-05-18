---
description: "Runs a full technical audit. Runs phase agents in parallel, then meta-agents for correlation and reports. Supports flags: --phase, --scope, --mode, --output."
argument-hint: "[--phase=security|performance|architecture|quality|experience|operations] [--scope=changed|module=<path>] [--mode=full|quick] [--output=report|tasks|both]"
allowed-tools: Read, Glob, Grep, Bash
---

# /sequoia audit

Runs the comprehensive technical audit. Orchestrates phase agents, correlation meta-agents, and generates complete deliverables.

## Precondition

`/sequoia init` must have been run previously. If there is no Project Map in Engram, ask the user to run init first.

## Execution flow

```
/sequoia audit
  │
  ├─ 1. Retrieve Project Map from Engram
  │     └─ If not found → ERROR: "Run /sequoia init first"
  │
  ├─ 2. Select agents based on flags + Project Map
  │     ├─ Without --phase → all applicable agents
  │     ├─ With --phase → only that agent + meta-agents
  │     └─ Non-applicable agents → skipped with reason
  │
  ├─ 3. Run phase agents
  │     ├─ Parallel: P1, P2, P3, P4 (no dependencies between them)
  │     ├─ After: P5, P6 (use P3 findings)
  │     └─ All applicable per Project Map
  │
  ├─ 4. Run meta-agents
  │     ├─ M1 sequoia-correlator (cross-references findings across phases)
  │     └─ M2 sequoia-reporter (calculates health scores + generates documents)
  │
  ├─ 5. Generate deliverables
  │     ├─ docs/sequoia/summary.md (score + root causes + trajectory + verified state + gaps)
  │     └─ docs/sequoia/tasks/{area}.md + index.md (self-contained tasks per area)
  │
  └─ 6. Persist in Engram
        ├─ Findings with timestamp
        ├─ Health scores
        └─ State snapshot for future diff
```

## Flag reference

| Flag | Values | Default | Description |
|------|---------|---------|-------------|
| `--phase` | `security` `performance` `architecture` `quality` `experience` `operations` | all | Run only a specific phase |
| `--scope` | `changed` `module=<path>` | entire project | Limit audit scope |
| `--mode` | `full` `quick` | `full` | Analysis depth |
| `--output` | `report` `tasks` `both` | `both` | Type of deliverable to generate |

## `--mode` differences

### `full` (default)
- All applicable agents
- Deep analysis per agent
- Findings of all severities
- Full performance budgets
- Module dependency map
- Full attack surface matrix
- Estimated time: 15-45 min depending on size

### `quick`
- Only 🔴 CRITICAL and 🟠 RISK findings
- No additional deliverables (budgets, maps, matrices)
- Agents reduced to highest-impact inspections
- No deep correlation (simplified correlator)
- Estimated time: 5-15 min depending on size

## `--scope` options

| Value | What it does |
|-------|----------|
| *(no flag)* | Audits the entire project |
| `changed` | Only files modified vs last commit. Uses `git diff --name-only HEAD` |
| `module=src/auth` | Only the indicated module and its subdirectories |

With `--scope=changed`, each agent only inspects files from the diff. Meta-agents correlate only against those findings.

## `--output` options

| Value | Generates |
|-------|--------|
| `report` | `docs/sequoia/summary.md` |
| `tasks` | `docs/sequoia/tasks/*.md` (area files + index) |
| `both` | All of the above (default) |

With `--output=report`, only `summary.md` is generated (tasks are skipped). With `--output=tasks`, only area task files are generated (`summary.md` is skipped). The default `both` generates all deliverables.

## Parallelism logic

### Can run in parallel (no dependencies):
- P1 Security, P2 Performance, P3 Architecture, P4 Quality

### Must run after P3:
- P5 Experience (uses architecture map)
- P6 Operations (uses architecture model)

### Meta-agents (always sequential):
- M1 Correlator → M2 Reporter (in that order; scoring is part of M2)

## Orchestrator delegation

The orchestrator delegates to each agent by providing:
1. The complete **Project Map**
2. The applicable **scope** (all, changed files, or module)
3. The **mode** (full or quick)
4. The standard **finding template** (from `references/finding-format.md`)

Each agent returns its findings in the standard format. The orchestrator does not interpret findings, only routes them.

## Generated deliverables

All are created in the configured directory (default: `docs/sequoia/`):

```
docs/sequoia/
├── summary.md                  # Health score + root causes + verified state + missing items + trajectory
└── tasks/
    ├── index.md                # Global dependency graph, priority tiers, risk estimate
    ├── security.md             # Self-contained task file (P1 findings)
    ├── performance.md          # Self-contained task file (P2 findings)
    ├── architecture.md         # Self-contained task file (P3 findings)
    ├── quality.md              # Self-contained task file (P4 findings)
    ├── experience.md           # Self-contained task file (P5 findings, if applicable)
    ├── operations.md           # Self-contained task file (P6 findings)
    └── i18n.md                 # Self-contained task file (P7 findings, if applicable)
```

Each area task file is self-contained: an implementing agent opens ONE file (~150-250 lines) instead of the full report.

## Usage examples

```bash
# Full audit
/sequoia audit

# Security only, quick mode
/sequoia audit --phase=security --mode=quick

# Changed files only, report only
/sequoia audit --scope=changed --output=report

# Deep module audit
/sequoia audit --scope=module=src/auth --mode=full

# Generate only quality tasks
/sequoia audit --phase=quality --output=tasks
```
