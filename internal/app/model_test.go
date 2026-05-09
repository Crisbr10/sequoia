package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sequoia-ai/adapters"
	"sequoia-ai/internal/app"
	"sequoia-ai/internal/model"
	"sequoia-ai/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

// mockAdapter is a minimal ToolAdapter for model tests.
type mockAdapter struct {
	id   string
	name string
}

func (m *mockAdapter) ID() string                     { return m.id }
func (m *mockAdapter) Name() string                   { return m.name }
func (m *mockAdapter) Detect() bool                   { return false }
func (m *mockAdapter) IsInstalled() bool              { return false }
func (m *mockAdapter) Install() error                 { return nil }
func (m *mockAdapter) Uninstall() error               { return nil }
func (m *mockAdapter) Status() adapters.AdapterStatus { return adapters.AdapterStatus{} }
func (m *mockAdapter) SkillsPath() string             { return "" }
func (m *mockAdapter) CommandsPath() string           { return "" }
func (m *mockAdapter) SystemPromptPath() string       { return "" }
func (m *mockAdapter) PromptStrategy() adapters.PromptStrategy {
	return adapters.StrategyMarkdownSections
}

var _ adapters.ToolAdapter = (*mockAdapter)(nil)

func TestNewModel_DefaultScreen(t *testing.T) {
	t.Parallel()

	m := app.NewModel("")
	assert.Equal(t, model.ScreenWelcome, m.Screen, "new model should default to ScreenWelcome")
}

func TestNewModel_PopulatesTools(t *testing.T) {
	t.Parallel()

	// Register a mock adapter so Tools is non-empty.
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original }()

	reg.Register(&mockAdapter{id: "claude-code", name: "Claude Code"})
	reg.Register(&mockAdapter{id: "opencode", name: "OpenCode"})

	m := app.NewModel("")
	require.Len(t, m.Tools, 2, "Tools should be populated from DefaultRegistry")

	// Verify ToolState wraps the adapter.
	assert.Equal(t, "claude-code", m.Tools[0].Adapter.ID())
	assert.Equal(t, "opencode", m.Tools[1].Adapter.ID())
}

func TestNewModel_ProgressChannel(t *testing.T) {
	t.Parallel()

	m := app.NewModel("")
	require.NotNil(t, m.Progress, "Progress channel should be allocated")
	assert.Equal(t, 64, cap(m.Progress), "Progress channel buffer capacity should be 64")
}

func TestNewModel_InitReturnsCmd(t *testing.T) {
	t.Parallel()

	m := app.NewModel("")
	cmd := m.Init()
	// Init may return nil (a valid tea.Cmd meaning "no initial command").
	// We just verify it compiles and doesn't panic.
	assert.Nil(t, cmd, "Init returns nil by default (no startup command)")
}

func TestModel_ImplementsBubbleteaModel(t *testing.T) {
	t.Parallel()

	m := app.NewModel("")
	var _ tea.Model = m // compile-time check
	_ = m.Init()
	_, _ = m.Update(nil)
	_ = m.View()
}

func TestWindowSizeMsg_UpdatesDimensions(t *testing.T) {
	t.Parallel()

	m := app.NewModel("")
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	updated, cmd := m.Update(msg)
	require.IsType(t, app.Model{}, updated, "Update should return app.Model")
	m2 := updated.(app.Model)
	assert.Equal(t, 120, m2.Width, "Width should be updated from WindowSizeMsg")
	assert.Equal(t, 40, m2.Height, "Height should be updated from WindowSizeMsg")
	assert.Nil(t, cmd, "WindowSizeMsg should not produce a command")
}

func TestKeyMsg_Q_Quits(t *testing.T) {
	t.Parallel()

	m := app.NewModel("")
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
	t.Parallel()

	m := app.NewModel("")
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	updated, cmd := m.Update(msg)
	require.NotNil(t, cmd, "pressing ctrl+c should return a tea.Cmd")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "pressing ctrl+c should produce tea.QuitMsg")
	_ = updated
}

func TestEmptyModel_CompilesAndRunsWithoutPanic(t *testing.T) {
	t.Parallel()

	m := app.NewModel("")

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
	t.Parallel()

	m := app.NewModel("")
	assert.Equal(t, model.ScreenWelcome, m.Screen, "initial screen should be Welcome")

	// Send NavigateMsg targeting ToolSelection.
	msg := tui.NavigateMsg{Target: model.ScreenToolSelection}
	updated, cmd := m.Update(msg)
	m2 := updated.(app.Model)
	assert.Equal(t, model.ScreenToolSelection, m2.Screen, "NavigateMsg should transition to ToolSelection")
	assert.Nil(t, cmd, "NavigateMsg should return no command")
}

func TestNavigateMsg_ToComplete_TransitionsScreen(t *testing.T) {
	t.Parallel()

	m := app.NewModel("")
	// Manually set to InstallProgress to simulate post-install.
	m.Screen = model.ScreenInstallProgress

	msg := tui.NavigateMsg{Target: model.ScreenComplete}
	updated, cmd := m.Update(msg)
	m2 := updated.(app.Model)
	assert.Equal(t, model.ScreenComplete, m2.Screen, "NavigateMsg should transition to Complete")
	assert.Nil(t, cmd, "NavigateMsg should return no command")
}

func TestUnknownMsg_DoesNotPanic(t *testing.T) {
	t.Parallel()

	m := app.NewModel("")

	type unknownMsg struct{}
	require.NotPanics(t, func() {
		_, _ = m.Update(unknownMsg{})
	}, "Update should not panic on unknown message types")
}

func TestWelcomeView_RendersContent(t *testing.T) {
	t.Parallel()

	m := app.NewModel("")
	view := m.View()

	// Should NOT be the default placeholder — screens are wired.
	assert.NotEqual(t, "Sequoia TUI — screen not yet implemented", view,
		"Welcome screen should render real content, not placeholder")
	assert.Contains(t, view, "Sequoia",
		"Welcome view should contain branding")
}

func TestWelcomeView_EnterNavigatesToToolSelection(t *testing.T) {
	t.Parallel()

	m := app.NewModel("")
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
	t.Parallel()

	// Register mock adapters, then navigate to ToolSelection.
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original }()

	reg.Register(&mockAdapter{id: "claude-code", name: "Claude Code"})
	reg.Register(&mockAdapter{id: "opencode", name: "OpenCode"})

	m := app.NewModel("")
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
	t.Parallel()

	m := app.NewModel("")
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
	t.Parallel()

	// Register one tool, none selected.
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original }()

	reg.Register(&mockAdapter{id: "claude-code", name: "Claude Code"})

	m := app.NewModel("")
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
	t.Parallel()

	// Register one tool and select it.
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original }()

	reg.Register(&mockAdapter{id: "claude-code", name: "Claude Code"})

	m := app.NewModel("")
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

func TestToolSelection_SpaceToggles(t *testing.T) {
	t.Parallel()

	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original }()

	reg.Register(&mockAdapter{id: "claude-code", name: "Claude Code"})

	m := app.NewModel("")
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
