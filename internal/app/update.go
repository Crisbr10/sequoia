package app

import (
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/pipeline"
	"github.com/Crisbr10/sequoia/internal/tui"
	"github.com/Crisbr10/sequoia/internal/tui/screens"

	tea "github.com/charmbracelet/bubbletea"
)

// Update dispatches incoming messages to the appropriate handler based on
// the current Screen. Global keybindings (q, ctrl+c, WindowSizeMsg) are
// handled at the top before screen-specific delegation.
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
			m.Quitting = true
			m.cancel()
			return m, tea.Quit
		}

		// Delegate to screen-specific key handler.
		return m.updateScreenKey(msg)
	}

	// Delegate non-key messages to screen-specific handler.
	return m.updateScreenMsg(msg)
}

// updateScreenKey delegates key messages to the active screen's handler.
func (m Model) updateScreenKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.Screen {
	case model.ScreenWelcome:
		newCursor, action := screens.WelcomeUpdate(msg, m.Cursor)
		m.Cursor = newCursor
		switch action {
		case "install":
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenToolSelection}
			}
		case "status":
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenStatus}
			}
		case "uninstall":
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenUninstall}
			}
		case "quit":
			m.Quitting = true
			m.cancel()
			return m, tea.Quit
		}
		return m, nil

	case model.ScreenToolSelection:
		newCursor, shouldToggle, action := screens.ToolSelectionUpdate(msg, m.Cursor, len(m.Tools))
		m.Cursor = newCursor
		if m.Cursor >= 0 && m.Cursor < len(m.Tools) && shouldToggle {
			m.Tools[m.Cursor].Selected = !m.Tools[m.Cursor].Selected
		}

		switch action {
		case "confirm":
			// Validate at least one tool selected.
			selected := countSelected(m.Tools)
			if selected == 0 {
				m.ErrorMsg = "Select at least one tool to continue"
				return m, nil
			}
			m.ErrorMsg = ""
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenConfiguration}
			}
		case "back":
			m.ErrorMsg = ""
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenWelcome}
			}
		case "quit":
			m.Quitting = true
			return m, tea.Quit
		}
		return m, nil

	case model.ScreenConfiguration:
		newActiveField, newConfig, action := screens.ConfigurationUpdate(msg, m.Cursor, m.Config, m.EngramAvailable)
		m.Cursor = newActiveField
		m.Config = newConfig

		switch action {
		case "confirm":
			m.ErrorMsg = ""
			return m, m.startPipeline("install")
		case "back":
			m.ErrorMsg = ""
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenToolSelection}
			}
		case "quit":
			m.Quitting = true
			return m, tea.Quit
		}
		return m, nil

	case model.ScreenInstallProgress:
		action := screens.InstallProgressUpdate(msg, m.InstallCompleted, m.InstallFailed, len(m.ProgressTools))
		switch action {
		case "quit":
			m.Quitting = true
			return m, tea.Quit
		case "success":
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenComplete}
			}
		case "fail":
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenError}
			}
		}
		return m, nil

	case model.ScreenComplete:
		return m, screens.CompleteUpdate(msg)

	case model.ScreenError:
		// Inline handler: rebuild pipeline on retry, not bare navigation.
		switch msg.Type {
		case tea.KeyEsc, tea.KeyLeft:
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenToolSelection}
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			switch msg.Runes[0] {
			case 'r':
				return m, m.startPipeline(m.OperationMode)
			case 'q':
				return m, tea.Quit
			}
		}
		return m, nil

	case model.ScreenStatus:
		newCursor, action := screens.StatusUpdate(msg, m.Cursor, len(m.Tools))
		m.Cursor = newCursor

		switch action {
		case "uninstall":
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenUninstall}
			}
		case "reinstall":
			// Mark all installed tools as selected so the pipeline picks them up.
			for i := range m.Tools {
				if m.Tools[i].Adapter.IsInstalled() {
					m.Tools[i].Selected = true
				}
			}
			return m, m.startPipeline("install")
		case "back":
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenWelcome}
			}
		case "update":
			// Placeholder — update functionality not yet implemented.
		}
		return m, nil

	case model.ScreenUninstall:
		// Confirmation mode: only y, n, and Esc matter.
		if m.UninstallConfirming {
			if msg.Type == tea.KeyEsc {
				m.UninstallConfirming = false
				return m, nil
			}
			if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
				switch msg.Runes[0] {
				case 'y':
					m.UninstallConfirming = false
					m.ErrorMsg = ""
					return m, m.startPipeline("uninstall")
				case 'n':
					m.UninstallConfirming = false
					return m, nil
				}
			}
			return m, nil
		}

		newCursor, shouldToggle, action := screens.UninstallUpdate(msg, m.Cursor, len(m.Tools))
		m.Cursor = newCursor
		if m.Cursor >= 0 && m.Cursor < len(m.Tools) && shouldToggle {
			m.Tools[m.Cursor].Selected = !m.Tools[m.Cursor].Selected
			m.ErrorMsg = ""
		}

		switch action {
		case "confirm":
			// Check at least one tool is selected and installed.
			if hasSelectedInstalled(m.Tools) {
				m.ErrorMsg = ""
				m.UninstallConfirming = true
			} else {
				m.ErrorMsg = "Select at least one installed tool to continue"
			}
			return m, nil
		case "back":
			m.UninstallConfirming = false
			// Navigate back to the source screen (Welcome or Status).
			// PreviousScreen defaults to ScreenWelcome (zero value), which is
			// correct when the user arrived from the Welcome screen directly.
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: m.PreviousScreen}
			}
		}
		return m, nil

	default:
		return m, nil
	}
}

// updateScreenMsg delegates non-key messages to the active screen's handler.
func (m Model) updateScreenMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.Screen {
	case model.ScreenInstallProgress:
		if progressMsg, ok := msg.(model.ProgressMsg); ok {
			newTools, completed, hasNewFailure := screens.ApplyProgressMsg(m.ProgressTools, progressMsg)
			m.ProgressTools = newTools
			m.InstallCompleted = completed
			if hasNewFailure {
				m.InstallFailed++
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
// Each tool gets the standard install steps (Skills, Commands, System Prompt)
// all in pending state.
func buildProgressTools(tools []model.ToolState) []screens.ProgressTool {
	stepNames := []string{"Skills", "Commands", "System Prompt"}
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
	return result
}

// buildUninstallProgressTools creates progress state for uninstall.
// Only tools that are BOTH selected and installed are included.
func buildUninstallProgressTools(tools []model.ToolState) []screens.ProgressTool {
	stepNames := []string{"Skills", "Commands", "System Prompt"}
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
	if mode == "install" {
		m.OperationMode = "install"
		m.ProgressTools = buildProgressTools(m.Tools)
	} else {
		m.OperationMode = "uninstall"
		m.ProgressTools = buildUninstallProgressTools(m.Tools)
	}
	m.InstallCompleted = 0
	m.InstallFailed = 0

	navigateCmd := func() tea.Msg {
		return tui.NavigateMsg{Target: model.ScreenInstallProgress}
	}

	var pipelineCmd tea.Cmd
	if mode == "install" {
		pipelineCmd = pipeline.RunInstall(m.ctx, m.Tools, m.Progress, m.Config.Language)
	} else {
		pipelineCmd = pipeline.RunUninstall(m.ctx, m.Tools, m.Progress, m.Config.Language)
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
