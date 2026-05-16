package i18n_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/internal/i18n"
)

// allKeys is the complete list of message key constants defined in keys.go.
// If a key is added to keys.go, it MUST be added to this slice so the test
// catches keys that are defined in code but missing from one of the catalogs.
var allKeys = []string{
	// Welcome screen
	i18n.MsgWelcomeMenuInstall,
	i18n.MsgWelcomeMenuStatus,
	i18n.MsgWelcomeMenuUninstall,
	i18n.MsgWelcomeMenuQuit,
	i18n.MsgWelcomeSubtitle,
	i18n.MsgWelcomeFooter,

	// Tool selection screen
	i18n.MsgToolSelectionTitle,
	i18n.MsgToolSelectionInstruction,
	i18n.MsgToolSelectionEmpty,
	i18n.MsgToolSelectionSelectedCount,

	// Configuration screen
	i18n.MsgConfigurationTitle,
	i18n.MsgConfigurationLanguageLabel,
	i18n.MsgConfigurationLanguageEN,
	i18n.MsgConfigurationLanguageES,
	i18n.MsgConfigurationPersistenceLabel,
	i18n.MsgConfigurationPersistenceEngram,
	i18n.MsgConfigurationPersistenceFiles,
	i18n.MsgConfigurationPersistenceBoth,
	i18n.MsgConfigurationEngramUnavailable,

	// Install progress screen
	i18n.MsgInstallProgressTitleInstall,
	i18n.MsgInstallProgressTitleUninstall,
	i18n.MsgInstallProgressSummaryInstall,
	i18n.MsgInstallProgressSummaryUninstall,

	// Complete screen
	i18n.MsgCompleteHeadingInstall,
	i18n.MsgCompleteHeadingUninstall,
	i18n.MsgCompleteHeadingUninstallWarnings,
	i18n.MsgCompleteInstalledItems,
	i18n.MsgCompleteUninstalledItems,
	i18n.MsgCompleteWarningsNote,
	i18n.MsgCompleteTryCommand,

	// Error screen
	i18n.MsgErrorHeadingInstall,
	i18n.MsgErrorHeadingUninstall,

	// Status screen
	i18n.MsgStatusTitle,
	i18n.MsgStatusEmpty,

	// Uninstall screen
	i18n.MsgUninstallTitle,
	i18n.MsgUninstallEmpty,
	i18n.MsgUninstallConfirmPrompt,
	i18n.MsgUninstallConfirmSuffix,

	// Validation messages
	i18n.MsgValidationSelectAtLeastOne,
	i18n.MsgValidationSelectAtLeastOneInstalled,

	// Default / fallback
	i18n.MsgDefaultPlaceholder,

	// Footer hints
	i18n.MsgFooterNavigateKeys,
	i18n.MsgFooterNavigate,
	i18n.MsgFooterToggleKey,
	i18n.MsgFooterToggle,
	i18n.MsgFooterConfirmKey,
	i18n.MsgFooterConfirm,
	i18n.MsgFooterBackKey,
	i18n.MsgFooterBack,
	i18n.MsgFooterQuitKey,
	i18n.MsgFooterQuit,
	i18n.MsgFooterTabKey,
	i18n.MsgFooterSwitchField,
	i18n.MsgFooterArrowsKeys,
	i18n.MsgFooterChangeOption,
	i18n.MsgFooterStatusScreenKey,
	i18n.MsgFooterStatusScreen,
	i18n.MsgFooterRetryFailed,
	i18n.MsgFooterBackToTools,
	i18n.MsgFooterQuitLabel,
	i18n.MsgFooterUpdateKey,
	i18n.MsgFooterUpdateLabel,
	i18n.MsgFooterReinstallKey,
	i18n.MsgFooterReinstallLabel,
	i18n.MsgFooterUninstallKey,
	i18n.MsgFooterUninstallLabel,
}

// TestKeys_AllExistInEnglishCatalog verifies that every key constant in
// keys.go has a corresponding entry in the English TOML catalog.
func TestKeys_AllExistInEnglishCatalog(t *testing.T) {
	t.Parallel()

	err := i18n.Init()
	require.NoError(t, err, "Init must succeed before key validation")

	for _, key := range allKeys {
		result := i18n.T(key, "en")
		assert.NotEqual(t, key, result,
			"key %q missing from English catalog: T() returned the key itself", key)
	}
}

// TestKeys_AllExistInSpanishCatalog verifies that every key constant in
// keys.go has a corresponding entry in the Spanish TOML catalog.
func TestKeys_AllExistInSpanishCatalog(t *testing.T) {
	t.Parallel()

	err := i18n.Init()
	require.NoError(t, err, "Init must succeed before key validation")

	for _, key := range allKeys {
		result := i18n.T(key, "es")
		assert.NotEqual(t, key, result,
			"key %q missing from Spanish catalog: T() returned the key itself", key)
	}
}

// TestKeys_TotalCount verifies we have the expected number of key constants.
// Update this count when adding or removing keys.
func TestKeys_TotalCount(t *testing.T) {
	t.Parallel()

	// Expected: 6 welcome + 4 tool_selection + 9 configuration + 4 install_progress
	// + 7 complete + 2 error + 2 status + 4 uninstall + 2 validation + 1 default
	// + 25 footer = 66 keys total
	expectedCount := 66
	assert.Len(t, allKeys, expectedCount,
		"key count mismatch; update allKeys slice and this assertion when adding/removing keys")
}
