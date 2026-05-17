---
description: "Initializes Sequoia in the project. Detects stack, builds the project map, identifies applicable agents, and persists context in Engram. Mandatory first step before any audit."
allowed-tools: Read, Glob, Grep, Bash
---

# /sequoia init

Initializes Sequoia in the current project. Builds the complete map that informs ALL subsequent agents.

## What it does

1. Runs the context agent (`sequoia-context`) as pre-flight
2. Builds the **Project Map** — the single source of truth about the project
3. Determines which agents apply and which don't (with explicit reason)
4. Persists the map in Engram for future sessions to retrieve

## Step-by-step workflow

### Step 1 — Scan project structure

```
Search for:
├── Dependency manifests: package.json, go.mod, Cargo.toml, pom.xml,
│   requirements.txt, pyproject.toml, Gemfile, *.csproj, composer.json
├── Configurations: tsconfig.*, vite.config.*, webpack.config.*, next.config.*,
│   docker-compose.*, Dockerfile, Makefile, build.gradle, .env*, *.yaml, *.toml
├── Frameworks: next/, pages/, src/app/, angular.json, nuxt.config.*
├── CI/CD: .github/workflows/, .gitlab-ci.yml, Jenkinsfile, .circleci/
└── Documentation: README*, CHANGELOG*, docs/, CONTRIBUTING*
```

### Step 2 — Analyze tech stack

Identify with evidence:
- **Primary language** (and secondary if any)
- **Framework** (React, Next.js, Express, Django, Spring, Gin, etc.)
- **Runtime** (Node, Deno, Bun, Go, Python, Java, .NET)
- **Bundler/Build** (Vite, Webpack, esbuild, Rollup, Turbopack, or none)
- **Test runner** (Jest, Vitest, pytest, Go testing, xUnit, or absent)
- **Package manager** (npm, pnpm, yarn, pip, cargo, go modules)

### Step 3 — Determine project paradigm

Classify into ONE primary (may have secondary):
- SPA | SSR | SSG | API REST | API GraphQL | CLI | Library | Monolith
- Microservices | Fullstack | Mobile | Desktop | Plugin

Justify the classification with repo evidence.

### Step 4 — Estimate project size

```markdown
| Metric | Value | How measured |
|---------|-------|---------------|
| Estimated LOC | ~N | file count × estimated average |
| Main modules | N | directories with own responsibility |
| Direct dependencies | N | from dependency manifest |
| Total dependencies | ~N | from lockfile if it exists |
```

### Step 5 — Verify existing infrastructure

| Aspect | Present | Partial | Absent |
|---------|----------|---------|---------|
| Unit tests | | | |
| Integration tests | | | |
| CI/CD pipeline | | | |
| Linting/Formatting | | | |
| Type checking | | | |
| Technical docs | | | |
| `.env.example` / env contract | | | |
| Docker / containerization | | | |

### Step 6 — Assess project maturity

Classification criteria:

- **Prototype**: no tests, no CI, minimal README, minimal deps, flat structure
- **Active development**: some tests, basic CI, modular structure, README with instructions
- **Production**: tests > 50%, CI with gates, operational docs, monitoring, stable releases

### Step 7 — Determine applicable agents

For each agent P1-P6, decide if it applies with reason:

```markdown
| Agent | Applies | Reason |
|--------|--------|--------|
| P1 Security | ✅ | Every project needs security review |
| P2 Performance | ✅ | [justify by type] |
| P3 Architecture | ✅ | Always (includes API design) |
| P4 Quality | ✅ | Always (includes dependencies) |
| P5 Experience | ❌ | Pure API without user interface |
| P6 Operations | ✅ | Always, adjusted by maturity (includes data) |
```

### Step 8 — Persist in Engram

Save as observation with:
- **title**: "Sequoia Project Map — {project-name}"
- **topic_key**: `sequoia/{project-name}/project-map`
- **type**: `architecture`
- **content**: the complete Project Map in markdown format

If a previous Project Map exists, compare and report significant changes.

## Output format: Project Map

```markdown
## Sequoia Project Map — {name}

**Date**: {date}
**Maturity**: {prototype | development | production}
**Paradigm**: {primary type}

### Detected stack
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

### Applicable agents
{table from step 7}

### Context notes
- {any relevant details that affect analysis}
```

## Handling ambiguous detection

When the stack is unclear:
1. List possible options with estimated confidence
2. Search for additional evidence (imports in source files, scripts in manifests)
3. If ambiguity persists, declare it explicitly: `[AMBIGUOUS: could be X or Y]`
4. Do NOT guess. Declared ambiguity is better than silent assumption.

If no dependency manifest is detected:
- Declare: "No standard dependency manifest detected"
- List which files were searched for and not found
- Evaluate whether it's a dependency-free project (loose scripts, bare code) or missing information

## Precondition

Does not require preconditions. This is ALWAYS the first command.

## Post-condition

The Project Map is persisted in Engram. The `audit`, `review`, `diff` and `fix` commands consume it automatically.
