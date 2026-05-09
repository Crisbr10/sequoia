# Design: Uninstall Command Safety (T-021)

## Technical Approach

Add `--yes`/`-y` flag to the `uninstall` subcommand and an interactive confirmation gate inside `runUninstall`. When `--yes` is absent and stdin is a terminal, prompt the user before removing files. When stdin is piped and `--yes` not set, return a clear error.

Resolve targets first, confirm once, then execute — no per-adapter prompts.

## Architecture Decisions

| # | Decision | Choice | Alternatives | Rationale |
|---|----------|--------|--------------|-----------|
| 1 | Signature change | Add `yes bool, in io.Reader` params | Config struct | Only 2 new params; config struct overkill for an internal function |
| 2 | Input reader | `fmt.Fscanln(in, &response)` | `bufio.Reader` | Matches proposal spec; sufficient for single-word "y"/"Y" input |
| 3 | Terminal check | Use existing `isTerminal()` via package var `isTerminalFn` | Pass `isTerm bool` param | Enables test override without changing call sites; follows install pattern |
| 4 | `--all` prompt | List each tool name before `[y/N]` | Generic "N tool(s)" message | Gives user full context about what will be removed |
| 5 | Abort behavior | `return nil` with "aborted" message (exit 0) | `return error` (exit 1) | User-initiated cancellation is not a failure; matches Unix conventions |

## Data Flow

```
sequoia uninstall --tool=X [--yes]
         │
         ▼
   newUninstallCmd() RunE
         │
         ├─ parse flags: toolID, all, yesFlag
         │
         ▼
   runUninstall(toolID, all, yesFlag, os.Stdin, out)
         │
         ├─ targetAdapters(toolID) ──→ targets []ToolAdapter
         ├─ filter installed, collect names
         │
         ├─ yesFlag == true? ──yes──→ skip prompt ──→ execute loop
         │
         ├─ isTerminalFn() == false? ──yes──→ error: "requires --yes"
         │
         ├─ prompt: "Remove Sequoia from {names}? [y/N]: "
         ├─ fmt.Fscanln(in, &response)
         ├─ response ∉ {"y","Y"}? ──yes──→ "aborted" (exit 0)
         │
         └─ execute loop: a.Uninstall() for each target
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `cmd/sequoia/main.go` | Modify | Add `yesFlag` to `newUninstallCmd()`; add `in io.Reader` param to `runUninstall()`; add confirmation gate; add `isTerminalFn` package var |
| `cmd/sequoia/main_test.go` | Modify | Add `TestUninstall*` table-driven cases: yes-flag, confirm-y, confirm-n, empty-input, piped-stdin-error, all-with-confirm, invalid-tool |

## Interfaces / Contracts

### `runUninstall` signature (after)

```go
// isTerminalFn wraps isTerminal for test override.
var isTerminalFn = isTerminal

func runUninstall(toolID string, all bool, yes bool, in io.Reader, out io.Writer) error
```

### Confirmation prompt formats

```
// Single tool
Remove Sequoia from Claude Code? [y/N]:

// Multiple tools (--all)
This will remove Sequoia from:
  Claude Code
  OpenCode
Continue? [y/N]:
```

### `newUninstallCmd` flag registration

```go
cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")
```

## Testing Strategy

All tests in `cmd/sequoia/main_test.go`, table-driven per project convention.

| # | Test | Approach | Type |
|---|------|----------|------|
| 1 | `--yes` bypasses prompt | `runUninstall("claude-code", false, true, nil, &out)` — expect "Done" | Unit |
| 2 | "y" confirms uninstall | `in=strings.NewReader("y\n")`, override `isTerminalFn=true` — expect "Removing" | Unit |
| 3 | "n" aborts (exit 0) | `in=strings.NewReader("n\n")` — expect "aborted", err==nil | Unit |
| 4 | Empty input aborts | `in=strings.NewReader("\n")` — expect "aborted" | Unit |
| 5 | Piped stdin errors | `isTerminalFn=false`, `yes=false` — expect "requires --yes" | Unit |
| 6 | `--all` lists tools | `runUninstall("", true, false, mockIn, &out)` — output includes tool names before prompt | Unit |
| 7 | Invalid tool still fails | `runUninstall("no-existe", false, true, nil, &out)` — expect "unknown adapter" | Unit |
| 8 | `--yes` flag registered | `cmd.SetArgs([]string{"uninstall", "--help"})` — output contains "--yes" | Integration |

Mock setup: set `isTerminalFn = func() bool { return true }` and restore via defer in each test that needs terminal behavior. Tests are NOT parallel when overriding `isTerminalFn`.

## Migration / Rollout

No migration required. `--yes` flag is additive; existing callers of `runUninstall` are internal and updated in the same commit. `go test ./...` validates zero regressions.

## Open Questions

None — all design questions resolved.
