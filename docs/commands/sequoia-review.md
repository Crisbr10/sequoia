---
description: "PR/diff-focused code review. Analyzes changed files, auto-selects relevant agents, detects impact on prior findings. Faster than audit, deeper than a linter."
argument-hint: "[--diff=HEAD~1..HEAD] [--pr=<number>] [--strict]"
allowed-tools: Read, Glob, Grep, Bash
---

# /sequoia review

PR-style code review. Analyzes recent changes, auto-selects relevant agents, and cross-references against prior findings.

## When to use

- Before merging a PR
- After a batch of changes
- As a pre-commit quality gate (with `--strict`)
- When you want fast feedback without a full audit

## Execution flow

```
/sequoia review
  │
  ├─ 1. Get changed files
  │     ├─ --diff=HEAD~3..HEAD → git diff in that range
  │     ├─ --pr=42 → gh pr diff 42
  │     └─ no flags → git diff HEAD~1..HEAD (last commit)
  │
  ├─ 2. Classify changed files by type
  │     ├── {auth,session,token,login} → P1 Security
  │     ├── {component,page,view,jsx,tsx,vue,svelte} → P5 Experience
  │     ├── {api,route,endpoint,controller,handler} → P3 Architecture
  │     ├── {model,schema,migration,entity,repository} → P3 Architecture, P6 Operations
  │     ├── {test,spec,__tests__} → P4 Quality
  │     ├── {Dockerfile,workflow,deploy,.github} → P6 Operations
  │     ├── {package.json,go.mod,Cargo.toml} → P4 Quality (deps)
  │     ├── {config,bundle,build,webpack,vite} → P2 Performance
  │     └── All files → P3 Architecture (always)
  │
  ├─ 3. Retrieve prior findings from Engram
  │     └─ Flag if changes touch areas with open findings
  │
  ├─ 4. Run selected agents (only on diff files)
  │
  ├─ 5. Generate focused output
  │     ├─ New findings from the diff
  │     ├─ Prior findings affected by the changes
  │     └─ Verdict: ✅ PASS | ⚠️ WARN | 🔴 BLOCK
  │
  └─ 6. Persist findings in Engram
```

## Flag reference

| Flag | Value | Default | Description |
|------|-------|---------|-------------|
| `--diff` | git range | `HEAD~1..HEAD` | Commit range to review |
| `--pr` | PR number | — | Gets PR diff via `gh` CLI |
| `--strict` | *(boolean flag)* | off | No tolerance for medium findings |

## `--strict` mode

With `--strict`:
- All 🟡 ATTENTION findings are reported (normally omitted in review)
- Findings that would normally be "acceptable debt" are marked as `WARN`
- If there is any 🟠 RISK, the verdict is `🔴 BLOCK`
- Useful as a merge gate on protected branches

Without `--strict`:
- Only 🔴 CRITICAL and 🟠 RISK are reported
- The verdict is `⚠️ WARN` if there are risks, `✅ PASS` if only attention

## Agent auto-selection table

| Pattern in changed files | Agents activated |
|------------------------------|-------------------|
| Auth/security files | P1, P4 |
| UI components / pages | P2, P3, P5 |
| Routes / endpoints | P3 |
| Models / migrations | P3, P6 |
| Tests | P4 |
| CI/CD / Docker / deploy | P6 |
| Dependency manifests | P4 |
| Build configuration | P2, P3 |
| **Always** | P3 (architecture) |

## Cross-reference with prior findings

For each changed file:
1. Search Engram for open findings on that file
2. If changes modify lines near a prior finding → mark `📋 PRIOR FINDING AFFECTED`
3. If changes resolve a prior finding → mark `✅ FINDING RESOLVED`
4. If changes don't touch the prior finding → do not mention (reduce noise)

## Output format

```markdown
## Sequoia Review — [range or PR]

**Files reviewed**: {N}
**Agents executed**: {list}

### 🔴 Blocking
{list of critical findings, if any}

### 🟠 Risks
{list of risk findings}

### 🟡 Attention {only with --strict}
{list of attention findings}

### 📋 Prior findings affected
{if changes touch areas with open findings}

### ✅ Resolved findings
{if changes fix prior findings}

---
**Verdict**: {✅ PASS | ⚠️ WARN | 🔴 BLOCK}
```

## Difference from `/sequoia audit`

| Aspect | audit | review |
|---------|-------|--------|
| Scope | entire project | only changed files |
| Agents | all applicable | only relevant to diff |
| Time | 15-45 min | 2-8 min |
| Depth | complete | focused on changes |
| Correlation | complete across phases | only between diff findings |
| Meta-agents | all | only simplified correlator |
| Output | full reports | focused findings + verdict |

## Examples

```bash
# Review last commit
/sequoia review

# Review last 3 commits
/sequoia review --diff=HEAD~3..HEAD

# Review a specific PR
/sequoia review --pr=42

# Strict review as merge gate
/sequoia review --pr=42 --strict
```
