# Tasks: v1.0 Audit Agents — P7 i18n, P4 Deep Deps, P3 Resilience

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 450–500 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (Agent Specs) → PR 2 (Go Wiring) → PR 3 (Templates) |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Agent docs only (no Go, no templates) | PR 1 | ~255 lines, standalone mergeable |
| 2 | InstallOpts + interface + pipeline | PR 2 | ~62 lines Go + tests, depends on nothing from PR 1 |
| 3 | Template propagation + doc wiring | PR 3 | ~125 lines, depends on PR 1 content + PR 2 interface |

---

## Milestone A: Agent Specs (no Go changes)

### TASK-A1 · Create P7 i18n agent document
**Severity**: high | **Category**: documentation | **Effort**: 2h | **Module**: docs/agents

**Why**: No agent audits i18n. P7 fills the gap with hardcoded string detection, RTL support checks, locale format verification, and translation key cross-referencing. Follows canonical pattern: YAML frontmatter → misión → methodology decision trees → checklists → calibración.

**Files**

| Acción | Archivo |
|--------|---------|
| ✨ NEW | `docs/agents/sequoia-i18n.md` |

**Acceptance Criteria**
- [ ] YAML frontmatter with `name: sequoia-i18n`, description, tools
- [ ] Misión section defining i18n audit scope
- [ ] Decision trees: hardcoded string detection (R2), locale formatting (R3), RTL support (R4)
- [ ] Checklists: translation key consistency (R5)
- [ ] Calibración de Libertad section

**Dependencies**: None

---

### TASK-A2 · Expand P4 quality agent with deep deps
**Severity**: medium | **Category**: documentation | **Effort**: 1h | **Module**: docs/agents

**Why**: P4 lacks CVE scanning methodology, license compliance across transitive trees, and SBOM generation workflow. Append Deep Dependencies section before Calibración in `sequoia-quality.md`. Covers CVE severity triage (R1), license scan (R2), SBOM methodology (R3).

**Files**

| Acción | Archivo |
|--------|---------|
| ✏️ EDIT | `docs/agents/sequoia-quality.md` (+60 lines before Calibración) |

**Acceptance Criteria**
- [ ] CVE multi-source advisory lookup section with severity scoping
- [ ] License compliance decision tree detecting copyleft in transitive deps
- [ ] CycloneDX/SPDX SBOM workflow documented (agent instruction, no Go code)
- [ ] Existing P4 content unmodified; section placed before Calibración

**Dependencies**: None

---

### TASK-A3 · Expand P3 architecture agent with resilience patterns
**Severity**: medium | **Category**: documentation | **Effort**: 45min | **Module**: docs/agents

**Why**: P3 audits architecture but not resilience patterns. Append before Calibración: circuit breaker detection at service boundaries (R1), retry/timeout audit with backoff/jitter (R2), graceful degradation and fallback assessment (R3).

**Files**

| Acción | Archivo |
|--------|---------|
| ✏️ EDIT | `docs/agents/sequoia-architecture.md` (+45 lines before Calibración) |

**Acceptance Criteria**
- [ ] Circuit breaker pattern detection decision tree (cascade risk)
- [ ] Retry+timeout audit checklist (unbounded, missing backoff/jitter)
- [ ] Graceful degradation section (fallbacks, cached responses, degraded modes)
- [ ] Section placed before Calibración; existing content untouched

**Dependencies**: None

---

## Milestone B: Go Wiring (compilable standalone)

### TASK-B1 · Add InstallOpts struct + update interface signatures
**Severity**: critical | **Category**: architecture | **Effort**: 30min | **Module**: adapters

**Why**: `Install()` and `Uninstall()` take no args; `_ = lang` in runner.go is a placeholder. Add `InstallOpts{Language string}` in `adapters/interface.go`, change signatures to `Install(InstallOpts) error` and `Uninstall(InstallOpts) error`. Breaking change — all adapters must update.

**Files**

| Acción | Archivo |
|--------|---------|
| ✏️ EDIT | `adapters/interface.go` (+InstallOpts struct, ~12 lines) |

**Acceptance Criteria**
- [ ] `InstallOpts` struct defined alongside `ToolAdapter` in interface.go
- [ ] `Install(InstallOpts) error` replaces `Install() error` in interface
- [ ] `Uninstall(InstallOpts) error` replaces `Uninstall() error` in interface
- [ ] Project does NOT compile yet (breaking change — fixed in B2)

**Dependencies**: None (start of breaking change chain)

---

### TASK-B2 · Update 7 adapter implementations + mock
**Severity**: high | **Category**: quality | **Effort**: 1h | **Module**: adapters

**Why**: Interface change in B1 breaks all 6 adapters + _template + test mock. Each adapter's `Install()` and `Uninstall()` must accept `InstallOpts`, use `_ = opts.Language` to satisfy unused-param linting.

**Files**

| Acción | Archivo |
|--------|---------|
| ✏️ EDIT | `adapters/claude/adapter.go` (Install+Uninstall sigs) |
| ✏️ EDIT | `adapters/opencode/adapter.go` |
| ✏️ EDIT | `adapters/cursor/adapter.go` |
| ✏️ EDIT | `adapters/gemini/adapter.go` |
| ✏️ EDIT | `adapters/codex/adapter.go` |
| ✏️ EDIT | `adapters/_template/adapter.go` |
| ✏️ EDIT | `adapters/registry_test.go` (mockAdapter sigs) |
| ✏️ EDIT | `internal/pipeline/runner_test.go` (testAdapter sigs) |

**Acceptance Criteria**
- [ ] All 7 adapter.go files compile with new Install/Uninstall signatures
- [ ] `_ = opts.Language` present in each adapter body
- [ ] mockAdapter and testAdapter updated
- [ ] `go vet ./...` clean

**Dependencies**: depends on TASK-B1 (interface change)

---

### TASK-B3 · Wire pipeline runner with InstallOpts
**Severity**: high | **Category**: architecture | **Effort**: 30min | **Module**: internal/pipeline

**Why**: Runner constructs `InstallOpts{Language: lang}` from existing lang param, removes `_ = lang` placeholder. Passes to `adapter.Install(opts)` and `adapter.Uninstall(opts)` in runInstallSteps/runUninstallSteps.

**Files**

| Acción | Archivo |
|--------|---------|
| ✏️ EDIT | `internal/pipeline/runner.go` (construct opts, ~6 lines) |

**Acceptance Criteria**
- [ ] `_ = lang` removed from RunInstall and RunUninstall
- [ ] `opts := adapters.InstallOpts{Language: lang}` before adapter calls
- [ ] `adapter.Install(opts)` and `adapter.Uninstall(opts)` compile
- [ ] `go test -race -count=1 ./...` — all 312+ tests pass

**Dependencies**: depends on B1+B2 (interface + adapters must compile first)

---

## Milestone C: Template Propagation (depends on A+B)

### TASK-C1 · Wire opencode canonical template
**Severity**: medium | **Category**: integration | **Effort**: 45min | **Module**: adapters/opencode/templates

**Why**: opencode's `skill.md.tmpl` (1931 lines) is canonical. Add: P7 to Agent Roster table (line ~96), Phase 2 selection matrix, health_score `i18n` category, weights in M2 scoring, P7 delegation section before M1, P7 full agent spec block after P6.

**Files**

| Acción | Archivo |
|--------|---------|
| ✏️ EDIT | `adapters/opencode/templates/skill.md.tmpl` (+12 lines) |
| ✏️ EDIT | `adapters/opencode/templates/agents-md-section.md.tmpl` (+1 row) |

**Acceptance Criteria**
- [ ] P7 row in Agent Roster table: `| P7 i18n | I18n, locale, RTL, translation keys | All /sequoia commands |`
- [ ] Phase 2 matrix includes P7 with trigger `**Siempre**`
- [ ] health_score categories includes `i18n: [0-100]`
- [ ] P7 delegation section rendered in full template
- [ ] agents-md-section.md.tmpl includes P7 in agent roster table

**Dependencies**: depends on A1 (P7 spec content) + B2 (interface stable)

---

### TASK-C2 · Propagate P7 to other 5 templates
**Severity**: medium | **Category**: integration | **Effort**: 30min | **Module**: adapters/*/templates

**Why**: Gemini's template mirrors opencode (full sections). Claude, cursor, codex use shorter templates (roster table + description only). _template gets P7 roster entry for future adapters. Pattern: replicate C1 changes scaled to each template's verbosity.

**Files**

| Acción | Archivo |
|--------|---------|
| ✏️ EDIT | `adapters/gemini/templates/skill.md.tmpl` (full: roster + matrix + section) |
| ✏️ EDIT | `adapters/claude/templates/skill.md.tmpl` (short: roster + description) |
| ✏️ EDIT | `adapters/cursor/templates/skill.md.tmpl` (short: roster + description) |
| ✏️ EDIT | `adapters/codex/templates/skill.md.tmpl` (short: roster + description) |
| ✏️ EDIT | `adapters/_template/templates/skill.md.tmpl` (roster row only) |

**Acceptance Criteria**
- [ ] All templates render without broken Go template syntax
- [ ] P7 appears in Agent Roster in every template
- [ ] Description YAML updated to mention i18n where applicable
- [ ] Short templates (claude/cursor/codex) only have roster + description; no full section

**Dependencies**: depends on C1 (canonical pattern established)

---

### TASK-C3 · Update golden files + template tests
**Severity**: medium | **Category**: testing | **Effort**: 30min | **Module**: adapters/*/templates_test.go

**Why**: Golden file tests verify template rendering output. Adding P7 content changes all 6 template outputs. Regenerate golden files or add P7-specific assertions. Update `go test -update` for each adapter. Verify all tests pass.

**Files**

| Acción | Archivo |
|--------|---------|
| ✏️ EDIT | `adapters/*/templates/testdata/golden/*.golden` (6 files regenerated) |
| ✏️ EDIT | `adapters/*/templates_test.go` (add P7 assertions, if applicable) |

**Acceptance Criteria**
- [ ] `go test -race -count=1 ./...` passes (312+)
- [ ] Golden files reflect P7 content in rendered output
- [ ] `go vet ./...` clean after regeneration

**Dependencies**: depends on C1+C2 (templates must be final before regenerating golden files)

---

### TASK-C4 · Update docs/SEQUOIA.md and docs/SKILL.md
**Severity**: low | **Category**: documentation | **Effort**: 20min | **Module**: docs

**Why**: Both general docs list agents. Add P7 i18n to agent tables and Phase 2 selection matrix. docs/SKILL.md mirrors the template but serves as standalone skill reference.

**Files**

| Acción | Archivo |
|--------|---------|
| ✏️ EDIT | `docs/SEQUOIA.md` (+15 lines: P7 in agent tables + detail section) |
| ✏️ EDIT | `docs/SKILL.md` (+10 lines: P7 in selection matrix + health_score + delegation) |

**Acceptance Criteria**
- [ ] P7 in agent roster/table in both files
- [ ] Phase 2 matrix includes P7
- [ ] Health score categories include i18n in docs/SKILL.md
- [ ] No broken markdown formatting

**Dependencies**: depends on A1 (P7 exists) + C1 (pattern reference)
