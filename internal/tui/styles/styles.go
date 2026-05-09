// Package styles provides the lipgloss theme for the Sequoia TUI.
package styles

import "github.com/charmbracelet/lipgloss"

// Sequoia color palette — inspired by sequoia tree bark and foliage.
var (
	colorBark     = lipgloss.Color("#8B4513") // saddle brown
	colorFoliage  = lipgloss.Color("#228B22") // forest green
	colorSky      = lipgloss.Color("#4682B4") // steel blue
	colorSunlight = lipgloss.Color("#DAA520") // goldenrod
	colorError    = lipgloss.Color("#DC143C") // crimson
	colorMuted    = lipgloss.Color("#696969") // dim gray
)

// Title returns a bold, large-text style for screen headers.
func Title() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(colorFoliage).
		MarginLeft(2).
		Padding(0, 1)
}

// Subtitle returns a secondary heading style.
func Subtitle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(colorSky).
		MarginLeft(2)
}

// Body returns the default body text style.
func Body() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D3D3D3"))
}

// Accent returns a highlighted style for interactive elements and key labels.
func Accent() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(colorSunlight)
}

// Error returns a red style for error messages and failed states.
func Error() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(colorError)
}

// Success returns a green style for success messages and completed states.
func Success() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(colorFoliage)
}

// Muted returns a dimmed style for secondary or disabled information.
func Muted() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(colorMuted)
}

// Highlight returns a bright, eye-catching style for important notices.
func Highlight() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(colorBark).
		Background(lipgloss.Color("#FFF8DC")) // cornsilk
}
