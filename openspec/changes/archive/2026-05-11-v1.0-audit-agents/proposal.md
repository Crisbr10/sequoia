# Proposal: v1.0 Audit Agents — P7 i18n, P4 Deep Deps, P3 Resilience

## Intent

Sequoia v0.1.0 has 6 phase agents (P1-P6) + 2 meta agents (M1-M2). Roadmap gap analysis identified 3 missing audit domains for v1.0: **internationalization** (no agent audits i18n), **deep dependency scanning** (CVE/license/SBOM not covered by P4), and **resilience patterns** (circuit breakers, retries, timeouts not audited by P3). All three are orthogonal additions — no existing agent behavior changes.

## Scope

### In Scope
- **P7 i18n**: New `docs/agents/sequoia-i18n.md` — hardcoded string detection, RTL support, locale formatting, translation extraction
- **P4 Deep Deps**: CVE scanning methodology, license compliance verification, SBOM generation workflow in existing `docs/agents/sequoia-quality.md`
- **P3 Resilience**: Circuit breaker detection, retry/timeout patterns, graceful degradation in existing `docs/agents/sequoia-architecture.md`
- **Template wiring**: Add P7 to all 5 adapter templates (opencode, gemini, cursor, codex, _template), roster tables, and Phase 2 selection matrices
- **Go wiring**: Wire `internal/model/types.go` Language type and `internal/pipeline/runner.go` lang parameter to i18n usage
- **Docs**: Update `docs/SEQUOIA.md` and `docs/SKILL.md` agent tables

### Out of Scope
- Auto-fix or auto-generation of translations
- SBOM generation Go logic (agent instruction only)
- Runtime i18n middleware implementation
- P7 execution engine changes (reuses existing delegation pattern)

## Capabilities

### New Capabilities
- `agent-p7-i18n`: I18n audit domain — hardcoded string detection, locale format verification, RTL support, translation key consistency

### Modified Capabilities
- `agent-p4-quality`: Gains deep dependency scan sub-domain (CVE, license compliance, SBOM generation)
- `agent-p3-architecture`: Gains resilience pattern audit sub-domain (circuit breakers, retries, timeouts, graceful degradation)

## Approach

Follow existing agent spec pattern: YAML frontmatter → misión → methodology decision trees → inspection checklists → calibración. P7 is green-field. P4/P3 are orthogonal sections appended before Calibración de Libertad within existing files. Templates re-use `{{.Version}}` data. Go wiring: pass `lang` through adapter.Install()/Uninstall(), replacing `_ = lang` placeholder. Add `Language` field to adapter interface, update 5 adapter implementations + tests.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `docs/agents/sequoia-i18n.md` | **New** | Full P7 agent specification |
| `docs/agents/sequoia-quality.md` | Modified | Deep Deps methodology section |
| `docs/agents/sequoia-architecture.md` | Modified | Resilience patterns section |
| `adapters/*/templates/skill.md.tmpl` (5) | Modified | P7 sections, P3 resilience, P4 deep-deps |
| `adapters/opencode/templates/agents-md-section.md.tmpl` | Modified | P7 roster entry |
| `docs/SEQUOIA.md`, `docs/SKILL.md` | Modified | P7 in agent tables |
| `internal/model/types.go` | Modified | Wire Language to i18n |
| `internal/pipeline/runner.go` | Modified | Pass lang through pipeline |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| P7 findings overlap with P4 (string quality) | Low | Clear domain boundaries: P7=i18n context (locale/RTL), P4=code quality |
| P3 resilience contradicts existing P3 findings | Low | Sub-domain expansion, continuous decision tree |
| Template drift across 5 adapters | Medium | Use _template as canonical, propagate consistently |
| Go interface change breaks adapters | Medium | Add `Language` field to InstallOpts; update all 5 adapters + tests |
| Test regression | Low | Additive change. `go test -race ./...` after each commit |

## Rollback Plan

1. Revert `docs/agents/sequoia-i18n.md` (delete if created)
2. `git checkout` individual files to pre-change state
3. Run `go test -race ./...` to verify 312+ tests pass
4. Restore engram artifacts to prior snapshot

## Dependencies

- Go 1.24.2, testify, bubbletea, teatest — no new external deps
- 312 existing tests must remain green
- Adapter interface `Install()` currently takes no args — needs `InstallOpts` struct with `Language`

## Success Criteria

- [ ] P7 i18n agent spec passes peer review for completeness
- [ ] P7 executes without modifying P1-P6 agent outputs
- [ ] P7 findings don't duplicate existing agent findings (unique i18n domain)
- [ ] P3 resilience doesn't contradict existing P3 architecture findings
- [ ] P4 deep deps doesn't contradict existing P4 quality findings
- [ ] All 312+ tests pass: `go test -race -count=1 ./...`
- [ ] Updated Project Map reflects new agent (P7) and sub-domains
- [ ] All 5 adapter templates synchronize correctly

## Effort Estimate

| Change | Effort |
|--------|--------|
| P7 i18n agent spec | 4-6h |
| P4 deep deps expansion | 3-4h |
| P3 resilience expansion | 2-3h |
| Template wiring (5 adapters) | 2-3h |
| Go code wiring + test updates | 2-3h |
| **Total** | **13-19h** |
