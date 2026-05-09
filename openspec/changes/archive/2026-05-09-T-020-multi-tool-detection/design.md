# Design: Multi-tool Detection (T-020)

## Technical Approach

Wire `sequoia status` to surface `AdapterStatus` fields currently ignored. Each adapter's `Install()` writes a `.sequoia-version` file into its skills directory; `Status()` reads it back. `EvalSymlinks` in `base()` resolves symlinked home directories. A new `ScanTools()` function in `cmd/sequoia/` provides structured status for all registered adapters, shared by CLI output and future TUI.

## Architecture Decisions

| Decision | Options | Tradeoff | Choice |
|----------|---------|----------|--------|
| Version file location | (A) Per-adapter skills dir, (B) Common `.sequoia/` dir, (C) Adapter base dir | A: cleaned by existing Uninstall, one file per adapter. B: single global file, but must track multiple adapters. C: pollutes tool config root. | **A** — `~/.claude/skills/sequoia/.sequoia-version` |
| Version file format | (A) Plain text, (B) YAML, (C) JSON | A: zero deps, a read is `os.ReadFile()` + `TrimSpace()`. B/C: overkill for a single string, add import overhead. | **A** — single-line version string |
| Missing version file | (A) Return `""`, (B) Error, (C) Log warning | A: backward-compatible with pre-T020 installs. B: breaks `status` on legacy installs. C: noise for expected case. | **A** — silent empty string |
| Path field source | (A) `SkillsPath()`, (B) `SystemPromptPath()`, (C) `base()` | A: where Sequoia files live, matches current impl. B: returns a file, not a "root path". C: tool config root, not Sequoia-specific. | **A** — `SkillsPath()` |
| `ScanTools()` location | (A) `cmd/sequoia/`, (B) `internal/detection/` | A: simple, avoids new package. B: cleaner separation but overkill for 5 lines. | **A** — `cmd/sequoia/` |
| Symlink resolution | (A) `base()`, (B) `Status()` | A: all path methods benefit (SkillsPath, CommandsPath, etc). B: only status benefits, other ops use stale paths. | **A** — `claudeBase()` / `opencodeBase()` |
| CLI columns | (A) ID/NAME/DETECTED/INSTALLED/VERSION/PATH, (B) NAME/PATH/INSTALLED/VERSION only | A: ID needed for `--tool` flag, DETECTED shows tool presence. B: cleaner but loses useful info. Table widths: 14/14/9/10/10/55. | **A** — 6 columns |

## Data Flow

```
Install() ──writes──→ skills/sequoia/.sequoia-version
                             │
Status() ──reads─────────────┘
    │                       │
    ├── Version ← file      │
    ├── Installed ← IsInstalled()
    └── Path ← SkillsPath() [EvalSymlinks resolved]
              │
ScanTools() ──collects──→ []AdapterStatus ──→ runStatus() ──→ table output
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `adapters/claude/paths.go` | Modify | Add `versionFilePath(base)`; wrap `claudeBase()` return with `filepath.EvalSymlinks()` |
| `adapters/claude/adapter.go` | Modify | `Status()` reads `.sequoia-version`; `Install()` writes it; `Uninstall()` removes it |
| `adapters/opencode/paths.go` | Modify | Add `versionFilePath(base)`; wrap `opencodeBase()` return with `filepath.EvalSymlinks()` |
| `adapters/opencode/adapter.go` | Modify | Mirror of claude changes |
| `cmd/sequoia/main.go` | Modify | `runStatus` uses `a.Status()`; add PATH/VERSION columns; new `ScanTools()` |
| `cmd/sequoia/main_test.go` | Modify | Tests for new columns, `ScanTools()`, version file round-trip, symlink resolution |
| `adapters/claude/adapter_test.go` | Modify | Tests for version file read/write, Status() populates Version |
| `adapters/opencode/adapter_test.go` | Modify | Mirror of claude tests |

## Interfaces / Contracts

`ScanTools()` is the only new exported symbol:
```go
// ScanTools returns structured installation status for all registered adapters.
// Each result includes name, path, installed state, and Sequoia version.
// The returned slice includes adapters even if not detected — callers filter.
func ScanTools() []adapters.AdapterStatus
```

`versionFilePath` is unexported per adapter:
```go
// versionFilePath returns the full path to the .sequoia-version marker.
func versionFilePath(base string) string {
    return filepath.Join(skillsPath(base), ".sequoia-version")
}
```

`AdapterStatus` (no changes — existing struct is sufficient):
```go
type AdapterStatus struct {
    Installed bool
    Version   string  // ← now populated from .sequoia-version
    Path      string  // ← now EvalSymlinks-resolved
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `versionFilePath()` | Table-driven: verify suffix `.skills/sequoia/.sequoia-version` |
| Unit | `Status()` reads version | Write `.sequoia-version` in temp skills dir, assert Version field |
| Unit | `Status()` missing file | No version file → Version empty string, no error |
| Unit | `EvalSymlinks` error fallback | Mock: create dir, assert base() returns path without panic |
| Unit | `ScanTools()` | Register mock adapter, assert returned slice has expected count/fields |
| Integration | `runStatus` output format | Capture stdout, assert header and column alignment |
| Integration | Version round-trip | `Install()` → `Status()` → assert Version matches `Version` const |
| Integration | Symlinked home | `os.Symlink()` temp dir → `EvalSymlinks` → resolved path in Status |

All tests use `t.TempDir()`, never mutate real `~/.claude/` or `~/.config/opencode/`. External test packages (`claude_test`, `opencode_test`) follow existing convention.

## Open Questions

None — all design decisions resolved.
