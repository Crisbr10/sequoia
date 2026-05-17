# Sequoia — Code Audit and Review Framework

> "A sequoia tree doesn't grow in haste. It grows with deep roots."
> Sequoia doesn't audit to sound smart. It audits so the project can be intervened.

---

## Vision

Sequoia is a comprehensive technical audit framework designed as a Claude Code plugin. It works as a team of specialized architects inspecting a project from multiple angles in parallel or in phases, without assuming specific technology, without enterprise filler, and with concrete evidence from the repository as the sole source of truth.

**Core principle**: every finding must be traceable to a real file, a real line, or a documentable absence. What cannot be verified is explicitly declared as unverifiable.

---

## Design Philosophy

| Principle | Description |
|-----------|-------------|
| **Evidence over opinion** | No finding without a repo citation. Not one. |
| **Context over dogma** | Sequoia detects the stack and adapts the analysis. It doesn't apply React rules to a Go project. |
| **Root cause over symptom** | Distinguishes between what is seen and what causes it. |
| **Absolute actionability** | Every recommendation has an owner, an acceptance criterion, and an estimated risk. |
| **Agent separation** | Each domain has its own agent. There's no agent that knows everything moderately; there are agents that know one domain deeply. |
| **Prioritizable debt** | Not all debt is equal. Sequoia classifies: blocking, high leverage, backlog, acceptable. |

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    SEQUOIA ORCHESTRATOR                      │
│         Detects context · Coordinates agents · Synthesizes   │
└────────────────────────┬────────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        ▼                ▼                ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ Context Agent│ │ Phase Agents │ │ Meta Agents  │
│  (pre-flight)│ │  (1-6)       │ │  (post-run)  │
└──────────────┘ └──────────────┘ └──────────────┘
```

### Layer 0 — Context Agent (Pre-flight)

Before any audit, `sequoia-context` runs automatically and builds the **Project Map**:

- Detected stack (language, framework, runtime, bundler)
- Dominant paradigm (SPA, SSR, API, CLI, monolith, microservice, etc.)
- Project size (LOC, modules, dependencies)
- Presence of tests, CI/CD, documentation
- Dependency health status
- Estimated project maturity (prototype / in development / production)

This map is passed as context to ALL subsequent agents. Each agent adapts its criteria according to the map.

### Layer 1 — Phase Agents

Six specialized agents, each owning a domain:

| ID | Agent | Domain |
|----|--------|---------|
| P1 | `sequoia-security` | Security, authentication, attack surface |
| P2 | `sequoia-performance` | Performance, bundles, metrics, load |
| P3 | `sequoia-architecture` | Architecture, scalability, patterns, boundaries |
| P4 | `sequoia-quality` | Testing, coverage, code quality, contracts |
| P5 | `sequoia-experience` | User experience, accessibility, flows |
| P6 | `sequoia-operations` | CI/CD, observability, operations, releases |

### Layer 2 — Meta Agents (Post-run)

| Agent | Role |
|--------|-----|
| `sequoia-correlator` | Cross-references findings across phases. Detects root causes that generate problems in multiple agents. |
| `sequoia-reporter` | Generates the master document and phase markdowns with uniform template. Calculates the Health Score by phase and globally. |

---

## Commands

### Main Commands

```bash
/sequoia init
```
Initializes Sequoia in the current project. Runs `sequoia-context`, builds the project map, detects which agents are relevant (not all apply to all projects), and persists the context in Engram.

---

```bash
/sequoia audit
```
Full audit. Runs the 6 phase agents in parallel (where there are no dependencies) and then the 2 meta-agents. Generates all deliverables.

**Flags**:
- `--phase=security|performance|architecture|quality|experience|operations` — Audits only one phase
- `--scope=changed` — Only audits files modified since last commit/PR
- `--scope=module=<path>` — Audits a specific module
- `--mode=full|quick` — Full: deep analysis. Quick: blocking and critical findings only
- `--output=report|tasks|both` — Type of deliverable to generate

---

```bash
/sequoia review
```
Code review mode. Designed for PR review or specific diff review. Runs a subset of relevant agents based on changed files. Faster than audit, deeper than a linter.

**Flags**:
- `--diff=HEAD~1..HEAD` — Commit range to review
- `--pr=<number>` — Review a specific PR (requires gh CLI)
- `--strict` — No tolerance for medium findings

---

```bash
/sequoia score
```
Generates the project Health Scorecard. Requires at least one prior audit. Shows evolution if history exists.

---

```bash
/sequoia report
```
Regenerates documents from cached findings. Does not re-run agents.

---

```bash
/sequoia fix <phase> [--task=<id>]
```
Generates an actionable task plan from a phase's findings, optimized for another agent to implement. Includes: minimum necessary context, candidate files, acceptance criteria, risk.

---

```bash
/sequoia diff
```
Compares current project state against the last recorded audit. Shows what improved, what worsened, what's new.

---

## Agents in Detail

---

### C1 · sequoia-context (Pre-flight)

**Purpose**: Build the project map that will inform all other agents.

**Inputs**:
- Directory structure
- Configuration files (any detected format)
- Dependency manifests
- README and internal documentation
- CI/CD workflows present

**Outputs: Project Map**

```markdown
## Project Map
- Stack: [detected]
- Runtime: [detected]
- Paradigm: [SPA / SSR / API REST / API GraphQL / CLI / Library / Fullstack / ...]
- Bundler/Build: [detected or absent]
- Test infra: [present / partial / absent]
- CI/CD: [present / partial / absent]
- Estimated maturity: [prototype / active development / production]
- Main modules: [list]
- Risk dependencies: [preliminary list]
- Applicable agents: [list of P1-P6 that apply to the context]
- Non-applicable agents: [list with reason — e.g. P5 does not apply, no user interface]
```

**Key improvement over the original**: The original prompt assumes frontend (mentions package.json, vite, hooks, stores). Sequoia-context removes that assumption and makes each agent adapt its questions to the real stack.

---

### P1 · sequoia-security

**Domain**: Security, authentication, authorization, attack surface, secrets handling.

**Context adaptation**:
- Frontend SPA: focus on tokens, XSS, CSRF, insecure storage, redirects
- API/Backend: focus on endpoint authentication, injection, rate limiting, RBAC
- Fullstack: both
- CLI tool: focus on credential handling, system permissions

**Specific inspections**:
- Token handling: where stored, how expired, how rotated
- PII persistence: what data is logged, cached, or exposed in errors
- Real vs cosmetic logout (UI only vs real session/token invalidation)
- Redirects: externally manipulable?
- XSS surface: unsanitized interpolation, dangerouslySetInnerHTML equivalents
- Secrets in code: hardcoded keys, tokens, passwords in tracked files
- Security headers if there's a custom server
- Front/back contracts needed for real hardening
- CORS, CSP, cookies (httpOnly, Secure, SameSite)
- Dependencies with known CVEs

**Additional deliverable**: Attack surface matrix — table of input vectors × current mitigation state.

---

### P2 · sequoia-performance

**Domain**: Performance, load times, render, memory, bundles, assets.

**Context adaptation**:
- Frontend: Core Web Vitals, bundle splitting, lazy loading, unnecessary renders
- Backend/API: endpoint latency, N+1 queries, cache, concurrency
- Fullstack: both
- CLI: startup time, memory usage

**Specific inspections**:
- Dependency weight: what enters the bundle that shouldn't?
- Unnecessary or full imports when only one function is used
- Eager vs lazy: what loads at startup that could be deferred?
- Oversized assets: images, fonts, static JSON
- Avoidable render work: computations in the render path that could be cached
- Perceived vs real loads: skeletons, optimistic UI, streaming
- Expensive queries without indexes
- Blocking operations on the critical path

**Additional deliverable**: Performance budget — table with measurable target metrics and verification method:

| Metric | Target | How to measure | Current state |
|---------|----------|------------|---------------|
| LCP | < 2.5s | Lighthouse / WebVitals | [verified] |
| TTI | < 3.5s | Lighthouse | [verified] |
| Bundle JS | < 300kb gzip | build output | [verified] |

---

### P3 · sequoia-architecture

**Domain**: System design, scalability, patterns, module boundaries, structural debt.

**Context adaptation**:
- SPA: stores, routing, state management, service layer
- Backend: domain layers, repositories, services, controllers
- Fullstack: integration between layers, internal contracts
- Library: public API, encapsulation, composability

**Specific inspections**:
- Functional duplication: same logic in multiple places
- Module boundaries: does each module have clear responsibility?
- Unnecessary coupling: who knows too much about whom?
- Missing runtime validation: blind trust in static types at runtime?
- Inconsistent conventions: is the same problem solved 3 different ways?
- Unnecessary public surface: internal APIs exposed without need
- Dangerous scaling points: what breaks first if the project grows 10x?
- God objects/modules: components or modules that know too much
- Stack-specific antipatterns detected

**Additional deliverable**: Module dependency map — textual diagram of who depends on whom, with identified coupling arrows.

---

### P4 · sequoia-quality

**Domain**: Testing, coverage, code quality, measurable technical debt.

**Context adaptation**:
- Detects the testing framework present (or its absence)
- Adapts recommendations based on project type and maturity

**Specific inspections**:
- Real test status: how many exist? do they pass? what do they cover?
- Missing infrastructure: test runner, fixtures, mocks, factories?
- Highest-risk modules without coverage (identified together with P3)
- Quality of existing tests: do they test behavior or implementation?
- Contract tests: is there validation that the API contract is fulfilled?
- Minimum smoke tests: is there something that verifies the project starts?
- Code debt: high cyclomatic complexity, long functions, duplication
- Lint/format: configured? runs in CI? has serious rules?

**Mandatory incremental strategy**: Don't propose "reach 80% coverage." Propose the minimum viable path: first smoke, then critical modules, then integration.

---

### P5 · sequoia-experience

**Domain**: User experience, accessibility, flows, responsive, onboarding, interactions.

**Context adaptation**:
- Only applies to projects with user interface (web, mobile, desktop, interactive CLI)
- For pure APIs: applies to developer experience (DX) as "API UX"

**Specific inspections**:
- Flow blocking: steps where the user can get stuck without an exit
- Errors without recovery message or actionable next step
- Poor loading states: eternal spinners, absence of feedback
- Real accessibility: ARIA roles, contrast, keyboard navigation, screen reader
- Responsive/mobile: does the layout break in real viewports?
- Onboarding friction: how many steps until the first perceived value?
- Modals, tables, forms, menus, tabs: does each have error, empty, and loading states?
- Forms: real-time validation? clear messages? error recovery?

**New inspection**: Developer Experience (DX) for library or API type projects — how hard is it to start using this code?

---

### P6 · sequoia-operations

**Domain**: CI/CD, operations, releases, monitoring, observability, repo guardrails.

**Context adaptation**:
- For small/personal projects: focus on minimum viable (lint in CI, env contract, basic monitoring)
- For production: focus on staging, previews, rollback, alerts, postmortem capability

**Specific inspections**:
- Missing scripts: start, build, test, lint, typecheck — do they all exist and work?
- Env contract: is there `.env.example` or similar? are all required variables documented?
- Tracked `.env`: are there environment files with real secrets in the repo?
- Git hooks: pre-commit, pre-push — do they exist? do they have serious guardrails?
- CI/CD: does it exist? does it run tests? does it block merge if it fails?
- Staging/previews: is there a validation environment before production?
- Basic monitoring: is there uptime monitoring? is it known when it goes down?
- Logger: is there structured logging? are errors observable?
- Observability: is there error tracking (Sentry equivalent)?
- Operational docs: is there a minimum runbook? how to deploy? how to rollback?
- Release policy: semver? changelog? tags?
- Dependency health: regularly updated? CVE scanning?

**Additional deliverable**: Minimum ownership — table of who is responsible for what in operations (even if it's one person).

---

### M1 · sequoia-correlator (Meta)

**Purpose**: Find cross-cutting root causes. Many symptoms in different phases have the same cause.

**Real example**:
- P3 detects: "no runtime validation"
- P1 detects: "user data is not sanitized before use"
- P4 detects: "no tests for data input modules"
- **Correlation**: the root cause is a single one — absence of an input validation layer

**Output**: List of root causes with their symptoms in multiple phases, prioritized by aggregate impact.

---

### M2 · sequoia-reporter (Meta)

**Purpose**: Generate final deliverables with uniform template. Calculate project health metrics by phase and globally.

**Deliverables**:
- `sequoia-master.md` — Master document with executive summary, severities by phase, roadmap
- `sequoia-phases/01-security.md` ... `06-operations.md` — One markdown per phase
- `sequoia-score.md` — Health scorecard

**Health Score by phase**:

```
🔴 CRITICAL   — Production or security blockers
🟠 RISK       — Serious problems without active solution
🟡 ATTENTION  — Prioritizable technical debt
🟢 HEALTHY    — Low level of serious findings
⚪ N/A         — Phase does not apply to the project
```

**Global Health Score**: weighted average (security and operations carry more weight).

---

## Non-Negotiable Rules (inherited and extended)

These rules apply to ALL agents without exception:

1. **Do not assume**. If something is not in the repo, declare it as absent, not invent it.
2. **Do not confirm claims without verifying**. Neither the user's nor internal documentation's.
3. **If not verifiable, say so explicitly** with the `[NOT VERIFIABLE]` label.
4. **Cite real files**. If you mention a file, it must exist with that path.
5. **No generic theory**. Repo evidence or silence.
6. **No decorative checklists**. Each item must be verifiable and actionable.
7. **No magic solutions**. Each recommendation includes impact, dependencies, and risk.
8. **If repo and prior documentation contradict, the repo prevails**.
9. **No destructive changes unless explicitly instructed**.
10. **Adapt to the project's real context**. Don't apply enterprise criteria to a prototype.

**New rule 11**: If an agent cannot verify something because it requires external access (infra, production DB, Sentry logs), it declares it as `[REQUIRES EXTERNAL ACCESS]` and describes what to verify and how.

**New rule 12**: If a recommendation only applies if the project grows (future scale), mark it with `[ONLY IF SCALING]`. Don't mix it with recommendations for the current state.

---

## Standard Finding Format

All agents use exactly this structure:

```markdown
### [PHASE-ID] · [Finding title]  [🔴 CRITICAL | 🟠 RISK | 🟡 ATTENTION]

**Status**: Confirmed | Partial | Not verifiable | Outdated

**Evidence**:
- `path/to/real/file.ext:line` — description of what was observed
- Detected behavior or absence

**Problem**:
What is wrong and why it matters technically. No generalities.

**Real Impact**:
What can happen in production if this continues.

**Minimum High-Leverage Recommendation**:
What concrete change to make first and why specifically that one.

**Dependencies / Blockers**:
Backend, infra, API contract, other modules, external team, etc.

**Implementation Risk**: Low | Medium | High
Reason for the estimated risk.

**Acceptance Criteria**:
How to verify that the finding was resolved.
```

---

## Mandatory Phase Template

Each phase document generated by `sequoia-reporter` uses exactly this structure:

```markdown
# Phase [N] — [Name]

**Agent**: sequoia-[name]
**Project**: [name]
**Audit date**: [date]
**Detected stack**: [from project map]

## 1. Phase objective

## 2. Inspection scope
What files, directories and configurations were reviewed.
What was left out of scope and why.

## 3. Verified current state
### What internal documentation says (if it exists)
### What was confirmed in code
### What was outdated, ambiguous, or not verifiable

## 4. Consolidated findings
Ordered by severity. No duplicates. No filler.

## 5. High-leverage missing items
Only technically justified improvements. With expected impact.

## 6. Task plan
Each task with: context, files, impact, dependencies, risk, acceptance criteria, priority.

## 7. Recommended implementation order
Sequence that minimizes risk and maximizes impact.

## 8. Phase risks and blockers

## 9. Phase closure checklist
Verifiable list of what must be true when this phase is "done."
```

---

## Final Deliverables

### Master Document: `sequoia-master.md`

```markdown
# Sequoia Audit Report — [Project]

**Date**: [date]
**Mode**: Full | Quick | Phase | Review
**Stack**: [detected]
**Estimated maturity**: [from context map]

## Executive Summary
Overall project state in 5-10 lines. No exaggeration.

## Health Scorecard
| Phase | Agent | Score | Blocking | High Leverage | Backlog |
|------|--------|-------|-------------|---------------|---------|
| Security | P1 | 🟠 | 2 | 3 | 1 |
| ... | | | | | |

## Top 10 Global Findings
The most critical, regardless of phase. Prioritized by impact × urgency.

## Cross-Cutting Root Causes
Output from sequoia-correlator.

## Suggested Roadmap
Ordered by: blocks production → high leverage → backlog → acceptable.

## Non-Applicable Phases
List with reason (from context map).
```

---

## Workflows

### Flow 1: New Full Audit

```
/sequoia init
  └─► sequoia-context → Project Map

/sequoia audit
  ├─► P1-P6 in parallel (where there are no dependencies)
  │     Each agent receives: Project Map + domain instruction
  ├─► M1 (sequoia-correlator) — after all agents
  └─► M2 (sequoia-reporter) — generates all documents + scores
```

### Flow 2: PR / Diff Review

```
/sequoia review --diff=HEAD~1..HEAD
  └─► sequoia-context (changed files only)
        └─► Automatic selection of relevant agents
              (based on what types of files changed)
        └─► Findings only on the diff
        └─► Flag if the change touches areas with open prior findings
```

### Flow 3: Incremental Audit (Re-run)

```
/sequoia diff
  └─► Compares current state vs last audit in Engram
        └─► Shows: resolved / new / worsened / unchanged
```

### Flow 4: Generate Tasks for Implementing Agent

```
/sequoia fix security
  └─► sequoia-reporter generates task plan from P1
        Format: each task self-sufficient, with minimum context for another agent
        No need to re-read the entire audit
```

---

## Configuration: `sequoia.config.json`

Optional file at the project root:

```json
{
  "project": "project-name",
  "maturity": "prototype | development | production",
  "agents": {
    "disabled": ["P5"],
    "focus": ["P1", "P6"]
  },
  "thresholds": {
    "security": "strict",
    "performance": "standard",
    "quality": "relaxed"
  },
  "outputs": {
    "dir": "docs/sequoia",
    "master": true,
    "phases": true,
    "scorecard": true
  },
  "context": {
    "stack": "auto",
    "entryPoints": ["src/main.ts", "app/index.tsx"],
    "excludeDirs": ["node_modules", "dist", ".git"]
  }
}
```

**Threshold levels by phase**:
- `strict` — zero tolerance for medium findings
- `standard` — only reports medium+ (default)
- `relaxed` — only reports critical and high risks

---

## Integration with Claude Code

### Suggested Hooks

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [{
          "type": "command",
          "command": "echo 'Sequoia: write operation detected'"
        }]
      }
    ]
  }
}
```

### Slash Commands (Skills)

| Command | Skill | Description |
|---------|-------|-------------|
| `/sequoia` | `sequoia:orchestrator` | Main entry point |
| `/sequoia-security` | `sequoia:security` | Security audit |
| `/sequoia-review` | `sequoia:review` | PR/diff review |
| `/sequoia-score` | `sequoia:scorecard` | Health scorecard |
| `/sequoia-fix` | `sequoia:fix` | Generates task plan |

### Integration with Engram (Persistent Memory)

Sequoia persists in Engram:
- The Project Map for each project
- The findings of each audit with timestamp
- The Health Score history to see evolution
- The generated tasks and their status

This enables:
- `sequoia diff`: compare current state vs previous audit
- The implementing agent remembers findings from previous sessions
- Project health evolution over time

---

## Improvements Over the Original Prompt

| Aspect | Original | Sequoia |
|---------|----------|---------|
| **Stack** | Assumes frontend (package.json, vite, hooks) | Auto-detect: any stack |
| **Phases** | 7 phases | 6 phase agents + 2 meta |
| **Review mode** | Full audit only | Audit + Review (PR/diff) + Incremental |
| **Correlation** | Does not exist | sequoia-correlator: cross-cutting root causes |
| **Scoring** | Does not exist | Health Scorecard by phase and global |
| **Persistence** | Does not exist | Engram: history, diff, evolution |
| **Configuration** | Does not exist | sequoia.config.json: thresholds, agents, outputs |
| **API DX** | Does not exist | P3 includes DX for API consumers |
| **PII/Data** | Mentioned in security | Dedicated data integrity checks in P6 |
| **Deps** | Mentioned in DevOps | Dedicated CVE, license, risk score checks |
| **Tasks for agent** | Static format | /sequoia fix: output optimized for implementer |
| **Project maturity** | Ignores context | Adapts criteria based on maturity (prototype vs production) |
| **Future findings** | Mixed with current | `[ONLY IF SCALING]` separates them explicitly |
| **External access** | Not considered | `[REQUIRES EXTERNAL ACCESS]` declares verifiable limits |

---

## Closing Principle

Sequoia is not a checklist. It is a team of agents that reason, correlate, and prioritize.

A project audited by Sequoia should end up with:
1. A map of real state — not aspirational
2. A list of findings traceable to real evidence
3. A roadmap any architect would sign off on
4. Tasks any implementing agent can execute without ambiguity
5. A score that evolves with the project

The goal is not to find problems. The goal is to leave the project understood, prioritized, and ready to intervene.
