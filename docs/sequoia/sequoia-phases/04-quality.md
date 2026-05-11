# P4 Quality — sequoia-ai v0.1.0

**Score**: 79/100 (B) | **Findings**: 15 (0C, 2H, 6M, 5L, 2I)

---

## Quality Metrics

| Metric | Value | Assessment |
|--------|-------|-----------|
| Test files | 64 (841 test functions) | ✅ Strong test culture |
| Test style | Behavioral (real filesystem I/O) | ✅ Commendable — see P4-015 |
| Coverage gaps | 5 untested error/paths | ⚠️ Concentrated in `cmd/sequoia/` |
| Error types | 1 sentinel error total | ⚠️ Immature error model |
| Go idioms | Go 1.24 but uses `interface{}` | ⚠️ Pre-1.18 style |
| File naming | 2 files with hyphens | ⚠️ Inconsistent with Go conventions |
| Linting | golangci-lint v2 (10 linters) | ✅ Active and configured |

---

## Findings

### P4-001 [HIGH] — Go Idioms: interface{} used instead of `any` (20+ occurrences)
**Evidence**: `adapters/common/template.go:13`, `adapters/codex/toml_merge.go`, `release_test.go:42`, etc.

Go 1.18 (2022) introduced `any` as an alias for `interface{}`. The project uses Go 1.24.2 but retains the legacy spelling in 20+ locations. The Go team recommends `any` for new code.

**Impact**: Reduces readability for developers trained on Go 1.18+ idioms. Signals outdated code style.

**Remediation**: Global find-replace: `interface{}` → `any`.

---

### P4-002 [HIGH] — Quality: 5 adapter Install() methods nearly identical (~250 duplicated lines)
**Evidence**: All `adapters/*/adapter.go` Install() methods

The install workflow is copy-pasted across 5 adapters with ~80% identical code. Steps 1-9 (staging, template render, stage, MkdirAll, skill installer, cmd installer, version file) are identical. Only system prompt injection differs. This is the root cause of P4-007 (inconsistent version file newline).

**Impact**: Bug fixes must be replicated 5 times manually. High risk of subtle divergence.

**Remediation**: Extract `common.InstallSkills()` — already captured in RC1.

---

### P4-003 [MEDIUM] — Naming: Duplicate sequoia marker constants in opencode package
**Evidence**: `adapters/opencode/adapter.go:15` vs `adapters/opencode/installer.go:10-11`

Two constants hold the identical string `"<!-- sequoia:start -->"` in the same package: `sequoiaMarker` (adapter.go) and `markerStart` (installer.go). Different functions use different constants. Update risk: one constant gets changed, the other doesn't.

**Impact**: IsInstalled() and isSequoiaManaged() could diverge silently if the marker format changes.

---

### P4-004 [MEDIUM] — Coverage Gap: isTerminal() has no direct test
**Evidence**: `cmd/sequoia/terminal.go:1-14`

No `_test.go` file for `terminal.go`. The `os.Stdin.Stat()` error path is never exercised. All CLI tests mock `isTerminalFn` to bypass it, leaving the actual function untested.

**Impact**: If `isTerminal()` regresses on certain platforms, CI won't catch it.

---

### P4-005 [MEDIUM] — Coverage Gap: runUninstall tests only cover confirmation gate
**Evidence**: `cmd/sequoia/main_test.go:318-328`

7 test functions exercise `runUninstall` but ALL test confirmation prompt logic only. None verify that `a.Uninstall()` is actually called or that files are removed. The yes-flag test calls `runUninstall("claude-code", false, true, ...)` which enters the loop but only verifies no "?" in output.

**Impact**: Regression in the uninstall loop (lines 326-336) would not be caught by existing tests.

---

### P4-006 [MEDIUM] — Error Handling: fmt.Fscanln return values silently discarded
**Evidence**: `cmd/sequoia/main.go:319`

Both count and error from `fmt.Fscanln` are discarded with `_, _`. If stdin is closed or broken, the error is lost and `response` remains `""`, which is indistinguishable from user pressing Enter. An I/O error is silently conflated with user cancellation.

**Impact**: Broken stdin pipe causes silent abort instead of surfacing the error.

**Remediation**: Check `n > 0` and handle `err != nil` separately.

---

### P4-007 [MEDIUM] — Naming: Inconsistent version file trailing newline across adapters
**Evidence**: `adapters/cursor/adapter.go:197` vs all others

4 of 5 adapters write `common.Version + "\n"`. Cursor writes only `common.Version` without trailing newline. Currently masked by `strings.TrimSpace()` in the reader, but fragile.

**Impact**: If a future reader expects the newline (e.g., `bufio.Scanner`), cursor installations would break.

---

### P4-008 [MEDIUM] — Coverage Gap: isSequoiaManaged() error path untested
**Evidence**: `adapters/opencode/installer.go:84-89`

`isSequoiaManaged()` is called by both GenerateAgentsMD and RemoveAgentsMD, but both catch `IsNotExist` before calling it. The permission-denied error path from `os.ReadFile` is never exercised.

**Impact**: Permission errors on filesystems with ACLs would propagate untested.

---

### P4-009 [LOW] — Error Handling: All 5 adapter Uninstall() methods discard os.Remove errors
**Evidence**: `adapters/claude/adapter.go:215-218` and all others

Best-effort removal is intentional, but ALL error types are discarded — including permission-denied, filesystem-full, and directory-not-empty. User gets "Done" even when files couldn't be removed.

**Impact**: Silent uninstall failures. No diagnostic information for debugging.

---

### P4-010 [LOW] — Go Idioms: Hyphens in Go file names (2 files)
**Evidence**: `internal/tui/screens/install-progress.go`, `internal/tui/screens/tool-selection.go`

Go convention specifies underscores, not hyphens, in file names. Their test files correctly use underscores (`install_progress_test.go`), creating inconsistency even within the same screen.

**Remediation**: Rename to `install_progress.go` and `tool_selection.go`.

---

### P4-011 [LOW] — Test Quality: Fields accessed without assertions in ScanTools tests
**Evidence**: `cmd/sequoia/main_test.go:249-251,463`

Tests access `r.Version` and `r.Installed` only to suppress the Go compiler's "declared and not used" error. No assertions made about their values. Tests give false sense of coverage.

**Impact**: If ScanTools() stops populating these fields, no test fails.

---

### P4-012 [LOW] — Error Handling: Only 1 sentinel error for entire adapter system
**Evidence**: `adapters/errors.go:1-7`

The entire adapter error model is `ErrUnknownAdapter`. Install failures, uninstall failures, path resolution errors — all use opaque `fmt.Errorf` wrappers. CLI cannot programmatically distinguish failure modes.

**Impact**: All errors surface identically to the user. No way to offer recovery paths.

---

### P4-013 [LOW] — Coverage Gap: Adapter paths.go functions untested directly
**Evidence**: All `adapters/*/paths.go:12-26`

The `xxxBase()` functions are exercised indirectly through adapter methods but their `os.UserHomeDir()` error path is never tested directly. Edge cases (empty string after resolution) are untested.

**Impact**: If `os.UserHomeDir()` changes behavior in a future Go version, no test catches the regression.

---

### P4-014 [INFO] — Documentation: ScanTools exported but unreachable from outside package main
**Evidence**: `cmd/sequoia/main.go:270-280`

`ScanTools` is exported (capital S) but Go does not allow importing `main` packages. It's only used in 2 test functions, never in production code. The exported name suggests public API intent that's impossible to fulfill.

**Remediation**: Unexport to `scanTools`.

---

### P4-015 [INFO] — Commendation: Test quality is predominantly behavioral ✅
**Evidence**: `adapters/common/installer_test.go:40-231`, all adapter tests

Sequoia's tests create real temporary directories, invoke the full lifecycle, and verify file-system outcomes. They test behavior (file contents, file existence) rather than implementation (mock calls, spy assertions). Error paths are tested for the installer lifecycle and section injection/removal. This is a commendable pattern that provides high confidence and resilience to refactoring.

**Impact**: Positive finding. This quality should be preserved as the codebase evolves.
