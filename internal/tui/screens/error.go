package screens

import (
	"fmt"
	"strings"

	"github.com/Crisbr10/sequoia/internal/tui/styles"
)

// ErrorView renders the post-installation error summary screen.
// It lists each tool with its success/failure status and shows
// error messages for failed tools.
// mode is the operation mode: "install" or "uninstall". Empty string defaults to "install".
func ErrorView(progressTools []ProgressTool, mode string) string {
	var b strings.Builder
	b.Grow(300 + 100*len(progressTools)) // Heading + ~100 per tool (with errors) + hints

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

		fmt.Fprintf(&b, "  %s %s\n", marker, styles.Body().Render(tool.ToolName))

		// Show failed steps with error messages.
		for _, step := range tool.Steps {
			if step.Status == StepFailed {
				fmt.Fprintf(&b, "      %s: %s\n",
					styles.Error().Render(step.Name),
					styles.Error().Render(step.Error))
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

// hasAnyFailedStep returns true if any step in the slice has status StepFailed.
func hasAnyFailedStep(steps []ProgressStep) bool {
	for _, s := range steps {
		if s.Status == StepFailed {
			return true
		}
	}
	return false
}
