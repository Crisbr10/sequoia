package app

import (
	"sequoia-ai/internal/model"
	"sequoia-ai/internal/tui"
	"sequoia-ai/internal/tui/screens"

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
		m.Screen = msg.Target
		return m, nil

	case tea.KeyMsg:
		// Global quit keybindings.
		if msg.String() == "q" || msg.Type == tea.KeyCtrlC {
			m.Quitting = true
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
		return m, screens.WelcomeUpdate(msg)

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
			// Build initial progress state from selected tools.
			m.ProgressTools = buildProgressTools(m.Tools)
			m.InstallCompleted = 0
			m.InstallFailed = 0
			m.ErrorMsg = ""
			return m, func() tea.Msg {
				return tui.NavigateMsg{Target: model.ScreenInstallProgress}
			}
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
		return m, screens.ErrorUpdate(msg)

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
		}
		return m, nil
	default:
		return m, nil
	}
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
			ToolName: ts.Adapter.Name(),
			Steps:    steps,
		})
	}
	return result
}
