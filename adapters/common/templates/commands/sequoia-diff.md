---
description: "Compares current project state against the last recorded audit. Shows: resolved, new, worsened, unchanged. Useful for tracking project evolution."
allowed-tools: Read, Glob, Grep
---

# /sequoia diff

Compares the current project state against the last audit recorded in Engram. Shows evolution: what improved, what worsened, what's new.

## Precondition

There must be at least one prior audit in Engram. If no prior audit exists, suggest running `/sequoia audit` first.

## What it does

1. Retrieves the last audit from Engram
2. Runs a quick scan of the current project state
3. Compares previous findings vs current state
4. Classifies each finding into an evolution category
5. Generates the diff report

## Comparison categories

| Category | Meaning | Icon |
|-----------|-------------|-------|
| **Resolved** | The previous finding no longer reproduces | ✅ |
| **New** | Problem that didn't exist in the previous audit | 🆕 |
| **Worsened** | The previous finding persists and has worsened | 🔻 |
| **Unchanged** | The previous finding persists unchanged | ⏸️ |
| **Partially resolved** | Improved but doesn't meet acceptance criteria | 🔸 |

## Execution flow

```
/sequoia diff
  │
  ├─ 1. Retrieve last audit from Engram
  │     ├─ Findings with timestamp
  │     ├─ Health scores
  │     └─ Project Map snapshot
  │
  ├─ 2. Verify changes in project structure
  │     ├─ New or deleted files since the last audit?
  │     ├─ Did the stack or dependencies change?
  │     └─ Did project maturity change?
  │
  ├─ 3. Re-verify each previous finding
  │     ├─ For each finding, read the cited files
  │     ├─ Is the evidence still present?
  │     ├─ Was the recommendation implemented?
  │     └─ Classify: resolved | unchanged | worsened | partial
  │
  ├─ 4. Detect new findings
  │     ├─ Quick scan of areas not previously covered
  │     ├─ Only 🔴 and 🟠 findings (not a full audit)
  │     └─ List as "new"
  │
  └─ 5. Generate evolution report
        ├─ Summary table by category
        ├─ Health score comparison
        └─ Persist result in Engram
```

## Verification methodology

For each previous finding:

1. **Read the file cited in the evidence** — does it still exist? same lines?
2. **Verify the acceptance criteria** — were they met?
3. **Cross-check with git blame/log** — were there commits touching that area?

Classification:
- If the file changed and the problem is gone → ✅ Resolved
- If the file changed but the problem partially persists → 🔸 Partial
- If the file hasn't changed → ⏸️ Unchanged
- If the file changed and there are additional problems → 🔻 Worsened

## Output format

```markdown
## Sequoia Diff — [Project]

**Previous audit**: [date]
**Current comparison**: [date]
**Time elapsed**: [days/weeks]

### Evolution summary

| Category | Count | Percentage |
|-----------|----------|------------|
| ✅ Resolved | {N} | {N}% |
| 🔸 Partial | {N} | {N}% |
| ⏸️ Unchanged | {N} | {N}% |
| 🔻 Worsened | {N} | {N}% |
| 🆕 New | {N} | {N}% |
| **Total** | **{N}** | **100%** |

### Health Score comparison

| Phase | Previous score | Current score | Trend |
|------|---------------|--------------|-----------|
| Security | 🟠 | 🟢 | ↗️ Improving |
| Performance | 🟡 | 🟡 | → Stable |
| ... | | | |

### Detail of resolved findings ✅
{list of findings with what changed}

### Detail of new findings 🆕
{only 🔴 and 🟠 findings detected in the quick scan}

### Detail of worsened findings 🔻
{findings where the problem grew or new risks were added}

### Global trend
📈 Improving | ➡️ Stable | 📉 Degrading

### Recommendation
{when to run the next full audit}
```

## When to use diff vs new audit

| Situation | Use |
|-----------|------|
| You implemented fixes and want to verify | `diff` |
| 1-2 weeks passed and you want tracking | `diff` |
| Major changes in the project | `audit` (new audit) |
| More than a month passed | `audit` (new audit) |
| New team member | `audit` (new audit) |
| Post-merge of large feature | `diff` first, `audit` if there are surprises |

## Obsolescence detection

If the last audit is more than 30 days old, diff shows a warning:
> ⚠️ The last audit is {N} days old. Findings may be outdated. Consider running `/sequoia audit` for a fresh audit.

If the Project Map changed significantly (new deps, framework change, etc.), diff recommends running a new `init` + `audit`.
