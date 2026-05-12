package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	figure "github.com/common-nighthawk/go-figure"
)

// Logo returns "Sequoia" rendered as ASCII art in forest green.
func Logo() string {
	fig := figure.NewFigure("Sequoia", "", true)
	raw := fig.String()
	style := lipgloss.NewStyle().Foreground(colorFoliage)
	lines := strings.Split(strings.TrimRight(raw, "\n"), "\n")
	var b strings.Builder
	for i, line := range lines {
		b.WriteString(style.Render(line))
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
