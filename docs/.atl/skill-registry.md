# Skill Registry — sequoia-ai

> Generated: 2026-05-09 | Mode: engram

## User Skills (`~/.claude/skills/`)

| Skill | Trigger |
|-------|---------|
| `go-testing` | Writing Go tests, teatest, golden files, Bubbletea TUI testing |
| `sdd-init` | Initialize SDD context |
| `sdd-explore` | Explore/investigate ideas before committing |
| `sdd-propose` | Create change proposal |
| `sdd-spec` | Write specifications with Given/When/Then |
| `sdd-design` | Technical design document |
| `sdd-tasks` | Break change into implementation task checklist |
| `sdd-apply` | Implement tasks from the change |
| `sdd-verify` | Validate implementation against specs |
| `sdd-archive` | Sync delta specs and archive completed change |
| `judgment-day` | Adversarial dual-agent code review |
| `branch-pr` | PR creation workflow |
| `issue-creation` | GitHub issue creation |
| `skill-creator` | Create new AI skills |
| `web-design-guidelines` | Web/UI design guidelines |

## Compact Rules

### go-testing
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

## Project Conventions (sequoia-ai)

- **Language**: Go 1.22+, module `sequoia-ai`
- **Adapters**: All implement `ToolAdapter` interface
- **Installer**: Prepare → Apply → Verify → Rollback (never skip rollback on error)
- **Paths**: Always `filepath.Join()` — never string concatenation (cross-platform)
- **TUI scope**: Install/config/status ONLY — no audit features in TUI (hard rule)
- **Comments**: Godoc on all exported symbols; inline comments only for non-obvious WHY
- **Commits**: Conventional Commits format, no AI attribution
