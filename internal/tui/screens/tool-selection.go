package screens

import (
	"fmt"
	"strings"

	"github.com/Crisbr10/sequoia/internal/i18n"
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

// ToolSelectionView renders the Tool Selection screen showing a checkbox list
// of detected tools. cursor indicates the currently highlighted item; errorMsg
// is displayed when validation fails (e.g., zero tools selected).
// lang is the current UI language (e.g., "en", "es").
func ToolSelectionView(tools []model.ToolState, cursor int, errorMsg string, lang string) string {
	var b strings.Builder

	// Title.
	b.WriteString(styles.Title().Render(i18n.T(i18n.MsgToolSelectionTitle, lang)))
	b.WriteString("\n\n")

	// Instruction.
	b.WriteString(styles.Body().Render("  " + i18n.T(i18n.MsgToolSelectionInstruction, lang)))
	b.WriteString("\n\n")

	// Tool list with checkboxes and cursor.
	if len(tools) == 0 {
		b.WriteString(styles.Muted().Render("  " + i18n.T(i18n.MsgToolSelectionEmpty, lang)))
		b.WriteString("\n")
	} else {
		for i, ts := range tools {
			checkbox := "[ ]"
			if ts.Selected {
				checkbox = styles.Success().Render("[x]")
			}

			cursorMark := "  "
			if i == cursor {
				cursorMark = styles.Accent().Render("▶ ")
			}

			fmt.Fprintf(&b, "%s%s %s\n",
				cursorMark,
				checkbox,
				styles.Body().Render(ts.Adapter.Name()))
		}
	}

	// Selection count.
	selected := countSelectedTools(tools)
	b.WriteString("\n")
	b.WriteString(styles.Muted().Render(
		fmt.Sprintf("  %s", i18n.T(i18n.MsgToolSelectionSelectedCount, lang, map[string]interface{}{
			"Selected": selected,
			"Total":    len(tools),
		})),
	))
	b.WriteString("\n\n")

	// Error message (if any).
	if errorMsg != "" {
		b.WriteString(styles.Error().Render(fmt.Sprintf("  %s", errorMsg)))
		b.WriteString("\n\n")
	}

	// Footer hints.
	b.WriteString(styles.Muted().Render("  "))
	b.WriteString(styles.Accent().Render(i18n.T(i18n.MsgFooterNavigateKeys, lang)))
	b.WriteString(styles.Muted().Render(i18n.T(i18n.MsgFooterNavigate, lang)))
	b.WriteString(styles.Accent().Render(i18n.T(i18n.MsgFooterToggleKey, lang)))
	b.WriteString(styles.Muted().Render(i18n.T(i18n.MsgFooterToggle, lang)))
	b.WriteString(styles.Accent().Render(i18n.T(i18n.MsgFooterConfirmKey, lang)))
	b.WriteString(styles.Muted().Render(i18n.T(i18n.MsgFooterConfirm, lang)))
	b.WriteString(styles.Accent().Render(i18n.T(i18n.MsgFooterBackKey, lang)))
	b.WriteString(styles.Muted().Render(i18n.T(i18n.MsgFooterBack, lang)))

	return b.String()
}

// countSelectedTools returns how many tools have Selected=true.
func countSelectedTools(tools []model.ToolState) int {
	n := 0
	for _, ts := range tools {
		if ts.Selected {
			n++
		}
	}
	return n
}

// ToolSelectionUpdate processes key events for the Tool Selection screen.
// Returns the new cursor position, whether the current tool should be toggled,
// and a navigation action string ("confirm", "back", "quit", or "").
// toolCount is the total number of tools available.
func ToolSelectionUpdate(msg tea.KeyMsg, cursor int, toolCount int) (newCursor int, shouldToggle bool, action string) {
	switch msg.Type {
	case tea.KeyUp:
		return wrapDecrement(cursor, toolCount), false, ""
	case tea.KeyDown:
		return wrapIncrement(cursor, toolCount), false, ""
	case tea.KeySpace:
		return cursor, true, ""
	case tea.KeyEnter:
		if toolCount == 0 {
			return cursor, false, "back"
		}
		return cursor, false, "confirm"
	case tea.KeyEsc, tea.KeyLeft:
		return cursor, false, "back"
	}

	// Handle rune-based keys.
	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		switch msg.Runes[0] {
		case ' ':
			return cursor, true, ""
		case 'j':
			return wrapIncrement(cursor, toolCount), false, ""
		case 'k':
			return wrapDecrement(cursor, toolCount), false, ""
		}
	}

	return cursor, false, ""
}

// wrapDecrement decrements n, wrapping to limit-1 if n <= 0.
func wrapDecrement(n, limit int) int {
	if limit <= 1 {
		return 0
	}
	n--
	if n < 0 {
		return limit - 1
	}
	return n
}

// wrapIncrement increments n, wrapping to 0 if n >= limit-1.
func wrapIncrement(n, limit int) int {
	if limit <= 1 {
		return 0
	}
	n++
	if n >= limit {
		return 0
	}
	return n
}
