# Design: Fix Uninstall Orphaned Files

## Technical Approach

Replace three `xxxBase(a.HomeDir())` calls with `a.base()` in Gemini and Codex adapters, and add an explicit branch for clean uninstalls in `CompleteView`. The `base()` method already handles production home-dir resolution (`os.UserHomeDir()` + `sync.Once` cache + symlink resolution) — the same method correctly used by `IsInstalled()`, `SkillsPath()`, and OpenCode adapters. No new abstractions; just wiring the adapters to the already-correct code path.

## Architecture Decisions

| Decision | Options | Chosen | Rationale |
|----------|---------|--------|-----------|
| Path resolution fix | a) `a.base()` b) fix `geminiBase`/`codexBase` to internally resolve `os.UserHomeDir()` | **a) `a.base()`** | `base()` already exists, caches, handles symlinks, and is the canonical method used by `IsInstalled()`/`SkillsPath()`/`Install()`. Option (b) duplicates logic and bypasses symlink resolution. |
| TUI conditional structure | a) `if/else if/else` b) nested `if` inside existing branches | **a) `if/else if/else`** | Reads linearly: warnings → clean uninstall → install. One level of nesting, mirrors the heading logic already above (lines 25-34). |
| Production-path test approach | a) `t.Setenv("HOME", tmp)` with `NewAdapter("")` b) `os.Setenv` + cleanup c) integration-only | **a) `t.Setenv` + `NewAdapter("")`** | Per-adapter `sync.Once` ensures test isolation. Go 1.20+ `t.Setenv` is parallel-safe. No global state mutation. On Windows use `USERPROFILE`; test helper abstracts platform. |

## Data Flow

**Before (Bug)**:
```
a.HomeDir() → "" (production) or tempDir (tests)
    ↓
geminiBase("") → filepath.Join("", ".gemini") → ".gemini"  ← RELATIVE to CWD
    ↓
Uninstall operates on ./gemini/sequoia/ → WRONG directory → orphaned files
```

**After (Fix)**:
```
a.base()
    ├─ homeDir=""? → os.UserHomeDir() → /home/user  (cached via sync.Once)
    └─ homeDir=tmp? → tmp                              (test path)
         ↓
ResolveSymlink(homeDir) → resolved home  (+ warning if symlinks)
         ↓
a.resolveBase(resolved) → geminiBase(resolved) → /home/user/.gemini
         ↓
Uninstall operates on /home/user/.gemini/sequoia/ → CORRECT directory
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `adapters/gemini/adapter.go:77` | Modify | `geminiBase(a.HomeDir())` → `a.base()` |
| `adapters/codex/adapter.go:80` | Modify | `codexBase(a.HomeDir())` → `a.base()` (Install) |
| `adapters/codex/adapter.go:175` | Modify | `codexBase(a.HomeDir())` → `a.base()` (Uninstall) |
| `internal/tui/screens/complete.go:73` | Modify | Add `else if mode == "uninstall"` branch for clean uninstalls |
| `internal/i18n/keys.go` | Modify | Add `MsgCompleteUninstalledItems` constant |
| `internal/i18n/translations/en.toml` | Modify | Add `complete.uninstalled_items` key |
| `internal/i18n/translations/es.toml` | Modify | Add `complete.uninstalled_items` key |
| `internal/i18n/keys_test.go` | Modify | Add new key to `allKeys`, bump expected count (65→66) |
| `adapters/gemini/adapter_test.go` | Modify | Add table-driven test with `homeDir=""` path |
| `adapters/codex/adapter_test.go` | Modify | Add table-driven test with `homeDir=""` path |
| `internal/tui/screens/complete_test.go` | Modify | Add golden test for `uninstall+warnedCount=0`, bump `allSucceed` golden |
| `internal/tui/screens/testdata/golden/` | Create/Update | New golden: `complete_uninstall_clean.txt`, update `complete_all_succeed.txt` |

## Interfaces / Contracts

No new interfaces. The only new exported constant:

```go
// keys.go — Complete screen section
MsgCompleteUninstalledItems = "complete.uninstalled_items"
```

Translation values:
- `en`: `"Uninstalled: Skills, Commands, System Prompt"`
- `es`: `"Desinstalado: Skills, Comandos, System Prompt"`

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit (adapter) | Production path: `homeDir=""` with controlled `USERPROFILE`/`HOME` | Table-driven: each case sets env, creates adapter with `NewAdapter("")`, installs, uninstalls, asserts sequoia dir removed. `t.Setenv` per subtest for isolation. |
| Unit (adapter) | Regression: existing `SetHomeDir(t.TempDir())` tests | No changes needed — existing tests pass as-is and continue to cover the test path. |
| Unit (TUI) | `CompleteView` with `mode="uninstall", warnedCount=0` | Assert rendered output contains `"Uninstalled"` text (not `"Installed"`). Golden file for the new variant. |
| Golden (TUI) | Uninstall clean variant | New golden file `complete_uninstall_clean.txt` generated via `UPDATE_GOLDEN=1`. |

**Production-path test helper pattern**:
```go
func testAdapterProdPath(t *testing.T, newFn func(string) Adapter) {
    t.Helper()
    tmp := t.TempDir()
    t.Setenv("USERPROFILE", tmp)  // Windows; "HOME" for Unix
    a := newFn("")                // empty → uses os.UserHomeDir() → tmp
    // ... install, uninstall, assert
}
```

## Symlink Considerations

`a.base()` calls `ResolveSymlink(homeDir)` before passing to `resolveBase`. When Gemini/Codex switch from raw `xxxBase(a.HomeDir())` to `a.base()`:

- **If `~/.gemini` is a symlink**: Uninstall now targets the RESOLVED path (same as `IsInstalled()` and `Install()` already do via `base()`). Consistency gain — Install and Uninstall operate on the same filesystem location.
- **If `$HOME` itself is a symlink** (e.g., macOS `/var/...` → `/Users/...`): `ResolveSymlink` resolves it, same behavior across all adapters. Was already the case for `Install()` via BaseAdapter.
- **No migration needed**: The bug meant uninstall NEVER targeted the correct home directory anyway. First correct uninstall is inherently clean.

## Migration / Rollout

No migration required. This is a pure bugfix — affected users have orphaned files at `./.gemini/sequoia/` and `./.codex/sequoia/` relative to wherever they ran `sequoia uninstall`. These orphaned dirs are harmless (will be recreated on next install) and are manually removable.

## Rollback Plan

Revert three adapter changes:
1. `adapters/gemini/adapter.go:77`: `a.base()` → `geminiBase(a.HomeDir())`
2. `adapters/codex/adapter.go:80`: `a.base()` → `codexBase(a.HomeDir())`
3. `adapters/codex/adapter.go:175`: `a.base()` → `codexBase(a.HomeDir())`

The i18n keys, TUI conditional, and tests are additive — they don't need rollback and are harmless (the new key is only used from the new branch).

## Open Questions

None — all technical decisions resolved.
