package screens

import (
	"fmt"
	"strings"

	"github.com/Crisbr10/sequoia/internal/i18n"
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

// UninstallView renders the Uninstall screen showing a checkbox list
// of installed tools. Tools that are not installed are not shown.
// The full tools array is iterated so cursor indexing matches the
// original ToolState positions; only installed tools are rendered.
// errorMsg, when non-empty, is rendered above the footer as a validation error.
// lang is the current UI language (e.g., "en", "es").
func UninstallView(tools []model.ToolState, cursor int, errorMsg string, lang string) string {
	var b strings.Builder

	// Title.
	b.WriteString(styles.Title().Render(i18n.T(i18n.MsgUninstallTitle, lang)))
	b.WriteString("\n\n")

	// Count installed tools to detect empty state.
	installedCount := 0
	for _, ts := range tools {
		if ts.Adapter.IsInstalled() {
			installedCount++
		}
	}

	if installedCount == 0 {
		b.WriteString(styles.Muted().Render("  " + i18n.T(i18n.MsgUninstallEmpty, lang)))
		b.WriteString("\n\n")

		// Error message (if any).
		if errorMsg != "" {
			b.WriteString(styles.Error().Render("  "))
			b.WriteString(styles.Error().Render(errorMsg))
			b.WriteString("\n\n")
		}

		// Only q available when nothing to uninstall.
		b.WriteString(styles.Muted().Render("  "))
		b.WriteString(styles.Accent().Render(i18n.T(i18n.MsgFooterQuitKey, lang)))
		b.WriteString(styles.Muted().Render(i18n.T(i18n.MsgFooterQuit, lang)))
		return b.String()
	}

	// Iterate full tools array; skip non-installed tools.
	// Cursor indexing uses the full array position so cursor wraps correctly
	// within the installed subset (see REQ-BUG-001).
	for i, ts := range tools {
		if !ts.Adapter.IsInstalled() {
			continue
		}
		b.WriteString(renderUninstallRow(ts, i, cursor == i))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Error message (if any).
	if errorMsg != "" {
		b.WriteString(styles.Error().Render("  "))
		b.WriteString(styles.Error().Render(errorMsg))
		b.WriteString("\n\n")
	}

	// Key hints.
	b.WriteString(styles.Muted().Render("  "))
	b.WriteString(styles.Accent().Render(i18n.T(i18n.MsgFooterNavigateKeys, lang)))
	b.WriteString(styles.Muted().Render(i18n.T(i18n.MsgFooterNavigate, lang)))
	b.WriteString(styles.Accent().Render(i18n.T(i18n.MsgFooterToggleKey, lang)))
	b.WriteString(styles.Muted().Render(i18n.T(i18n.MsgFooterToggle, lang)))
	b.WriteString(styles.Accent().Render(i18n.T(i18n.MsgFooterConfirmKey, lang)))
	b.WriteString(styles.Muted().Render(i18n.T(i18n.MsgFooterConfirm, lang)))
	b.WriteString(styles.Accent().Render(i18n.T(i18n.MsgFooterBackKey, lang)))
	b.WriteString(styles.Muted().Render(i18n.T(i18n.MsgFooterBack, lang)))
	b.WriteString(styles.Accent().Render(i18n.T(i18n.MsgFooterQuitKey, lang)))
	b.WriteString(styles.Muted().Render(i18n.T(i18n.MsgFooterQuit, lang)))

	return b.String()
}

// filterInstalled returns only tools where IsInstalled() is true.
func filterInstalled(tools []model.ToolState) []model.ToolState {
	var result []model.ToolState
	for _, ts := range tools {
		if ts.Adapter.IsInstalled() {
			result = append(result, ts)
		}
	}
	return result
}

// renderUninstallRow renders a single uninstall checkbox row.
func renderUninstallRow(ts model.ToolState, _ int, highlighted bool) string {
	// Checkbox.
	var checkbox string
	if ts.Selected {
		checkbox = styles.Success().Render("[x]")
	} else {
		checkbox = styles.Muted().Render("[ ]")
	}

	// Cursor indicator.
	prefix := "  "
	if highlighted {
		prefix = styles.Accent().Render("▶ ")
	}

	return fmt.Sprintf("%s%s %s",
		prefix,
		checkbox,
		styles.Body().Render(ts.Adapter.Name()),
	)
}

// UninstallUpdate handles key events on the Uninstall screen.
// It returns the new cursor position, whether to toggle the current selection,
// and an action string: "confirm" (Enter), "back" (Esc/Left), or "" (no action).
// tools is the full []ToolState array. Navigation skips non-installed tools
// and wraps within the installed subset (REQ-BUG-001).
func UninstallUpdate(msg tea.KeyMsg, cursor int, tools []model.ToolState) (int, bool, string) {
	// Count installed tools for boundary checks.
	installedCount := 0
	for _, ts := range tools {
		if ts.Adapter.IsInstalled() {
			installedCount++
		}
	}

	switch msg.Type {
	case tea.KeyUp:
		if installedCount == 0 {
			return cursor, false, ""
		}
		// Find the previous installed tool, wrapping around.
		for {
			cursor--
			if cursor < 0 {
				cursor = len(tools) - 1
			}
			if tools[cursor].Adapter.IsInstalled() {
				break
			}
		}
		return cursor, false, ""

	case tea.KeyDown:
		if installedCount == 0 {
			return cursor, false, ""
		}
		// Find the next installed tool, wrapping around.
		for {
			cursor++
			if cursor >= len(tools) {
				cursor = 0
			}
			if tools[cursor].Adapter.IsInstalled() {
				break
			}
		}
		return cursor, false, ""

	case tea.KeySpace:
		return cursor, true, ""

	case tea.KeyEnter:
		if installedCount == 0 {
			return cursor, false, ""
		}
		return cursor, false, "confirm"

	case tea.KeyEsc, tea.KeyLeft:
		return cursor, false, "back"

	case tea.KeyRunes:
		if len(msg.Runes) == 0 {
			return cursor, false, ""
		}
		switch msg.Runes[0] {
		case ' ':
			return cursor, true, ""
		case 'j':
			if installedCount == 0 {
				return cursor, false, ""
			}
			for {
				cursor++
				if cursor >= len(tools) {
					cursor = 0
				}
				if tools[cursor].Adapter.IsInstalled() {
					break
				}
			}
			return cursor, false, ""

		case 'k':
			if installedCount == 0 {
				return cursor, false, ""
			}
			for {
				cursor--
				if cursor < 0 {
					cursor = len(tools) - 1
				}
				if tools[cursor].Adapter.IsInstalled() {
					break
				}
			}
			return cursor, false, ""

		default:
			return cursor, false, ""
		}
	}

	return cursor, false, ""
}

// RenderConfirmPrompt returns the styled confirmation prompt for the Uninstall screen.
func RenderConfirmPrompt(lang string) string {
	return styles.Accent().Render("  "+i18n.T(i18n.MsgUninstallConfirmPrompt, lang)) +
		styles.Muted().Render(i18n.T(i18n.MsgUninstallConfirmSuffix, lang))
}
