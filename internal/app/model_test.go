package app_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/testutil"
	"github.com/Crisbr10/sequoia/internal/app"
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui"
	"github.com/Crisbr10/sequoia/internal/tui/screens"

	tea "github.com/charmbracelet/bubbletea"
)

// registryMu protects concurrent swaps of the global adapters.DefaultRegistry
// during parallel tests. Tests that mutate DefaultRegistry must lock this
// mutex around the swap+NewModel critical section.
var registryMu sync.Mutex

// installedMock creates a MockAdapter with Detect and IsInstalled returning the given value.
func installedMock(id, name string, installed bool) *testutil.MockAdapter {
	return &testutil.MockAdapter{
		IDVal:           id,
		NameVal:         name,
		DetectFunc:      func() bool { return installed },
		IsInstalledFunc: func() bool { return installed },
	}
}

func TestNewModel_StoresVersion(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "v9.9.9-test")
	assert.Equal(t, "v9.9.9-test", m.Version, "NewModel should store the version string")
}

// TestNewModel_EngramAvailableDefaultsFalse verifies that after construction,
// EngramAvailable is false — detection happens asynchronously via detectEngram().
func TestNewModel_EngramAvailableDefaultsFalse(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	assert.False(t, m.EngramAvailable,
		"EngramAvailable should default to false; detection happens asynchronously via Init()")
}

func TestNewModel_DefaultScreen(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	assert.Equal(t, model.ScreenWelcome, m.Screen, "new model should default to ScreenWelcome")
}

func TestNewModel_PopulatesTools(t *testing.T) {
	// NOT parallel: mutates global adapters.DefaultRegistry.

	// Register a mock adapter so Tools is non-empty.
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})
	reg.Register(&testutil.MockAdapter{IDVal: "opencode", NameVal: "OpenCode"})

	m := app.NewModel("", "test")
	require.Len(t, m.Tools, 2, "Tools should be populated from DefaultRegistry")

	// Verify ToolState wraps the adapter.
	assert.Equal(t, "claude-code", m.Tools[0].Adapter.ID())
	assert.Equal(t, "opencode", m.Tools[1].Adapter.ID())
}

func TestNewModel_ProgressChannel(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	require.NotNil(t, m.Progress, "Progress channel should be allocated")
	assert.Equal(t, 64, cap(m.Progress), "Progress channel buffer capacity should be 64")
}

func TestNewModel_InitReturnsCmd(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	cmd := m.Init()
	// Init now returns a tea.Batch wrapping detectEngram (async engram detection).
	assert.NotNil(t, cmd, "Init should return detecEngram batch command for async detection")
}

func TestModel_ImplementsBubbleteaModel(_ *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	var _ tea.Model = m // compile-time check
	_ = m.Init()
	_, _ = m.Update(nil)
	_ = m.View()
}

func TestWindowSizeMsg_UpdatesDimensions(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	updated, cmd := m.Update(msg)
	require.IsType(t, app.Model{}, updated, "Update should return app.Model")
	m2 := updated.(app.Model)
	assert.Equal(t, 120, m2.Width, "Width should be updated from WindowSizeMsg")
	assert.Equal(t, 40, m2.Height, "Height should be updated from WindowSizeMsg")
	assert.Nil(t, cmd, "WindowSizeMsg should not produce a command")
}

func TestKeyMsg_Q_Quits(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	updated, cmd := m.Update(msg)
	require.NotNil(t, cmd, "pressing q should return a tea.Cmd")
	// tea.Quit() returns a function; calling it gives tea.QuitMsg.
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "pressing q should produce tea.QuitMsg")
	_ = updated
}

func TestKeyMsg_CtrlC_Quits(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	updated, cmd := m.Update(msg)
	require.NotNil(t, cmd, "pressing ctrl+c should return a tea.Cmd")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "pressing ctrl+c should produce tea.QuitMsg")
	_ = updated
}

func TestEmptyModel_CompilesAndRunsWithoutPanic(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")

	// These should not panic.
	require.NotPanics(t, func() {
		_ = m.Init()
	})

	require.NotPanics(t, func() {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	})

	require.NotPanics(t, func() {
		_ = m.View()
	})
}

func TestNavigateMsg_TransitionsScreen(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	assert.Equal(t, model.ScreenWelcome, m.Screen, "initial screen should be Welcome")

	// Send NavigateMsg targeting ToolSelection.
	msg := tui.NavigateMsg{Target: model.ScreenToolSelection}
	updated, cmd := m.Update(msg)
	m2 := updated.(app.Model)
	assert.Equal(t, model.ScreenToolSelection, m2.Screen, "NavigateMsg should transition to ToolSelection")
	assert.Nil(t, cmd, "NavigateMsg should return no command")
}

func TestNavigateMsg_ToComplete_TransitionsScreen(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	// Manually set to InstallProgress to simulate post-install.
	m.Screen = model.ScreenInstallProgress

	msg := tui.NavigateMsg{Target: model.ScreenComplete}
	updated, cmd := m.Update(msg)
	m2 := updated.(app.Model)
	assert.Equal(t, model.ScreenComplete, m2.Screen, "NavigateMsg should transition to Complete")
	assert.Nil(t, cmd, "NavigateMsg should return no command")
}

func TestUnknownMsg_DoesNotPanic(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")

	type unknownMsg struct{}
	require.NotPanics(t, func() {
		_, _ = m.Update(unknownMsg{})
	}, "Update should not panic on unknown message types")
}

func TestWelcomeView_RendersContent(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	view := m.View()

	// Should NOT be the default placeholder — screens are wired.
	assert.NotEqual(t, "Sequoia TUI — screen not yet implemented", view,
		"Welcome screen should render real content, not placeholder")
	assert.Contains(t, view, "Menu",
		"Welcome view should contain the main menu")
}

func TestWelcomeView_EnterNavigatesToToolSelection(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	msg := tea.KeyMsg{Type: tea.KeyEnter}

	updated, cmd := m.Update(msg)
	require.NotNil(t, cmd, "Enter on Welcome should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Enter should produce NavigateMsg")
	assert.Equal(t, model.ScreenToolSelection, nav.Target,
		"Enter on Welcome should navigate to ToolSelection")
	_ = updated
}

func TestToolSelectionView_RendersCheckboxes(t *testing.T) {
	// Register mock adapters, then navigate to ToolSelection.
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})
	reg.Register(&testutil.MockAdapter{IDVal: "opencode", NameVal: "OpenCode"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenToolSelection
	// NewModel("") selects all tools — deselect first to test [ ] rendering.
	m.Tools[0].Selected = false

	view := m.View()
	assert.NotEqual(t, "Sequoia TUI — screen not yet implemented", view,
		"ToolSelection should render real content")
	assert.Contains(t, view, "[ ]",
		"ToolSelection should show unselected checkboxes")
	assert.Contains(t, view, "[x]",
		"ToolSelection should show selected checkboxes")
	assert.Contains(t, view, "Claude Code",
		"ToolSelection should list tools by name")
}

func TestToolSelection_EscNavigatesToWelcome(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	m.Screen = model.ScreenToolSelection
	msg := tea.KeyMsg{Type: tea.KeyEsc}

	updated, cmd := m.Update(msg)
	require.NotNil(t, cmd, "Esc on ToolSelection should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Esc should produce NavigateMsg")
	assert.Equal(t, model.ScreenWelcome, nav.Target,
		"Esc should navigate back to Welcome")
	_ = updated
}

func TestToolSelection_EnterWithNoSelectionShowsError(t *testing.T) {
	// Register one tool, none selected.
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenToolSelection
	// Ensure no tool is selected.
	for i := range m.Tools {
		m.Tools[i].Selected = false
	}

	// Press Enter with no selection.
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := m.Update(msg)

	// Should return nil cmd (error shown inline) and set ErrorMsg.
	assert.Nil(t, cmd, "Enter with no selection should not navigate")
	m2 := updated.(app.Model)
	assert.NotEmpty(t, m2.ErrorMsg, "Error message should be set when no tools selected")
}

func TestToolSelection_EnterWithSelectionNavigatesToConfiguration(t *testing.T) {
	// Register one tool and select it.
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenToolSelection
	// Select the first tool.
	if len(m.Tools) > 0 {
		m.Tools[0].Selected = true
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "Enter with selection should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Enter should produce NavigateMsg")
	assert.Equal(t, model.ScreenConfiguration, nav.Target,
		"Enter with selection should navigate to Configuration")
	_ = updated
}

func TestConfigurationView_RendersLanguageAndPersistence(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenConfiguration
	m.EngramAvailable = true
	m.Config = model.TUIConfig{Language: "en", Persistence: "engram"}

	view := m.View()
	assert.NotEqual(t, "Sequoia TUI — screen not yet implemented", view)
	assert.Contains(t, view, "Language", "Configuration should show language selector")
	assert.Contains(t, view, "Engram", "Configuration should show persistence option")
}

func TestConfiguration_EnterConfirmBuildsProgressAndNavigates(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenConfiguration
	// Select the tool so that confirm builds progress.
	m.Tools[0].Selected = true

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "Enter on Configuration should produce a tea.Cmd")
	m2 := updated.(app.Model)
	require.NotEmpty(t, m2.ProgressTools, "ProgressTools should be built from selected tools")
	assert.Equal(t, "Claude Code", m2.ProgressTools[0].ToolName)
	assert.Equal(t, 0, m2.InstallCompleted)
	assert.Equal(t, 0, m2.InstallFailed)
}

func TestConfiguration_EscGoesBackToToolSelection(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	m.Screen = model.ScreenConfiguration

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "Esc on Configuration should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Esc should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenToolSelection, nav.Target)
	_ = updated
}

func TestInstallProgressView_RendersProgressTable(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenInstallProgress
	m.ProgressTools = []screens.ProgressTool{
		{
			ToolID:   "claude-code",
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepPending},
				{Name: "Commands", Status: screens.StepRunning},
				{Name: "System Prompt", Status: screens.StepDone},
			},
		},
	}
	m.InstallCompleted = 0

	view := m.View()
	assert.NotEqual(t, "Sequoia TUI — screen not yet implemented", view)
	assert.Contains(t, view, "Installing", "InstallProgress should show Installing title")
	assert.Contains(t, view, "Claude Code", "InstallProgress should show tool name")
}

func TestInstallProgress_QQuitsFromProgress(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	m.Screen = model.ScreenInstallProgress

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "q on InstallProgress should produce a command")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "q on InstallProgress should produce tea.QuitMsg")
	_ = updated
}

func TestCompleteView_RendersSuccess(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenComplete
	m.ProgressTools = []screens.ProgressTool{
		{
			ToolID:   "claude-code",
			ToolName: "Claude Code",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepDone},
				{Name: "System Prompt", Status: screens.StepDone},
			},
		},
	}

	view := m.View()
	assert.NotEqual(t, "Sequoia TUI — screen not yet implemented", view)
	assert.Contains(t, view, "Complete", "Complete view should render success")
}

func TestComplete_RKeyNavigatesToStatus(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	m.Screen = model.ScreenComplete

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "r on Complete should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "r should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenStatus, nav.Target)
	_ = updated
}

func TestComplete_QKeyQuits(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	m.Screen = model.ScreenComplete

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "q on Complete should produce a command")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "q on Complete should produce tea.QuitMsg")
	_ = updated
}

func TestStatusView_RendersToolTable(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenStatus

	view := m.View()
	assert.NotEqual(t, "Sequoia TUI — screen not yet implemented", view)
	assert.Contains(t, view, "Claude Code", "Status should show tool name")
	assert.Contains(t, view, "Status", "Status should show title")
}

func TestStatus_DKeyNavigatesToUninstall(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenStatus
	m.Cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "d on Status should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "d should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenUninstall, nav.Target)
	_ = updated
}

func TestStatus_RKeyNavigatesToReinstall(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenStatus
	m.Cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "r on Status should produce a command")
	// startPipeline returns tea.Batch containing navigate + install + poll commands.
	result := cmd()
	batch, ok := result.(tea.BatchMsg)
	require.True(t, ok, "r on Status should produce tea.BatchMsg, got %T", result)
	require.NotEmpty(t, batch, "batch should contain commands")

	// The first command in the batch should be the NavigateMsg.
	navResult := batch[0]()
	nav, ok := navResult.(tui.NavigateMsg)
	require.True(t, ok, "first batched command should produce NavigateMsg, got %T", navResult)
	assert.Equal(t, model.ScreenInstallProgress, nav.Target)

	// Verify the model populated ProgressTools.
	m2 := updated.(app.Model)
	assert.NotEmpty(t, m2.ProgressTools, "reinstall should populate ProgressTools")
	assert.Equal(t, "install", m2.OperationMode, "reinstall should set OperationMode to install")
}

func TestStatus_UKeyNoOp(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenStatus
	m.Cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}
	updated, cmd := m.Update(msg)

	assert.Nil(t, cmd, "u on Status should produce no command (placeholder)")
	_ = updated
}

func TestUninstallView_RendersCheckboxList(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(installedMock("claude-code", "Claude Code", true))

	m := app.NewModel("", "test")
	m.Screen = model.ScreenUninstall
	m.Tools[0].Selected = false

	view := m.View()
	assert.NotEqual(t, "Sequoia TUI — screen not yet implemented", view)
	assert.Contains(t, view, "Uninstall", "Uninstall should show title")
}

func TestUninstall_EnterConfirmsWhenToolSelected(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	// Tool must be installed for confirmation to work.
	reg.Register(installedMock("claude-code", "Claude Code", true))

	m := app.NewModel("", "test")
	m.Screen = model.ScreenUninstall
	m.Tools[0].Selected = true
	m.UninstallConfirming = false

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := m.Update(msg)

	assert.Nil(t, cmd, "Enter on Uninstall should not produce navigation command (confirmation mode)")
	m2 := updated.(app.Model)
	assert.True(t, m2.UninstallConfirming, "Enter should activate confirmation mode")
}

func TestUninstall_SpaceTogglesSelection(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(installedMock("claude-code", "Claude Code", true))

	m := app.NewModel("", "test")
	m.Screen = model.ScreenUninstall
	m.Tools[0].Selected = false

	msg := tea.KeyMsg{Type: tea.KeySpace}
	updated, cmd := m.Update(msg)

	assert.Nil(t, cmd, "Space should not produce a command on Uninstall")
	m2 := updated.(app.Model)
	assert.True(t, m2.Tools[0].Selected, "Space should toggle tool to selected")
}

func TestUninstall_EscGoesBackToStatus(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	m.Screen = model.ScreenUninstall

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "Esc on Uninstall should produce a command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Esc should produce NavigateMsg")
	assert.Equal(t, model.ScreenWelcome, nav.Target)
	_ = updated
}

func TestUninstallConfirm_YConfirmsAndStartsPipeline(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(installedMock("claude-code", "Claude Code", true))

	m := app.NewModel("", "test")
	m.Screen = model.ScreenUninstall
	m.UninstallConfirming = true
	m.Tools[0].Selected = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "y on confirmation should produce a tea.Cmd")
	m2 := updated.(app.Model)
	assert.False(t, m2.UninstallConfirming, "UninstallConfirming should be cleared after y")
	assert.NotEmpty(t, m2.ProgressTools, "ProgressTools should be populated for uninstall")
}

func TestUninstallConfirm_NCancelsConfirmation(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	m.Screen = model.ScreenUninstall
	m.UninstallConfirming = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updated, cmd := m.Update(msg)

	assert.Nil(t, cmd, "n on confirmation should NOT produce a command")
	m2 := updated.(app.Model)
	assert.False(t, m2.UninstallConfirming, "UninstallConfirming should be cleared after n")
}

func TestErrorView_RendersFailedTools(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "fail-tool", NameVal: "Fail Tool"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenError
	m.ProgressTools = []screens.ProgressTool{
		{
			ToolID:   "fail-tool",
			ToolName: "Fail Tool",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepFailed, Error: "disk full"},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}

	view := m.View()
	assert.NotEqual(t, "Sequoia TUI — screen not yet implemented", view)
	assert.Contains(t, view, "Failed", "Error should show failure")
	assert.Contains(t, view, "disk full", "Error should show error message")
}

func TestUpdateScreenMsg_ProgressMsgSuccessTransition(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "test-tool", NameVal: "Test Tool"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenInstallProgress
	m.ProgressTools = []screens.ProgressTool{
		{
			ToolID:   "test-tool",
			ToolName: "Test Tool",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepDone},
				{Name: "System Prompt", Status: screens.StepDone},
			},
		},
	}
	m.InstallCompleted = 0
	m.InstallFailed = 0

	// Send a ProgressMsg that completes the remaining steps.
	msg := model.ProgressMsg{ToolID: "test-tool", Step: "System Prompt", Done: true}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "completing last step should produce a transition command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "completion should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenComplete, nav.Target)

	m2 := updated.(app.Model)
	assert.Equal(t, 1, m2.InstallCompleted, "completed count should be incremented")
}

func TestUpdateScreenMsg_ProgressMsgFailTransition(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "fail-tool", NameVal: "Fail Tool"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenInstallProgress
	m.ProgressTools = []screens.ProgressTool{
		{
			ToolID:   "fail-tool",
			ToolName: "Fail Tool",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepPending},
				{Name: "Commands", Status: screens.StepPending},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}
	m.InstallCompleted = 0
	m.InstallFailed = 0

	// Send a ProgressMsg with an error — this should mark the step as failed.
	msg := model.ProgressMsg{ToolID: "fail-tool", Step: "Skills", Done: true, Error: "disk full"}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "failure should produce a transition command")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "failure should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenError, nav.Target)

	m2 := updated.(app.Model)
	assert.Equal(t, 1, m2.InstallFailed, "failed count should be incremented")
}

func TestUpdateScreenMsg_ProgressMsgContinuesPolling(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "test-tool", NameVal: "Test Tool"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenInstallProgress
	m.ProgressTools = []screens.ProgressTool{
		{
			ToolID:   "test-tool",
			ToolName: "Test Tool",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepPending},
				{Name: "Commands", Status: screens.StepPending},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}

	// Send a ProgressMsg that marks one step as done, but tool isn't fully complete.
	msg := model.ProgressMsg{ToolID: "test-tool", Step: "Skills", Done: true}
	updated, cmd := m.Update(msg)

	require.NotNil(t, cmd, "in-progress message should produce a polling command")
	_ = updated
}

func TestUpdateScreenMsg_NonInstallProgressScreen_NoOp(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	m.Screen = model.ScreenWelcome

	msg := model.ProgressMsg{ToolID: "test", Step: "Skills", Done: true}
	updated, cmd := m.Update(msg)

	assert.Nil(t, cmd, "ProgressMsg on non-InstallProgress screen should be no-op")
	_ = updated
}

func TestView_DefaultPlaceholder(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "test-tool", NameVal: "Test Tool"})

	m := app.NewModel("", "test")
	// Set screen to an invalid value.
	m.Screen = model.Screen(999)

	view := m.View()
	assert.Equal(t, "Sequoia TUI — screen not yet implemented", view,
		"invalid screen should show placeholder text")
}

func TestModel_UninstallConfirmView_ShowsPrompt(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(installedMock("claude-code", "Claude Code", true))

	m := app.NewModel("", "test")
	m.Screen = model.ScreenUninstall
	m.UninstallConfirming = true

	view := m.View()
	assert.Contains(t, view, "y/N", "Uninstall confirmation should show y/N prompt")
	assert.Contains(t, view, "Remove Sequoia", "Uninstall confirmation should mention Remove Sequoia")
}

func TestUpdateScreenKey_StatusQQuits(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	m.Screen = model.ScreenStatus

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "q on Status should produce a command")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "q on Status should produce tea.QuitMsg")
}

func TestUpdateScreenKey_UninstallQQuits(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	m.Screen = model.ScreenUninstall

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "q on Uninstall should produce a command")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "q on Uninstall should produce tea.QuitMsg")
}

func TestUpdateScreenKey_CompleteQQuits(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	m := app.NewModel("", "test")
	m.Screen = model.ScreenComplete

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "q on Complete should produce a command")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "q on Complete should produce tea.QuitMsg")
}

// TestNewModel_NoExecLookPath verifies that NewModel does not block on exec.LookPath.
// The call was moved to an async Bubbletea command (detectEngram) so the TUI
// renders immediately without blocking.
func TestNewModel_NoExecLookPath(t *testing.T) {
	// NOT parallel: reads adapters.DefaultRegistry via NewModel().

	start := time.Now()
	_ = app.NewModel("", "test")
	elapsed := time.Since(start)

	// NewModel should return in well under 100ms when no exec.LookPath blocks.
	// Even on slow CI, 50ms provides generous margin for adapter registration.
	assert.Less(t, elapsed, 50*time.Millisecond,
		"NewModel should not block on exec.LookPath; took %v", elapsed)
}

func TestToolSelection_SpaceToggles(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"})

	m := app.NewModel("", "test")
	m.Screen = model.ScreenToolSelection
	// Initially not selected.
	m.Tools[0].Selected = false

	msg := tea.KeyMsg{Type: tea.KeySpace}
	updated, cmd := m.Update(msg)
	assert.Nil(t, cmd, "Space should not produce navigation command")
	m2 := updated.(app.Model)
	assert.True(t, m2.Tools[0].Selected, "Space should toggle tool to selected")

	// Toggle again.
	updated2, _ := m2.Update(msg)
	m3 := updated2.(app.Model)
	assert.False(t, m3.Tools[0].Selected, "Space should toggle tool back to unselected")
}
