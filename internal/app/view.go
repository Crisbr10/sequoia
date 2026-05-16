package app

import (
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/screens"
)

// View renders the TUI by delegating to the active screen's render function.
// Each screen returns a string that Bubbletea paints to the terminal.
func (m Model) View() string {
	switch m.Screen {
	case model.ScreenWelcome:
		return screens.WelcomeView(m.Version, m.Cursor)
	case model.ScreenToolSelection:
		return screens.ToolSelectionView(m.Tools, m.Cursor, m.ErrorMsg)
	case model.ScreenConfiguration:
		return screens.ConfigurationView(m.Config, m.Cursor, m.EngramAvailable)
	case model.ScreenInstallProgress:
		return screens.InstallProgressView(m.ProgressTools, m.InstallCompleted, len(m.ProgressTools), m.OperationMode)
	case model.ScreenComplete:
		return screens.CompleteView(m.ProgressTools, m.OperationMode, m.InstallWarned)
	case model.ScreenError:
		return screens.ErrorView(m.ProgressTools, m.OperationMode)
	case model.ScreenStatus:
		return screens.StatusView(m.Tools, m.Cursor)
	case model.ScreenUninstall:
		view := screens.UninstallView(m.Tools, m.Cursor, m.ErrorMsg)
		if m.UninstallConfirming {
			view += "\n" + screens.RenderConfirmPrompt()
		}
		return view
	default:
		return "Sequoia TUI — screen not yet implemented"
	}
}
