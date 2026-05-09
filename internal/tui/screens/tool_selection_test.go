package screens_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	tea "github.com/charmbracelet/bubbletea"

	"sequoia-ai/internal/model"
	"sequoia-ai/internal/tui/screens"
)

func TestToolSelectionView_ShowsCheckboxes(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code"}, Selected: false},
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode"}, Selected: true},
		{Adapter: &dummyAdapter{id: "gemini", name: "Gemini CLI"}, Selected: false},
	}
	view := screens.ToolSelectionView(tools, 0, "")

	// Check that tools are listed by name.
	for _, ts := range tools {
		assert.Contains(t, view, ts.Adapter.Name(),
			"Tool Selection should list tool %q", ts.Adapter.Name())
	}

	// Check that selection state is visually represented with checkboxes.
	assert.Contains(t, view, "[x]", "Selected tool (OpenCode) should show [x]")
	assert.Contains(t, view, "[ ]", "Unselected tools should show [ ]")

	// Unselected count should be at least 2.
	emptyCount := strings.Count(view, "[ ]")
	assert.GreaterOrEqual(t, emptyCount, 2, "Should have at least 2 unselected checkboxes")
}

func TestToolSelectionView_ShowsSelectionCount(t *testing.T) {
	t.Parallel()

	// 1 of 3 tools selected.
	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code"}, Selected: true},
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode"}, Selected: false},
		{Adapter: &dummyAdapter{id: "gemini", name: "Gemini CLI"}, Selected: false},
	}
	view := screens.ToolSelectionView(tools, 0, "")

	assert.Contains(t, view, "1 of 3", "Should show selection count")
}

func TestToolSelectionView_ShowsErrorWhenPresent(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code"}, Selected: false},
	}
	errorMsg := "Select at least one tool to continue"
	view := screens.ToolSelectionView(tools, 0, errorMsg)

	assert.Contains(t, view, errorMsg, "Should display error message")
}

func TestToolSelectionView_ShowsNavigationHints(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code"}, Selected: false},
	}
	view := screens.ToolSelectionView(tools, 0, "")

	// Should show keyboard hints for navigation.
	hints := []string{"↑/↓", "j/k", "Space", "Enter", "Esc"}
	for _, hint := range hints {
		assert.Contains(t, view, hint,
			"Tool Selection should show hint %q", hint)
	}
}

func TestToolSelectionUpdate_SpaceTogglesSelection(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeySpace}
	newCursor, shouldToggle, action := screens.ToolSelectionUpdate(msg, 0, 3)

	assert.Equal(t, 0, newCursor, "Cursor should not change on Space")
	assert.True(t, shouldToggle, "Space should trigger toggle")
	assert.Empty(t, action, "Space should not trigger navigation action")
}

func TestToolSelectionUpdate_UpArrowMovesCursor(t *testing.T) {
	t.Parallel()

	// Cursor at index 1, pressing up should move to 0.
	msg := tea.KeyMsg{Type: tea.KeyUp}
	newCursor, shouldToggle, action := screens.ToolSelectionUpdate(msg, 1, 3)

	assert.Equal(t, 0, newCursor, "Up arrow should decrement cursor")
	assert.False(t, shouldToggle)
	assert.Empty(t, action)
}

func TestToolSelectionUpdate_DownArrowMovesCursor(t *testing.T) {
	t.Parallel()

	// Cursor at index 0, pressing down should move to 1.
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newCursor, shouldToggle, action := screens.ToolSelectionUpdate(msg, 0, 3)

	assert.Equal(t, 1, newCursor, "Down arrow should increment cursor")
	assert.False(t, shouldToggle)
	assert.Empty(t, action)
}

func TestToolSelectionUpdate_JKKeysMoveCursor(t *testing.T) {
	t.Parallel()

	// 'j' moves down.
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newCursor, _, _ := screens.ToolSelectionUpdate(msg, 0, 3)
	assert.Equal(t, 1, newCursor, "'j' should move cursor down")

	// 'k' moves up.
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newCursor, _, _ = screens.ToolSelectionUpdate(msg, 1, 3)
	assert.Equal(t, 0, newCursor, "'k' should move cursor up")
}

func TestToolSelectionUpdate_WrapsAtTop(t *testing.T) {
	t.Parallel()

	// Cursor at 0, pressing up should wrap to last item.
	msg := tea.KeyMsg{Type: tea.KeyUp}
	newCursor, _, _ := screens.ToolSelectionUpdate(msg, 0, 3)

	assert.Equal(t, 2, newCursor, "Up arrow at top should wrap to last item")
}

func TestToolSelectionUpdate_WrapsAtBottom(t *testing.T) {
	t.Parallel()

	// Cursor at last item, pressing down should wrap to first item.
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newCursor, _, _ := screens.ToolSelectionUpdate(msg, 2, 3)

	assert.Equal(t, 0, newCursor, "Down arrow at bottom should wrap to first item")
}

func TestToolSelectionUpdate_EnterWithSelectionConfirms(t *testing.T) {
	t.Parallel()

	// When tools are selected, Enter should confirm.
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, _, action := screens.ToolSelectionUpdate(msg, 0, 3)

	assert.Equal(t, "confirm", action, "Enter should confirm selection")
}

func TestToolSelectionUpdate_EnterEmptySelectionReturnsBack(t *testing.T) {
	t.Parallel()

	// With 0 tools, Enter should not confirm (caller validates count).
	// The ToolSelectionUpdate itself doesn't know about selection state —
	// it just returns "confirm". The app validates.
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, _, action := screens.ToolSelectionUpdate(msg, 0, 0)

	// With 0 tools, Enter shouldn't confirm — there's nothing to select.
	assert.NotEqual(t, "confirm", action, "Enter with 0 tools cannot confirm")
}

func TestToolSelectionUpdate_EscReturnsBack(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, _, action := screens.ToolSelectionUpdate(msg, 0, 3)

	assert.Equal(t, "back", action, "Esc should return back to Welcome")
}

func TestToolSelectionUpdate_LeftArrowReturnsBack(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	_, _, action := screens.ToolSelectionUpdate(msg, 0, 3)

	assert.Equal(t, "back", action, "Left arrow should return back to Welcome")
}

func TestToolSelectionUpdate_QReturnsQuit(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, _, action := screens.ToolSelectionUpdate(msg, 0, 3)

	assert.Equal(t, "quit", action, "q should quit")
}

func TestToolSelectionView_NonEmptyView(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code"}, Selected: false},
	}
	view := screens.ToolSelectionView(tools, 0, "")

	assert.NotEmpty(t, view, "Tool Selection view should not be empty")
	lines := strings.Split(strings.TrimSpace(view), "\n")
	assert.GreaterOrEqual(t, len(lines), 3, "Tool Selection view should span at least 3 lines")
}
