# Flow: Full Audit

Workflow for comprehensive audits on medium and large projects.

## Preconditions

- `/sequoia init` executed and Project Map available in Engram
- If the init is more than 7 days old, refresh with a quick re-init

## Flow diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    /sequoia audit                                │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                    ┌──────▼──────┐
                    │ 1. REFRESH  │ Quick re-scan of Project Map
                    │   CONTEXT   │ (verify it's still current)
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ 2. SELECT   │ Applicable agents per
                    │   AGENTS    │ Project Map + flags
                    └──────┬──────┘
                           │
               ┌───────────┼───────────┐
               │           │           │
        ┌──────▼──────┐    │    ┌──────▼──────┐
        │ 3a. BATCH 1 │    │    │ 3a. BATCH 1 │
        │ P1 Security │    │    │ P4 Quality  │
        │ P2 Perform. │    │    └──────┬──────┘
        │ P3 Archit.  │    │           │
        └──────┬──────┘    │           │
               │           │           │
               └─────┬─────┘           │
                     │    (parallel)   │
                     └────────┬────────┘
                              │
                     ┌────────▼────────┐
                     │ 3b. BATCH 2     │
                     │ P5 Experience   │
                     │ P6 Operations   │
                     │ (use P3 output) │
                     └────────┬────────┘
                              │
               ┌──────────────┼──────────────┐
               │                             │
        ┌──────▼──────┐              ┌───────▼──────┐
        │ 4a. M1      │─────────────│ 4b. M2       │
        │ CORRELATOR  │             │ REPORTER     │
        │ cross-phase │             │ scoring+docs │
        └─────────────┘             └──────────────┘
                                            │
                                   ┌────────▼────────┐
                                   │ 5. DELIVERABLES │
                                   │ master.md       │
                                   │ phases/*.md     │
                                   │ score.md        │
                                   │ tasks.md        │
                                   └────────┬────────┘
                                            │
                                   ┌────────▼────────┐
                                   │ 6. ENGRAM       │
                                   │ Persist:        │
                                   │ findings+score  │
                                   │ + snapshot      │
                                   └─────────────────┘
```

## Per-step detail

### Step 1 — Context Refresh (~1-2 min)

Quick re-scan to verify the Project Map is still current:
- Were new dependencies added?
- Did the directory structure change?
- Are there relevant new files?

If significant changes are detected → re-run the corresponding init step.

### Step 2 — Agent Selection (~instant)

Determine agents to run:
- Without `--phase` → all marked as "applies" in the Project Map
- With `--phase` → only that agent
- With `--scope` → all applicable, but each limits its scope

### Step 3 — Phase agents (~10-25 min total)

**Batch 1 (parallel)** — no dependencies between them:
| Agent | Estimated time | Produces |
|--------|----------------|---------|
| P1 Security | 3-8 min | Security findings + attack matrix |
| P2 Performance | 3-8 min | Performance findings + budget |
| P3 Architecture | 5-10 min | Architecture findings + API design + dep map |
| P4 Quality | 3-6 min | Quality findings + deps + testing strategy |

**Batch 2 (after P3)** — use architecture output:
| Agent | Estimated time | Produces |
|--------|----------------|---------|
| P5 Experience | 3-6 min | UX + product findings |
| P6 Operations | 3-6 min | DevOps + data + infra findings |

### Step 4 — Meta-agents (~3-5 min total)

Always sequential in this order:

1. **M1 Correlator** (~1-2 min): Cross-references findings across phases, detects root causes
2. **M2 Reporter** (~1-2 min): Calculates health score by phase and global + generates all documents

### Step 5 — Deliverables (~1 min)

Generation of markdown files in `docs/sequoia/`.

### Step 6 — Engram (~instant)

Persist:
- Findings with timestamp and current commit hash
- Health scores for history
- State snapshot for future `/sequoia diff`

## Edge case decisions

### New dependencies detected during audit
If an agent discovers deps not mapped in init:
1. Note them as findings (P4 Quality)
2. Do not stop the audit
3. Suggest re-init at the end of the report

### Ambiguous or mixed stack (monorepo)
1. Run init for each sub-project if they are independent
2. If they share code, audit the shared module as cross-cutting
3. Reporter separates findings by sub-project

### Agent that cannot verify something
The agent marks the finding as `[NOT VERIFIABLE]` or `[REQUIRES EXTERNAL ACCESS]`.
The correlator does NOT correlate unverifiable findings. The reporter includes them in a separate section.

### Project without tests and without CI
This is not an error. P4 and P6 report the absence as findings.
The reporter marks those phases according to the real state, not aspirational.

## Total time estimate

| Project size | full | quick |
|----------------|------|-------|
| Small (< 50 files) | 10-15 min | 5-8 min |
| Medium (50-200 files) | 15-30 min | 8-15 min |
| Large (> 200 files) | 30-45 min | 12-20 min |

*With `--scope=module`, subtract ~60% of estimated time.*
