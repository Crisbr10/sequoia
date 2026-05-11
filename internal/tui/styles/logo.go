package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// logoLines contains the SEQUOIA logo in a figlet box-drawing style.
// Row 6 is Q's exclusive tail — all other letters stop at row 5.
// The tail hangs below and to the right of Q's closing arc, making Q
// unmistakably different from O (which has no row 6 content).
var logoLines = []string{
	`   ███████╗███████╗ ██████╗ ██╗   ██╗ ██████╗ ██╗ █████╗ `,
	`   ██╔════╝██╔════╝██╔═══██╗██║   ██║██╔═══██╗██║██╔══██╗`,
	`   ███████╗█████╗  ██║   ██║██║   ██║██║   ██║██║███████║`,
	`   ╚════██║██╔══╝  ██║   ██║██║   ██║██║   ██║██║██╔══██║`,
	`   ███████║███████╗╚██████╔╝╚██████╔╝╚██████╔╝██║██║  ██║`,
	`   ╚══════╝╚══════╝ ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝╚═╝  ╚═╝`,
	`                           ╚══╝                             `,
	`               Audit Skills Installer — Sequoia AI          `,
}

// logoColors assigns a lipgloss color to each logo row (top → bottom).
// The Q tail row (6) uses bright goldenrod so it visually pops.
var logoColors = []lipgloss.Color{
	colorFoliage,              // row 0 — forest green
	lipgloss.Color("#2E8B57"), // row 1 — sea green
	lipgloss.Color("#3CB371"), // row 2 — medium sea green
	colorBark,                 // row 3 — bark brown
	lipgloss.Color("#A0522D"), // row 4 — sienna
	lipgloss.Color("#CD853F"), // row 5 — peru
	colorSunlight,             // row 6 — Q tail: bright goldenrod (unmissable)
	colorMuted,                // row 7 — tagline
}

// Logo returns the SEQUOIA logo with a top-to-bottom gradient.
// Q's tail on row 6 is the key visual differentiator from O.
func Logo() string {
	var b strings.Builder
	for i, line := range logoLines {
		color := logoColors[i]
		style := lipgloss.NewStyle().Foreground(color)
		b.WriteString(style.Render(line))
		if i < len(logoLines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
