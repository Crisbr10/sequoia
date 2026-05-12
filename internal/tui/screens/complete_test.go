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
	"github.com/Crisbr10/sequoia/internal/tui"
	"github.com/Crisbr10/sequoia/internal/tui/screens"
)

func TestCompleteView_ShowsSuccessHeading(t *testing.T) {
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

	view := screens.CompleteView(tools, "")
	assert.Contains(t, view, "Installation Complete", "Complete screen should show success heading")
}

func TestCompleteView_ListsInstalledTools(t *testing.T) {
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
				{Name: "Commands", Status: screens.StepDone},
				{Name: "System Prompt", Status: screens.StepDone},
			},
		},
	}

	view := screens.CompleteView(tools, "")

	// All tool names should appear.
	assert.Contains(t, view, "Claude Code", "Complete screen should list Claude Code")
	assert.Contains(t, view, "OpenCode", "Complete screen should list OpenCode")

	// Success indicators for installed tools.
	assert.Contains(t, view, "✅", "Complete screen should show success indicators")
}

func TestCompleteView_ShowsWhatWasInstalled(t *testing.T) {
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

	view := screens.CompleteView(tools, "")

	// Should mention what was installed: skills, commands, system prompt.
	assert.Contains(t, view, "Skills", "Complete screen should mention Skills were installed")
	assert.Contains(t, view, "Commands", "Complete screen should mention Commands were installed")
	assert.Contains(t, view, "System Prompt", "Complete screen should mention System Prompt was installed")
}

func TestCompleteView_ShowsFirstCommandHint(t *testing.T) {
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

	view := screens.CompleteView(tools, "")

	// Should show a hint for the first command to try.
	assert.Contains(t, view, "Try running", "Complete screen should show a 'Try running' hint")
}

func TestCompleteView_ShowsKeyHints(t *testing.T) {
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

	view := screens.CompleteView(tools, "")

	// Should show keyboard navigation hints.
	assert.Contains(t, view, "r", "Complete screen should show 'r' key hint")
	assert.Contains(t, view, "q", "Complete screen should show 'q' key hint")
	assert.Contains(t, view, "Status", "Complete screen should hint that r goes to Status")
	assert.Contains(t, view, "Quit", "Complete screen should hint that q quits")
}

func TestCompleteView_InstallModeShowsInstallationComplete(t *testing.T) {
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

	view := screens.CompleteView(tools, "install")
	assert.Contains(t, view, "Installation Complete", "install mode should show 'Installation Complete'")
}

func TestCompleteView_UninstallModeShowsUninstallationComplete(t *testing.T) {
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

	view := screens.CompleteView(tools, "uninstall")
	assert.Contains(t, view, "Uninstallation Complete", "uninstall mode should show 'Uninstallation Complete'")
}

func TestCompleteView_EmptyModeDefaultsToInstallationComplete(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
			},
		},
	}

	view := screens.CompleteView(tools, "")
	assert.Contains(t, view, "Installation Complete", "empty mode should default to 'Installation Complete'")
}

func TestCompleteUpdate_RReturnsNavigateToStatus(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	cmd := screens.CompleteUpdate(msg)

	require.NotNil(t, cmd, "r key should produce a command on Complete screen")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "r should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenStatus, nav.Target,
		"r on Complete should navigate to Status")
}

func TestCompleteUpdate_QReturnsQuit(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	cmd := screens.CompleteUpdate(msg)

	require.NotNil(t, cmd, "q key should produce a command on Complete screen")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "q key should produce tea.QuitMsg, got %T", result)
}

func TestCompleteUpdate_CtrlCReturnsQuit(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	cmd := screens.CompleteUpdate(msg)

	require.NotNil(t, cmd, "ctrl+c should produce a command on Complete screen")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "ctrl+c should produce tea.QuitMsg, got %T", result)
}

func TestCompleteUpdate_UnknownKeyReturnsNil(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	cmd := screens.CompleteUpdate(msg)

	assert.Nil(t, cmd, "Unknown key should produce no command on Complete screen")
}

func TestCompleteView_Golden_AllSucceed(t *testing.T) {
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
				{Name: "Commands", Status: screens.StepDone},
				{Name: "System Prompt", Status: screens.StepDone},
			},
		},
	}
	view := screens.CompleteView(tools, "")

	golden := goldenPath("complete_all_succeed.txt")
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

func TestCompleteView_Golden_PartialSuccess(t *testing.T) {
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
				{Name: "Commands", Status: screens.StepFailed, Error: "permission denied"},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}
	// ProgressTools passed to Complete may include partially failed tools
	// (those that were retried successfully in error→retry flow).
	view := screens.CompleteView(tools, "")

	golden := goldenPath("complete_partial.txt")
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

func TestCompleteView_NonEmptyView(t *testing.T) {
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
	view := screens.CompleteView(tools, "")

	assert.NotEmpty(t, view, "Complete view should not be empty")
	lines := strings.Split(strings.TrimSpace(view), "\n")
	assert.GreaterOrEqual(t, len(lines), 3, "Complete view should span at least 3 lines")
}
