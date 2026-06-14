// Package tui_test — Router dispatch tests (REQ-TUI-07).
//
// These tests verify the Router.DispatchKey method routes key messages to
// the correct per-screen handler and returns the expected tea.Cmd. They
// use a real *app.Model as the KeyHandler (which satisfies the interface
// via wrapper methods in internal/app/model.go).
//
// RED on main / PR6: the Router type does not exist yet, so the test file
// fails to compile. The compile error is the RED signal: the test references
// production code that is not present.
//
// GREEN after PR7 commit 2 (Router shell + KeyHandler interface): the file
// compiles, the Router exists, and DispatchKey delegates to updateScreenKey
// for behavior preservation. All tests pass.
//
// After PR7 commit 10 (delete updateScreenKey): tests still pass because
// the Router's per-screen methods now handle dispatch directly.
package tui_test

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

func TestNavigateMsg_StillExists(t *testing.T) {
	t.Parallel()

	// Verify NavigateMsg type is defined and can be created.
	msg := tui.NavigateMsg{Target: model.ScreenWelcome}

	require.Equal(t, model.ScreenWelcome, msg.Target,
		"NavigateMsg.Target should be ScreenWelcome")

	// Verify it can target other screens too.
	msg2 := tui.NavigateMsg{Target: model.ScreenToolSelection}
	require.Equal(t, model.ScreenToolSelection, msg2.Target,
		"NavigateMsg should support other screen targets")
}

// TestRouter_DispatchKey_WelcomeEnter_NavigatesToToolSelection — REQ-TUI-07
//
// GIVEN m.Screen == ScreenWelcome
// WHEN  the user presses Enter (Welcome's "install" action)
// THEN  Router.DispatchKey returns a cmd that produces
//
//	NavigateMsg{Target: ScreenToolSelection}
func TestRouter_DispatchKey_WelcomeEnter_NavigatesToToolSelection(t *testing.T) {
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Screen = model.ScreenWelcome

	router := tui.NewRouter()
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := router.DispatchKey(&m, msg)

	require.NotNil(t, cmd, "Enter on Welcome should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Welcome Enter should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenToolSelection, nav.Target,
		"Welcome Enter should navigate to ScreenToolSelection")
}

// TestRouter_DispatchKey_ToolSelectionEsc_NavigatesToWelcome — REQ-TUI-07
//
// GIVEN m.Screen == ScreenToolSelection
// WHEN  the user presses Esc (ToolSelection's "back" action)
// THEN  Router.DispatchKey returns a cmd that produces
//
//	NavigateMsg{Target: ScreenWelcome}
func TestRouter_DispatchKey_ToolSelectionEsc_NavigatesToWelcome(t *testing.T) {
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Screen = model.ScreenToolSelection

	router := tui.NewRouter()
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := router.DispatchKey(&m, msg)

	require.NotNil(t, cmd, "Esc on ToolSelection should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Esc on ToolSelection should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenWelcome, nav.Target,
		"Esc on ToolSelection should navigate to ScreenWelcome")
}

// TestRouter_DispatchKey_InstallProgressQ_Quits — REQ-TUI-07
//
// GIVEN m.Screen == ScreenInstallProgress
// WHEN  the user presses q
// THEN  Router.DispatchKey returns tea.Quit (the package-level sentinel
//
//	that signals the Bubbletea runtime to exit).
func TestRouter_DispatchKey_InstallProgressQ_Quits(t *testing.T) {
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Screen = model.ScreenInstallProgress

	router := tui.NewRouter()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := router.DispatchKey(&m, msg)

	// tea.Quit is a package-level function; we assert it is non-nil and
	// does not panic when invoked. The original updateScreenKey returns
	// the package-level tea.Quit; the Router must preserve that.
	require.NotNil(t, cmd, "q on InstallProgress should produce tea.Quit")
	// tea.Quit when called returns tea.QuitMsg{} (not nil) — that is the
	// Bubbletea signal to exit. The test asserts the function produces
	// a QuitMsg.
	result := cmd()
	_, isQuit := result.(tea.QuitMsg)
	assert.True(t, isQuit, "tea.Quit should produce tea.QuitMsg, got %T", result)
}

// TestRouter_DispatchKey_CompleteLeft_NavigatesToPreviousScreen — REQ-TUI-07
//
// GIVEN m.Screen == ScreenComplete
//       AND m.PreviousScreen == ScreenUninstall
// WHEN  the user presses ← (tea.KeyLeft)
// THEN  Router.DispatchKey returns a cmd that produces
//
//	NavigateMsg{Target: ScreenUninstall}
func TestRouter_DispatchKey_CompleteLeft_NavigatesToPreviousScreen(t *testing.T) {
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Screen = model.ScreenComplete
	m.PreviousScreen = model.ScreenUninstall

	router := tui.NewRouter()
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	_, cmd := router.DispatchKey(&m, msg)

	require.NotNil(t, cmd, "← on Complete should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "← on Complete should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenUninstall, nav.Target,
		"← on Complete should navigate to PreviousScreen")
}

// TestRouter_DispatchKey_ErrorLeft_NavigatesToToolSelection — REQ-TUI-07
//
// GIVEN m.Screen == ScreenError
// WHEN  the user presses ← (tea.KeyLeft)
// THEN  Router.DispatchKey returns a cmd that produces
//
//	NavigateMsg{Target: ScreenToolSelection} (PR4 contract).
//
// Note: PR4 changed this to use m.PreviousScreen in the production code,
// but the test exercises the REQ-TUI-07 dispatch surface; the test only
// asserts that the router returns a NavigateMsg whose target is the
// expected screen. The full source-aware back nav is tested in
// internal/app/error_nav_test.go.
func TestRouter_DispatchKey_ErrorLeft_NavigatesToToolSelection(t *testing.T) {
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Screen = model.ScreenError
	m.PreviousScreen = model.ScreenToolSelection

	router := tui.NewRouter()
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	_, cmd := router.DispatchKey(&m, msg)

	require.NotNil(t, cmd, "← on Error should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "← on Error should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenToolSelection, nav.Target,
		"← on Error should navigate to the expected screen")
}

// TestRouter_DispatchKey_StatusBack_NavigatesToWelcome — REQ-TUI-07
//
// GIVEN m.Screen == ScreenStatus
// WHEN  the user presses Esc (Status's "back" action)
// THEN  Router.DispatchKey returns a cmd that produces
//
//	NavigateMsg{Target: ScreenWelcome}
func TestRouter_DispatchKey_StatusBack_NavigatesToWelcome(t *testing.T) {
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Screen = model.ScreenStatus

	router := tui.NewRouter()
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := router.DispatchKey(&m, msg)

	require.NotNil(t, cmd, "Esc on Status should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Esc on Status should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenWelcome, nav.Target,
		"Esc on Status should navigate to ScreenWelcome")
}

// TestRouter_DispatchKey_UninstallEsc_CancelsConfirmation — REQ-TUI-07
//
// GIVEN m.Screen == ScreenUninstall
//       AND m.UninstallConfirming == true
// WHEN  the user presses Esc
// THEN  Router.DispatchKey returns nil (cancels the confirmation prompt)
//
// The Esc handler in confirmation mode returns (m, nil) — no navigation,
// just clears the confirmation state. The full state transition is
// verified by the existing model_test.go tests; this Router test
// confirms the dispatch surfaces the right contract.
//
// Note: the state is checked on the RETURNED model (m2), not the input
// m. This is because the legacy updateScreenKey has a value receiver
// and returns a modified copy. After commits 3-9, the per-screen
// handler will mutate via the pointer so the local m will also reflect
// the change; the assertion is written to be correct in both regimes
// by checking the returned model.
func TestRouter_DispatchKey_UninstallEsc_CancelsConfirmation(t *testing.T) {
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Screen = model.ScreenUninstall
	m.UninstallConfirming = true

	router := tui.NewRouter()
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	m2, cmd := router.DispatchKey(&m, msg)

	assert.Nil(t, cmd, "Esc on Uninstall confirmation should produce nil cmd")
	// Verify the state was cleared on the returned model.
	require.NotNil(t, m2, "DispatchKey should return a non-nil model")
	modelVal, ok := m2.(app.Model)
	require.True(t, ok, "DispatchKey should return app.Model, got %T", m2)
	assert.False(t, modelVal.UninstallConfirming,
		"Esc on Uninstall confirmation should clear UninstallConfirming")
}
