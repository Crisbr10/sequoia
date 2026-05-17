# Flow: PR Review

Focused flow for PR and diff review.

## Trigger

The user runs `/sequoia review` with optional flags.

## Flow diagram

```
/sequoia review [--diff=X] [--pr=N] [--strict]
  │
  ├─ 1. GET DIFF
  │     ├─ --diff=HEAD~3..HEAD → git diff --stat + content
  │     ├─ --pr=42 → gh pr diff 42
  │     └─ default → git diff HEAD~1..HEAD
  │
  ├─ 2. CLASSIFY FILES
  │     └─ Map each changed file to agent types
  │
  ├─ 3. AUTO-SELECT AGENTS
  │     ├─ Consult file→agent mapping table
  │     └─ Minimum: P3 Architecture (always runs)
  │
  ├─ 4. RETRIEVE PRIOR FINDINGS
  │     ├─ Search Engram for findings affecting diff files
  │     └─ Mark prior findings that changes could resolve or affect
  │
  ├─ 5. RUN AGENTS (only on diff files)
  │     ├─ All in parallel (no dependencies in review)
  │     └─ Each agent only analyzes changed files
  │
  ├─ 6. GENERATE FOCUSED OUTPUT
  │     ├─ New findings from the diff
  │     ├─ Prior findings affected/resolved
  │     └─ Final verdict
  │
  └─ 7. PERSIST
        └─ Save review findings in Engram (separate from audit)
```

## Auto-selection table: file type → agents

| File pattern | Agents activated | Justification |
|-------------------|-------------------|---------------|
| `**/auth/**`, `**/session/**`, `**/middleware/**` | P1 | Auth changes are always security-sensitive |
| `**/*.jsx`, `**/*.tsx`, `**/*.vue`, `**/*.svelte`, `**/*.html` | P2, P5 | UI components: performance + experience |
| `**/api/**`, `**/routes/**`, `**/controllers/**`, `**/handlers/**` | P1, P3 | Endpoints: security + architecture + API design |
| `**/models/**`, `**/schema/**`, `**/migrations/**`, `**/entities/**` | P3, P6 | Data: architecture + operations |
| `**/*.test.*`, `**/*.spec.*`, `**/__tests__/**` | P4 | Tests: quality |
| `Dockerfile*`, `**/.github/**`, `**/deploy/**`, `docker-compose.*` | P6 | Infra: operations |
| `package.json`, `go.mod`, `Cargo.toml`, `requirements.txt` | P4 | Deps: always review dependency changes (part of quality) |
| `vite.config.*`, `webpack.config.*`, `tsconfig.*` | P2, P3 | Build config: performance + architecture |
| `*.css`, `*.scss`, `*.less` | P2, P5 | Styles: performance + experience |
| `**/*.md` | — | Documentation: don't audit (unless API docs) |
| **Any other** | P3 | Architecture always has something to say |

## Selection rules

1. **P3 Architecture always runs** — every change has architectural impact
2. **P4 Quality runs if the manifest changed** — new deps = new risk
3. **P1 Security runs if there's auth, endpoints, or input handling** — security is non-negotiable
4. **P5 Experience only if there's UI** — if the diff is backend-only, skip
5. **Meta-agents**: only simplified correlator, no full reporter

## Focused output format

```markdown
## Sequoia Review

**Range**: [diff range or PR #N]
**Files**: {N} changed files
**Agents**: {list of agents executed}

---

### 🔴 Blocking
{critical findings from the diff}

### 🟠 Risks
{risk findings from the diff}

### 🟡 Attention [--strict only]
{medium findings, only in strict mode}

---

### 📋 Prior findings affected

| Prior finding | Status | Detail |
|----------------|--------|---------|
| [P1-002] Token without expiration | 🔸 Partially resolved | Added expiration but missing rotation |
| [P3-005] God module auth | ⏸️ Not touched | The diff doesn't modify auth/index.ts |

### ✅ Findings resolved by this diff
{if the change fixes a prior finding}

---

**Verdict**: {✅ PASS | ⚠️ WARN | 🔴 BLOCK}
**Suggested review**: {optional comment on the change as a whole}
```

## Verdict logic

| Condition | Verdict |
|-----------|-----------|
| No 🔴 or 🟠 findings | ✅ PASS |
| Only 🟠 findings and `--strict` not active | ⚠️ WARN |
| 🔴 findings | 🔴 BLOCK |
| 🟠 findings and `--strict` active | 🔴 BLOCK |
| No findings but prior findings worsened | ⚠️ WARN |

## Integration with prior findings

1. **Retrieve** from Engram the findings that cite diff files
2. **Verify** if the changed lines are near prior findings
3. **Mark** each prior finding as:
   - `✅ Potentially resolved` — if the diff touches the lines cited in the evidence
   - `🔸 Partially affected` — if the diff touches the file but not the exact lines
   - `🔻 Worsened` — if the diff aggravates the problem
   - `⏸️ Not touched` — if the diff is unrelated
4. Only report those marked as resolved, affected, or worsened (reduce noise)

## Flags in areas with open findings

If the diff touches files with open 🔴 or 🟠 findings, add to the output:

> ⚠️ **This diff modifies areas with open findings:**
> - `src/auth/handler.ts` has [P1-002] 🔴 Token without expiration
> - `src/api/routes.ts` has [P3-001] 🟠 Endpoints without pagination
>
> Verify that changes don't worsen these findings.
