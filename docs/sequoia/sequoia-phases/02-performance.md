# P2 Performance — sequoia-ai v0.1.0

**Score**: 82/100 (B) | **Findings**: 9 (0C, 0H, 3M, 3L, 3I)

---

## Performance Budget (CLI Tool)

| Metric | Target | Assessment |
|--------|--------|-----------|
| Startup time | <500ms | ✅ Pass — init() registration is ~50μs |
| Memory baseline | <100MB | ✅ Pass — templates loaded are ~70KB |
| Binary size | ~1.5MB | ⚠️ ~75KB avoidable bloat from duplication |

---

## Findings

### P2-001 [MEDIUM] — Memory: Duplicate command templates across 5 adapters
**Evidence**: `adapters/claude/embed.go:6`, `adapters/opencode/embed.go:6`, etc.

The 5 command markdown files (~4.9KB each) are byte-identical but embedded independently in all 5 adapters via `//go:embed templates`. Results in ~25 duplicated files (~70KB) in the binary. These could live in `adapters/common/` once.

**Impact**: ~50-70KB binary bloat, ~5% of estimated binary size.

---

### P2-002 [MEDIUM] — Memory: Duplicate InjectSection/RemoveSection in claude + gemini
**Evidence**: `adapters/claude/installer.go:17-88`, `adapters/gemini/installer.go:17-88`

72 lines of byte-identical code duplicated between the two adapters. Both use identical marker-based section injection. Should be moved to `adapters/common/`.

**Impact**: ~5-10KB binary bloat; maintenance cost — one bug fix requires changing 2 files.

---

### P2-003 [MEDIUM] — Allocations: Lipgloss styles re-created every View() frame
**Evidence**: `internal/tui/styles/styles.go:17-22`

Every style function (`Title()`, `Subtitle()`, etc.) calls `lipgloss.NewStyle()` and builds a new immutable style object on each call. In the TUI render loop (60fps), this generates hundreds of allocations per second. Lipgloss styles should be cached as package-level variables.

**Impact**: ~500-900 allocs/second in TUI mode; unnecessary GC pressure in render loop.

---

### P2-004 [LOW] — Allocations: String concatenation with + in InjectSection
**Evidence**: `adapters/claude/installer.go:22`

`section := markerStart + "\n" + strings.TrimRight(...) + "\n" + markerEnd + "\n"` creates up to 6 intermediate strings. Use `strings.Builder`.

**Impact**: ~5-8 extra allocations per InjectSection call — negligible for config files <10KB.

---

### P2-005 [LOW] — Allocations: Unnecessary []byte ↔ string round-trips
**Evidence**: `adapters/claude/installer.go:24-40`

Pattern: `os.ReadFile` ([]byte) → `string(raw)` → process → `[]byte(result)` → `os.WriteFile`. Creates 2 allocs of file-size each. Work directly with `[]byte` using `bytes` package.

**Impact**: 2 extra allocations (~2-10KB) per injection — negligible.

---

### P2-006 [LOW] — Computation: No template parse caching in RenderTemplate
**Evidence**: `adapters/common/template.go:13-26`

Each call re-parses templates with `template.New(name).Parse()`. Templates are fixed — could cache in `sync.Map` after first parse. Cost is microsecond-level per call.

**Impact**: ~10-15 re-parses per install (~100-500μs total) — negligible for one-shot CLI use.

---

### P2-007 [INFO] — Dependencies: BurntSushi/toml only used by codex adapter
**Evidence**: `adapters/codex/adapter.go:10`, `go.mod`

BurntSushi/toml is only imported by the Codex adapter. Users without Codex CLI still pay the binary size cost (~100-200KB). Could be moved to a build tag.

**Impact**: ~100-200KB binary bloat for non-Codex users. Large effort to fix (>8h).

---

### P2-008 [INFO] — Dependencies: TUI stack has 14 indirect dependencies
**Evidence**: `go.mod:14-35`

Bubbletea + Lipgloss pull 14 indirect dependencies for terminal rendering. Inherent to any Bubbletea-based TUI. Headless mode (`--no-tui`) still pays the compilation cost.

**Impact**: ~500KB-1MB binary overhead. Large effort to separate via build tags (>8h).

---

### P2-009 [INFO] — Startup: Eager adapter registration via init()
**Evidence**: `cmd/sequoia/main.go:20-24`, each adapter's `init()`

5 adapters register at startup via `init()`. Pattern is correct (database/sql style). Cost is ~50μs for registration + ~70KB templates loaded. Acceptable for a CLI tool.

**Impact**: Negligible. This is idiomatic Go for plugin systems.
