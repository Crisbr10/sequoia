// Package styles provides the lipgloss theme for the Sequoia TUI.
package styles

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// Sequoia color palette — inspired by sequoia tree bark and foliage.
var (
	colorBark     = lipgloss.Color("#8B4513") // saddle brown
	colorFoliage  = lipgloss.Color("#228B22") // forest green
	colorSky      = lipgloss.Color("#4682B4") // steel blue
	colorSunlight = lipgloss.Color("#DAA520") // goldenrod
	colorError    = lipgloss.Color("#DC143C") // crimson
	colorMuted    = lipgloss.Color("#696969") // dim gray
)

// Cached lipgloss styles — built once lazily via sync.Once.
// Lipgloss styles are immutable after construction, so sharing
// a single instance across all callers is safe and eliminates
// per-frame heap allocations at 60fps.
var (
	titleStyle     lipgloss.Style
	titleOnce      sync.Once
	subtitleStyle  lipgloss.Style
	subtitleOnce   sync.Once
	bodyStyle      lipgloss.Style
	bodyOnce       sync.Once
	accentStyle    lipgloss.Style
	accentOnce     sync.Once
	errorStyle     lipgloss.Style
	errorOnce      sync.Once
	successStyle   lipgloss.Style
	successOnce    sync.Once
	mutedStyle     lipgloss.Style
	mutedOnce      sync.Once
	highlightStyle lipgloss.Style
	highlightOnce  sync.Once
)

// Title returns a bold, large-text style for screen headers.
func Title() lipgloss.Style {
	titleOnce.Do(func() {
		titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorFoliage).
			MarginLeft(2).
			Padding(0, 1)
	})
	return titleStyle
}

// Subtitle returns a secondary heading style.
func Subtitle() lipgloss.Style {
	subtitleOnce.Do(func() {
		subtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSky).
			MarginLeft(2)
	})
	return subtitleStyle
}

// Body returns the default body text style.
func Body() lipgloss.Style {
	bodyOnce.Do(func() {
		bodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D3D3D3"))
	})
	return bodyStyle
}

// Accent returns a highlighted style for interactive elements and key labels.
func Accent() lipgloss.Style {
	accentOnce.Do(func() {
		accentStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSunlight)
	})
	return accentStyle
}

// Error returns a red style for error messages and failed states.
func Error() lipgloss.Style {
	errorOnce.Do(func() {
		errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorError)
	})
	return errorStyle
}

// Success returns a green style for success messages and completed states.
func Success() lipgloss.Style {
	successOnce.Do(func() {
		successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorFoliage)
	})
	return successStyle
}

// Muted returns a dimmed style for secondary or disabled information.
func Muted() lipgloss.Style {
	mutedOnce.Do(func() {
		mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)
	})
	return mutedStyle
}

// Highlight returns a bright, eye-catching style for important notices.
func Highlight() lipgloss.Style {
	highlightOnce.Do(func() {
		highlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBark).
			Background(lipgloss.Color("#FFF8DC")) // cornsilk
	})
	return highlightStyle
}
