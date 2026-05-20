# Design: Consolidate go-figure Usage

## Technical Approach

Replace the runtime `sync.Once` + `figure.NewFigure("Sequoia", "", true)` in `Logo()` with a compile-time `const rawLogo` containing the identical 6-line ASCII art. The lipgloss styling (forest green per-line) is preserved; only the source of the raw text changes from dynamic generation to a string literal.

## Architecture Decisions

| Option | Tradeoff | Decision |
|--------|----------|----------|
| **A: Embed raw ASCII + style at runtime** | Preserves styling at call site; const is human-readable; zero behavioral change | **Chosen** |
| B: Embed fully styled output as const | Fewer runtime operations, but opaque ANSI escapes baked in; harder to visually verify art matches | Rejected — couples art to color constant |
| C: Generate art at build time via `go generate` | Over-engineered for 6 static lines; adds tooling dependency | Rejected |

## Data Flow

```
const rawLogo (6-line ASCII)  ──→  Logo() splits per line  ──→  lipgloss.NewStyle().Foreground(colorFoliage).Render(line)  ──→  return styled string
```

No goroutine synchronization needed — const reads are inherently data-race-free.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/tui/styles/logo.go` | Modify | Add `const rawLogo`, simplify `Logo()` body, drop `sync` + `go-figure` imports |
| `go.mod` | Modify | Remove `github.com/common-nighthawk/go-figure` line |
| `go.sum` | Modify | Remove go-figure checksums (via `go mod tidy`) |
| `internal/tui/screens/testdata/golden/welcome_standard.txt` | Regenerate | Golden file via `UPDATE_GOLDEN=1` |
| `internal/tui/screens/testdata/golden/welcome_cursor_status.txt` | Regenerate | Golden file via `UPDATE_GOLDEN=1` |

## Target Code

**New `logo.go`**:

```go
package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// rawLogo is the 6-line ASCII art for "Sequoia" originally produced
// by figure.NewFigure("Sequoia", "", true).String().
const rawLogo = `  ____                                   _
 / ___|    ___    __ _   _   _    ___   (_)   __ _
 \___ \   / _ \  / _` + "`" + ` | | | | |  / _ \  | |  / _` + "`" + ` |
  ___) | |  __/ | (_| | | |_| | | (_) | | | | (_| |
 |____/   \___|  \__, |  \__,_|  \___/  |_|  \__,_|
                    |_|`

// Logo returns "Sequoia" rendered as ASCII art in forest green.
func Logo() string {
	style := lipgloss.NewStyle().Foreground(colorFoliage)
	lines := strings.Split(rawLogo, "\n")
	var b strings.Builder
	for i, line := range lines {
		b.WriteString(style.Render(line))
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
```

Note: backtick characters in the const require Go string literal escaping (` + "`" + `). The raw art is captured verbatim from the golden file output — 6 lines, no trailing newline.

## Golden File Strategy

After code change, regenerate:

```powershell
$env:UPDATE_GOLDEN="1"; go test ./internal/tui/screens/...
```

Verify by running without the flag:

```powershell
go test ./internal/tui/screens/...
```

## Test Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit (`logo_test.go`) | NonEmpty, MultiLine, ANSI, idempotency, goroutine safety | All 5 tests pass unchanged — const satisfies every invariant previously ensured by `sync.Once` |
| Golden (`welcome_golden_test.go`) | Welcome screen rendering with cursor at index 0 and 1 | Regenerate via `UPDATE_GOLDEN=1`; diff manually before committing |

## Implementation Steps

1. Replace `internal/tui/styles/logo.go` with const-based version
2. `go mod tidy` (removes go-figure from go.mod/go.sum)
3. `$env:UPDATE_GOLDEN="1"; go test ./internal/tui/screens/...`
4. `go test ./internal/tui/styles/...` (verify logo unit tests)
5. `go test -race ./...` (full regression)
6. Visual diff golden files before commit

## Migration / Rollout

No migration required. Rollback: revert `logo.go` + `go mod tidy` + `git checkout testdata/golden/`.

## Open Questions

None.
