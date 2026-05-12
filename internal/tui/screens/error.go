package screens

import (
	"fmt"
	"strings"

	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui"
	"github.com/Crisbr10/sequoia/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

// ErrorView renders the post-installation error summary screen.
// It lists each tool with its success/failure status and shows
// error messages for failed tools.
// mode is the operation mode: "install" or "uninstall". Empty string defaults to "install".
func ErrorView(progressTools []ProgressTool, mode string) string {
	var b strings.Builder

	// Resolve heading based on mode.
	heading := "❌  Installation Failed"
	if mode == "uninstall" {
		heading = "❌  Uninstallation Failed"
	}

	// Failure heading.
	b.WriteString(styles.Error().Render(heading))
	b.WriteString("\n\n")

	// Per-tool status list.
	for _, tool := range progressTools {
		allDone := allStepsDone(tool.Steps)
		hasFailed := hasAnyFailedStep(tool.Steps)

		marker := ""
		switch {
		case hasFailed:
			marker = styles.Error().Render("❌")
		case allDone:
			marker = styles.Success().Render("✅")
		default:
			marker = styles.Muted().Render("⚠️")
		}

		b.WriteString(fmt.Sprintf("  %s %s\n", marker, styles.Body().Render(tool.ToolName)))

		// Show failed steps with error messages.
		for _, step := range tool.Steps {
			if step.Status == StepFailed {
				b.WriteString(fmt.Sprintf("      %s: %s\n",
					styles.Error().Render(step.Name),
					styles.Error().Render(step.Error),
				))
			}
		}
	}

	b.WriteString("\n")

	// Retry / navigation options.
	b.WriteString(styles.Muted().Render("  "))
	b.WriteString(styles.Accent().Render("r"))
	b.WriteString(styles.Muted().Render(" — Retry failed  "))
	b.WriteString(styles.Accent().Render("Esc"))
	b.WriteString(styles.Muted().Render(" — Back to tools  "))
	b.WriteString(styles.Accent().Render("q"))
	b.WriteString(styles.Muted().Render(" — Quit"))

	return b.String()
}

// ErrorUpdate processes key events for the Error screen.
// r navigates to InstallProgress for retry, Esc goes back to ToolSelection, q quits.
func ErrorUpdate(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyLeft:
		return func() tea.Msg {
			return tui.NavigateMsg{Target: model.ScreenToolSelection}
		}
	case tea.KeyCtrlC:
		return tea.Quit
	}

	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		switch msg.Runes[0] {
		case 'r':
			return func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenInstallProgress}
			}
		case 'q':
			return tea.Quit
		}
	}

	return nil
}

// hasAnyFailedStep returns true if any step in the slice has status StepFailed.
func hasAnyFailedStep(steps []ProgressStep) bool {
	for _, s := range steps {
		if s.Status == StepFailed {
			return true
		}
	}
	return false
}
