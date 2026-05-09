package screens_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tea "github.com/charmbracelet/bubbletea"

	"sequoia-ai/internal/model"
	"sequoia-ai/internal/tui"
	"sequoia-ai/internal/tui/screens"
)

func TestErrorView_ShowsFailureHeading(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepFailed, Error: "permission denied"},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}

	view := screens.ErrorView(tools)
	assert.Contains(t, view, "Installation Failed", "Error screen should show failure heading")
	assert.Contains(t, view, "❌", "Error screen should show failure indicator")
}

func TestErrorView_ListsSucceededAndFailedTools(t *testing.T) {
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
				{Name: "Commands", Status: screens.StepFailed, Error: "mkdir failed"},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}

	view := screens.ErrorView(tools)

	// Both tool names should appear.
	assert.Contains(t, view, "Claude Code", "Error screen should list Claude Code")
	assert.Contains(t, view, "OpenCode", "Error screen should list OpenCode")

	// Success and failure indicators should both be present.
	assert.Contains(t, view, "✅", "Error screen should show success indicator for succeeded tools")
	assert.Contains(t, view, "❌", "Error screen should show failure indicator for failed tools")
}

func TestErrorView_ShowsErrorMessagesForFailedTools(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "OpenCode",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepFailed, Error: "mkdir /home/user/.opencode: permission denied"},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}

	view := screens.ErrorView(tools)

	// The error message from the failed step should be visible.
	assert.Contains(t, view, "permission denied", "Error screen should show the error message")
	assert.Contains(t, view, "Commands", "Error screen should show the failed step name")
}

func TestErrorView_ShowsRetryOption(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepFailed, Error: "network error"},
				{Name: "Commands", Status: screens.StepPending},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}

	view := screens.ErrorView(tools)

	assert.Contains(t, view, "r", "Error screen should show 'r' key hint")
	assert.Contains(t, view, "Retry", "Error screen should hint that r retries")
	assert.Contains(t, view, "q", "Error screen should show 'q' key hint")
}

func TestErrorUpdate_RReturnsNavigateToInstallProgress(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	cmd := screens.ErrorUpdate(msg)

	require.NotNil(t, cmd, "r key should produce a command on Error screen")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "r should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenInstallProgress, nav.Target,
		"r on Error should navigate to InstallProgress for retry")
}

func TestErrorUpdate_EscReturnsBackToToolSelection(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	cmd := screens.ErrorUpdate(msg)

	require.NotNil(t, cmd, "Esc should produce a command on Error screen")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Esc should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenToolSelection, nav.Target,
		"Esc on Error should navigate back to ToolSelection")
}

func TestErrorUpdate_LeftArrowReturnsBackToToolSelection(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	cmd := screens.ErrorUpdate(msg)

	require.NotNil(t, cmd, "Left arrow should produce a command on Error screen")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Left arrow should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenToolSelection, nav.Target,
		"Left arrow on Error should navigate back to ToolSelection")
}

func TestErrorUpdate_QReturnsQuit(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	cmd := screens.ErrorUpdate(msg)

	require.NotNil(t, cmd, "q key should produce a command on Error screen")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "q key should produce tea.QuitMsg, got %T", result)
}

func TestErrorUpdate_CtrlCReturnsQuit(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	cmd := screens.ErrorUpdate(msg)

	require.NotNil(t, cmd, "ctrl+c should produce a command on Error screen")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "ctrl+c should produce tea.QuitMsg, got %T", result)
}

func TestErrorUpdate_UnknownKeyReturnsNil(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	cmd := screens.ErrorUpdate(msg)

	assert.Nil(t, cmd, "Unknown key should produce no command on Error screen")
}

func TestErrorView_Golden_MixedResults(t *testing.T) {
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
				{Name: "Commands", Status: screens.StepFailed, Error: "mkdir /home/user/.opencode: permission denied"},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}
	view := screens.ErrorView(tools)

	golden := goldenPath("error_mixed.txt")
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

func TestErrorView_NonEmptyView(t *testing.T) {
	t.Parallel()

	tools := []screens.ProgressTool{
		{
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepFailed, Error: "network error"},
				{Name: "Commands", Status: screens.StepPending},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}
	view := screens.ErrorView(tools)

	assert.NotEmpty(t, view, "Error view should not be empty")
	lines := strings.Split(strings.TrimSpace(view), "\n")
	assert.GreaterOrEqual(t, len(lines), 3, "Error view should span at least 3 lines")
}
