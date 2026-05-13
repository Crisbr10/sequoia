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

func TestStatusView_ShowsInstalledToolsWithCheckmark(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
	}
	view := screens.StatusView(tools, 0, "en")

	assert.Contains(t, view, "Claude Code", "Status should show tool name")
	assert.Contains(t, view, "✅", "Status should show installed indicator")
}

func TestStatusView_ShowsNotInstalledToolsWithCross(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode", inst: false}, Selected: false},
	}
	view := screens.StatusView(tools, 0, "en")

	assert.Contains(t, view, "OpenCode", "Status should show tool name")
	assert.Contains(t, view, "❌", "Status should show not-installed indicator")
}

func TestStatusView_ShowsVersionAndPath(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true, ver: "v0.1.0", path: "/home/user/.claude"}, Selected: false},
	}
	view := screens.StatusView(tools, 0, "en")

	assert.Contains(t, view, "v0.1.0", "Status should show version")
	assert.Contains(t, view, "/home/user/.claude", "Status should show install path")
}

func TestStatusView_ShowsMixedState(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode", inst: false}, Selected: false},
	}
	view := screens.StatusView(tools, 0, "en")

	// Both checkmark and cross should appear.
	assert.Contains(t, view, "✅", "Status should show installed indicator")
	assert.Contains(t, view, "❌", "Status should show not-installed indicator")
	assert.Contains(t, view, "Claude Code", "Status should list Claude Code")
	assert.Contains(t, view, "OpenCode", "Status should list OpenCode")
}

func TestStatusView_ShowsEmptyMessageWhenNoAdapters(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{}
	view := screens.StatusView(tools, 0, "en")

	assert.Contains(t, view, "No adapters registered", "Status should show empty message when no adapters")
}

func TestStatusView_ShowsKeyHints(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
	}
	view := screens.StatusView(tools, 0, "en")

	assert.Contains(t, view, "u", "Status should show 'u' key hint")
	assert.Contains(t, view, "r", "Status should show 'r' key hint")
	assert.Contains(t, view, "d", "Status should show 'd' key hint")
	assert.Contains(t, view, "q", "Status should show 'q' key hint")
}

func TestStatusView_ShowsNonEmptyView(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
	}
	view := screens.StatusView(tools, 0, "en")

	assert.NotEmpty(t, view, "Status view should not be empty")
	lines := strings.Split(strings.TrimSpace(view), "\n")
	assert.GreaterOrEqual(t, len(lines), 3, "Status view should span at least 3 lines")
}

// StatusUpdate tests.

func TestStatusUpdate_DKeyReturnsUninstall(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newCursor, action := screens.StatusUpdate(msg, 0, 3)

	assert.Equal(t, 0, newCursor, "Cursor should not change on 'd'")
	assert.Equal(t, "uninstall", action, "'d' should return uninstall action")
}

func TestStatusUpdate_RKeyReturnsReinstall(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newCursor, action := screens.StatusUpdate(msg, 0, 3)

	assert.Equal(t, 0, newCursor, "Cursor should not change on 'r'")
	assert.Equal(t, "reinstall", action, "'r' should return reinstall action")
}

func TestStatusUpdate_UKeyReturnsUpdate(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}
	newCursor, action := screens.StatusUpdate(msg, 0, 3)

	assert.Equal(t, 0, newCursor, "Cursor should not change on 'u'")
	assert.Equal(t, "update", action, "'u' should return update action")
}

func TestStatusUpdate_UpArrowMovesCursor(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyUp}
	newCursor, action := screens.StatusUpdate(msg, 1, 3)

	assert.Equal(t, 0, newCursor, "Up arrow should decrement cursor")
	assert.Empty(t, action, "Up arrow should not trigger action")
}

func TestStatusUpdate_DownArrowMovesCursor(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyDown}
	newCursor, action := screens.StatusUpdate(msg, 0, 3)

	assert.Equal(t, 1, newCursor, "Down arrow should increment cursor")
	assert.Empty(t, action, "Down arrow should not trigger action")
}

func TestStatusUpdate_JKKeysMoveCursor(t *testing.T) {
	t.Parallel()

	// 'j' moves down.
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newCursor, _ := screens.StatusUpdate(msg, 0, 3)
	assert.Equal(t, 1, newCursor, "'j' should move cursor down")

	// 'k' moves up.
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newCursor, _ = screens.StatusUpdate(msg, 1, 3)
	assert.Equal(t, 0, newCursor, "'k' should move cursor up")
}

func TestStatusUpdate_WrapsAtTop(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyUp}
	newCursor, _ := screens.StatusUpdate(msg, 0, 3)

	assert.Equal(t, 2, newCursor, "Up arrow at top should wrap to last item")
}

func TestStatusUpdate_WrapsAtBottom(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyDown}
	newCursor, _ := screens.StatusUpdate(msg, 2, 3)

	assert.Equal(t, 0, newCursor, "Down arrow at bottom should wrap to first item")
}

func TestStatusUpdate_UnknownKeyReturnsEmptyAction(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	newCursor, action := screens.StatusUpdate(msg, 0, 3)

	assert.Equal(t, 0, newCursor, "Cursor should not change on unknown key")
	assert.Empty(t, action, "Unknown key should return empty action")
}

func TestStatusUpdate_ZeroToolsEmptyActions(t *testing.T) {
	t.Parallel()

	// With 0 tools, arrow keys should not move cursor.
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newCursor, action := screens.StatusUpdate(msg, 0, 0)

	assert.Equal(t, 0, newCursor, "Cursor should stay at 0 with no tools")
	assert.Empty(t, action, "No action with 0 tools")

	// d/r/u should still produce actions even with 0 tools.
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	_, action = screens.StatusUpdate(msg, 0, 0)
	assert.Equal(t, "uninstall", action, "'d' should return uninstall even with 0 tools")
}

func TestStatusView_Golden_AllInstalled(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true, ver: "v0.1.0", path: "/home/user/.claude"}, Selected: false},
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode", inst: true, ver: "v0.1.0", path: "/home/user/.config/opencode"}, Selected: false},
	}
	view := screens.StatusView(tools, 0, "en")

	golden := goldenPath("status_all_installed.txt")
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

func TestStatusView_Golden_Empty(t *testing.T) {
	tools := []model.ToolState{}
	view := screens.StatusView(tools, 0, "en")

	golden := goldenPath("status_empty.txt")
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

func TestStatusView_Golden_Mixed(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true, ver: "v0.1.0", path: "/home/user/.claude"}, Selected: false},
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode", inst: false}, Selected: false},
	}
	view := screens.StatusView(tools, 0, "en")

	golden := goldenPath("status_mixed.txt")
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
