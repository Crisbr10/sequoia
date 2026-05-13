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

func TestUninstallView_ShowsInstalledToolsOnly(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode", inst: false}, Selected: false},
	}
	view := screens.UninstallView(tools, 0, "", "en")

	assert.Contains(t, view, "Claude Code", "Uninstall should show installed Claude Code")
	assert.NotContains(t, view, "OpenCode", "Uninstall should NOT show not-installed OpenCode")
}

func TestUninstallView_ShowsEmptyMessageWhenNothingInstalled(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode", inst: false}, Selected: false},
	}
	view := screens.UninstallView(tools, 0, "", "en")

	assert.Contains(t, view, "Nothing to uninstall", "Uninstall should show empty message")
}

func TestUninstallView_ShowsCheckboxesForSelection(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
		{Adapter: &dummyAdapter{id: "gemini", name: "Gemini CLI", inst: true}, Selected: true},
	}
	view := screens.UninstallView(tools, 0, "", "en")

	// Selected tool should show [x].
	assert.Contains(t, view, "[x]", "Uninstall should show selected checkbox")

	// Unselected tool should show [ ].
	assert.Contains(t, view, "[ ]", "Uninstall should show unselected checkbox")
}

func TestUninstallView_ShowsKeyHints(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
	}
	view := screens.UninstallView(tools, 0, "", "en")

	assert.Contains(t, view, "Space", "Uninstall should show Space key hint")
	assert.Contains(t, view, "Enter", "Uninstall should show Enter key hint")
	assert.Contains(t, view, "q", "Uninstall should show 'q' key hint")
}

func TestUninstallView_ZeroInstalledShowsOnlyQHint(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode", inst: false}, Selected: false},
	}
	view := screens.UninstallView(tools, 0, "", "en")

	// When nothing is installed, Enter/Space should NOT be shown.
	assert.Contains(t, view, "q", "Uninstall should still show 'q' key hint")
	assert.NotContains(t, view, "Space", "Uninstall with nothing installed should not show Space hint")
	assert.NotContains(t, view, "Enter", "Uninstall with nothing installed should not show Enter hint")
}

func TestUninstallView_ShowsNonEmptyView(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
	}
	view := screens.UninstallView(tools, 0, "", "en")

	assert.NotEmpty(t, view, "Uninstall view should not be empty")
	lines := strings.Split(strings.TrimSpace(view), "\n")
	assert.GreaterOrEqual(t, len(lines), 3, "Uninstall view should span at least 3 lines")
}

func TestUninstallView_ShowsError(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
	}

	// View with error message.
	view := screens.UninstallView(tools, 0, "Select at least one installed tool to continue", "en")
	assert.Contains(t, view, "Select at least one installed tool to continue",
		"Uninstall view should render the error message when provided")

	// View without error message should not show placeholder.
	viewNoErr := screens.UninstallView(tools, 0, "", "en")
	assert.NotContains(t, viewNoErr, "Select at least one installed tool",
		"Uninstall view should not show error when message is empty")
}

// UninstallUpdate tests.

func TestUninstallUpdate_SpaceTogglesSelection(t *testing.T) {
	t.Parallel()

	_ = []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
	}
	msg := tea.KeyMsg{Type: tea.KeySpace}
	newCursor, shouldToggle, action := screens.UninstallUpdate(msg, 0, 1)

	assert.Equal(t, 0, newCursor, "Cursor should not change on Space")
	assert.True(t, shouldToggle, "Space should trigger toggle")
	assert.Empty(t, action, "Space should not trigger action")
}

func TestUninstallUpdate_UpArrowMovesCursor(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyUp}
	newCursor, shouldToggle, action := screens.UninstallUpdate(msg, 1, 3)

	assert.Equal(t, 0, newCursor, "Up arrow should decrement cursor")
	assert.False(t, shouldToggle)
	assert.Empty(t, action)
}

func TestUninstallUpdate_DownArrowMovesCursor(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyDown}
	newCursor, shouldToggle, action := screens.UninstallUpdate(msg, 0, 3)

	assert.Equal(t, 1, newCursor, "Down arrow should increment cursor")
	assert.False(t, shouldToggle)
	assert.Empty(t, action)
}

func TestUninstallUpdate_JKKeysMoveCursor(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newCursor, _, _ := screens.UninstallUpdate(msg, 0, 3)
	assert.Equal(t, 1, newCursor, "'j' should move cursor down")

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newCursor, _, _ = screens.UninstallUpdate(msg, 1, 3)
	assert.Equal(t, 0, newCursor, "'k' should move cursor up")
}

func TestUninstallUpdate_WrapsAtTop(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyUp}
	newCursor, _, _ := screens.UninstallUpdate(msg, 0, 3)

	assert.Equal(t, 2, newCursor, "Up arrow at top should wrap to last item")
}

func TestUninstallUpdate_WrapsAtBottom(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyDown}
	newCursor, _, _ := screens.UninstallUpdate(msg, 2, 3)

	assert.Equal(t, 0, newCursor, "Down arrow at bottom should wrap to first item")
}

func TestUninstallUpdate_EnterConfirms(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, _, action := screens.UninstallUpdate(msg, 0, 3)

	assert.Equal(t, "confirm", action, "Enter should confirm selection")
}

func TestUninstallUpdate_EnterWithZeroToolsDoesNotConfirm(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, _, action := screens.UninstallUpdate(msg, 0, 0)

	assert.NotEqual(t, "confirm", action, "Enter with 0 tools should not confirm")
}

func TestUninstallUpdate_EscReturnsBack(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, _, action := screens.UninstallUpdate(msg, 0, 3)

	assert.Equal(t, "back", action, "Esc should return back to Status")
}

func TestUninstallUpdate_LeftArrowReturnsBack(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	_, _, action := screens.UninstallUpdate(msg, 0, 3)

	assert.Equal(t, "back", action, "Left arrow should return back to Status")
}

func TestUninstallUpdate_SpaceRuneToggles(t *testing.T) {
	t.Parallel()

	// Some terminals send space as a rune (' ') instead of tea.KeySpace.
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	newCursor, shouldToggle, action := screens.UninstallUpdate(msg, 0, 1)

	assert.Equal(t, 0, newCursor, "Cursor should not change on space rune")
	assert.True(t, shouldToggle, "Space rune should trigger toggle")
	assert.Empty(t, action, "Space rune should not trigger action")
}

func TestUninstallUpdate_UnknownKeyReturnsEmptyAction(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	_, shouldToggle, action := screens.UninstallUpdate(msg, 0, 3)

	assert.False(t, shouldToggle, "Unknown key should not toggle")
	assert.Empty(t, action, "Unknown key should return empty action")
}

func TestUninstallView_ShowsEscBackHint(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
	}
	view := screens.UninstallView(tools, 0, "", "en")

	assert.Contains(t, view, "Esc back",
		"Uninstall view footer should show 'Esc back' hint for back navigation")
}

func TestUninstallView_Golden_InstalledTools(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode", inst: false}, Selected: false},
		{Adapter: &dummyAdapter{id: "gemini", name: "Gemini CLI", inst: true}, Selected: true},
	}
	view := screens.UninstallView(tools, 0, "", "en")

	golden := goldenPath("uninstall_installed_tools.txt")
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

func TestUninstallView_Golden_NothingInstalled(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode", inst: false}, Selected: false},
	}
	view := screens.UninstallView(tools, 0, "", "en")

	golden := goldenPath("uninstall_nothing_installed.txt")
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
