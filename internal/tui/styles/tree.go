package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SequoiaTree returns a detailed pixel-art sequoia tree using full-block and
// half-block Unicode characters. Three colour bands mirror the tree's anatomy:
//   - Fresh bright-green rings between canopy layers (new growth)
//   - Forest-green main canopy mass
//   - Saddle-brown trunk and root flare
func SequoiaTree() string {
	canopy := lipgloss.NewStyle().Foreground(colorFoliage)
	fresh := lipgloss.NewStyle().Foreground(lipgloss.Color("#90EE90"))
	bark := lipgloss.NewStyle().Foreground(colorBark)
	roots := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B3410"))

	lines := []string{
		canopy.Render(`                   ▄█▄                   `),
		canopy.Render(`                 ▄█████▄                 `),
		canopy.Render(`                ▄███████▄                `),
		fresh.Render( `               ▀▀▀▀▀▀▀▀▀▀▀               `),
		canopy.Render(`              ▄███████████▄              `),
		canopy.Render(`             ▄█████████████▄             `),
		canopy.Render(`            ▄███████████████▄            `),
		fresh.Render( `           ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀           `),
		canopy.Render(`          ▄█████████████████▄            `),
		canopy.Render(`         ▄███████████████████▄           `),
		canopy.Render(`        ▄█████████████████████▄          `),
		fresh.Render( `       ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀         `),
		canopy.Render(`      ▄███████████████████████▄          `),
		canopy.Render(`     ▄█████████████████████████▄         `),
		canopy.Render(`    ▄███████████████████████████▄        `),
		bark.Render(  `               ██████████               `),
		bark.Render(  `               ██████████               `),
		bark.Render(  `               ██████████               `),
		bark.Render(  `               ██████████               `),
		bark.Render(  `               ██████████               `),
		roots.Render( `          ▄████████████████▄            `),
		roots.Render( `       ▄██████████████████████▄         `),
	}

	return strings.Join(lines, "\n") + "\n"
}
