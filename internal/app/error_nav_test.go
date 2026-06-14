// Package app — error screen PreviousScreen navigation tests (REQ-TUI-04).
//
// Verifies that the Error screen's Esc/Left handler returns a NavigateMsg
// whose target is the source screen recorded in m.PreviousScreen, NOT the
// hardcoded model.ScreenToolSelection. This is the source-aware back nav
// pattern already used by the Uninstall screen at update.go:217.
package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/internal/app"
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

// newErrorModel builds a Model placed on the Error screen with the given
// previous screen. Used by the two tests in this file.
func newErrorModel(t *testing.T, prev model.Screen) app.Model {
	t.Helper()
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Screen = model.ScreenError
	m.PreviousScreen = prev
	return m
}

// TestErrorScreen_EscUsesPreviousScreen — REQ-TUI-04 (discriminating case).
//
// GIVEN m.PreviousScreen == ScreenUninstall (NOT the hardcoded
// ScreenToolSelection) AND m.Screen == ScreenError
// WHEN  the user presses Esc
// THEN  the returned cmd produces tui.NavigateMsg{Target: ScreenUninstall}
//
// RED on main: the inline Error handler hardcodes model.ScreenToolSelection
// at update.go:128, so the assertion nav.Target == ScreenUninstall fails.
func TestErrorScreen_EscUsesPreviousScreen(t *testing.T) {
	m := newErrorModel(t, model.ScreenUninstall)

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "Esc on Error screen should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Esc on Error should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenUninstall, nav.Target,
		"Esc on Error should navigate to PreviousScreen (ScreenUninstall), not the hardcoded ScreenToolSelection")
}

// TestErrorScreen_LeftUsesPreviousScreen — REQ-TUI-04 (companion: KeyLeft).
//
// Same discriminating setup as the Esc test, but sends tea.KeyLeft.
// The Error handler's switch case at update.go:126 covers BOTH Esc and
// Left, so this test pins both branches of the same case.
func TestErrorScreen_LeftUsesPreviousScreen(t *testing.T) {
	m := newErrorModel(t, model.ScreenUninstall)

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "Left on Error screen should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Left on Error should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenUninstall, nav.Target,
		"Left on Error should navigate to PreviousScreen (ScreenUninstall), not the hardcoded ScreenToolSelection")
}

// TestErrorScreen_EscFromToolSelectionPreviousScreenStillWorks — REQ-TUI-04 (contrast).
//
// GIVEN m.PreviousScreen == ScreenToolSelection (the case that happens to
// match the OLD hardcoded behavior) AND m.Screen == ScreenError
// WHEN  the user presses Esc
// THEN  navigation still targets ScreenToolSelection
//
// This test guards the common case: when a user arrived at Error from
// ToolSelection, the new source-aware back nav must still route back to
// ToolSelection. The fix changes the dispatch from a hardcoded value to
// m.PreviousScreen, and when PreviousScreen == ToolSelection the result
// is identical. The first two tests prove the new behavior actually fires;
// this test proves we did not regress the dominant path.
func TestErrorScreen_EscFromToolSelectionPreviousScreenStillWorks(t *testing.T) {
	m := newErrorModel(t, model.ScreenToolSelection)

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "Esc on Error screen should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Esc on Error should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenToolSelection, nav.Target,
		"when PreviousScreen is ToolSelection, the new source-aware handler must still route there (no regression of the common case)")
}
