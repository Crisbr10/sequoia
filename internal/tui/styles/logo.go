package styles

import (
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	figure "github.com/common-nighthawk/go-figure"
)

var (
	logoOnce   sync.Once
	cachedLogo string
)

// Logo returns "Sequoia" rendered as ASCII art in forest green.
// The logo is generated once and cached for subsequent calls.
func Logo() string {
	logoOnce.Do(func() {
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
		cachedLogo = b.String()
	})
	return cachedLogo
}
