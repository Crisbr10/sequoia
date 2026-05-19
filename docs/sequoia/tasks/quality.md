# Quality Tasks — Sequoia Audit

## Context
Sequoia has 56 test files and claims 90%+ coverage across ~19,200 LOC. The audit found that coverage numbers mask real quality gaps: zero-assertion tests inflate metrics, the TUI router package has no coverage, and cancellation/rollback paths are completely untested. The codebase carries one abandoned dependency (go-figure, 5 years stale), and the shared `adapters/common` package has only 64.8% coverage despite being the foundation for all 5 adapters. The audit identified 11 quality findings: 1 HIGH, 6 MEDIUM, 4 LOW.

**Key files**: `adapters/common/base_adapter.go`, `adapters/common/installer.go`, `adapters/codex/installer.go`, `internal/app/model_test.go`, `go.mod`

## Priority Tiers

### Tier 1 — Immediate (HIGH)

| ID | Task | Effort | Blocks |
|----|------|--------|--------|
| P4-005 | Propagate session file write errors in Codex config merge | small | — |

### Tier 2 — Short Term (MEDIUM)

| ID | Task | Effort |
|----|------|--------|
| MERGED-001 | Replace go-figure with lightweight ASCII art (~20 lines) | small |
| P4-002 | ~~Improve adapters/common test coverage from 64.8% to >80%~~ ✅ DONE via RC-001 | medium |
| P4-003 | Remove zero-assertion tests, add real behavior assertions | small |
| P4-004 | ~~Refactor BaseAdapter.Install() into smaller testable units~~ ✅ DONE via RC-001 | medium |
| P4-009 | Add tests for TUI router (currently [no statements]) | small |

### Tier 3 — Long Term (LOW)

| ID | Task | Effort |
|----|------|--------|
| P4-006 | Add build tag to _template directory to exclude from compilation | small |
| P4-010 | Replace defer-recover with select-ok pattern for channel safety | small |
| P4-011 | ~~Add context-cancellation rollback tests for Install()~~ ✅ DONE via RC-001 | medium |
| P4-012 | Remove global DefaultRegistry mutation from TUI tests | medium |

---

## Detailed Tasks

### P4-005 — Propagate session file write errors in Codex config merge
- **Severity**: HIGH (recalibrated from MEDIUM)
- **Evidence**: `adapters/codex/installer.go:36-38` — error from `AtomicWriteFile(path+".sequoia-session", []byte(suffix), 0o644)` is silently discarded
- **Problem**: If the session tracking file write fails but the backup succeeded, `RemoveConfig` during uninstall falls back to scanning legacy backup names — potentially restoring the wrong backup or failing entirely. Combined with RC-001 (BaseAdapter monolith), this forms part of a pattern of silent data corruption.
- **Fix**:
  1. Collect the error via `adapter.AddWarning(fmt.Sprintf("session file write failed: %v", err))`
  2. Or: return the error wrapped with context: `return fmt.Errorf("merge config: write session: %w", err)`
  3. Add test: mock AtomicWriteFile to fail, verify error is surfaced (not swallowed)
- **Verification**: `go test ./adapters/codex/ -run TestMergeConfig_SessionWriteFails -v`
- **References**: Go best practice: never silently discard errors

### MERGED-001 — Replace go-figure with lightweight ASCII art
- **Severity**: MEDIUM (consolidated from P1-003, P2-002, P4-001)
- **Evidence**: `go.mod:9` — `github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be`; `internal/tui/styles/logo.go:20` — single call: `figure.NewFigure("Sequoia", "", true)`
- **Problem**: go-figure is abandoned (last release June 2021, 5 years stale), linked at ~2MB binary cost, used once to render the word "Sequoia" in the Welcome screen. No security patches, no Go version compatibility guarantees.
- **Fix**:
  1. Hand-craft an ASCII art string (~30 lines) for the "Sequoia" logo
  2. Replace `figure.NewFigure("Sequoia", "", true)` with the hardcoded string
  3. Remove `go-figure` from `go.mod` via `go mod tidy`
  4. Update golden test files for Welcome screen output
- **Verification**: Binary size should drop by ~1.5-3MB. `go mod tidy` removes dependency. Welcome screen golden tests pass.
- **References**: OWASP A06:2021; SLSA provenance; Go dependency hygiene

### P4-002 — Improve adapters/common test coverage from 64.8% to >80% ✅ DONE via RC-001
- **Resolution**: RC-001 added 35 new tests (20 error-path + 15 mock FS) to adapters/common, raising coverage from 64.8% to 84.6%. All error branches in Install(), Uninstall(), Detect(), IsInstalled() are now covered. Context cancellation rollback paths are tested at all 5 checkpoints.

### P4-003 — Remove zero-assertion tests, add real behavior assertions
- **Severity**: MEDIUM
- **Evidence**: `internal/app/model_test.go:98` — `func TestModel_ImplementsBubbleteaModel(_ *testing.T)` — parameter is unnamed `_`, no assertions
- **Problem**: This test has zero `assert.*` or `require.*` calls. It counts toward coverage statistics without verifying behavior. If `m.Init()`, `m.Update()`, or `m.View()` panicked, the test would only fail via Go runtime — not through proper assertions. Creates false confidence.
- **Fix**:
  1. Rename test parameter to `t *testing.T`
  2. Add assertions: `assert.NotNil(t, m.Init())`, `assert.NotEmpty(t, m.View())`
  3. Search codebase for all tests with unnamed `*testing.T`: `grep -r "func Test.*(_ \*testing\.T)"`
  4. Fix each: add at least one meaningful assertion
- **Verification**: All tests have named `t *testing.T` parameter and at least one assertion. Coverage numbers remain meaningful.
- **References**: Go testing best practices: every test should verify at least one concrete behavior

### P4-004 — Refactor BaseAdapter.Install() into smaller testable units ✅ DONE via RC-001
- **Resolution**: RC-001 decomposed the 482-line BaseAdapter into orchestration logic with 4 delegate structs (PathResolver, Detector, PromptManager, BackupPathBuilder). Install() now delegates through named composition. Each extracted struct has its own test suite. Context cancellation is tested at all 5 checkpoints.

### P4-009 — Add tests for TUI router
- **Severity**: MEDIUM
- **Evidence**: `$ go test -cover ./internal/tui/` → `coverage: [no statements]`
- **Problem**: The `internal/tui` package contains `router.go` and `router_test.go`, but coverage shows no statements — meaning either the router has no executable statements or tests don't exercise any paths. The screens sub-package has 88.8% coverage, but the router that connects them is unprotected.
- **Fix**:
  1. Read `internal/tui/router.go` and `router_test.go` to understand current state
  2. If router has no executable code: move screen-to-screen navigation logic from screens to router
  3. If tests exist but don't cover: add test cases for each screen transition (welcome→tool-selection→configuration→install-progress→complete; also error and status paths)
  4. Test keypress routing: simulate key events, verify correct next screen
- **Verification**: `go test -cover ./internal/tui/` shows >70% coverage. Screen transitions tested.
- **References**: Test coverage should include all modules with executable code

### P4-006 — Add build tag to _template directory
- **Severity**: LOW (absorbed by RC-007)
- **Evidence**: `adapters/_template/` — Go files compile into binary but are scaffolding, not functional adapters
- **Problem**: Template files register into adapter registry and reference non-existent embed directories. While they don't cause runtime errors, they add dead code paths to the binary and confuse contributors about whether the template is documentation or functional code.
- **Fix**:
  1. Add `//go:build ignore` to all `.go` files in `adapters/_template/`
  2. Or: move to `docs/adapter-template/` outside Go module path
  3. Combined with RC-007 (_template rewrite)
- **Verification**: `go build ./...` succeeds. Template files excluded from binary.
- **References**: Go convention: scaffold/template code should use build tags

### P4-010 — Replace defer-recover with select-ok pattern
- **Severity**: LOW (absorbed by MERGED-002)
- **Evidence**: `internal/pipeline/runner.go:258-261,275-279` — `safeClose()` and `sendProgress()` both use `defer func() { recover() }()`
- **Problem**: recover() catches ALL panics including nil pointer dereference and index out of bounds, silently converting them to no-ops. A real bug in a goroutine would be swallowed with no stack trace, making debugging nearly impossible.
- **Fix**:
  1. In `sendProgress()`: replace `defer/recover` with `select { case ch <- msg: return true; default: return false }`
  2. In `safeClose()`: use `sync.Once` to track whether channel has been closed: `closeOnce.Do(func() { close(ch) })`
  3. Remove both defer-recover blocks
- **Verification**: Send on closed channel no longer panics. Real bugs (nil dereference) produce stack traces instead of being swallowed.
- **References**: Effective Go: recover should only catch expected panics

### P4-011 — Add context-cancellation rollback tests for Install() ✅ DONE via RC-001
- **Resolution**: RC-001 added 6 context-cancellation tests using a deterministic checkpointCtx wrapper. Tests verify: pre-cancelled context (no work), post-staging cancel (clean staging), post-skill cancel (skill rolled back), post-commands cancel (both rolled back), post-system-prompt cancel (both rolled back), and full success (all 5 checkpoints pass).

### P4-012 — Remove global DefaultRegistry mutation from TUI tests
- **Severity**: LOW (absorbed by RC-005)
- **Evidence**: `internal/app/model_test.go:24-34` — 28/39 tests mutate `adapters.DefaultRegistry` with manual mutex
- **Problem**: Tests explicitly declare `// NOT parallel` and use manual mutex to serialize access to global DefaultRegistry. This slows the test suite and the mutex pattern is fragile — forgetting the lock causes data races at test time.
- **Fix**:
  1. Create `NewModel(registry *adapters.Registry)` constructor (dependency injection)
  2. Each test creates its own Registry, adds mock adapters, passes to NewModel
  3. Remove `registryMu` and `// NOT parallel` comments
  4. Enable `t.Parallel()` on all model tests
- **Verification**: `go test -race -count=1 ./internal/app/` passes with `t.Parallel()` enabled. No mutex needed.
- **References**: Go testing: prefer dependency injection over global state
