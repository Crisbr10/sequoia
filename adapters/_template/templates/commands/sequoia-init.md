---
description: Initialize Sequoia in the project. Detects stack, builds the
  Project Map, identifies applicable agents, and persists context. This is
  the required first step before any audit.
allowed-tools: Read, Glob, Grep, Bash
---

# /sequoia init

Initialize Sequoia in the current project. Builds the complete Project Map
that informs ALL subsequent agents.

## What It Does

1. Runs the context agent (`sequoia-context`) as a pre-flight
2. Builds the **Project Map** — the single source of truth about the project
3. Determines which agents apply and which don't (with explicit reasoning)
4. Persists the map so future sessions can retrieve it

## Workflow

### Step 1 — Scan project structure

Search for:
- Dependency manifests: package.json, go.mod, Cargo.toml, pom.xml,
  requirements.txt, pyproject.toml, Gemfile, *.csproj, composer.json
- Configurations: tsconfig.*, vite.config.*, next.config.*, Dockerfile,
  .env*, *.yaml, *.toml
- Frameworks: next/, pages/, src/app/, angular.json, nuxt.config.*
- CI/CD: .github/workflows/, .gitlab-ci.yml, Jenkinsfile
- Documentation: README*, CHANGELOG*, docs/, CONTRIBUTING*

### Step 2 — Analyze tech stack

Identify with evidence:
- Primary language (and secondary if any)
- Framework (React, Next.js, Express, Django, Spring, Gin, etc.)
- Runtime (Node, Deno, Bun, Go, Python, Java, .NET)
- Bundler/Build (Vite, Webpack, esbuild, Rollup, or none)
- Test runner (Jest, Vitest, pytest, Go testing, or absent)
- Package manager (npm, pnpm, yarn, pip, cargo, go modules)

### Step 3 — Determine project paradigm

Classify into ONE primary (may have secondary):
SPA | SSR | SSG | API REST | API GraphQL | CLI | Library | Monolith |
Microservices | Fullstack | Mobile | Desktop | Plugin

Justify classification with evidence from the repo.

### Step 4 — Estimate project size

| Metric | Value | How measured |
|--------|-------|-------------|
| Estimated LOC | ~N | file count x average estimate |
| Main modules | N | directories with own responsibility |
| Direct dependencies | N | from dependency manifest |
| Total dependencies | ~N | from lockfile if present |

### Step 5 — Check existing infrastructure

| Aspect | Present | Partial | Absent |
|--------|---------|---------|--------|
| Unit tests | | | |
| Integration tests | | | |
| CI/CD pipeline | | | |
| Linting/Formatting | | | |
| Type checking | | | |
| Technical docs | | | |
| .env.example | | | |
| Containerization | | | |

### Step 6 — Assess project maturity

- **Prototype**: no tests, no CI, minimal README, flat structure
- **Active development**: some tests, basic CI, modular structure
- **Production**: tests >50%, CI with gates, monitoring, stable releases

### Step 7 — Determine applicable agents

For each agent P1-P6, decide applicability with reasoning:

| Agent | Applies | Reason |
|-------|---------|--------|
| P1 Security | ✅ | Every project needs security review |
| P2 Performance | | |
| P3 Architecture | ✅ | Always (includes API design) |
| P4 Quality | ✅ | Always (includes dependencies) |
| P5 Experience | | |
| P6 Operations | ✅ | Always, adjusted by maturity |

### Step 8 — Persist

Save as a persistent observation:
- **title**: "Sequoia Project Map — {project-name}"
- **type**: architecture
- **content**: the complete Project Map in markdown format

If a previous Project Map exists, compare and notify of significant changes.

## Output format: Project Map

```markdown
## Sequoia Project Map — {name}

**Date**: {date}
**Maturity**: {prototype | development | production}
**Paradigm**: {primary type}

### Detected Stack
- Language: {detected}
- Framework: {detected or "none"}
- Runtime: {detected}
- Bundler: {detected or "none"}
- Test runner: {detected or "ABSENT"}
- Package manager: {detected}

### Size
- LOC: ~{N}
- Modules: {list}
- Dependencies: {N} direct, ~{N} total

### Infrastructure
- Tests: {present | partial | absent}
- CI/CD: {present | partial | absent}
- Docs: {present | partial | absent}
- Lint: {present | absent}

### Applicable Agents
{table from step 7}

### Context Notes
- {any relevant detail affecting analysis}
```

## Handling ambiguous detection

When the stack is unclear:
1. List possible options with estimated confidence
2. Search for additional evidence (imports in source, scripts in manifests)
3. If ambiguity persists, declare it explicitly: `[AMBIGUOUS: could be X or Y]`
4. Do NOT guess. Declared ambiguity is better than silent assumption.

If no dependency manifest is detected:
- Declare: "No standard dependency manifest detected"
- List which files were searched for and not found
- Assess whether it's a dependency-free project or missing information

## Precondition

No preconditions. This is ALWAYS the first command.

## Post-condition

The Project Map is persisted. The `audit`, `review`, `diff`, and `fix`
commands consume it automatically.
