package app

import (
	"github.com/Crisbr10/sequoia/internal/i18n"
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/screens"
)

// View renders the TUI by delegating to the active screen's render function.
// Each screen returns a string that Bubbletea paints to the terminal.
func (m Model) View() string {
	lang := string(m.Config.Language)
	switch m.Screen {
	case model.ScreenWelcome:
		return screens.WelcomeView(m.Version, m.Cursor, lang)
	case model.ScreenToolSelection:
		return screens.ToolSelectionView(m.Tools, m.Cursor, m.ErrorMsg, lang)
	case model.ScreenConfiguration:
		return screens.ConfigurationView(m.Config, m.Cursor, m.EngramAvailable, lang)
	case model.ScreenInstallProgress:
		return screens.InstallProgressView(m.ProgressTools, m.InstallCompleted, len(m.ProgressTools), m.OperationMode, lang)
	case model.ScreenComplete:
		return screens.CompleteView(m.ProgressTools, m.OperationMode, m.InstallWarned, lang)
	case model.ScreenError:
		return screens.ErrorView(m.ProgressTools, m.OperationMode, lang)
	case model.ScreenStatus:
		return screens.StatusView(m.Tools, m.Cursor, lang)
	case model.ScreenUninstall:
		view := screens.UninstallView(m.Tools, m.Cursor, m.ErrorMsg, lang)
		if m.UninstallConfirming {
			view += "\n" + screens.RenderConfirmPrompt(lang)
		}
		return view
	default:
		return i18n.T(i18n.MsgDefaultPlaceholder, lang)
	}
}
