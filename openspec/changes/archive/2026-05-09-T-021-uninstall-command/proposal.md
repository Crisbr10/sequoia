# Proposal: Uninstall Command Safety (T-021)

## Intent

`sequoia uninstall` currently runs immediately without confirmation. A destructive command that removes files from `~/.claude/`, `~/.config/opencode/`, and restores backups needs a safety gate. This adds `--yes` flag and interactive confirmation prompt.

## Scope

### In Scope
- `--yes` / `-y` flag on `uninstall` subcommand (skip confirmation)
- Interactive confirmation prompt when `--yes` absent and stdin is a terminal
- Graceful error when stdin is non-interactive and `--yes` not set
- Tests: `--yes` bypass, confirm-yes, confirm-no abort, `--all --yes`, invalid tool, piped stdin error

### Out of Scope
- TUI uninstall (Phase 5)
- Adapter `Uninstall()` method changes (already correct)
- Backup/restore logic (already in common installer)

## Capabilities

### New Capabilities
- `uninstall-confirmation`: Interactive confirmation prompt with `--yes` override for safe removal of Sequoia from AI tools

### Modified Capabilities
None — existing adapter and CLI specs unaffected.

## Approach

1. **Add `--yes` flag** to `newUninstallCmd()` via `cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, ...)`
2. **Inject `yesFlag` into `runUninstall()`** — signature becomes `runUninstall(toolID string, all, yesFlag bool, out io.Writer) error`
3. **Confirmation gate** in `runUninstall()` before the loop:
   - If `yesFlag`: skip prompt, proceed
   - If stdin is NOT a terminal: return error "piped input requires --yes"
   - If stdin IS a terminal: prompt "Remove Sequoia from {N} tool(s)? [y/N]: " via `fmt.Fscanln`
   - Any answer other than "y"/"Y" aborts with "aborted" message
4. **Refactor `runUninstall` to separate concerns**: determine targets first, confirm once, then execute
5. **Tests** using `bytes.Buffer` as mock stdin: table-driven with `{name, args, stdinInput, wantErrContains, wantOutputContains}`

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/sequoia/main.go` | Modified | `newUninstallCmd`: add `--yes` flag. `runUninstall`: confirmation gate, signature change |
| `cmd/sequoia/main_test.go` | Modified | New table-driven tests for confirmation flow |
| `openspec/changes/T-021-uninstall-command/specs/uninstall-confirmation/spec.md` | New | Delta spec per new capability |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Stdin mock doesn't behave like real terminal | Low | Use `bytes.Buffer` via `cmd.SetIn()` — same as install tests |
| Breaking existing callers of `runUninstall` | Low | Signature change is internal to `cmd/sequoia` package; `go test` catches |

## Rollback Plan

1. `git revert`: signature is internal, `go test ./...` passes on revert
2. `--yes` flag removed from help text, no user-facing artifact remains
3. No filesystem side effects — confirmation is read-only until user says yes

## Dependencies

- **T-020** (Multi-tool Detection): DONE — provides working `runUninstall`, adapter `Uninstall()`, and test infrastructure

## Success Criteria

- [ ] `sequoia uninstall --tool=claude-code --yes` removes without prompting
- [ ] `sequoia uninstall --tool=claude-code` (no --yes, terminal) prompts and respects "n" abort
- [ ] `sequoia uninstall --all --yes` loops all installed tools silently
- [ ] `sequoia uninstall --tool=claude-code` with piped stdin returns error mentioning `--yes`
- [ ] `go test ./cmd/sequoia/... -run Uninstall` passes all new cases
- [ ] `go test ./...` passes with zero regressions
