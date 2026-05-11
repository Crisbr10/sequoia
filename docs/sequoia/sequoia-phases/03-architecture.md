# P3 Architecture — sequoia-ai v0.1.0

**Score**: 65/100 (C+) | **Findings**: 13 (0C, 4H, 7M, 2L)

---

## Architecture Overview

```
cmd/sequoia/         → CLI entry (Cobra commands)
  ↓ uses
adapters/             → ToolAdapter interface + Registry + 6 implementations
  ├── common/         → shared installer pipeline (Prepare→Apply→Verify→Rollback)
  ├── claude/         → adapter para Claude Code (~/.claude/)
  ├── opencode/       → adapter para OpenCode (~/.config/opencode/)
  ├── cursor/         → adapter para Cursor (~/.cursor/rules/)
  ├── gemini/         → adapter para Gemini CLI (~/.gemini/)
  ├── codex/          → adapter para OpenAI Codex (~/.codex/)
  └── _template/      → template para nuevos adapters
internal/             → TUI app, model, pipeline, screens
plugin/               → plugin system (separate concern — cleanly isolated)
```

**Dependency direction**: Clean — no circular dependencies. `cmd → adapters + internal`, `internal → adapters`, `adapters → common`. ✅

---

## Findings

### P3-001 [HIGH] — Duplication: 507+ duplicated lines across all 6 adapters
**Evidence**: All `adapters/*/adapter.go` files

All 6 adapters share 70-80% identical code in their `Install()` methods. The shared install workflow (staging → template render → stage commands → install skills → install commands → write version) is copy-pasted with only the system prompt step varying. This violates DRY at scale — every enhancement must be replicated across 6 files.

**Impact**: High maintenance burden. Risk of drift (one adapter missing a fix). Adding a 7th adapter copies all boilerplate again.

**Remediation**: Extract `common.InstallSkills(cfg)` with a system-prompt callback parameter.

---

### P3-002 [HIGH] — Duplication: All xxxBase() functions share identical symlink resolution
**Evidence**: `adapters/*/paths.go:12-26` (all 5 adapters)

The homeDir resolution + `filepath.EvalSymlinks` fallback logic is copy-pasted 5 times. Only the final `filepath.Join(resolved, ".toolname")` differs. If `os.UserHomeDir()` ever needs caching or symlink resolution changes, all 5 files must be updated identically.

**Impact**: Any change to home directory resolution requires modifying 5-6 files. Risk of behavioral divergence.

**Remediation**: Extract `common.BaseResolver(homeDir, relativeDir string)`.

---

### P3-004 [HIGH] — Duplication: 25 command template files byte-identical across adapters
**Evidence**: All `adapters/*/templates/commands/` directories

Five command files × 5 adapters = 25 identical copies. Each adapter embeds all 5 files via `//go:embed templates`. Binary contains 5 copies of identical content. Editing a command file requires updating 5 template directories.

**Impact**: Binary bloat + maintenance burden. One directory inevitably gets missed on updates.

**Remediation**: Centralize command templates in `adapters/common/embed.go`.

---

### P3-007 [HIGH] — Breaking Change: Adding any interface method breaks 12+ files
**Evidence**: `adapters/interface.go:36-58`, `adapters/registry_test.go:13-35`

Go interfaces have no default methods. Adding a single method to `ToolAdapter` breaks: 6 concrete adapters, 1 mockAdapter in tests, the `_template` adapter, and all adapter-specific test mocks. Total blast radius: ~12+ files must be updated atomically.

**Impact**: Discourages interface evolution. Future improvements to the adapter contract require coordinated multi-file changes.

**Remediation**: Split into smaller interfaces (`Installable`, `Detectable`, `Statusable`) and compose in registry.

---

### P3-003 [MEDIUM] — Duplication: 5 identical install.go templateData structs
**Evidence**: `adapters/*/install.go:5-7` (4 of 5 adapters are identical)

Four adapters define an identical `templateData` struct with just a `Version string` field. Only Codex adds extra fields. This type could live once in `adapters/common/`.

**Impact**: Adding a new template variable requires editing 5 files instead of 1.

---

### P3-005 [MEDIUM] — API Design: Path query methods leak implementation details
**Evidence**: `adapters/interface.go:51-57`

`SkillsPath()`, `CommandsPath()`, `SystemPromptPath()` expose internal filesystem structure to consumers. Only used by `Status()` and TUI display. If an adapter changes its directory layout, consumers break. Violates Interface Segregation Principle.

**Impact**: Interface has 11 methods. Path changes become breaking changes. These methods should be internal to each adapter.

---

### P3-006 [MEDIUM] — API Design: PromptStrategy is a type-code enum, not encapsulated behavior
**Evidence**: `adapters/interface.go:5-20`

`PromptStrategy()` returns an int-based enum that callers must switch on. Currently no consumer uses it programmatically — it's purely informational. Should be a descriptive string or removed from the interface.

**Impact**: If consumers ever need to react to strategy, they'd need switch statements for each new strategy. Currently low impact.

---

### P3-008 [MEDIUM] — God Object: main.go mixes 12 concerns in 373 lines
**Evidence**: `cmd/sequoia/main.go:1-373`

Single file contains: version detection, command tree, install/status/uninstall/version commands, TUI launcher, status table formatter, adapter selector, confirmation flow handler. Each concern belongs in its own file following `cmd/` best practices.

**Impact**: Changes to unrelated features require editing the same file. Testing individual handlers is harder.

---

### P3-010 [MEDIUM] — Error Handling: Single sentinel error — no structured error types
**Evidence**: `adapters/errors.go:5-7`, `adapters/common/installer.go:65-110`

The installer's 4-phase lifecycle produces generic `fmt.Errorf` wrappers. No way to programmatically distinguish Prepare failure from Verify failure, or permission error from missing source. TUI screens can only display raw error strings.

**Impact**: TUI cannot offer phase-specific recovery options (retry? skip? abort?).

---

### P3-012 [MEDIUM] — Coupling: Cursor adapter uses `../` for Detect()
**Evidence**: `adapters/cursor/adapter.go:47`

While all other adapters check their base directory directly with `os.Stat(base)`, Cursor's Detect() resolves `filepath.Join(base, "..")`. This is fragile — if `cursorBase()` changes, the `..` navigation breaks silently.

**Impact**: Inconsistency between adapters. Path navigation is fragile to restructuring.

---

### P3-009 [LOW] — Coupling: internal/model imports adapters.ToolAdapter directly
**Evidence**: `internal/model/types.go:6,33`

The domain model depends on the adapter infrastructure layer. Mild Clean Architecture violation. Both are in the same module and compiled together — compile-time concern only.

**Impact**: Low. Pragmatic for Go. If ToolAdapter changes, internal/model breaks at compile time.

---

### P3-011 [LOW] — Plugin Architecture: Register() silently replaces duplicates
**Evidence**: `adapters/registry.go:25-40`

If two adapters accidentally share an ID, only the last registered survives with no warning. The test `TestRegistry_RegisterDuplicate_ReplacesExisting` confirms this is by design, but a log message would aid debugging.

**Impact**: Low — Go's init() ordering via blank import makes this deterministic. But zero feedback on accidental duplicate IDs.

---

### P3-013 [LOW] — Positive Finding: Plugin and Adapter packages cleanly isolated
**Evidence**: `plugin/interface.go` vs `adapters/interface.go`

`plugin/` and `adapters/` both define ID/Name patterns but serve completely different domains (audit agent extensions vs tool installation targets). Zero cross-imports. Clean separation confirmed.

**Impact**: None. This is a positive architectural quality to preserve.
