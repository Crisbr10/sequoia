# Sequoia Action Plan — sequoia-ai v0.1.0

> Generated from `/sequoia audit` on 2026-05-11
> Health Score: 🟢 78/100 (B)

---

## Immediate (Critical + High)

### 🔴 RC1: Extract shared adapter boilerplate into `adapters/common/`
**Effort**: 4–6h | **Fixes**: P3-001, P3-002, P3-004, P2-001, P2-002, P4-002, P4-003, P4-007

- [ ] **TASK-001**: Create `adapters/common/base.go` — `BaseResolver(homeDir, relativeDir string)` to eliminate 5 identical `xxxBase()` functions
- [ ] **TASK-002**: Create `adapters/common/adapter_installer.go` — `InstallSkills(cfg AdapterInstallConfig)` helper that encapsulates staging, template rendering, skill install, command install, and version file writing. Accepts a `SystemPromptInjector` callback for the adapter-specific step
- [ ] **TASK-003**: Move `templateData` struct to `adapters/common/` as `TemplateData`
- [ ] **TASK-004**: Centralize command template embedding in `adapters/common/embed.go` using `//go:embed templates/commands`; expose via `common.CommandTemplates()`
- [ ] **TASK-005**: Move `InjectSection`/`RemoveSection` from `claude/installer.go` and `gemini/installer.go` to `adapters/common/sections.go`
- [ ] **TASK-006**: Add `"Version\n"` trailing newline to `cursor/adapter.go:197` (consistency fix)
- [ ] **TASK-007**: Refactor all 5 adapters to use the new common helpers
- [ ] **TASK-008**: Remove duplicated `embed.go` files from each adapter package
- [ ] **TASK-009**: Remove duplicated template directories from each adapter
- [ ] **TASK-010**: Update `_template/adapter.go` to reflect the reduced boilerplate
- [ ] **TASK-011**: Run full test suite across all adapters — verify no regressions

### 🔴 RC3: Make ToolAdapter interface evolvable
**Effort**: 3–5h | **Fixes**: P3-005, P3-006, P3-007

- [ ] **TASK-012**: Split `ToolAdapter` into smaller interfaces: `Installable`, `Detectable`, `Statusable` — compose them in the Registry where needed
- [ ] **TASK-013**: Remove `SkillsPath()`, `CommandsPath()`, `SystemPromptPath()` from the interface; make them internal to each adapter
- [ ] **TASK-014**: Replace `PromptStrategy()` int enum with a descriptive `StrategyDescription() string` method
- [ ] **TASK-015**: Update `registry_test.go` mockAdapter and all test mocks
- [ ] **TASK-016**: Run full test suite — verify all 12+ files compile correctly

### P4-002: Refactor duplicated Install() methods
**Effort**: Included in RC1 above | **Fixes**: P4-002 (duplicate Install methods)

---

## Short-term (Medium)

### 🟡 RC2: Structured error types for install lifecycle
**Effort**: 2–3h | **Fixes**: P3-010, P4-012

- [ ] **TASK-017**: Define error types in `adapters/common/errors.go`: `ErrPrepareWriteAccess`, `ErrApplySourceMissing`, `ErrVerifyFileNotFound`
- [ ] **TASK-018**: Wrap installer phase errors with typed errors using `fmt.Errorf("...: %w", ErrPrepareWriteAccess)`
- [ ] **TASK-019**: Update TUI screens to inspect `errors.Is()` and show phase-specific recovery options

### 🟡 RC5: Strengthen supply chain verification
**Effort**: 2–3h | **Fixes**: P1-001, P1-002, P1-003

- [ ] **TASK-020**: Add `-SkipChecksum` warning to install script documentation — make checksum non-optional unless user explicitly opts out
- [ ] **TASK-021**: In `action.yml`, add SHA-256 verification step for versioned downloads
- [ ] **TASK-022**: Document pipe-to-shell risks in README and getting-started guide; provide alternative download-verify-run path

### 🟡 RC4: Fill test coverage gaps in `cmd/sequoia/`
**Effort**: 2–3h | **Fixes**: P4-004, P4-005, P4-008, P4-011, P4-013

- [ ] **TASK-023**: Add test for `isTerminal()` — test the `os.Stdin.Stat()` error path
- [ ] **TASK-024**: Add integration test for `runUninstall` that verifies `a.Uninstall()` is actually called
- [ ] **TASK-025**: Add test for `isSequoiaManaged()` permission-denied error path
- [ ] **TASK-026**: Add proper assertions for `Version` and `Installed` fields in ScanTools tests
- [ ] **TASK-027**: Add direct tests for adapter `xxxBase()` error paths (UserHomeDir failure)

### P4-006: Handle Fscanln errors in uninstall confirmation
**Effort**: <30min | **Fixes**: P4-006

- [ ] **TASK-028**: In `cmd/sequoia/main.go:319`, check `n > 0` and handle Fscanln error separately

### P4-009: Log uninstall errors instead of silently discarding
**Effort**: <1h | **Fixes**: P4-009

- [ ] **TASK-029**: In all 5 adapter Uninstall() methods, accumulate `os.Remove` errors and return them if any files couldn't be removed

### P3-008: Decompose `cmd/sequoia/main.go`
**Effort**: 2–3h | **Fixes**: P3-008

- [ ] **TASK-030**: Extract `install.go` with `runInstall` and `newInstallCmd`
- [ ] **TASK-031**: Extract `uninstall.go` with `runUninstall` and `newUninstallCmd`
- [ ] **TASK-032**: Extract `status.go` with `runStatus` and `newStatusCmd`
- [ ] **TASK-033**: Extract `tui.go` with `runTUI`
- [ ] **TASK-034**: Keep `main.go` with only `main()`, `newRootCmd()`, `init()`, `targetAdapters()`

### P2-003: Cache Lipgloss styles at package level
**Effort**: <1h | **Fixes**: P2-003

- [ ] **TASK-035**: Convert style functions in `internal/tui/styles/styles.go` to package-level `var` declarations (e.g., `var TitleStyle = lipgloss.NewStyle().Bold(true)...`)

### P3-012: Fix Cursor adapter Detect() to use direct directory check
**Effort**: <30min | **Fixes**: P3-012

- [ ] **TASK-036**: Replace `filepath.Join(base, "..")` with a dedicated `cursorRoot()` or check the actual cursor directory

### P4-010: Fix Go file naming (hyphens to underscores)
**Effort**: <30min | **Fixes**: P4-010

- [ ] **TASK-037**: Rename `internal/tui/screens/install-progress.go` → `install_progress.go`
- [ ] **TASK-038**: Rename `internal/tui/screens/tool-selection.go` → `tool_selection.go`
- [ ] **TASK-039**: Update all imports and references (if any cross-file references exist)

### P4-014: Unexport ScanTools
**Effort**: <15min | **Fixes**: P4-014

- [ ] **TASK-040**: Rename `ScanTools` → `scanTools` in `cmd/sequoia/main.go` and its tests

---

## Long-term (Low + Info)

### 🟢 RC6: Modernize Go idioms
**Effort**: <1h | **Fixes**: P4-001, P4-010

- [ ] **TASK-041**: Replace all `interface{}` with `any` in 20+ locations
- [ ] **TASK-042**: Already covered by TASK-037/038 above

### P2-004/005/006: Micro-optimizations (string concat, byte conversion, template caching)
**Effort**: 1–2h | **Fixes**: P2-004, P2-005, P2-006

- [ ] **TASK-043**: Use `strings.Builder` in InjectSection string concatenation
- [ ] **TASK-044**: Work with `[]byte` directly in InjectSection (eliminate string round-trips)
- [ ] **TASK-045**: Add `sync.Map` for template caching in `RenderTemplate`

### P2-007/008: Build tag separation (large effort, optional)
**Effort**: >8h (optional — only if binary size becomes a concern)

- [ ] **TASK-046**: [OPTIONAL] Move TUI to `//go:build tui` tag for headless-only builds
- [ ] **TASK-047**: [OPTIONAL] Move Codex adapter to `//go:build codex` tag to exclude BurntSushi/toml

### P3-011: Log warning on duplicate adapter registration
**Effort**: <30min | **Fixes**: P3-011

- [ ] **TASK-048**: Add `log.Printf("adapter %q replaced existing registration", id)` in `Registry.Register()`

### P3-009: Decouple `internal/model` from `adapters`
**Effort**: <1h | **Fixes**: P3-009

- [ ] **TASK-049**: Define a local `AdapterInfo` interface in `internal/model/` and use it instead of `adapters.ToolAdapter`

---

## Summary

| Priority | Tasks | Estimated Effort |
|----------|-------|-----------------|
| 🔴 Immediate | 16 tasks (RC1 + RC3) | 7–11h |
| 🟡 Short-term | 24 tasks | 8–12h |
| 🟢 Long-term | 9 tasks | 3–5h |
| **Total** | **49 tasks** | **18–28h** |

### Quick Wins (<2h each, high impact)
- TASK-035: Cache Lipgloss styles (30 min)
- TASK-028: Handle Fscanln errors (15 min)
- TASK-036: Fix Cursor adapter Detect() (30 min)
- TASK-037-039: Fix file naming (30 min)
- TASK-040: Unexport ScanTools (15 min)
- TASK-006: Fix cursor version newline (5 min)
