package screens

import (
	"fmt"
	"strings"

	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

// StatusView renders the Status screen with a table showing per-tool
// installation state: name, installed indicator, and version.
func StatusView(tools []model.ToolState, cursor int) string {
	var b strings.Builder

	// Title.
	b.WriteString(styles.Title().Render("Status"))
	b.WriteString("\n\n")

	if len(tools) == 0 {
		b.WriteString(styles.Muted().Render("  No adapters registered"))
		b.WriteString("\n\n")
	} else {
		// Per-tool status rows.
		for i, ts := range tools {
			b.WriteString(renderStatusRow(ts, cursor == i))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Key hints.
	b.WriteString(styles.Muted().Render("  "))
	b.WriteString(styles.Accent().Render("↑/↓ j/k"))
	b.WriteString(styles.Muted().Render(" navigate  "))
	b.WriteString(styles.Accent().Render("u"))
	b.WriteString(styles.Muted().Render(" update  "))
	b.WriteString(styles.Accent().Render("r"))
	b.WriteString(styles.Muted().Render(" reinstall  "))
	b.WriteString(styles.Accent().Render("d"))
	b.WriteString(styles.Muted().Render(" uninstall  "))
	b.WriteString(styles.Accent().Render("Esc"))
	b.WriteString(styles.Muted().Render(" back  "))
	b.WriteString(styles.Accent().Render("q"))
	b.WriteString(styles.Muted().Render(" quit"))

	return b.String()
}

// renderStatusRow renders a single tool status row with cursor highlight.
func renderStatusRow(ts model.ToolState, highlighted bool) string {
	status := ts.Adapter.Status()
	installed := status.Installed

	// Cursor indicator.
	prefix := styles.Muted().Render("  ")
	if highlighted {
		prefix = styles.Accent().Render("▶ ")
	}

	// Installed indicator.
	marker := styles.Error().Render("❌")
	if installed {
		marker = styles.Success().Render("✅")
	}

	// Version.
	version := status.Version
	if version == "" {
		version = "—"
	}

	line := fmt.Sprintf("%s%s %s  %s",
		prefix,
		marker,
		styles.Body().Render(ts.Adapter.Name()),
		styles.Muted().Render(version),
	)
	return line
}

// StatusUpdate handles key events on the Status screen.
// It returns the new cursor position and an action string.
// Actions: "uninstall" (d), "reinstall" (r), "update" (u), "back" (Esc/Left), or "".
func StatusUpdate(msg tea.KeyMsg, cursor int, numTools int) (int, string) {
	switch msg.Type {
	case tea.KeyUp:
		if numTools == 0 {
			return cursor, ""
		}
		cursor--
		if cursor < 0 {
			cursor = numTools - 1
		}
		return cursor, ""

	case tea.KeyDown:
		if numTools == 0 {
			return cursor, ""
		}
		cursor++
		if cursor >= numTools {
			cursor = 0
		}
		return cursor, ""

	case tea.KeyEsc, tea.KeyLeft:
		return cursor, "back"

	case tea.KeyRunes:
		if len(msg.Runes) == 0 {
			return cursor, ""
		}
		switch msg.Runes[0] {
		case 'j':
			if numTools == 0 {
				return cursor, ""
			}
			cursor++
			if cursor >= numTools {
				cursor = 0
			}
			return cursor, ""

		case 'k':
			if numTools == 0 {
				return cursor, ""
			}
			cursor--
			if cursor < 0 {
				cursor = numTools - 1
			}
			return cursor, ""

		case 'd':
			return cursor, "uninstall"

		case 'r':
			return cursor, "reinstall"

		case 'u':
			return cursor, "update"

		default:
			return cursor, ""
		}
	}

	return cursor, ""
}
