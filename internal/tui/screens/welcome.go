// Package screens provides rendering and interaction logic for each TUI screen
// in the Sequoia installer. Screens are pure functions that receive model state
// and return view strings or navigation commands.
package screens

import (
	"fmt"
	"strings"

	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui"
	"github.com/Crisbr10/sequoia/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

// WelcomeView renders the Welcome screen showing branding, version, and
// detected tools with their install status. tools is the adapter snapshot
// from the root model; version is the Sequoia release string.
func WelcomeView(tools []model.ToolState, version string) string {
	var b strings.Builder

	// Sequoia tree pixel art above the logo.
	b.WriteString(styles.SequoiaTree())
	b.WriteString("\n")

	// ASCII logo.
	b.WriteString(styles.Logo())
	b.WriteString("\n")

	// Version line.
	b.WriteString(styles.Muted().Render(fmt.Sprintf("  %s", version)))
	b.WriteString("\n\n")

	// Tagline.
	b.WriteString(styles.Body().Render("  Audit quality for AI coding tools"))
	b.WriteString("\n\n")

	// Detected tools header.
	b.WriteString(styles.Subtitle().Render("Detected Tools"))
	b.WriteString("\n\n")

	// Tool list with install status.
	if len(tools) == 0 {
		b.WriteString(styles.Muted().Render("  (no tools detected)"))
		b.WriteString("\n")
	} else {
		for _, ts := range tools {
			name := ts.Adapter.Name()
			installed := ts.Adapter.IsInstalled()
			statusIcon := "✗ not installed"
			if installed {
				statusIcon = styles.Success().Render("✓ installed")
			}
			b.WriteString(fmt.Sprintf("    %s  %s\n",
				styles.Body().Render(name),
				statusIcon,
			))
		}
	}

	b.WriteString("\n")

	// Footer hint.
	b.WriteString(styles.Muted().Render(
		fmt.Sprintf("  %s %s  %s",
			styles.Accent().Render("Enter"),
			styles.Muted().Render("to continue ―"),
			styles.Muted().Render("q"),
		),
	))
	b.WriteString(styles.Muted().Render(" to quit"))

	return b.String()
}

// WelcomeUpdate processes key events for the Welcome screen.
// Enter/RightArrow navigate to ToolSelection; q/ctrl+c quit;
// any other key is ignored (returns nil).
func WelcomeUpdate(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEnter, tea.KeyRight:
		return func() tea.Msg {
			return tui.NavigateMsg{Target: model.ScreenToolSelection}
		}
	case tea.KeyCtrlC:
		return tea.Quit
	}

	// Handle rune-based keys.
	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		switch msg.Runes[0] {
		case 'q':
			return tea.Quit
		}
	}

	return nil
}
