package screens

import (
	"fmt"
	"strings"

	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui"
	"github.com/Crisbr10/sequoia/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

// CompleteView renders the post-installation success summary screen.
// It lists each tool that was installed with its completion status and
// shows a hint for the first command to try.
// mode is the operation mode: "install" or "uninstall". Empty string defaults to "install".
func CompleteView(progressTools []ProgressTool, mode string) string {
	var b strings.Builder

	// Resolve heading based on mode.
	heading := "✅  Installation Complete!"
	if mode == "uninstall" {
		heading = "✅  Uninstallation Complete!"
	}

	// Success heading.
	b.WriteString(styles.Success().Render(heading))
	b.WriteString("\n\n")

	// Tools with their install status.
	for _, tool := range progressTools {
		allDone := allStepsDone(tool.Steps)
		var marker string
		if !allDone {
			marker = styles.Muted().Render("⚠️")
		} else {
			marker = styles.Success().Render("✅")
		}

		fmt.Fprintf(&b, "  %s %s\n", marker, styles.Body().Render(tool.ToolName))
	}

	b.WriteString("\n")

	// What was installed.
	b.WriteString(styles.Body().Render("  Installed: Skills, Commands, System Prompt"))
	b.WriteString("\n\n")

	// First command hint.
	b.WriteString(styles.Highlight().Render("  Try running: sequoia status"))
	b.WriteString("\n\n")

	// Key hints.
	b.WriteString(styles.Muted().Render("  "))
	b.WriteString(styles.Accent().Render("r"))
	b.WriteString(styles.Muted().Render(" — Status screen  "))
	b.WriteString(styles.Accent().Render("q"))
	b.WriteString(styles.Muted().Render(" — Quit"))

	return b.String()
}

// CompleteUpdate processes key events for the Complete screen.
// r navigates to the Status screen; q quits.
func CompleteUpdate(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyCtrlC:
		return tea.Quit
	}

	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		switch msg.Runes[0] {
		case 'r':
			return func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenStatus}
			}
		case 'q':
			return tea.Quit
		}
	}

	return nil
}
