# SEQ-REF: Delegar escritura de archivos a agentes de fase

**Objetivo**: Cada agente de fase (P1-P6) escribe su propio `.sequoia/{phase}.md` al terminar. M2 solo genera `summary.md`. Se elimina `tasks/`.

**Archivos base** (todos bajo `C:\Users\Usuario\.config\opencode\`):
- `skills\sequoia\SKILL.md`
- `commands\sequoia-audit.md`
- `commands\sequoia-fix.md`
- `commands\sequoia-dev.md`

---

## Tarea 1: SEQ-REF-001 — Phase 3: cada agente escribe su archivo

**Archivo**: `skills\sequoia\SKILL.md`
**Depende de**: nada
**Por qué**: Núcleo del refactor. Sin esto los agentes siguen dependiendo de M2 para persistir.

**Cambio**: Reemplazar el bloque `**Each agent must**:` (líneas 135-142) por el siguiente:

OLD (líneas 135-142):
```
**Each agent must**:
1. Scan files relevant to its domain
2. Document each finding with concrete evidence
3. Classify severity based on real impact in THIS project
4. Limit findings to those with direct evidence
5. Deliver findings in the standard format

**Checkpoint**: If an agent produces no findings, explicitly report "no findings in domain". It's not an error — it's information.
```

NEW:
```
**Each agent must**:
1. Scan files relevant to its domain
2. Document each finding with concrete evidence
3. Classify severity based on real impact in THIS project
4. Limit findings to those with direct evidence
5. **Write `.sequoia/{phase}.md`** — a self-contained file with ALL findings and tasks — BEFORE returning to the orchestrator
6. Deliver findings in the standard format to the orchestrator

**Phase Agent Output File** — each agent writes to `.sequoia/{phase}.md`:

Phase → file mapping:
- P1 Security → `.sequoia/security.md`
- P2 Performance → `.sequoia/performance.md`
- P3 Architecture → `.sequoia/architecture.md`
- P4 Quality → `.sequoia/quality.md`
- P5 Experience → `.sequoia/experience.md`
- P6 Operations → `.sequoia/operations.md`

File format (self-contained, no external references needed):

\`\`\`markdown
# {Phase} Audit — {project_name}

## Score: {score}/100

### Critical Findings
{full detail with evidence — file, line, code, explanation, impact}

### High Findings
{full detail with evidence}

### Medium Findings
{concise — title, file, line, one-line impact}

### Low Findings
{summary only — title and file}

## Tasks
### [{PHASE}-{NNN}] · {Actionable title}
**Status**: ⏳ Pending
**Priority**: 🔴 Blocking | 🟠 High leverage | 🟡 Backlog
**Source finding(s)**: {finding IDs this task resolves}
**Minimum context**: {3-5 lines explaining WHAT is wrong and WHY it matters}
**Files involved**:
- \`path/to/file\` — {role}
**What to do**:
1. {concrete step}
2. {concrete step}
**Acceptance criteria**:
- [ ] {verifiable condition}
\`\`\`

**Checkpoint**: If an agent produces no findings, it MUST STILL write the file with "## No findings in domain — {phase}" and an empty Tasks section.
```

---

## Tarea 2: SEQ-REF-002 — Eliminar template "Phase Document" de M2

**Archivo**: `skills\sequoia\SKILL.md`
**Depende de**: SEQ-REF-001
**Por qué**: La template "Phase Document" (líneas 1847-1868) estaba en M2 porque antes M2 la generaba. Ahora la genera cada phase agent. La template ya está incluida en SEQ-REF-001. Hay que eliminar la copia obsoleta.

**Cambio**: Eliminar líneas 1847-1868 (el bloque `### Phase Document (One per Phase)` completo).

OLD (líneas 1847-1868):
```
### Phase Document (One per Phase)

\`\`\`markdown
# {Phase} Audit — {project_name}

## Score: {score}/100

### Critical Findings
{critical_findings_with_details}

### High Findings
{high_findings_with_details}

### Medium Findings
{medium_findings_concise}

### Low Findings
{low_findings_summary}

## Recommendations
{ordered_recommendations}
\`\`\`
```

NEW: (eliminar el bloque completo — no se reemplaza con nada).

---

## Tarea 3: SEQ-REF-003 — M2 Reporter solo genera summary.md

**Archivo**: `skills\sequoia\SKILL.md`
**Depende de**: SEQ-REF-001
**Por qué**: M2 ya no genera archivos de fase. Su única salida es summary.md con correlation embebida.

**Cambio**: Reemplazar desde `**Input**: Correlated findings + Project Map` (línea 168) hasta el inicio de Phase 6 (línea 216).

OLD (líneas 168-216):
```
**Input**: Correlated findings + Project Map

**Actions**:
1. **Calculate Health Score** by category and globally:

\`\`\`
health_score:
  global: [0-100]
  categories:
    security: [0-100]
    performance: [0-100]
    architecture: [0-100]
    quality: [0-100]
    experience: [0-100|N/A]
    operations: [0-100|N/A]

  methodology: >
    score = 100 − Σ(severity_weight × scope_multiplier), floored at 0
    severity_weight: critical=15, high=8, medium=4, low=2, info=0
    scope_multiplier: 1.0 (isolated finding) | 1.5 (shared root cause across ≥2 findings)
    See references/scoring-criteria.md for full formula, grade table, and worked example.
\`\`\`

2. **Generate prioritized action plan**:

\`\`\`yaml
action_plan:
  immediate:  # critical + high, ordered by dependencies
    - finding_id: [ID]
      action: [what to do]
      blocks: [IDs that are unblocked by resolving this]
  short_term: # medium
    - finding_id: [ID]
      action: [what to do]
  long_term:  # low + info
    - finding_id: [ID]
      action: [what to do]
\`\`\`

3. **Generate final report** with structure:
   - Executive summary (3-5 sentences)
   - Health Score with breakdown
   - Critical and high findings (with full evidence)
   - Identified root causes
   - Prioritized action plan
   - Findings by category (full detail)

**Output**: Complete report + Health Score + Action plan
```

NEW:
```
**Input**: Correlated findings (from M1) + Project Map

**Actions**:
1. **Calculate Health Score** by category and globally (methodology unchanged from `references/scoring-criteria.md`)
2. **Embed M1 correlation results**: root causes, dedup summary, severity recalibrations — as section inside summary.md (no separate file)
3. **Generate prioritized action plan** referencing tasks from `.sequoia/{phase}.md` files
4. **Write `.sequoia/summary.md`** — the ONLY file M2 produces

**Output**: `.sequoia/summary.md` with structure:

\`\`\`markdown
# Sequoia Audit Report — {project_name}

**Date**: {date} | **Stack**: {stack} | **Size**: {size} | **Maturity**: {maturity}

## Global Health Score: {score}/100 — {classification}

### Scores by Phase
| Phase | Score | Classification | Findings | File |
|-------|-------|----------------|-----------|------|
| Security | {score} | {class} | {count} | [security.md](security.md) |
| Performance | {score} | {class} | {count} | [performance.md](performance.md) |
| Architecture | {score} | {class} | {count} | [architecture.md](architecture.md) |
| Quality | {score} | {class} | {count} | [quality.md](quality.md) |
| Experience | {score} | {class} | {count} | [experience.md](experience.md) |
| Operations | {score} | {class} | {count} | [operations.md](operations.md) |

### Root Causes (from M1 Correlator)
{correlation chains — embedded, not a separate file}

### Prioritized Roadmap
| Priority | Task ID | Title | Phase | Resolves | Effort |
|----------|---------|-------|-------|----------|--------|
| P0 | {ID} | {title} | {phase} | {finding IDs} | {effort} |

### Phase Details
See individual phase files for full findings and tasks:
- [security.md](security.md)
- [performance.md](performance.md)
- [architecture.md](architecture.md)
- [quality.md](quality.md)
- [experience.md](experience.md)
- [operations.md](operations.md)
\`\`\`
```

---

## Tarea 4: SEQ-REF-004 — Simplificar Phase 6 Delivery

**Archivo**: `skills\sequoia\SKILL.md`
**Depende de**: SEQ-REF-003
**Por qué**: Quitar referencia obsoleta a `.sequoia/tasks/`.

**Cambio**: En Phase 6 Delivery (líneas 217-225), reemplazar el ítem 5.

OLD (línea 224):
```
5. **Tasks** are automatically generated in `.sequoia/tasks/`
```

NEW:
```
5. **Files generated**: `.sequoia/summary.md` and `.sequoia/{phase}.md` for each applicable phase
```

---

## Tarea 5: SEQ-REF-005 — Actualizar Implementation Notes

**Archivo**: `skills\sequoia\SKILL.md`
**Depende de**: SEQ-REF-003
**Por qué**: La nota sobre persistencia (línea 312) dice que todo está en memoria hasta el final. Ya no es cierto.

**Cambio**: Reemplazar la línea 312.

OLD (línea 312):
```
- All state is kept in memory during the session. Between sessions, audit artifacts persist in `.sequoia/audit_YYYY-MM-DD/` inside the audited project. Engram serves as backup storage.
```

NEW:
```
- Phase agents write `.sequoia/{phase}.md` on completion — findings are persisted immediately, not only at the end.
- M2 writes `.sequoia/summary.md` after correlation.
- Between sessions, audit artifacts persist in `.sequoia/audit_YYYY-MM-DD/`. Engram serves as backup storage.
```

---

## Tarea 6: SEQ-REF-006 — Actualizar sequoia-audit.md

**Archivo**: `commands\sequoia-audit.md`
**Depende de**: SEQ-REF-001..005
**Por qué**: El comando audit es el entry point. Debe reflejar el nuevo flujo.

### Cambio 6a — Step 3 del flujo (líneas 33-35)

OLD:
```
  ├─ 3. Run phase agents
  │     ├─ Parallel: P1, P2, P3, P4 (no dependencies between them)
  │     ├─ After: P5, P6 (use P3 findings)
  │     └─ All applicable per Project Map
```

NEW:
```
  ├─ 3. Run phase agents
  │     ├─ Parallel: P1, P2, P3, P4 (no dependencies between them)
  │     │   └─ Each agent writes `.sequoia/{phase}.md` on completion
  │     ├─ After: P5, P6 (use P3 findings)
  │     │   └─ Each agent writes `.sequoia/{phase}.md` on completion
  │     └─ All applicable per Project Map
```

### Cambio 6b — Step 5 del flujo (líneas 41-43)

OLD:
```
  ├─ 5. Generate deliverables
  │     ├─ .sequoia/summary.md (score + root causes + trajectory + verified state + gaps)
  │     └─ .sequoia/tasks/{area}.md + index.md (self-contained tasks per area)
```

NEW:
```
  ├─ 5. Generate deliverables
  │     └─ .sequoia/summary.md (score + root causes embedded + action plan + phase file refs)
```

### Cambio 6c — Sección Generated deliverables (líneas 146-159)

OLD:
```
## Generated deliverables

All are created in the configured directory (default: `.sequoia/`):

\`\`\`
.sequoia/
├── summary.md                  # Health score + root causes + verified state + missing items + trajectory
└── tasks/
    ├── index.md                # Global dependency graph, priority tiers, risk estimate
    ├── security.md             # Self-contained task file (P1 findings)
    ├── performance.md          # Self-contained task file (P2 findings)
    ├── architecture.md         # Self-contained task file (P3 findings)
    ├── quality.md              # Self-contained task file (P4 findings)
    ├── experience.md           # Self-contained task file (P5 findings, if applicable)
    └── operations.md           # Self-contained task file (P6 findings)
\`\`\`

Each area task file is self-contained: an implementing agent opens ONE file (~150-250 lines) instead of the full report.
```

NEW:
```
## Generated deliverables

All are created in `.sequoia/`:

\`\`\`
.sequoia/
├── summary.md          # M2 — health score + root causes embedded + action plan + phase file refs
├── security.md         # P1 — findings + tasks (self-contained)
├── performance.md      # P2 — findings + tasks (self-contained)
├── architecture.md     # P3 — findings + tasks (self-contained)
├── quality.md          # P4 — findings + tasks (self-contained)
├── experience.md       # P5 — findings + tasks (self-contained)
└── operations.md       # P6 — findings + tasks (self-contained)
\`\`\`

Each phase file is self-contained: an implementing agent opens ONE file to understand
all findings and tasks for that domain. No cross-referencing required.
```

### Cambio 6d — Task Generation Requirements (líneas 163-172)

OLD:
```
## 🔴 Task Generation Requirements (NON-NEGOTIABLE)

When the Reporter (M2) generates tasks, every single task MUST include:

1. **\`[TASK-ID]\`** — format \`{PHASE}-{NNN}\` (e.g., \`P1-003\`, \`P3-012\`, \`M1-001\`). This is the task's permanent reference for all Sequoia commands and dependency tracking.
2. **\`**Status**\`** — \`⏳ Pending\` or \`✅ Resolved\`. New tasks always default to \`⏳ Pending\`. The status enables progress tracking across audits and \`/sequoia-dev\` sessions.

**A task without both \`[TASK-ID]\` and \`**Status**\` is INVALID.** The Reporter MUST NOT output a task unless both fields are present. The orchestrator MUST reject any task file that is missing either field and request regeneration.

The task format is defined in the Sequoia SKILL.md (\`Task Plan Format\` section). The Reporter agent (\`docs/agents/sequoia-reporter.md\`) contains the canonical Task Plan Format with \`id\` and \`status\` as the first two fields.
```

NEW:
```
## 🔴 Task Generation Requirements (NON-NEGOTIABLE)

When a phase agent generates tasks in `.sequoia/{phase}.md`, every single task MUST include:

1. **\`[TASK-ID]\`** — format \`{PHASE}-{NNN}\` (e.g., \`P1-003\`, \`P3-012\`). This is the task's permanent reference for all Sequoia commands and dependency tracking.
2. **\`**Status**\`** — \`⏳ Pending\` or \`✅ Resolved\`. New tasks always default to \`⏳ Pending\`. The status enables progress tracking across audits and \`/sequoia-dev\` sessions.

**A task without both \`[TASK-ID]\` and \`**Status**\` is INVALID.** The phase agent MUST NOT output a task unless both fields are present. The orchestrator MUST reject any phase file that contains tasks missing either field and request regeneration.

The task format is defined in the Sequoia SKILL.md (\`Task Plan Format\` section).
```

---

## Tarea 7: SEQ-REF-007 — Actualizar paths en sequoia-fix.md

**Archivo**: `commands\sequoia-fix.md`
**Depende de**: SEQ-REF-006
**Por qué**: `/sequoia fix` regenera tareas. Sus paths de salida referencian `tasks/` que ya no existe.

**Cambio**: Reemplazar sección Output (líneas 139-153).

OLD:
```
## Output

Generates task files under `.sequoia/tasks/` using the same format and structure as `/sequoia audit`:

\`\`\`
.sequoia/tasks/
├── index.md           # Global dependency graph, priority tiers, risk estimate
├── security.md        # Security tasks with full evidence
├── architecture.md    # Architecture tasks with full evidence
├── performance.md     # Performance tasks with full evidence
├── quality.md         # Quality tasks with full evidence
└── operations.md      # Operations tasks with full evidence
\`\`\`

When filtering by phase (`/sequoia fix security`), only the corresponding area file is generated. When running `all`, all area files + `index.md` are generated.

Each task follows the standard template defined in Sequoia's SKILL.md. See `/sequoia audit` for the primary workflow.
```

NEW:
```
## Output

Generates task content appended to `.sequoia/{phase}.md` files using the same format as `/sequoia audit`:

\`\`\`
.sequoia/
├── security.md        # Security tasks appended to existing findings
├── architecture.md    # Architecture tasks
├── performance.md     # Performance tasks
├── quality.md         # Quality tasks
└── operations.md      # Operations tasks
\`\`\`

When filtering by phase (`/sequoia fix security`), only the corresponding file is updated.
When running `all`, all phase files are updated.

Each task follows the standard template defined in Sequoia's SKILL.md — Task Plan Format.
```

---

## Tarea 8: SEQ-REF-008 — Actualizar phase mapping en sequoia-dev.md

**Archivo**: `commands\sequoia-dev.md`
**Depende de**: SEQ-REF-007
**Por qué**: `/sequoia-dev` busca archivos de tarea por fase. La tabla de mapeo debe apuntar a la nueva ubicación aplanada.

### Cambio 8a — Tabla de phase mapping (líneas 197-207)

OLD:
```
| Prefix | File            |
| ------ | --------------- |
| P1     | security.md     |
| P2     | performance.md  |
| P3     | architecture.md |
| P4     | quality.md      |
| P5     | experience.md   |
| P6     | operations.md   |
| M1     | index.md        |
| M2     | index.md        |
```

NEW:
```
| Prefix | Phase        | File            | Path                      |
|--------|-------------|-----------------|---------------------------|
| P1     | Security     | security.md     | .sequoia/security.md      |
| P2     | Performance  | performance.md  | .sequoia/performance.md   |
| P3     | Architecture | architecture.md | .sequoia/architecture.md  |
| P4     | Quality      | quality.md      | .sequoia/quality.md       |
| P5     | Experience   | experience.md   | .sequoia/experience.md    |
| P6     | Operations   | operations.md   | .sequoia/operations.md    |
| M1     | Correlator   | summary.md      | .sequoia/summary.md       |
| M2     | Reporter     | summary.md      | .sequoia/summary.md       |
```

### Cambio 8b — Config default paths.tasks_dir (línea 145)

OLD:
```
  tasks_dir: docs/sequoia/tasks/
```

NEW:
```
  tasks_dir: .sequoia/
```

### Cambio 8c — Lookup path (línea 212)

OLD:
```
Lookup path:

\`\`\`text
{paths.tasks_dir}/{area_file}
\`\`\`
```

NEW:
```
Lookup path:

\`\`\`text
{paths.tasks_dir}{area_file}
\`\`\`

Example: `.sequoia/security.md`
```

---

## Orden de ejecución

| # | Tarea | Archivo | Depende de |
|---|-------|---------|------------|
| 1 | SEQ-REF-001 | skills\sequoia\SKILL.md | — |
| 2 | SEQ-REF-002 | skills\sequoia\SKILL.md | SEQ-REF-001 |
| 3 | SEQ-REF-003 | skills\sequoia\SKILL.md | SEQ-REF-001 |
| 4 | SEQ-REF-004 | skills\sequoia\SKILL.md | SEQ-REF-003 |
| 5 | SEQ-REF-005 | skills\sequoia\SKILL.md | SEQ-REF-003 |
| 6 | SEQ-REF-006 | commands\sequoia-audit.md | SEQ-REF-001..005 |
| 7 | SEQ-REF-007 | commands\sequoia-fix.md | SEQ-REF-006 |
| 8 | SEQ-REF-008 | commands\sequoia-dev.md | SEQ-REF-007 |

## Verificación post-implementación

- `skills\sequoia\SKILL.md` no debe contener `tasks/` ni `tasks_dir`
- `skills\sequoia\SKILL.md` Phase 3 debe mencionar `.sequoia/{phase}.md`
- `skills\sequoia\SKILL.md` Phase 5 debe mencionar que M2 solo genera `summary.md`
- `skills\sequoia\SKILL.md` no debe tener la template "Phase Document" duplicada
- `commands\sequoia-audit.md` deliverables debe mostrar estructura aplanada
- `commands\sequoia-fix.md` output debe referenciar `.sequoia/{phase}.md`
- `commands\sequoia-dev.md` phase mapping debe tener columna Path con `.sequoia/`
