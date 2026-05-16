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
// warnedCount is the number of tools that completed with non-fatal warnings.
func CompleteView(progressTools []ProgressTool, mode string, warnedCount int) string {
	var b strings.Builder

	// Resolve heading based on mode.
	heading := "✅  Installation Complete!"
	if mode == "uninstall" {
		if warnedCount > 0 {
			heading = fmt.Sprintf("⚠️  Uninstall Completed with Warnings — %d tool(s) had issues", warnedCount)
		} else {
			heading = "✅  Uninstallation Complete!"
		}
	}

	// Success heading.
	if warnedCount > 0 && mode == "uninstall" {
		b.WriteString(styles.Accent().Render(heading))
	} else {
		b.WriteString(styles.Success().Render(heading))
	}
	b.WriteString("\n\n")

	// Tools with their install status.
	for _, tool := range progressTools {
		allDone := allStepsDone(tool.Steps)
		hasWarnings := hasWarningSteps(tool.Steps)
		var marker string
		switch {
		case hasWarnings:
			marker = styles.Accent().Render("⚠")
		case !allDone:
			marker = styles.Muted().Render("⚠️")
		default:
			marker = styles.Success().Render("✅")
		}

		fmt.Fprintf(&b, "  %s %s\n", marker, styles.Body().Render(tool.ToolName))

		// Show warning details for steps that completed with warnings.
		for _, step := range tool.Steps {
			if step.Error != "" && step.Status == StepDone {
				fmt.Fprintf(&b, "      %s: %s\n",
					styles.Muted().Render(step.Name),
					styles.Muted().Render(step.Error))
			}
		}
	}

	b.WriteString("\n")

	// What was installed or uninstalled.
	if mode == "uninstall" && warnedCount > 0 {
		b.WriteString(styles.Muted().Render("  Some files could not be removed — check permissions and try again."))
	} else if mode == "uninstall" {
		b.WriteString(styles.Body().Render("  Uninstalled: Skills, Commands, System Prompt"))
	} else {
		b.WriteString(styles.Body().Render("  Installed: Skills, Commands, System Prompt"))
	}
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

// hasWarningSteps returns true if any step in the slice completed with
// a warning (StepDone with non-empty Error).
func hasWarningSteps(steps []ProgressStep) bool {
	for _, s := range steps {
		if s.Status == StepDone && s.Error != "" {
			return true
		}
	}
	return false
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
