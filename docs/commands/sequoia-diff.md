---
description: "Compares current project state against the last recorded audit. Shows: resolved, new, worsened, unchanged. Supports --baseline, --since, --skip-scan, --only-verified."
argument-hint: "[--baseline=<audit-id>] [--since=<date>] [--skip-scan] [--only-verified]"
allowed-tools: Read, Glob, Grep
---

# /sequoia diff

Compares the current project state against the last audit recorded in Engram. Shows evolution: what improved, what worsened, what's new.

## Precondition

There must be at least one prior audit in Engram, **less than 60 days old**. If no prior audit exists, suggest running `/sequoia audit` first.

**Obsolescence policy**:
- 30-60 days old → Warning: "Baseline audit is {N} days old. Results may be incomplete."
- >60 days old → Error: "Baseline audit is too old (>60 days). Run `/sequoia audit` to establish a fresh baseline."

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
│     │   • >5 new files → moderate change
│     │   • >20 new files → major change (recommend full audit)
│     ├─ Did the stack or dependencies change?
│     └─ Did project maturity change?
│
├─ 3. Re-verify each previous finding
│     ├─ For each finding, read the cited files
│     ├─ Does the file still exist? If NOT → "resolved by deletion"
│     ├─ Is the evidence still present?
│     ├─ Was the recommendation implemented?
│     └─ Classify: resolved | unchanged | worsened | partial
│
├─ 4. Quick scan for new findings (unless --skip-scan)
│     ├─ Targeted scan: each agent checks only its highest-signal patterns
│     ├─ Only 🔴 and 🟠 findings (not a full audit)
│     └─ List as "new"

### Quick Scan Specification

| Agent | What it checks in quick scan | What it skips |
|-------|------------------------------|---------------|
| P1 Security | Hardcoded secrets (regex), missing CSP headers, SQL injection patterns in changed files | Full attack surface matrix, PII audit, token rotation analysis |
| P2 Performance | N+1 patterns in new code, missing image dimensions, render-blocking resources | Full bundle budgets, Core Web Vitals simulation, waterfall analysis |
| P3 Architecture | New circular imports, new god object signals in changed modules | Full dependency matrix, coupling analysis, API contract audit |
| P4 Quality | New CVEs in updated dependencies, new assertion-less tests | Transitive license tree, SBOM generation, full test smell audit |
| P5 Experience | Missing alt text in new components, div onClick without semantics | Full flow analysis, conversion funnel, WCAG AA compliance audit |
| P6 Operations | Secrets in new CI/CD configs, missing health checks in new services | Full pipeline audit, resilience patterns, backup verification |
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
- If the file no longer exists → ✅ Resolved by deletion
- If the file changed but the problem partially persists → 🔸 Partial
- If the file hasn't changed → ⏸️ Unchanged
- If the file changed and ANY of these criteria apply → 🔻 Worsened:
  1. **Severity increased**: finding severity escalated (medium → high)
  2. **Scope expanded**: affected lines increased by >50%
  3. **Impact increased**: new files/modules are now affected
  4. **New related problems**: changes introduced additional issues in the same area

## Flags

| Flag | Description |
|------|-------------|
| `--baseline=<audit-id>` | Compare against a specific audit by ID instead of the most recent |
| `--since=<date>` | Compare against the last audit before a given date (ISO 8601) |
| `--skip-scan` | Skip the quick scan for new findings — only re-verify previous findings |
| `--only-verified` | Only show findings that were verified (resolved/unchanged), skip new detections |

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

**Current score calculation**: Base score from previous audit, adjusted by:
- Resolved findings: +severity_weight per finding
- New findings: -severity_weight per finding
- Worsened findings: -(severity_weight × 1.5)

| Phase | Previous score | Current score | Trend |
|------|---------------|--------------|-----------|
| 🔒 Security | 72 | 78 | ↗️ +6 |
| ⚡ Performance | 50 | 45 | ↘️ -5 |
| 🏗️ Architecture | 88 | 92 | ↗️ +4 |
| ✅ Quality | 62 | 62 | → 0 |
| 🎨 Experience | 75 | 80 | ↗️ +5 |
| 🔧 Operations | 35 | 42 | ↗️ +7 |
| **GLOBAL** | **65** | **70** | **↗️ +5** |

### Trend Analysis

| Metric | Value |
|--------|-------|
| Total resolved | {N} (↑ from previous period) |
| Total worsened | {N} (↓ from previous period) |
| Net health change | +{N} or -{N} |
| Resolution rate | {resolved}/{total_previous} = {N}% |
| New finding rate | {new}/{total_current} = {N}% |

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
