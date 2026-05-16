// Package i18n provides internationalization support for the Sequoia TUI installer.
// It wraps go-i18n/v2 with embedded TOML catalogs and a minimal public API:
// Init() to load catalogs, T() to look up messages, and Initialized() to gate
// language-dependent UI features.
package i18n

// Message key constants for every user-facing string in the Sequoia TUI.
// Keys follow dot-notation: "screen.element" for scannability and collision
// prevention. Each constant maps to an entry in the en.toml and es.toml catalogs.

// --- Welcome Screen ---
const (
	MsgWelcomeMenuInstall   = "welcome.menu_install"
	MsgWelcomeMenuStatus    = "welcome.menu_status"
	MsgWelcomeMenuUninstall = "welcome.menu_uninstall"
	MsgWelcomeMenuQuit      = "welcome.menu_quit"
	MsgWelcomeSubtitle      = "welcome.subtitle"
	MsgWelcomeFooter        = "welcome.footer"
)

// --- Tool Selection Screen ---
const (
	MsgToolSelectionTitle         = "tool_selection.title"
	MsgToolSelectionInstruction   = "tool_selection.instruction"
	MsgToolSelectionEmpty         = "tool_selection.empty"
	MsgToolSelectionSelectedCount = "tool_selection.selected_count"
)

// --- Configuration Screen ---
const (
	MsgConfigurationTitle             = "configuration.title"
	MsgConfigurationLanguageLabel     = "configuration.language_label"
	MsgConfigurationLanguageEN        = "configuration.language_en"
	MsgConfigurationLanguageES        = "configuration.language_es"
	MsgConfigurationPersistenceLabel  = "configuration.persistence_label"
	MsgConfigurationPersistenceEngram = "configuration.persistence_engram"
	MsgConfigurationPersistenceFiles  = "configuration.persistence_files"
	MsgConfigurationPersistenceBoth   = "configuration.persistence_both"
	MsgConfigurationEngramUnavailable = "configuration.engram_unavailable"
)

// --- Install Progress Screen ---
const (
	MsgInstallProgressTitleInstall     = "install_progress.title_install"
	MsgInstallProgressTitleUninstall   = "install_progress.title_uninstall"
	MsgInstallProgressSummaryInstall   = "install_progress.summary_install"
	MsgInstallProgressSummaryUninstall = "install_progress.summary_uninstall"
	MsgInstallBackupDir                = "install_progress.backup_dir"
)

// --- Complete Screen ---
const (
	MsgCompleteHeadingInstall           = "complete.heading_install"
	MsgCompleteHeadingUninstall         = "complete.heading_uninstall"
	MsgCompleteHeadingUninstallWarnings = "complete.heading_uninstall_warnings"
	MsgCompleteInstalledItems           = "complete.installed_items"
	MsgCompleteUninstalledItems         = "complete.uninstalled_items"
	MsgCompleteWarningsNote             = "complete.warnings_note"
	MsgCompleteTryCommand               = "complete.try_command"
)

// --- Error Screen ---
const (
	MsgErrorHeadingInstall   = "error.heading_install"
	MsgErrorHeadingUninstall = "error.heading_uninstall"
)

// --- Status Screen ---
const (
	MsgStatusTitle = "status.title"
	MsgStatusEmpty = "status.empty"
)

// --- Uninstall Screen ---
const (
	MsgUninstallTitle         = "uninstall.title"
	MsgUninstallEmpty         = "uninstall.empty"
	MsgUninstallConfirmPrompt = "uninstall.confirm_prompt"
	MsgUninstallConfirmSuffix = "uninstall.confirm_suffix"
)

// --- Common Validation Messages ---
const (
	MsgValidationSelectAtLeastOne          = "validation.select_at_least_one"
	MsgValidationSelectAtLeastOneInstalled = "validation.select_at_least_one_installed"
)

// --- Default / Fallback ---
const (
	MsgDefaultPlaceholder = "default.placeholder"
)

// --- Footer Hints (reusable across screens) ---
const (
	MsgFooterNavigateKeys    = "footer.navigate_keys"
	MsgFooterNavigate        = "footer.navigate"
	MsgFooterToggleKey       = "footer.toggle_key"
	MsgFooterToggle          = "footer.toggle"
	MsgFooterConfirmKey      = "footer.confirm_key"
	MsgFooterConfirm         = "footer.confirm"
	MsgFooterBackKey         = "footer.back_key"
	MsgFooterBack            = "footer.back"
	MsgFooterQuitKey         = "footer.quit_key"
	MsgFooterQuit            = "footer.quit"
	MsgFooterTabKey          = "footer.tab_key"
	MsgFooterSwitchField     = "footer.switch_field"
	MsgFooterArrowsKeys      = "footer.arrows_keys"
	MsgFooterChangeOption    = "footer.change_option"
	MsgFooterStatusScreenKey = "footer.status_screen_key"
	MsgFooterStatusScreen    = "footer.status_screen"
	MsgFooterRetryFailed     = "footer.retry_failed"
	MsgFooterBackToTools     = "footer.back_to_tools"
	MsgFooterQuitLabel       = "footer.quit_label"
	MsgFooterUpdateKey       = "footer.update_key"
	MsgFooterUpdateLabel     = "footer.update_label"
	MsgFooterReinstallKey    = "footer.reinstall_key"
	MsgFooterReinstallLabel  = "footer.reinstall_label"
	MsgFooterUninstallKey    = "footer.uninstall_key"
	MsgFooterUninstallLabel  = "footer.uninstall_label"
)
