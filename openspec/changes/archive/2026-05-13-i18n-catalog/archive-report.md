# Archive Report: i18n-catalog

**Change**: i18n-catalog
**Archived**: 2026-05-13
**Verdict**: PASS ✅ (0 CRITICAL, 3 WARNING, 3 SUGGESTION)
**Delivery**: 3 stacked PRs (feature-branch-chain)
**Artifact Store**: Hybrid (engram + openspec)

---

## Change Summary

Built full i18n infrastructure for the Sequoia TUI across 3 stacked PRs.

**PR 1** (i18n engine foundation): `internal/i18n/` package with `go-i18n/v2`, 65-key en+es TOML catalogs, `T()`/`Init()`/`Initialized()` API, 13 unit tests.

**PR 2** (TUI string migration): All 8 screens migrated to `i18n.T()` calls, language selector re-enabled (gated on `i18n.Initialized()`), `TODO(i18n)` markers removed, 14 golden files regenerated, 2 skipped tests unskipped.

**PR 3** (Adapter template wiring): `RenderTemplateLang()` with language fallback added to `template.go`, `_ = opts.Language` removed from all adapters (base_adapter, codex, gemini, _template), main specs updated.

---

## Artifact Traceability (Engram)

| Artifact | Observation ID | Topic Key |
|----------|---------------|-----------|
| Proposal | #292 | `sdd/i18n-catalog/proposal` |
| Spec (delta) | #293 | `sdd/i18n-catalog/spec` |
| Design | #294 | `sdd/i18n-catalog/design` |
| Tasks | #295 | `sdd/i18n-catalog/tasks` |
| Apply Progress | #296 | `sdd/i18n-catalog/apply-progress` |
| Verify Report | #298 | `sdd/i18n-catalog/verify-report` |
| **Archive Report** | (this save) | `sdd/i18n-catalog/archive-report` |

---

## Specs Synced to Main (openspec)

| Domain | Action | Details |
|--------|--------|---------|
| `i18n-engine` | **Created** | New spec with 3 requirements: Bundle Initialization, T() Accessor, Initialized Guard (9 scenarios) |
| `tui-install-flow` | **Modified** | Updated Configuration Screen requirement (language selector now visible when initialized, 8 updated scenarios); Golden Files requirement reversed (now include language labels); Added Configuration Tests Re-Enabled requirement (1 new scenario) |
| `go-wiring` | Already updated | Modified during apply phase (task 3.4) — Adapters Use Language + RenderTemplateLang requirements (no delta sync needed) |
| `template-wiring` | Already updated | Modified during apply phase (task 3.4) — Language-Aware Template Resolution requirement (no delta sync needed) |

---

## Archive Contents (Filesystem)

```
openspec/changes/archive/2026-05-13-i18n-catalog/
├── proposal.md          ✅ — Change intent, scope, approach, risks
├── design.md            ✅ — Architecture decisions, data flow, API contracts
├── tasks.md             ✅ — 18/18 tasks complete across 3 phases
├── verify-report.md     ✅ — Verification verdict + issues
└── archive-report.md    ✅ — This file
```

---

## Verification Summary

| Metric | Result |
|--------|--------|
| Tasks complete | 18/18 (100%) |
| Build | Clean (`go build ./...`) |
| Vet | Clean (`go vet ./...`) |
| Tests | 18/18 packages PASS |
| Coverage (i18n) | 82.8% |
| Coverage (screens) | 87.7% |
| Golden files | 14 regenerated, all pass |
| Spec compliance | 15/20 COMPLIANT, 4/20 PARTIAL, 0 FAILING |
| Critical issues | 0 |
| Warnings | 3 (non-blocking) |
| Suggestions | 3 (nice-to-have) |

---

## Implementation Details

### New Packages
- `internal/i18n/` — Bundle, T(), Init(), Initialized(), 65 message keys, en+es TOML catalogs

### New Files (10)
- `internal/i18n/bundle.go`, `bundle_test.go`, `keys.go`, `keys_test.go`
- `internal/i18n/translations/en.toml`, `es.toml`, `embed.go`
- `internal/tui/screens/main_test.go`
- `adapters/common/testdata/test.en.tmpl`

### Modified Files (25+)
- `go.mod` — `go-i18n/v2 v2.6.1`, `x/text v0.32.0`
- `internal/app/model.go`, `view.go`, `update.go`
- 8 screen `*.go` files + 8 `*_test.go` files
- 14 golden files in `testdata/golden/`
- `adapters/common/template.go`, `template_test.go`, `base_adapter.go`
- `adapters/codex/adapter.go`, `gemini/adapter.go`, `_template/adapter.go`
- `adapters/registry_test.go`, `internal/tui/screens/helpers_test.go`
- `openspec/specs/go-wiring/spec.md`, `template-wiring/spec.md`

### Dependencies Added
- `github.com/nicksnyder/go-i18n/v2` v2.6.1 (MIT, 2.6k stars)
- `golang.org/x/text` v0.32.0 (upgraded from indirect v0.3.8)

---

## SDD Cycle Complete

The i18n-catalog change has been fully planned, implemented, verified, and archived. All artifacts are preserved in both engram memory and the openspec filesystem archive. Ready for the next change.
