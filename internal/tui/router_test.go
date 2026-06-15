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
//
//	AND m.PreviousScreen == ScreenUninstall
//
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

// TestRouter_DispatchKey_ErrorLeft_NavigatesToPreviousScreen_SourceAware —
// REQ-TUI-07 source-aware back nav for the Error screen.
//
// GIVEN m.Screen == ScreenError
//
//	AND m.PreviousScreen == ScreenUninstall
//
// WHEN  the user presses ← (tea.KeyLeft)
// THEN  Router.DispatchKey returns a cmd that produces
//
//	NavigateMsg{Target: ScreenUninstall}
//
// This test pins the source-aware behavior that
// TestRouter_DispatchKey_ErrorLeft_NavigatesToToolSelection cannot: that
// test sets m.PreviousScreen = ScreenToolSelection, which coincides with
// the previously hardcoded target, so it passes on the buggy code by
// accident. This test sets m.PreviousScreen = ScreenUninstall to prove
// the Router actually consults the source screen and is not hardcoded.
func TestRouter_DispatchKey_ErrorLeft_NavigatesToPreviousScreen_SourceAware(t *testing.T) {
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Screen = model.ScreenError
	m.PreviousScreen = model.ScreenUninstall

	router := tui.NewRouter()
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	_, cmd := router.DispatchKey(&m, msg)

	require.NotNil(t, cmd, "← on Error should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "← on Error should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenUninstall, nav.Target,
		"← on Error should navigate to PreviousScreen (source-aware)")
}

// TestRouter_DispatchKey_ErrorEsc_NavigatesToPreviousScreen_SourceAware —
// companion for tea.KeyEsc on the Error screen.
//
// GIVEN m.Screen == ScreenError
//
//	AND m.PreviousScreen == ScreenStatus
//
// WHEN  the user presses Esc (tea.KeyEsc)
// THEN  Router.DispatchKey returns a cmd that produces
//
//	NavigateMsg{Target: ScreenStatus}
//
// The handleError switch case groups tea.KeyEsc and tea.KeyLeft into a
// single branch, so this test exercises the second key of that branch
// with a different PreviousScreen value to triangulate the source-aware
// behavior (Uninstall vs Status).
func TestRouter_DispatchKey_ErrorEsc_NavigatesToPreviousScreen_SourceAware(t *testing.T) {
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Screen = model.ScreenError
	m.PreviousScreen = model.ScreenStatus

	router := tui.NewRouter()
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := router.DispatchKey(&m, msg)

	require.NotNil(t, cmd, "Esc on Error should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Esc on Error should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenStatus, nav.Target,
		"Esc on Error should navigate to PreviousScreen (source-aware)")
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
//
//	AND m.UninstallConfirming == true
//
// WHEN  the user presses Esc
// THEN  Router.DispatchKey returns nil (cancels the confirmation prompt)
//
// The Esc handler in confirmation mode returns (h, nil) — no navigation,
// just clears the confirmation state. The full state transition is
// verified by the existing model_test.go tests; this Router test
// confirms the dispatch surfaces the right contract.
//
// PR7 commit 9 note: the per-screen handleUninstall mutates state via
// the KeyHandler interface pointer, so the LOCAL m reflects the state
// change. The returned model is *app.Model (pointer), not app.Model
// (value) — the per-screen handlers return h directly.
func TestRouter_DispatchKey_UninstallEsc_CancelsConfirmation(t *testing.T) {
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Screen = model.ScreenUninstall
	m.UninstallConfirming = true

	router := tui.NewRouter()
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := router.DispatchKey(&m, msg)

	assert.Nil(t, cmd, "Esc on Uninstall confirmation should produce nil cmd")
	// Verify the state was cleared on the local m (mutated via pointer).
	assert.False(t, m.UninstallConfirming,
		"Esc on Uninstall confirmation should clear UninstallConfirming")
}

// -- REQ-TUI-RES-01: table-driven branch tests for the 7 handleX methods ------
//
// The tests below pin every key-dispatch branch of the per-screen handlers
// in Router. They are GREEN-by-design against the existing Router code
// (no production change). The 10 pre-existing TestRouter_DispatchKey_*
// tests above stay untouched; the new tests add coverage for branches
// that were implicit (e.g. handleWelcome on the "status" / "uninstall"
// cursor positions, handleError's r-retry, etc.).
//
// Each sub-case is named with a stable ID (W-1, TS-1, etc.) so failures
// are grep-friendly across the suite.

// toolInfoMockFor returns a minimal model.ToolInfo implementation backed
// by the ID/IsInstalled state. It avoids importing the full adapter
// stack for tests that only need to drive the IsInstalled branch.
type toolInfoMock struct {
	id        string
	installed bool
}

func (t toolInfoMock) ID() string        { return t.id }
func (t toolInfoMock) Name() string      { return t.id }
func (t toolInfoMock) IsInstalled() bool { return t.installed }
func (t toolInfoMock) Status() model.ToolStatus {
	return model.ToolStatus{Installed: t.installed}
}
func (t toolInfoMock) Detect() bool { return t.installed }

func toolInfoMockFor(id string, installed bool) toolInfoMock {
	return toolInfoMock{id: id, installed: installed}
}

// routerToolsWithSentinel builds a model whose Tools slice contains
// a single installed tool with Selected=true (used by tests that need
// a clear "has installed + selected" fixture for handleUninstall /
// handleStatus / handleToolSelection paths). The caller MUST set
// m.Screen after calling this helper — the helper leaves it at the
// NewModel default (ScreenWelcome) so the screen choice stays
// explicit at the call site.
func routerToolsWithSentinel(t *testing.T, installed, selected bool) app.Model {
	t.Helper()
	m := app.NewModel("", "test", adapters.NewRegistry())
	m.Tools = []model.ToolState{
		{Adapter: toolInfoMockFor("tool-a", installed), Selected: selected},
	}
	return m
}

// TestRouter_DispatchKey_Welcome_AllActions — REQ-TUI-RES-01
//
// Exercises the four key branches of handleWelcome (router.go:186):
// W-1 Enter on Status cursor → NavigateMsg{ScreenStatus}
// W-2 Enter on Uninstall cursor → NavigateMsg{ScreenUninstall}
// W-3 Enter on Quit cursor → tea.Quit + m.Quitting=true + m.cancel fired
// W-4 Up → nil cmd, cursor wraps
func TestRouter_DispatchKey_Welcome_AllActions(t *testing.T) {
	router := tui.NewRouter()

	t.Run("W-1_EnterOnStatusCursor_NavigatesToStatus", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenWelcome
		m.Cursor = 1 // WelcomeMenuStatus

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyEnter})

		require.NotNil(t, cmd, "Enter on Status cursor should produce a command")
		result := cmd()
		nav, ok := result.(tui.NavigateMsg)
		require.True(t, ok, "Enter on Status cursor should produce NavigateMsg, got %T", result)
		assert.Equal(t, model.ScreenStatus, nav.Target,
			"Enter on Status cursor should navigate to ScreenStatus")
	})

	t.Run("W-2_EnterOnUninstallCursor_NavigatesToUninstall", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenWelcome
		m.Cursor = 2 // WelcomeMenuUninstall

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyEnter})

		require.NotNil(t, cmd, "Enter on Uninstall cursor should produce a command")
		result := cmd()
		nav, ok := result.(tui.NavigateMsg)
		require.True(t, ok, "Enter on Uninstall cursor should produce NavigateMsg, got %T", result)
		assert.Equal(t, model.ScreenUninstall, nav.Target,
			"Enter on Uninstall cursor should navigate to ScreenUninstall")
	})

	t.Run("W-3_EnterOnQuitCursor_QuitsAndCancels", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenWelcome
		m.Cursor = 3 // WelcomeMenuQuit

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyEnter})

		require.NotNil(t, cmd, "Enter on Quit cursor should produce tea.Quit")
		result := cmd()
		_, isQuit := result.(tea.QuitMsg)
		assert.True(t, isQuit, "Enter on Quit cursor should produce tea.QuitMsg, got %T", result)
		assert.True(t, m.Quitting, "Enter on Quit cursor should set m.Quitting=true")
	})

	t.Run("W-4_UpCursor_MovesWithoutNavigation", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenWelcome
		m.Cursor = 0

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyUp})

		assert.Nil(t, cmd, "Up on Welcome should not produce a command")
		// Cursor wraps from 0 to WelcomeMenuCount-1 (3).
		assert.Equal(t, 3, m.Cursor, "Up at first item should wrap to last")
	})
}

// TestRouter_DispatchKey_ToolSelection_AllBranches — REQ-TUI-RES-01
//
// Exercises the five key branches of handleToolSelection (router.go:228):
// TS-1 Enter with ≥1 selected → StartPipeline("install")
// TS-2 Enter with 0 selected → ErrorMsg + nil cmd
// TS-3 Space → toggles Selected
// TS-4 Up → nil cmd, cursor decrements
// TS-5 unknown rune → nil cmd
func TestRouter_DispatchKey_ToolSelection_AllBranches(t *testing.T) {
	router := tui.NewRouter()

	t.Run("TS-1_EnterWithSelection_StartsPipeline", func(t *testing.T) {
		m := routerToolsWithSentinel(t, false, true) // installed=false, selected=true
		m.Screen = model.ScreenToolSelection

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyEnter})

		require.NotNil(t, cmd, "Enter with selection should produce a pipeline command")
		// StartPipeline also sets OperationMode and builds ProgressTools.
		assert.Equal(t, "install", m.OperationMode, "Enter with selection should set OperationMode=install")
		assert.NotEmpty(t, m.ProgressTools, "Enter with selection should build ProgressTools")
	})

	t.Run("TS-2_EnterWithNoSelection_SetsErrorAndReturnsNil", func(t *testing.T) {
		// installed=false, selected=false — pass Enter with 0 selected.
		m := routerToolsWithSentinel(t, false, false)
		m.Screen = model.ScreenToolSelection

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyEnter})

		assert.Nil(t, cmd, "Enter with no selection should not produce a command")
		assert.Contains(t, m.ErrorMsg, "Select at least one",
			"Enter with no selection should set ErrorMsg mentioning selection")
	})

	t.Run("TS-3_SpaceTogglesSelected", func(t *testing.T) {
		// installed=false, selected=false (initial state).
		m := routerToolsWithSentinel(t, false, false)
		m.Screen = model.ScreenToolSelection

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeySpace})

		assert.Nil(t, cmd, "Space should not produce a navigation command")
		assert.True(t, m.Tools[0].Selected, "Space should toggle Selected to true")
	})

	t.Run("TS-4_UpCursor_MovesWithoutNavigation", func(t *testing.T) {
		// Use 2 tools to give the cursor room to decrement.
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Tools = []model.ToolState{
			{Adapter: toolInfoMockFor("tool-a", false)},
			{Adapter: toolInfoMockFor("tool-b", false)},
		}
		m.Screen = model.ScreenToolSelection
		m.Cursor = 1

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyUp})

		assert.Nil(t, cmd, "Up on ToolSelection should not produce a command")
		assert.Equal(t, 0, m.Cursor, "Up should decrement cursor from 1 to 0")
	})

	t.Run("TS-5_UnknownKey_ReturnsNil", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenToolSelection

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

		assert.Nil(t, cmd, "Unknown key on ToolSelection should not produce a command")
	})
}

// TestRouter_DispatchKey_InstallProgress_TransitionAndIgnored — REQ-TUI-RES-01
//
// Exercises the key branches of handleInstallProgress (router.go:273).
// The router-level surface for this screen is q→QuitCmd, CtrlC→QuitCmd,
// and any other key→nil. The "success" / "fail" auto-transitions are
// driven by ProgressMsg (m.updateScreenMsg) rather than the router, so
// they are not reachable through DispatchKey. This test pins the
// router-visible contract.
func TestRouter_DispatchKey_InstallProgress_TransitionAndIgnored(t *testing.T) {
	router := tui.NewRouter()

	t.Run("IP-1_Q_QuitsViaQuitCmd", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenInstallProgress

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		require.NotNil(t, cmd, "q on InstallProgress should produce a command")
		result := cmd()
		_, isQuit := result.(tea.QuitMsg)
		assert.True(t, isQuit, "q on InstallProgress should produce tea.QuitMsg, got %T", result)
	})

	t.Run("IP-2_CtrlC_QuitsViaQuitCmd", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenInstallProgress

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyCtrlC})

		require.NotNil(t, cmd, "CtrlC on InstallProgress should produce a command")
		result := cmd()
		_, isQuit := result.(tea.QuitMsg)
		assert.True(t, isQuit, "CtrlC on InstallProgress should produce tea.QuitMsg, got %T", result)
	})

	t.Run("IP-3_UnknownKey_ReturnsNil", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenInstallProgress

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

		assert.Nil(t, cmd, "Unknown key on InstallProgress should not produce a command")
	})
}

// TestRouter_DispatchKey_Complete_AllBranches — REQ-TUI-RES-01
//
// Exercises the six key branches of handleComplete (router.go:301):
// C-1 Esc + PreviousScreen=Uninstall → NavigateMsg{ScreenUninstall}
// C-2 Esc + PreviousScreen=Welcome → NavigateMsg{ScreenWelcome} (source-aware)
// C-3 r → NavigateMsg{ScreenStatus}
// C-4 CtrlC → QuitCmd
// C-5 q → QuitCmd
// C-6 unknown → nil
func TestRouter_DispatchKey_Complete_AllBranches(t *testing.T) {
	router := tui.NewRouter()

	t.Run("C-1_Esc_PreviousScreenUninstall_NavigatesToUninstall", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenComplete
		m.PreviousScreen = model.ScreenUninstall

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyEsc})

		require.NotNil(t, cmd, "Esc on Complete should produce a command")
		result := cmd()
		nav, ok := result.(tui.NavigateMsg)
		require.True(t, ok, "Esc on Complete should produce NavigateMsg, got %T", result)
		assert.Equal(t, model.ScreenUninstall, nav.Target,
			"Esc on Complete should navigate to PreviousScreen (Uninstall)")
	})

	t.Run("C-2_Esc_PreviousScreenWelcome_NavigatesToWelcome", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenComplete
		m.PreviousScreen = model.ScreenWelcome

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyEsc})

		require.NotNil(t, cmd, "Esc on Complete should produce a command")
		result := cmd()
		nav, ok := result.(tui.NavigateMsg)
		require.True(t, ok, "Esc on Complete should produce NavigateMsg, got %T", result)
		assert.Equal(t, model.ScreenWelcome, nav.Target,
			"Esc on Complete should navigate to PreviousScreen (Welcome zero-value fallback)")
	})

	t.Run("C-3_R_NavigatesToStatus", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenComplete

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

		require.NotNil(t, cmd, "r on Complete should produce a command")
		result := cmd()
		nav, ok := result.(tui.NavigateMsg)
		require.True(t, ok, "r on Complete should produce NavigateMsg, got %T", result)
		assert.Equal(t, model.ScreenStatus, nav.Target,
			"r on Complete should navigate to ScreenStatus")
	})

	t.Run("C-4_CtrlC_QuitsViaQuitCmd", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenComplete

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyCtrlC})

		require.NotNil(t, cmd, "CtrlC on Complete should produce a command")
		result := cmd()
		_, isQuit := result.(tea.QuitMsg)
		assert.True(t, isQuit, "CtrlC on Complete should produce tea.QuitMsg, got %T", result)
	})

	t.Run("C-5_Q_QuitsViaQuitCmd", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenComplete

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		require.NotNil(t, cmd, "q on Complete should produce a command")
		result := cmd()
		_, isQuit := result.(tea.QuitMsg)
		assert.True(t, isQuit, "q on Complete should produce tea.QuitMsg, got %T", result)
	})

	t.Run("C-6_UnknownKey_ReturnsNil", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenComplete

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

		assert.Nil(t, cmd, "Unknown key on Complete should not produce a command")
	})
}

// TestRouter_DispatchKey_Error_AllBranches — REQ-TUI-RES-01
//
// Exercises the four key branches of handleError (router.go:340):
// E-1 r with OperationMode=install → StartPipeline("install")
// E-2 CtrlC → QuitCmd
// E-3 q → QuitCmd
// E-4 unknown → nil
func TestRouter_DispatchKey_Error_AllBranches(t *testing.T) {
	router := tui.NewRouter()

	t.Run("E-1_R_RetriesWithCurrentOperationMode", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenError
		m.OperationMode = "install"

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

		require.NotNil(t, cmd, "r on Error should produce a pipeline command")
		// StartPipeline("install") rebuilds ProgressTools and OperationMode.
		assert.Equal(t, "install", m.OperationMode, "r on Error should preserve OperationMode=install")
		assert.NotEmpty(t, m.ProgressTools, "r on Error should rebuild ProgressTools")
	})

	t.Run("E-2_CtrlC_QuitsViaQuitCmd", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenError

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyCtrlC})

		require.NotNil(t, cmd, "CtrlC on Error should produce a command")
		result := cmd()
		_, isQuit := result.(tea.QuitMsg)
		assert.True(t, isQuit, "CtrlC on Error should produce tea.QuitMsg, got %T", result)
	})

	t.Run("E-3_Q_QuitsViaQuitCmd", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenError

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		require.NotNil(t, cmd, "q on Error should produce a command")
		result := cmd()
		_, isQuit := result.(tea.QuitMsg)
		assert.True(t, isQuit, "q on Error should produce tea.QuitMsg, got %T", result)
	})

	t.Run("E-4_UnknownKey_ReturnsNil", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenError

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

		assert.Nil(t, cmd, "Unknown key on Error should not produce a command")
	})
}

// TestRouter_DispatchKey_Status_AllActions — REQ-TUI-RES-01
//
// Exercises the four key branches of handleStatus (router.go:374):
// S-1 d → NavigateMsg{ScreenUninstall}
// S-2 r with cursor on installed tool → clears selections + selects cursor + StartPipeline
// S-3 r with no installed tool → ErrorMsg + nil
// S-4 Up → nil cmd, cursor changes
func TestRouter_DispatchKey_Status_AllActions(t *testing.T) {
	router := tui.NewRouter()

	t.Run("S-1_D_NavigatesToUninstall", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenStatus

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

		require.NotNil(t, cmd, "d on Status should produce a command")
		result := cmd()
		nav, ok := result.(tui.NavigateMsg)
		require.True(t, ok, "d on Status should produce NavigateMsg, got %T", result)
		assert.Equal(t, model.ScreenUninstall, nav.Target,
			"d on Status should navigate to ScreenUninstall")
	})

	t.Run("S-2_R_ReinstallOnInstalledTool_ClearsAndSelectsCursor", func(t *testing.T) {
		// Two tools, cursor on the first one (installed=true).
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Tools = []model.ToolState{
			{Adapter: toolInfoMockFor("tool-a", true)},
			{Adapter: toolInfoMockFor("tool-b", false)},
		}
		// Pre-select tool-b to verify the "clear" branch fires.
		m.Tools[1].Selected = true
		m.Screen = model.ScreenStatus
		m.Cursor = 0

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

		require.NotNil(t, cmd, "r on Status with installed tool should produce a pipeline command")
		assert.True(t, m.Tools[0].Selected, "r on installed cursor should select the cursor tool")
		assert.False(t, m.Tools[1].Selected, "r on installed cursor should clear other selections")
	})

	t.Run("S-3_R_ReinstallWithNoInstalledTool_SetsError", func(t *testing.T) {
		// Cursor on a non-installed tool.
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Tools = []model.ToolState{
			{Adapter: toolInfoMockFor("tool-a", false)},
		}
		m.Screen = model.ScreenStatus
		m.Cursor = 0

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

		assert.Nil(t, cmd, "r on Status with no installed tool should not produce a command")
		assert.Contains(t, m.ErrorMsg, "installed tool",
			"r on Status with no installed tool should set ErrorMsg mentioning installed tool")
	})

	t.Run("S-4_UpCursor_MovesWithoutNavigation", func(t *testing.T) {
		// Two tools so the cursor has room to decrement.
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Tools = []model.ToolState{
			{Adapter: toolInfoMockFor("tool-a", true)},
			{Adapter: toolInfoMockFor("tool-b", true)},
		}
		m.Screen = model.ScreenStatus
		m.Cursor = 1

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyUp})

		assert.Nil(t, cmd, "Up on Status should not produce a command")
		assert.Equal(t, 0, m.Cursor, "Up should decrement cursor from 1 to 0")
	})
}

// TestRouter_DispatchKey_Uninstall_ConfirmationAndSelection — REQ-TUI-RES-01
//
// Exercises the six key branches of handleUninstall (router.go:422):
// U-1 confirming=true + y → StartPipeline("uninstall") + confirming=false
// U-2 confirming=true + n → confirming=false + nil
// U-3 Enter with ≥1 installed+selected → confirming=true
// U-4 Enter with 0 installed+selected → ErrorMsg + nil
// U-5 Esc with confirming=false → NavigateMsg{PreviousScreen}
// U-6 Space on installed tool → toggles Selected
func TestRouter_DispatchKey_Uninstall_ConfirmationAndSelection(t *testing.T) {
	router := tui.NewRouter()

	t.Run("U-1_Y_StartsUninstallPipeline", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenUninstall
		m.UninstallConfirming = true

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

		require.NotNil(t, cmd, "y on Uninstall confirmation should produce a pipeline command")
		assert.False(t, m.UninstallConfirming, "y on Uninstall confirmation should clear UninstallConfirming")
		assert.Equal(t, "uninstall", m.OperationMode, "y on Uninstall confirmation should set OperationMode=uninstall")
	})

	t.Run("U-2_N_CancelsConfirmation", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenUninstall
		m.UninstallConfirming = true

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

		assert.Nil(t, cmd, "n on Uninstall confirmation should not produce a command")
		assert.False(t, m.UninstallConfirming, "n on Uninstall confirmation should clear UninstallConfirming")
	})

	t.Run("U-3_EnterWithInstalledSelected_SetsConfirming", func(t *testing.T) {
		// installed=true, selected=true → HasSelectedInstalled() is true.
		m := routerToolsWithSentinel(t, true, true)
		m.Screen = model.ScreenUninstall

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyEnter})

		assert.Nil(t, cmd, "Enter on Uninstall selection should not produce a command")
		assert.True(t, m.UninstallConfirming, "Enter with installed+selected should set UninstallConfirming=true")
	})

	t.Run("U-4_EnterWithNoInstalledSelected_SetsError", func(t *testing.T) {
		// installed=true, selected=false → HasSelectedInstalled() is false.
		m := routerToolsWithSentinel(t, true, false)
		m.Screen = model.ScreenUninstall

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyEnter})

		assert.Nil(t, cmd, "Enter on Uninstall with no installed selected should not produce a command")
		assert.Contains(t, m.ErrorMsg, "installed tool",
			"Enter on Uninstall with no installed selected should set ErrorMsg")
		assert.False(t, m.UninstallConfirming, "Enter with no installed selected should NOT set UninstallConfirming")
	})

	t.Run("U-5_EscNavigatesToPreviousScreen", func(t *testing.T) {
		m := app.NewModel("", "test", adapters.NewRegistry())
		m.Screen = model.ScreenUninstall
		m.PreviousScreen = model.ScreenStatus

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeyEsc})

		require.NotNil(t, cmd, "Esc on Uninstall selection should produce a command")
		result := cmd()
		nav, ok := result.(tui.NavigateMsg)
		require.True(t, ok, "Esc on Uninstall selection should produce NavigateMsg, got %T", result)
		assert.Equal(t, model.ScreenStatus, nav.Target,
			"Esc on Uninstall should navigate to PreviousScreen (source-aware)")
	})

	t.Run("U-6_SpaceTogglesSelectedOnInstalledTool", func(t *testing.T) {
		// installed=true, selected=false (initial).
		m := routerToolsWithSentinel(t, true, false)
		m.Screen = model.ScreenUninstall

		_, cmd := router.DispatchKey(&m, tea.KeyMsg{Type: tea.KeySpace})

		assert.Nil(t, cmd, "Space on Uninstall should not produce a command")
		assert.True(t, m.Tools[0].Selected, "Space on installed tool should toggle Selected to true")
	})
}
