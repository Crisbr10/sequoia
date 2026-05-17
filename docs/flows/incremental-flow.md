# Flow: Incremental Audit

Flow for re-audit and tracking of project evolution.

## When to use

| Situation | Command |
|-----------|---------|
| After implementing fixes from previous audit | `/sequoia diff` |
| Weekly/biweekly periodic health check | `/sequoia diff` |
| Post-merge of large feature | `/sequoia diff` → if there are surprises, `audit` |
| Significant changes in the project | `/sequoia audit` (new complete) |
| More than 30 days since last audit | `/sequoia audit` (new complete) |

## Incremental diff flow

```
/sequoia diff
  │
  ├─ 1. RETRIEVE PREVIOUS AUDIT
  │     ├─ Findings from Engram (most recent)
  │     ├─ Health scores
  │     └─ State snapshot (commit hash, structure)
  │
  ├─ 2. DETECT STALENESS
  │     ├─ How many commits since last audit?
  │     ├─ How many files changed?
  │     └─ Did the stack or structure change significantly?
  │
  ├─ 3. RE-VERIFY PREVIOUS FINDINGS
  │     ├─ For each prior finding:
  │     │   ├─ Read files cited in evidence
  │     │   ├─ Compare current state vs snapshot
  │     │   └─ Classify: ✅ | 🔸 | ⏸️ | 🔻
  │     └─ Generate classification table
  │
  ├─ 4. QUICK SCAN FOR NEW FINDINGS
  │     ├─ Only in areas changed since last audit
  │     ├─ Only 🔴 CRITICAL and 🟠 RISK
  │     └─ Not a full audit: quick sweep
  │
  ├─ 5. CALCULATE EVOLUTION
  │     ├─ Previous score vs current score (estimated)
  │     ├─ Trend: ↗️ Improving | → Stable | ↘️ Degrading
  │     └─ Resolution velocity (findings resolved / time)
  │
  └─ 6. GENERATE EVOLUTION REPORT
        └─ Diff format (see sequoia-diff.md)
```

## Staleness detection

```markdown
| Indicator | Green | Yellow | Red |
|-----------|-------|----------|------|
| Days since last audit | < 14 | 14-30 | > 30 |
| Commits since last audit | < 20 | 20-50 | > 50 |
| Files changed | < 15% | 15-40% | > 40% |
| Dep changes | 0 | 1-3 | > 3 |
| Structure change | No | Minor | Significant |
```

- **All green**: incremental diff is sufficient
- **Any yellow**: diff + attention to those areas
- **Any red**: recommend new full audit

## Incremental scope

The quick scan only re-audits areas that changed:

1. **Get file diff** from the last audit commit
2. **Filter** to source code files (exclude generated, vendor, lockfiles)
3. **For each changed file**, run only the agents relevant to that file type
4. **Do not re-run** agents on unchanged files

This reduces scan time from ~15-30 min to ~3-8 min.

## Evolution scoring

### Phase scoring

Compare previous score with current estimate:

```
🟢 → 🟢 = → Stable (maintains health)
🟡 → 🟢 = ↗️ Improving (resolved debt)
🟠 → 🟢 = ↗️↗️ Significant improvement
🟠 → 🟡 = ↗️ Improving
🟢 → 🟡 = ↘️ Slightly degrading
🟡 → 🟠 = ↘️ Degrading
🟢 → 🟠 = ↘️↘️ Significant degradation
🟢 → 🔴 = 🔻 Critical (requires immediate action)
```

### Global trend

```
Improvement rate = (resolved + partial) / total_prior_findings

📈 Improving:  rate > 30%
➡️ Stable:     rate 10-30%
📉 Degrading:  rate < 10%  OR  new > resolved
```

### Velocity score

```markdown
| Metric | Formula | Interpretation |
|---------|---------|----------------|
| Resolution rate | resolved / prior_findings | % progress |
| New debt rate | new / weeks_elapsed | appearance velocity |
| Net trend | (resolved - new) / weeks | net balance |
```

## Complete vs incremental audit

```
                 ┌──────────────────────────────┐
                 │  How much changed since the   │
                 │  last audit?                   │
                 └──────────┬───────────────────┘
                            │
                 ┌──────────▼───────────────────┐
                 │  Did stack or structure       │
              ┌──┤  change significantly?        │
              │  └──────────┬───────────────────┘
              │             │
         Yes  │        No   │
              │             │
    ┌─────────▼──┐   ┌─────▼──────────┐
    │ NEW FULL   │   │ Is staleness   │
    │ AUDIT      │   │ red?           │
    └────────────┘   └───┬──────┬─────┘
                         │      │
                    Yes  │   No │
                    ┌────▼──┐ ┌─▼──────────┐
                    |FULL   │ │ INCREMENTAL │
                    |AUDIT  │ │ DIFF        │
                    └───────┘ └─────────────┘
```

## Diff persistence

Each diff is saved in Engram with:
- **title**: "Sequoia Diff — {project} — {date}"
- **topic_key**: `sequoia/{project}/diff-{timestamp}`
- **type**: `architecture`
- **content**: complete diff result

This enables building an evolution history. The scorecard can show trends across multiple diffs.

## Integration with full audit

Diffs do NOT replace complete audits. They are complementary:

- **Complete audit**: baseline, exhaustive discovery, deep correlation
- **Incremental diff**: tracking, fix verification, early degradation detection

Suggested cadence:
- Complete audit: monthly or after major changes
- Incremental diff: weekly or post-fix
