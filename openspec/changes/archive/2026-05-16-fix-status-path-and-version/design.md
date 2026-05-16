# Design: Fix Status Path Display and Welcome Version

## Technical Approach

Two independent changes in the same fileset: (1) remove the path column from the TUI status
`renderStatusRow()` format string — reducing visual noise without touching the headless CLI
PATH column; (2) extract `resolveVersion()` from `newVersionCmd` so `runTUI()` can pass the
resolved build version to the Welcome screen instead of the raw `"0.1.0-dev"` fallback.

## Architecture Decisions

| # | Option | Tradeoff | Decision |
|---|--------|----------|----------|
| 1 | Remove `Path` from adapter/model types vs. drop only from TUI render | Removing the field breaks adapter tests that validate `Status().Path` and `ScanTools` expectations | **Keep `Path`** in `ToolStatus` and `AdapterStatus`; remove only from `renderStatusRow()` format |
| 2 | Unify `cmd/sequoia.Version` with `common.Version` vs. keep separate | `common.Version` is the *content* version written to disk during install; `cmd/sequoia.Version` is the CLI binary version. They serve different lifecycles | **Keep separate** |
| 3 | Inline resolution in `runTUI()` vs. extract `resolveVersion()` | Inlining duplicates 12 lines; extraction gives one source of truth | **Extract** `resolveVersion(raw string) string` in `cmd/sequoia/main.go` |
| 4 | Remove PATH from headless `runStatus` vs. preserve | Headless serves scripting/CI users who benefit from full install paths | **Preserve** headless PATH column |
| 5 | Create a `version` package vs. keep in `main.go` | New package is overkill for an unexported 15-line helper used only in two call sites within the same file | **Keep in main.go** |

## Data Flow

```
                    resolveVersion(Version)
                   /                      \
     newVersionCmd()                      runTUI()
     (prints resolved)                    (passes to app.NewModel)
                                               |
                                          WelcomeView
                                          (renders whatever it receives)
```

TUI status path removal:
```
     Adapter.Status() → ToolStatus{Path: kept}
              |
     renderStatusRow() → format: prefix + marker + name + version  (4 fields, path dropped)
              |
     StatusView() → renders to terminal
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/tui/screens/status.go` | Modify | `renderStatusRow()`: remove `path` variable (lines 76-79) and the path `%s` from Sprintf (line 81-87), changing 5-field to 4-field format. `StatusView()` Godoc comment (line 15): drop "and path". |
| `internal/tui/screens/status_test.go` | Modify | Rename `TestStatusView_ShowsVersionAndPath` → `TestStatusView_ShowsVersion`; remove path assertion (line 51). |
| `internal/tui/screens/testdata/golden/status_all_installed.txt` | Modify | Remove `/home/user/.claude` (line 3) and `/home/user/.config/opencode` (line 4) path strings from golden output. |
| `internal/tui/screens/testdata/golden/status_mixed.txt` | Modify | Remove `/home/user/.claude` path (line 3); replace `—  —` with single `—` (line 4, OpenCode not-installed row). |
| `cmd/sequoia/main.go` | Modify | Add `resolveVersion(raw string) string` (lines after `newVersionCmd`). Simplify `newVersionCmd()` RunE (lines 209-224) to `resolveVersion(Version)`. Update `runTUI()` line 404: `Version` → `resolveVersion(Version)`. |

## Interfaces / Contracts

No new exported interfaces. The new `resolveVersion` is unexported:

```go
// resolveVersion resolves the CLI version. If raw is "0.1.0-dev",
// it attempts to read the real version from debug.ReadBuildInfo().
// Otherwise, it returns raw unchanged.
func resolveVersion(raw string) string
```

`renderStatusRow` signature unchanged — still receives `model.ToolState` and `bool`.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `resolveVersion` edge cases | Table-driven: `"0.1.0"` (pass-through), `"0.1.0-dev"` (resolves via build info), `""` (pass-through). Since `debug.ReadBuildInfo()` depends on build context, tests validate the function compiles and returns non-empty for dev builds. |
| Unit | TUI status no longer shows path | Rename `TestStatusView_ShowsVersionAndPath` → `TestStatusView_ShowsVersion`; assert `v0.1.0` present, path absent. |
| Golden | Status screen output matches new format | Run `UPDATE_GOLDEN=1 go test ./internal/tui/screens/` after code changes to regenerate `status_all_installed.txt` and `status_mixed.txt`. |
| Integration | `runTUI` receives resolved version | Existing `TestVersionCmd_DevVersionResolves` already validates resolution logic. New test verifies `resolveVersion("0.1.0-dev")` returns non-empty, non-"(devel)" string. |

## Migration / Rollout

No migration required. No data persisted, no schema changes. Rollback: `git revert`.

## Open Questions

None — all design decisions confirmed by proposal and codebase review.
