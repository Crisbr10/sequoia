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
