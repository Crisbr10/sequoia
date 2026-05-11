// Package app_test contains integration tests for the full TUI install flow,
// driving the Model through screen transitions and verifying view output.
package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/internal/app"
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/screens"

	tea "github.com/charmbracelet/bubbletea"
)

// sendKey sends a key message through Update and chains any resulting
// commands (except waitForProgress — skipped to avoid channel blocking).
func sendKey(m app.Model, key tea.KeyMsg, maxCmd int) app.Model {
	updated, cmd := m.Update(key)
	current := updated.(app.Model)
	return safeProcessCmd(current, cmd, maxCmd)
}

// safeProcessCmd recursively executes a tea.Cmd and feeds results through
// Update. It skips nil and returns when max iterations are reached.
func safeProcessCmd(m app.Model, cmd tea.Cmd, remaining int) app.Model {
	if cmd == nil || remaining <= 0 {
		return m
	}
	result := cmd()
	if result == nil {
		return m
	}
	updated, nextCmd := m.Update(result)
	m = updated.(app.Model)
	return safeProcessCmd(m, nextCmd, remaining-1)
}

func TestIntegration_FullInstallFlow_ScreenSequence(t *testing.T) {
	// Setup: register a mock adapter.
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&mockAdapter{id: "test-tool", name: "Test Tool"})

	m := app.NewModel("")

	// Verify initial state.
	assert.Equal(t, model.ScreenWelcome, m.Screen)
	assert.Contains(t, m.View(), "Menu")

	// Welcome → ToolSelection.
	m = sendKey(m, tea.KeyMsg{Type: tea.KeyEnter}, 3)
	assert.Equal(t, model.ScreenToolSelection, m.Screen)

	// ToolSelection → Configuration (tool is selected by default).
	m = sendKey(m, tea.KeyMsg{Type: tea.KeyEnter}, 3)
	assert.Equal(t, model.ScreenConfiguration, m.Screen)
	assert.Contains(t, m.View(), "Language")

	// Configuration confirm builds ProgressTools.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(app.Model)
	require.NotEmpty(t, m.ProgressTools, "ProgressTools should be populated from confirm")
	assert.Equal(t, "Test Tool", m.ProgressTools[0].ToolName)
}

func TestIntegration_ProgressToComplete_Transition(t *testing.T) {
	// Test the progress → complete transition by directly applying
	// ProgressMsg to the model state and checking the auto-transition.
	// We bypass the full Update path to avoid blocking on waitForProgress.
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&mockAdapter{id: "test-tool", name: "Test Tool"})

	m := app.NewModel("")
	m.Screen = model.ScreenInstallProgress

	// Populate ProgressTools with 3 pending steps.
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
	m.InstallCompleted = 0
	m.InstallFailed = 0

	// Mark all steps as done using ApplyProgressMsg directly.
	steps := []string{"Skills", "Commands", "System Prompt"}
	for _, step := range steps {
		newTools, completed, hasNewFailure := screens.ApplyProgressMsg(m.ProgressTools, model.ProgressMsg{
			ToolID: "test-tool", Step: step, Done: true, Error: "",
		})
		m.ProgressTools = newTools
		m.InstallCompleted = completed
		if hasNewFailure {
			m.InstallFailed++
		}
	}

	// Check auto-transition — all tools complete, no failures.
	action := screens.InstallProgressUpdate(nil, m.InstallCompleted, m.InstallFailed, len(m.ProgressTools))
	assert.Equal(t, "success", action, "all tools complete should trigger success transition")

	// Verify Complete view renders when screen is set.
	m.Screen = model.ScreenComplete
	view := m.View()
	assert.Contains(t, view, "Complete", "Complete view should render")
}

func TestIntegration_ProgressWithError_Transition(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&mockAdapter{id: "fail-tool", name: "Fail Tool"})

	m := app.NewModel("")
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

	// Apply an error progress message.
	newTools, completed, hasNewFailure := screens.ApplyProgressMsg(m.ProgressTools, model.ProgressMsg{
		ToolID: "fail-tool", Step: "Commands", Done: true,
		Error: "permission denied",
	})
	m.ProgressTools = newTools
	m.InstallCompleted = completed
	if hasNewFailure {
		m.InstallFailed++
	}

	// Check that failedCount is incremented.
	assert.Equal(t, 1, m.InstallFailed, "error should increment failed count")

	// Check auto-transition: 1 tool total, 0 completed, 1 failed → fail.
	action := screens.InstallProgressUpdate(nil, m.InstallCompleted, m.InstallFailed, len(m.ProgressTools))
	assert.Equal(t, "fail", action, "failed tool should trigger fail transition")

	// Verify Error view.
	m.Screen = model.ScreenError
	view := m.View()
	assert.Contains(t, view, "Failed", "Error view should show failure")
	assert.Contains(t, view, "permission denied", "Error view should show error message")
}

func TestIntegration_FlowWithNoToolsSelected_ShowsError(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&mockAdapter{id: "test-tool", name: "Test Tool"})

	m := app.NewModel("")
	m = sendKey(m, tea.KeyMsg{Type: tea.KeyEnter}, 3)
	require.Equal(t, model.ScreenToolSelection, m.Screen)

	// Deselect the tool.
	m.Tools[0].Selected = false

	// Try to advance with no selection.
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := updated.(app.Model)

	assert.Nil(t, cmd, "Enter with no selection should not navigate")
	assert.NotEmpty(t, m2.ErrorMsg, "should show error when no tools selected")
	assert.Equal(t, model.ScreenToolSelection, m2.Screen, "should stay on ToolSelection")
}

func TestIntegration_WelcomeView_ShowsMenu(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&mockAdapter{id: "tool-a", name: "Tool A"})
	reg.Register(&mockAdapter{id: "tool-b", name: "Tool B"})

	m := app.NewModel("")
	view := m.View()

	// Welcome screen shows the main menu, not the individual tool list.
	assert.Contains(t, view, "Install", "Welcome should show Install menu option")
	assert.Contains(t, view, "Status", "Welcome should show Status menu option")
	assert.Contains(t, view, "Uninstall", "Welcome should show Uninstall menu option")
	assert.Equal(t, model.ScreenWelcome, m.Screen)
}

func TestIntegration_CompleteView_ShowsNextSteps(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&mockAdapter{id: "test-tool", name: "Test Tool"})

	m := app.NewModel("")
	m.Screen = model.ScreenComplete

	view := m.View()
	assert.Contains(t, view, "Complete", "Complete view should render")
	assert.Contains(t, view, "sequoia status", "Complete should show next command hint")
}

func TestIntegration_QuitFromAnyScreen_ReturnsQuitMsg(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&mockAdapter{id: "test-tool", name: "Test Tool"})

	m := app.NewModel("")

	// Test quit from Welcome.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	require.NotNil(t, cmd)
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "q should produce tea.QuitMsg from Welcome")

	// Test quit from other screens.
	m.Screen = model.ScreenToolSelection
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	require.NotNil(t, cmd)
	result = cmd()
	_, ok = result.(tea.QuitMsg)
	assert.True(t, ok, "q should produce tea.QuitMsg from any screen")
}

func TestIntegration_ErrorScreen_RetryReturnsToProgress(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&mockAdapter{id: "test-tool", name: "Test Tool"})

	m := app.NewModel("")
	m.Screen = model.ScreenError

	// Press 'r' for retry.
	m = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}, 3)
	// The Error screen's 'r' navigates to InstallProgress (for retry).
	assert.Equal(t, model.ScreenInstallProgress, m.Screen,
		"r on Error should navigate to InstallProgress for retry")
}

func TestIntegration_ErrorRecovery_FullFlow(t *testing.T) {
	// Full error recovery flow test:
	// 1. Simulate: install fails → Error screen → retry (r key) → InstallProgress → Complete
	// 2. Verify screen transitions: InstallProgress → Error → InstallProgress → Complete
	// 3. Verify error messages are displayed on Error screen

	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&mockAdapter{id: "fail-tool", name: "Fail Tool"})

	m := app.NewModel("")
	m.Screen = model.ScreenInstallProgress

	// === PHASE 1: Simulate a partial installation with a failure ===
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
	m.InstallCompleted = 0
	m.InstallFailed = 1

	// Verify auto-transition to Error screen.
	action := screens.InstallProgressUpdate(nil, m.InstallCompleted, m.InstallFailed, len(m.ProgressTools))
	assert.Equal(t, "fail", action, "failed tool should trigger fail transition")
	m.Screen = model.ScreenError

	// === PHASE 2: Verify Error screen displays the failure ===
	errorView := m.View()
	assert.Contains(t, errorView, "Failed", "Error screen should show failure heading")
	assert.Contains(t, errorView, "disk full", "Error screen should show the error message")
	assert.Contains(t, errorView, "Fail Tool", "Error screen should list the failed tool")
	assert.Contains(t, errorView, "r", "Error screen should show retry option")

	// === PHASE 3: Retry — press 'r' to go back to InstallProgress ===
	m = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}, 3)
	assert.Equal(t, model.ScreenInstallProgress, m.Screen,
		"r on Error should navigate to InstallProgress for retry")

	// === PHASE 4: Simulate successful retry — reset progress and apply success ===
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

	// Apply progress messages for all steps as successful.
	steps := []string{"Skills", "Commands", "System Prompt"}
	for _, step := range steps {
		newTools, completed, hasNewFailure := screens.ApplyProgressMsg(m.ProgressTools, model.ProgressMsg{
			ToolID: "fail-tool", Step: step, Done: true, Error: "",
		})
		m.ProgressTools = newTools
		m.InstallCompleted = completed
		if hasNewFailure {
			m.InstallFailed++
		}
	}

	// Verify auto-transition to Complete after all steps succeed.
	action = screens.InstallProgressUpdate(nil, m.InstallCompleted, m.InstallFailed, len(m.ProgressTools))
	assert.Equal(t, "success", action, "retry success should trigger success transition")
	m.Screen = model.ScreenComplete

	// === PHASE 5: Verify Complete screen ===
	completeView := m.View()
	assert.Contains(t, completeView, "Complete", "Complete screen should render after retry success")
	assert.Contains(t, completeView, "Fail Tool", "Complete screen should list the tool")
	assert.Contains(t, completeView, "✅", "Complete screen should show success indicators")
}

func TestIntegration_ErrorRecovery_MultipleFailuresRetryAll(t *testing.T) {
	// Simulate two tools failing, then retry succeeds for both.

	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&mockAdapter{id: "tool-a", name: "Tool A"})
	reg.Register(&mockAdapter{id: "tool-b", name: "Tool B"})

	m := app.NewModel("")
	m.Screen = model.ScreenInstallProgress

	m.ProgressTools = []screens.ProgressTool{
		{
			ToolID:   "tool-a",
			ToolName: "Tool A",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepFailed, Error: "permission denied"},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
		{
			ToolID:   "tool-b",
			ToolName: "Tool B",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepDone},
				{Name: "Commands", Status: screens.StepFailed, Error: "mkdir failed"},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}
	m.InstallCompleted = 0
	m.InstallFailed = 2

	// Both failed → Error screen.
	m.Screen = model.ScreenError
	errorView := m.View()
	assert.Contains(t, errorView, "Tool A", "Error screen should show Tool A")
	assert.Contains(t, errorView, "Tool B", "Error screen should show Tool B")
	assert.Contains(t, errorView, "permission denied", "Error screen should show Tool A's error")
	assert.Contains(t, errorView, "mkdir failed", "Error screen should show Tool B's error")

	// Retry.
	m = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}, 3)
	assert.Equal(t, model.ScreenInstallProgress, m.Screen)

	// Reset for retry.
	m.ProgressTools = []screens.ProgressTool{
		{
			ToolID:   "tool-a",
			ToolName: "Tool A",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepPending},
				{Name: "Commands", Status: screens.StepPending},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
		{
			ToolID:   "tool-b",
			ToolName: "Tool B",
			Steps: []screens.ProgressStep{
				{Name: "Skills", Status: screens.StepPending},
				{Name: "Commands", Status: screens.StepPending},
				{Name: "System Prompt", Status: screens.StepPending},
			},
		},
	}
	m.InstallCompleted = 0
	m.InstallFailed = 0

	// Apply success for both tools.
	for _, toolID := range []string{"tool-a", "tool-b"} {
		for _, step := range []string{"Skills", "Commands", "System Prompt"} {
			newTools, completed, hasNewFailure := screens.ApplyProgressMsg(m.ProgressTools, model.ProgressMsg{
				ToolID: toolID, Step: step, Done: true, Error: "",
			})
			m.ProgressTools = newTools
			m.InstallCompleted = completed
			if hasNewFailure {
				m.InstallFailed++
			}
		}
	}

	action := screens.InstallProgressUpdate(nil, m.InstallCompleted, m.InstallFailed, len(m.ProgressTools))
	assert.Equal(t, "success", action, "retry of both tools should trigger success")
	m.Screen = model.ScreenComplete

	completeView := m.View()
	assert.Contains(t, completeView, "Complete")
	assert.Contains(t, completeView, "Tool A")
	assert.Contains(t, completeView, "Tool B")
}

func TestIntegration_StatusAndUninstall_Flow(t *testing.T) {
	reg := &adapters.Registry{}
	original := adapters.DefaultRegistry
	registryMu.Lock()
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = original; registryMu.Unlock() }()

	reg.Register(&mockAdapter{id: "test-tool", name: "Test Tool"})

	m := app.NewModel("")
	m.Screen = model.ScreenStatus

	view := m.View()
	assert.Contains(t, view, "Test Tool", "Status should show tool name")

	// Press 'd' for uninstall.
	m = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, 3)
	assert.Equal(t, model.ScreenUninstall, m.Screen, "d should navigate to Uninstall")
}
