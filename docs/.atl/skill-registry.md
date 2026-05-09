# Skill Registry — Sequoia

> Generated: 2026-05-09 | Mode: hybrid (engram + openspec)

## User Skills

### `~/.config/opencode/skills/` (preferred — most up-to-date)

| Skill | Trigger | Path |
|-------|---------|------|
| `sdd-init` | Initialize SDD context, testing capabilities, registry, and persistence | `~/.config/opencode/skills/sdd-init/SKILL.md` |
| `sdd-explore` | Explore SDD ideas before committing to a change | `~/.config/opencode/skills/sdd-explore/SKILL.md` |
| `sdd-propose` | Create an SDD change proposal with intent, scope, and approach | `~/.config/opencode/skills/sdd-propose/SKILL.md` |
| `sdd-spec` | Write SDD delta specs with requirements and scenarios (Given/When/Then) | `~/.config/opencode/skills/sdd-spec/SKILL.md` |
| `sdd-design` | Create technical design document with architecture decisions | `~/.config/opencode/skills/sdd-design/SKILL.md` |
| `sdd-tasks` | Break an SDD change into implementation tasks | `~/.config/opencode/skills/sdd-tasks/SKILL.md` |
| `sdd-apply` | Implement SDD tasks from specs and design (RED-GREEN-REFACTOR when strict_tdd) | `~/.config/opencode/skills/sdd-apply/SKILL.md` |
| `sdd-verify` | Execute tests and prove implementation matches specs, design, and tasks | `~/.config/opencode/skills/sdd-verify/SKILL.md` |
| `sdd-archive` | Sync delta specs to main specs and archive a completed change | `~/.config/opencode/skills/sdd-archive/SKILL.md` |
| `sdd-onboard` | Walk users through the SDD workflow on the real codebase | `~/.config/opencode/skills/sdd-onboard/SKILL.md` |
| `go-testing` | Go tests, teatest, golden files, Bubbletea TUI testing | `~/.config/opencode/skills/go-testing/SKILL.md` |
| `judgment-day` | Blind dual review with fix-rejudge loop | `~/.config/opencode/skills/judgment-day/SKILL.md` |
| `branch-pr` | PR creation workflow with issue-first checks | `~/.config/opencode/skills/branch-pr/SKILL.md` |
| `issue-creation` | GitHub issue creation with issue-first checks | `~/.config/opencode/skills/issue-creation/SKILL.md` |
| `skill-creator` | Create LLM-first skills with valid frontmatter | `~/.config/opencode/skills/skill-creator/SKILL.md` |
| `skill-registry` | Create or update the project skill registry | `~/.config/opencode/skills/skill-registry/SKILL.md` |
| `chained-pr` | Split oversized PRs (>400 lines) into chained/stacked reviews | `~/.config/opencode/skills/chained-pr/SKILL.md` |
| `cognitive-doc-design` | Design docs that reduce cognitive load (guides, READMEs, RFCs, architecture) | `~/.config/opencode/skills/cognitive-doc-design/SKILL.md` |
| `comment-writer` | Write warm, direct collaboration comments (PR feedback, issues, reviews) | `~/.config/opencode/skills/comment-writer/SKILL.md` |
| `work-unit-commits` | Plan commits as reviewable work units | `~/.config/opencode/skills/work-unit-commits/SKILL.md` |

### `~/.claude/skills/` (mirror — same skills, older versions)

| Skill | Trigger |
|-------|---------|
| `go-testing` | Go tests, teatest, golden files |
| `sdd-*` (init, explore, propose, spec, design, tasks, apply, verify, archive) | SDD phases |
| `judgment-day` | Dual-agent code review |
| `branch-pr` | PR creation |
| `issue-creation` | GitHub issue creation |
| `skill-creator` | Create new AI skills |

### `~/.agents/skills/` (domain-specific engineering skills)

| Skill | Trigger |
|-------|---------|
| `api-design-principles` | REST/GraphQL API design, review specs, establish standards |
| `brainstorming` | Creative work — features, components, behavior changes (MUST use before implementation) |
| `diagnose` | Hard bugs, performance regressions (Reproduce → Minimise → Hypothesise → Instrument → Fix) |
| `email-best-practices` | Email features, SPF/DKIM/DMARC, compliance, webhooks, transactional vs marketing |
| `evolution-api` | WhatsApp integration via Evolution API |
| `expo-dev-client` | Expo development client builds |
| `find-skills` | Discover and install agent skills |
| `grill-me` | Interview user relentlessly about a plan — resolve decision tree |
| `grill-with-docs` | Grilling session that updates CONTEXT.md and ADRs inline |
| `improve-codebase-architecture` | Find deepening opportunities, refactoring, module consolidation |
| `next-best-practices` | Next.js file conventions, RSC boundaries, data patterns, async APIs |
| `next-cache-components` | Next.js 16 Cache Components, PPR, use cache, cacheLife, cacheTag |
| `nextjs-supabase-auth` | Supabase Auth with Next.js App Router |
| `python-testing-patterns` | Pytest, fixtures, mocking, TDD |
| `react-doctor` | Review React changes, catch issues early |
| `react-email` | HTML email templates with React components |
| `requesting-code-review` | Verify work meets requirements before merging |
| `resend` | Resend email API (transactional, webhooks, templates) |
| `setup-matt-pocock-skills` | Configure agent skills block for repo context (issue tracker, labels, docs) |
| `supabase-postgres-best-practices` | Postgres performance optimization, queries, schema design |
| `tdd` | Test-driven development with red-green-refactor loop |
| `test-driven-development` | Write tests before implementation code |
| `to-issues` | Break plan/spec/PRD into independently-grabbable issues |
| `to-prd` | Convert conversation context into PRD |
| `triage` | Triage issues through state machine (triage roles) |
| `twilio-communications` | SMS, voice, WhatsApp, 2FA via Twilio |
| `ui-ux-pro-max` | UI/UX design (50+ styles, 161 palettes, 57 fonts, 99 UX guidelines) |
| `vercel-react-best-practices` | React/Next.js performance optimization patterns |
| `vercel-react-native-skills` | React Native/Expo performance, animations, native modules |
| `write-a-skill` | Create new agent skills with proper structure |
| `zoom-out` | Broader context or higher-level perspective on code |

## Project Skills

### `gentle-ai/skills/` (Gentle AI project-specific)

| Skill | Trigger | Path |
|-------|---------|------|
| `gentle-ai-issue-creation` | Creating a GitHub issue, reporting a bug, requesting a feature | `gentle-ai/skills/issue-creation/SKILL.md` |
| `gentle-ai-branch-pr` | Creating a PR, opening a PR, preparing changes for review | `gentle-ai/skills/branch-pr/SKILL.md` |
| `gentle-ai-chained-pr` | PR >400 lines, stacked/chained pull requests | `gentle-ai/skills/chained-pr/SKILL.md` |
| `cognitive-doc-design` | Writing docs that must reduce cognitive load | `gentle-ai/skills/cognitive-doc-design/SKILL.md` |
| `comment-writer` | Drafting human comments, PR feedback, issue replies | `gentle-ai/skills/comment-writer/SKILL.md` |
| `work-unit-commits` | Splitting implementation work into deliverable commits | `gentle-ai/skills/work-unit-commits/SKILL.md` |

## Compact Rules

### go-testing (applies to both sequoia-ai and gentle-ai)
- Always use table-driven tests (`[]struct{ name, input, want }`)
- Use `t.Run(tc.name, ...)` for subtests
- Golden files go in `testdata/golden/` — update with `-update` flag
- Bubbletea TUI: use `teatest.NewTestModel()`, send keys via `tm.Send()`
- Prefer `require.*` (fatal) over `assert.*` (continue) for setup failures
- Integration tests: use temp dirs (`t.TempDir()`), never mutate real `~/.claude/`
- Test file naming: `foo_test.go` in same package for white-box, `foo_integration_test.go` for black-box

### SDD Phases (sdd-apply context)
- Strict TDD Mode: RED → GREEN → REFACTOR cycle mandatory
- Write failing test first, then implement
- No implementation without a failing test that proves the behavior

### chained-pr
- Maximum 400 lines per PR
- Each PR must be independently reviewable
- Stack order documented in PR descriptions
- Base each stacked PR on the previous one

### branch-pr
- Issue must exist before PR (issue-first)
- Conventional Commits format
- No AI attribution in commits

## Project Conventions (sequoia-ai)

- **Language**: Go 1.24+, module `sequoia-ai`
- **Adapters**: All implement `ToolAdapter` interface; registered via `init()` (database/sql pattern)
- **Installer**: Prepare → Apply → Verify → Rollback (never skip rollback on error)
- **Paths**: Always `filepath.Join()` — never string concatenation (cross-platform)
- **TUI scope**: Install/config/status ONLY — no audit features in TUI (hard rule)
- **Progress**: Buffered channel (64), `tea.Cmd` goroutine, never blocks UI
- **Comments**: Godoc on all exported symbols; inline comments only for non-obvious WHY
- **Commits**: Conventional Commits format, no AI attribution
- **Architecture**: Hexagonal/Ports & Adapters, screen enum switch for TUI dispatch

## Agent Convention Files

| File | Role |
|------|------|
| `sequoia/.golangci.yaml` | Go linting configuration (golangci-lint v2) |
| `gentle-ai/AGENTS.md` | Agent skills index for Gentle AI project |
| `gentle-ai/openspec/config.yaml` | Gentle AI SDD configuration |
| `sequoia/openspec/config.yaml` | Sequoia SDD configuration (hybrid mode, strict_tdd) |
