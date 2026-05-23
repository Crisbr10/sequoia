# Tasks: Global Mutable State Removal (CORR-002)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~240 (additions + deletions) |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | Single PR (mechanical changes under budget, tightly coupled phases) |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Medium

> 18 files, ~240 diff lines. Mechanical replacements dominate; risk is Medium due to breadth, not volume.
> Registry_test.go needs structural changes (not pure find-replace). Single PR feasible.

---

## Phase 1: Test Migration (TDD Red — break tests before production code)

- [ ] 1.1 `internal/app/model_test.go`: Replace 28 `adapters.DefaultRegistry` → `adapters.NewRegistry()`. **Acceptance**: `go test ./internal/app/` fails (DefaultRegistry gone) OR passes (all callers use NewRegistry()).
- [ ] 1.2 `adapters/registry_test.go`: Replace 4 `DefaultRegistry` refs (TestFactory_NewAdapter_KnownID L105-110, TestFactory_NewAdapter_UnknownID L119) with explicit `r := adapters.NewRegistry()`, manual `r.Register(a)`, `r.Get()`. **Acceptance**: tests still pass with no global.
- [ ] 1.3 `adapters/common/shared_test.go` L36 + 4 test files: Change 11 `common.CommandFiles` → `common.CommandFiles()` across: `shared_test.go` (1), `base_adapter_test.go` (4), `base_adapter_error_test.go` (4 L222,244,269,701 index access also needs `()` update), `base_adapter_mockfs_test.go` (1), `command_template_test.go` (2). **Acceptance**: compilation fails (CommandFiles still a var).

## Phase 2: Production Code (TDD Green — make tests pass)

- [ ] 2.1 `adapters/common/commands.go`: Convert `var CommandFiles = []string{...}` to unexported `var commandFiles` + exported `func CommandFiles() []string { return append([]string{}, commandFiles...) }`. **Acceptance**: Phase 1.3 tests compile and pass.
- [ ] 2.2 `adapters/common/base_adapter.go` (L323,386,466) + `adapters/_template/adapter.go` (L189,226,271) + `adapters/codex/adapter.go` (L107,140,190): Add `()` to 9 `common.CommandFiles` refs. **Acceptance**: `go build ./adapters/...` succeeds.
- [ ] 2.3 `adapters/registry.go`: Delete `DefaultRegistry` variable (L40-44). Update L9 doc comment: remove "init() functions" reference. **Acceptance**: `go build ./...` compiles; Phase 1.1-1.2 tests pass.
- [ ] 2.4 Delete `init()` blocks from 6 adapter packages: `claude/adapter.go` L27-29, `codex/adapter.go` L30-32, `cursor/adapter.go` L25-27, `gemini/adapter.go` L31-33, `opencode/adapter.go` L27-29, `_template/adapter.go` L43-45. **Acceptance**: `go build ./...` compiles; zero `func init()` in `adapters/*/adapter.go`.
- [ ] 2.5 `cmd/sequoia/main.go` L22: Update import comment — remove `init() +` mention. **Acceptance**: comment reflects DI-only pattern.

## Phase 3: Documentation

- [ ] 3.1 `CONTRIBUTING.md`: Rewrite Step 2 to show `reg := adapters.NewRegistry(); claude.RegisterIn(reg)` pattern. Remove Step 7 blank import instruction. **Acceptance**: Doc shows DI pattern with zero `DefaultRegistry` or `init()` mentions.

## Phase 4: Verification

- [ ] 4.1 `adapters/common/shared_test.go`: Add `TestCommandFiles_Immutability` — call `CommandFiles()`, mutate `result[0] = "evil"`, call again, assert original list returned. **Acceptance**: test passes.
- [ ] 4.2 Run `go test -race -count=1 ./...` — zero failures across all packages. **Acceptance**: all Spec scenarios satisfied (explicit registry, no auto-registration, parallel isolation, defensive copy).
- [ ] 4.3 Run `go vet ./...` — zero issues. **Acceptance**: no vet warnings.
- [ ] 4.4 Grep verification: `rg "DefaultRegistry" --include="*.go"` returns zero in production + test code. `rg "func init\(\)" adapters/*/adapter.go` returns zero. **Acceptance**: zero global mutable state references.
