# Proposal: Remove I18n — English Only

## Intent

Sequoia's i18n subsystem (`internal/i18n/`, go-i18n/v2, Spanish/English TOML catalogs) is dead weight. The language selector on the Configuration screen is cosmetic — all adapters use it only to pick between `templates/{en,es}/` subdirectories. Agent instruction `.md` files are written in mixed Spanish/English, reducing AI comprehension. Removing i18n simplifies the codebase (~815 lines deleted, ~720 lines simplified), eliminates 2 dependencies, and rewrites all skill docs in pure English for better agent performance.

## Scope

### In Scope
- Delete `internal/i18n/` package (7 files, ~815 lines)
- Hardcode English strings in 8 TUI screen view functions; remove `lang string` param
- Remove `Language` type/constants/field from `model/types.go`; remove `InstallOpts.Language`
- Remove `lang` parameter from `pipeline.RunInstall`/`RunUninstall`
- Remove `RenderTemplateLang`; consolidate template paths to single language (drop `templates/es/`)
- Delete `agent-p7-i18n` spec, doc, and roster references
- Remove `go-i18n/v2` and `golang.org/x/text` from go.mod; keep `BurntSushi/toml`
- Regenerate 15 golden test files; update ~100 test assertions
- Rewrite ~30 agent/skill `.md` files in pure English (~4,500 lines)
- Pass `go test -race -count=1 ./...`

### Out of Scope
- README.md, ARCHITECTURE.md, SEQUOIA.md, release notes, installer docs, getting-started, FAQ
- Adapters: claude, cursor, gemini, opencode (confirmed no `opts.Language` references)
- BurntSushi/toml (used by Codex adapter for TOML merging — must keep)

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- **i18n-engine** — REMOVED. Entire spec deleted.
- **agent-p7-i18n** — REMOVED. Spec, agent doc, roster entry deleted.
- **tui-core** — Screen views drop `lang string` parameter; all strings hardcoded English.
- **tui-management** — Configuration screen loses language selector (2-field → 1-field).
- **tui-pipeline** — `RunInstall`/`RunUninstall` drop `lang` parameter.
- **go-wiring** — `InstallOpts.Language` removed; `Language` type removed; `go-i18n/v2` and `x/text` removed.
- **template-wiring** — `RenderTemplateLang` removed; `templates/es/` deleted; P7 references stripped from all templates.

## Approach

Two-track, sequential:

**Track 1 — Code**: Delete i18n package → inline hardcoded English strings in 8 screens → remove `lang` params and `Language` model fields → remove deps + `go mod tidy` → update tests → regenerate golden files → verify test suite green.

**Track 2 — Docs**: Rewrite ~30 `.md` skill/agent files in native English → delete `sequoia-i18n.md` → strip P7 from agent roster tables in all adapter templates → verify template rendering.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| 15 golden files need regeneration | Moderate | Compile-safe; regeneration script + manual review |
| Configuration screen UI layout change (2→1 field) | Low | Single-field cursor simplifies navigation logic |
| Template path consolidation affects adapter behavior | Low | All adapters tested in CI; base_adapter handles single path |
| `go mod tidy` removes unexpected indirect deps | Low | Verify full build + test suite after tidy |

## Rollback Plan

All changes in a git branch. Revert to main if needed. Deleted i18n package recoverable from git history.

## Success Criteria

- [x] `go build ./...` compiles with zero errors
- [x] `go test -race -count=1 ./...` passes with zero failures (core)
- [x] No `i18n` import remains in any `.go` file
- [x] `go-i18n/v2` absent from `go.mod`
- [x] 15 golden test files regenerated and match expected English-only output
- [x] All ~30 skill docs rewritten in pure English with no Spanish text
- [x] `docs/agents/sequoia-i18n.md` deleted
- [x] `openspec/specs/i18n-engine/` and `openspec/specs/agent-p7-i18n/` archived
