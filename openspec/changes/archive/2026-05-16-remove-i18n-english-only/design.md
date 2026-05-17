# Design: Remove I18n — English Only

## Technical Approach

Two tracks, sequential. Track 1 (code): delete i18n package → inline English strings in 8 screen views → remove `lang` from all signatures → strip `Language` from model/types and `InstallOpts` → remove 2 Go deps → regenerate 15 golden files → verify `go test -race ./...` passes. Track 2 (docs): rewrite ~30 skill `.md` files in pure English, delete P7 agent files, strip P7 from adapter templates.

## Architecture Decisions

### Decision 1: Hardcoded English String Strategy

| Option | Tradeoff | Decision |
|--------|----------|----------|
| A) Inline strings at call site | Direct, self-documenting, no imports | **Chosen** |
| B) Convert keys.go constants to hold English values | Centralized but pointless without translation | Rejected |
| C) New `messages.go` package | Over-engineering for 66 strings | Rejected |

**Rationale**: The 66 message keys serve 66 unique UI positions across 8 screens. Each screen uses ~8–15 strings. Inlining makes view code self-documenting — you read `"Install"` not `i18n.T(i18n.MsgWelcomeMenuInstall, lang)`. Key indirection without translation is pure overhead.

### Decision 2: Language Field Removal Order

| Option | Tradeoff | Decision |
|--------|----------|----------|
| A) Top-down (model → screens → render) | Structural but breaks compilation at each step | Rejected |
| B) Bottom-up (render → screens → model) | Incremental but mismatched intermediates | Rejected |
| C) Side-by-side (all at once) | Fastest; compiler guarantees correctness | **Chosen** |

### Decision 3: Golden File Regeneration

| Option | Tradeoff | Decision |
|--------|----------|----------|
| A) `UPDATE_GOLDEN=1 go test` to auto-regenerate | Automated, consistent, uses existing infra | **Chosen** |
| B) Manually edit each file | Error-prone with ANSI codes | Rejected |
| C) Delete and regenerate | Same as A with extra steps | Rejected |

### Decision 4: Template Path Consolidation

| Option | Tradeoff | Decision |
|--------|----------|----------|
| A) Move templates to root | No lang dirs exist — moot | N/A |
| C) Replace `RenderTemplateLang` with `RenderTemplate` | Simplest; lang resolution is dead code | **Chosen** |

### Decision 5: Doc Rewriting Strategy

| Option | Tradeoff | Decision |
|--------|----------|----------|
| A) Rewrite each file in place | Preserves git history, per-file precision | **Chosen** |
| B) New English versions, delete old | Extra churn, loses blame continuity | Rejected |

## File Change Map (Track 1)

| # | File | Action | Description |
|---|------|--------|-------------|
| 1 | `internal/i18n/*` (7 files) | **Delete** | Entire package |
| 2 | `internal/model/types.go` | Modify | Remove `Language` type, constants, field |
| 3-10 | `internal/tui/screens/*.go` (8 files) | Modify | Drop `lang` param, inline English strings |
| 11-13 | `internal/app/{model,view,update}.go` | Modify | Drop i18n imports, lang params |
| 14 | `internal/pipeline/runner.go` | Modify | Drop `lang` param from RunInstall/RunUninstall |
| 15 | `adapters/interface.go` | Modify | Remove `Language` from `InstallOpts` |
| 16-17 | `adapters/common/{template,base_adapter}.go` | Modify | Delete `RenderTemplateLang`, use `RenderTemplate` |
| 18-19 | `adapters/{codex,_template}/adapter.go` | Modify | Replace `RenderTemplateLang` |
| 20-21 | Test files | Delete/Modify | Remove interface_test.go, update model_test.go |
| 22 | `go.mod` | Modify | Remove `go-i18n/v2` and `x/text`; `go mod tidy` |
| 23 | Golden files (15) | Regenerate | `UPDATE_GOLDEN=1 go test` |

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| **Remove** | `internal/i18n/bundle_test.go` (213 lines) | Delete with package |
| **Remove** | `internal/i18n/keys_test.go` (146 lines) | Delete |
| **Remove** | `adapters/interface_test.go` (44 lines) | Delete |
| **Remove** | `TestConfigurationView_ShowsLanguageOptions` | Delete |
| **Modify** | Screen tests passing `"en"` | Drop 3rd arg |
| **Modify** | `internal/app/model_test.go` | Drop `Language:"en"` expectation |
| **Regenerate** | 15 screen golden files | `UPDATE_GOLDEN=1 go test` |
| **Verify** | Full suite | `go test -race -count=1 ./...` |
