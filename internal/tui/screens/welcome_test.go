// Package screens_test provides tests for the TUI screen implementations.
package screens_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Crisbr10/sequoia/internal/tui/screens"
)

func TestWelcomeView_ContainsBrandingElements(t *testing.T) {
	t.Parallel()

	version := "v0.1.0"
	view := screens.WelcomeView(version, 0)

	assert.Contains(t, view, "Sequoia", "Welcome screen should contain Sequoia branding")
	assert.Contains(t, view, version, "Welcome screen should display the version")
	assert.Contains(t, view, "navigate", "Welcome screen should show navigation hint")
	assert.Contains(t, view, "quit", "Welcome screen should show quit hint")

	assert.NotEmpty(t, view, "Welcome view should not be empty")
	lines := strings.Split(strings.TrimSpace(view), "\n")
	assert.GreaterOrEqual(t, len(lines), 5, "Welcome view should span at least 5 lines of content")
}

func TestWelcomeView_ShowsMenuOptions(t *testing.T) {
	t.Parallel()

	view := screens.WelcomeView("v0.1.0", 0)

	assert.Contains(t, view, "Install", "Welcome view should list Install option")
	assert.Contains(t, view, "Status", "Welcome view should list Status option")
	assert.Contains(t, view, "Uninstall", "Welcome view should list Uninstall option")
	assert.Contains(t, view, "Quit", "Welcome view should list Quit option")
}

func TestWelcomeView_HighlightsCursor(t *testing.T) {
	t.Parallel()

	// Cursor on Install (0): the Install label should be preceded by the cursor marker.
	view0 := screens.WelcomeView("v0.1.0", 0)
	assert.Contains(t, view0, "▶", "cursor marker should appear when cursor=0")

	// Cursor on Status (1): different visual from cursor=0.
	view1 := screens.WelcomeView("v0.1.0", 1)
	assert.NotEqual(t, view0, view1, "view should differ based on cursor position")
}

// --- WelcomeUpdate ---

func TestWelcomeUpdate_DownKeyIncrementsCursor(t *testing.T) {
	t.Parallel()

	newCursor, action := screens.WelcomeUpdate(tea.KeyMsg{Type: tea.KeyDown}, 0)
	assert.Equal(t, 1, newCursor, "Down key should increment cursor")
	assert.Equal(t, "", action, "Down key should produce no action")
}

func TestWelcomeUpdate_UpKeyDecrementsCursor(t *testing.T) {
	t.Parallel()

	newCursor, action := screens.WelcomeUpdate(tea.KeyMsg{Type: tea.KeyUp}, 2)
	assert.Equal(t, 1, newCursor, "Up key should decrement cursor")
	assert.Equal(t, "", action, "Up key should produce no action")
}

func TestWelcomeUpdate_DownWrapsAround(t *testing.T) {
	t.Parallel()

	// Cursor at last item (Quit = 3) → should wrap to 0.
	newCursor, _ := screens.WelcomeUpdate(tea.KeyMsg{Type: tea.KeyDown}, screens.WelcomeMenuCount-1)
	assert.Equal(t, 0, newCursor, "Down at last item should wrap to 0")
}

func TestWelcomeUpdate_UpWrapsAround(t *testing.T) {
	t.Parallel()

	// Cursor at first item (0) → Up should wrap to last.
	newCursor, _ := screens.WelcomeUpdate(tea.KeyMsg{Type: tea.KeyUp}, 0)
	assert.Equal(t, screens.WelcomeMenuCount-1, newCursor, "Up at first item should wrap to last")
}

func TestWelcomeUpdate_EnterOnInstallReturnsInstall(t *testing.T) {
	t.Parallel()

	_, action := screens.WelcomeUpdate(tea.KeyMsg{Type: tea.KeyEnter}, screens.WelcomeMenuInstall)
	assert.Equal(t, "install", action, "Enter on Install should return 'install'")
}

func TestWelcomeUpdate_EnterOnStatusReturnsStatus(t *testing.T) {
	t.Parallel()

	_, action := screens.WelcomeUpdate(tea.KeyMsg{Type: tea.KeyEnter}, screens.WelcomeMenuStatus)
	assert.Equal(t, "status", action, "Enter on Status should return 'status'")
}

func TestWelcomeUpdate_EnterOnUninstallReturnsUninstall(t *testing.T) {
	t.Parallel()

	_, action := screens.WelcomeUpdate(tea.KeyMsg{Type: tea.KeyEnter}, screens.WelcomeMenuUninstall)
	assert.Equal(t, "uninstall", action, "Enter on Uninstall should return 'uninstall'")
}

func TestWelcomeUpdate_EnterOnQuitReturnsQuit(t *testing.T) {
	t.Parallel()

	_, action := screens.WelcomeUpdate(tea.KeyMsg{Type: tea.KeyEnter}, screens.WelcomeMenuQuit)
	assert.Equal(t, "quit", action, "Enter on Quit should return 'quit'")
}

func TestWelcomeUpdate_JKeyIncrementsCursor(t *testing.T) {
	t.Parallel()

	newCursor, action := screens.WelcomeUpdate(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, 0)
	assert.Equal(t, 1, newCursor, "j key should increment cursor")
	assert.Equal(t, "", action, "j key should produce no action")
}

func TestWelcomeUpdate_KKeyDecrementsCursor(t *testing.T) {
	t.Parallel()

	newCursor, action := screens.WelcomeUpdate(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, 2)
	assert.Equal(t, 1, newCursor, "k key should decrement cursor")
	assert.Equal(t, "", action, "k key should produce no action")
}

func TestWelcomeUpdate_UnknownKeyReturnsNoChange(t *testing.T) {
	t.Parallel()

	newCursor, action := screens.WelcomeUpdate(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, 1)
	assert.Equal(t, 1, newCursor, "Unknown key should not change cursor")
	assert.Equal(t, "", action, "Unknown key should produce no action")
}
