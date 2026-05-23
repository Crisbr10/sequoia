# error-context-wrapping

> Source: CORR-004 — Common Package Split (archived 2026-05-22)
> Domain: MODIFIED — P4-003 bare error returns

All bare `return err` statements in `adapters/common/strategy.go` SHALL be wrapped with operation context including file path, operation name, and adapter identifier.

## Requirements

| # | Requirement | Strength |
|---|------------|----------|
| R13 | Every `return err` in `strategy.go` wrapped via `fmt.Errorf("op %s path %s: %w", op, path, err)` | MUST |
| R14 | Wrapping uses `fmt.Errorf` + `%w` verb (preserves error chain for `errors.Is`/`errors.As`) | MUST |
| R15 | Zero bare `return err` in `strategy.go` after refactor | MUST |

## Scenarios

### Wrapped error includes file path
- GIVEN `AtomicWriteFile("/etc/config.yaml", ...)` fails with permission denied
- WHEN the error propagates to the caller
- THEN the error message contains `"/etc/config.yaml"` and `"write"`
- AND `errors.Is(err, os.ErrPermission)` returns true

### Wrapped error preserves chain
- GIVEN `InjectMarkdownSection()` fails inside `MkdirAll`
- WHEN the wrapped error is checked with `errors.Is(err, os.ErrPermission)`
- THEN the check succeeds through the wrapping layer
