# Design: v1.0 Audit Agents — P7 i18n, P4 Deep Deps, P3 Resilience

## Architecture Decisions

### D1: InstallOpts Struct Location
**Decision**: `adapters/interface.go` — alongside `ToolAdapter` interface.
**Rationale**: The struct is part of the adapter contract. Putting it in `adapters/interface.go` keeps the contract self-contained. `internal/model/types.go` is for TUI domain types — InstallOpts is an adapter concern.
**Rejected**: internal/model (wrong layer), adapters/common (less discoverable).

### D2: Interface Compatibility
**Decision**: **Option A (Breaking)** — change signatures to `Install(InstallOpts) error` and `Uninstall(InstallOpts) error`.
**Rationale**: Cleanest approach. The existing code already has the `_ = lang` placeholder anticipating this. All 5 adapters + _template need updating, but the change is mechanical (add parameter, use `_ = opts.Language`). Test mocks update similarly.
**Rejected**: Option B (InstallWithOpts) — would leave dead code. Option C (context.Context) — implicit, non-obvious.

### D3: Template Propagation Strategy
**Decision**: **Sequential with test verification** — edit opencode first (canonical, most complete), then propagate pattern to gemini, claude, cursor, codex, _template.
**Rationale**: opencode's skill.md.tmpl (1931 lines) is the canonical full template. Edit it first, verify it works, then mechanically propagate to others. The shorter templates (cursor, codex, _template) only need roster table + description updates, not full agent sections.
**Verification**: Golden file tests for each adapter verify template rendering. Add P7-specific assertions.

### D4: Test Verification
**Decision**: Extend existing golden file tests with P7 content assertions. Add template-specific assertions.
**Files affected**: `adapters/*/templates_test.go` + golden files in `adapters/*/templates/testdata/golden/`.
**Approach**: 
1. Update golden files (or add assertions that P7 appears in rendered output)
2. Add `templateData` tests verifying P7-related template variables
3. Run `go test -race ./...` after each adapter edit

### D5: Pipeline Runner Changes
**Decision**: Construct `InstallOpts{Language: lang}` in runner.go, pass to adapter.Install() and adapter.Uninstall().
**Rationale**: The `lang` parameter already flows through the runner. Minimal change: remove `_ = lang`, construct InstallOpts, pass to adapter calls.
**TUI impact**: None. The progress channel pattern is unchanged.

## Component Diagram

```
TUIConfig.Language (string "en"/"es")
        │
        ▼
pipeline.Runner.RunInstall(ctx, tools, ch, lang)
        │
        ▼
InstallOpts{Language: lang}
        │
        ▼
adapter.Install(opts InstallOpts) ──► template rendering
adapter.Uninstall(opts InstallOpts)
```

## File Change Map

### NEW Files
| File | Purpose |
|------|---------|
| `docs/agents/sequoia-i18n.md` | P7 agent specification (~150 lines) |

### MODIFIED Files — Go Code
| File | Change | Lines |
|------|--------|-------|
| `adapters/interface.go` | Add `InstallOpts` struct; update `Install()`/`Uninstall()` signatures | +12, ~6 |
| `adapters/claude/adapter.go` | Update `Install()`/`Uninstall()` to accept `InstallOpts` | ~4 |
| `adapters/opencode/adapter.go` | Same | ~4 |
| `adapters/cursor/adapter.go` | Same | ~4 |
| `adapters/gemini/adapter.go` | Same | ~4 |
| `adapters/codex/adapter.go` | Same | ~4 |
| `adapters/_template/adapter.go` | Same | ~4 |
| `internal/pipeline/runner.go` | Remove `_ = lang`, construct `InstallOpts` | ~6 |
| `internal/model/types.go` | Add `LangEN.Locale()` method or keep as-is | ~0-3 |

### MODIFIED Files — Agent Docs
| File | Change | Lines |
|------|--------|-------|
| `docs/agents/sequoia-quality.md` | Add Deep Deps section (CVE, license, SBOM) | +60 |
| `docs/agents/sequoia-architecture.md` | Add Resilience section (circuit breakers, retries, degradation) | +45 |

### MODIFIED Files — Templates
| File | Change | Lines |
|------|--------|-------|
| `adapters/opencode/templates/skill.md.tmpl` | P7 in roster, health_score, weights, correlation, delegation | +10 |
| `adapters/opencode/templates/agents-md-section.md.tmpl` | P7 in agent roster table | +1 |
| `adapters/gemini/templates/skill.md.tmpl` | Same as opencode + P3/P4 expanded sections | +8 |
| `adapters/claude/templates/skill.md.tmpl` | P7 roster + description update | +2 |
| `adapters/cursor/templates/skill.md.tmpl` | P7 roster + description update | +2 |
| `adapters/codex/templates/skill.md.tmpl` | P7 roster + description update | +2 |
| `adapters/_template/templates/skill.md.tmpl` | P7 roster | +1 |

### MODIFIED Files — Docs
| File | Change | Lines |
|------|--------|-------|
| `docs/SEQUOIA.md` | P7 in agent tables + detail section | +15 |
| `docs/SKILL.md` | P7 in agent selection + roster tables | +10 |

## Test Strategy

### Unit Tests (NEW)
- `adapters/interface_test.go` — InstallOpts struct tests
- `docs/agents/sequoia-i18n_test.go` — Agent spec validation

### Modified Tests
- All adapter `adapter_test.go` files — update mock `Install()`/`Uninstall()` signatures
- All adapter `installer_test.go` files — update calls to `Install(InstallOpts{})`
- `internal/pipeline/runner_test.go` — verify InstallOpts construction
- Golden file tests — update golden files or add assertion for P7 content

### Integration Tests
- `internal/app/integration_test.go` — verify language flows through pipeline
- Template rendering tests — verify P7 appears in rendered output

### Verification
```bash
go test -race -count=1 ./...    # All 312+ tests must pass
go vet ./...                     # Clean
golangci-lint run                # Clean
```

## Rollback Plan
1. `git checkout` all modified files to pre-change state
2. Delete `docs/agents/sequoia-i18n.md`
3. Run `go test -race ./...` to verify 312+ tests pass
