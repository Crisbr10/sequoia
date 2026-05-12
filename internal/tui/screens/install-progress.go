package screens

import (
	"fmt"
	"strings"

	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

// StepStatus indicates the current state of an install step.
type StepStatus int

const (
	// StepPending means the step has not started yet.
	StepPending StepStatus = iota
	// StepRunning means the step is currently executing.
	StepRunning
	// StepDone means the step completed successfully.
	StepDone
	// StepFailed means the step encountered an error.
	StepFailed
)

// ProgressStep represents the render state for a single install step.
type ProgressStep struct {
	// Name is the human-readable step name (e.g., "Skills", "Commands", "System Prompt").
	Name string
	// Status is the current execution state of this step.
	Status StepStatus
	// Error contains the error message when Status is StepFailed.
	Error string
}

// ProgressTool represents per-tool install progress data for the display.
type ProgressTool struct {
	// ToolID is the adapter's unique ID (e.g., "claude-code").
	ToolID string
	// ToolName is the display name of the tool being installed.
	ToolName string
	// Steps tracks the progress of each installation step for this tool.
	Steps []ProgressStep
}

// InstallProgressView renders the Install Progress screen showing per-tool
// step-by-step progress. completedCount is the number of fully-finished tools;
// totalCount is the total number of tools being installed.
// mode is the operation mode: "install" or "uninstall". Empty string defaults to "install".
func InstallProgressView(tools []ProgressTool, completedCount, totalCount int, mode string) string {
	var b strings.Builder

	// Resolve labels based on mode.
	titleLabel := "Installing"
	progressLabel := "Installing"
	if mode == "uninstall" {
		titleLabel = "Uninstalling"
		progressLabel = "Uninstalling"
	}

	// Title.
	b.WriteString(styles.Title().Render(titleLabel))
	b.WriteString("\n\n")

	// Overall progress summary.
	summary := fmt.Sprintf("  %s %d of %d tools...", progressLabel, completedCount, totalCount)
	b.WriteString(styles.Body().Render(summary))
	b.WriteString("\n\n")

	// Per-tool progress blocks.
	for _, tool := range tools {
		b.WriteString(renderToolProgress(tool))
		b.WriteString("\n")
	}

	// Footer hints.
	b.WriteString(styles.Muted().Render("  "))
	b.WriteString(styles.Accent().Render("q"))
	b.WriteString(styles.Muted().Render(" quit"))

	return b.String()
}

// renderToolProgress renders a single tool's progress block with step rows.
func renderToolProgress(tool ProgressTool) string {
	var b strings.Builder

	// Tool name header.
	b.WriteString(styles.Subtitle().Render(fmt.Sprintf("  %s", tool.ToolName)))
	b.WriteString("\n")

	// Step rows.
	for _, step := range tool.Steps {
		b.WriteString(renderStepRow(step))
		b.WriteString("\n")
	}

	return b.String()
}

// renderStepRow renders a single step row with appropriate status indicator.
func renderStepRow(step ProgressStep) string {
	prefix := "    " // indent for step rows

	switch step.Status {
	case StepPending:
		return fmt.Sprintf("%s%s %s",
			prefix,
			styles.Muted().Render("[ ]"),
			styles.Muted().Render(step.Name))

	case StepRunning:
		// Use the first spinner frame as a static indicator.
		spinner := styles.Accent().Render("⠋")
		return fmt.Sprintf("%s%s %s",
			prefix,
			spinner,
			styles.Body().Render(step.Name))

	case StepDone:
		return fmt.Sprintf("%s%s %s",
			prefix,
			styles.Success().Render("[✓]"),
			styles.Success().Render(step.Name))

	case StepFailed:
		errorLine := fmt.Sprintf("%s%s %s",
			prefix,
			styles.Error().Render("[✗]"),
			styles.Error().Render(step.Name))
		if step.Error != "" {
			errorLine += "\n" + fmt.Sprintf("%s      %s",
				prefix,
				styles.Error().Render(step.Error))
		}
		return errorLine

	default:
		return fmt.Sprintf("%s%s %s",
			prefix,
			styles.Muted().Render("[?]"),
			styles.Muted().Render(step.Name))
	}
}

// InstallProgressUpdate processes messages for the Install Progress screen.
// completedCount is the number of fully-finished tools. failedCount is the
// number of tools that encountered a critical failure. totalCount is the
// total number of tools being installed.
//
// Returns an action: "success" (all done, no failures), "fail" (some failed),
// "quit" (user quit), or "" (continue).
func InstallProgressUpdate(msg tea.Msg, completedCount, failedCount, totalCount int) string {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.Type == tea.KeyCtrlC {
			return "quit"
		}
		return ""

	case model.ProgressMsg:
		// Progress messages update the display state but don't trigger
		// screen transitions by themselves. The caller accumulates progress
		// and re-evaluates completion after each message.
		_ = msg // consumed by caller for state updates
		return ""
	}

	// Auto-transition check: all tools either completed or failed.
	finished := completedCount + failedCount
	if finished >= totalCount && totalCount > 0 {
		if failedCount > 0 {
			return "fail"
		}
		return "success"
	}

	return ""
}

// ApplyProgressMsg updates a ProgressTool's step based on a ProgressMsg.
// Returns the updated tools slice, the new completedCount, and whether
// this message introduced a new failure.
func ApplyProgressMsg(tools []ProgressTool, msg model.ProgressMsg) ([]ProgressTool, int, bool) {
	newFailed := false
	completedCount := 0

	for i := range tools {
		// Match by ToolID first, fall back to ToolName for backwards compat.
		if tools[i].ToolID != msg.ToolID && tools[i].ToolName != msg.ToolID && !toolNameMatches(tools[i], msg.ToolID) {
			continue
		}

		// Find the matching step.
		for j := range tools[i].Steps {
			if tools[i].Steps[j].Name == msg.Step {
				if msg.Error != "" {
					tools[i].Steps[j].Status = StepFailed
					tools[i].Steps[j].Error = msg.Error
					newFailed = true
				} else if msg.Done {
					tools[i].Steps[j].Status = StepDone
				} else {
					tools[i].Steps[j].Status = StepRunning
				}
				break
			}
		}
	}

	// Count completed tools (all steps done).
	for _, tool := range tools {
		if allStepsDone(tool.Steps) {
			completedCount++
		}
	}

	return tools, completedCount, newFailed
}

// toolNameMatches checks if a ProgressTool's name matches a tool ID.
func toolNameMatches(tool ProgressTool, toolID string) bool {
	return tool.ToolName == toolID
}

// allStepsDone returns true when every step has status StepDone.
func allStepsDone(steps []ProgressStep) bool {
	for _, s := range steps {
		if s.Status != StepDone {
			return false
		}
	}
	return len(steps) > 0
}
