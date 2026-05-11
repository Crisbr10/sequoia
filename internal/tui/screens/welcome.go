// Package screens provides rendering and interaction logic for each TUI screen
// in the Sequoia installer. Screens are pure functions that receive model state
// and return view strings or navigation commands.
package screens

import (
	"fmt"
	"strings"

	"github.com/Crisbr10/sequoia/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

// Welcome menu item indices.
const (
	WelcomeMenuInstall   = 0
	WelcomeMenuStatus    = 1
	WelcomeMenuUninstall = 2
	WelcomeMenuQuit      = 3
	WelcomeMenuCount     = 4
)

// welcomeMenuLabels maps menu index → display label.
var welcomeMenuLabels = [WelcomeMenuCount]string{
	"Install",
	"Status",
	"Uninstall",
	"Quit",
}

// WelcomeView renders the Welcome/Home screen: logo, version, and a
// navigable main menu. cursor is the currently highlighted menu item index.
func WelcomeView(version string, cursor int) string {
	var b strings.Builder

	// ASCII logo with gradient.
	b.WriteString(styles.Logo())
	b.WriteString("\n\n")

	// Version line.
	b.WriteString(styles.Muted().Render(fmt.Sprintf("  %s", version)))
	b.WriteString("\n\n")

	// Main menu.
	b.WriteString(styles.Subtitle().Render("  Menu"))
	b.WriteString("\n\n")
	for i, label := range welcomeMenuLabels {
		if i == cursor {
			b.WriteString(styles.Accent().Render("  ▶ " + label))
		} else {
			b.WriteString(styles.Body().Render("    " + label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.Muted().Render("  j/k ↑/↓: navigate  Enter: select  q: quit"))

	return b.String()
}

// WelcomeUpdate handles key events on the Welcome screen.
// Returns the new cursor position and an action string:
//   - "install"   → navigate to tool selection
//   - "status"    → navigate to status screen
//   - "uninstall" → navigate to uninstall screen
//   - "quit"      → quit the application
//   - ""          → no navigation (cursor moved or key ignored)
func WelcomeUpdate(msg tea.KeyMsg, cursor int) (int, string) {
	switch msg.Type {
	case tea.KeyUp:
		return wrapMenuDecrement(cursor), ""
	case tea.KeyDown:
		return wrapMenuIncrement(cursor), ""

	case tea.KeyEnter, tea.KeyRight:
		return cursor, welcomeMenuAction(cursor)

	case tea.KeyRunes:
		if len(msg.Runes) == 0 {
			return cursor, ""
		}
		switch msg.Runes[0] {
		case 'j':
			return wrapMenuIncrement(cursor), ""
		case 'k':
			return wrapMenuDecrement(cursor), ""
		}
	}

	return cursor, ""
}

// welcomeMenuAction maps a cursor position to its navigation action string.
func welcomeMenuAction(cursor int) string {
	switch cursor {
	case WelcomeMenuInstall:
		return "install"
	case WelcomeMenuStatus:
		return "status"
	case WelcomeMenuUninstall:
		return "uninstall"
	case WelcomeMenuQuit:
		return "quit"
	}
	return ""
}

func wrapMenuIncrement(n int) int {
	n++
	if n >= WelcomeMenuCount {
		return 0
	}
	return n
}

func wrapMenuDecrement(n int) int {
	n--
	if n < 0 {
		return WelcomeMenuCount - 1
	}
	return n
}
