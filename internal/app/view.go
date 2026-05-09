package app

import (
	"sequoia-ai/internal/model"
	"sequoia-ai/internal/tui/screens"
)

// View renders the TUI by delegating to the active screen's render function.
// Each screen returns a string that Bubbletea paints to the terminal.
func (m Model) View() string {
	switch m.Screen {
	case model.ScreenWelcome:
		return screens.WelcomeView(m.Tools, Version)
	case model.ScreenToolSelection:
		return screens.ToolSelectionView(m.Tools, m.Cursor, m.ErrorMsg)
	case model.ScreenConfiguration:
		return screens.ConfigurationView(m.Config, m.Cursor, m.EngramAvailable)
	case model.ScreenInstallProgress:
		return screens.InstallProgressView(m.ProgressTools, m.InstallCompleted, len(m.ProgressTools))
	case model.ScreenComplete:
		return screens.CompleteView(m.ProgressTools)
	case model.ScreenError:
		return screens.ErrorView(m.ProgressTools)
	default:
		return "Sequoia TUI — screen not yet implemented"
	}
}
