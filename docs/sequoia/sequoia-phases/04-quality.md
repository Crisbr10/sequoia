# P4 — Code Quality & Maintainability Audit

> **Agent**: P4 sequoia-quality
> **Date**: 2026-05-12
> **Target**: Sequoia CLI — Go 1.24.2, Cobra + Bubbletea + Lipgloss
> **Scope**: 85+ Go source files, 64 test files, 48 dependencies

---

## 📊 Dependency Health Table

| Dependency | Version | Status | Risk | Notes |
|---|---|---|---|---|
| github.com/spf13/cobra | v1.10.2 | ✅ Active | Low | Well-maintained CLI framework |
| github.com/charmbracelet/bubbletea | v1.3.10 | ✅ Active | Low | Active TUI framework |
| github.com/charmbracelet/lipgloss | v1.1.0 | ✅ Active | Low | Styles library, same org |
| github.com/stretchr/testify | v1.11.1 | ✅ Active | Low | Standard Go test library |
| github.com/BurntSushi/toml | v1.6.0 | ✅ Stable | Low | Mature TOML parser |
| gopkg.in/yaml.v3 | v3.0.1 | ✅ Stable | Low | Mature YAML parser |
| github.com/common-nighthawk/go-figure | v0.0.0-20210622… | 🟡 Stale | Low | ASCII art — last update 2021-06. No logic risk |
| **27 indirect deps** | various | ✅ All Active | Low | All from charmbracelet/x ecosystem or stdlib extensions |

**Verdict**: No abandoned packages. No known CVEs. All indirect deps are transitive from active direct deps. The go-figure ASCII art library is feature-complete and frozen — acceptable risk.

---

## 📈 Test Coverage Summary

`coverage.out` contains data **only** from `cmd/sequoia/main.go` and `cmd/sequoia/terminal.go` (76 coverage lines). All other packages (`internal/`, `adapters/`) have **zero recorded coverage**. The coverage tool was run against the main package only — not the full project.

| Metric | Value |
|---|---|
| Test files | 64 |
| Test functions | ~85 identified |
| Coverage modules measured | 1/12 (main package only) |
| Unmeasured packages | `adapters/*` (6 sub-packages), `internal/app`, `internal/pipeline`, `internal/tui/*`, `internal/model`, `plugin/*`, `adapters/common` |

**Assessment**: The coverage run is incomplete. Run `go test -coverprofile=coverage.out ./...` to measure all packages.

---

## 🔍 Findings

### [P4-001] · Massive code duplication across 5 adapter packages [🔴 CRÍTICO]
**State**: Confirmed
**Evidence**: 
- `adapters/claude/installer.go:17-87` ≡ `adapters/gemini/installer.go:17-87` — Byte-for-byte identical `InjectSection`/`RemoveSection`
- `adapters/cursor/installer.go:18-90` ≡ `adapters/opencode/installer.go:18-90` ≡ `adapters/_template/installer.go:27-101` — Byte-for-byte identical `GenerateRulesMD`/`RemoveRulesMD`/`isSequoiaManaged`
- `adapters/claude/adapter.go:125-207` ≈ `adapters/opencode/adapter.go:125-209` ≈ `adapters/cursor/adapter.go:120-204` ≈ `adapters/gemini/adapter.go:121-203` ≈ `adapters/_template/adapter.go:150-241` — `Install()` method ~90% identical across all 5 adapters
- `adapters/claude/adapter.go:210-227` ≈ `adapters/opencode/adapter.go:212-229` ≈ `adapters/cursor/adapter.go:207-224` — `Uninstall()` method ~85% identical
- `sequoiaMarker` constant duplicated in `claude/adapter.go:15`, `opencode/adapter.go:15`, `gemini/adapter.go:13`
- `markerStart`/`markerEnd` constants duplicated in `claude/installer.go:10-11`, `gemini/installer.go:10-11`, `opencode/installer.go:10-11`, `cursor/installer.go:10-11`, `_template/installer.go:10-11`
**Problem**: ~300 lines of byte-for-byte identical code duplicated across 5 adapter packages, plus ~500 lines of near-identical `Install`/`Uninstall`/`Status` methods. The duplicate installer functions (`InjectSection`, `GenerateRulesMD`, `RemoveRulesMD`, `isSequoiaManaged`) should live in `adapters/common/`. The adapter `Install`/`Uninstall` flows share a pattern (stage templates → create dirs → install skills → install commands → inject/replace system prompt → write version) that can be factored into a `common.InstallFlow` helper accepting a strategy enum.
**Real Impact**: Every bug fix in the installer logic must be applied 3–5 times. Adding a new adapter requires copy-pasting ~200 lines of boilerplate. The `_template` directory exists specifically because the duplication is acknowledged but not resolved. Risk of adapter drift increases with each change.
**Minimal High-Leverage Recommendation**: 
1. Move `InjectSection`/`RemoveSection` and `GenerateRulesMD`/`RemoveRulesMD` to `adapters/common/installer_strategies.go`
2. Extract the shared Install flow into `common.InstallFlow(config InstallFlowConfig)` in `adapters/common/install_flow.go`
**Dependencies/Blockers**: Must preserve all existing behavior — add tests for the refactored common functions first.
**Implementation Risk**: Medium — affects all 5 production adapters. Requires careful test migration.
**Acceptance Criteria**:
- [ ] `InjectSection`/`RemoveSection` exists exactly once in `adapters/common/`
- [ ] `GenerateRulesMD`/`RemoveRulesMD`/`isSequoiaManaged` exists exactly once in `adapters/common/`
- [ ] No adapter `.go` file is byte-for-byte identical to another
- [ ] All existing adapter tests pass without modification
- [ ] Adding a new adapter requires ≤50 lines of unique code (not counting templates)

---

### [P4-002] · `_template` package is importable and self-registers to production registry [🔴 CRÍTICO]
**State**: Confirmed
**Evidence**: `adapters/_template/adapter.go:30-32` — `init()` function calls `adapters.DefaultRegistry.Register(&Adapter{})` with `ID()="template"`, `Name()="Template Tool"`. The package has valid `package template` declaration and `embed.FS` at `adapters/_template/embed.go:6`.
**Problem**: The `_template` directory is a valid, compilable Go package. Its `init()` function registers a "Template Tool" adapter into the global `DefaultRegistry`. If anyone adds `_ "github.com/Crisbr10/sequoia/adapters/template"` or if a future refactor renames the underscore, the `template` adapter would appear in production with `ID()="template"` — polluting the status table, install targets, and TUI screens. The underscore prefix only prevents `go build ./...` from including it; it does not prevent explicit imports.
**Real Impact**: Production registry pollution. A user running `sequoia status` could see "Template Tool" listed. An accidental import would register real-but-broken install/uninstall logic into production.
**Minimal High-Leverage Recommendation**: Add `//go:build ignore` at the top of every file in `adapters/_template/`, or move the directory to `docs/examples/template-adapter/` outside the module path.
**Dependencies/Blockers**: None — pure additive change.
**Implementation Risk**: Low — the template is not imported by any production code.
**Acceptance Criteria**:
- [ ] `go build ./...` does not compile `adapters/_template/`
- [ ] `sequoia status` never shows "Template Tool"
- [ ] Template files remain accessible for documentation purposes

---

### [P4-003] · `opts.Language` is plumbed through entire stack but never used [🟠 RIESGO]
**State**: Confirmed
**Evidence**: 
- `adapters/interface.go:28` — `InstallOpts.Language` field defined
- `adapters/claude/adapter.go:126` — `_ = opts.Language` (ignored)
- `adapters/opencode/adapter.go:126` — `_ = opts.Language` (ignored)
- `adapters/cursor/adapter.go:121` — `_ = opts.Language` (ignored)
- `adapters/gemini/adapter.go:122` — `_ = opts.Language` (ignored)
- `adapters/codex/adapter.go:127` — `_ = opts.Language` (ignored)
- `adapters/_template/adapter.go:151` — `_ = opts.Language` (ignored)
- `internal/pipeline/runner.go:78` — `opts := adapters.InstallOpts{Language: lang}` (passed but unused)
- `internal/app/model_test.go:36,37` — `_ = opts.Language` (test mock)
- `internal/app/model_internal_test.go:29,30` — `_ = opts.Language` (test mock)
**Problem**: The `Language` field is accepted in the interface and passed through the entire pipeline (TUI → Model → Pipeline → Adapter) but every adapter discards it with `_ = opts.Language`. This is dead plumbing — ~15 locations with `_ = opts.Language` across 10 files. It signals incomplete i18n implementation. See P7 findings for the full i18n impact.
**Real Impact**: The `InstallOpts` struct carries dead weight through every function call. The TUI Configuration screen offers a language selector that produces no observable change in installed content. Users who select "es" get the same English templates as "en".
**Minimal High-Leverage Recommendation**: Either (a) implement template localization using the language parameter to select `.tmpl` files from language-specific subdirectories, or (b) remove the `Language` field and the configuration screen selector until i18n is ready.
**Dependencies/Blockers**: P7 i18n agent should coordinate.
**Implementation Risk**: Low — removal is trivial; implementation requires template translation work.
**Acceptance Criteria**:
- [ ] No `_ = opts.Language` in production adapter code OR language selector affects template output
- [ ] Configuration screen reflects actual capability

---

### [P4-004] · Incomplete test coverage — 0% on all non-main packages [🟠 RIESGO]
**State**: Confirmed
**Evidence**: `coverage.out` (76 lines) contains only `mode: set` entries for `cmd/sequoia/main.go` and `cmd/sequoia/terminal.go`. No coverage data for `adapters/` (6 sub-packages), `internal/app`, `internal/pipeline`, `internal/tui/`, `internal/model`, `plugin/`, `adapters/common/`.
**Problem**: While 64 test files exist and ~85 test functions are written, the `go test -coverprofile` command was only run against the main package. The coverage data is so narrow it provides no visibility into adapter, pipeline, model, or TUI coverage. Key untested areas based on code inspection:
- `adapters/common/installer.go` — `Rollback()`, `Verify()`, `Prepare()` have tests but unknown coverage
- `internal/pipeline/runner.go` — `RunStatus()` tested, `RunInstall`/`RunUninstall` tested but `sendProgress()` is a helper
- `cmd/sequoia/main.go` — `runUninstall()` lines 288-294 unreached (0 count), `runStatus()` partially covered
**Real Impact**: Unknown coverage gaps in critical install/uninstall logic. Regression risk when refactoring adapter code (P4-001).
**Minimal High-Leverage Recommendation**: Run `go test -coverprofile=coverage.out ./...` and regenerate `coverage.out`. Add coverage enforcement to CI with `-cover` flag.
**Dependencies/Blockers**: None.
**Implementation Risk**: Low.
**Acceptance Criteria**:
- [ ] `coverage.out` contains data for all packages
- [ ] Coverage report identifies packages below target (suggest 70% minimum)
- [ ] CI workflow includes coverage check

---

### [P4-005] · `runInstallSteps` and `runUninstallSteps` are 95% identical [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `internal/pipeline/runner.go:75` vs `internal/pipeline/runner.go:161` — both functions share identical structure: iterate `defaultStepNames`, send "running" progress, call adapter method, send completion/error messages. Only difference: `adapter.Install(opts)` on line 92 vs `adapter.Uninstall(opts)` on line 176.
**Problem**: 85 lines of duplicated code. Any change to step ordering, progress reporting format, or error handling must be applied twice. The `defaultStepNames` constant is shared correctly, but the step runner logic is not.
**Real Impact**: Maintenance burden. If a new step is added (e.g., "Validation"), both functions must be updated. Risk of divergence.
**Minimal High-Leverage Recommendation**: Extract `runSteps(ctx, t, ch, lang, fn)` where `fn` is `func(adapters.ToolAdapter, adapters.InstallOpts) error`. Pass either `adapter.Install` or `adapter.Uninstall` as the function argument.
**Dependencies/Blockers**: None.
**Implementation Risk**: Low — mechanical refactor, tests already cover both paths.
**Acceptance Criteria**:
- [ ] Single `runSteps` function exists in `pipeline/runner.go`
- [ ] `TestRunInstall_*` and `TestRunUninstall_*` tests pass unchanged
- [ ] Line count of `runner.go` reduced by at least 40 lines

---

### [P4-006] · 4 duplicate mock adapter implementations across test packages [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: 
- `internal/app/model_test.go:26-44` — `mockAdapter` (11 methods)
- `internal/app/model_internal_test.go:19-37` — `stubAdapter` (11 methods)
- `internal/tui/screens/helpers_test.go:11-27` — `dummyAdapter` (11 methods)
- `internal/pipeline/runner_test.go:22-84` — `testAdapter` (11 methods + concurrency-safe counters)
**Problem**: Four test packages each define their own `ToolAdapter` mock, totaling ~150 lines of duplicated boilerplate. The `mockAdapter`, `stubAdapter`, and `dummyAdapter` are functionally identical (11 stub methods, no state). `testAdapter` is slightly more complex (mutex-protected call counters, install error injection). These should live in a shared `adapters/adapters_test` or `internal/testutil` package.
**Real Impact**: Adding a new method to `ToolAdapter` interface requires updating 4 mock types plus the 2 production adapter stubs. Compile-failure risk when interface changes.
**Minimal High-Leverage Recommendation**: Create `adapters/testutil/mock_adapter.go` with a configurable `MockAdapter` struct that all test packages import. Support `InstallFunc`/`UninstallFunc` field injection for behavior customization.
**Dependencies/Blockers**: None — internal refactor only.
**Implementation Risk**: Low — search-and-replace with a shared import.
**Acceptance Criteria**:
- [ ] Single `MockAdapter` type in a shared test package
- [ ] All 4 test packages import the shared mock
- [ ] `_ adapters.ToolAdapter = (*MockAdapter)(nil)` compile-time check still passes

---

### [P4-007] · Fragile recursion limit in test command helper [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `internal/app/integration_test.go:31` — `safeProcessCmd(m, cmd, remaining int)` with hardcoded `maxCmd` limits (3 or 5) passed from `sendKey()`. If a tea.Batch produces more than `remaining` nested commands, the chain is silently truncated, and assertions on final state may pass incorrectly or fail confusingly.
**Problem**: The `remaining` parameter is an arbitrary depth limit. Tests that call `sendKey(m, key, 3)` assume the command chain will resolve within 3 iterations. If a future change adds an intermediate command (e.g., a validation step), the test will silently stop processing at depth 3, and assertions may fail with no indication it's a depth-limit issue rather than a logic bug.
**Real Impact**: Brittle tests — adding a new tea.Cmd in the pipeline chain can break integration tests without any error message indicating the real cause. Developer time wasted debugging false failures.
**Minimal High-Leverage Recommendation**: Replace recursion with an explicit loop (e.g., `maxIterations := 20`) and log a warning or fail the test if the limit is reached: `t.Fatalf("command chain exceeded %d iterations — possible infinite loop or new intermediate command", maxIterations)`.
**Dependencies/Blockers**: None.
**Implementation Risk**: Low — the function is test-only.
**Acceptance Criteria**:
- [ ] `safeProcessCmd` uses an explicit loop instead of recursion
- [ ] Exceeding the limit produces a clear test failure with descriptive message

---

### [P4-008] · Go version `1.24.2` in go.mod is unnecessarily restrictive [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `go.mod:3` — `go 1.24.2`. The codebase uses no Go 1.24 features: no generic type aliases, no `slices`/`maps` standard library additions from 1.21–1.24, no `range over func` iterators (1.22/1.23). The code compiles with standard patterns available since Go 1.19+.
**Problem**: The `go` directive sets the minimum toolchain version. Declaring `1.24.2` prevents users with Go 1.22 or 1.23 from building the project, even though the code is compatible. The project context says "min 1.22" which contradicts the go.mod.
**Real Impact**: Users on Go 1.22 or 1.23 cannot `go install` or `go build` this project. The readme claims Go 1.22 minimum but go.mod requires 1.24.2.
**Minimal High-Leverage Recommendation**: Lower `go 1.24.2` to `go 1.22.0` in go.mod. Run `go mod tidy -go=1.22` to regenerate. Verify builds on Go 1.22 CI matrix.
**Dependencies/Blockers**: Check if any indirect dep requires >=1.23.
**Implementation Risk**: Low — the codebase has no 1.24-specific features.
**Acceptance Criteria**:
- [ ] `go.mod` specifies `go 1.22` (or lowest version that all deps support)
- [ ] README min version matches go.mod
- [ ] CI successfully builds with Go 1.22

---

### [P4-009] · Unreachable completion in `runInstallSteps` failure path with empty `defaultStepNames` [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `internal/pipeline/runner.go:96-98` — `if len(defaultStepNames) > 0 { sendProgress(...) }`. The `defaultStepNames` variable on line 18 is a package-level `var` (mutable) initialized to `[]string{"Skills", "Commands", "System Prompt"}`, so it is always non-empty in normal execution. The `len > 0` guard is unreachable dead code unless the variable is modified externally.
**Problem**: The `if len(defaultStepNames) > 0` check implies a design intent to support empty step lists, but the code never exercises this path. If `defaultStepNames` were ever emptied (e.g., by a test or a malicious package), the error would be silently swallowed — `adapter.Install()` fails but no error progress message is sent, and the TUI would never show the failure. This is both dead code AND a silent-error-swallowing bug when triggered.
**Real Impact**: Low in practice (the variable is never modified). However, it represents a latent bug: the error path silently drops the error when there are no steps. The fix should either remove the guard (making the code panic on empty steps, which is better than silent error swallowing) or handle the empty case explicitly.
**Minimal High-Leverage Recommendation**: Either remove the `len > 0` check (if `defaultStepNames` is always expected to be non-empty) or send an error progress message with step `"install"` when no step names are defined.
**Dependencies/Blockers**: None.
**Implementation Risk**: Low.
**Acceptance Criteria**:
- [ ] No silent error swallowing path exists in `runInstallSteps` or `runUninstallSteps`
- [ ] Empty `defaultStepNames` either panics (developer error) or reports error correctly

---

### [P4-010] · `renderUninstallConfirm` bypasses screen abstraction [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `internal/app/view.go:40-43` — `renderUninstallConfirm()` calls `styles.Accent().Render(...)` and `styles.Muted().Render(...)` directly from the `app` package. All other view rendering is delegated to `internal/tui/screens/*` (e.g., `screens.WelcomeView()`, `screens.StatusView()`). This function is tested in `model_internal_test.go:144-149` (`TestRenderUninstallConfirm_ContainsPrompt`).
**Problem**: The `app` package has a direct dependency on `internal/tui/styles` for a single function, while all other rendering is done through the `screens` package. This breaks the rendering architecture's single-responsibility pattern — the `app` package should own state management (Model + Update + View dispatch), not rendering logic.
**Real Impact**: Minimal — one 2-line function. Architectural inconsistency that makes the codebase harder to reason about.
**Minimal High-Leverage Recommendation**: Move `renderUninstallConfirm()` to `internal/tui/screens/uninstall.go` and export it as `RenderConfirmPrompt()`. Update `View()` to call `screens.RenderConfirmPrompt()`.
**Dependencies/Blockers**: None.
**Implementation Risk**: Low — mechanical move.
**Acceptance Criteria**:
- [ ] `app` package no longer imports `internal/tui/styles`
- [ ] `renderUninstallConfirm` tests are in the `screens` test package
- [ ] `View()` delegate only — no inline rendering logic

---

### [P4-011] · `_template` package has 20 TODO comments in "shipping" code [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: 
- `adapters/_template/adapter.go` — lines 2, 7, 47, 53, 58, 69, 76, 90, 97, 219, 262 (11 TODOs)
- `adapters/_template/paths.go` — lines 13, 34, 40, 46, 51, 68 (6 TODOs)
- `adapters/_template/installer.go` — lines 19, 61 (2 TODOs)
- `adapters/_template/install.go` — line 6 (1 TODO)
**Problem**: The _template directory is intentionally a boilerplate starting point, so TODOs are expected. However, combined with P4-002, these TODOs exist in code that could compile into a production binary. A static analysis tool like `golangci-lint` with `godox` linter would flag all 20 as issues if the package were included in scanning.
**Real Impact**: Only relevant if `_template` is accidental imported (see P4-002). Currently negligible.
**Minimal High-Leverage Recommendation**: Resolved automatically by applying P4-002 (build tag exclusion). No separate action needed.
**Dependencies/Blockers**: P4-002.
**Implementation Risk**: N/A — resolved transitively.
**Acceptance Criteria**:
- [ ] TODOs do not appear in any compiled binary
- [ ] `golangci-lint` (or `godox` linter) does not flag template TODOs in production packages

---

### [P4-012] · Test `TestWaitForProgress_ContextCancellationIgnored` tests a non-behavior [🟡 ATENCIÓN]
**State**: Confirmed
**Evidence**: `internal/app/model_internal_test.go:231-249` — Creates a context, cancels it, but the test acknowledges `waitForProgress` doesn't use context. The test verifies that a message is returned "regardless of external state" which is trivially true since `waitForProgress` only reads from a channel and has zero awareness of the context. `_ = ctx` on line 242 explicitly discards the context, confirming it's unused.
**Problem**: The test name implies it validates context cancellation behavior but actually tests nothing meaningful — it just verifies basic channel receive works. The context is created, cancelled, and discarded. This test provides no signal and could be deleted without losing any coverage.
**Real Impact**: Zero — test passes but provides no value. Slightly increases test runtime and cognitive load.
**Minimal High-Leverage Recommendation**: Remove the test or replace it with a meaningful test that verifies `waitForProgress` returns `nil` when the channel is closed (already covered by `TestWaitForProgress_ClosedChannel_ReturnsNil`).
**Dependencies/Blockers**: None.
**Implementation Risk**: Low.
**Acceptance Criteria**:
- [ ] Removed or replaced with a test that exercises actual behavior

---

## 📊 Summary Table

| ID | Severity | Title | Risk |
|---|---|---|---|
| P4-001 | 🔴 CRÍTICO | Massive code duplication across 5 adapter packages | High |
| P4-002 | 🔴 CRÍTICO | `_template` package is importable and self-registers | Low |
| P4-003 | 🟠 RIESGO | `opts.Language` plumbed but never used (dead i18n) | Low |
| P4-004 | 🟠 RIESGO | Incomplete test coverage — 0% on non-main packages | Low |
| P4-005 | 🟡 ATENCIÓN | `runInstallSteps`/`runUninstallSteps` 95% identical | Low |
| P4-006 | 🟡 ATENCIÓN | 4 duplicate mock adapter implementations | Low |
| P4-007 | 🟡 ATENCIÓN | Fragile recursion limit in test helper | Low |
| P4-008 | 🟡 ATENCIÓN | Go 1.24.2 go.mod is unnecessarily restrictive | Low |
| P4-009 | 🟡 ATENCIÓN | Dead code guard + silent error swallowing in pipeline | Low |
| P4-010 | 🟡 ATENCIÓN | `renderUninstallConfirm` bypasses screen abstraction | Low |
| P4-011 | 🟡 ATENCIÓN | `_template` has 20 TODOs in compilable code | N/A |
| P4-012 | 🟡 ATENCIÓN | Meaningless test for non-behavior | Low |

---

## ✅ Positive Findings

1. **Error wrapping is consistent** — All production errors use `fmt.Errorf("context: %w", err)`. No bare error returns found.
2. **No panics in production code** — The only `panic()` is in a third-party test file (`docs/Ejemplos de Skills/engram/`).
3. **Good interface design** — `ToolAdapter` interface in `adapters/interface.go` is clean, well-documented, and has exactly the right surface area for the contract.
4. **Registry pattern is testable** — `DefaultRegistry` is a mutable global, but tests properly lock (`registryMu`) and restore it. Tests do NOT run in parallel when mutating the registry.
5. **`common.Installer` has proper rollback** — The four-phase Prepare→Apply→Verify→Rollback pattern with backup/restore is well-implemented.
6. **Context cancellation is respected** — Pipeline goroutines check `ctx.Done()` at launch and between steps. Channels are always closed.
7. **Go doc comments are thorough** — Every exported symbol has a doc comment. Private helpers like `hasSelectedInstalled` and `countSelected` are also documented.
8. **Naming is consistent** — Follows Go conventions (camelCase unexported, PascalCase exported). No stuttering (`adapters.Adapter` would be bad; using `ToolAdapter` is good).
9. **Build tags used for templates** — `//go:embed templates` is used correctly in each adapter's `embed.go`.
10. **No goroutine leaks detected** — `sync.WaitGroup` used correctly in pipeline. Channel lifecycle (create → fill → close) is clean.

---

## 🎯 Recommended Action Priority

1. **Immediate**: Apply P4-002 (`_template` build tag) — 5 minutes, eliminates production registry risk.
2. **Short-term**: Start P4-001 (installer consolidation) with `InjectSection`/`RemoveSection` and `GenerateRulesMD`/`RemoveRulesMD` moves — these are byte-identical and zero-risk to move.
3. **Short-term**: Fix P4-008 (go.mod version) — verify and lower.
4. **Medium-term**: Apply P4-004 (full coverage run), P4-005 (pipeline DRY), P4-006 (shared mock).
5. **Later**: P4-003 (i18n — coordinate with P7), P4-007/P4-010/P4-012 (minor test/architecture improvements).
