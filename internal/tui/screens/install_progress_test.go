package screens_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/screens"
)

func TestInstallProgressView_ShowsPendingSteps(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepPending},
				{Name: "Commands", Status: screens.StepPending},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}

	view := screens.InstallProgressView(tools, 0, 1, "")

	// All steps should show pending indicator.
	assert.Contains(t, view, "[ ] Skills", "Pending Skills step should show [ ]")
	assert.Contains(t, view, "[ ] Commands", "Pending Commands step should show [ ]")
	assert.Contains(t, view, "[ ] System Prompt", "Pending System Prompt step should show [ ]")

	// Tool name should be visible.
	assert.Contains(t, view, "Claude Code", "Tool name should be displayed")
}

func TestInstallProgressView_ShowsRunningStep(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepRunning},
				{Name: "Commands", Status: screens.StepPending},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}

	view := screens.InstallProgressView(tools, 0, 1, "")

	// Running step should show a spinner-like indicator (⋯ or similar animation marker).
	hasSpinner := strings.Contains(view, "⋯") || strings.Contains(view, "⠋") ||
		strings.Contains(view, "⠙") || strings.Contains(view, "⠹") || strings.Contains(view, "⠸") ||
		strings.Contains(view, "⠼") || strings.Contains(view, "⠴") || strings.Contains(view, "⠦") ||
		strings.Contains(view, "⠧") || strings.Contains(view, "⠇") || strings.Contains(view, "⠏") ||
		strings.Contains(view, "…") || strings.Contains(view, "*") || strings.Contains(view, "|") ||
		strings.Contains(view, "/") || strings.Contains(view, "-") || strings.Contains(view, "\\")
	assert.True(t, hasSpinner, "Running step should show a spinner/animation indicator, got: %s", view)

	// Skills step should not show [ ] or [✓] — it's running.
	// It should show a different indicator.
	assert.NotContains(t, view, "[ ] Skills", "Running Skills step should not show [ ]")
	assert.NotContains(t, view, "[✓] Skills", "Running Skills step should not show [✓] yet")
}

func TestInstallProgressView_ShowsCompletedStep(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepDone},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}

	view := screens.InstallProgressView(tools, 0, 1, "")

	// Done steps should show checkmark.
	assert.Contains(t, view, "[✓] Skills", "Completed Skills step should show [✓]")
	assert.Contains(t, view, "[✓] Commands", "Completed Commands step should show [✓]")

	// Pending step should still show [].
	assert.Contains(t, view, "[ ] System Prompt", "Pending step should still show [ ]")
}

func TestInstallProgressView_ShowsFailedStep(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "OpenCode",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepFailed, Error: "permission denied"},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}

	view := screens.InstallProgressView(tools, 0, 1, "")

	// Failed step should show error indicator and message.
	assert.Contains(t, view, "[✗] Commands", "Failed step should show [✗]")
	assert.Contains(t, view, "permission denied", "Failed step should show error message")
}

func TestInstallProgressView_ShowsMultiToolProgress(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepRunning},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
		{
			ToolName: "OpenCode",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepDone},
				{Name: "System Prompt", Status: screens.StepDone},
			},
		},
	}

	view := screens.InstallProgressView(tools, 1, 2, "")

	// Both tool names should appear.
	assert.Contains(t, view, "Claude Code", "Should show first tool name")
	assert.Contains(t, view, "OpenCode", "Should show second tool name")

	// Overall progress should be indicated.
	assert.Contains(t, view, "1 of 2", "Should show progress count")
}

func TestInstallProgressView_ShowsOverallProgress(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepDone},
				{Name: "System Prompt", Status: screens.StepDone},
			},
		},
		{
			ToolName: "OpenCode",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepRunning},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
		{
			ToolName: "Gemini CLI",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepPending},
				{Name: "Commands", Status: screens.StepPending},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}

	view := screens.InstallProgressView(tools, 1, 3, "")
	assert.Contains(t, view, "Installing", "Should indicate installation is in progress")
	assert.Contains(t, view, "1 of 3", "Should show overall progress")
}

func TestInstallProgressView_AllDone(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepDone},
				{Name: "System Prompt", Status: screens.StepDone},
			},
		},
	}

	view := screens.InstallProgressView(tools, 1, 1, "")

	// All steps should show [✓].
	assert.Contains(t, view, "[✓] Skills", "All-done Skills should show [✓]")
	assert.Contains(t, view, "[✓] Commands", "All-done Commands should show [✓]")
	assert.Contains(t, view, "[✓] System Prompt", "All-done System Prompt should show [✓]")
}

func TestInstallProgressView_InstallModeShowsInstalling(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepRunning},
			},
		},
	}

	view := screens.InstallProgressView(tools, 0, 1, "install")
	assert.Contains(t, view, "Installing", "install mode should show 'Installing' title")
	assert.Contains(t, view, "Installing 0 of 1", "install mode should show 'Installing N of M' summary")
}

func TestInstallProgressView_UninstallModeShowsUninstalling(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "OpenCode",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepRunning},
			},
		},
	}

	view := screens.InstallProgressView(tools, 0, 1, "uninstall")
	assert.Contains(t, view, "Uninstalling", "uninstall mode should show 'Uninstalling' title")
	assert.Contains(t, view, "Uninstalling 0 of 1", "uninstall mode should show 'Uninstalling N of M' summary")
}

func TestInstallProgressView_EmptyModeDefaultsToInstalling(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepRunning},
			},
		},
	}

	view := screens.InstallProgressView(tools, 0, 1, "")
	assert.Contains(t, view, "Installing", "empty mode should default to 'Installing' title")
}

func TestInstallProgressUpdate_SuccessAutoTransitions(t *testing.T) {
	t.Parallel()

	// When all tools are complete, the update should return "success".
	action := screens.InstallProgressUpdate(nil, 3, 0, 3)
	assert.Equal(t, "success", action, "All tools complete should trigger success transition")
}

func TestInstallProgressUpdate_FailureAutoTransitions(t *testing.T) {
	t.Parallel()

	// When all tools are done (2 completed, 1 failed), transition to error.
	action := screens.InstallProgressUpdate(nil, 2, 1, 3)
	assert.Equal(t, "fail", action, "Failed tool should trigger fail transition")
}

func TestInstallProgressUpdate_InProgressReturnsEmpty(t *testing.T) {
	t.Parallel()

	// When some tools are still in progress, no transition.
	action := screens.InstallProgressUpdate(nil, 1, 0, 3)
	assert.Empty(t, action, "In-progress should not trigger transition")
}

func TestInstallProgressUpdate_QReturnsQuit(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	action := screens.InstallProgressUpdate(msg, 0, 0, 1)
	assert.Equal(t, "quit", action, "q should quit from progress screen")
}

func TestInstallProgressView_Golden_Standard(t *testing.T) {
	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepRunning},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
		{
			ToolName: "OpenCode",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepDone},
				{Name: "System Prompt", Status: screens.StepDone},
			},
		},
	}
	view := screens.InstallProgressView(tools, 1, 2, "")

	golden := goldenPath("install_progress_standard.txt")
	if updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0755))
		require.NoError(t, os.WriteFile(golden, []byte(view), 0644))
		t.Logf("updated golden file: %s", golden)
		return
	}

	expected, err := os.ReadFile(golden)
	require.NoError(t, err, "golden file missing — run with UPDATE_GOLDEN=1 to generate")
	assert.Equal(t, string(expected), view, "golden file mismatch — run with UPDATE_GOLDEN=1 to regenerate")
}

func TestInstallProgressView_Golden_Error(t *testing.T) {
	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepFailed, Error: "mkdir /home/user/.claude: permission denied"},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}
	view := screens.InstallProgressView(tools, 0, 1, "")

	golden := goldenPath("install_progress_error.txt")
	if updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0755))
		require.NoError(t, os.WriteFile(golden, []byte(view), 0644))
		t.Logf("updated golden file: %s", golden)
		return
	}

	expected, err := os.ReadFile(golden)
	require.NoError(t, err, "golden file missing — run with UPDATE_GOLDEN=1 to generate")
	assert.Equal(t, string(expected), view, "golden file mismatch — run with UPDATE_GOLDEN=1 to regenerate")
}

func TestInstallProgressUpdate_ProgressMsgUpdatesState(t *testing.T) {
	t.Parallel()

	// A ProgressMsg should be accepted by the update function.
	// In PR 3, the state management is in the app layer, but the screen
	// update function should return the ProgressMsg data so the caller can update state.
	progressMsg := model.ProgressMsg{
		ToolID: "claude-code",
		Step:   "Skills",
		Done:   true,
		Error:  "",
	}

	// Verify ProgressMsg is a valid tea.Msg (it implements the interface).
	var _ tea.Msg = progressMsg

	// The update function should handle ProgressMsg and return no action (not done yet).
	action := screens.InstallProgressUpdate(progressMsg, 0, 0, 1)
	assert.Empty(t, action, "Single step completion should not trigger screen transition")
}

func TestInstallProgressView_NonEmptyView(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{Steps: []screens.ProgressStep{
			{Name: "Skills", Status: screens.StepPending},
		}},
	}
	view := screens.InstallProgressView(tools, 0, 1, "")

	assert.NotEmpty(t, view, "Progress view should not be empty")
	lines := strings.Split(strings.TrimSpace(view), "\n")
	assert.GreaterOrEqual(t, len(lines), 3, "Progress view should span at least 3 lines")
}
