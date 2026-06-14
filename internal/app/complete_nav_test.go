// Package app_test — Complete screen back-navigation tests (REQ-TUI-06).
//
// Verifies that the Complete screen's ← / Esc handlers emit a NavigateMsg
// whose target is the source screen recorded in m.PreviousScreen, NOT a
// hardcoded value. The Complete dispatch was inlined into updateScreenKey
// by PR1 (REQ-TUI-01) to give the handler access to m.cancel; PR6 extends
// that same inlined block to handle the back-navigation keys, mirroring
// the source-aware pattern already used by the Uninstall screen
// (update.go:217) and added for the Error screen in PR4.
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

// newCompleteModel builds a Model placed on the Complete screen with the
// given previous screen. Used by the three tests in this file.
func newCompleteModel(_ *testing.T, prev model.Screen) app.Model {
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Screen = model.ScreenComplete
	m.PreviousScreen = prev
	return m
}

// TestComplete_LeftArrow_NavigatesToPreviousScreen — REQ-TUI-06 (discriminating case).
//
// GIVEN m.PreviousScreen == ScreenUninstall (NOT a hardcoded target) AND
//
//	m.Screen == ScreenComplete
//
// WHEN  the user presses ← (tea.KeyLeft)
// THEN  the returned cmd produces tui.NavigateMsg{Target: ScreenUninstall}
//
// RED on the PR1 base: the inlined Complete block added by PR1 handles
// only `q`, `Ctrl+C`, and `r`. For `←` it falls through to the default
// `return m, nil`, so the require.NotNil assertion below fails.
func TestComplete_LeftArrow_NavigatesToPreviousScreen(t *testing.T) {
	m := newCompleteModel(t, model.ScreenUninstall)

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "← on Complete screen should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "← on Complete should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenUninstall, nav.Target,
		"← on Complete should navigate to PreviousScreen (ScreenUninstall), not the hardcoded ScreenStatus used by `r`")
}

// TestComplete_Esc_NavigatesToPreviousScreen — REQ-TUI-06 (companion: Esc).
//
// Same discriminating setup as the ← test, but sends tea.KeyEsc. The
// Complete handler must handle both keys in a single case to keep the
// behavior symmetric (Esc and ← are interchangeable back-nav triggers
// throughout the TUI: see update.go:126 for the Error screen's matching
// case and the Welcome/ToolSelection screens' `back` actions).
func TestComplete_Esc_NavigatesToPreviousScreen(t *testing.T) {
	m := newCompleteModel(t, model.ScreenUninstall)

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "Esc on Complete screen should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Esc on Complete should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenUninstall, nav.Target,
		"Esc on Complete should navigate to PreviousScreen (ScreenUninstall), not the hardcoded ScreenStatus used by `r`")
}

// TestComplete_LeftArrow_FromWelcomePreviousScreen_StillWorks — REQ-TUI-06 (contrast).
//
// GIVEN m.PreviousScreen == ScreenWelcome (the most common case: a fresh
//
//	install followed by the user pressing ← to return to the menu)
//
// AND  m.Screen == ScreenComplete
// WHEN  the user presses ←
// THEN  navigation still targets ScreenWelcome
//
// This test guards the common case. When PreviousScreen is ScreenWelcome,
// the new source-aware handler must still route there — proving we did not
// regress the dominant path while adding the new behavior for other
// PreviousScreen values. The first two tests prove the new behavior
// actually fires; this test proves the common path stays correct.
//
// RED on the PR1 base for the same reason as the discriminating tests:
// the inlined Complete block has no case for ←, so the require.NotNil
// assertion below fails.
func TestComplete_LeftArrow_FromWelcomePreviousScreen_StillWorks(t *testing.T) {
	m := newCompleteModel(t, model.ScreenWelcome)

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "← on Complete screen should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "← on Complete should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenWelcome, nav.Target,
		"when PreviousScreen is ScreenWelcome, the new source-aware handler must still route there (no regression of the common case)")
}
