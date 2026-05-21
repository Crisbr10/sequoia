package styles

import (
	"strings"
	"sync"

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

// renderedLogo is the cached, pre-rendered ANSI-colored ASCII art logo.
// Built once on first Logo() call via sync.Once, so tests that set FORCE_COLOR
// before the first call trigger color output correctly.
var (
	renderedLogo string
	logoOnce     sync.Once
)

// Logo returns "Sequoia" rendered as ASCII art in forest green.
// The output is cached after the first call; all subsequent calls
// return the cached string with zero heap allocations.
func Logo() string {
	logoOnce.Do(func() {
		style := lipgloss.NewStyle().Foreground(colorFoliage)
		lines := strings.Split(rawLogo, "\n")
		var b strings.Builder
		b.Grow(500) // Known size: ~6 lines of 60-70 chars each with ANSI codes.
		for i, line := range lines {
			b.WriteString(style.Render(line))
			if i < len(lines)-1 {
				b.WriteByte('\n')
			}
		}
		renderedLogo = b.String()
	})
	return renderedLogo
}
