package app

import (
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/screens"
	"github.com/Crisbr10/sequoia/internal/tui/styles"
)

// View renders the TUI by delegating to the active screen's render function.
// Each screen returns a string that Bubbletea paints to the terminal.
func (m Model) View() string {
	switch m.Screen {
	case model.ScreenWelcome:
		return screens.WelcomeView(Version, m.Cursor)
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
	case model.ScreenStatus:
		return screens.StatusView(m.Tools, m.Cursor)
	case model.ScreenUninstall:
		view := screens.UninstallView(m.Tools, m.Cursor)
		if m.UninstallConfirming {
			view += "\n" + renderUninstallConfirm()
		}
		return view
	default:
		return "Sequoia TUI — screen not yet implemented"
	}
}

// renderUninstallConfirm returns the confirmation prompt shown when the user
// presses Enter on the Uninstall screen with selected tools.
func renderUninstallConfirm() string {
	return styles.Accent().Render("  Remove Sequoia from selected tools?") +
		styles.Muted().Render(" [y/N]")
}
