package app

import (
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/pipeline"
	"github.com/Crisbr10/sequoia/internal/tui"
	"github.com/Crisbr10/sequoia/internal/tui/screens"

	tea "github.com/charmbracelet/bubbletea"
)

// quitCmd returns the tea.Quit command for screen-level quit branches
// after marking the model as quitting and cancelling the pipeline context.
// REQ-TUI-01 requires every screen-level quit branch to invoke m.cancel();
// centralizing the pattern here guarantees it is never accidentally
// dropped during future refactors of updateScreenKey.
func (m *Model) quitCmd() tea.Cmd {
	m.Quitting = true
	m.cancel()
	return tea.Quit
}

// Update dispatches incoming messages to the appropriate handler based on
// the current Screen. Global keybindings (q, ctrl+c, WindowSizeMsg) are
// handled at the top before screen-specific delegation.
//
// REQ-TUI-07: key dispatch is delegated to the Router in
// internal/tui/router.go. The Router operates on a KeyHandler interface
// (satisfied by *Model) and routes the message to the per-screen handler
// for the active Screen. The current PR7 commit 2 (GREEN shell) keeps
// the legacy updateScreenKey function as the dispatch surface via the
// temporary UpdateScreenKey bridge method on KeyHandler; commits 3-9
// migrate each screen to a per-screen handleX method on the Router;
// commit 10 removes the legacy function.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tui.NavigateMsg:
		m.Cursor = 0
		m.ErrorMsg = ""
		m.PreviousScreen = m.Screen
		m.Screen = msg.Target
		return m, nil

	case tea.KeyMsg:
		// Global quit keybindings.
		if msg.String() == "q" || msg.Type == tea.KeyCtrlC {
			return m, m.quitCmd()
		}

		// Delegate key dispatch to the Router. After commits 3-9, the
		// Router's per-screen handleX methods will mutate m in place via
		// the pointer; in commit 2 (GREEN shell), the Router delegates
		// back to the legacy updateScreenKey which returns a modified
		// copy. The type assertion below handles both cases: when the
		// Router returns *app.Model (pointer, future commits), the
		// assertion fails and we keep the in-place mutation; when it
		// returns app.Model (value, commit 2 shell), the assertion
		// succeeds and we update m to the modified copy.
		var router tui.Router
		m2, cmd := router.DispatchKey(&m, msg)
		if modelVal, ok := m2.(Model); ok {
			m = modelVal
		}
		return m, cmd
	}

	// Delegate non-key messages to screen-specific handler.
	return m.updateScreenMsg(msg)
}

// updateScreenKey delegates key messages to the active screen's handler.
//
// PR7 commits 3-8 (migrated Welcome, ToolSelection, InstallProgress,
// Complete, Error, Status): the corresponding cases have been extracted
// to per-screen handleX methods on the Router. PR7 commit 9 (this
// commit) extracts the ScreenUninstall case to Router.handleUninstall.
// The Router's DispatchKey now routes ALL seven screens through their
// respective handleX methods. This function is now a thin shell that
// only contains the default case (for any future screen not yet
// handled by the Router); PR7 commit 10 will delete it entirely along
// with the temporary UpdateScreenKey bridge.
func (m Model) updateScreenKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	_ = msg
	return m, nil
}

// updateScreenMsg delegates non-key messages to the active screen's handler.
func (m Model) updateScreenMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.Screen {
	case model.ScreenInstallProgress:
		if progressMsg, ok := msg.(model.ProgressMsg); ok {
			newTools, completed, hasNewFailure, hasNewWarning := screens.ApplyProgressMsg(m.ProgressTools, progressMsg)
			m.ProgressTools = newTools
			m.InstallCompleted = completed
			if hasNewFailure {
				m.InstallFailed++
			}
			if hasNewWarning {
				m.InstallWarned++
			}

			// Check for auto-transition after applying progress.
			action := screens.InstallProgressUpdate(nil, m.InstallCompleted, m.InstallFailed, len(m.ProgressTools))
			switch action {
			case "success":
				return m, func() tea.Msg {
					return tui.NavigateMsg{Target: model.ScreenComplete}
				}
			case "fail":
				return m, func() tea.Msg {
					return tui.NavigateMsg{Target: model.ScreenError}
				}
			}

			// Keep polling the channel for more progress messages.
			return m, waitForProgress(m.Progress)
		}
		return m, nil
	default:
		return m, nil
	}
}

// hasSelectedInstalled returns true if at least one tool that is both
// selected and installed exists.
func hasSelectedInstalled(tools []model.ToolState) bool {
	for _, t := range tools {
		if t.Selected && t.Adapter.IsInstalled() {
			return true
		}
	}
	return false
}

// countSelected returns the number of tools with Selected=true.
func countSelected(tools []model.ToolState) int {
	n := 0
	for _, t := range tools {
		if t.Selected {
			n++
		}
	}
	return n
}

// buildProgressTools creates the initial progress state for selected tools.
// Each tool gets a single "Installing" step in pending state, matching the
// single-step pipeline (pipeline.InstallSteps).
func buildProgressTools(tools []model.ToolState) []screens.ProgressTool {
	stepNames := pipeline.InstallSteps
	var result []screens.ProgressTool
	for _, ts := range tools {
		if !ts.Selected {
			continue
		}
		steps := make([]screens.ProgressStep, len(stepNames))
		for i, name := range stepNames {
			steps[i] = screens.ProgressStep{
				Name:   name,
				Status: screens.StepPending,
			}
		}
		result = append(result, screens.ProgressTool{
			ToolID:   ts.Adapter.ID(),
			ToolName: ts.Adapter.Name(),
			Steps:    steps,
		})
	}

	result = append(result, screens.ProgressTool{
		ToolID:   "codegraph",
		ToolName: "CodeGraph",
		Steps: []screens.ProgressStep{
			{Name: "Installing", Status: screens.StepPending},
			{Name: "Configuring", Status: screens.StepPending},
		},
	})

	return result
}

// buildUninstallProgressTools creates progress state for uninstall.
// Only tools that are BOTH selected and installed are included.
// Uses a single "Uninstalling" step to match the single-phase uninstall
// dispatched by runUninstallSteps in the pipeline.
func buildUninstallProgressTools(tools []model.ToolState) []screens.ProgressTool {
	stepNames := []string{"Uninstalling"}
	var result []screens.ProgressTool
	for _, ts := range tools {
		if !ts.Selected || !ts.Adapter.IsInstalled() {
			continue
		}
		steps := make([]screens.ProgressStep, len(stepNames))
		for i, name := range stepNames {
			steps[i] = screens.ProgressStep{
				Name:   name,
				Status: screens.StepPending,
			}
		}
		result = append(result, screens.ProgressTool{
			ToolID:   ts.Adapter.ID(),
			ToolName: ts.Adapter.Name(),
			Steps:    steps,
		})
	}
	return result
}

// startPipeline builds the progress state, starts the pipeline (install or
// uninstall), and returns the batched tea commands that begin execution.
// All entry points to the InstallProgress screen MUST use this method so
// that ProgressTools, counters, and polling are set up consistently.
func (m *Model) startPipeline(mode string) tea.Cmd {
	// Allocate a fresh channel on every invocation. The previous channel
	// (if any) was closed by the prior pipeline run. Reusing it would cause
	// a panic in sendProgress. See REQ-BUG-002.
	m.Progress = make(chan model.ProgressMsg, model.ProgressChannelBufferSize)

	if mode == "install" {
		m.OperationMode = "install"
		m.ProgressTools = buildProgressTools(m.Tools)
	} else {
		m.OperationMode = "uninstall"
		m.ProgressTools = buildUninstallProgressTools(m.Tools)
	}
	m.InstallCompleted = 0
	m.InstallFailed = 0
	m.InstallWarned = 0

	navigateCmd := func() tea.Msg {
		return tui.NavigateMsg{Target: model.ScreenInstallProgress}
	}

	var pipelineCmd tea.Cmd
	if mode == "install" {
		pipelineCmd = pipeline.RunInstall(m.ctx, m.Tools, m.Progress)
	} else {
		pipelineCmd = pipeline.RunUninstall(m.ctx, m.Tools, m.Progress)
	}

	return tea.Batch(navigateCmd, pipelineCmd, waitForProgress(m.Progress))
}

// waitForProgress returns a tea.Cmd that reads the next model.ProgressMsg
// from the buffered channel. When the channel is closed (and drained),
// it returns nil, stopping the polling loop.
func waitForProgress(ch <-chan model.ProgressMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil // channel closed — stop polling
		}
		return msg
	}
}
